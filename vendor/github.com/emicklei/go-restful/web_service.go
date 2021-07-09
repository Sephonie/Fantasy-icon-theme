package restful

import (
	"errors"
	"os"
	"reflect"
	"sync"

	"github.com/emicklei/go-restful/log"
)

// Copyright 2013 Ernest Micklei. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

// WebService holds a collection of Route values that bind a Http Method + URL Path to a function.
type WebService struct {
	rootPath       string
	pathExpr       *pathExpression // cached compilation of rootPath as RegExp
	routes         []Route
	produces       []string
	consumes       []string
	pathParameters []*Parameter
	filters        []FilterFunction
	documentation  string
	apiVersion     string

	typeNameHandleFunc TypeNameHandleFunction

	dynamicRoutes bool

	// protects 'routes' if dynamic routes are enabled
	routesLock sync.RWMutex
}

func (w *WebService) SetDynamicRoutes(enable bool) {
	w.dynamicRoutes = enable
}

// TypeNameHandleFunction declares functions that can handle translating the name of a sample object
// into the restful documentation for the service.
type TypeNameHandleFunction func(sample interface{}) string

// TypeNameHandler sets the function that will convert types to strings in the parameter
// and model definitions. If not set, the web service will invoke
// reflect.TypeOf(object).String().
func (w *WebService) TypeNameHandler(handler TypeNameHandleFunction) *WebService {
	w.typeNameHandleFunc = handler
	return w
}

// reflectTypeName is the default TypeNameHandleFunction and for a given object
// returns the name that Go identifies it with (e.g. "string" or "v1.Object") via
// the reflection API.
func reflectTypeName(sample interface{}) string {
	return reflect.TypeOf(sample).String()
}

// compilePathExpression ensures that the path is compiled into a RegEx for those routers that need it.
func (w *WebService) compilePathExpression() {
	compiled, err := newPathExpression(w.rootPath)
	if err != nil {
		log.Printf("[restful] invalid path:%s because:%v", w.rootPath, err)
		os.Exit(1)
	}
	w.pathExpr = compiled
}

// ApiVersion sets the API version for documentation purposes.
func (w *WebService) ApiVersion(apiVersion string) *WebService {
	w.apiVersion = apiVersion
	return w
}

// Version returns the API version for documentation purposes.
func (w *WebService) Version() string { return w.apiVersion }

// Path specifies the root URL template path of the WebService.
// All Routes will be relative to this path.
func (w *WebService) Path(root string) *WebService {
	w.rootPath = root
	if len(w.rootPath) == 0 {
		w.rootPath = "/"
	}
	w.compilePathExpression()
	return w
}

// Param adds a PathParameter to document parameters used in the root path.
func (w *WebService) Param(parameter *Parameter) *WebService {
	if w.pathParameters == nil {
		w.pathParameters = []*Parameter{}
	}
	w.pathParameters = append(w.pathParameters, parameter)
	return w
}

// PathParameter creates a new Parameter of kind Path for documentation purposes.
// It is initialized as required with string as its DataType.
func (w *WebService) PathParameter(name, description string) *Parameter {
	return PathParameter(name, description)
}

// PathParameter creates a new Parameter of kind Path for documentation purposes.
// It is initialized as required with string as its DataType.
func PathParameter(name, description string) *Parameter {
	p := &Parameter{&ParameterData{Name: name, Description: description, Required: true, DataType: "string"}}
	p.bePath()
	return p
}

// QueryParameter creates a new Parameter of kind Query for documentation purposes.
// It is initialized as not required with string as its DataType.
func (w *WebService) QueryParameter(name, description string) *Parameter {
	return QueryParameter(name, description)
}

// QueryParameter creates a new Parameter of kind Query for documentation purposes.
// It is initialized