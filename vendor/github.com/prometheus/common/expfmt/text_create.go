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
		return written, fmt