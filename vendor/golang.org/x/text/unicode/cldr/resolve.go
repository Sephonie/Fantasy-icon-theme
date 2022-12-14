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
						return nil, fmt.Errorf("cldr: type selector or default value needed for element %s", m[1])
					}
				}
				for ; i < v.Len(); i++ {
					vi := v.Index(i)
					key, err := findField(vi.Elem(), m[2])
					if err != nil {
						return nil, err
					}
					key = reflect.Indirect(key)
					if key.Kind() == reflect.String && key.String() == m[3] {
						break
					}
				}
			}
			if i == v.Len() || v.Index(i).IsNil() {
				return nil, fmt.Errorf("no %s found with %s==%s", m[1], m[2], m[3])
			}
			e = v.Index(i).Interface().(Elem)
		case reflect.Ptr:
			if v.IsNil() {
				return nil, fmt.Errorf("cldr: element %q not found within element %q", m[1], e.GetCommon().name)
			}
			var ok bool
			if e, ok = v.Interface().(Elem); !ok {
				return nil, fmt.Errorf("cldr: %q is not an XML element", m[1])
			} else if m[2] != "" || m[3] != "" {
				return nil, fmt.Errorf("cldr: no type selector allowed for element %s", m[1])
			}
		default:
			return nil, fmt.Errorf("cldr: %q is not an XML element", m[1])
		}
	}
	return e, nil
}

const absPrefix = "//ldml/"

func (cldr *CLDR) resolveAlias(e Elem, src, path string) (res Elem, err error) {
	if src != "locale" {
		if !strings.HasPrefix(path, absPrefix) {
			return nil, fmt.Errorf("cldr: expected absolute path, found %q", path)
		}
		path = path[len(absPrefix):]
		if e, err = cldr.resolve(src); err != nil {
			return nil, err
		}
	}
	return walkXPath(e, path)
}

func (cldr *CLDR) resolveAndMergeAlias(e Elem) error {
	alias := e.GetCommon().Alias
	if alias == nil {
		return nil
	}
	a, err := cldr.resolveAlias(e, alias.Source, alias.Path)
	if err != nil {
		return fmt.Errorf("%v: error evaluating path %q: %v", getPath(e), alias.Path, err)
	}
	// Ensure alias node was already evaluated. TODO: avoid double evaluation.
	err = cldr.resolveAndMergeAlias(a)
	v := reflect.ValueOf(e).Elem()
	for i := iter(reflect.ValueOf(a).Elem()); !i.done(); i.next() {
		if vv := i.value(); vv.Kind() != reflect.Ptr || !vv.IsNil() {
			if _, attr := xmlName(i.field()); !attr {
				v.FieldByIndex(i.index).Set(vv)
			}
		}
	}
	return err
}

func (cldr *CLDR) aliasResolver() visitor {
	return func(v reflect.Value) (err error) {
		if e, ok := v.Addr().Interface().(Elem); ok {
			err = cldr.resolveAndMergeAlias(e)
			if err == nil && blocking[e.GetCommon().name] {
				return stopDescent
			}
		}
		return err
	}
}

// elements within blocking elements do not inherit.
// Taken from CLDR's supplementalMetaData.xml.
var blocking = map[string]bool{
	"identity":         true,
	"supplementalData": true,
	"cldrTest":         true,
	"collation":        true,
	"transform":        true,
}

// Distinguishing attributes affect inheritance; two elements with different
// distinguishing attributes are treated as different for purposes of inheritance,
// except when such attributes occur in the indicated elements.
// Taken from CLDR's supplementalMetaData.xml.
var distinguishing = map[string][]string{
	"key":        nil,
	"request_id": nil,
	"id":         nil,
	"registry":   nil,
	"alt":        nil,
	"iso4217":    nil,
	"iso3166":    nil,
	"mzone":      nil,
	"from":       nil,
	"to":         nil,
	"type": []string{
		"abbreviationFallback",
		"default",
		"mapping",
		"measurementSystem",
		"preferenceOrdering",
	},
	"numberSystem": nil,
}

func in(set []string, s string) bool {
	for _, v := range set {
		if v == s {
			return true
		}
	}
	return false
}

