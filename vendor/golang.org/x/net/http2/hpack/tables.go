// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hpack

import (
	"fmt"
)

// headerFieldTable implements a list of HeaderFields.
// This is used to implement the static and dynamic tables.
type headerFieldTable struct {
	// For static tables, entries are never evicted.
	//
	// For dynamic tables, entries are evicted from ents[0] and added to the end.
	// Each entry has a unique id that starts at one and increments for each
	// entry that is added. This unique id is stable across evictions, meaning
	// it can be used as a pointer to a specific entry. As in hpack, unique ids
	// are 1-based. The unique id for ents[k] is k + evictCount + 1.
	//
	// Zero is not a valid unique id.
	//
	// evictCount should not overflow in any remotely practical situation. In
	// practice, we will have one dynamic table per HTTP/2 connection. If we
	// assume a very powerful server that handles 1M QPS per connection and each
	// request adds (then evicts) 100 entries from the table, it would still take
	// 2M years for evictCount to overflow.
	ents       []HeaderField
	evictCount uint64

	// byName maps a HeaderField name to the unique id of the newest entry with
	// the same name. See above for a definition of "unique id".
	byName map[string]uint64

	// byNameValue maps a HeaderField name/value pair to the unique id of the newest
	// entry with the same name and value. See above for a definition of "unique id".
	byNameValue map[pairNameValue]uint64
}

type pairNameValue struct {
	name, value string
}

func (t *headerFieldTable) init() {
	t.byName = make(map[string]uint64)
	t.byNameValue = make(map[pairNameValue]uint64)
}

// len reports the number of entries in the table.
func (t *headerFieldTable) len() int {
	return len(t.ents)
}

// addEntry adds a new entry.
func (t *headerFieldTable) addEntry(f HeaderField) {
	id := uint64(t.len()) + t.evictCount + 1
	t.byName[f.Name] = id
	t.byNameValue[pairNameValue{f.Name, f.Value}] = id
	t.ents = append(t.ents, f)
}

// evictOldest evicts the n oldest entries in the table.
func (t *headerFieldTable) evictOldest(n int) {
	if n > t.len() {
		panic(fmt.Sprintf("evictOldest(%v) on table with %v entries", n, t.len()))
	}
	for k := 0; k < n; k++ {
		f := t.ents[k]
		id := t.evictCount + uint64(k) + 1
		if t.byName[f.Name] == id {
			delete(t.byName, f.Name)
		}
		if p := (pairNameValue{f.Name, f.Value}); t.byNameValue[p] == id {
			delete(t.byNameValue, p)
		}
	}
	copy(t.ents, t.ents[n:])
	for k := t.len() - n; k < t.len(); k++ {
		t.ents[k] = HeaderField{} // so strings can be garbage collected
	}
	t.ents = t.ents[:t.len()-n]
	if t.evictCount+uint64(n) < t.evictCount {
		panic("evictCount overflow")
	}
	t.evictCount += uint64(n)
}

// search finds f in the table. If there is no match, i is 0.
// If both name and value match, i is the matched index and nameValueMatch
// becomes true. If only name matches, i points to that index and
// nameValueMatch becomes false.
//
// The returned index is a 1-based HPACK index. For dynamic tables, HPACK says
// that index 1 should be the newest entry, but t.ents[0] is the oldest entry,
// meaning t.ents is reversed for dynamic tables. Hence, when t is a dynamic
// table, the return value i actually refers to the entry t.ents[t.len()-i].
//
// All tables are assumed to be a dynamic tables except for the global
// staticTable pointer.
//
// See Section 2.3.3.
func (t *headerFieldTable) search(f HeaderField) (i uint64, nameValueMatch bool) {
	if !f.Sensitive {
		if id := t.byNameValue[pairNameValue{f.Name, f.Value}]; id != 0 {
			return t.idToIndex(id), true
		}
	}
	if id := t.byName[f.Name]; id != 0 {
		return t.idToIndex(id), false
	}
	return 0, false
}

// idToIndex converts a unique id to an HPACK index.
// See Section 2.3.3.
func (t *headerFieldTable) idToIndex(id uint64) uint64 {
	if id <= t.evictCount {
		panic(fmt.Sprintf("id (%v) <= evictCount (%v)", id, t.evictCount))
	}
	k := id - t.evictCount - 1 // convert id to an index t.ents[k]
	if t != staticTable {
		return uint64(t.len()) - k // dynamic table
	}
	return k + 1
}

// http://tools.ietf.org/html/draft-ietf-httpbis-header-compression-07#appendix-B
var staticTable = newStaticTable()
var staticTableEntries = [...]HeaderField{
	{Name: ":authority"},
	{Name: ":method", Value: "GET"},
	{Name: ":method", Value: "POST"},
	{Name: ":path", Value: "/"},
	{Name: ":path", Value: "/index.html"},
	{Name: ":scheme", Value: "http"},
	{Name: ":scheme", Value: "https"},
	{Name: ":status", Value: "200"},
	{Name: ":status", Value: "204"},
	{Name: ":status", Value: "206"},
	{Name: ":status", Value: "304"},
	{Name: ":status", Value: "400"},
	{Name: ":status", Value: "404"},
	{Name: ":status", Value: "500"},
	{Name: "accept-charset"},
	{Name: "accept-encoding", Value: "gzip, deflate"},
	{Name: "accept-language"},
	{Name: "accept-ranges"},
	{Name: "accept"},
	{Name: "access-control-allow-origin"},
	{Name: "age"},
	{Name: "allow"},
	{Name: "authorization"},
	{Name: "cache-control"},
	{Name: "content-disposition"},
	{Name: "content-encoding"},
	{Name: "content-language"},
	{Name: "content-length"},
	{Name: "content-location"},
	{Name: "content-range"},
	{Name: "content-type"},
	{Name: "cookie