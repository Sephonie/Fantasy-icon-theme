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

	// Initialize result levels to paragraph embedding level.
	p.resultLevels = make([]level, p.Len())
	setLevels(p.resultLevels, p.embeddingLevel)

	// 2) Explicit levels and directions
	// Rules X1-X8.
	p.determineExplicitEmbeddingLevels()

	// Rule X9.
	// We do not remove the embeddings, the overrides, the PDFs, and the BNs
	// from the string explicitly. But they are not copied into isolating run
	// sequences when they are created, so they are removed for all
	// practical purposes.

	// Rule X10.
	// Run remainder of algorithm one isolating run sequence at a time
	for _, seq := range p.determineIsolatingRunSequences() {
		// 3) resolving weak types
		// Rules W1-W7.
		seq.resolveWeakTypes()

		// 4a) resolving paired brackets
		// Rule N0
		resolvePairedBrackets(seq)

		// 4b) resolving neutral types
		// Rules N1-N3.
		seq.resolveNeutralTypes()

		// 5) resolving implicit embedding levels
		// Rules I1, I2.
		seq.resolveImplicitLevels()

		// Apply the computed levels and types
		seq.applyLevelsAndTypes()
	}

	// Assign appropriate levels to 'hide' LREs, RLEs, LROs, RLOs, PDFs, and
	// BNs. This is for convenience, so the resulting level array will have
	// a value for every character.
	p.assignLevelsToCharactersRemovedByX9()
}

// determineMatchingIsolates determines the matching PDI for each isolate
// initiator and vice versa.
//
// Definition BD9.
//
// At the end of this function:
//
//  - The member variable matchingPDI is set to point to the index of the
//    matching PDI character for each isolate initiator character. If there is
//    no matching PDI, it is set to the length of the input text. For other
//    characters, it is set to -1.
//  - The member variable matchingIsolateInitiator is set to point to the
//    index of the matching isolate initiator character for each PDI character.
//    If there is no matching isolate initiator, or the character is not a PDI,
//    it is set to -1.
func (p *paragraph) determineMatchingIsolates() {
	p.matchingPDI = make([]int, p.Len())
	p.matchingIsolateInitiator = make([]int, p.Len())

	for i := range p.matchingIsolateInitiator {
		p.matchingIsolateInitiator[i] = -1
	}

	for i := range p.matchingPDI {
		p.matchingPDI[i] = -1

		if t := p.resultTypes[i]; t.in(LRI, RLI, FSI) {
			depthCounter := 1
			for j := i + 1; j < p.Len(); j++ {
				if u := p.resultTypes[j]; u.in(LRI, RLI, FSI) {
					depthCounter++
				} else if u == PDI {
					if depthCounter--; depthCounter == 0 {
						p.matchingPDI[i] = j
						p.matchingIsolateInitiator[j] = i
						break
					}
				}
			}
			if p.matchingPDI[i] == -1 {
				p.matchingPDI[i] = p.Len()
			}
		}
	}
}

// determineParagraphEmbeddingLevel reports the resolved paragraph direction of
// the substring limited by the given range [start, end).
//
// Determines the paragraph level based on rules P2, P3. This is also used
// in rule X5c to find if an FSI should resolve to LRI or RLI.
func (p *paragraph) determineParagraphEmbeddingLevel(start, end int) level {
	var strongType Class = unknownClass

	// Rule P2.
	for i := start; i < end; i++ {
		if t := p.resultTypes[i]; t.in(L, AL, R) {
			strongType = t
			break
		} else if t.in(FSI, LRI, RLI) {
			i = p.matchingPDI[i] // skip over to the matching PDI
			if i > end {
				log.Panic("assert (i <= end)")
			}
		}
	}
	// Rule P3.
	switch strongType {
	case unknownClass: // none found
		// default embedding level when no strong types found is 0.
		return 0
	case L:
		return 0
	default: // AL, R
		return 1
	}
}

const maxDepth = 125

