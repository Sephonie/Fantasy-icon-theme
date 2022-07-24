// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/text/internal/colltab"
)

// This file contains code for detecting contractions and generating
// the necessary tables.
// Any Unicode Collation Algorithm (UCA) table entry that has more than
// one rune one the left-hand side is called a contraction.
// See http://www.unicode.org/reports/tr10/#Contractions for more details.
//
// We define the following terms:
//   initial:     a rune that appears as the first rune in a contraction.
//   suffix:      a sequence of runes succeeding the initial rune
//                in a given contraction.
//   non-initial: a rune that appears in a suffix.
//
// A rune may be both an initial and a non-initial and may be so in
// many contractions.  An initial may typically also appear by itself.
// In case of ambiguities, the UCA requires we match the longest
// contraction.
//
// Many contraction rules share the same set of possible suffixes.
// We store sets of suffixes in a trie that associates an index with
// each suffix in the set.  This index can be used to look up a
// collation element associated with the (starter rune, suffix) pair.
//
// The trie is defined on a UTF-8 byte sequence.
// The overall trie is represented as an array of ctEntries.  Each node of the trie
// is represented as a subsequence of ctEntries, where each entry corresponds to
// a possible match of a next character in the search string.  An entry
// also includes the length and offset to the next sequence of entries
// to check in case of a match.

const (
	final   = 0
	noIndex = 0xFF
)

// ctEntry associates to a matching byte an offset and/or next sequence of
// bytes to check. A ctEntry c is called final if a match means that the
// longest suffix has been found.  An entry c is final if c.N == 0.
// A single final entry can match a range of characters to an offset.
// A non-final entry always matches a single byte. Note that a non-final
// entry might still resemble a completed suffix.
// Examples:
// The suffix strings "ab" and "ac" can be represented as:
// []ctEntry{
//     {'a', 1, 1, noIndex},  // 'a' by itself does not match, so i is 0xFF.
//     {'b', 'c', 0, 1},   // "ab" -> 1, "ac" -> 2
// }
//
// The suffix strings "ab", "abc", "abd", and "abcd" can be represented as:
// []ctEntry{
//     {'a', 1, 1, noIndex}, // 'a' must be followed by 'b'.
//     {'b', 1, 2, 1},    // "ab" -> 1, may be followed by 'c' or 'd'.
//     {'d', 'd', final, 3},  // "abd" -> 3
//     {'c', 4, 1, 2},    // "abc" -> 2, may be followed by 'd'.
//     {'d', 'd', final, 4},  // "abcd" -> 4
// }
// See genStateTests in contract_test.go for more examples.
type ctEntry struct {
	L uint8 // non-final: byte value to match; final: lowest match in range.
	H uint8 // non-final: relative index to next block; final: highest match in range.
	N uint8 // non-final: length of next block; final: final
	I uint8 // result offset. Will be noIndex if more bytes are needed to complete.
}

// contractTrieSet holds a set of contraction tries. The tries are stored
// consecutively in the entry field.
type contractTrieSet []struct{ l, h, n, i uint8 }

// ctHandle is used to identify a trie in the trie set, consisting in an offset
// in the array and the size of the first node.
type ctHandle struct {
	index, n int
}

// appendTrie adds a new trie for the given suffixes to the trie set and returns
// a handle to it.  The handle will be invalid on error.
func appendTrie(ct *colltab.ContractTrieSet, suffixes []string) (ctHandle, error) {
	es := make([]stridx, len(suffixes))
	for i, s := range suffixes {
		es[i].str = s
	}
	sort.Sort(offsetSort(es))
	for i := range es {
		es[i].index = i + 1
	}
	sort.Sort(genidxSort(es))
	i := len(*ct)
	n, err := genStates(ct, es)
	if err != nil {
		*ct = (*ct)[:i]
		return ctHandle{}, err
	}
	return ctHandle{i, n}, nil
}

// genStates generates ctEntries for a given suffix set and returns
// the number of entries for the first node.
func genStates(ct *colltab.ContractTrieSet, sis []stridx) (int, error) {
	if len(sis) == 0 {
		return 0, fmt.Errorf("genStates: list of suffices must be non-empty")
	}
	start := len(*ct)
	// create entries for differing first bytes.
	for _, si := range sis {
		s := si.str
		if len(s) == 0 {
			continue
		}
		added := false
		c := s[0]
		if len(s) > 1 {
			for j := len(*ct) - 1; j >= start; j-- {
				if (*ct)[j].L == c {
					added = true
					break
				}
			}
			if !added {
				*ct = append(*ct, ctEntry{L: c, I: noIndex})
			}
		} else {
			for j := len(*ct) - 1; j >= start; j-- {
				// Update the offset for longer suffixes with the same byte.
				if (*ct)[j].L == c {
					(*ct)[j].I = uint8(si.index)
					added = true
				}
				// Extend range of final ctEntry, if possible.
				if (*ct)[j].H+1 == c {
					(*ct)[j].H = c
					added = true
				}
			}
			if !added {
				*ct = append(*ct, ctEntry{L: c, H: c, N: final, I: uint8(si.index)})
			}
		}
	}
	n := len(*ct) - start
	// Append nodes f