// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cldr

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

// A Decoder loads an archive of CLDR data.
type Decoder struct {
	dirFilter     []string
	sectionFilter []string
	loader        Loader
	cldr          *CLDR
	curLocale     string
}

// SetSectionFilter takes a list top-level LDML element names to which
// evaluation of LDML should be limited.  It automatically calls SetDirFilter.
func (d *Decoder) SetSectionFilter(filter ...string) {
	d.sectionFilter = filter
	// TODO: automatically set dir filter
}

// SetDirFilter limits the loading of LDML XML files of the specied directories.
// Note that sections may be split across directories differently for different CLDR versions.
// For more robust code, use SetSectionFilter.
func (d *Decoder) SetDirFilter(dir ...string) {
	d.dirFilter = dir
}

// A Loader provides access to the files of a CLDR archive.
type Loader interface {
	Len() int
	Path(i int) string
	Reader(i int) (io.ReadCloser, error)
}

var fileRe = regexp.MustCompile(`.*[/\\](.*)[/\\](.*)\.xml`)

// Decode loads and decodes the files represented by l.
func (d *Decoder) Decode(l Loader) (cldr *CLDR, err error) {
	d.cldr = makeCLDR()
	for i := 0; i < l.Len(); i++ {
		fname := l.Path(i)
		if m := fileRe.FindStringSubmatch(fname); m != nil {
			if len(d.dirFilter) > 0 && !in(d.dirFi