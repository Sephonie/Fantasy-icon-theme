// Package jlexer contains a JSON lexer implementation.
//
// It is expected that it is mostly used with generated parser code, so the interface is tuned
// for a parser that knows what kind of data is expected.
package jlexer

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// tokenKind determines type of a token.
type tokenKind byte

const (
	tokenUndef  tokenKind = iota // No token.
	tokenDelim                   // Delimiter: one of '{', '}', '[' or ']'.
	tokenString                  // A string literal, e.g. "abc\u1234"
	tokenNumber                  // Number literal, e.g. 1.5e5
	tokenBool                    // Boolean literal: true or false.
	tokenNull                    // null keyword.
)

// token describes a single token: type, position in the input and value.
type token struct {
	kind tokenKind // Type of a token.

	boolValue  bool   // Value if a boolean literal token.
	byteValue  []byte // Raw value of a token.
	delimValue byte
}

// Lexer is a JSON lexer: it iterates over JSON tokens in a byte slice.
type Lexer struct {
	Data []byte // Input data given to the lexer.

	start int   // Start of the current token.
	pos   int   // Current unscanned position in the input stream.
	token token // Last scanned token, if token.kind != tokenUndef.

	firstElement bool // Whether current element is the first in array or an object.
	wantSep      byte // A comma or a colon character, which need to occur before a token.

	UseMultipleErrors bool          // If we want to use multiple errors.
	fatalError        error         // Fatal error occurred during lexing. It is usually a syntax error.
	multipleErrors    []*LexerError // Semantic errors occurred during lexing. Marshalling will be continued after finding this errors.
}

// FetchToken scans the input for the next token.
func (r *Lexer) FetchToken() {
	r.token.kind = tokenUndef
	r.start = r.pos

	// Check if r.Data has r.pos element
	// If it doesn't, it mean corrupted input data
	if len(r.Data) < r.pos {
		r.errParse("Unexpected end of data")
		return
	}
	// Determine the type of a token by skipping whitespace and reading the
	// first character.
	for _, c := range r.Data[r.pos:] {
		switch c {
		case ':', ',':
			if r.wantSep == c {
				r.pos++
				r.start++
				r.wantSep = 0
			} else {
				r.errSyntax()
			}

		case ' ', '\t', '\r', '\n':
			r.pos++
			r.start++

		case '"':
			if r.wantSep != 0 {
				r.errSyntax()
			}

			r.token.kind = tokenString
			r.fetchString()
			return

		case '{', '[':
			if r.wantSep != 0 {
				r.errSyntax()
			}
			r.firstElement = true
			r.token.kind = tokenDelim
			r.token.delimValue = r.Data[r.pos]
			r.pos++
			return

		case '}', ']':
			if !r.firstElement && (r.wantSep != ',') {
				r.errSyntax()
			}
			r.wantSep = 0
			r.token.kind = tokenDelim
			r.token.delimValue = r.Data[r.pos]
			r.pos++
			return

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			if r.wantSep != 0 {
				r.errSyntax()
			}
			r.token.kind = tokenNumber
			r.fetchNumber()
			return

		case 'n':
			if r.wantSep != 0 {
				r.errSyntax()
			}

			r.token.kind = tokenNull
			r.fetchNull()
			return

		case 't':
			if r.wantSep != 0 {
				r.errSyntax()
			}

			r.token.kind = tokenBool
			r.token.boolValue = true
			r.fetchTrue()
			return

		case 'f':
			if r.wantSep !=