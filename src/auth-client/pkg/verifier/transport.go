package verifier

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/perpetua1g0d/bmstu-diploma/src/auth-client/internal/config"
	"github.com/perpetua1g0d/bmstu-diploma/src/auth-client/internal/metrics"
)

// type VerifierMiddleware struct {
// 	verifier *Verifier
// 	cfg      *config.Config
// 	scope    string

// 	defaultRT http.RoundTripper
// }

// func NewVerifierTransport(verifier *Verifier, scope string) *VerifierMiddleware {
// 	return &VerifierMiddleware{
// 		verifier:  verifier,
// 		scope:     scope,
// 		defaultRT: http.DefaultTransport,
// 	}
// }

func VerifySQLMiddleware(next http.HandlerFunc, verifier *Verifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := verifier.cfg

		var verifyEnabled bool
		var verifyResult = "ok"
		verifyEnabled = getVerifyEnabled(cfg)
		verifyStart := time.Now()
		defer func() {
			verifyDuration := float64(time.Since(verifyStart).Milliseconds())
			metrics.TokenVerifyTotal.WithLabelValues(verifyResult, strconv.FormatBool(verifyEnabled), cfg.ClientID).Inc()
			metrics.TokenVerifyDuration.WithLabelValues(verifyResult, strconv.FormatBool(verifyEnabled), cfg.ClientID).Observe(verifyDuration)

		}()

		requiredRoles := []string{"RO"}
		dbQuery := strings.ToUpper(r.URL.Query().Get("sql"))
		if strings.Contains(dbQuery, "INSERT") || strings.Contains(dbQuery, "UPDATE") || strings.Contains(dbQuery, "DELETE") {
			requiredRoles = []string{"RW"}
		}

		log.Printf("query: %s, required_roles=%v", dbQuery, requiredRoles)

		if verifyEnabled {
			token := r.Header.Get("X-S2I-Token")
			if token == "" {
				verifyResult = "missing_token"
				respondError(w, "missing token", http.StatusUnauthorized)
				return
			}

			if verifyErr := verifier.verifyToken(token, requiredRoles); verifyErr != nil {
				log.Printf("failed to verify token: %v", verifyErr)
				verifyResult = "permissions_denied"
				respondError(w, "forbidden: token has no required roles", http.StatusUnauthorized)
				return
			}
		}
	}
}

func getVerifyEnabled(cfg *config.Config) bool {
	loaded := cfg.VerifyAuthEnabled.Load()
	if loaded == nil {
		log.Printf("config pointer[VerifyAuthEnabled] is empty! Veryfy is enabled as fallback")
		return true
	}

	return *loaded
}

func respondError(w http.ResponseWriter, message string, code int) {
	if code != http.StatusOK {
		log.Printf("request failed: status: %d, message %s", code, message)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(RespErr{Error: message})
}

type RespErr struct {
	Error string `json:"error"`
}
