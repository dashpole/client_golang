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
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"regexp"
	"time"

	prombridge "go.opentelemetry.io/contrib/bridges/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 12 buckets per histogram
const histogramCardinality = 10000
const counterCardinality = 10000
const gaugeCardinality = 10000

func main() {
	// expose a minimum set of memory metrics to be able to accurately measure memory usage.
	memoryUsageRegistry := prometheus.NewRegistry()
	memoryUsageRegistry.MustRegister(collectors.NewGoCollector(
		collectors.WithGoCollectorMemStatsMetricsDisabled(),
		collectors.WithoutGoCollectorRuntimeMetrics(regexp.MustCompile(".*")),
		collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/memory/classes/total:bytes")}),
	))

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

	// Expose /memorymetrics HTTP endpoint.
	http.Handle(
		"/memorymetrics", promhttp.HandlerFor(
			memoryUsageRegistry,
			promhttp.HandlerOpts{}),
	)

	//Expose /metrics HTTP endpoint.
	http.Handle(
		"/metrics", promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{}),
	)

	ctx := context.Background()
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithProducer(prombridge.NewMetricProducer(prombridge.WithGatherer(registry))),
			metric.WithInterval(time.Second))),
	)
	defer meterProvider.Shutdown(ctx)
	otel.SetMeterProvider(meterProvider)
	// To test: curl localhost:8080/metrics
	// This is also for the profiling endpoint
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
