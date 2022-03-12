// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.8

package http2

import (
	"crypto/tls"
	"io"
	"net/http"
)

func cloneTLSConfig(c *tls.Config) *tls.Config {
	c2 := c.Clone()
	c2.GetClientCertificate = c.GetClientCertificate // golang.org/issue/19264
	return c2
}

var _ http.Pusher = (*responseWriter)(nil)

// Push implements http.Pusher.
func (w *responseW