// This stack will store the embedding levels and override and isolated
// statuses
type directionalStatusStack struct {
	stackCounter        int
	embeddingLevelStack [maxDepth + 1]level
	overrideStatusStack [maxDepth + 1]Class
	isolateStatusStack  [maxDepth + 1]bool
}

func (s *directionalStatusStack) empty()     { s.stackCounter = 0 }
func (s *directionalStatusStack) pop()       { s.stackCounter-- }
func (s *directionalStatusStack) depth() int { return s.stackCounter }

func (s *directionalStatusStack) push(level level, overrideStatus Class, isolateStatus bool) {
	s.embeddingLevelStack[s.stackCounter] = level
	s.overrideStatusStack[s.stackCounter] = overrideStatus
	s.isolateStatusStack[s.stackCounter] = isolateStatus
	s.stackCounter++
}

func (s *directionalStatusStack) lastEmbeddingLevel() level {
	return s.embeddingLevelStack[s.stackCounter-1]
}

func (s *directionalStatusStack) lastDirectionalOverrideStatus() Class {
	return s.overrideStatusStack[s.stackCounter-1]
}

func (s *directionalStatusStack) lastDirectionalIsolateStatus() bool {
	return s.isolateStatusStack[s.stackCounter-1]
}

// Determine explicit levels using rules X1 - X8
func (p *paragraph) determineExplicitEmbeddingLevels() {
	var stack directionalStatusStack
	var overflowIsolateCount, overflowEmbeddingCount, validIsolateCount int

	// Rule X1.
	stack.push(p.embeddingLevel, ON, false)

	for i, t := range p.resultTypes {
		// Rules X2, X3, X4, X5, X5a, X5b, X5c
		switch t {
		case RLE, LRE, RLO, LRO, RLI, LRI, FSI:
			isIsolate := t.in(RLI, LRI, FSI)
			isRTL := t.in(RLE, RLO, RLI)

			// override if this is an FSI that resolves to RLI
			if t == FSI {
				isRTL = (p.determineParagraphEmbeddingLevel(i+1, p.matchingPDI[i]) == 1)
			}
			if isIsolate {
				p.resultLevels[i] = stack.lastEmbeddingLevel()
				if stack.lastDirectionalOverrideStatus() != ON {
					p.resultTypes[i] = stack.lastDirectionalOverrideStatus()
				}
			}

			var newLevel level
			if isRTL {
				// least greater odd
				newLevel = (stack.lastEmbeddingLevel() + 1) | 1
			} else {
				// least greater even
				newLevel = (stack.lastEmbeddingLevel() + 2) &^ 1
			}

			if newLevel <= maxDepth && overflowIsolateCount == 0 && overflowEmbeddingCount == 0 {
				if isIsolate {
					validIsolateCount++
				}
				// Push new embedding level, override status, and isolated
				// status.
				// No check for valid stack counter, since the level check
				// suffices.
				switch t {
				case LRO:
					stack.push(newLevel, L, isIsolate)
				case RLO:
					stack.push(newLevel, R, isIsolate)
				default:
					stack.push(newLevel, ON, isIsolate)
				}
				// Not really part of the spec
				if !isIsolate {
					p.resultLevels[i] = newLevel
				}
			} else {
				// This is an invalid explicit formatting character,
				// so apply the "Otherwise" part of rules X2-X5b.
				if isIsolate {
					overflowIsolateCount++
				} else { // !isIsolate
					if overflowIsolateCount == 0 {
						overflowEmbeddingCount++
					}
				}
			}

		// Rule X6a
		case PDI:
			if overflowIsolateCount > 0 {
				overflowIsolateCount--
			} else if validIsolateCount == 0 {
				// do nothing
			} else {
				overflowEmbeddingCount = 0
				for !stack.lastDirectionalIsolateStatus() {
					stack.pop()
				}
				stack.pop()
				validIsolateCount--
			}
			p.resultLevels[i] = stack.lastEmbeddingLevel()

		// Rule X7
		case PDF:
			// Not really part of the spec
			p.resultLevels[i] = stack.lastEmbeddingLevel()

			if overflowIsolateCount > 0 {
				// do nothing
			} else if overflowEmbeddingCount > 0 {
				overflowEmbeddingCount--
			} else if !stack.lastDirectionalIsolateStatus() && stack.depth() >= 2 {
				stack.pop()
			}

		case B: // paragraph separator.
			// Rule X8.

			// These values are reset for clarity, in this implementation B
			// can only occur as the last code in the array.
			stack.empty()
			overflowIsolateCount = 0
			overflowEmbeddingCount = 0
			validIsolateCount = 0
			p.resultLevels[i] = p.embeddingLevel

		default:
			p.resultLevels[i] = stack.lastEmbeddingLevel()
			if stack.lastDirectionalOverrideStatus() != ON {
				p.resultTypes[i] = stack.lastDirectionalOverrideStatus()
			}
		}
	}
}

