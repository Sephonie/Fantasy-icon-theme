// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

// This file generates derivative tables based on the language package itself.

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/text/internal/gen"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/cldr"
)

var (
	test = flag.Bool("test", false,
		"test existing tables; can be used to compare web data with package data.")

	draft = flag.String("draft",
		"contributed",
		`Minimal draft requirements (approved, contributed, provisional, unconfirmed).`)
)

func main() {
	gen.Init()

	// Read the CLDR zip file.
	r := gen.OpenCLDRCoreZip()
	defer r.Close()

	d := &cldr.Decoder{}
	data, err := d.DecodeZip(r)
	if err != nil {
		log.Fatalf("DecodeZip: %v", err)
	}

	w := gen.NewCodeWriter()
	defer func() {
		buf := &bytes.Buffer{}

		if _, err = w.WriteGo(buf, "language", ""); err != nil {
			log.Fatalf("Error formatting file index.go: %v", err)
		}

		// Since we're generating a table for our own package we need to rewrite
		// doing the equivalent of go fmt -r 'language.b -> b'. Using
		// bytes.Replace will do.
		out := bytes.Replace(buf.Bytes(), []byte("language."), nil, -1)
		if err := ioutil.WriteFile("index.go", out, 0600); err != nil {
			log.Fatalf("Could not create file index.go: %v", err)
		}
	}()

	m := map[language.Tag]bool{}
	for _, lang := range data.Locales() {
		// We include all locales unconditionally to be consistent with en_US.
		// We want en_US, even though it has no data associated with it.

		// TODO: put any of the languages for which no data exists at the end
		// of the index. This allows all components based on ICU to use that
		// as the cutoff point.
		// if x := data.RawLDML(lang); false ||
		// 	x.LocaleDisplayNames != nil ||
		// 	x.Characters != nil ||
		// 	x.Delimiters != nil ||
		// 	x.Measurement != nil ||
		// 	x.Dates != nil ||
		// 	x.Numbers != nil ||
		// 	x.Units != nil ||
		// 	x.ListPatterns != nil ||
		// 	x.Collations != nil ||
		// 	x.Segmentations != nil ||
		// 	x.Rbnf != nil ||
		// 	x.Annotations != nil ||