// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run makexml.go -output xml.go

// Package cldr provides a parser for LDML and related XML formats.
// This package is intended to be used by the table generation tools
// for the various internationalization-related packages.
// As the XML types are generated from the CLDR DTD, and as the CLDR standard
// is periodically amended, this package may change considerably over time.
// This mostly means that data may appear and disappear between versions.
// That is, old code should keep compiling for newer versions, but data
// may have moved or changed.
// CLDR version 22 is the first version supported by this package.
// Older versions may not work.
package cldr // import "golang.org/x/text/unicode/cldr"

import (
	"fmt"
	"sort"
)

// CLDR provides access to parsed data of the Unicode Common Locale Data Repository.
type CLDR struct {
	parent   map[string][]string
	locale   map[string]*LDML
	resolved map[string]*LDML
	bcp47    *LDMLBCP47
	supp     *SupplementalData
}

func makeCLDR() *CLDR {
	return &CLDR{
		parent:   make(map[string][]string),
		locale:   make(map[string]*LDML),
		resolved: make(map[string]*LDML),
		bcp47:    &LDMLBCP47{},
		supp:     &SupplementalData{},
	}
}

// BCP47 returns the parsed BCP47 LDML data. If no such data was parsed, nil is returned.
func (cldr *CLDR) BCP47() *LDMLBCP47 {
	return nil
}

// Draft indicates the draft level of an element.
type Draft int

const (
	Approved Draft = iota
	Contributed
	Provisional
	Unconfirmed
)

var drafts = []string{"unconfirmed", "provisional", "contributed", "approved", ""}

// ParseDraft returns the Draft value corresponding to the given string. The
// empty string corresponds to Approved.
func ParseDraft(level string) (Draft, error) {
	if level == "" {
		return Approved, nil
	}
	for i, s := range drafts {
		if level == s {
			return Unconfirmed - Draft(i), nil
		}
	}
	return Approved, fmt.Errorf("cldr: unknown draft level %q", level)
}

func (d Draft) String() string {
	return drafts[len(drafts)-1-int(d)]
}

// SetDraftLevel sets which draft levels to include in 