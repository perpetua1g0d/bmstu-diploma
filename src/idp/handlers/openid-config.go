package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/config"
)

func OpenIDConfigHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"issuer":                                cfg.Issuer,
			"token_endpoint":                        cfg.Issuer + "/realms/infra2infra/protocol/openid-connect/token",
			"jwks_uri":                              cfg.Issuer + "/realms/infra2infra/protocol/openid-connect/certs",
			"grant_types_supported":                 []string{grantTypeTokenExchange},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
