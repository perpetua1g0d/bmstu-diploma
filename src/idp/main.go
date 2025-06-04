package main

import (
	"context"
	"log"
	"net/http"

	"github.com/perpetua1g0d/bmstu-diploma/idp/handlers"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/config"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/db"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/jwks"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx := context.Background()

	cfg := config.Load()
	keyPair := jwks.GenerateKeyPair()

	permissions := map[string]map[string][]string{
		"service-a": {"postgres-a": {"RO", "RW"}},
		"service-b": {"postgres-b": {"RO"}},
	}
	repository := db.NewRepository(permissions)

	controllerOpts := &handlers.ControllerOpts{
		Cfg:        cfg,
		Keys:       keyPair,
		Repository: repository,
	}
	controller, err := handlers.NewController(ctx, controllerOpts)
	if err != nil {
		log.Fatalf("Failed to create token controller: %v", err)
	}

	tokenHandler, err := controller.NewTokenHandler(ctx)
	if err != nil {
		log.Fatalf("Failed to create token handler: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/realms/service2infra/.well-known/openid-configuration", controller.OpenIDConfigHandler())
	mux.HandleFunc("/realms/service2infra/protocol/openid-connect/token", tokenHandler)
	mux.HandleFunc("/realms/service2infra/protocol/openid-connect/certs", controller.CertsHandler())

	mux.HandleFunc("/update_permissions", controller.NewUpdatePermissionsHandler(ctx))
	mux.HandleFunc("/get_permissions", controller.NewGetPermissionsHandler(ctx))

	log.Printf("idp OIDC server started on %s", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, mux))
}
