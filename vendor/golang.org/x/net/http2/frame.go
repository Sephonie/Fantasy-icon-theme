// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"golang.org/x/net/http2/hpack"
	"golang.org/x/net/lex/httplex"
)

const frameHeaderLen = 9

var padZeros = make([]byte, 255) // zeros for padding

// A FrameType is a registered frame type as defined in
// http://http2.github.io/http2-spec/#rfc.section.11.2
type FrameType uint8

const (
	FrameData         FrameType = 0x0
	FrameHeaders      FrameType = 0x1
	FramePriority     FrameType = 0x2
	FrameRSTStream    FrameType = 0x3
	FrameSettings     FrameType = 0x4
	FramePushPromise  FrameType = 0x5
	FramePing         FrameType = 0x6
	FrameGoAway       FrameType = 0x7
	FrameWindowUpdate FrameType = 0x8
	FrameContinuation FrameType = 0x9
)

var frameName = map[FrameType]string{
	FrameData:         "DATA",
	FrameHeaders:      "HEADERS",
	FramePriority:     "PRIORITY",
	FrameRSTStream:    "RST_STREAM",
	FrameSettings:     "SETTINGS",
	FramePushPromise:  "PUSH_PROMISE",
	FramePing:         "PING",
	FrameGoAway:       "GOAWAY",
	FrameWindowUpdate: "WINDOW_UPDATE",
	FrameContinuation: "CONTINUATION",
}

func (t FrameType) String() string {
	if s, ok := frameName[t]; ok {
		return s
	}
	return fmt.Sprintf("UNKNOWN_FRAME_TYPE_%d", uint8(t))
}

// Flags is a bitmask of HTTP/2 flags.
// The meaning of flags varies depending on the frame type.
type Flags uint8

// Has reports whether f contains all (0 or more) flags in v.
func (f Flags) Has(v Flags) bool {
	return (f & v) == v
}

// Frame-specific FrameHeader flag bits.
const (
	// Data Frame
	FlagDataEndStream Flags = 0x1
	FlagDataPadded    Flags = 0x8

	// Headers Frame
	FlagHeadersEndStream  Flags = 0x1
	FlagHeadersEndHeaders Flags = 0x4
	FlagHeadersPadded     Flags = 0x8
	FlagHeadersPriority   Flags = 0x20

	// Settings Frame
	FlagSettingsAck Flags = 0x1

	// Ping Frame
	FlagPingAck Flags = 0x1

	// Continuation Frame
	FlagContinuationEndHeaders Flags = 0x4

	FlagPushPromiseEndHeaders Flags = 0x4
	FlagPushPromisePadded     Flags = 0x8
)

var flagName = map[FrameType]map[Flags]string{
	FrameData: {
		FlagDataEndStream: "END_STREAM",
		FlagDataPadded:    "PADDED",
	},
	FrameHeaders: {
		FlagHeadersEndStream:  "END_STREAM",
		FlagHeadersEndHeaders: "END_HEADERS",
		FlagHeadersPadded:     "PADDED",
		FlagHeadersPriority:   "PRIORITY",
	},
	FrameSettings: {
		FlagSettingsAck: "ACK",
	},
	FramePing: {
		FlagPingAck: "ACK",
	},
	FrameContinuation: {
		FlagContinuationEndHeaders: "END_HEADERS",
	},
	FramePushPromise: {
		FlagPushPromiseEndHeaders: "END_HEADERS",
		FlagPushPromisePadded:     "PADDED",
	},
}

// a frameParser parses a frame given its FrameHeader and payload
// bytes. The length of payload will always equal fh.Length (which
// might be 0).
type frameParser func(fc *frameCache, fh FrameHeader, payload []byte) (Frame, error)

