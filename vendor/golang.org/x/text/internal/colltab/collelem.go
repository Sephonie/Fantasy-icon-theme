
// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colltab

import (
	"fmt"
	"unicode"
)

// Level identifies the collation comparison level.
// The primary level corresponds to the basic sorting of text.
// The secondary level corresponds to accents and related linguistic elements.
// The tertiary level corresponds to casing and related concepts.
// The quaternary level is derived from the other levels by the
// various algorithms for handling variable elements.
type Level int

const (
	Primary Level = iota
	Secondary
	Tertiary
	Quaternary
	Identity

	NumLevels
)

const (
	defaultSecondary = 0x20
	defaultTertiary  = 0x2
	maxTertiary      = 0x1F
	MaxQuaternary    = 0x1FFFFF // 21 bits.
)

// Elem is a representation of a collation element. This API provides ways to encode
// and decode Elems. Implementations of collation tables may use values greater
// or equal to PrivateUse for their own purposes.  However, these should never be
// returned by AppendNext.
type Elem uint32

const (
	maxCE       Elem = 0xAFFFFFFF
	PrivateUse       = minContract
	minContract      = 0xC0000000
	maxContract      = 0xDFFFFFFF
	minExpand        = 0xE0000000
	maxExpand        = 0xEFFFFFFF
	minDecomp        = 0xF0000000
)

type ceType int

const (
	ceNormal           ceType = iota // ceNormal includes implicits (ce == 0)
	ceContractionIndex               // rune can be a start of a contraction
	ceExpansionIndex                 // rune expands into a sequence of collation elements
	ceDecompose                      // rune expands using NFKC decomposition
)

func (ce Elem) ctype() ceType {
	if ce <= maxCE {
		return ceNormal
	}
	if ce <= maxContract {
		return ceContractionIndex
	} else {
		if ce <= maxExpand {
			return ceExpansionIndex
		}
		return ceDecompose
	}
	panic("should not reach here")
	return ceType(-1)
}

// For normal collation elements, we assume that a collation element either has
// a primary or non-default secondary value, not both.
// Collation elements with a primary value are of the form
// 01pppppp pppppppp ppppppp0 ssssssss
//   - p* is primary collation value
//   - s* is the secondary collation value
// 00pppppp pppppppp ppppppps sssttttt, where
//   - p* is primary collation value
//   - s* offset of secondary from default value.
//   - t* is the tertiary collation value
// 100ttttt cccccccc pppppppp pppppppp
//   - t* is the tertiar collation value
//   - c* is the canonical combining class
//   - p* is the primary collation value
// Collation elements with a secondary value are of the form
// 1010cccc ccccssss ssssssss tttttttt, where
//   - c* is the canonical combining class
//   - s* is the secondary collation value
//   - t* is the tertiary collation value
// 11qqqqqq qqqqqqqq qqqqqqq0 00000000
//   - q* quaternary value
const (
	ceTypeMask              = 0xC0000000
	ceTypeMaskExt           = 0xE0000000
	ceIgnoreMask            = 0xF00FFFFF
	ceType1                 = 0x40000000
	ceType2                 = 0x00000000
	ceType3or4              = 0x80000000
	ceType4                 = 0xA0000000
	ceTypeQ                 = 0xC0000000
	Ignore                  = ceType4
	firstNonPrimary         = 0x80000000
	lastSpecialPrimary      = 0xA0000000
	secondaryMask           = 0x80000000
	hasTertiaryMask         = 0x40000000
	primaryValueMask        = 0x3FFFFE00
	maxPrimaryBits          = 21
	compactPrimaryBits      = 16
	maxSecondaryBits        = 12
	maxTertiaryBits         = 8
	maxCCCBits              = 8
	maxSecondaryCompactBits = 8
	maxSecondaryDiffBits    = 4
	maxTertiaryCompactBits  = 5
	primaryShift            = 9
	compactSecondaryShift   = 5
	minCompactSecondary     = defaultSecondary - 4
)

