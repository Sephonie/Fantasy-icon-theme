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

func (s *scopeStack) pop() {
	n := len(*s)
	*s = (*s)[:n-1]
}

func (s *scopeStack) push(e scopeStackElem) {
	*s = append(*s, e)
}

func (s *scopeStack) top() *scopeStackElem {
	return &(*s)[len(*s)-1]
}

func (s scopeStack) describe() string {
	desc := ""
	if len(s) > 1 {
		desc = "(" + s[1].value.Type().String() + ")"
	}
	for i, v := range s {
		if i < 2 {
			// First layer on stack is not real; second is handled specially above.
			continue
		}
		if v.key == "" {
			desc += fmt.Sprintf(".%v", v.value.Type())
		} else {
			desc += fmt.Sprintf(".%v", v.key)
		}
	}
	return desc
}

// Formats src & dest as indices for printing.
func (s *scope) setIndices(src, dest int) {
	s.srcStack.top().key = fmt.Sprintf("[%v]", src)
	s.destStack.top().key = fmt.Sprintf("[%v]", dest)
}

// Formats src & dest as map keys for printing.
func (s *scope) setKeys(src, dest interface{}) {
	s.srcStack.top().key = fmt.Sprintf(`["%v"]`, src)
	s.destStack.top().key = fmt.Sprintf(`["%v"]`, dest)
}

// Convert continues a conversion.
func (s *scope) Convert(src, dest interface{}, flags FieldMatchingFlags) error {
	return s.converter.Convert(src, dest, flags, s.meta)
}

// DefaultConvert continues a conversion, performing a default conversion (no conversion func)
// for the current stack frame.
func (s *scope) DefaultConvert(src, dest interface{}, flags FieldMatchingFlags) error {
	return s.converter.DefaultConvert(src, dest, flags, s.meta)
}

// SrcTag returns the tag of the struct containing the current source item, if any.
func (s *scope) SrcTag() reflect.StructTag {
	return s.srcStack.top().tag
}

// DestTag returns the tag of the struct containing the current dest item, if any.
func (s *scope) DestTag() reflect.StructTag {
	return s.destStack.top().tag
}

// Flags returns the flags with which the current conversion was started.
func (s *scope) Flags() FieldMatchingFlags {
	return s.flags
}

// Meta returns the meta object that was originally passed to Convert.
func (s *scope) Meta() *Meta {
	return s.meta
}

// describe prints the path to get to the current (source, dest) values.
func (s *scope) describe() (src, dest string) {
	return s.srcStack.describe(), s.destStack.describe()
}

// error makes an error that includes information about where we were in the objects
// we were asked to convert.
func (s *scope) errorf(message string, args ...interface{}) error {
	srcPath, destPath := s.describe()
	where := fmt.Sprintf("converting %v to %v: ", srcPath, destPath)
	return fmt.Errorf(where+message, args...)
}

// Verifies whether a conversion function has a correct signature.
func verifyConversionFunctionSignature(ft reflect.Type) error {
	if ft.Kind() != reflect.Func {
		return fmt.Errorf("expected func, got: %v", ft)
	}
	if ft.NumIn() != 3 {
		return fmt.Errorf("expected three 'in' params, got: %v", ft)
	}
	if ft.NumOut() != 1 {
		return fmt.Errorf("expected one 'out' param, got: %v", ft)
	}
	if ft.In(0).Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer arg for 'in' param 0, got: %v", ft)
	}
	if ft.In(1).Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer arg for 'in' param 1, got: %v", ft)
	}
	scopeType := Scope(nil)
	if e, a := reflect.TypeOf(&scopeType).Elem(), ft.In(2); e != a {
		return fmt.Errorf("expected '%v' arg for 'in' param 2, got '%v' (%v)", e, a, ft)
	}
	var forErrorType error
	// This convolution is necessary, otherwise TypeOf picks up on the fact
	// that forErrorType is nil.
	errorType := reflect.TypeOf(&forErrorType).Elem()
	if ft.Out(0) != errorType {
		return fmt.Errorf("expected error return, got: %v", ft)
	}
	return nil
}

// RegisterConversionFunc registers a conversion func with the
// Converter. conversionFunc must take three parameters: a pointer to the input
// type, a pointer to the output type, and a conversion.Scope (which should be
// used if recursive conversion calls are desired).  It must return an error.
//
// Example:
// c.RegisterConversionFunc(
//         func(in *Pod, out *v1.Pod, s Scope) error {
//                 // conversion logic...
//                 return nil
//          })
func (c *Converter) RegisterConversionFunc(conversionFunc interface{}) error {
	return c.conversionFuncs.Add(conversionFunc)
}

// Similar to RegisterConversionFunc, but registers conversion function that were
// automatically generated.
func (c *Converter) RegisterGeneratedConversionFunc(conversionFunc interface{}) error {
	return c.generatedConversionFuncs.Add(conversionFunc)
}

// RegisterIgnoredConversion registers a "no-op" for conversion, where any requested
// conversion between from and to is ignored.
func (c *Converter) RegisterIgnoredConversion(from, to interface{}) error {
	typeFrom := reflect.TypeOf(from)
	typeTo := reflect.TypeOf(to)
	if reflect.TypeOf(from).Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer arg for 'from' param 0, got: %v", typeFrom)
	}
	if typeTo.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer arg for 'to' param 1, got: %v", typeTo)
	}
	c.ignoredConversions[typePair{typeFrom.Elem(), typeTo.Elem()}] = struct{}{}
	return nil
}

// IsConversionIgnored returns true if the specified objects should be dropped during
// conversion.
func (c *Converter) IsConversionIgnored(inType, outType reflect.Type) bool {
	_, found := c.ignoredConversions[typePair{inType, outType}]
	return found
}

func (c *Converter) HasConversionFunc(inType, outType reflect.Type) bool {
	_, found := c.conversionFuncs.fns[typePair{inType, outType}]
	return found
}

func (c *Converter) ConversionFuncValue(inType, outType reflect.Type) (reflect.Value, bool) {
	value, found := c.conversionFuncs.fns[typePair{inType, outType}]
	return value, found
}

// SetStructFieldCopy registers a correspondence. Whenever a struct field is encountered
// which has a type and name matching srcFieldType and srcFieldName, it wil be copied
// into the field in the destination struct matching destFieldType & Name, if such a
// field exists.
// May be called multiple times, even for the same source field & type--all applicable
// copies will be performed.
func (c *Converter) SetStructFieldCopy(srcFieldType interface{}, srcFieldName string, destFieldType interface{}, destFieldName string) error {
	st := reflect.TypeOf(srcFieldType)
	dt := reflect.TypeOf(destFieldType)
	srcKey := typeNamePair{st, srcFieldName}
	destKey := typeNamePair{dt, destFieldName}
	c.structFieldDests[srcKey] = append(c.structFieldDests[srcKey], destKey)
	c.structFieldSources[destKey] = append(c.structFieldSources[destKey], srcKey)
	return nil
}

// RegisterInputDefaults registers a field name mapping function, used when converting
// from maps to structs. Inputs to the conversion methods are checked for this type and a mapping
// applied automatically if the input matches in. A set of default flags for the input conversion
// may also be provided, which will be used when no explicit flags are requested.
func (c *Converter) RegisterInputDefaults(in interface{}, fn FieldMappingFunc, defaultFlags FieldMatchingFl