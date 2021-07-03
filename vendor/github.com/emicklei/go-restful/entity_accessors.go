package restful

// Copyright 2015 Ernest Micklei. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"
	"sync"
)

// EntityReaderWriter can read and write values using an encoding such as JSON,XML.
type EntityReaderWriter interface {
	// Read a serialized version of the value from the request.
	// The Request may have a decompressing reader. Depends on Content-Encoding.
	Read(req *Request, v interface{}) error

	// Write a serialized version of the value on the response.
	// The Response may have a compressing writer. Depends on Accept-Encoding.
	// status should be a valid Http Status code
	Write(resp *Response, status int, v interface{}) error
}

// entityAccessRegistry is a singleton
var entityAccessRegistry = &entityReaderWriters{
	protection: new(sync.RWMutex),
	accessors:  map[string]EntityReaderWriter{},
}

// entityReaderWriters associates MIME to an EntityReaderWriter
type entityReaderWriters struct {
	protection *sync.RWMutex
	accessors  map[string]EntityReaderWriter
}

func init() {
	RegisterEntityAccessor