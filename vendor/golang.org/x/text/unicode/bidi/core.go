// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bidi

import "log"

// This implementation is a port based on the reference implementation found at:
// http://www.unicode.org/Public/PROGRAMS/BidiReferenceJava/
//
// described in Unicode Bidirectional Algorithm (UAX #9).
//
// Input:
// There are two levels of input to the algorithm, since clients may prefer to
// supply some information from out-of-band sources rather than relying on the
// default behavior.
//
// - Bidi class array
// - Bidi class array, with externally supplied base line direction
//
// Output:
// Output is separated into several stages:
//
//  - levels array over entire paragraph
//  - reordering array over entire paragraph
//  - levels array over line
//  - reordering array over line
//
// Note that for conformance to the Unicode Bidirectional Algorithm,
// implementations are only required to generate correct reordering and
// character directionality (odd or even levels) over a line. Generating
// identical level arrays over a line is not required. Bidi explicit format
// codes (LRE, RLE, LRO, RLO, PDF) and BN can be assigned arbitrary levels and
// positions as long as the rest of the input is properly reordered.
//
// As the algorithm is defined to operate on a single paragraph at a time, this
// implementation is written to handle single paragraphs. Thus rule P1 is
// presumed by this implementation-- the data provided to the implementation is
// assumed to be a single paragraph, and either contains no 'B' codes, or a
// single 'B' code at the end of the input. 'B' is allowed as input to
// illustrate how the algorithm assigns it a level.
//
// Also note that rules L3 and L4 depend on the rendering engine that uses the
// result of the bidi algorithm. This implementation assumes that the rendering
// engine expects combining marks in visual order (e.g. to the left of their
// base character in RTL runs) and that it adjusts the glyphs used to render
// mirrored characters that are in RTL runs so that they render appropriately.

// level is the embedding level of a character. Even embedding levels indicate
// left-to-right order and odd levels indicate right-to-left order. The special
// level of -1 is reserved for undefined order.
type level int8

const implicitLevel level = -1

// in returns if x is equal to any of the values in set.
func (c Class) in(set ...Class) bool {
	for _, s := range set {
		if c == s {
			return true
		}
	}
	return false
}

// A paragraph contains the state of a paragraph.
type paragraph struct {
	initialTypes []Class

	// Arrays of properties needed for paired bracket evaluation in N0
	pairTypes  []bracketType // paired Bracket types for paragraph
	pairValues []rune        // rune for opening bracket or pbOpen and pbClose; 0 for pbNone

	embeddingLevel level // default: = implicitLevel;

	// at the paragraph levels
	resultTypes  []Class
	resultLevels []level

	// Index of matching PDI for isolate initiator characters. For other
	// characters, the value of matchingPDI will be set to -1. For isolate
	// initiators with no matching PDI, matchingPDI will be set to the length of
	// the input string.
	matchingPDI []int

	// Index of matching isolate initiator for PDI characters. For other
	// characters, and for PDIs with no matching isolate initiator, the value of
	// matchingIsolateInitiator will be set to -1.
	matchingIsolateInitiator []int
}

// newParagraph initializes a paragraph. The user needs to supply a few arrays
// corresponding to the preprocessed text input. The types correspond to the
// Unicode BiDi classes for each rune. pairTypes indicates the bracket type for
// each rune. pairValues provides a unique bracket class identifier for each
// rune (suggested is the rune of the open bracket for opening and matching
// close brackets, after normalization). The embedding levels are optional, but
// may be supplied to encode embedding levels of styled text.
//
// TODO: return an error.
func newParagraph(types []Class, pairTypes []bracketType, pairValues []rune, levels level) *paragraph {
	validateTypes(types)
	validatePbTypes(pairTypes)
	validatePbValues(pairValues, pairTypes)
	validateParagraphEmbeddingLevel(levels)

	p := &paragraph{
		initialTypes:   append([]Class(nil), types...),
		embeddingLevel: levels,

		pairTypes:  pairTypes,
		pairValues: pairValues,

		resultTypes: append([]Class(nil), types...),
	}
	p.run()
	return p
}

func (p *paragraph) Len() int { return len(p.initialTypes) }

// The algorithm. Does not include line-based processing (Rules L1, L2).
// These are applied later in the line-based phase of the algorithm.
func (p *paragraph) run() {
	p.determineMatchingIsolates()

	// 1) determining the paragraph level
	// Rule P1 is the requirement for entering this algorithm.
	// Rules P2, P3.
	// If no externally supplied paragraph embedding level, use default.
	if p.embeddingLevel == implicitLevel {
		p.embeddingLevel = p.determineParagraphEmbeddingLevel(0, p.Len())
	}

	// Initialize result levels to paragr