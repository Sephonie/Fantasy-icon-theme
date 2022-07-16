// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build // import "golang.org/x/text/collate/build"

import (
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/internal/colltab"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

// TODO: optimizations:
// - expandElem is currently 20K. By putting unique colElems in a separate
//   table and having a byte array of indexes into this table, we can reduce
//   the total size to about 7K. By also factoring out the length bytes, we
//   can reduce this to about 6K.
// - trie valueBlocks are currently 100K. There are a lot of sparse blocks
//   and many consecutive values with the same stride. This can be further
//   compacted.
// - Compress secondary weights into 8 bits.
// - Some LDML specs specify a context element. Currently we simply concatenate
//   those.  Context can be implemented using the contraction trie. If Builder
//   could analyze and detect when using a context makes sense, there is no
//   need to expose this construct in the API.

// A Builder builds a root collation table.  The user must specify the
// collation elements for each entry.  A common use will be to base the weights
// on those specified in the allkeys* file as provided by the UCA or CLDR.
type Builder struct {
	index  *trieBuilder
	root   ordering
	locale []*Tailoring
	t      *table
	err    error
	built  bool

	minNonVar int // lowest primary recorded for a variable
	varTop    int // highest primary recorded for a non-variable

	// indexes used for reusing expansions and contractions
	expIndex map[string]int      // positions of expansions keyed by their string representation
	ctHandle map[string]ctHandle // contraction handles keyed by a concatenation of the suffixes
	ctElem   map[string]int      // contraction elements keyed by their string representation
}

// A Tailoring builds a collation table based on another collation table.
// The table is defined by specifying tailorings to the underlying table.
// See http://unicode.org/reports/tr35/ for an overview of tailoring
// collation tables.  The CLDR contains pre-defined tailorings for a variety