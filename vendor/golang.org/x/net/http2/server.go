// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO: turn off the serve goroutine when idle, so
// an idle conn only has the readFrames goroutine active. (which could
// also be optimized probably to pin less memory in crypto/tls). This
// would involve tracking when the serve goroutine is active (atomic
// int32 read/CAS probably?) and starting it up when frames arrive,
// and shutting it down when all handlers exit. the occasional PING
// packets could use time.AfterFunc to call sc.wakeStartServeLoop()
// (which is a no-op if already running) and then queue the PING write
// as normal. The serve loop would then exit in most cases (if no
// Handlers running) and not be woken up again until the PING packet
// returns.

// TODO (maybe): add a mechanism for Handlers to going into
// half-closed-local mode (rw.(io.Closer) test?) but not exit their
// handler, and continue to be able to read from the
// Request.Body. This would be a somewhat semantic change from HTTP/1
// (or at least what we expose in net/http), so I'd probably want to
// add it there too. For now, this package says that returning from
// the Handler ServeHTTP function means you're both done reading and
// done writing, without a way to stop just one or the other.

package http2

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2/hpack"
)

const (
	prefaceTimeout        = 10 * time.Second
	firstSettingsTimeout  = 2 * time.Second // should be in-flight with preface anyway
	handlerChunkWriteSize = 4 << 10
	defaultMaxStreams     = 250 // TODO: make this 100 as the GFE seems to?
)

var (
	errClientDisconnected = errors.New("client disconnected")
	errClosedBody         = errors.New("body closed by handler")
	errHandlerComplete    = errors.New("http2: request body closed due to handler exit