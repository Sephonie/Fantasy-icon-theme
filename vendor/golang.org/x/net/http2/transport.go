
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Transport code.

package http2

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	mathrand "math/rand"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2/hpack"
	"golang.org/x/net/idna"
	"golang.org/x/net/lex/httplex"
)

const (
	// transportDefaultConnFlow is how many connection-level flow control
	// tokens we give the server at start-up, past the default 64k.
	transportDefaultConnFlow = 1 << 30

	// transportDefaultStreamFlow is how many stream-level flow
	// control tokens we announce to the peer, and how many bytes
	// we buffer per stream.
	transportDefaultStreamFlow = 4 << 20

	// transportDefaultStreamMinRefresh is the minimum number of bytes we'll send
	// a stream-level WINDOW_UPDATE for at a time.
	transportDefaultStreamMinRefresh = 4 << 10

	defaultUserAgent = "Go-http-client/2.0"
)

// Transport is an HTTP/2 Transport.
//
// A Transport internally caches connections to servers. It is safe
// for concurrent use by multiple goroutines.
type Transport struct {
	// DialTLS specifies an optional dial function for creating
	// TLS connections for requests.
	//
	// If DialTLS is nil, tls.Dial is used.
	//
	// If the returned net.Conn has a ConnectionState method like tls.Conn,
	// it will be used to set http.Response.TLS.
	DialTLS func(network, addr string, cfg *tls.Config) (net.Conn, error)

	// TLSClientConfig specifies the TLS configuration to use with
	// tls.Client. If nil, the default configuration is used.
	TLSClientConfig *tls.Config

	// ConnPool optionally specifies an alternate connection pool to use.
	// If nil, the default is used.
	ConnPool ClientConnPool

	// DisableCompression, if true, prevents the Transport from
	// requesting compression with an "Accept-Encoding: gzip"
	// request header when the Request contains no existing
	// Accept-Encoding value. If the Transport requests gzip on
	// its own and gets a gzipped response, it's transparently
	// decoded in the Response.Body. However, if the user
	// explicitly requested gzip it is not automatically
	// uncompressed.
	DisableCompression bool

	// AllowHTTP, if true, permits HTTP/2 requests using the insecure,
	// plain-text "http" scheme. Note that this does not enable h2c support.
	AllowHTTP bool

	// MaxHeaderListSize is the http2 SETTINGS_MAX_HEADER_LIST_SIZE to
	// send in the initial settings frame. It is how many bytes
	// of response headers are allowed. Unlike the http2 spec, zero here
	// means to use a default limit (currently 10MB). If you actually
	// want to advertise an ulimited value to the peer, Transport
	// interprets the highest possible value here (0xffffffff or 1<<32-1)
	// to mean no limit.
	MaxHeaderListSize uint32

	// t1, if non-nil, is the standard library Transport using
	// this transport. Its settings are used (but not its
	// RoundTrip method, etc).
	t1 *http.Transport

	connPoolOnce  sync.Once
	connPoolOrDef ClientConnPool // non-nil version of ConnPool
}

func (t *Transport) maxHeaderListSize() uint32 {
	if t.MaxHeaderListSize == 0 {
		return 10 << 20
	}
	if t.MaxHeaderListSize == 0xffffffff {
		return 0
	}
	return t.MaxHeaderListSize
}

func (t *Transport) disableCompression() bool {
	return t.DisableCompression || (t.t1 != nil && t.t1.DisableCompression)
}

var errTransportVersion = errors.New("http2: ConfigureTransport is only supported starting at Go 1.6")

// ConfigureTransport configures a net/http HTTP/1 Transport to use HTTP/2.
// It requires Go 1.6 or later and returns an error if the net/http package is too old
// or if t1 has already been HTTP/2-enabled.
func ConfigureTransport(t1 *http.Transport) error {
	_, err := configureTransport(t1) // in configure_transport.go (go1.6) or not_go16.go
	return err
}

func (t *Transport) connPool() ClientConnPool {
	t.connPoolOnce.Do(t.initConnPool)
	return t.connPoolOrDef
}

func (t *Transport) initConnPool() {
	if t.ConnPool != nil {
		t.connPoolOrDef = t.ConnPool
	} else {
		t.connPoolOrDef = &clientConnPool{t: t}
	}
}

