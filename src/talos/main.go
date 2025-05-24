package main

import (
	"log"
	"net/http"

	"github.com/perpetua1g0d/bmstu-diploma/talos/handlers"
	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/config"
	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/jwks"
)

func main() {
	cfg := config.Load()
	keyPair := jwks.GenerateKeyPair()

	http.HandleFunc("/realms/infra2infra/.well-known/openid-configuration", handlers.OpenIDConfigHandler(cfg))
	http.HandleFunc("/realms/infra2infra/protocol/openid-connect/token", handlers.TokenHandler(cfg, keyPair))
	http.HandleFunc("/realms/infra2infra/protocol/openid-connect/certs", handlers.CertsHandler(keyPair))

	log.Printf("Talos OIDC server started on %s", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, nil))
}