var frameParsers = map[FrameType]frameParser{
	FrameData:         parseDataFrame,
	FrameHeaders:      parseHeadersFrame,
	FramePriority:     parsePriorityFrame,
	FrameRSTStream:    parseRSTStreamFrame,
	FrameSettings:     parseSettingsFrame,
	FramePushPromise:  parsePushPromise,
	FramePing:         parsePingFrame,
	FrameGoAway:       parseGoAwayFrame,
	FrameWindowUpdate: parseWindowUpdateFrame,
	FrameContinuation: parseContinuationFrame,
}

func typeFrameParser(t FrameType) frameParser {
	if f := frameParsers[t]; f != nil {
		return f
	}
	return parseUnknownFrame
}

// A FrameHeader is the 9 byte header of all HTTP/2 frames.
//
// See http://http2.github.io/http2-spec/#FrameHeader
type FrameHeader struct {
	valid bool // caller can access []byte fields in the Frame

	// Type is the 1 byte frame type. There are ten standard frame
	// types, but extension frame types may be written by WriteRawFrame
	// and will be returned by ReadFrame (as UnknownFrame).
	Type FrameType

	// Flags are the 1 byte of 8 potential bit flags per frame.
	// They are specific to the frame type.
	Flags Flags

	// Length is the length of the frame, not including the 9 byte header.
	// The maximum size is one byte less than 16MB (uint24), but only
	// frames up to 16KB are allowed without peer agreement.
	Length uint32

	// StreamID is which stream this frame is for. Certain frames
	// are not stream-specific, in which case this field is 0.
	StreamID uint32
}

// Header returns h. It exists so FrameHeaders can be embedded in other
// specific frame types and implement the Frame interface.
func (h FrameHeader) Header() FrameHeader { return h }

func (h FrameHeader) String() string {
	var buf bytes.Buffer
	buf.WriteString("[FrameHeader ")
	h.writeDebug(&buf)
	buf.WriteByte(']')
	return buf.String()
}

func (h FrameHeader) writeDebug(buf *bytes.Buffer) {
	buf.WriteString(h.Type.String())
	if h.Flags != 0 {
		buf.WriteString(" flags=")
		set := 0
		for i := uint8(0); i < 8; i++ {
			if h.Flags&(1<<i) == 0 {
				continue
			}
			set++
			if set > 1 {
				buf.WriteByte('|')
			}
			name := flagName[h.Type][Flags(1<<i)]
			if name != "" {
				buf.WriteString(name)
			} else {
				fmt.Fprintf(buf, "0x%x", 1<<i)
			}
		}
	}
	if h.StreamID != 0 {
		fmt.Fprintf(buf, " stream=%d", h.StreamID)
	}
	fmt.Fprintf(buf, " len=%d", h.Length)
}

func (h *FrameHeader) checkValid() {
	if !h.valid {
		panic("Frame accessor called on non-owned Frame")
	}
}

func (h *FrameHeader) invalidate() { h.valid = false }

// frame header bytes.
// Used only by ReadFrameHeader.
var fhBytes = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, frameHeaderLen)
		return &buf
	},
}

// ReadFrameHeader reads 9 bytes from r and returns a FrameHeader.
// Most users should use Framer.ReadFrame instead.
func ReadFrameHeader(r io.Reader) (FrameHeader, error) {
	bufp := fhBytes.Get().(*[]byte)
	defer fhBytes.Put(bufp)
	return readFrameHeader(*bufp, r)
}

func readFrameHeader(buf []byte, r io.Reader) (FrameHeader, error) {
	_, err := io.ReadFull(r, buf[:frameHeaderLen])
	if err != nil {
		return FrameHeader{}, err
	}
	return FrameHeader{
		Length:   (uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2])),
		Type:     FrameType(buf[3]),
		Flags:    Flags(buf[4]),
		StreamID: binary.BigEndian.Uint32(buf[5:]) & (1<<31 - 1),
		valid:    true,
	}, nil
}

// A Frame is the base interface implemented by all frame types.
// Callers will generally type-assert the specific frame type:
// *HeadersFrame, *SettingsFrame, *WindowUpdateFrame, etc.
//
// Frames are only valid until the next call to Framer.ReadFrame.
type Frame interface {
	Header() FrameHeader

	// invalidate is called by Framer.ReadFrame to make this
	// frame's buffers as being invalid, since the subsequent
	// frame will reuse them.
	invalidate()
}