type isolatingRunSequence struct {
	p *paragraph

	indexes []int // indexes to the original string

	types          []Class // type of each character using the index
	resolvedLevels []level // resolved levels after application of rules
	level          level
	sos, eos       Class
}

func (i *isolatingRunSequence) Len() int { return len(i.indexes) }

func maxLevel(a, b level) level {
	if a > b {
		return a
	}
	return b
}

// Rule X10, second bullet: Determine the start-of-sequence (sos) and end-of-sequence (eos) types,
// 			 either L or R, for each isolating run sequence.
func (p *paragraph) isolatingRunSequence(indexes []int) *isolatingRunSequence {
	length := len(indexes)
	types := make([]Class, length)
	for i, x := range indexes {
		types[i] = p.resultTypes[x]
	}

	// assign level, sos and eos
	prevChar := indexes[0] - 1
	for prevChar >= 0 && isRemovedByX9(p.initialTypes[prevChar]) {
		prevChar--
	}
	prevLevel := p.embeddingLevel
	if prevChar >= 0 {
		prevLevel = p.resultLevels[prevChar]
	}

	var succLevel level
	lastType := types[length-1]
	if lastType.in(LRI, RLI, FSI) {
		succLevel = p.embeddingLevel
	} else {
		// the first character after the end of run sequence
		limit := indexes[length-1] + 1
		for ; limit < p.Len() && isRemovedByX9(p.initialTypes[limit]); limit++ {

		}
		succLevel = p.embeddingLevel
		if limit < p.Len() {
			succLevel = p.resultLevels[limit]
		}
	}
	level := p.resultLevels[indexes[0]]
	return &isolatingRunSequence{
		p:       p,
		indexes: indexes,
		types:   types,
		level:   level,
		sos:     typeForLevel(maxLevel(prevLevel, level)),
		eos:     typeForLevel(maxLevel(succLevel, level)),
	}
}

