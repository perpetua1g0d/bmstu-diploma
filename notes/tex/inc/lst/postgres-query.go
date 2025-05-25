type QueryRequest struct {
	SQL    string `json:"sql"`
	Params []any  `json:"params"`
}

func NewQueryHandler(ctx context.Context, authClient *auth_client.AuthClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming request: %s %s", r.Method, r.URL)

		cfg := config.GetConfig()

		if cfg.VerifyAuthEnabled {
			token := r.Header.Get("X-I2I-Token")
			if token == "" {
				respondError(w, "missing token", http.StatusUnauthorized)
				return
			}

			requiredRole := "RO"
			sqlQuery := strings.ToUpper(r.URL.Query().Get("sql")
			if !strings.Contains(sqlQuery), "SELECT") {
				requiredRole = "RW"
			}

			if verifyErr := authClient.VerifyToken(token, []string{requiredRole}); verifyErr != nil {
				log.Printf("failed to verify token: %v", verifyErr)
				respondError(w, "forbidden: token has no required roles", http.StatusUnauthorized)
				return
			}

			log.Printf("successfully verified incoming token")
		}

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
}
