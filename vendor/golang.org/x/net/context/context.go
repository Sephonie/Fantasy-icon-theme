// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package context defines the Context type, which carries deadlines,
// cancelation signals, and other request-scoped values across API boundaries
// and between processes.
// As of Go 1.7 this package is available in the standard library under the
// name context.  https://golang.org/pkg/context.
//
// Incoming requests to a server should create a Context, and outgoing calls to
// servers should accept a Context. The chain of function calls between must
// propagate the Contex