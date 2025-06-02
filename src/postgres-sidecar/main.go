package main

import (
	"context"
	"log"
	"net/http"

	auth_verifier "github.com/perpetua1g0d/bmstu-diploma/auth-client/pkg/verifier"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/config"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/handlers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const idpAddress = "http://idp.idp.svc.cluster.local:80"

const benchmarksResultsFile = "/var/log/results.csv"

var (
	tokenSignedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "client_token_sign_requests_total",
		Help: "Total number of requests signed with token",
	}, []string{"scope", "result", "enabled", "service_name"})
	tokenSignDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "client_token_sign_duration_milliseconds",
		Help:    "Duration of token signing in milliseconds",
		Buckets: []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000},
	}, []string{"scope", "result", "enabled", "service_name"})
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()

	verifier, err := auth_verifier.NewVerifier(ctx, cfg.ServiceName, cfg.VerifyAuthEnabled)
	if err != nil {
		log.Fatalf("failed to create verifier: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/reload_config", verifier.NewRealodHandler())

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc(cfg.ServiceEndpoint, handlers.NewQueryHandler(ctx, cfg, verifier))

	log.Printf("Starting %s on :8080 (verify: %v)", cfg.ServiceName, cfg.VerifyAuthEnabled)
	log.Fatal(http.ListenAndServe(":8080", mux)) // root handler promhttp.Handler()
}
