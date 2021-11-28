// Copyright 2013 The Prometheus Authors
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

package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// AlertNameLabel is the name of the label containing the an alert's name.
	AlertNameLabel = "alertname"

	// ExportedLabelPrefix is the prefix to prepend to the label names present in
	// exported metrics if a label of the same name is added by the server.
	ExportedLabelPrefix = "exported_"

	// MetricNameLabel is the label name indicating the metric name of a
	// timeseries.
	MetricNameLabel = "__name__"

	// SchemeLabel is the name of the label that holds the scheme on which to
	// scrape a target.
	SchemeLabel = "__scheme__"

	// AddressLabel is the name of the label that holds the address of
	// a scrape target.
	AddressLabel = "__address__"

	// MetricsPathLabel is the name of the label that holds the path on which to
	// scrape a target.
	MetricsPathLabel = "__metrics_path__"

	// ReservedLabelPrefix is a prefix which is not legal in user-supplied
	// label names.
	ReservedLabelPrefix = "__"

	// MetaLabelPrefix is a prefix for labels that provide meta information.
	// Labels with this prefix are used for intermediate label processing and
	// will not be attached to time series.
	MetaLabelPrefix = "__meta_"

	// TmpLabelPrefix is a prefix for temporary labels as part of relabelling.
	// Labels with this prefix are used for intermediate label processing and
	// will not be attached to time series. This is reserved for use in
	// Prometheus configuration files by users.
	TmpLabelPrefix = "__tmp_"

	// ParamLabelPrefix is a prefix for labels that provide URL parameters
	// used to scrape a target.
	ParamLabelPrefix = "__param_"

	// JobLabel is the label name indicating the job from which a timeseries
	// was scraped.
	JobLabel = "job"

	// InstanceLabel is the label name used for the instance label.
	InstanceLabel = "instance"

	// BucketLabel is used for the label that defines the upper bound of a
	// bucket of a histogram ("le" -> "less or equal").
	BucketLabel = "le"

	// QuantileLabel is used for the label that defines the quantile in a
	// summary.
	QuantileLabel = "quantile"
)

// LabelNameRE is a regular expression matching valid label names. Note that the
// IsValid method of LabelName performs the same check but faster than a match
// with this regular expression.
var LabelNameRE = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")

// A LabelName is a key for a LabelSet or Metric.  It has a value associated
// therewith.
type LabelName string

// IsValid is true iff th