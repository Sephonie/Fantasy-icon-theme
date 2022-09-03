// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colltab

import (
	"unicode"
	"unicode/utf8"
)

// NewNumericWeighter wraps w to replace individual digits to sort based on their
// numeric value.
//
// Weighter w must have a free primary weight after the primary weight for 9.
// If this is not the case, numeric value will sort at the same primary level
// as the first primary sorting after 9.
func NewNumericWeighter(w Weighter) Weighter {
	getElem := func(s string) Elem {
		elems, _ := w.AppendNextString(nil, s)
		return elems[0]
	}
	nine := getElem("9")

	// Numbers should order before zero, but the DUCET has no room for this.
	// TODO: move before zero once we use fractional collation elements.
	ns, _ := MakeElem(nine.Primary()+1, nine.Secondary(), int(nine.Tertiary()), 0)

	return &numericWeighter{
		Weighter: w,

		// We assume that w sorts digits of different kinds in order of numeric
		// value and that the tertiary weight order is preserved.
		//
		// TODO: evaluate whether it is worth basing the ranges on the Elem
		// encoding itself once the move to fractional weights is complete.
		zero:          getElem("0"),
		zeroSpecialLo: getElem("０"), // U+FF10 FULLWIDTH DIGIT ZERO
		zeroSpecialHi: getElem("₀"), // U+2080 SUBSCRIPT ZERO
		nine:          nine,
		nineSpecialHi: getElem("₉"), // U+2089 SUBSCRIPT NINE
		numberStart:   ns,
	}
}

// A numericWeighter translates a stream of digits into a stream of weights
// representing the numeric value.
type numericWeighter struct {
	Weighter

	// The Elems below all demarcate boundaries of specific ranges. With the
	// current element encoding digits are in two ranges: normal (default
	// tertiary value) and special. For most languages, digits have collation
	// elements in the normal range.
	//
	// Note: the range tests are very specific for the element encoding used by
	// this implementation. The tests in collate_test.go are designed to fail
	// if this code is not updated when an encoding has changed.

	zero          Elem // normal digit zero
	zeroSpecialLo Elem // special digit zero, low tertiary value
	zeroSpecialHi Elem // special digit zero, high tertiary value
	nine          Elem // normal digit nine
	nineSpecialHi Elem // special digit nine
	numberStart   Elem
}

// AppendNext calls the namesake of the underlying weigher, but replaces single
// digits with weights representing their value.
func (nw *numericWeighter) AppendNext(buf []Elem, s []byte) (ce []Elem, n int) {
	ce, n = nw.Weighter.AppendNext(buf, s)
	nc := numberConverter{
		elems: buf,
		w:     nw,
		b:     s,
	}
	isZero, ok := nc.checkNextDigit(ce)
	if !ok {
		return ce, n
	}
	// ce might have been grown already, so take it instead of buf.
	nc.init(ce, len(buf), isZero)
	for n < len(s) {
		ce, sz := nw.Weighter.AppendNext(nc.elems, s[n:])
		nc.b = s
		n += sz
		if !nc.update(ce) {
			break
		}
	}
	return nc.result(), n
}

// AppendNextString calls the namesake of the underlying weigher, but replaces
// single digits with weights representing the