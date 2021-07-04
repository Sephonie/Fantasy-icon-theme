package restful

// Copyright 2013 Ernest Micklei. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
)

// RouterJSR311 implements the flow for matching Requests to Routes (and consequently Resource Functions)
// as specified by the JSR311 http://jsr311.java.net/nonav/releases/1.1/spec/spec.html.
// RouterJSR311 implements the Router interface.
// Concept of locators is not implemented.
type RouterJSR311 struct{}

// SelectRoute is part of the Router interface and returns the best match
// for the WebService and its Route for the given Request.
func (r RouterJSR311) SelectRoute(
	webServices []*WebService,
	httpRequest *http.Request) (selectedService *WebService, selectedRoute *Route, err error) {

	// Identify the root resource class (WebService)
	dispatcher, finalMatch, err := r.detectDispatcher(httpRequest.URL.Path, webServices)
	if err != nil {
		return nil, nil, NewError(http.StatusNotFound, "")
	}
	// Obtain the set of candidate methods (Routes)
	routes := r.selectRoutes(dispatcher, finalMatch)
	if len(routes) == 0 {
		return dispatcher, nil, NewError(http.StatusNotFound, "404: Page Not Found")
	}

	// Identify