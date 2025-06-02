package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	auth_client "github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/auth/client"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/config"

	_ "github.com/lib/pq"
)

type QueryRequest struct {
	SQL    string `json:"sql"`
	Params []any  `json:"params"`
}

func NewQueryHandler(ctx context.Context, cfg *config.Config, authClient *auth_client.AuthClient) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("Incoming request: %s %s", r.Method, r.URL)

		// cclient := aauth.New

		var verifyEnabled bool
		var verifyResult = "ok"
		scope := cfg.ServiceName
		verifyEnabled = getVerifyEnabled(cfg)
		verifyStart := time.Now()
		if verifyEnabled {
			token := r.Header.Get("X-I2I-Token")
			if token == "" {
				verifyResult = "missing_token"
				verifyDuration := float64(time.Since(verifyStart).Milliseconds())
				tokenVerifyTotal.WithLabelValues(scope, verifyResult, strconv.FormatBool(verifyEnabled), cfg.ServiceName).Inc()
				tokenVerifyDuration.WithLabelValues(scope, verifyResult, strconv.FormatBool(verifyEnabled), cfg.ServiceName).Observe(verifyDuration)

				respondError(w, "missing token", http.StatusUnauthorized)
				return
			}

			requiredRole := "RO"
			if !strings.Contains(strings.ToUpper(r.URL.Query().Get("sql")), "SELECT") {
				requiredRole = "RW"
			}

			if verifyErr := authClient.VerifyToken(token, []string{requiredRole}); verifyErr != nil {
				log.Printf("failed to verify token: %v", verifyErr)
				verifyResult = "permissions_denied"
				verifyDuration := float64(time.Since(verifyStart).Milliseconds())
				tokenVerifyTotal.WithLabelValues(scope, verifyResult, strconv.FormatBool(verifyEnabled), cfg.ServiceName).Inc()
				tokenVerifyDuration.WithLabelValues(scope, verifyResult, strconv.FormatBool(verifyEnabled), cfg.ServiceName).Observe(verifyDuration)

				respondError(w, "forbidden: token has no required roles", http.StatusUnauthorized)
				return
			}
		}

		verifyDuration := float64(time.Since(verifyStart).Milliseconds())
		tokenVerifyTotal.WithLabelValues(scope, verifyResult, strconv.FormatBool(verifyEnabled), cfg.ServiceName).Inc()
		tokenVerifyDuration.WithLabelValues(scope, verifyResult, strconv.FormatBool(verifyEnabled), cfg.ServiceName).Observe(verifyDuration)

		db, err := sql.Open("postgres", fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.PostgresHost,
			cfg.PostgresPort,
			cfg.PostgresUser,
			cfg.PostgresPassword,
			cfg.PostgresDB,
		))
		if err != nil {
			respondError(w, "database connection failed", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		var req QueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		start := time.Now()
		defer func() {
			operation := "read"
			if !strings.Contains(strings.ToUpper(req.SQL), "SELECT") {
				operation = "write"
			}
			durationMs := float64(time.Since(start).Milliseconds())
			dbQueryDuration.WithLabelValues(operation, cfg.ServiceName).Observe(durationMs)
		}()

		rows, err := db.Query(req.SQL, req.Params...)
		if err != nil {
			respondError(w, fmt.Sprintf("query failed: %v", err), http.StatusBadRequest)
			return
		}
		defer rows.Close()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"latency": time.Since(start).String(),
		})
	}

	return baseMetricsMiddleware(handler, cfg.ServiceName)
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
