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
			if r.wantSep != 0 {
				r.errSyntax()
			}

			r.token.kind = tokenBool
			r.token.boolValue = false
			r.fetchFalse()
			return

		default:
			r.errSyntax()
			return
		}
	}
	r.fatalError = io.EOF
	return
}

// isTokenEnd returns true if the char can follow a non-delimiter token
func isTokenEnd(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '[' || c == ']' || c == '{' || c == '}' || c == ',' || c == ':'
}

// fetchNull fetches and checks remaining bytes of null keyword.
func (r *Lexer) fetchNull() {
	r.pos += 4
	if r.pos > len(r.Data) ||
		r.Data[r.pos-3] != 'u' ||
		r.Data[r.pos-2] != 'l' ||
		r.Data[r.pos-1] != 'l' ||
		(r.pos != len(r.Data) && !isTokenEnd(r.Data[r.pos])) {

		r.pos -= 4
		r.errSyntax()
	}
}

// fetchTrue fetches and checks remaining bytes of true keyword.
func (r *Lexer) fetchTrue() {
	r.pos += 4
	if r.pos > len(r.Data) ||
		r.Data[r.pos-3] != 'r' ||
		r.Data[r.pos-2] != 'u' ||
		r.Data[r.pos-1] != 'e' ||
		(r.pos != len(r.Data) && !isTokenEnd(r.Data[r.pos])) {

		r.pos -= 4
		r.errSyntax()
	}
}

// fetchFalse fetches and checks remaining bytes of false keyword.
func (r *Lexer) fetchFalse() {
	r.pos += 5
	if r.pos > len(r.Data) ||
		r.Data[r.pos-4] != 'a' ||
		r.Data[r.pos-3] != 'l' ||
		r.Data[r.pos-2] != 's' ||
		r.Data[r.pos-1] != 'e' ||
		(r.pos != len(r.Data) && !isTokenEnd(r.Data[r.pos])) {

		r.pos -= 5
		r.errSyntax()
	}
}

// fetchNumber scans a number literal token.
func (r *Lexer) fetchNumber() {
	hasE := false
	afterE := false
	hasDot := false

	r.pos++
	for i, c := range r.Data[r.pos:] {
		switch {
		case c >= '0' && c <= '9':
			afterE = false
		case c == '.' && !hasDot:
			hasDot = true
		case (c == 'e' || c == 'E') && !hasE:
			hasE = true
			hasDot = true
			afterE = true
		case (c == '+' || c == '-') && afterE:
			afterE = false
		default:
			r.pos += i
			if !isTokenEnd(c) {
				r.errSyntax()
			} else {
				r.token.byteValue = r.Data[r.start:r.pos]
			}
			return
		}
	}

	r.pos = len(r.Data)
	r.token.byteValue = r.Data[r.start:]
}

// findStringLen tries to scan into the string literal for ending quote char to determine required size.
// The size will be exact if no escapes are present and may be inexact if there are escaped chars.
func findStringLen(data []byte) (hasEscapes bool, length int) {
	delta := 0

	for i := 0; i < len(data); i++ {
		switch data[i] {
		case '\\':
			i++
			delta++
			if i < len(data) && data[i] == 'u' {
				delta++
			}
		case '"':
			return (delta > 0), (i - delta)
		}
	}

	return false, len(data)
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	var val rune
	for i := 2; i < len(s) && i < 6; i++ {
		var v byte
		c := s[i]
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			v = c - '0'
		case 'a', 'b', 'c', 'd', 'e', 'f':
			v = c - 'a' + 10
		case 'A', 'B', 'C', 'D', 'E', 'F':
			v = c - 'A' + 10
		default:
			return -1
		}

		val <<= 4
		val |= rune(v)
	}
	return val
}

