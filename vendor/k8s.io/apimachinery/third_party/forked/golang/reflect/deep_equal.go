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

// Below here is forked