// A Framer reads and writes Frames.
type Framer struct {
	r         io.Reader
	lastFrame Frame
	errDetail error

	// lastHeaderStream is non-zero if the last frame was an
	// unfinished HEADERS/CONTINUATION.
	lastHeaderStream uint32

	maxReadSize uint32
	headerBuf   [frameHeaderLen]byte

	// TODO: let getReadBuf be configurable, and use a less memory-pinning
	// allocator in server.go to minimize memory pinned for many idle conns.
	// Will probably also need to make frame invalidation have a hook too.
	getReadBuf func(size uint32) []byte
	readBuf    []byte // cache for default getReadBuf

	maxWriteSize uint32 // zero means unlimited; TODO: implement

	w    io.Writer
	wbuf []byte

	// AllowIllegalWrites permits the Framer's Write methods to
	// write frames that do not conform to the HTTP/2 spec. This
	// permits using the Framer to test other HTTP/2
	// implementations' conformance to the spec.
	// If false, the Write methods will prefer to return an error
	// rather than comply.
	AllowIllegalWrites bool

	// AllowIllegalReads permits the Framer's ReadFrame method
	// to return non-compliant frames or frame orders.
	// This is for testing and permits using the Framer to test
	// other HTTP/2 implementations' conformance to the spec.
	// It is not compatible with ReadMetaHeaders.
	AllowIllegalReads bool

	// ReadMetaHeaders if non-nil causes ReadFrame to merge
	// HEADERS and CONTINUATION frames together and return
	// MetaHeadersFrame instead.
	ReadMetaHeaders *hpack.Decoder

	// MaxHeaderListSize is the http2 MAX_HEADER_LIST_SIZE.
	// It's used only if ReadMetaHeaders is set; 0 means a sane default
	// (currently 16MB)
	// If the limit is hit, MetaHeadersFrame.Truncated is set true.
	MaxHeaderListSize uint32

	// TODO: track which type of frame & with which flags was sent
	// last. Then return an error (unless AllowIllegalWrites) if
	// we're in the middle of a header block and a
	// non-Continuation or Continuation on a different stream is
	// attempted to be written.

	logReads, logWrites bool

	debugFramer       *Framer // only use for logging written writes
	debugFramerBuf    *bytes.Buffer
	debugReadLoggerf  func(string, ...interface{})
	debugWriteLoggerf func(string, ...interface{})

	frameCache *frameCache // nil if frames aren't reused (default)
}

func (fr *Framer) maxHeaderListSize() uint32 {
	if fr.MaxHeaderListSize == 0 {
		return 16 << 20 // sane default, per docs
	}
	return fr.MaxHeaderListSize
}

func (f *Framer) startWrite(ftype FrameType, flags Flags, streamID uint32) {
	// Write the FrameHeader.
	f.wbuf = append(f.wbuf[:0],
		0, // 3 bytes of length, filled in in endWrite
		0,
		0,
		byte(ftype),
		byte(flags),
		byte(streamID>>24),
		byte(streamID>>16),
		byte(streamID>>8),
		byte(streamID))
}

func (f *Framer) endWrite() error {
	// Now that we know the final size, fill in the FrameHeader in
	// the space previously reserved for it. Abuse append.
	length := len(f.wbuf) - frameHeaderLen
	if length >= (1 << 24) {
		return ErrFrameTooLarge
	}
	_ = append(f.wbuf[:0],
		byte(length>>16),
		byte(length>>8),
		byte(length))
	if f.logWrites {
		f.logWrite()
	}

	n, err := f.w.Write(f.wbuf)
	if err == nil && n != len(f.wbuf) {
		err = io.ErrShortWrite
	}
	return err
}