// processEscape processes a single escape sequence and returns number of bytes processed.
func (r *Lexer) processEscape(data []byte) (int, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("syntax error at %v", string(data))
	}

	c := data[1]
	switch c {
	case '"', '/', '\\':
		r.token.byteValue = append(r.token.byteValue, c)
		return 2, nil
	case 'b':
		r.token.byteValue = append(r.token.byteValue, '\b')
		return 2, nil
	case 'f':
		r.token.byteValue = append(r.token.byteValue, '\f')
		return 2, nil
	case 'n':
		r.token.byteValue = append(r.token.byteValue, '\n')
		return 2, nil
	case 'r':
		r.token.byteValue = append(r.token.byteValue, '\r')
		return 2, nil
	case 't':
		r.token.byteValue = append(r.token.byteValue, '\t')
		return 2, nil
	case 'u':
		rr := getu4(data)
		if rr < 0 {
			return 0, errors.New("syntax error")
		}

		read := 6
		if utf16.IsSurrogate(rr) {
			rr1 := getu4(data[read:])
			if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
				read += 6
				rr = dec
			} else {
				rr = unicode.ReplacementChar
			}
		}
		var d [4]byte
		s := utf8.EncodeRune(d[:], rr)
		r.token.byteValue = append(r.token.byteValue, d[:s]...)
		return read, nil
	}

	return 0, errors.New("syntax error")
}

// fetchString scans a string literal token.
func (r *Lexer) fetchString() {
	r.pos++
	data := r.Data[r.pos:]

	hasEscapes, length := findStringLen(data)
	if !hasEscapes {
		r.token.byteValue = data[:length]
		r.pos += length + 1
		return
	}

	r.token.byteValue = make([]byte, 0, length)
	p := 0
	for i := 0; i < len(data); {
		switch data[i] {
		case '"':
			r.pos += i + 1
			r.token.byteValue = append(r.token.byteValue, data[p:i]...)
			i++
			return

		case '\\':
			r.token.byteValue = append(r.token.byteValue, data[p:i]...)
			off, err := r.processEscape(data[i:])
			if err != nil {
				r.errParse(err.Error())
				return
			}
			i += off
			p = i

		default:
			i++
		}
	}
	r.errParse("unterminated string literal")
}

// scanToken scans the next token if no token is currently available in the lexer.
func (r *Lexer) scanToken() {
	if r.token.kind != tokenUndef || r.fatalError != nil {
		return
	}

	r.FetchToken()
}

// consume resets the current token to allow scanning the next one.
func (r *Lexer) consume() {
	r.token.kind = tokenUndef
	r.token.delimValue = 0
}

// Ok returns true if no error (including io.EOF) was encountered during scanning.
func (r *Lexer) Ok() bool {
	return r.fatalError == nil
}

const maxErrorContextLen = 13

func (r *Lexer) errParse(what string) {
	if r.fatalError == nil {
		var str string
		if len(r.Data)-r.pos <= maxErrorContextLen {
			str = string(r.Data)
		} else {
			str = string(r.Data[r.pos:r.pos+maxErrorContextLen-3]) + "..."
		}
		r.fatalError = &LexerError{
			Reason: what,
			Offset: r.pos,
			Data:   str,
		}
	}
}

func (r *Lexer) errSyntax() {
	r.errParse("syntax error")
}

func (r *Lexer) errInvalidToken(expected string) {
	if r.fatalError != nil {
		return
	}
	if r.UseMultipleErrors {
		r.pos = r.start
		r.consume()
		r.SkipRecursive()
		switch expected {
		case "[":
			r.token.delimValue = ']'
			r.token.kind = tokenDelim
		case "{":
			r.token.delimValue = '}'
			r.token.kind = tokenDelim
		}
		r.addNonfatalError(&LexerError{
			Reason: fmt.Sprintf("expected %s", expected),
			Offset: r.start,
			Data:   string(r.Data[r.start:r.pos]),
		})
		return
	}

	var str string
	if len(r.token.byteValue) <= maxErrorContextLen {
		str = string(r.token.byteValue)
	} else {
		str = string(r.token.byteValue[:maxErrorContextLen-3]) + "..."
	}
	r.fatalError = &LexerError{
		Reason: fmt.Sprintf("expected %s", expected),
		Offset: r.pos,
		Data:   str,
	}
}

func (r *Lexer) GetPos() int {
	return r.pos
}

// Delim consumes a token and verifies that it is the given delimiter.
func (r *Lexer) Delim(c byte) {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() || r.token.delimValue != c {
		r.consume() // errInvali