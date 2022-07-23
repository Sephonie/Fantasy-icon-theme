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
// also includes the length and offset