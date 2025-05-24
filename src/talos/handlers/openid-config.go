package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/config"
)

func OpenIDConfigHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"issuer":                                cfg.Issuer,
			"token_endpoint":                        cfg.Issuer + "/protocol/openid-connect/token",
			"jwks_uri":                              cfg.Issuer + "/protocol/openid-connect/certs",
			"grant_types_supported":                 []string{"urn:ietf:params:oauth:grant-type:token-exchange"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
