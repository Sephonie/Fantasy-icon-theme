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
		log.Panicf("value %v must be a struct", v)
	}
	i := fieldIter{
		v:     v,
		index: []int{0},
		n:     []int{v.NumField()},
	}
	i.descent()
	return i
}

func (i *fieldIter) descent() {
	for f := i.field(); f.Anonymous && f.Type.NumField() > 0; f = i.field() {
		i.index = append(i.index, 0)
		i.n = append(i.n, f.Type.NumField())
	}
}

func (i *fieldIter) done() bool {
	return len(i.index) == 1 && i.index[0] >= i.n[0]
}

func skip(f reflect.StructField) bool {
	return !f.Anonymous && (f.Name[0] < 'A' || f.Name[0] > 'Z')
}

func (i *fieldIter) next() {
	for {
		k := len(i.index) - 1
		i.index[k]++
		if i.index[k] < i.n[k] {
			if !skip(i.field()) {
				break
			}
		} else {
			if k == 0 {
				return
			}
			i.index = i.index[:k]
			i.n = i.n[:k]
		}
	}
	i.descent()
}

func (i *fieldIter) value() reflect.Value {
	return i.v.FieldByIndex(i.index)
}

func (i *fieldIter) field() reflect.StructField {
	return i.v.Type().FieldByIndex(i.index)
}

type visitor func(v reflect.Value) error

var stopDescent = fmt.Errorf("do not recurse")

func (f visitor) visit(x interface{}) error {
	return f.visitRec(reflect.ValueOf(x))
}

// visit recursively calls f on all nodes in v.
func (f visitor) visitRec(v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		return f.visitRec(v.Elem())
	}
	if err := f(v); err != nil {
		if err == stopDescent {
			return nil
		}
		return err
	}
	switch v.Kind() {
	case reflect.Struct:
		for i := iter(v); !i.done(); i.next() {
			if err := f.visitRec(i.value()); err != nil {
				return err
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if err := f.visitRec(v.Index(i)); err != nil {
				return err
			}
		}
	}
	return nil
}

// getPath is used for error reporting purposes only.
func getPath(e Elem) string {
	if e == nil {
		return "<nil>"
	}
	if e.enclosing() == nil {
		return e.GetCommon().name
	}
	if e.GetCommon().Type == "" {
		return fmt.Sprintf("%s.%s", getPath(e.enclosing()), e.GetCommon().name)
	}
	return fmt.Sprintf("%s.%s[type=%s]", getPath(e.enclosing()), e.GetCommon().name, e.GetCommon().Type)
}

// xmlName returns the xml name of the element or attribute
func xmlName(f reflect.StructField) (name string, attr bool) {
	tags := strings.Split(f.Tag.Get("xml"), ",")
	for _, s := range tags {
		attr = attr || s == "attr"
	}
	return tags[0], attr
}

func findField(v reflect.Value, key string) (reflect.Value, error) {
	v = reflect.Indirect(v)
	for i := iter(v); !i.done(); i.next() {
		if n, _ := xmlName(i.field()); n == key {
			return i.value(), nil
		}
	}
	return reflect.Value{}, fmt.Errorf("cldr: no field %q in element %#v", key, v.Interface())
}

var xpathPart = regexp.MustCompile(`(\pL+)(?:\[@(\pL+)='([\w-]+)'\])?`)

func walkXPath(e Elem, path string) (res Elem, err error) {
	for _, c := range strings.Split(path, "/") {
		if c == ".." {
			if e = e.enclosing(); e == nil {
				panic("path ..")
				return nil, fmt.Errorf(`cldr: ".." moves past root in path %q`, path)
			}
			continue
		} else if c == "" {
			continue
		}
		m := xpathPart.FindStringSubmatch(c)
		if len(m) == 0 || len(m[0]) != len(c) {
			return nil, fmt.Errorf("cldr: syntax error in path component %q", c)
		}
		v, err := findField(reflect.ValueOf(e), m[1])
		if err != nil {
			return nil, err
		}
		switch v.Kind() {
		case reflect.Slice:
			i := 0
			if m[2] != "" || v.Len() > 1 {
				if m[2] == "" {
					m[2] = "type"
					if m[3] = e.GetCommon().Default(); m[3] == "" {
						return nil, fmt.