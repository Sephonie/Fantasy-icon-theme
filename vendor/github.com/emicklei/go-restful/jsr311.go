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

	// Identify the method (Route) that will handle the request
	route, ok := r.detectRoute(routes, httpRequest)
	return dispatcher, route, ok
}

// ExtractParameters is used to obtain the path parameters from the route using the same matching
// engine as the JSR 311 router.
func (r RouterJSR311) ExtractParameters(route *Route, webService *WebService, urlPath string) map[string]string {
	webServiceExpr := webService.pathExpr
	webServiceMatches := webServiceExpr.Matcher.FindStringSubmatch(urlPath)
	pathParameters := r.extractParams(webServiceExpr, webServiceMatches)
	routeExpr := route.pathExpr
	routeMatches := routeExpr.Matcher.FindStringSubmatch(webServiceMatches[len(webServiceMatches)-1])
	routeParams := r.extractParams(routeExpr, routeMatches)
	for key, value := range routeParams {
		pathParameters[key] = value
	}
	return pathParameters
}

func (RouterJSR311) extractParams(pathExpr *pathExpression, matches []string) map[string]string {
	params := map[string]string{}
	for i := 1; i < len(matches); i++ {
		if len(pathExpr.VarNames) >= i {
			params[pathExpr.VarNames[i-1]] = matches[i]
		}
	}
	return params
}

// http://jsr311.java.net/nonav/releases/1.1/spec/spec3.html#x3-360003.7.2
func (r RouterJSR311) detectRoute(routes []Route, httpRequest *http.Request) (*Route, error) {
	ifOk := []Route{}
	for _, each := range routes {
		ok := true
		for _, fn := range each.If {
			if !fn(httpRequest) {
				ok = false
				break
			}
		}
		if ok {
			ifOk = append(ifOk, each)
		}
	}
	if len(ifOk) == 0 {
		if trace {
			traceLogger.Printf("no Route found (from %d) that passes conditional checks", len(routes))
		}
		return nil, NewError(http.StatusNotFound, "404: Not Found")
	}

	// http method
	methodOk := []Route{}
	for _, each := range ifOk {
		if httpRequest.Method == each.Method {
			methodOk = append(methodOk, each)
		}
	}
	if len(methodOk) == 0 {
		if trace {
			traceLogger.Printf("no Route found (in %d routes) that matches HTTP method %s\n", len(routes), httpRequest.Method)
		}
		return nil, NewError(http.StatusMethodNotAllowed, "405: Method Not Allowed")
	}
	inputMediaOk := methodOk

	// content-type
	contentType := httpRequest.Header.Get(HEADER_ContentType)
	inputMediaOk = []Route{}
	for _, each := range methodOk {
		if each.matchesContentType(contentType) {
			inputMediaOk = append(inputMediaOk, each)
		}
	}
	if len(inputMediaOk) == 0 {
		if trace {
			traceLogger.Printf("no Route found (from %d) that matches HTTP Content-Type: %s\n", len(methodOk), contentType)
		}
		return nil, NewError(http.StatusUnsupportedMediaType, "415: Unsupported Media Type")
	}

	// accept
	outputMediaOk := []Route{}
	accept := httpRequest.Header.Get(HEADER_Accept)
	if len(accept) == 0 {
		accept = "*/*"
	}
	for _, each := range inputMediaOk {
		if each.matchesAccept(accept) {
			outputMediaOk = append(outputMediaOk, each)
		}
	}
	if len(outputMediaOk) == 0 {
		if trace {
			traceLogger.Printf("no Route found (from %d) that matches HTTP Accept: %s\n", len(inputMediaOk), accept)
		}
		return nil, NewError(http.StatusNotAcceptable, "406: Not Acceptable")
	}
	// return r.bestMatchByMedia(outputMediaOk, contentType, accept), nil
	return &outputMediaOk[0], nil
}

// http://jsr311.java.net/nonav/releases/1.1/spec/spec3.html#x3-360003.7.2
// n/m > n/* > */*
func (r RouterJSR311) bestMatchByMedia(routes []Route, contentType string, accept string) *Route {
	// TODO
	return &routes[0]
}

// http://jsr311.java.net/nonav/releases/1.1/spec/spec3.html#x3-360003.7.2  (step 2)
func 