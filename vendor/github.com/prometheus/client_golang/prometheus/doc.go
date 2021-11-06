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

// Package prometheus provides metrics primitives to instrument code for
// monitoring. It also offers a registry for metrics. Sub-packages allow to
// expose the registered metrics via HTTP (package promhttp) or push them to a
// Pushgateway (package push).
//
// All exported functions and methods are safe to be used concurrently unless
//specified otherwise.
//
// A Basic Example
//
// As a starting point, a very basic usage example:
//
//    package main
//
//    import (
//    	"net/http"
//
//    	"github.com/prometheus/client_golang/prometheus"
//    	"github.com/prometheus/client_golang/prometheus/promhttp"
//    )
//
//    var (
//    	cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
//    		Name: "cpu_temperature_celsius",
//    		Help: "Current temperature of the CPU.",
//    	})
//    	hdFailures = prometheus.NewCounterVec(
//    		prometheus.CounterOpts{
//    			Name: "hd_errors_total",
//    			Help: "Number of hard-disk errors.",
//    		},
//    		[]string{"device"},
//    	)
//    )
//
//    func init() {
//    	// Metrics have to be registered to be exposed:
//    	prometheus.MustRegister(cpuTemp)
//    	prometheus.MustRegister(hdFailures)
//    }
//
//    func main() {
//    	cpuTemp.Set(65.3)
//    	hdFailures.With(prometheus.Labels{"device":"/dev/sda"}).Inc()
//
//    	// The Handler function provides a default handler to expose metrics
//    	// via an HTTP server. "/metrics" is the usual endpoint for that.
//    	http.Handle("/metrics", promhttp.Handler())
//    	http.ListenAndServe(":8080", nil)
//    }
//
//
// This is a complete program that exports two metrics, a Gauge and a Counter,
// the latter with a label attached to turn it into a (one-dimensional) vector.
//
// Metrics
//
// The number of exported identifiers in this package might appear a bit
// overwhelming. Hovever, in addition to the basic plumbing shown in the example
// above, you only need to understand the different metric types and their
// vector versions for basic usage.
//
// Above, you have already touched the Counter and the Gauge. There are two more
// advanced metric types: the Summary and Histogram. A more thorough description
// of those four metric types can be found in the Prometheus docs:
// https://prometheus.io/docs/concepts/metric_types/
//
// A fifth "type" of metric is Untyped. It behaves like a Gauge, but signals the
// Prometheus server not to assume anything about its type.
//
// In addition to the fundamental metric types Gauge, Counter, Summary,
// Histogram, and Untyped, a very important part of the Prometheus data model is
// the partitioning of samples along dimensions called labels, which results in
// metric vectors. The fundamental types are GaugeVec, CounterVec, SummaryVec,
// HistogramVec, an