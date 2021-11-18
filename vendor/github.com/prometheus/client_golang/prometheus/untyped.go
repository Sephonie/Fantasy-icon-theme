// Copyright 2014 The Prometheus Authors
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

// Untyped is a Metric that represents a single numerical value that can
// arbitrarily go up and down.
//
// An Untyped metric works the same as a Gauge. The only difference is that to
// no type information is implied.
//
// To create Untyped instances, use NewUntyped.
type Untyped interface {
	Metric
	Collector

	// Set sets the Untyped metric to an arbitrary value.
	Set(float64)
	// Inc increments the Untyped metric by 1.
	Inc()
	// Dec decrements the Untyped metric by 1.
	Dec()
	// Add adds the given value to the Untyped metric. (The value can be
	// negative, resulting in a decrease.)
	Add(float64)
	// Sub subtracts the given value from the Untyped metric. (The value can
	// be negative, resulting in an increase.)
	Sub(float64)
}

// UntypedOpts is an alias for Opts. See there for doc comments.
type UntypedOpts Opts

// NewUntyped creates a new Untyped metric from the provided UntypedOpts.
func NewUntyped(opts UntypedOpts) Untyped {
	return newValue(NewDesc(
		BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		nil,
		opts.ConstLabels,
	), UntypedValue, 0)
}

// UntypedVec is a Collector that bundles a set of Untyped metrics that all
// share the same Desc, but have different values for their variable
// labels. This is used if you want to count the same thing partitioned by
// various dimensions. Create instances with NewUntypedVec.
type UntypedVec struct {
	*MetricVec
}

// NewUntypedVec creates a new UntypedVec based on the provided UntypedOpts and
// partitioned by the given label names. At least one label name must be
// provided.
func NewUntypedVec(opts UntypedOpts, labelNames []string) *UntypedVec {
	desc := NewDesc(
		BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		labelNames,
		opts.ConstLabels,
	)
	return &UntypedVec{
		MetricVec: newMet