// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright (c) 2013, The Prometheus Authors
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Package promhttp contains functions to create http.Handler instances to
// expose Prometheus metrics via HTTP. In later versions of this package, it
// will also contain tooling to instrument instances of http.Handler and
// http.RoundTripper.
//
// promhttp.Handler acts on the prometheus.DefaultGatherer. With HandlerFor,
// you can create a handler for a custom registry or anything that implements
// the Gatherer interface. It also allows to create handlers that act
// differently on errors or allow to log errors.
package promhttp

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/prometheus/common/expfmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	contentTypeHeader     = "Content-Type"
	contentLengthHeader   = "Content-Length"
	contentEncodingHeader = "Content-Encoding"
	acceptEncodingHeader  = "Accept-Encoding"
)

var bufPool sync.Pool

func getBuf() *bytes.Buffer {
	buf := bufPool.Get()
	if buf == nil {
		return &bytes.Buffer{}
	}
	return buf.(*bytes.Buffer)
}

func giveBuf(buf *bytes.Buffer) {
	buf.Reset()
	bufPool.Put(buf)
}

// Handler returns an HTTP handler for the prometheus.DefaultGatherer. The
// Handler uses the default HandlerOpts, i.e. report the first error as an HTTP
// error, no error logging, and compression if requested by the client.
//
// If you want to create a Handler for the DefaultGatherer with different
// HandlerOpts, create it with HandlerFor with prometheus.DefaultGatherer and
// your desired HandlerOpts.
func Handler() http.Handler {
	return HandlerFor(prometheus.DefaultGatherer, HandlerOpts{})
}

// HandlerFor returns an http.Handler for the provided Gatherer. The behavior
// of the Handler is defined by the provided HandlerOpts.
func HandlerFor(reg prometheus.Gatherer, opts HandlerOpts) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mfs, err := reg.Gather()
		if err != nil {
			if opts.ErrorLog != nil {
				opts.ErrorLog.Println("error gathering metrics:", err)
			}
			switch opts.ErrorHandling {
			case PanicOnError:
				panic(err)
			case ContinueOnError:
				if len(mfs) == 0 {
					http.Error(w, "No metrics gathered, last error:\n\n"+err.Error(), http.StatusInternalServerError)
					return
				}
			case HTTPErrorOnError:
				http.Error(w, "An error has occurred during metrics gathering:\n\n"+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		contentType := expfmt.Negotiate(req.Header)
		buf := getBuf()
		defer giveBuf(buf)
		writer, encoding := decorateWriter(req, buf, opts.DisableCompression)
		enc := expfmt.NewEncoder(writer, contentType)
		var lastErr error
		for _, mf := range mfs {
			if err := enc.En