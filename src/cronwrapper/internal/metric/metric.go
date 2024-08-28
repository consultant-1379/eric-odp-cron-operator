// Package metric defines the prometheus metrics for the service and
// provides method for setting up.
package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	servicePrefix = "eric_oss_odp_cronwrapper"
)

var (
	// Registry The prometheus registry.
	Registry = prometheus.NewRegistry()

	// RequestsTotal total number of processed requests.
	RequestsTotal prometheus.Counter
	// RequestsFailedTotal total number of failed requests.
	RequestsFailedTotal prometheus.Counter
	// AnalysisParametersTotal number of analysis parameters returned from text analyzer.
	AnalysisParametersTotal prometheus.Gauge
	// HTTPResponseSizeBytes response size for HTTP requests.
	HTTPResponseSizeBytes *prometheus.HistogramVec
	// RequestDurationSeconds time spent to process words across application from
	// incoming request to response.
	RequestDurationSeconds *prometheus.HistogramVec
)

func createMetrics() {
	RequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: servicePrefix,
			Name:      "requests_total",
			Help:      "Total number of processed requests",
		})
	RequestsFailedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: servicePrefix,
			Name:      "requests_failed_total",
			Help:      "Total number of failed requests",
		})
	AnalysisParametersTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: servicePrefix,
			Name:      "analysis_parameters_total",
			Help:      "Number of analysis parameters returned from text analyzer",
		})
	HTTPResponseSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: servicePrefix,
			Name:      "http_response_size_bytes",
			Help:      "Response size for HTTP requests",
			Buckets:   prometheus.ExponentialBuckets(100, 2, 5), //nolint
		},
		[]string{"handler"},
	)
	RequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: servicePrefix,
			Name:      "request_duration_seconds",
			Help:      "Time spent to process words across application from incoming request to response",
			Buckets:   prometheus.ExponentialBuckets(0.5, 2, 5), //nolint
		},
		[]string{"handler", "method"},
	)
}

func registerMetrics() {
	Registry.MustRegister(RequestsTotal)
	Registry.MustRegister(RequestsFailedTotal)
	Registry.MustRegister(AnalysisParametersTotal)
	Registry.MustRegister(HTTPResponseSizeBytes)
	Registry.MustRegister(RequestDurationSeconds)
}

// SetupMetric function creates the metrics and registers them.
func SetupMetric() {
	createMetrics()
	registerMetrics()
}
