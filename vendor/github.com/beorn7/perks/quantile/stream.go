// Package quantile computes approximate quantiles over an unbounded data
// stream within low memory and CPU bounds.
//
// A small amount of accuracy is traded to achieve the above properties.
//
// Multiple streams can be merged before calling Query to generate a single set
// of results. This is meaningful when the streams represent the same type of
// data. See Merge and Samples.
//
// For more detailed information about the algorithm used, see:
//
// Effective Computation of Biased Quantiles over Data Streams
//
// http://www.cs.rutgers.edu/~muthu/bquant.pdf
package quantile

import (
	"math"
	"sort"
)

// Sample holds an observed value and meta information for compression. JSON
// tags have been added for convenience.
type Sample struct {
	Value float64 `json:",string"`
	Width float64 `json:",string"`
	Delta float64 `json:",string"`
}

// Samples represents a slice of samples. It implements sort.Interface.
type Samples []Sample

func (a Samples) Len() int           { return len(a) }
func (a Samples) Less(i, j int) bool { return a[i].Value < a[j].Value }
func (a Samples) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type invariant func(s *stream, r float64) float64

// NewLowBiased returns an initialized Stream for low-biased quantiles
// (e.g. 0.01, 0.1, 0.5) where the needed quantiles are not known a priori, but
// error guarantees can still be given even for the lower ranks of the data
// distribution.
//
// The provided epsilon is a relative error, i.e. the true quantile of a value
// returned by a query is guaranteed to be within (1±Epsilon)*Quantile.
//
// See http://www.cs.rutgers.edu/~muthu/bquant.pdf for time, space, and error
// properties.
func NewLowBiased(epsilon float64) *Stream {
	ƒ := func(s *stream, r float64) float64 {
		return 2 * epsilon * r
	}
	return newStream(ƒ)
}

// NewHighBiased returns an initialized Stream for high-biased quantiles
// (e.g. 0.01, 0.1, 0.5) where the needed quantiles are not known a priori, but
// error guarantees can still be given even for the higher ranks of the data
// distribution.
//
// The provided epsilon is a relative error, i.e. the true quantile of a value
// returned by a query is guaranteed to be within 1-(1±Epsilon)*(1-Quantile).
//
// See http://www.cs.rutgers.edu/~muthu/bquant.pdf for time, space, and error
// properties.
func NewHighBiased(epsilon float64) *Stream {
	ƒ := func(s *stream, r float64) float64 {
		return 2 * epsilon * (s.n - r)
	}
	return newStream(ƒ)
}

// NewTargeted returns an initialized Stream concerned with a particular set of
// quantile values that are supplied a priori. Knowing these a priori reduces
// space and computation time. The targets map maps the desired quantiles to
// their absolute errors, i.e. the true quantile of a value returned by a query
// is guaranteed to be within (Quantile±Epsilon).
//
// See http://www.cs.rutgers.edu/~muthu/bquant.pdf for time, space, and error properties.
func NewTargeted(targets map[float64]float64) *Stream {
	ƒ := func(s *stream, r float64) float64 {
		var m = math.MaxFloat64
		var f float64
		for quantile, epsilon := range targets {
			if quantile*s.n <= r {
				f = (2 * epsilon * r) / quantile
			} else {
				f = (2 * epsilon * (s.n - r)) / (1 - quantile)
			}
			if f < m {
				m = f
			}
		}
		return m
	}
	return newStream(ƒ)
}

// Stream computes quantiles for a stream of float64s. It is not thread-safe by
// design. Take care when using across multiple goroutines.
type Stream struct {
	*stream
	b      Samples
	sorted bool
}

func newStream(ƒ invariant) *Stream {
	x := &stream{ƒ: ƒ}
	return &Stream{x, make(Samples, 0, 500), true}
}

// Insert inserts v into the stream.
func (s *Stream) Insert(v float64) {
	s.insert(Sample{Value: v, Width: 1})
}

func (s *Stream) insert(sample Sample) {
	s.b = append(s.b, sample)
	s.sorted = false
	if len(s.b) == cap(s.b) {
		s.flush()
	}
}

// Query returns the computed qth percentiles value. If s was created with
// NewTargeted, and q is not in the set of quantiles provided a priori, Query
// will return an unspecified result.
func (s *Stream) Query(q float64) float64 {
	if !s.flushed() {
		// Fast path when there hasn't been enough data for a flush;
		// this also yields better accuracy for small sets of data.
		l := len(s.b)
		if l == 0 {
			return 0
		}
		i := int(math.Ceil(float64(l) * q))
		if i > 0 {
			i -= 1
		}
		s.maybeSort()
		return s.b[i].Value
	}
	s.flush()
	return s.stream.query(q)
}

// Merge merges samples into the underlying streams samples. This is handy when
// merging multiple streams from separate threads, database shards, etc.
//
// ATTENTION: This method is broken and does not yield correct results. The
// underlying algorithm is not capable of merging streams correctly.
func (s *Stream) Merge(samples Samples) {
	sort.Sort(samples)
	s.stream.merge(samples)
}

// Reset reinitializes and clears the list reusing the samples buffer memory.
func (s *Stream) Reset() {
	s.stream.reset()
	s.b = s.b[:0]
}

// Samples returns stream samples held by s.
func 