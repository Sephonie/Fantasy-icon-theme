// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

import (
	"unicode/utf8"

	"golang.org/x/text/transform"
)

// Reset implements the Reset method of the transform.Transformer interface.
func (Form) Reset() {}

// Transform implements the Transform method of the transform.Transformer
// interface. It may need to write segments of up to MaxSegmentSize at once.
// Users should either catch ErrShortDst and allow dst to grow or have dst be at
// least of size MaxTransformChunkSize to be guaranteed of progress.
func (f Form) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	n := 0
	// Cap the maximum number of src bytes to check.
	b := src
	eof := atEOF
	if ns := len(dst); ns < len(b) {
		err = transform.ErrShortDst
		eof = false
		b = b[:ns]
	}
	i, ok := formTable[f].quickSpan(inputBytes(b), n, len(b), eof)
	n += copy(dst[n:], b[n:i])
	if !ok {
		nDst, nSrc, err = f.transform(dst[n:], src[n:], atEOF)
		return nDst + n, nSrc + n, err
	}
	if n < len(src) && !atEOF {
		err = transform.ErrShortSrc
	}
	return n, n, err
}

func flushTransform(rb *reorderBuffer) bool {
	// Write out (must fully fit in dst, or else it is an ErrShortDst).
	if len(rb.out) < rb.nrune*utf8.UTFMax {
		return false
	}
	rb.out = rb.out[rb.flushCopy(rb.out):]
	return true
}

var errs = []error{nil, transform.ErrShortDst, transform.ErrShortSrc}

// transform implements the transform.Transfor