package restful

import (
	"bytes"
	"strings"
)

// Copyright 2018 Ernest Micklei. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

// PathProcessor is extra behaviour that a Router can provide to extract path parameters from the path.
// If a Router does not implement this interface then the default behaviour will be used.
type PathProcessor interface {
	// ExtractParameters gets the path parameters defined in the route and webService from the urlPath
	ExtractParameters(route *Route, webService *WebService, urlPath string) map[string]string
}

type defaultPathProcessor struct{}

// Extract the parameters from the request url path
func (d defaultPathProcessor) ExtractParameters(r *Route, _ *WebService, urlPath string) map[string]string {
	urlParts := tokenizePath(urlPath)
	pathParameters := map[string]string{}
	for i, k