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
	return c.inputDefaultFlags[t], &Meta{
		KeyNameMapping: c.inputFieldMappingFuncs[t],
	}
}

// Convert_Slice_byte_To_Slice_byte prevents recursing into every byte
func Convert_Slice_byte_To_Slice_byte(in *[]byte, out *[]byte, s Scope) error {
	if *in == nil {
		*out = nil
		return nil
	}
	*out = make([]byte, len(*in))
	copy(*out, *in)
	return nil
}

// Scope is passed to conversion funcs to allow them to continue an ongoing conversion.
// If multiple converters exist in the system, Scope will allow you to use the correct one
// from a conversion function--that is, the one your conversion function was called by.
type Scope interface {
	// Call Convert to convert sub-objects. Note that if you call it with your own exact
	// parameters, you'll run out of stack space before anything useful happens.
	Convert(src, dest interface{}, flags FieldMatchingFlags) error

	// DefaultConvert performs the default conversion, without calling a conversion func
	// on the current stack frame. This makes it safe to call from a conversion func.
	DefaultConvert(src, dest interface{}, flags FieldMatchingFlags) error

	// SrcTags and DestTags contain the struct tags that src and dest had, respectively.
	// If the enclosing object was not a struct, then these will contain no tags, of course.
	SrcTag() reflect.StructTag
	DestTag() reflect.StructTag

	// Flags returns the flags with which the conversion was started.
	Flags() FieldMatchingFlags

	// Meta returns any information originally passed to Convert.
	Meta() *Meta
}

// FieldMappingFunc can convert an input field value into different values, depending on
// the value of the source or destination struct tags.
type FieldMappingFunc func(key string, sourceTag, destTag reflect.StructTag) (source string, dest string)

func NewConversionFuncs() ConversionFuncs {
	return ConversionFuncs{fns: make(map[typePair]reflect.Value)}
}

type ConversionFuncs struct {
	fns map[typePair]reflect.Value
}

// Add adds the provided conversion functions to the lookup table - they must have the signature
// `func(type1, type2, Scope) error`. Functions are added in the order passed and will override
// previously registered pairs.
func (c ConversionFuncs) Add(fns ...interface{}) error {
	for _, fn := range fns {
		fv := reflect.ValueOf(fn)
		ft := fv.Type()
		if err := verifyConversionFunctionSignature(ft); err != nil {
			return err
		}
		c.fns[typePair{ft.In(0).Elem(), ft.In(1).Elem()}] = fv
	}
	return nil
}

// Merge returns a new ConversionFuncs that contains all conversions from
// both other and c, with other conversions taking precedence.
func (c ConversionFuncs) Merge(other ConversionFuncs) ConversionFuncs {
	merged := NewConversionFuncs()
	for k, v := range c.fns {
		merged.fns[k] = v
	}
	for k, v := range other.fns {
		merged.fns[k] = v
	}
	return merged
}

// Meta is supplied by Scheme, when it calls Convert.
type Meta struct {
	// KeyNameMapping is an optional function which may map the listed key (field name)
	// into a source and destination value.
	KeyNameMapping FieldMappingFunc
	// Context is an optional field that callers may use to pass info to conversion functions.
	Context interface{}
}

// scope contains information about an ongoing conversion.
type scope struct {
	converter *Converter
	meta      *Meta
	flags     FieldMatchingFlags

	// srcStack & destStack are separate because they may not have a 1:1
	// relationship.
	srcStack  scopeStack
	destStack scopeStack
}

type scopeStackElem struct {
	tag   reflect.StructTag
	value reflect.Value
	key   string
}

type scopeStack []scopeStackElem

f