package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/perpetua1g0d/bmstu-diploma/business-service/config"
	auth_signer "github.com/perpetua1g0d/bmstu-diploma/src/auth-client/pkg/signer"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/lib/pq"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	scopes := []string{cfg.InitTarget}
	signer, err := auth_signer.NewTokenSigner(ctx, cfg.ServiceName, scopes, cfg.SignAuthEnabled)
	if err != nil {
		log.Fatalf("failed to create signer: %v", err)
	}

	service := NewService(cfg, signer)
	defer service.db.Close()

	mux.Handle("/benchmark/start", http.HandlerFunc(service.benchmarkHandler))
	mux.Handle("/benchmark/stop", http.HandlerFunc(service.stopBenchmarkHandler))
	mux.Handle("/reload_config", signer.NewRealodHandler())
	mux.Handle("/refresh_tokens", signer.NewRefreshTokensHandler())

	go func() {
		for {
			time.Sleep(10 * time.Second)
			if service.benchmark.running {
				continue
			}
			service.sendRegularQuery()
		}
	}()

	log.Printf("Starting %s on :8080 (sign: %v)", cfg.ServiceName, cfg.SignAuthEnabled)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (s *Service) sendRegularQuery() {
	status, err := s.sendBenchmarkQuery("light")
	if err != nil {
		log.Printf("Regular query failed: %d, error: %v", status, err)
	} else {
		log.Printf("Regular query completed")
	}
}
