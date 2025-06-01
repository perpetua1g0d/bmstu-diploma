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
)
