package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/auth"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/config"

	_ "github.com/lib/pq"
)

type QueryRequest struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params"`
}

func QueryHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()

	// Динамическая проверка авторизации
	if cfg.AuthEnabled {
		token := r.Header.Get("X-I2I-Token")
		if token == "" {
			respondError(w, "missing token", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(token, cfg.JWTSecret)
		if err != nil {
			respondError(w, "invalid token", http.StatusForbidden)
			return
		}

		requiredRole := "RO"
		if strings.Contains(strings.ToUpper(r.URL.Query().Get("sql")), "SELECT") {
			requiredRole = "RO"
		} else {
			requiredRole = "RW"
		}

		if !auth.HasRequiredRole(claims, requiredRole) {
			respondError(w, "unauthorized: token permissions are not satisfied", http.StatusForbidden)
			return
		}
	}

	// Подключение к PostgreSQL
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
		respondError(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Выполнение запроса
	start := time.Now()
	rows, err := db.Query(req.SQL, req.Params...)
	if err != nil {
		respondError(w, fmt.Sprintf("query failed: %v", err), http.StatusBadRequest)
		return
	}
	defer rows.Close()

	// Формирование ответа
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"latency": time.Since(start).String(),
	})
}

func respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
