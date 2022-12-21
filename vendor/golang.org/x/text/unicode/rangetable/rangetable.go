// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rangetable provides utilities for creating and inspecting
// unicode.RangeTables.
package rangetable

import (
	"sort"
	"unicode"
)

// New creates a RangeTable from the given runes, which may contain duplicates.
func New(r ...rune) *unicode.RangeTable {
	if len(r) == 0 {
		return &unicode.RangeTable{}
	}

	sort.Sort(byRune(r))

	// Remove duplicates.
	k := 1
	for i := 1; i < len(r); i++ {
		if r[k-1] != r[i] {
			r[k] = r[i]
			k++
		}
	}

	var rt unicode.RangeTable
	for _, r := range r[:k] {
		if r <= 0xFFFF {
			rt.R16 = append(rt.R16, unicode.Range16{Lo: uint16(r), Hi: uint16(r), Stride: 1})
		} else {
			rt.R32 = append(rt.R32, unicode.Range32{Lo: uint32(r), Hi: uint32(r), Stride: 1})
		}
	}

	// Optimize RangeTable.
	r