// Resolving weak types Rules W1-W7.
//
// Note that some weak types (EN, AN) remain after this processing is
// complete.
func (s *isolatingRunSequence) resolveWeakTypes() {

	// on entry, only these types remain
	s.assertOnly(L, R, AL, EN, ES, ET, AN, CS, B, S, WS, ON, NSM, LRI, RLI, FSI, PDI)

	// Rule W1.
	// Changes all NSMs.
	preceedingCharacterType := s.sos
	for i, t := range s.types {
		if t == NSM {
			s.types[i] = preceedingCharacterType
		} else {
			if t.in(LRI, RLI, FSI, PDI) {
				preceedingCharacterType = ON
			}
			preceedingCharacterType = t
		}
	}

	// Rule W2.
	// EN does not change at the start of the run, because sos != AL.
	for i, t := range s.types {
		if t == EN {
			for j := i - 1; j >= 0; j-- {
				if t := s.types[j]; t.in(L, R, AL) {
					if t == AL {
						s.types[i] = AN
					}
					break
				}
			}
		}
	}

	// Rule W3.
	for i, t := range s.types {
		if t == AL {
			s.types[i] = R
		}
	}

	// Rule W4.
	// Since there must be values on both sides for this rule to have an
	// effect, the scan skips the first and last value.
	//
	// Although the scan proceeds left to right, and changes the type
	// values in a way that would appear to affect the computations
	// later in the scan, there is actually no problem. A change in the
	// current value can only affect the value to its immediate right,
	// and only affect it if it is ES or CS. But the current value can
	// only change if the value to its right is not ES or CS. Thus
	// either the current value will not change, or its change will have
	// no effect on the remainder of the analysis.

	for i := 1; i < s.Len()-1; i++ {
		t := s.types[i]
		if t == ES || t == CS {
			prevSepType := s.types[i-1]
			succSepType := s.types[i+1]
			if prevSepType == EN && succSepType == EN {
				s.types[i] = EN
			} else if s.types[i] == CS && prevSepType == AN && succSepType == AN {
				s.types[i] = AN
			}
		}
	}

	// Rule W5.
	for i, t := range s.types {
		if t == ET {
			// locate end of sequence
			runStart := i
			runEnd := s.findRunLimit(runStart, ET)

			// check values at ends of sequence
			t := s.sos
			if runStart > 0 {
				t = s.types[runStart-1]
			}
			if t != EN {
				t = s.eos
				if runEnd < len(s.types) {
					t = s.types[runEnd]
				}
			}
			if t == EN {
				setTypes(s.types[runStart:runEnd], EN)
			}
			// continue at end of sequence
			i = runEnd
		}
	}

	// Rule W6.
	for i, t := range s.types {
		if t.in(ES, ET, CS) {
			s.types[i] = ON
		}
	}

	// Rule W7.
	for i, t := range s.types {
		if t == EN {
			// set default if we reach start of run
			prevStrongType := s.sos
			for j := i - 1; j >= 0; j-- {
				t = s.types[j]
				if t == L || t == R { // AL's have been changed to R
					prevStrongType = t
					break
				}
			}
			if prevStrongType == L {
				s.types[i] = L
			}
		}
	}
}

// 6) resolving neutral types Rules N1-N2.
func (s *isolatingRunSequence) resolveNeutralTypes() {

	// on entry, only these types can be in resultTypes
	s.assertOnly(L, R, EN, AN, B, S, WS, ON, RLI, LRI, FSI, PDI)

	for i, t := range s.types {
		switch t {
		case WS, ON, B, S, RLI, LRI, FSI, PDI:
			// find bounds of run of neutrals
			runStart := i
			runEnd := s.findRunLimit(runStart, B, S, WS, ON, RLI, LRI, FSI, PDI)

			// determine effective types at ends of run
			var leadType, trailType Class

			// Note that the character found can only be L, R, AN, or
			// EN.
			if runStart == 0 {
				leadType = s.sos
			} else {
				leadType = s.types[runStart-1]
				if leadType.in(AN, EN) {
					leadType = R
				}
			}
			if runEnd == len(s.types) {
				trailType = s.eos
			} else {
				trailType = s.types[runEnd]
				if trailType.in(AN, EN) {
					trailType = R
				}
			}

			var resolvedType Class
			if leadType == trailType {
				// Rule N1.
				resolvedType = leadType
			} else {
				// Rule N2.
				// Notice the embedding level of the run is used, not
				// the paragraph embedding level.
				resolvedType = typeForLevel(s.level)
			}

			setTypes(s.types[runStart:runEnd], resolvedType)

			// skip over run of (former) neutrals
			i = runEnd
		}
	}
}

func setLevels(levels []level, newLevel level) {
	for i := range levels {
		levels[i] = newLevel
	}
}

func setTypes(types []Class, newType Class) {
	for i := range types {
		types[i] = newType
	}
}

// 7) resolving implicit embedding levels Rules I1, I2.
func (s *isolatingRunSequence) resolveImplicitLevels() {

	// on entry, only these types can be in resultTypes
	s.assertOnly(L, R, EN, AN)

	s.resolvedLevels = make([]level, len(s.types))
	setLevels(s.resolvedLevels, s.level)

	if (s.level & 1) == 0 { // even level
		for i, t := range s.types {
			// Rule I1.
			if t == L {
				// no change
			} else if t == R {
				s.resolvedLevels[i] += 1
			} else { // t == AN || t == EN
				s.resolvedLevels[i] += 2
			}
		}
	} 