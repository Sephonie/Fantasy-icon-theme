// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

import (
	"fmt"
	"sort"
)

// The Coverage interface is used to define the level of coverage of an
// internationalization service. Note that not all types are supported by all
// services. As lists may be generated on the fly, it is recommended that users
// of a Coverage cache the results.
type Coverage interface {
	// Tags returns the list of supported tags.
	Tags() []Tag

	// BaseLanguages returns the list of supported base languages.
	BaseLanguages() []Base

	// Scripts returns the list of supported scripts.
	Scripts() []Script

	// Regions returns the list of supported regions.
	Regions() []Region
}

var (
	// Supported defines a Coverage that lists all supported subtags. Tags
	// always returns nil.
	Supported Coverage = allSubtags{}
)

// TODO:
// - Support Variants, numbering systems.
// - CLDR coverage levels.
// - Set of common tags defined in this package.

type allSubtags struct{}

// Regions returns the list of supported regions. As all regions are in a
// consecutive range, it simply returns a slice of numbers in increasing order.
// The "undefined" region is not returned.
func (s allSubtags) Regions() []Region {
	reg := make([]Region, numRegions)
	for i := range reg {
		reg[i] = Region{regionID(i + 1)}
	}
	return reg
}

// Scripts returns the list of supported scripts. As all scripts are in a
// consecutive range, it simply returns a slice of numbers in increasing order.
// The "undefined" script is not returned.
func (s allSubtags) Scripts() []Script {
	scr := make([]Script, numScripts)
	for i := range scr {
		scr[i] = Script{scriptID(i + 1)}
	}
	return scr
}

// BaseLanguages returns the list of all supported base languages. It generates
// the list by traversing the internal structures.
func (s allSubtags) BaseLanguages() []Base {
	base := make([]Base, 0, numLanguages)
	for i := 0; i < langNoIndexOffset; i++ {
		// We include