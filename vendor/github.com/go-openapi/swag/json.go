// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package swag

import (
	"bytes"
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
)

// nullJSON represents a JSON object with null type
var nullJSON = []byte("null")

// DefaultJSONNameProvider the default cache for types
var DefaultJSONNameProvider = NewNameProvider()

const comma = byte(',')

var closers = map[byte]byte{
	'{': '}',
	'[': ']',
}

type ejMarshaler interface {
	MarshalEasyJSON(w *jwriter.Writer)
}

type ejUnmarshaler interface {
	UnmarshalEasyJSON(w *jlexer.Lexer)
}

// WriteJSON writes json data, prefers finding an appropriate interface to short-circuit the marshaller
// so it takes the fastest option available.
func WriteJSON(data interface{}) ([]byte, error) {
	if d, ok := data.(ejMarshaler); ok {
		jw := new(jwriter.Writer)
		d.MarshalEasyJSON(jw)
		return jw.BuildBytes()
	}
	if d, ok := data.(json.Marshaler); ok {
		return d.MarshalJSON()
	}
	return json.Marshal(data)
}

// ReadJSON reads json data, prefers finding an appropriate interface to short-circuit the unmarshaller
// so it takes the fastes option available
func ReadJSON(data []byte, value interface{}) error {
	if d, ok := value.(ejUnmarshaler); ok {
		jl := &jlexer.Lexer{Data: data}
		d.UnmarshalEasyJSON(jl)
		return jl.Error()
	}
	if d, ok := value.(json.Unmarshaler); ok {
		return d.UnmarshalJSON(data)
	}
	return json.Unmarshal(data, value)
}

// DynamicJSONToStruct converts an untyped json structure into a struct
func DynamicJSONToStruct(data interface{}, target interface{}) error {
	// TODO: convert straight to a json typed map  (mergo + iterate?)
	b, err := WriteJSON(data)
	if err != nil {
		return err
	}
	if err := ReadJSON(b, target); err != nil {
		return err
	}
	return nil
}

// ConcatJSON concatenates multiple json objects efficiently
func ConcatJSON(blobs ...[]byte) []byte {
	if len(blobs) == 0 {
		return nil
	}

	last := len(blobs) - 1
	for blobs[last] == nil || bytes.Equal(blobs[last], nullJSON) {
		// strips trailing null objects
		last = last - 1
		if l