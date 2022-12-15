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

var lastChar = rune