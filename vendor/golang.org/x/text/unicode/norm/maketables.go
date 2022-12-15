// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Normalization table generator.
// Data read from the web.
// See forminfo.go for a description of the trie values associated with each rune.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/internal/gen"
	"golang.org/x/text/internal/triegen"
	"golang.org/x/text/internal/ucd"
)

func main() {
	gen.Init()
	loadUnicodeData()
	compactCCC()
	loadCompositionExclusions()
	completeCharFields(FCanonical)
	completeCharFields(FCompatibility)
	computeNonStarterCounts()
	verifyComputed()
	printChars()
	testDerived()
	printTestdata()
	makeTables()
}

var (
	tablelist = flag.String("tables",
		"all",
		"comma-separated list of which tables to generate; "+
			"can be 'decomp', 'recomp', 'info' and 'all'")
	test = flag.Bool("test",
		false,
		"test existing tables against DerivedNormalizationProps and generate test data for regression testing")
	verbose = flag.Bool("verbose",
		false,
		"write data to stdout as it is parsed")
)

const MaxChar = 0x10FFFF // anything above this shouldn't exist

// Quick Check properties of runes allow us to quickly
// determine whether a rune may occur in a normal form.
// For a given normal form, a rune may be guaranteed to occur
// verbatim (QC=Yes), may or may not combine with another
// rune (QC=Maybe), or may not occur (QC=No).
type QCResult int

const (
	QCUnknown QCResult = iota
	QCYes
	QCNo
	QCMaybe
)

func (r QCResult) String() string {
	switch r {
	case QCYes:
		return "Yes"
	case QCNo:
		return "No"
	case QCMaybe:
		return "Maybe"
	}
	return "***UNKNOWN***"
}

const (
	FCanonical     = iota // NFC or NFD
	FCompatibility        // NFKC or NFKD
	FNumberOfFormTypes
)

const (
	MComposed   = iota // NFC or NFKC
	MDecomposed        // NFD or NFKD
	MNumberOfModes
)

// This contains only the properties we're interested in.
type Char struct {
	name          string
	codePoint     rune  // if zero, this index is not a valid code point.
	ccc           uint8 // canonical combining class
	origCCC       uint8
	excludeInComp bool // from CompositionExclusions.txt
	compatDecomp  bool // it has a compatibility expansion

	nTrailingNonStarters uint8
	nLeadingNonStarters  uint8 // must be equal to trailing if non-zero

	forms [FNumberOfFormTypes]FormInfo // For FCanonical and FCompatibility

	state State
}

var chars = make([]Char, MaxChar+1)
var cccMap = make(map[uint8]uint8)

func (c Char) String() string {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "%U [%s]:\n", c.codePoint, c.name)
	fmt.Fprintf(buf, "  ccc: %v\n", c.ccc)
	fmt.Fprintf(buf, "  excludeInComp: %v\n", c.excludeInComp)
	fmt.Fprintf(buf, "  compatDecomp: %v\n", c.compatDecomp)
	fmt.Fprintf(buf, "  state: %v\n", c.state)
	fmt.Fprintf(buf, "  NFC:\n")
	fmt.Fprint(buf, c.forms[FCanonical])
	fmt.Fprintf(buf, "  NFKC:\n")
	fmt.Fprint(buf, c.forms[FCompatibility])

	return buf.String()
}

// In UnicodeData.txt, some ranges are marked like this:
//	3400;<CJK Ideograph Extension A, First>;Lo;0;L;;;;;N;;;;;
//	4DB5;<CJK Ideograph Extension A, Last>;Lo;0;L;;;;;N;;;;;
// parseCharacter keeps a state variable indicating the weirdness.
type State int

const (
	SNormal State = iota // known to be zero for the type
	SFirst
	SLast
	SMissing
)

var lastChar = rune('\u0000')

func (c Char) isValid() bool {
	return c.codePoint != 0 && c.state != SMissing
}

type FormInfo struct {
	quickCheck [MNumberOfModes]QCResult // index: MComposed or MDecomposed
	verified   [MNumberOfModes]bool     // index: MComposed or MDecomposed

	combinesForward  bool // May combine with rune on the right
	combinesBackward bool // May combine with rune on the left
	isOneWay         bool // Never appears in result
	inDecomp         bool // Some decompositions result in this char.
	decomp           Decomposition
	expandedDecomp   Decomposition
}

