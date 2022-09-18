
// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Language tag table generator.
// Data read from the web.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/internal/gen"
	"golang.org/x/text/internal/tag"
	"golang.org/x/text/unicode/cldr"
)

var (
	test = flag.Bool("test",
		false,
		"test existing tables; can be used to compare web data with package data.")
	outputFile = flag.String("output",
		"tables.go",
		"output file for generated tables")
)

var comment = []string{
	`
lang holds an alphabetically sorted list of ISO-639 language identifiers.
All entries are 4 bytes. The index of the identifier (divided by 4) is the language tag.
For 2-byte language identifiers, the two successive bytes have the following meaning:
    - if the first letter of the 2- and 3-letter ISO codes are the same:
      the second and third letter of the 3-letter ISO code.
    - otherwise: a 0 and a by 2 bits right-shifted index into altLangISO3.
For 3-byte language identifiers the 4th byte is 0.`,
	`
langNoIndex is a bit vector of all 3-letter language codes that are not used as an index
in lookup tables. The language ids for these language codes are derived directly
from the letters and are not consecutive.`,
	`
altLangISO3 holds an alphabetically sorted list of 3-letter language code alternatives
to 2-letter language codes that cannot be derived using the method described above.
Each 3-letter code is followed by its 1-byte langID.`,
	`
altLangIndex is used to convert indexes in altLangISO3 to langIDs.`,
	`
langAliasMap maps langIDs to their suggested replacements.`,
	`
script is an alphabetically sorted list of ISO 15924 codes. The index
of the script in the string, divided by 4, is the internal scriptID.`,
	`
isoRegionOffset needs to be added to the index of regionISO to obtain the regionID
for 2-letter ISO codes. (The first isoRegionOffset regionIDs are reserved for
the UN.M49 codes used for groups.)`,
	`
regionISO holds a list of alphabetically sorted 2-letter ISO region codes.
Each 2-letter codes is followed by two bytes with the following meaning:
    - [A-Z}{2}: the first letter of the 2-letter code plus these two 
                letters form the 3-letter ISO code.
    - 0, n:     index into altRegionISO3.`,
	`
regionTypes defines the status of a region for various standards.`,
	`
m49 maps regionIDs to UN.M49 codes. The first isoRegionOffset entries are
codes indicating collections of regions.`,
	`
m49Index gives indexes into fromM49 based on the three most significant bits
of a 10-bit UN.M49 code. To search an UN.M49 code in fromM49, search in
   fromM49[m49Index[msb39(code)]:m49Index[msb3(code)+1]]
for an entry where the first 7 bits match the 7 lsb of the UN.M49 code.
The region code is stored in the 9 lsb of the indexed value.`,
	`
fromM49 contains entries to map UN.M49 codes to regions. See m49Index for details.`,
	`
altRegionISO3 holds a list of 3-letter region codes that cannot be
mapped to 2-letter codes using the default algorithm. This is a short list.`,
	`
altRegionIDs holds a list of regionIDs the positions of which match those
of the 3-letter ISO codes in altRegionISO3.`,
	`
variantNumSpecialized is the number of specialized variants in variants.`,
	`
suppressScript is an index from langID to the dominant script for that language,
if it exists.  If a script is given, it should be suppressed from the language tag.`,
	`
likelyLang is a lookup table, indexed by langID, for the most likely
scripts and regions given incomplete information. If more entries exist for a
given language, region and script are the index and size respectively
of the list in likelyLangList.`,
	`
likelyLangList holds lists info associated with likelyLang.`,
	`
likelyRegion is a lookup table, indexed by regionID, for the most likely
languages and scripts given incomplete information. If more entries exist
for a given regionID, lang and script are the index and size respectively
of the list in likelyRegionList.
TODO: exclude containers and user-definable regions from the list.`,
	`
likelyRegionList holds lists info associated with likelyRegion.`,
	`
likelyScript is a lookup table, indexed by scriptID, for the most likely
languages and regions given a script.`,
	`
matchLang holds pairs of langIDs of base languages that are typically
mutually intelligible. Each pair is associated with a confidence and
whether the intelligibility goes one or both ways.`,
	`
matchScript holds pairs of scriptIDs where readers of one script
can typically also read the other. Each is associated with a confidence.`,
	`
nRegionGroups is the number of region groups.`,
	`
regionInclusion maps region identifiers to sets of regions in regionInclusionBits,
where each set holds all groupings that are directly connected in a region
containment graph.`,
	`
regionInclusionBits is an array of bit vectors where every vector represents
a set of region groupings.  These sets are used to compute the distance
between two regions for the purpose of language matching.`,
	`
regionInclusionNext marks, for each entry in regionInclusionBits, the set of
all groups that are reachable from the groups set in the respective entry.`,
}

