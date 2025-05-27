func sendBenchmarkQuery(cfg *config.Config, authClient *auth_client.AuthClient) {
	target := fmt.Sprintf("http://%s.%s.svc.cluster.local:8080%s",
		cfg.InitTarget,
		cfg.InitTarget,
		cfg.ServiceEndpoint,
	)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"sql":    `INSERT INTO log (message) VALUES ($1)`,
		"params": []interface{}{fmt.Sprintf("Write from %s, ts: %s", cfg.Namespace, time.Now())},
	})

	req, err := http.NewRequest("POST", target, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("failed to create post request: %v", err)
		return
	}

	if cfg.SignAuthEnabled {
		token, err := authClient.Token(cfg.InitTarget)
		if err != nil {
			log.Fatalf("failed to issue token in auth client on scope %s: %v", cfg.InitTarget, err)
			return
		}
		req.Header.Set("X-I2I-Token", token)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Do(req)

	errMsg := handlers.RespErr{}
	var respBytes []byte
	if resp != nil && resp.Body != nil {
		respBytes, _ = io.ReadAll(resp.Body)
		_ = json.Unmarshal(respBytes, &errMsg)
	}

	if err != nil {
		log.Fatalf("Initial query failed: %v; errMsg: %s", err, errMsg.Error)
		return
	}
	defer resp.Body.Close()
}
