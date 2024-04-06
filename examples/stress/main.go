// Copyright 2022 The Prometheus Authors
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

// A simple example of how to record a latency metric with exemplars, using a fictional id
// as a prometheus label.

package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 12 buckets per histogram
const histogramCardinality = 1000
const counterCardinality = 1000
const gaugeCardinality = 1000

func main() {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "histogram",
		Help: "A histogram",
	}, []string{"label"})
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "counter",
		Help: "A counter",
	}, []string{"label"})
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gauge",
		Help: "A gauge",
	}, []string{"label"})

	// Create non-global registry.
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		histogram,
		counter,
		gauge,
	)

	go func() {
		for {
			for i := 0; i < histogramCardinality; i++ {
				// One observation for each histogram bucket
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(0.001)
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(0.008)
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(0.011)
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(0.03)
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(0.07)
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(0.11)
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(0.3)
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(0.7)
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(1.1)
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(3)
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(7)
				histogram.WithLabelValues(fmt.Sprintf("%v", i)).Observe(20)
			}
			for i := 0; i < counterCardinality; i++ {
				counter.WithLabelValues(fmt.Sprintf("%v", i)).Add(0.01)
			}
			for i := 0; i < gaugeCardinality; i++ {
				gauge.WithLabelValues(fmt.Sprintf("%v", i)).Set(0.01)
			}
			time.Sleep(time.Second)
		}
	}()

	// Expose /metrics HTTP endpoint using the created custom registry.
	http.Handle(
		"/metrics", promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{}),
	)
	// To test: curl localhost:8080/metrics
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
