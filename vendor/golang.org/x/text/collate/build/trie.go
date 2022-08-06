// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The trie in this file is used to associate the first full character
// in a UTF-8 string to a collation element.
// All but the last byte in a UTF-8 byte sequence are
// used to look up offsets in the index table to be used for the next byte.
// The last byte is used to index into a table of collation elements.
// This file contains the code for the generation of the trie.

package build

import (
	"fmt"
	"hash/fnv"
	"io"
	"reflect"
)

const (
	blockSize   = 64
	blockOffset = 2 // Subtract 2 blocks to compensate for the 0x80 added to continuation bytes.
)

type trieHandle struct {
	lookupStart uint16 // offset in table for first byte
	valueStart  uint16 // offset in table for first byte
}

type trie struct {
	index  []uint16
	values []uint32
}

// trieNode is the intermediate trie structure used for generating a trie.
type trieNode struct {
	index    []*trieNode
	value    []uint32
	b        byte
	refValue uint16
	refIndex uint16
}

func newNode() *trieNode {
	return &trieNode{
		index: make([]*trieNode, 64),
		value: make([]uint32, 12