func (f *Framer) logWrite() {
	if f.debugFramer == nil {
		f.debugFramerBuf = new(bytes.Buffer)
		f.debugFramer = NewFramer(nil, f.debugFramerBuf)
		f.debugFramer.logReads = false // we log it ourselves, saying "wrote" below
		// Let us read anything, even if we accidentally wrote it
		// in the wrong order:
		f.debugFramer.AllowIllegalReads = true
	}
	f.debugFramerBuf.Write(f.wbuf)
	fr, err := f.debugFramer.ReadFrame()
	if err != nil {
		f.debugWriteLoggerf("http2: Framer %p: failed to decode just-written frame", f)
		return
	}
	f.debugWriteLoggerf("http2: Framer %p: wrote %v", f, summarizeFrame(fr))
}

func (f *Framer) writeByte(v byte)     { f.wbuf = append(f.wbuf, v) }
func (f *Framer) writeBytes(v []byte)  { f.wbuf = append(f.wbuf, v...) }
func (f *Framer) writeUint16(v uint16) { f.wbuf = append(f.wbuf, byte(v>>8), byte(v)) }
func (f *Framer) writeUint32(v uint32) {
	f.wbuf = append(f.wbuf, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

const (
	minMaxFrameSize = 1 << 14
	maxFrameSize    = 1<<24 - 1
)

// SetReuseFrames allows the Framer to reuse Frames.
// If called on a Framer, Frames returned by calls to ReadFrame are only
// valid until the next call to ReadFrame.
func (fr *Framer) SetReuseFrames() {
	if fr.frameCache != nil {
		return
	}
	fr.frameCache = &frameCache{}
}

type frameCache struct {
	dataFrame DataFrame
}

func (fc *frameCache) getDataFrame() *DataFrame {
	if fc == nil {
		return &DataFrame{}
	}
	return &fc.dataFrame
}

// NewFramer returns a Framer that writes frames to w and reads them from r.
func NewFramer(w io.Writer, r io.Reader) *Framer {
	fr := &Framer{
		w:                 w,
		r:                 r,
		logReads:          logFrameReads,
		logWrites:         logFrameWrites,
		debugReadLoggerf:  log.Printf,
		debugWriteLoggerf: log.Printf,
	}
	fr.getReadBuf = func(size uint32) []byte {
		if cap(fr.readBuf) >= int(size) {
			return fr.readBuf[:size]
		}
		fr.readBuf = make([]byte, size)
		return fr.readBuf
	}
	fr.SetMaxReadFrameSize(maxFrameSize)
	return fr
}

// SetMaxReadFrameSize sets the maximum size of a frame
// that will be read by a subsequent call to ReadFrame.
// It is the caller's responsibility to advertise this
// limit with a SETTINGS frame.
func (fr *Framer) SetMaxReadFrameSize(v uint32) {
	if v > maxFrameSize {
		v = maxFrameSize
	}
	fr.maxReadSize = v
}

// ErrorDetail returns a more detailed error of the last error
// returned by Framer.ReadFrame. For instance, if ReadFrame
// returns a StreamError with code PROTOCOL_ERROR, ErrorDetail
// will say exactly what was invalid. ErrorDetail is not guaranteed
// to return a non-nil value and like the rest of the http2 package,
// its return value is not protected by an API compatibility promise.
// ErrorDetail is reset after the next call to ReadFrame.
func (fr *Framer) ErrorDetail() error {
	return fr.errDetail
}

// ErrFrameTooLarge is returned from Framer.ReadFrame when the peer
// sends a frame that is larger than declared with SetMaxReadFrameSize.
var ErrFrameTooLarge = errors.New("http2: frame too large")

// terminalReadFrameError reports whether err is an unrecoverable
// error from ReadFrame and no other frames should be read.
func terminalReadFrameError(err error) bool {
	if _, ok := err.(StreamError); ok {
		return false
	}
	return err != nil
}

// ReadFrame reads a single frame. The returned Frame is only valid
// until the next call to ReadFrame.
//
// If the frame is larger than previously set with SetMaxReadFrameSize, the
// returned error is ErrFrameTooLarge. Other errors may be of type
// ConnectionError, StreamError, or anything else from the underlying
// reader.
func (fr *Framer) ReadFrame() (Frame, error) {
	fr.errDetail = nil
	if fr.lastFrame != nil {
		fr.lastFrame.invalidate()
	}
	fh, err := readFrameHeader(fr.headerBuf[:], fr.r)
	if err != nil {
		return nil, err
	}
	if fh.Length > fr.maxReadSize {
		return nil, ErrFrameTooLarge
	}
	payload := fr.getReadBuf(fh.Length)
	if _, err := io.ReadFull(fr.r, payload); err != nil {
		return nil, err
	}
	f, err := typeFrameParser(fh.Type)(fr.frameCache, fh, payload)
	if err != nil {
		if ce, ok := err.(connError); ok {
			return nil, fr.connError(ce.Code, ce.Reason)
		}
		return nil, err
	}
	if err := fr.checkFrameOrder(f); err != nil {
		return nil, err
	}
	if fr.logReads {
		fr.debugReadLoggerf("http2: Framer %p: read %v", fr, summarizeFrame(f))
	}
	if fh.Type == FrameHeaders && fr.ReadMetaHeaders != nil {
		return fr.readMetaFrame(f.(*HeadersFrame))
	}
	return f, nil
}

// connError returns ConnectionError(code) but first
// stashes away a public reason to the caller can optionally relay it
// to the peer before hanging up on them. This might help others debug
// their implementations.
func (fr *Framer) connError(code ErrCode, reason string) error {
	fr.errDetail = errors.New(reason)
	return ConnectionError(code)
}

// checkFrameOrder reports an error if f is an invalid frame to return
// next from ReadFrame. Mostly it checks whether HEADERS and
// CONTINUATION frames are contiguous.
func (fr *Framer) checkFrameOrder(f Frame) error {
	last := fr.lastFrame
	fr.lastFrame = f
	if fr.AllowIllegalReads {
		return nil
	}

	fh := f.Header()
	if fr.lastHeaderStream != 0 {
		if fh.Type != FrameContinuation {
			return fr.connError(ErrCodeProtocol,
				fmt.Sprintf("got %s for stream %d; expected CONTINUATION following %s for stream %d",
					fh.Type, fh.StreamID,
					last.Header().Type, fr.lastHeaderStream))
		}
		if fh.StreamID != fr.lastHeaderStream {
			return fr.connError(ErrCodeProtocol,
				fmt.Sprintf("got CONTINUATION for stream %d; expected stream %d",
					fh.StreamID, fr.lastHeaderStream))
		}
	} else if fh.Type == FrameContinuation {
		return fr.connError(ErrCodeProtocol, fmt.Sprintf("unexpected CONTINUATION for stream %d", fh.StreamID))
	}

	switch fh.Type {
	case FrameHeaders, FrameContinuation:
		if fh.Flags.Has(FlagHeadersEndHeaders) {
			fr.lastHeaderStream = 0
		} else {
			fr.lastHeaderStream = fh.StreamID
		}
	}

	return nil
}

// A DataFrame conveys arbitrary, variable-length sequences of octets
// associated with a stream.
// See http://http2.github.io/http2-spec/#rfc.section.6.1
type DataFrame struct {
	FrameHeader
	data []byte
}

func (f *DataFrame) StreamEnded() bool {
	return f.FrameHeader.Flags.Has(FlagDataEndStream)
}

// Data returns the frame's data octets, not including any padding
// size byte or padding suffix bytes.
// The caller must not retain the returned memory past the next
// call to ReadFrame.
func (f *DataFrame) Data() []byte {
	f.checkValid()
	return f.data
}

func parseDataFrame(fc *frameCache, fh FrameHeader, payload []byte) (Frame, error) {
	if fh.StreamID == 0 {
		// DATA frames MUST be associated with a stream. If a
		// DATA frame is received whose stream identifier
		// field is 0x0, the recipient MUST respond with a
		// connection error (Section 5.4.1) of type
		// PROTOCOL_ERROR.
		return nil, connError{ErrCodeProtocol, "DATA frame with stream ID 0"}
	}
	f := fc.getDataFrame()
	f.FrameHeader = fh

	var padSize byte
	if fh.Flags.Has(FlagDataPadded) {
		var err error
		payload, padSize, err = readByte(payload)
		if err != nil {
			return nil, err
		}
	}
	if int(padSize) > len(payload) {
		// If the length of the padding is greater than the
		// length of the frame payload, the recipient MUST
		// treat this as a connection error.
		// Filed: https://github.com/http2/http2-spec/issues/610
		return nil, connError{ErrCodeProtocol, "pad size larger than data payload"}
	}
	f.data = payload[:len(payload)-int(padSize)]
	return f, nil
}

var (
	errStreamID    = errors.New("invalid stream ID")
	errDepStreamID = errors.New("invalid dependent stream ID")
	errPadLength   = errors.New("pad length too large")
	errPadBytes    = errors.New("padding bytes must all be zeros unless AllowIllegalWrites is enabled")
)

func validStreamIDOrZero(streamID uint32) bool {
	return streamID&(1<<31) == 0
}

func validStreamID(streamID uint32) bool {
	return streamID != 0 && streamID&(1<<31) == 0
}

// WriteData writes a DATA frame.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility not to violate the maximum frame size
// and to not call other Write methods concurrently.
func (f *Framer) WriteData(streamID uint32, endStream bool, data []byte) error {
	return f.WriteDataPadded(streamID, endStream, data, nil)
}

// WriteData writes a DATA frame with optional padding.
//
// If pad is nil, the padding bit is not sent.
// The length of pad must not exceed 255 bytes.
// The bytes of pad must all be zero, unless f.AllowIllegalWrites is set.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility not to violate the maximum frame size
// and to not call other Write methods concurrently.
func (f *Framer) WriteDataPadded(streamID uint32, endStream bool, data, pad []byte) error {
	if !validStreamID(streamID) && !f.AllowIllegalWrites {
		return errStreamID
	}
	if len(pad) > 0 {
		if len(pad) > 255 {
			return errPadLength
		}
		if !f.AllowIllegalWrites {
			for _, b := range pad {
				if b != 0 {
					// "Padding octets MUST be set to zero when sending."
					return errPadBytes
				}
			}
		}
	}
	var flags Flags
	if endStream {
		flags |= FlagDataEndStream
	}
	if pad != nil {
		flags |= FlagDataPadded
	}
	f.startWrite(FrameData, flags, streamID)
	if pad != nil {
		f.wbuf = append(f.wbuf, byte(len(pad)))
	}
	f.wbuf = append(f.wbuf, data...)
	f.wbuf = append(f.wbuf, pad...)
	return f.endWrite()
}

// A SettingsFrame conveys configuration parameters that affect how
// endpoints communicate, such as preferences and constraints on peer
// behavior.
//
// See http://http2.github.io/http2-spec/#SETTINGS
type SettingsFrame struct {
	FrameHeader
	p []byte
}

func parseSettingsFrame(_ *frameCache, fh FrameHeader, p []byte) (Frame, error) {
	if fh.Flags.Has(FlagSettingsAck) && fh.Length > 0 {
		// When this (ACK 0x1) bit is set, the payload of the
		// SETTINGS frame MUST be empty. Receipt of a
		// SETTINGS frame with the ACK flag set and a length
		// field value other than 0 MUST be treated as a
		// connection error (Section 5.4.1) of type
		// FRAME_SIZE_ERROR.
		return nil, ConnectionError(ErrCodeFrameSize)
	}
	if fh.StreamID != 0 {
		// SETTINGS frames always apply to a connection,
		// never a single stream. The stream identifier for a
		// SETTINGS frame MUST be zero (0x0).  If an endpoint
		// receives a SETTINGS frame whose stream identifier
		// field is anything other than 0x0, the endpoint MUST
		// respond with a connection error (Section 5.4.1) of
		// type PROTOCOL_ERROR.
		return nil, ConnectionError(ErrCodeProtocol)
	}
	if len(p)%6 != 0 {
		// Expecting even number of 6 byte settings.
		return nil, ConnectionError(ErrCodeFrameSize)
	}
	f := &SettingsFrame{FrameHeader: fh, p: p}
	if v, ok := f.Value(SettingInitialWindowSize); ok && v > (1<<31)-1 {
		// Values above the maximum flow control window size of 2^31 - 1 MUST
		// be treated as a connection error (Section 5.4.1) of type
		// FLOW_CONTROL_ERROR.
		return nil, ConnectionError(ErrCodeFlowControl)
	}
	return f, nil
}

func (f *SettingsFrame) IsAck() bool {
	return f.FrameHeader.Flags.Has(FlagSettingsAck)
}

func (f *SettingsFrame) Value(s SettingID) (v uint32, ok bool) {
	f.checkValid()
	buf := f.p
	for len(buf) > 0 {
		settingID := SettingID(binary.BigEndian.Uint16(buf[:2]))
		if settingID == s {
			return binary.BigEndian.Uint32(buf[2:6]), true
		}
		buf = buf[6:]
	}
	return 0, false
}

// ForeachSetting runs fn for each setting.
// It stops and returns the first error.
func (f *SettingsFrame) ForeachSetting(fn func(Setting) error) error {
	f.checkValid()
	buf := f.p
	for len(buf) > 0 {
		if err := fn(Setting{
			SettingID(binary.BigEndian.Uint16(buf[:2])),
			binary.BigEndian.Uint32(buf[2:6]),
		}); err != nil {
			return err
		}
		buf = buf[6:]
	}
	return nil
}

// WriteSettings writes a SETTINGS frame with zero or more settings
// specified and the ACK bit not set.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility to not call other Write methods concurrently.
func (f *Framer) WriteSettings(settings ...Setting) error {
	f.startWrite(FrameSettings, 0, 0)
	for _, s := range settings {
		f.writeUint16(uint16(s.ID))
		f.writeUint32(s.Val)
	}
	return f.endWrite()
}

// WriteSettingsAck writes an empty SETTINGS frame with the ACK bit set.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility to not call other Write methods concurrently.
func (f *Framer) WriteSettingsAck() error {
	f.startWrite(FrameSettings, FlagSettingsAck, 0)
	return f.endWrite()
}

// A PingFrame is a mechanism for measuring a minimal round trip time
// from the sender, as well as determining whether an idle connection
// is still functional.
// See http://http2.github.io/http2-spec/#rfc.section.6.7
type PingFrame struct {
	FrameHeader
	Data [8]byte
}

func (f *PingFrame) IsAck() bool { return f.Flags.Has(FlagPingAck) }

func parsePingFrame(_ *frameCache, fh FrameHeader, payload []byte) (Frame, error) {
	if len(payload) != 8 {
		return nil, ConnectionError(ErrCodeFrameSize)
	}
	if fh.StreamID != 0 {
		return nil, ConnectionError(ErrCodeProtocol)
	}
	f := &PingFrame{FrameHeader: fh}
	copy(f.Data[:], payload)
	return f, nil
}

func (f *Framer) WritePing(ack bool, data [8]byte) error {
	var flags Flags
	if ack {
		flags = FlagPingAck
	}
	f.startWrite(FramePing, flags, 0)
	f.writeBytes(data[:])
	return f.endWrite()
}

// A GoAwayFrame informs the remote peer to stop creating streams on this connection.
// See http://http2.github.io/http2-spec/#rfc.section.6.8
type GoAwayFrame struct {
	FrameHeader
	LastStreamID uint32
	ErrCode      ErrCode
	debugData    []byte
}

// DebugData returns any debug data in the GOAWAY frame. Its contents
// are not defined.
// The caller must not retain the returned memory past the next
// call to ReadFrame.
func (f *GoAwayFrame) DebugData() []byte {
	f.checkValid()
	return f.debugData
}

func parseGoAwayFrame(_ *frameCache, fh FrameHeader, p []byte) (Frame, error) {
	if fh.StreamID != 0 {
		return nil, ConnectionError(ErrCodeProtocol)
	}
	if len(p) < 8 {
		return nil, ConnectionError(ErrCodeFrameSize)
	}
	return &GoAwayFrame{
		FrameHeader:  fh,
		LastStreamID: binary.BigEndian.Uint32(p[:4]) & (1<<31 - 1),
		ErrCode:      ErrCode(binary.BigEndian.Uint32(p[4:8])),
		debugData:    p[8:],
	}, nil
}

func (f *Framer) WriteGoAway(maxStreamID uint32, code ErrCode, debugData []byte) error {
	f.startWrite(FrameGoAway, 0, 0)
	f.writeUint32(maxStreamID & (1<<31 - 1))
	f.writeUint32(uint32(code))
	f.writeBytes(debugData)
	return f.endWrite()
}

// An UnknownFrame is the frame type returned when the frame type is unknown
// or no specific frame type parser exists.
type UnknownFrame struct {
	FrameHeader
	p []byte
}

// Payload returns the frame's payload (after the header).  It is not
// valid to call this method after a subsequent call to
// Framer.ReadFrame, nor is it valid to retain the returned slice.
// The memory is owned by the Framer and is invalidated when the next
// frame is read.
func (f *UnknownFrame) Payload() []byte {
	f.checkValid()
	return f.p
}

func parseUnknownFrame(_ *frameCache, fh FrameHeader, p []byte) (Frame, error) {
	return &UnknownFrame{fh, p}, nil
}

// A WindowUpdateFrame is used to implement flow control.
// See http://http2.github.io/http2-spec/#rfc.section.6.9
type WindowUpdateFrame struct {
	FrameHeader
	Increment uint32 // never read with high bit set
}

func parseWindowUpdateFrame(_ *frameCache, fh FrameHeader, p []byte) (Frame, error) {
	if len(p) != 4 {
		return nil, ConnectionError(ErrCodeFrameSize)
	}
	inc := binary.BigEndian.Uint32(p[:4]) & 0x7fffffff // mask off high reserved bit
	if inc == 0 {
		// A receiver MUST treat the receipt of a
		// WINDOW_UPDATE frame with an flow control window
		// increment of 0 as a stream error (Section 5.4.2) of
		// type PROTOCOL_ERROR; errors on the connection flow
		// control window MUST be treated as a connection
		// error (Section 5.4.1).
		if fh.StreamID == 0 {
			return nil, ConnectionError(ErrCodeProtocol)
		}
		return nil, streamError(fh.StreamID, ErrCodeProtocol)
	}
	return &WindowUpdateFrame{
		FrameHeader: fh,
		Increment:   inc,
	}, nil
}

// WriteWindowUpdate writes a WINDOW_UPDATE frame.
// The increment value must be between 1 and 2,147,483,647, inclusive.
// If the Stream ID is zero, the window update applies to the
// connection as a whole.
func (f *Framer) WriteWindowUpdate(streamID, incr uint32) error {
	// "The legal range for the increment to the flow control window is 1 to 2^31-1 (2,147,483,647) octets."
	if (incr < 1 || incr > 2147483647) && !f.AllowIllegalWrites {
		return errors.New("illegal window increment value")
	}
	f.startWrite(FrameWindowUpdate, 0, streamID)
	f.writeUint32(incr)
	return f.endWrite()
}

// A HeadersFrame is used to open a stream and additionally carries a
// header block fragment.
type HeadersFrame struct {
	FrameHeader

	// Priority is set if FlagHeadersPriority is set in the FrameHeader.
	Priority PriorityParam

	heade