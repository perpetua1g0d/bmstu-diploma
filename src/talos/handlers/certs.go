package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/jwks"
)

func CertsHandler(keys *jwks.KeyPair) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwks := map[string]interface{}{
			"keys": []interface{}{
				keys.JWKSPublicKey(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}
}
