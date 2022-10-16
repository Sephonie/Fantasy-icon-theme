
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bidirule implements the Bidi Rule defined by RFC 5893.
//
// This package is under development. The API may change without notice and
// without preserving backward compatibility.
package bidirule

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/bidi"
)

// This file contains an implementation of RFC 5893: Right-to-Left Scripts for
// Internationalized Domain Names for Applications (IDNA)
//