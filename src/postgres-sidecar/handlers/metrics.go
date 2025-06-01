package handlers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
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

	dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_duration_milliseconds",
		Help:    "Duration of database queries in milliseconds",
		Buckets: []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000},
	}, []string{"operation", "service_name"}) // todo: add target
)

var (
	tokenVerifyTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "client_token_verify_requests_total",
		Help: "Total number of requests verified with token",
	}, []string{"scope", "result", "enabled"})
	tokenVerifyDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "client_token_verify_duration_milliseconds",
		Help:    "Duration of token verification in milliseconds",
		Buckets: []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000},
	}, []string{"scope", "result", "enabled"})
)