// attrKey computes a key based on the distinguishable attributes of
// an element and it's values.
func attrKey(v reflect.Value, exclude ...string) string {
	parts := []string{}
	ename := v.Interface().(Elem).GetCommon().name
	v = v.Elem()
	for i := iter(v); !i.done(); i.next() {
		if name, attr := xmlName(i.field()); attr {
			if except, ok := distinguishing[name]; ok && !in(exclude, name) && !in(except, ename) {
				v := i.value()
				if v.Kind() == reflect.Ptr {
					v = v.Elem()
				}
				if v.IsValid() {
					parts = append(parts, fmt.Sprintf("%s=%s", name, v.String()))
				}
			}
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, ";")
}

// Key returns a key for e derived from all distinguishing attributes
// except those specified by exclude.
func Key(e Elem, exclude ...string) string {
	return attrKey(reflect.ValueOf(e), exclude...)
}

// linkEnclosing sets the enclosing element as well as the name
// for all sub-elements of child, recursively.
func linkEnclosing(parent, child Elem) {
	child.setEnclosing(parent)
	v := reflect.ValueOf(child).Elem()
	for i := iter(v); !i.done(); i.next() {
		vf := i.value()
		if vf.Kind() == reflect.Slice {
			for j := 0; j < vf.Len(); j++ {
				linkEnclosing(child, vf.Index(j).Interface().(Elem))
			}
		} else if vf.Kind() == reflect.Ptr && !vf.IsNil() && vf.Elem().Kind() == reflect.Struct {
			linkEnclosing(child, vf.Interface().(Elem))
		}
	}
}

func setNames(e Elem, name string) {
	e.setName(name)
	v := reflect.ValueOf(e).Elem()
	for i := iter(v); !i.done(); i.next() {
		vf := i.value()
		name, _ = xmlName(i.field())
		if vf.Kind() == reflect.Slice {
			for j := 0; j < vf.Len(); j++ {
				setNames(vf.Index(j).Interface().(Elem), name)
			}
		} else if vf.Kind() == reflect.Ptr && !vf.IsNil() && vf.Elem().Kind() == reflect.Struct {
			setNames(vf.Interface().(Elem), name)
		}
	}
}

// deepCopy copies elements of v recursively.  All elements of v that may
// be modified by inheritance are explicitly copied.
func deepCopy(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() || v.Elem().Kind() != reflect.Struct {
			return v
		}
		nv := reflect.New(v.Elem().Type())
		nv.Elem().Set(v.Elem())
		deepCopyRec(nv.Elem(), v.Elem())
		return nv
	case reflect.Slice:
		nv := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			deepCopyRec(nv.Index(i), v.Index(i))
		}
		return nv
	}
	panic("deepCopy: must be called with pointer or slice")
}

// deepCopyRec is only called by deepCopy.
func deepCopyRec(nv, v reflect.Value) {
	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if name, attr := xmlName(t.Field(i)); name != "" && !attr {
				deepCopyRec(nv.Field(i), v.Field(i))
			}
		}
	} else {
		nv.Set(deepCopy(v))
	}
}

// newNode is used to insert a missing node during inheritance.
func (cldr *CLDR) newNode(v, enc reflect.Value) reflect.Value {
	n := reflect.New(v.Type())
	for i := iter(v); !i.done(); i.next() {
		if name, attr := xmlName(i.field()); name == "" || attr {
			n.Elem().FieldByIndex(i.index).Set(i.value())
		}
	}
	n.Interface().(Elem).GetCommon().setEnclosing(enc.Addr().Interface().(Elem))
	return n
}

// v, parent must be pointers to struct
func (cldr *CLDR) inheritFields(v, parent reflect.Value) (res reflect.Value, err error) {
	t := v.Type()
	nv := reflect.New(t)
	nv.Elem().Set(v)
	for i := iter(v); !i.done(); i.next() {
		vf := i.value()
		f := i.field()
		name, attr := xmlName(f)
		if name == "" || attr {
			continue
		}
		pf := parent.FieldByIndex(i.index)
		if blocking[name] {
			if vf.IsNil() {
				vf = pf
			}
			nv.Elem().FieldByIndex(i.index).Set(deepCopy(vf))
			continue
		}
		switch f.Type.Kind() {
		case reflect.Ptr:
			if f.Type.Elem().Kind() == reflect.Struct {
				if !vf.IsNil() {
					if vf, err =