func (f FormInfo) String() string {
	buf := bytes.NewBuffer(make([]byte, 0))

	fmt.Fprintf(buf, "    quickCheck[C]: %v\n", f.quickCheck[MComposed])
	fmt.Fprintf(buf, "    quickCheck[D]: %v\n", f.quickCheck[MDecomposed])
	fmt.Fprintf(buf, "    cmbForward: %v\n", f.combinesForward)
	fmt.Fprintf(buf, "    cmbBackward: %v\n", f.combinesBackward)
	fmt.Fprintf(buf, "    isOneWay: %v\n", f.isOneWay)
	fmt.Fprintf(buf, "    inDecomp: %v\n", f.inDecomp)
	fmt.Fprintf(buf, "    decomposition: %X\n", f.decomp)
	fmt.Fprintf(buf, "    expandedDecomp: %X\n", f.expandedDecomp)

	return buf.String()
}

type Decomposition []rune

func parseDecomposition(s string, skipfirst bool) (a []rune, err error) {
	decomp := strings.Split(s, " ")
	if len(decomp) > 0 && skipfirst {
		decomp = decomp[1:]
	}
	for _, d := range decomp {
		point, err := strconv.ParseUint(d, 16, 64)
		if err != nil {
			return a, err
		}
		a = append(a, rune(point))
	}
	return a, nil
}

func loadUnicodeData() {
	f := gen.OpenUCDFile("UnicodeData.txt")
	defer f.Close()
	p := ucd.New(f)
	for p.Next() {
		r := p.Rune(ucd.CodePoint)
		char := &chars[r]

		char.ccc = uint8(p.Uint(ucd.CanonicalCombiningClass))
		decmap := p.String(ucd.DecompMapping)

		exp, err := parseDecomposition(decmap, false)
		isCompat := false
		if err != nil {
			if len(decmap) > 0 {
				exp, err = parseDecomposition(decmap, true)
				if err != nil {
					log.Fatalf(`%U: bad decomp |%v|: "%s"`, r, decmap, err)
				}
				isCompat = true
			}
		}

		char.name = p.String(ucd.Name)
		char.codePoint = r
		char.forms[FCompatibility].decomp = exp
		if !isCompat {
			char.forms[FCanonical].decomp = exp
		} else {
			char.compatDecomp = true
		}
		if len(decmap) > 0 {
			char.forms[FCompatibility].decomp = exp
		}
	}
	if err := p.Err(); err != nil {
		log.Fatal(err)
	}
}

// compactCCC converts the sparse set of CCC values to a continguous one,
// reducing the number of bits needed from 8 to 6.
func compactCCC() {
	m := make(map[uint8]uint8)
	for i := range chars {
		c := &chars[i]
		m[c.ccc] = 0
	}
	cccs := []int{}
	for v, _ := range m {
		cccs = append(cccs, int(v))
	}
	sort.Ints(cccs)
	for i, c := range cccs {
		cccMap[uint8(i)] = uint8(c)
		m[uint8(c)] = uint8(i)
	}
	for i := range chars {
		c := &chars[i]
		c.origCCC = c.ccc
		c.ccc = m[c.ccc]
	}
	if len(m) >= 1<<6 {
		log.Fatalf("too many difference CCC values: %d >= 64", len(m))
	}
}

// CompositionExclusions.txt has form:
// 0958    # ...
// See http://unicode.org/reports/tr44/ for full explanation
func loadCompositionExclusions() {
	f := gen.OpenUCDFile("CompositionExclusions.txt")
	defer f.Close()
	p := ucd.New(f)
	for p.Next() {
		c := &chars[p.Rune(0)]
		if c.excludeInComp {
			log.Fatalf("%U: Duplicate entry in exclusions.", c.codePoint)
		}
		c.excludeInComp = true
	}
	if e := p.Err(); e != nil {
		log.Fatal(e)
	}
}

// hasCompatDecomp returns true if any of the recursive
// decompositions contains a compatibility expansion.
// In this case, the character may not occur in 