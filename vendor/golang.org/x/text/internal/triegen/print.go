// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package triegen

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
)

// print writes all the data structures as well as the code necessary to use the
// trie to w.
func (b *builder) print(w io.Writer) error {
	b.Stats.NValueEntries = len(b.ValueBlocks) * blockSize
	b.Stats.NValueBytes = len(b.ValueBlocks) * blockSize * b.ValueSize
	b.Stats.NIndexEntries = len(b.IndexBlocks) * blockSize
	b.Stats.NIndexBytes = len(b.IndexBlocks) * blockSize * b.IndexSize
	b.Stats.NHandleBytes = len(b.Trie) * 2 * b.IndexSize

	// If we only have one root trie, all starter blocks are at position 0 and
	// we can access the arrays directly.
	if len(b.Trie) == 1 {
		// At this point we cannot refer to the generated tables directly.
		b.ASCIIBlock = b.Name + "Values"
		b.StarterBlock = b.Name + "Index"
	} else {
		// Otherwise we need to have explicit starter indexes in the trie
		// structure.
		b.ASCIIBlock = "t.ascii"
		b.StarterBlock = "t.utf8Start"
	}

	b.SourceType = "[]byte"
	if err := lookupGen.Execute(w, b); err != nil {
		return err
	}

	b.SourceType = "string"
	if err := lookupGen.Execute(w, b); err != nil {
		return err
	}

	if err := trieGen.Execute(w, b); err != nil {
		return err
	}

	for _, c := range b.Compactions {
		if err := c.c.Print(w); err != nil {
			return err
		}
	}

	return nil
}

func printValues(n int, values []uint64) string {
	w := &bytes.Buffer{}
	boff := n * blockSize
	fmt.Fprintf(w, "\t// Block %#x, offset %#x", n, boff)
	var newline bool
	for i, v := range values {
		if i%6 == 0 {
			newline = true
		}
		if v != 0 {
			if newline {
				fmt.Fprintf(w, "\n")
				newline = false
			}
			fmt.Fprintf(w, "\t%#02x:%#04x, ", boff+i, v)
		}
	}
	return w.String()
}

func printIndex(b *builder, nr int, n *node) string {
	w := &bytes.Buffer{}
	boff := nr * blockSize
	fmt.Fprintf(w, "\t// Block %#x, offset %#x", nr, boff)
	var newline bool
	for i, c := range n.children {
		if i%8 == 0 {
			newline = true
		}
		if c != nil {
			v := b.Compactions[c.index.compaction].Offset + uint32(c.index.index)
			if v != 0 {
				if newline {
					fmt.Fprintf(w, "\n")
					newline = false
				}
				fmt.Fprintf(w, "\t%#02x:%#02x, ", boff+i, v)
			}
		}
	}
	return w.String()
}

var (
	trieGen = template.Must(template.New("trie").Funcs(template.FuncMap{
		"printValues": printValues,
		"printIndex":  printIndex,
		"title":       strings.Title,
		"dec":         func(x int) int { return x - 1 },
		"psize": func(n int) string {
			return fmt.Sprintf("%d bytes (%.2f KiB)", n, float64(n)/1024)
		},
	}).Parse(trieTemplate))
	lookupGen = template.Must(template.New("lookup").Parse(lookupTemplate))
)

// TODO: consider the return type of lookup. It could be uint64, even if the
// internal value type is smaller. We will have to verify this with the
// performance of unicode/norm, which is very sensitive to such changes.
const trieTemplate = `{{$b := .}}{{$multi := gt (len .Trie) 1}}
// {{.Name}}Trie. Total size: {{psize .Size}}. Checksum: {{printf "%08x" .Checksum}}.
type {{.Name}}Trie struct { {{if $multi}}
	ascii []{{.ValueType}} // index for ASCII bytes
	utf8Start  []{{.IndexType}} // index for UTF-8 bytes >= 0xC0
{{end}}}

func new{{title .Name}}Trie(i int) *{{.Name}}Trie { {{if $multi}}
	h := {{.Name}}TrieHandles[i]
	return &{{.Name}}Trie{ {{.Name}}Values[uint32(h.ascii)<<6:], {{.N