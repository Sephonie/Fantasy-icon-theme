package restful

// Copyright 2013 Ernest Micklei. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

import (
	"net/http"
)

// DefaultContainer is a restful.Container that uses http.DefaultServeMux
var DefaultContainer *Container

func init() {
	DefaultContainer = NewContainer()
	DefaultContainer.ServeMux = http.DefaultServeMux
}

// If set the true then panics will not be caught to return HTTP 500.
// In that case, Route functions are responsible for handling any error situation.
// Default value is false = recover from panics. This has performance im