// TODO: consider changing some of these structures to tries. This can reduce
// memory, but may increase the need for memory allocations. This could be
// mitigated if we can piggyback on language tags for common cases.

func failOnError(e error) {
	if e != nil {
		log.Panic(e)
	}
}

type setType int

const (
	Indexed setType = 1 + iota // all elements must be of same size
	Linear
)

type stringSet struct {
	s              []string
	sorted, frozen bool

	// We often need to update values after the creation of an index is completed.
	// We include a convenience map for keeping track of this.
	update map[string]string
	typ    setType // used for checking.
}

func (ss *stringSet) clone() stringSet {
	c := *ss
	c.s = append([]string(nil), c.s...)
	return c
}

func (ss *stringSet) setType(t setType) {
	if ss.typ != t && ss.typ != 0 {
		log.Panicf("type %d cannot be assigned as it was already %d", t, ss.typ)
	}
}

// parse parses a whitespace-separated string and initializes ss with its
// components.
func (ss *stringSet) parse(s string) {
	scan := bufio.NewScanner(strings.NewReader(s))
	scan.Split(bufio.ScanWords)
	for scan.Scan() {
		ss.add(scan.Text())
	}
}

func (ss *stringSet) assertChangeable() {
	if ss.frozen {
		log.Panic("attempt to modify a frozen stringSet")
	}
}

func (ss *stringSet) add(s string) {
	ss.assertChangeable()
	ss.s = append(ss.s, s)
	ss.sorted = ss.frozen
}

func (ss *stringSet) freeze() {
	ss.compact()
	ss.frozen = true
}

func (ss *stringSet) compact() {
	if ss.sorted {
		return
	}
	a := ss.s
	sort.Strings(a)
	k := 0
	for i := 1; i < len(a); i++ {
		if a[k] != a[i] {
			a[k+1] = a[i]
			k++
		}
	}
	ss.s = a[:k+1]
	ss.sorted = ss.frozen
}

type funcSorter struct {
	fn func(a, b string) bool
	sort.StringSlice
}

func (s funcSorter) Less(i, j int) bool {
	return s.fn(s.StringSlice[i], s.StringSlice[j])
}

func (ss *stringSet) sortFunc(f func(a, b string) bool) {
	ss.compact()
	sort.Sort(funcSorter{f, sort.StringSlice(ss.s)})
}

func (ss *stringSet) remove(s string) {
	ss.assertChangeable()
	if i, ok := ss.find(s); ok {
		copy(ss.s[i:], ss.s[i+1:])
		ss.s = ss.s[:len(ss.s)-1]
	}
}

func (ss *stringSet) replace(ol, nu string) {
	ss.s[ss.index(ol)] = nu
	ss.sorted = ss.frozen
}

func (ss *stringSet) index(s string) int {
	ss.setType(Indexed)
	i, ok := ss.find(s)
	if !ok {
		if i < len(ss.s) {
			log.Panicf("find: item %q is not in list. Closest match is %q.", s, ss.s[i])
		}
		log.Panicf("find: item %q is not in list", s)

	}
	return i
}

func (ss *stringSet) find(s string) (int, bool) {
	ss.compact()
	i := sort.SearchStrings(ss.s, s)
	return i, i != len(ss.s) && ss.s[i] == s
}

func (ss *stringSet) slice() []string {
	ss.compact()
	return ss.s
}

func (ss *stringSet) updateLater(v, key string) {
	if ss.update == nil {
		ss.update = map[string]string{}
	}
	ss.update[v] = key
}

// join joins the string and ensures that all entries are of the same length.
func (ss *stringSet) join() string {
	ss.setType(Indexed)
	n := len(ss.s[0])
	for _, s := range ss.s {
		if len(s) != n {
			log.Panicf("join: not all entries are of the same length: %q", s)
		}
	}
	ss.s = append(ss.s, strings.Repeat("\xff", n))
	return strings.Join(ss.s, "")
}

