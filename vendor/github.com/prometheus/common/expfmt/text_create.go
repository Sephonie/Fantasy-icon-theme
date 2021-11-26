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

package expfmt

import (
	"fmt"
	"io"
	"math"
	"strings"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
)

// MetricFamilyToText converts a MetricFamily proto message into text format and
// writes the resulting lines to 'out'. It returns the number of bytes written
// and any error encountered. The output will have the same order as the input,
// no further sorting is performed. Furthermore, this function assumes the input
// is already sanitized and does not perform any sanity checks. If the input
// contains duplicate metrics or invalid metric or label names, the conversion
// will result in invalid text format output.
//
// This method fulfills the type 'prometheus.encoder'.
func MetricFamilyToText(out io.Writer, in *dto.MetricFamily) (int, error) {
	var written int

	// Fail-fast checks.
	if len(in.Metric) == 0 {
		return written, fmt.Errorf("MetricFamily has no metrics: %s", in)
	}
	name := in.GetName()
	if name == "" {
		return written, fmt.Errorf("MetricFamily has no name: %s", in)
	}

	// Comments, first HELP, then TYPE.
	if in.Help != nil {
		n, err := fmt.Fprintf(
			out, "# HELP %s %s\n",
			name, escapeString(*in.Help, false),
		)
		written += n
		if err != nil {
			return written, err
		}
	}
	metricType := in.GetType()
	n, err := fmt.Fprintf(
		out, "# TYPE %s %s\n",
		name, strings.ToLower(metricType.String()),
	)
	written += n
	if err != nil {
		return written, err
	}

	// Finally the samples, one line for each.
	for _, metric := range in.Metric {
		switch metricType {
		case dto.MetricType_COUNTER:
			if metric.Counter == nil {
				return written, fmt.Errorf(
					"expected counter in metric %s %s", name, metric,
				)
			}
			n, err = writeSample(
				name, metric, "", "",
				metric.Counter.GetValue(),
				out,
			)
		case dto.MetricType_GAUGE:
			if metric.Gauge == nil {
				return written, fmt.Errorf(
					"expected gauge in metric %s %s", name, metric,
				)
			}
			n, err = writeSample(
				name, metric, "", "",
				metric.Gauge.GetValue(),
				out,
			)
		case dto.MetricType_UNTYPED:
			if metric.Untyped == nil {
				return written, fmt.Errorf(
					"expected untyped in metric %s %s", name, metric,
				)
			}
			n, err = writeSample(
				name, metric, "", "",
				metric.Untyped.GetValue(),
				out,
			)
		case dto.MetricType_SUMMARY:
			if metric.Summary == nil {
				return written, fmt.Errorf(
					"expected summary in metric %s %s", name, metric,
				)
			}
			for _, q := range metric.Summary.Quantile {
				n, err = writeSample(
					name, metric,
					model.QuantileLabel, fmt.Sprint(q.GetQuantile()),
					q.GetValue(),
					out,
				)
				written += n
				if err != nil {
					return written, err
				}
			}
			n, err = writeSample(
				name+"_sum", metric, "", "",
				metric.Summary.GetSampleSum(),
				out,
			)
			if err != nil {
				return written, err
			}
			written += n
			n, err = writeSample(
				name+"_count", metric, "", "",
				float64(metric.Summary.GetSampleCount()),
				out,
			)
		case dto.MetricType_HISTOGRAM:
			if metric.Histogram == nil {
				return written, fmt.Errorf(
					"expected histogram in metric %s %s", name, metric,
				)
			}
			infSeen := false
			for _, q := range metric.Histogram.Bucket {
				n, err = writeSample(
					name+"_bucket", metric,
					model.BucketLabel, fmt.Sprint(q.GetUpperBound()),
					float64(q.GetCumulativeCount()),
					out,
				)
				written += n
				if err != nil {
					return written, err
				}
				if math.IsInf(q.GetUpperBound(), +1) {
					infSeen = true
				}
			}
			if !infSeen {
				n, err = writeSample(
					name+"_bucket", metric,
					model.BucketLabel, "+Inf",
					float64(metric.Histogram.GetSampleCount()),
					out,
				)
				if err != nil {
					return written, err
				}
				written += n
			}
			n, err = writeSample(
				name+"_sum", metric, "", "",
				metric.Histogram.GetSampleSum(),
				out,
			)
			if err != nil {
				return written, err
			}
			written += n
			n, err = writeSample(
				name+"_count", metric, "", "",
				float64(metric.Histogram.GetSampleCount()),
				out,
			)
		default:
			return written, fmt.Errorf(
				"unexpected type in metric %s %s", name, metric,
			)
		}
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

// writeSample writes a single sample in text format to out, given the metric
// name, the metric pro