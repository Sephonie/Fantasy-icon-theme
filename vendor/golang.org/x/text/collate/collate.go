// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO: remove hard-coded versions when we have implemented fractional weights.
// The current implementation is incompatible with later CLDR versions.
//go:generate go run maketables.go -cldr=23 -unicode=6.2.0

// Package collate contains types for comparing and sorting Unicode strings
// according to a given collation order.
package collate // import "golang.org/x/text/collate"

import (
	"bytes"
	"strings"

	"golang.org/x/text/internal/colltab"
	"golang.org/x/text/language"
)

// Collator provides functionality for comparing strings for a given
// collation order.
type Collator struct {
	options

	sorter sorter

	_iter [2]iter
}

func (c *Collator) iter(i int) *iter {
	// TODO: evaluate performance for making the second iterator optional.
	return &c._iter[i]
}

// Supported returns the list of languages for which collating differs from it