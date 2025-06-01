package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/jwks"
)

func CertsHandler(keys *jwks.KeyPair) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		jwks := keys.JWKS()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}

	return baseMetricsMiddleware(handler)
}