// ClientConn is the state of a single HTTP/2 client connection to an
// HTTP/2 server.
type ClientConn struct {
	t         *Transport
	tconn     net.Conn             // usually *tls.Conn, except specialized impls
	tlsState  *tls.ConnectionState // nil only for specialized impls
	singleUse bool                 // whether being used for a single http.Request

	// readLoop goroutine fields:
	readerDone chan struct{} // closed on error
	readerErr  error         // set before readerDone is closed

	idleTimeout time.Duration // or 0 for never
	idleTimer   *time.Timer

	mu              sync.Mutex // guards following
	cond            *sync.Cond // hold mu; broadcast on flow/closed changes
	flow            flow       // our conn-level flow control quota (cs.flow is per stream)
	inflow          flow       // peer's conn-level flow control
	closed          bool
	wantSettingsAck bool                     // we sent a SETTINGS frame and haven't heard back
	goAway          *GoAwayFrame             // if non-nil, the GoAwayFrame we received
	goAwayDebug     string                   // goAway frame's debug data, retained as a string
	streams         map[uint32]*clientStream // client-initiated
	nextStreamID    uint32
	pendingRequests int                       // requests blocked and waiting to be sent because len(streams) == maxConcurrentStreams
	pings           map[[8]byte]chan struct{} // in flight ping data to notification channel
	bw              *bufio.Writer
	br              *bufio.Reader
	fr              *Framer
	lastActive      time.Time
	// Settings from peer: (also guarded by mu)
	maxFrameSize          uint32
	maxConcurrentStreams  uint32
	peerMaxHeaderListSize uint64
	initialWindowSize     uint32

	hbuf    bytes.Buffer // HPACK encoder writes into this
	henc    *hpack.Encoder
	freeBuf [][]byte

	wmu  sync.Mutex // held while writing; acquire AFTER mu if holding both
	werr error      // first write error that has occurred
}

// clientStream is the state for a single HTTP/2 stream. One of these
// is created for each Transport.RoundTrip call.
type clientStream struct {
	cc            *ClientConn
	req           *http.Request
	trace         *clientTrace // or nil
	ID            uint32
	resc          chan resAndError
	bufPipe       pipe // buffered pipe with the flow-controlled response payload
	startedWrite  bool // started request body write; guarded by cc.mu
	requestedGzip bool
	on100         func() // optional code to run if get a 100 continue response

	flow        flow  // guarded by cc.mu
	inflow      flow  // guarded by cc.mu
	bytesRemain int64 // -1 means unknown; owned by transportResponseBody.Read
	readErr     error // sticky read error; owned by transportResponseBody.Read
	stopReqBody error // if non-nil, stop writing req body; guarded by cc.mu
	didReset    bool  // whether we sent a RST_STREAM to the server; guarded by cc.mu

	peerReset chan struct{} // closed on peer reset
	resetErr  error         // populated before peerReset is closed

	done chan struct{} // closed when stream remove from cc.streams map; close calls guarded by cc.mu

	// owned by clientConnReadLoop:
	firstByte    bool // got the first response byte
	pastHeaders  bool // got first MetaHeadersFrame (actual headers)
	pastTrailers bool // got optional second MetaHeadersFrame (trailers)

	trailer    http.Header  // accumulated trailers
	resTrailer *http.Header // client's Response.Trailer
}

// awaitRequestCancel waits for the user to cancel a request or for the done
// channel to be signaled. A non-nil error is returned only if the request was
// canceled.
func awaitRequestCancel(req *http.Request, done <-chan struct{}) error {
	ctx := reqContext(req)
	if req.Cancel == nil && ctx.Done() == nil {
		return nil
	}
	select {
	case <-req.Cancel:
		return errRequestCanceled
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// awaitRequestCancel waits for the user to cancel a request, its context to
// expire, or for the request to be done (any way it might be removed from the
// cc.streams map: peer reset, successful completion, TCP connection breakage,
// etc). If the request is canceled, then cs will be canceled and closed.
func (cs *clientStream) awaitRequestCancel(req *http.Request) {
	if err := awaitRequestCancel(req, cs.done); err != nil {
		cs.cancelStream()
		cs.bufPipe.CloseWithError(err)
	}
}

func (cs *clientStream) cancelStream() {
	cc := cs.cc
	cc.mu.Lock()
	didReset := cs.didReset
	cs.didReset = true
	cc.mu.Unlock()

	if !didReset {
		cc.writeStreamReset(cs.ID, ErrCodeCancel, nil)
		cc.forgetStreamID(cs.ID)
	}
}

// checkResetOrDone reports any error sent in a RST_STREAM frame by the
// server, or errStreamClosed if the stream is complete.
func (cs *clientStream) checkResetOrDone() error {
	select {
	case <-cs.peerReset:
		return cs.resetErr
	case <-cs.done:
		return errStreamClosed
	default:
		return nil
	}
}