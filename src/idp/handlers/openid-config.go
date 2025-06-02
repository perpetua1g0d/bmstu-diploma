package handlers

import (
	"encoding/json"
	"net/http"
)

func (ctl *Controller) OpenIDConfigHandler() http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"issuer":                                ctl.cfg.Issuer,
			"token_endpoint":                        ctl.cfg.Issuer + "/realms/infra2infra/protocol/openid-connect/token",
			"jwks_uri":                              ctl.cfg.Issuer + "/realms/infra2infra/protocol/openid-connect/certs",
			"grant_types_supported":                 []string{grantTypeTokenExchange},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	return handler
}