func makeImplicitCE(primary int) Elem {
	return ceType1 | Elem(primary<<primaryShift) | defaultSecondary
}

// MakeElem returns an Elem for the given values.  It will return an error
// if the given combination of values is invalid.
func MakeElem(primary, secondary, tertiary int, ccc uint8) (Elem, error) {
	if w := primary; w >= 1<<maxPrimaryBits || w < 0 {
		return 0, fmt.Errorf("makeCE: primary weight out of bounds: %x >= %x", w, 1<<maxPrimaryBits)
	}
	if w := secondary; w >= 1<<maxSecondaryBits || w < 0 {
		return 0, fmt.Errorf("makeCE: secondary weight out of bounds: %x >= %x", w, 1<<maxSecondaryBits)
	}
	if w := tertiary; w >= 1<<maxTertiaryBits || w < 0 {
		return 0, fmt.Errorf("makeCE: tertiary weight out of bounds: %x >= %x", w, 1<<maxTertiaryBits)
	}
	ce := Elem(0)
	if primary != 0 {
		if ccc != 0 {
			if primary >= 1<<compactPrimaryBits {
				return 0, fmt.Errorf("makeCE: primary weight with non-zero CCC out of bounds: %x >= %x", primary, 1<<compactPrimaryBits)
			}
			if secondary != defaultSecondary {
				return 0, fmt.Errorf("makeCE: cannot combine non-default secondary value (%x) with non-zero CCC (%x)", secondary, ccc)
			}
			ce = Elem(tertiary << (compactPrimaryBits + maxCCCBits))
			ce |= Elem(ccc) << compactPrimaryBits
			ce |= Elem(primary)
			ce |= ceType3or4
		} else if tertiary == defaultTertiary {
			if secondary >= 1<<maxSecondaryCompactBits {
				return 0, fmt.Errorf("makeCE: secondary weight with non-zero primary out of bounds: %x >= %x", secondary, 1<<maxSecondaryCompactBits)
			}
			ce = Elem(primary<<(maxSecondaryCompactBits+1) + secondary)
			ce |= ceType1
		} else {
			d := secondary - defaultSecondary + maxSecondaryDiffBits
			if d >= 1<<maxSecondaryDiffBits || d < 0 {
				return 0, fmt.Errorf("makeCE: secondary weight diff out of bounds: %x < 0 || %x > %x", d, d, 1<<maxSecondaryDiffBits)
			}
			if tertiary >= 1<<maxTertiaryCompactBits {
				return 0, fmt.Errorf("makeCE: tertiary weight with non-zero primary out of bounds: %x > %x", tertiary, 1<<maxTertiaryCompactBits)
			}
			ce = Elem(primary<<maxSecondaryDiffBits + d)
			ce = ce<<maxTertiaryCompactBits + Elem(tertiary)
		}
	} else {
		ce = Elem(secondary<<maxTertiaryBits + tertiary)
		ce += Elem(ccc) << (maxSecondaryBits + maxTertiaryBits)
		ce |= ceType4
	}
	return ce, nil
}

// MakeQuaternary returns an Elem with the given quaternary value.
func MakeQuaternary(v int) Elem {
	return ceTypeQ | Elem(v<<primaryShift)
}

// Mask sets weights for any level smaller than l to 0.
// The resulting Elem can be used to test for equality with
// other Elems to which the same mask has been applied.
func (ce Elem) Mask(l Level) uint32 {
	return 0
}

// CCC returns the canonical combining class associated with the underlying character,
// if applicable, or 0 otherwise.
func (ce Elem) CCC() uint8 {
	if ce&ceType3or4 != 0 {
		if ce&ceType4 == ceType3or4 {
			return uint8(ce >> 16)
		}
		return uint8(ce >> 20)
	}
	return 0
}

// Primary returns the primary collation weight for ce.
func (ce Elem) Primary() int {
	if ce >= firstNonPrimary {
		if ce > lastSpecialPrimary {
			return 0
		}
		return int(uint16(ce))
	}
	return int(ce&primaryValueMask) >> primaryShift
}