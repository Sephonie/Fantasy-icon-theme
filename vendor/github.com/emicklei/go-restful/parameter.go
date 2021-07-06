package restful

// Copyright 2013 Ernest Micklei. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

const (
	// PathParameterKind = indicator of Request parameter type "path"
	PathParameterKind = iota

	// QueryParameterKind = indicator of Request parameter type "query"
	QueryParameterKind

	// BodyParameterKind = indicator of Request parameter type "body"
	BodyParameterKind

	// HeaderParameterKind = indicator of Request parameter type "header"
	HeaderParameterKind

	// FormParameterKind = indicator of Request parameter type "form"
	FormParameterKind

	// CollectionFormatCSV comma separated values `foo,bar`
	CollectionFormatCSV = CollectionFormat("csv")

	// CollectionFormatSSV space separated values `foo bar`
	CollectionFormatSSV = CollectionFormat("ssv")

	// CollectionFormatTSV tab separated values `foo\tbar`
	CollectionFormatTSV = CollectionFormat("tsv")

	// CollectionFormatPipes pipe separated values `foo|bar`
	CollectionFormatPipes = CollectionFormat("pipes")

	// CollectionFormatMulti corresponds to multiple parameter instances instead of multiple values for a single
	// instance `foo=bar&foo=baz`. This is valid only for QueryParameters and FormParameters
	CollectionFormatMulti = CollectionFormat("multi")
)

type CollectionFormat string

func (cf CollectionFormat) String() string {
	return string(cf)
}

// Parameter is for documententing the parameter used in a Http Request
// ParameterData kinds are Path,Query and Body
type Parameter struct {
	data *ParameterData
}

// ParameterData represents the state of a Parameter.
// It is ma