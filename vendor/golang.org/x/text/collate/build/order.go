// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/internal/colltab"
	"golang.org/x/text/unicode/norm"
)

type logicalAnchor int

const (
	firstAnchor logicalAnchor = -1
	noAnchor                  = 0
	lastAnchor                = 1
)

// entry is used to keep track of a single entry in the collation element table
// during building. Examples of entries can be found in the Default Unicode
// Collation Element Table.
// See http://www.unicode.org/Public/UCA/6.0.0/allkeys.txt.
type entry struct {
	str    string // same as string(runes)
	runes  []rune
	elems  []rawCE // the collation elements
	extend string  // weights of extend to be appended to elems
	before bool    // weights relative to next instead of previous.
	lock   bool    // entry is used in extension and can no longer be moved.

	// prev, next, and level are used to keep track of tailorings.
	prev, next *entry
	level      colltab.Level // next differs at this level
	skipRemove bool          // do not unlink when removed

	decompose bool // can use NFKD decomposition to generate elems
	exclude   bool // do not include in table
	implicit  bool // derived, is not included in the list
	modified  bool // entry was modified in tailoring
	logical   logicalAnchor

	expansionIndex    int // used to store index into expansion table
	contractionHandle ctHandle
	contractionIndex  int // index into contraction elements
}

func (e *entry) String() string {
	return fmt.Sprintf("%X (%q) -> %X (ch:%x; ci:%d, ei:%d)",
		e.runes, e.str, e.elems, e.contractionHandle, e.contractionIndex, e.expansionIndex)
}

func (e *entry) skip() bool {
	return e.contraction()
}

func (e *entry) expansion() bool {
	return !e.decompose && len(e.elems) > 1
}

func (e *entry) contraction() bool {
	return len(e.runes) > 1
}

func (e *entry) contractionStarter() bool {
	return e.contractionHandle.n != 0
}

// nextIndexed gets the next entry that needs to be stored in the table.
// It returns the entry and the collation level at which the next entry differs
// from the current entry.
// Entries that can be explicitly derived and logical reset positions are
// examples of entries that will not be indexed.
func (e *entry) nextIndexed() (*entry, colltab.Level) {
	level := e.level
	for e = e.next; e != nil && (e.exclude || len(e.elems) == 0); e = e.next {
		if e.level < level {
			level = e.level
		}
	}
	return e, level
}

// remove unlinks entry e from the sorted chain and clears the collation
// elemen