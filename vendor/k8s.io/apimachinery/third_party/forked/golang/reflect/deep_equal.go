// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package reflect is a fork of go's standard library reflection package, which
// allows for deep equal with equality functions defined.
package reflect

import (
	"fmt"
	"reflect"
	"strings"
)

// Equalities is a map from type to a function comparing two values of
// that type.
type Equalities map[reflect.Type]reflect.Value

// For convenience, panics on errrors
func EqualitiesOrDie(funcs ...interface{}) Equalities {
	e := Equalities{}
	if err := e.AddFuncs(funcs...); err != nil {
		panic(err)
	}
	return e
}

// AddFuncs is a shortcut for multiple calls to AddFunc.
func (e Equalities) AddFuncs(funcs ...interface{}) error {
	for _, f := range funcs {
		if err := e.AddFunc(f); err != nil {
			return err
		}
	}
	return nil
}

// AddFunc uses func as an equality function: it must take
// two parameters of the same type, and return a boolean.
func (e Equalities) AddFunc(eqFunc interface{}) error {
	fv := reflect.ValueOf(eqFunc)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		return fmt.Errorf("expected func, got: %v", ft)
	}
	if ft.NumIn() != 2 {
		return fmt.Errorf("expected three 'in' params, got: %v", ft)
	}
	if ft.NumOut() != 1 {
		return fmt.Errorf("expected one 'out' param, got: %v", ft)
	}
	if ft.In(0) != ft.In(1) {
		return fmt.Errorf("expected arg 1 and 2 to have same type, but got %v", ft)
	}
	var forReturnType bool
	boolType := reflect.TypeOf(forReturnType)
	if ft.Out(0) != boolType {
		return fmt.Errorf("expected bool return, got: %v", ft)
	}
	e[ft.In(0)] = fv
	return nil
}

// Below here is forked from go's reflect/deepequal.go

// During deepValueEqual, must keep track of checks that are
// in progress.  The comparison algorithm assumes that all
// checks in progress are true when it reencounters them.
// Visited comparisons are stored in a map indexed by visit.
type visit struct {
	a1  uintptr
	a2  uintptr
	typ reflect.Type
}

// unexportedTypePanic is thrown when you use this DeepEqual on something that has an
// unexported type. It indicates a programmer error, so should not occur at runtime,
// which is why it's not public and thus impossible to catch.
type unexportedTypePanic []reflect.Type

func (u unexportedTypePanic) Error() string { return u.String() }
func (u unexportedTypePanic) String() string {
	strs := make([]string, len(u))
	for i, t := range u {
		strs[i] = fmt.Sprintf("%v", t)
	}
	return "an unexported field was encountered, nested like this: " + strings.Join(strs, " -> ")
}

func makeUsefulPanic(v reflect.Value) {
	if x := recover(); x != nil {
		if u, ok := x.(unexportedTypePanic); ok {
			u = append(unexportedTypePanic{v.Type()}, u...)
			x = u
		}
		panic(x)
	}
}

// Tests for deep equality using reflected types. The map argument tracks
// comparisons that have already been seen, which allows short circuiting on
// recursive types.
func (e Equalities) deepValueEqual(v1, v2 reflect.Value, visited map[visit]bool, depth int) bool {
	defer makeUsefulPanic(v1)

	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}
	if v1.Type() != v2.Type() {
		return false
	}
	if fv, ok := e[v1.Type()]; ok {
		return fv.Call([]reflect.Value{v1, v2})[0].Bool()
	}

	hard := func(k reflect.Kind) bool {
		switch k {
		case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
			return true
		}
		return false
	}

	if v1.CanAddr() && v2.CanAddr() && hard(v1.Kind()) {
		addr1 := v1.UnsafeAddr()
		addr2 := v2.UnsafeAddr()
		if addr1 > addr2 {
			// Canonicalize order to reduce number of entries in visited.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are identical ...
		if addr1 == addr2 {
			return true
		}

		// ... or already seen
		typ := v1.Type()
		v := visit{addr1, addr2, typ}
		if visited[v] {
			return true
		}

		// Remember for later.
		visited[v] = true
	}

	switch v1.Kind() {
	case reflect.Array:
		// We don't need to check length here because length is part of
		// an array's type, which has already been filtered for.
		for i := 0; i < v1.Len(); i++ {
			if !e.deepValueEqual(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if (v1.IsNil() || v1.Len() == 0) != (v2.IsNil() || v2.Len() == 0) {
			return false
		}
		if v1.IsNil() || v1.Len() == 0 {
			return true
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for i := 0; i < v1.Len(); i++ {
			if !e.deepValueEqual(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		return e.deepValueEqual(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Ptr:
		return e.deepValueEqual(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			if !e.deepValueEqual(v1.Field(i), v2.Field(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Map:
		if (v1.IsNil() || v1.Len() == 0) != (v2.IsNil() || v2.Len() == 0) {
			return false
		}
		if v1.IsNil() || v1.Len() == 0 {
			return true
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Poi