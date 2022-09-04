// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colltab

import (
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Table holds all collation data for a given collation ordering.
type Table struct {
	Index Trie // main trie

	// expansion info
	ExpandElem []uint32

	// contraction info
	ContractTries  ContractTrieSet
	ContractElem   []uint32
	MaxContractLen int
	VariableTop    uint32
}

func (t *Table) AppendNext(w []Elem, b []byte) (res []Elem, n int) {
	return t.appendNext(w, source{bytes: b})
}

func (t *Table) AppendNextString(w []Elem, s string) (res []Elem, n int) {
	return t.appendNext(w, source{str: s})
}

func (t *Table) Start(p int, b []byte) int {
	// TODO: implement
	panic("not implemented")
}

func (t *Table) StartString(p int, s string) int {
	// TODO: implement
	panic("not implemented")
}

func (t *Table) Domain() []string {
	// TODO: implement
	panic("not implemented")
}

func (t *Table) Top() uint32 {
	return t.VariableTop
}

type source struct {
	str   string
	bytes []byte
}

func (src *source) lookup(t *Table) (ce Elem, sz int) {
	if src.bytes == nil {
		return t.Index.lookupString(src.str)
	}
	return t.Index.lookup(src.bytes)
}

func (src *source) tail(sz int) {
	if src.bytes == nil {
		src.str = src.str[sz:]
	} else {
		src.bytes = src.bytes[sz:]
	}
}

func (src *source) nfd(buf []byte, end int) []byte {
	if src.bytes == nil {
		return norm.NFD.AppendString(buf[:0], src.str[:end])
	}
	return norm.NFD.Append(buf[:0], src.bytes[:end]...)
}

func (src *source) rune() (r rune, sz int) {
	if src.bytes == nil {
		return utf8.DecodeRuneInString(src.str)
	}
	return utf8.DecodeRune(src.bytes)
}

func (src *source) properties(f norm.Form) norm.Properties {
	if src.bytes == nil {
		return f.PropertiesString(src.str)
	}
	return f.Properties(src.bytes)
}

// appendNext appends the weights corresponding to the next rune or
// contraction in s.  If a contraction is matched to a discontinuous
// sequence of runes, the weights for the interstitial runes are
// appended as well.  It returns a new slice that includes the appended
// weights and the number of bytes consumed from s.
func (t *Table) appendNext(w []Elem, src source) (res []Elem, n int) {
	ce, sz := src.lookup(t)
	tp := ce.ctype()
	if tp == ceNormal {
		if ce == 0 {
			r, _ := src.rune()
			const (
				hangulSize  = 3
				firstHangul = 0xAC00
				lastHangul  = 0xD7A3
			)
			if r >= firstHangul && r <= lastHangul {
				// TODO: performance can be considerably improved here.
				n = sz
				var buf [16]byte // Used for decomposing Hangul.
				for b := src.nfd(buf[:0], hangulSize); len(b) > 0; b = b[sz:] {
					ce, sz = t.Index.lookup(b)
					w = append(w, ce)
				}
				return w, n
			}
			ce = makeImplicitCE(implicitPrimary(r))
		}
		w = append(w, ce)
	} else if tp == ceExpansionIndex {
		w = t.appendExpansion(w, ce)
	} else if tp == ceContractionIndex {
		n := 0
		src.tail(sz)
		if src.bytes == nil {
			w, n = t.matchContractionString(w, ce, src.str)
		} else {
			w, n = t.matchContraction(w, ce,