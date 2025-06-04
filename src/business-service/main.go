package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/perpetua1g0d/bmstu-diploma/business-service/config"
	auth_signer "github.com/perpetua1g0d/bmstu-diploma/src/auth-client/pkg/signer"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()

	mux := http.NewServeMux()

	scopes := []string{cfg.InitTarget}
	signer, err := auth_signer.NewTokenSigner(ctx, cfg.ServiceName, scopes, cfg.SignAuthEnabled)
	if err != nil {
		log.Fatalf("failed to create signer: %v", err)
	}

	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/reload_config", signer.NewRealodHandler())
	mux.Handle("/refresh_tokens", signer.NewRefreshTokensHandler())

	authTransport := auth_signer.NewAuthTransport(signer, cfg.InitTarget)
	go func() {
		time.Sleep(30 * time.Second) // wait for app is up
		httpClient := &http.Client{
			Timeout:   5 * time.Second,
			Transport: authTransport,
		}
		for {
			time.Sleep(10 * time.Second)
			sendDBQuery(cfg, httpClient)
		}
	}()

	log.Printf("Starting %s on :8080 (sign: %v)", cfg.ServiceName, cfg.SignAuthEnabled)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