// ianaEntry holds information for an entry in the IANA Language Subtag Repository.
// All types use the same entry.
// See http://tools.ietf.org/html/bcp47#section-5.1 for a description of the various
// fields.
type ianaEntry struct {
	typ            string
	description    []string
	scope          string
	added          string
	preferred      string
	deprecated     string
	suppressScript string
	macro          string
	prefix         []string
}

type builder struct {
	w    *gen.CodeWriter
	hw   io.Writer // MultiWriter for w and w.Hash
	data *cldr.CLDR
	supp *cldr.SupplementalData

	// indices
	locale      stringSet // common locales
	lang        stringSet // canonical language ids (2 or 3 letter ISO codes) with data
	langNoIndex stringSet // 3-letter ISO codes with no associated data
	script      stringSet // 4-letter ISO codes
	region      stringSet // 2-letter ISO or 3-digit UN M49 codes
	variant     stringSet // 4-8-alphanumeric variant code.

	// Region codes that are groups with their corresponding group IDs.
	groups map[int]index

	// langInfo
	registry map[string]*ianaEntry
}

type index uint

func newBuilder(w *gen.CodeWriter) *builder {
	r := gen.OpenCLDRCoreZip()
	defer r.Close()
	d := &cldr.Decoder{}
	data, err := d.DecodeZip(r)
	failOnError(err)
	b := builder{
		w:    w,
		hw:   io.MultiWriter(w, w.Hash),
		data: data,
		supp: data.Supplemental(),
	}
	b.parseRegistry()
	return &b
}

func (b *builder) parseRegistry() {
	r := gen.OpenIANAFile("assignments/language-subtag-registry")
	defer r.Close()
	b.registry = make(map[string]*ianaEntry)

	scan := bufio.NewScanner(r)
	scan.Split(bufio.ScanWords)
	var record *ianaEntry
	for more := scan.Scan(); more; {
		key := scan.Text()
		more = scan.Scan()
		value := scan.Text()
		switch key {
		case "Type:":
			record = &ianaEntry{typ: value}
		case "Subtag:", "Tag:":
			if s := strings.SplitN(value, "..", 2); len(s) > 1 {
				for a := s[0]; a <= s[1]; a = inc(a) {
					b.addToRegistry(a, record)
				}
			} else {
				b.addToRegistry(value, record)
			}
		case "Suppress-Script:":
			record.suppressScript = value
		case "Added:":
			record.added = value
		case "Deprecated:":
			record.deprecated = value
		case "Macrolanguage:":
			record.macro = value
		case "Preferred-Value:":
			record.preferred = value
		case "Prefix:":
			record.prefix = append(record.prefix, value)
		case "Scope:":
			record.scope = value
		case "Description:":
			buf := []byte(value)
			for more = scan.Scan(); more; more = scan.Scan() {
				b := scan.Bytes()
				if b[0] == '%' || b[len(b)-1] == ':' {
					break
				}
				buf = append(buf, ' ')
				buf = append(buf, b...)
			}
			record.description = append(record.description, string(buf))
			continue
		default:
			continue
		}
		more = scan.Scan()
	}
	if scan.Err() != nil {
		log.Panic(scan.Err())
	}
}

func (b *builder) addToRegistry(key string, entry *ianaEntry) {
	if info, ok := b.registry[key]; ok {
		if info.typ != "language" || entry.typ != "extlang" {
			log.Fatalf("parseRegistry: tag %q already exists", key)
		}
	} else {
		b.registry[key] = entry
	}
}

var commentIndex = make(map[string]string)

func init() {
	for _, s := range comment {
		key := strings.TrimSpace(strings.SplitN(s, " ", 2)[0])
		commentIndex[key] = s
	}
}

func (b *builder) comment(name string) {
	if s := commentIndex[name]; len(s) > 0 {
		b.w.WriteComment(s)
	} else {
		fmt.Fprintln(b.w)
	}
}

func (b *builder) pf(f string, x ...interface{}) {
	fmt.Fprintf(b.hw, f, x...)
	fmt.Fprint(b.hw, "\n")
}

func (b *builder) p(x ...interface{}) {
	fmt.Fprintln(b.hw, x...)
}

func (b *builder) addSize(s int) {