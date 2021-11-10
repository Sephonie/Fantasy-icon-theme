// Copyright 2015 The Prometheus Authors
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

package prometheus

import (
	"fmt"
	"math"
	"sort"
	"sync/atomic"

	"github.com/golang/protobuf/proto"

	dto "github.com/prometheus/client_model/go"
)

// A Histogram counts individual observations from an event or sample stream in
// configurable buckets. Similar to a summary, it also provides a sum of
// observations and an observation count.
//
// On the Prometheus server, quantiles can be calculated from a Histogram using
// the histogram_quantile function in the query language.
//
// Note that Histograms, in contrast to Summaries, can be aggregated with the
// Prometheus query language (see the documentation for detailed
// procedures). However, Histograms require the user to pre-define suitable
// buckets, and they are in general less accurate. The Observe method of a
// Histogram has a very low performance overhead in comparison with the Observe
// method of a Summary.
//
// To create Histogram instances, use NewHistogram.
type Histogram interface {
	Metric
	Collector

	// Observe adds a single observation to the histogram.
	Observe(float64)
}

// bucketLabel is used for the label that defines the upper bound of a
// bucket of a histogram ("le" -> "less or equal").
const bucketLabel = "le"

// DefBuckets are the default Histogram buckets. The default buckets are
// tailored to broadly measure the response time (in seconds) of a network
// service. Most likely, however, you will be required to define buckets
// customized to your use case.
var (
	DefBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

	errBucketLabelNotAllowed = fmt.Errorf(
		"%q is not allowed as label name in histograms", bucketLabel,
	)
)

// LinearBuckets creates 'count' buckets, each 'width' wide, where the lowest
// bucket has an upper bound of 'start'. The final +Inf bucket is not counted
// and not included in the returned slice. The returned slice is meant to be
// used for the Buckets field of HistogramOpts.
//
// The function panics if 'count' is zero or negative.
func LinearBuckets(start, width float64, count int) []float64 {
	if count < 1 {
		panic("LinearBuckets needs a positive count")
	}
	buckets := make([]float64, count)
	for i := range buckets {
		buckets[i] = start
		start += width
	}
	return buckets
}

// ExponentialBuckets creates 'count' buckets, where the lowest bucket has an
// upper bound of 'start' and each following bucket's upper bound is 'factor'
// times the previous bucket's upper bound. The final +Inf bucket is not counted
// and not included in the returned slice. The returned slice is meant to be
// used for the Buckets field of HistogramOpts.
//
// The function panics if 'count' is 0 or negative, if 'start' is 0 or negative,
// or if 'factor' is less than or equal 1.
func ExponentialBuckets(start, factor float64, count int) []float64 {
	if count < 1 {
		panic("ExponentialBuckets needs a positive count")
	}
	if start <= 0 {
		panic("ExponentialBuckets needs a positive start value")
	}
	if factor <= 1 {
		panic("ExponentialBuckets needs a factor greater than 1")
	}
	buckets := make([]float64, count)
	for i := range buckets {
		buckets[i] = start
		start *= factor
	}
	return buckets
}

// HistogramOpts bundles the options for creating a Histogram metric. It is
// mandatory to set Name and Help to a non-empty string. All other fields are
// optional and can safely be left at their zero value.
type HistogramOpts struct {
	// Namespace, Subsystem, and Name are components of the fully-qualified
	// name of the Histogram (created by joining these components with
	// "_"). Only Name is mandatory, the others merely help structuring the
	// name. Note that the fully-qualified name of the Histogram must be a
	// valid Prometheus metric name.
	Namespace string
	Subsystem string
	Name      string

	// Help provides information about this Histogram. Mandatory!
	//
	// Metrics with the same fully-qualified name must have the same Help
	// string.
	Help string

	// ConstLabels are used to attach fixed labels to this
	// Histogram. Histograms with the same fully-qualified name must have the
	// same label names in their ConstLabels.
	//
	// Note that in most cases, labels have a value that varies during the
	// lifetime of a process. Those labels are usually managed with a
	// HistogramVec. ConstLabels serve only special purposes. One is for the
	// special case where the value of a label does not change during the
	// lifetime of a process, e.g. if the revision of the running binary is
	// put into a label. Another, more advanced purpose is if more than one
	// Collector needs to collect Histograms with the same fully-qualified
	// name. In that case, those Summaries must differ in the values of
	// their ConstLabels. See the Collector examples.
	//
	// If the value of a label never changes (not even between binaries),
	// that label most likely should not be a label at all (but part of the
	// metric name).
	ConstLabels Labels

	// Buckets defines the buckets into which observations are counted. Each
	// element in the slice is the upper inclusive bound of a bucket. The
	// values must be sorted in strictly increasing order. There is no need
	// to add a highest bucket with +Inf bound, it will be added
	// implicitly. The default value is DefBuckets.
	Buckets []float64
}

// NewHistogram creates a new Histogram based on the provided HistogramOpts. It
// panics if the buckets in HistogramOpts are not in strictly increasing order.
func NewHistogram(opts HistogramOpts) Histogram {
	return newHistogram(
		NewDesc(
			BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
			opts.Help,
			nil,
			opts.ConstLabels,
		),
		opts,
	)
}

func newHistogram(desc *Desc, opts HistogramOpts, labelValues ...string) Histogram {
	if len(desc.variableLabels) != len(labelValues) {
		panic(errInconsistentCardinality)
	}

	for _, n := range desc.variableLabels {
		if n == bucketLabel {
			panic(errBucketLabelNotAllowed)
		}
	}
	for _, lp := range desc.constLabelPairs {
		if lp.GetName() == bucketLabel {
			panic(errBucketLabelNotAllowed)
		}
	}

	if len(opts.Buckets) == 0 {
		opts.Buckets = DefBuckets
	}

	h := &histogram{
		desc:        desc,
		upperBounds: opts.Buckets,
		labelPairs:  makeLabelPairs(desc, labelValues),
	}
	for i, upperBound := range h.upperBounds {
		if i < len(h.upperBounds)-1 {
			if upperBound >= h.upperBounds[i+1] {
				panic(fmt.Errorf(
					"histogram buckets must be in increasing order: %f >= %f",
					upperBound, h.upperBounds[i+1],
				))
			}
		} else {
			if math.IsInf(upperBound, +1) {
				// The +Inf bucket is implicit. Remove it here.
				h.upperBounds = h.upperBounds[:i]
			}
		}
	}
	// Finally we know the final length of h.upperBounds and can make counts.
	h.counts = make([]uint64, len(h.upperBounds))

	h.init(h) // Init self-collection.
	return h
}

type histogram struct {
	// sumBits contains the bits of the float64 representing the sum of all
	// observations. sumBits and count have to go first in the struct to
	// guarantee alignment for atomic operations.
	// http://golang.org/pkg/sync/atomic/#pkg-note-BUG
	sumBits uint64
	count   uint64

	selfCollector
	// Note that there is no mutex required.

	desc *Desc

	upperBounds []float64
	counts      []uint64

	labelPairs []*dto.LabelPair
}

func (h *histogram) Desc() *Desc {
	return h.desc
}

func (h *histogram) Observe(v float64) {
	// TODO(beorn7): For small numbers of buckets (<30), a linear search is
	// slightly faster than the binary search. If we really care, we could
	// switch from one search strategy to the other depending on the number
	// of buckets.
	//
	// Microbenchmarks (BenchmarkHistogramNoLabels):
	// 11 buckets: 38.3 ns/op linear - binary 48.7 ns/op