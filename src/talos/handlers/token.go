package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/config"
	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/jwks"
)

type TokenRequest struct {
	GrantType        string `form:"grant_type"`
	SubjectTokenType string `form:"subject_token_type"`
	SubjectToken     string `form:"subject_token"`
	Scope            string `form:"scope"`
}

func TokenHandler(cfg *config.Config, keys *jwks.KeyPair) http.HandlerFunc {
	rolesDB := map[string]map[string][]string{
		"postgres-a": {
			"postgres-b": {"read", "write"},
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Псевдокод для примера
		namespace := "postgres-a" // Должно извлекаться из k8s токена
		targetScope := r.FormValue("scope")

		// Проверка прав
		if allowedRoles, ok := rolesDB[namespace][targetScope]; !ok {
			http.Error(w, `{"error":"access_denied"}`, http.StatusForbidden)
			return
		} else {
			token := jwks.GenerateJWT(keys, map[string]interface{}{
				"iss":   cfg.Issuer,
				"sub":   namespace,
				"aud":   targetScope,
				"roles": allowedRoles,
			})

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"access_token": token,
				"token_type":   "Bearer",
			})
		}
	}
}
