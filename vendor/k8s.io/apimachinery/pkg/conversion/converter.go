/*
Copyright 2014 The Kubernetes Authors.

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

package conversion

import (
	"fmt"
	"reflect"
)

type typePair struct {
	source reflect.Type
	dest   reflect.Type
}

type typeNamePair struct {
	fieldType reflect.Type
	fieldName string
}

// DebugLogger allows you to get debugging messages if necessary.
type DebugLogger interface {
	Logf(format string, args ...interface{})
}

type NameFunc func(t reflect.Type) string

var DefaultNameFunc = func(t reflect.Type) string { return t.Name() }

type GenericConversionFunc func(a, b interface{}, scope Scope) (bool, error)

// Converter knows how to convert one type to another.
type Converter struct {
	// Map from the conversion pair to a function which can
	// do the conversion.
	conversionFuncs          ConversionFuncs
	generatedConversionFuncs ConversionFuncs

	// genericConversions are called during normal conversion to offer a "fast-path"
	// that avoids all reflection. These methods are not called outside of the .Convert()
	// method.
	genericConversions []GenericConversionFunc

	// Set of conversions that should be treated as a no-op
	ignoredConversions map[typePair]struct{}

	// This is a map from a source field type and name, to a list of destination
	// field type and name.
	structFieldDests map[typeNamePair][]typeNamePair

	// Allows for the opposite lookup of structFieldDests. So that SourceFromDest
	// copy flag also works. So this is a map of destination field name, to potential
	// source field name and type to look for.
	structFieldSources map[typeNamePair][]typeNamePair

	// Map from an input type to a function which can apply a key name mapping
	inputFieldMappingFuncs map[reflect.Type]FieldMappingFunc

	// Map from an input type to a set of default conversion flags.
	inputDefaultFlags map[reflect.Type]FieldMatchingFlags

	// If non-nil, will be called to print helpful debugging info. Quite verbose.
	Debug DebugLogger

	// nameFunc is called to retrieve the name of a type; this name is used for the
	// purpose of deciding whether two types match or not (i.e., will we attempt to
	// do a conversion). The default returns the go type name.
	nameFunc func(t reflect.Type) string
}

// NewConverter creates a new Converter object.
func NewConverter(nameFn NameFunc) *Converter {
	c := &Converter{
		conversionFuncs:          NewConversionFuncs(),
		generatedConversionFuncs: NewConversionFuncs(),
		ignoredConversions:       make(map[typePair]struct{}),
		nameFunc:                 nameFn,
		structFieldDests:         make(map[typeNamePair][]typeNamePair),
		structFieldSources:       make(map[typeNamePair][]typeNamePair),

		inputFieldMappingFuncs: make(map[reflect.Type]FieldMappingFunc),
		inputDefaultFlags:      make(map[reflect.Type]FieldMatchingFlags),
	}
	c.RegisterConversionFunc(Convert_Slice_byte_To_Slice_byte)
	return c
}

// AddGenericConversionFunc adds a function that accepts the ConversionFunc call pattern
// (for two conversion types) to the converter. These functions are checked first during
// a normal conversion, but are otherwise not called. Use AddConversionFuncs when registering
// typed conversions.
func (c *Converter) AddGenericConversionFunc(fn GenericConversionFunc) {
	c.genericConversions = append(c.genericConversions, fn)
}

// WithConversions returns a Converter that is a copy of c but with the additional
// fns merged on top.
func (c *Converter) WithConversions(fns ConversionFuncs) *Converter {
	copied := *c
	copied.conversionFuncs = c.conversionFuncs.Merge(fns)
	return &copied
}

// DefaultMeta returns the conversion FieldMappingFunc and meta for a given type.
func (c *Converter) DefaultMeta(t reflect.Type) (FieldMatchingFlags, *Meta) {
	return c.inputDefaultFlags