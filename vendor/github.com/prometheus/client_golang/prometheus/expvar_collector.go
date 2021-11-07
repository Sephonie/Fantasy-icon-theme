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

import (
	"encoding/json"
	"expvar"
)

type expvarCollector struct {
	exports map[string]*Desc
}

// NewExpvarCollector returns a newly allocated expvar Collector that still has
// to be registered with a Prometheus registry.
//
// An expvar Collector collects metrics from the expvar interface. It provides a
// quick way to expose numeric values that are already exported via expvar as
// Prometheus metrics. Note that the data models of expvar and Prometheus are
// fundamentally different, and that the expvar Collector is inherently slower
// than native Prometheus metrics. Thus, the expvar Collector is probably great
// for experiments and prototying, but you should seriously consider a more
// direct implementation of Prometheus metrics for monitoring production
// systems.
//
// The exports map has the following meaning:
//
// The keys in the map correspond to expvar keys, i.e. for every expvar key you
// want to export as Prometheus metric, you need an entry in the exports
// map. The descriptor mapped to each key describes how to export the expvar
// value. It defines the name and the help string of the Prometheus metric
// proxying the expvar value. The type will always be Untyped.
//
// For descriptors without variable labels, the expvar value must be a number or
// a bool. The number is then directly exported as the Prometheus sample
// value. (For a bool, 'false' translates to 0 and 'true' to 1). Expvar values
// that are not numbers or bools are silently ignored.
//
// If the descriptor has one variable label, the expvar value must be an expvar
// map. The k