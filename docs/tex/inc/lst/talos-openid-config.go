func OpenIDConfigHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenEndpointPath := "/realms/infra2infra/protocol/openid-connect/token"
		certsEndpointPath := "/realms/infra2infra/protocol/openid-connect/certs"
		response := map[string]interface{}{
			"issuer":                                cfg.Issuer,
			"token_endpoint":                        cfg.Issuer + tokenEndpointPath,
			"jwks_uri":                              cfg.Issuer + certsEndpointPath,
			"grant_types_supported":                 []string{grantTypeTokenExchange},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
