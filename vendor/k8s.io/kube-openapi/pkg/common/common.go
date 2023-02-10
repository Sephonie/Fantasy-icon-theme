/*
Copyright 2016 The Kubernetes Authors.

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

package common

import (
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/go-openapi/spec"
)

// OpenAPIDefinition describes single type. Normally these definitions are auto-generated using gen-openapi.
type OpenAPIDefinition struct {
	Schema       spec.Schema
	Dependencies []string
}

type ReferenceCallback func(path string) spec.Ref

// GetOpenAPIDefinitions is collection of all definitions.
type GetOpenAPIDefinitions func(ReferenceCallback) map[string]OpenAPIDefinition

// OpenAPIDefinitionGetter gets openAPI definitions for a given type. If a type implements this interface,
// the definition returned by it will be used, otherwise the auto-generated definitions will be used. See
// GetOpenAPITypeFormat for more information about trade-offs of using this interface or GetOpenAPITypeFormat method when
// possible.
type OpenAPIDefinitionGetter interface {
	OpenAPIDefinition() *OpenAPIDefinition
}

type PathHandler interface {
	Handle(path string, handler http.Handler)
}

// Config is set of configuration for openAPI spec generation.
type Config struct {
	// List of supported protocols such as https, http, etc.
	ProtocolList []string

	// Info is general information about the API.
	Info *spec.Info

	// DefaultResponse wi