package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Everything here is to send logs to prometheus
// that will be shown inside Grafana dashboard
// Main crux of all logs and monitoring should happen here

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_ms",
			Help:    "HTTP request duration in milliseconds",
			Buckets: []float64{50, 100, 200, 300, 500, 800, 1000, 1500, 2000, 3000, 5000},
		},
		[]string{"path", "method"},
	)

	RequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_errors_total",
			Help: "Total number of HTTP request errors",
		},
		[]string{"path", "method", "status"},
	)

	DBQueriesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
	)

	DBQueryDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_ms",
			Help:    "Database query duration in milliseconds",
			Buckets: []float64{50, 100, 200, 300, 500, 800, 1000, 1500},
		},
	)

	DBQueryErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_query_errors_total",
			Help: "Total number of database query errors",
		},
	)

	ExternalAPICallsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "external_api_calls_total",
			Help: "Total number of external API calls",
		},
	)

	ExternalAPICallDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "external_api_call_duration_ms",
			Help:    "External API call duration in milliseconds",
			Buckets: []float64{100, 300, 500, 800, 1000, 1500, 2000, 3000},
		},
	)

	ExternalAPICallErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "external_api_call_errors_total",
			Help: "Total number of external API call errors",
		},
	)

	ProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "processing_duration_ms",
			Help:    "Processing duration in milliseconds",
			Buckets: []float64{50, 100, 200, 300, 500, 800, 1000},
		},
	)

	ProcessingErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "processing_errors_total",
			Help: "Total number of processing errors",
		},
	)

	LogsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logs_total",
			Help: "Total number of logs by level",
		},
		[]string{"level"},
	)
)
