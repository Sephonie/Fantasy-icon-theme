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
// elements. e may not be at the front or end of the list. This should always
// be the case, as the front and end of the list are always logical anchors,
// which may not be removed.
func (e *entry) remove() {
	if e.logical != noAnchor {
		log.Fatalf("may not remove anchor %q", e.str)
	}
	// TODO: need to set e.prev.level to e.level if e.level is smaller?
	e.elems = nil
	if !e.skipRemove {
		if e.prev != nil {
			e.prev.next = e.next
		}
		if e.next != nil {
			e.next.prev = e.prev
		}
	}
	e.skipRemove = false
}

// insertAfter inserts n after e.
func (e *entry) insertAfter(n *entry) {
	if e == n {
		panic("e == anchor")
	}
	if e == nil {
		panic("unexpected nil anchor")
	}
	n.remove()
	n.decompose = false // redo decomposition test

	n.next = e.next
	n.prev = e
	if e.next != nil {
		e.next.prev = n
	}
	e.next = n
}

// insertBefore inserts n before e.
func (e *entry) insertBefore(n *entry) {
	if e == n {
		panic("e == anchor")
	}
	if e == nil {
		panic("unexpected nil anchor")
	}
	n.remove()
	n.decompose = false // redo decomposition test

	n.prev = e.prev
	n.next = e
	if e.prev != nil {
		e.prev.next = n
	}
	e.prev = n
}

func (e *entry) encodeBase() (ce uint32, err error) {
	switch {
	case e.expansion():
		ce, err = makeExpandIndex(e.expansionIndex)
	default:
		if e.decompose {
			log.Fatal("decompose should be handled elsewhere")
		}
		ce, err = makeCE(e.elems[0])
	}
	return
}

func (e *entry) encode() (ce uint32, err error) {
	if e.skip() {
		log.Fatal("cannot build colElem for entry that should be skipped")
	}
	switch {
	case e.decompose:
		t1 := e.elems[0].w[2]
		t2 := 0
		if len(e.elems) > 1 {
			t2 = e.elems[1].w[2]
		}
		ce, err = makeDecompose(t1, t2)
	case e.contractionStarter():
		ce, err = makeContractIndex(e.contractionHandle, e.contractionIndex)
	default:
		if len(e.runes) > 1 {
			log.Fatal("colElem: contractions are handled in contraction trie")
		}
		ce, err = e.encodeBase()
	}
	return
}

// entryLess returns true if a sorts before b and false otherwise.
func entryLess(a, b *entry) bool {
	if res, _ := compareWeights(a.elems, b.elems); res != 0 {
		return res == -1
	}
	if a.logical != noAnchor {
		return a.logical == firstAnchor
	}
	if b.logical != noAnchor {
		return b.logical == lastAnchor
	}
	return a.str < b.str
}

type sortedEntries []*entry

func (s sortedEntries) Len() int {
	return len(s)
}

func (s sortedEntries) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortedEntries) Less(i, j int) bool {
	return entryLess(s[i], s[j])
}

type ordering struct {
	id       string
	entryMap map[string]*entry
	ordered  []*entry
	handle   *trieHandle
}

// insert inserts e into both entryMap and ordered.
// Note that insert simply appends e to ordered.  To reattain a sorted
// order, o.sort() should be called.
func (o *ordering) insert(e *entry) {
	if e.logical == noAnchor {
		o.entryMap[e.str] = e
	} else {
		// Use key format as used in UCA rules.
		o.entryMap[fmt.Sprintf("[%s]", e.str)] = e
		// Also add index entry for XML format.
		o.entryMap[fmt.Sprintf("<%s/>", strings.Replace(e.str, " ", "_", -1))] = e
	}
	o.ordered = append(o.ordered, e)
}

// newEntry creates a new entry for the given info and inserts it into
// the index.
func (o *ordering) newEntry(s string, ces []rawCE) *entry {
	e := &entry{
		runes: []rune(s),
		elems: ces,
		str:   s,
	}
	o.insert(e)
	return e
}

// find looks up and returns the entry for the given string.
// It returns nil if str is not in the index and if an implicit value
// cannot be derived, that is, if str represents more than one rune.
func (o *ordering) find(str string) *en