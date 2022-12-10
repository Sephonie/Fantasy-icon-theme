// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cldr

// This file implements the various inheritance constructs defined by LDML.
// See http://www.unicode.org/reports/tr35/#Inheritance_and_Validity
// for more details.

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// fieldIter iterates over fields in a struct. It includes
// fields of embedded structs.
type fieldIter struct {
	v        reflect.Value
	index, n []int
}

func iter(v reflect.Value) fieldIter {
	if v.Kind() != reflect.Struct {
		log.Panic