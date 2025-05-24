package handlers

import (
	"encoding/json"
	"net/http"
	"talos/pkg/jwks"
)

func TokenHandler(cfg *config.Config, keys *jwks.KeyPair) http.HandlerFunc {
	roles := map[string][]string{
		"postgres-a": {"read", "write"},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Псевдокод для проверки токена Kubernetes
		// token := parseK8SToken(r.FormValue("subject_token"))
		// namespace := extractNamespace(token)

		namespace := "postgres-a" // Заглушка
		targetScope := r.FormValue("scope")

		if _, ok := roles[namespace]; !ok {
			http.Error(w, `{"error":"access_denied"}`, http.StatusForbidden)
			return
		}

		// Генерация JWT (псевдокод)
		claims := map[string]interface{}{
			"iss":   cfg.Issuer,
			"sub":   namespace,
			"aud":   targetScope,
			"roles": roles[namespace],
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"access_token": "generated-jwt-token",
			"token_type":   "Bearer",
		})
	}
}
