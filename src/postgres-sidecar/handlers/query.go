package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	auth_verifier "github.com/perpetua1g0d/bmstu-diploma/auth-client/pkg/verifier"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/config"

	_ "github.com/lib/pq"
)

type QueryRequest struct {
	SQL    string `json:"sql"`
	Params []any  `json:"params"`
}

func NewQueryHandler(ctx context.Context, cfg *config.Config, db *sql.DB, tokenVerifier *auth_verifier.Verifier) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req QueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		start := time.Now()
		rows, err := db.Query(req.SQL, req.Params...)
		if err != nil {
			respondError(w, fmt.Sprintf("query failed: %v", err), http.StatusBadRequest)
			return
		}
		defer rows.Close()

		func() {
			operation := "read"
			if !strings.Contains(strings.ToUpper(req.SQL), "SELECT") {
				operation = "write"
			}
			durationMs := float64(time.Since(start).Milliseconds())
			dbQueryDuration.WithLabelValues(operation, cfg.ServiceName).Observe(durationMs)
		}()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"latency": time.Since(start).String(),
		})
	}

	authorizedHandler := auth_verifier.VerifySQLMiddleware(handler, tokenVerifier)
	return baseMetricsMiddleware(authorizedHandler, cfg.ServiceName)
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
