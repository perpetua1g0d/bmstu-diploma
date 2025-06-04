package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TokenSignedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "client_token_sign_requests_total",
		Help: "Total number of requests signed with token",
	}, []string{"scope", "result", "enabled", "service_name"})
	TokenSignDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "client_token_sign_duration_milliseconds",
		Help:    "Duration of token signing in milliseconds",
		Buckets: []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000},
	}, []string{"scope", "result", "enabled", "service_name"})
)

var (
	TokenVerifyTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "client_token_verify_requests_total",
		Help: "Total number of requests verified with token",
	}, []string{"result", "enabled", "service_name"})
	TokenVerifyDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "client_token_verify_duration_milliseconds",
		Help:    "Duration of token verification in milliseconds",
		Buckets: []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000},
	}, []string{"result", "enabled", "service_name"})
)
