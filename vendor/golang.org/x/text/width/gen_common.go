// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

// This code is shared between the main code generator and the test code.

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"golang.org/x/text/internal/gen"
	"golang.org/x/text/internal/ucd"
)

var (
	outputFile = flag.String("out", "tables.go", "output file")
)

var typeMap = map[string]elem{
	"A":  tagAmbiguous,
	"N":  tagNeutral,
	"Na": tagNarrow,
	"W":  tagWide,
	"F":  tagFullwidth,
	"H":  tagHalfwidth,
}

// getWidthData calls f for every entry for which it is defined.
//
// f may be called multiple times for the same rune. The last call to f is the
// correct value. f is not called for all runes. 