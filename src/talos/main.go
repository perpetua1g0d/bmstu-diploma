package main

import (
	"context"
	"log"
	"net/http"

	"github.com/perpetua1g0d/bmstu-diploma/talos/handlers"
	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/config"
	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/jwks"
)

func main() {
	ctx := context.Background()

	cfg := config.Load()
	keyPair := jwks.GenerateKeyPair()

	tokenHandler, err := handlers.NewTokenHandler(ctx, cfg, keyPair)
	if err != nil {
		log.Fatalf("Failed to create token handler: %v", err)
	}

	http.HandleFunc("/realms/infra2infra/.well-known/openid-configuration", handlers.OpenIDConfigHandler(cfg))
	http.HandleFunc("/realms/infra2infra/protocol/openid-connect/token", tokenHandler)
	http.HandleFunc("/realms/infra2infra/protocol/openid-connect/certs", handlers.CertsHandler(keyPair))

	log.Printf("Talos OIDC server started on %s", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, nil))
}
