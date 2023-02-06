/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file was autogenerated by set-gen. Do not edit it manually!

package sets

import (
	"reflect"
	"sort"
)

// sets.Int is a set of ints, implemented via map[int]struct{} for minimal memory consumption.
type Int map[int]Empty

// New creates a Int from a list of values.
func NewInt(items ...int) Int {
	ss := Int{}
	ss.Insert(items...)
	return ss
}

// IntKeySet creates a Int from a keys of a map[int](? extends interface{}).
// If the value passed in is not actually a map, this will panic.
func IntKeySet(theMap interface{}) Int {
	v := reflect.ValueOf(theMap)
	ret := Int{}

	for _, keyValue := range v.MapKeys() {
		ret.Insert(keyValue.Interface().(int))
	}
	return ret
}

// Insert adds items to the set.
func (s Int) Insert(items ...int) {
	for _, item := range items {
		s[item] = Empty{}
	}
}

// Delete removes all items from the set.
func (s Int) Delete(items ...int) {
	for _, item := range items {
		delete(s, item)
	}
}

// Has returns true if and only if item is contained in the set.
func (s Int) Has(item int) bool {
	_, contained := s[item]
	return contained
}

// HasAll returns true if and only if all items are contained in the set.
func (s Int) HasAll(items ...int) bool {
	f