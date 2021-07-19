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

package spec

import (
	"encoding/json"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/swag"
)

const (
	basic       = "basic"
	apiKey      = "apiKey"
	oauth2      = "oauth2"
	implicit    = "implicit"
	password    = "password"
	application = "application"
	accessCode  = "accessCode"
)

// BasicAuth creates a basic auth security scheme
func BasicAuth() *SecurityScheme {
	return &SecurityScheme{SecuritySchemeProps: SecuritySchemeProps{Type: basic}}
}

// APIKeyAuth creates an api key auth security scheme
func APIKeyAuth(fieldName, valueSource string) *SecurityScheme {
	return &SecurityScheme{SecuritySchemeProps: SecuritySchemeProps{Type: apiKey, Name: fieldName, In: valueSource}}
}

// OAuth2Implicit creates an implicit flow oauth2 security scheme
func OAuth2Implicit(authorizationURL string) *SecurityScheme {
	return &SecurityScheme{SecuritySchemeProps: SecuritySchemeProps{
		Type:             oauth2,
		Flow:             implicit,
		AuthorizationURL: authorizationURL,
	}}
}

// OAuth2Password creates a password flow oauth2 security scheme
func OAuth2Password(tokenURL string) *SecurityScheme {
	return &SecurityScheme{SecuritySchemeProps: SecuritySchemeProps{
		Type:     oauth2,
		Flow:     password,
		TokenURL: tokenURL,
	}}
}

// OAuth2Application creates an application flow oauth2 security scheme
func OAuth2Application(tokenURL string) *SecurityScheme {
	return &SecurityScheme{SecuritySchemeProps: SecuritySchemeProps{
		Type:     oauth2,
		Flow:     application,
		TokenURL: tokenURL,
	}}
}

// OAuth2AccessToken c