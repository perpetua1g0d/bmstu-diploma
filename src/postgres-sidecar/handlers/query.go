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

	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/config"
	auth_verifier "github.com/perpetua1g0d/bmstu-diploma/src/auth-client/pkg/verifier"

	_ "github.com/lib/pq"
)

type QueryRequest struct {
	SQL    string `json:"sql"`
	Params []any  `json:"params"`
}

func NewQueryHandler(ctx context.Context, cfg *config.Config, db *sql.DB, tokenVerifier *auth_verifier.Verifier) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("start process /query request")

		var req QueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("decoded request body: %+v", req)

		start := time.Now()
		rows, err := db.Query(req.SQL, req.Params...)
		if err != nil {
			respondError(w, fmt.Sprintf("query failed: %v", err), http.StatusBadRequest)
			return
		}
		defer rows.Close()
		operation := "read"
		if !strings.Contains(strings.ToUpper(req.SQL), "SELECT") {
			operation = "write"
		}
		durationMs := float64(time.Since(start).Milliseconds())
		dbQueryDuration.WithLabelValues(operation, cfg.ServiceName).Observe(durationMs)

		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"latency": time.Since(start).String(),
		}); err != nil {
			respondError(w, fmt.Sprintf("response wirte failed: %v", err), http.StatusInternalServerError)
			return
		}
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
