package handlers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	tokenIssuedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "idp_token_issued_total",
		Help: "Total number of tokens issued",
	}, []string{"result", "client_id", "scope"})

	tokenIssueDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "idp_token_issue_duration_milliseconds",
		Help:    "Duration of token issuing in milliseconds",
		Buckets: []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000},
	}, []string{"result", "client_id", "scope"})

	// base HTTP metriccs:
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status", "service_name"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_milliseconds",
		Help:    "Duration of HTTP requests in milliseconds",
		Buckets: []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000},
	}, []string{"method", "path", "service_name"})

	httpRequestSize = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "http_request_size_bytes",
		Help:       "Size of HTTP requests",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"method", "path", "service_name"})

	httpResponseSize = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "http_response_size_bytes",
		Help:       "Size of HTTP responses",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"method", "path", "service_name"})
)
