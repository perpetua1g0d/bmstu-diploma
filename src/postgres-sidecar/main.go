package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/config"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/handlers"
)

const talosAddress = "http://talos.talos.svc.cluster.local:80"

func main() {
	cfg := config.GetConfig()

	// Автоматический начальный запрос
	if cfg.InitTarget != "" && cfg.InitQuery != "" {
		go func() {
			for {
				tokenExchange()
				time.Sleep(5 * time.Second)
				sendInitialQuery(cfg)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	// Настройка HTTP-сервера
	http.HandleFunc(cfg.ServiceEndpoint, handlers.QueryHandler)
	log.Printf("Starting %s on :8080 (Auth: %v)", cfg.ServiceName, cfg.AuthEnabled)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func tokenExchange() {
	k8sToken := ""
	certs, _ := getTalosCerts()
	log.Printf("got talos certs: %v", certs)
	token := getTalosToken(k8sToken)
	log.Printf("got talos token: %s", token)
}

func getTalosCerts() (jose.JSONWebKeySet, error) {
	const talosCertEndpoint = talosAddress + "/realms/infra2infra/protocol/openid-connect/certs"

	req, err := http.NewRequest(http.MethodGet, talosCertEndpoint, nil)
	if err != nil {
		log.Fatalf("failed to create talos certs request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)

	var respBytes []byte
	if resp != nil && resp.Body != nil {
		respBytes, _ = io.ReadAll(resp.Body)
	}
	if err != nil {
		log.Fatalf("failed to get talos certs: %v; respBody: %s", err, string(respBytes))
	}
	defer resp.Body.Close()

	var jwks jose.JSONWebKeySet
	if marshalErr := json.Unmarshal(respBytes, &jwks); marshalErr != nil {
		log.Fatalf("failed to unmarshal certs: %v; body: %s", marshalErr, string(respBytes))
	}

	return jwks, nil
}

func getTalosToken(k8sToken string) string {
	const talosTokenEndpoint = talosAddress + "/realms/infra2infra/protocol/openid-connect/token"

	v := url.Values{}
	v.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	v.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt:kubernetes")
	v.Set("subject_token", k8sToken)
	v.Set("scope", os.Getenv("INIT_TARGET_SERVICE"))
	body := v.Encode()

	req, err := http.NewRequest(http.MethodPost, talosTokenEndpoint, strings.NewReader(body))
	if err != nil {
		log.Fatalf("failed to create talos token request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)

	var respBytes []byte
	if resp != nil && resp.Body != nil {
		respBytes, _ = io.ReadAll(resp.Body)
	}
	if err != nil {
		log.Fatalf("failed to get talos token: %v; respBody: %s", err, string(respBytes))
	}
	defer resp.Body.Close()

	var tokenResp map[string]string
	if marshalErr := json.Unmarshal(respBytes, &tokenResp); marshalErr != nil {
		log.Fatalf("failed to unmarshal token: %v; body: %s", marshalErr, string(respBytes))
	}

	token := tokenResp["access_token"]
	log.Printf("Talos response token exp_in: %s", tokenResp["expires_in"])

	return token
}

func sendInitialQuery(cfg *config.Config) {
	target := fmt.Sprintf("http://%s.%s.svc.cluster.local:8080%s",
		cfg.InitTarget,
		cfg.InitTarget,
		cfg.ServiceEndpoint,
	)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"sql":    `INSERT INTO log (message) VALUES ($1)`,
		"params": []interface{}{fmt.Sprintf("Init from %s, ts: %s", cfg.Namespace, time.Now())},
	})

	req, _ := http.NewRequest("POST", target, bytes.NewBuffer(reqBody))
	if cfg.AuthEnabled {
		req.Header.Set("X-I2I-Token", os.Getenv("TOKEN_JWT"))
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	errMsg := handlers.RespErr{}
	respBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(respBytes, &errMsg)
	if err != nil {
		log.Printf("Initial query failed: %v; errMsg: %s", err, errMsg.Error)
		return
	}
	defer resp.Body.Close()

	log.Printf("Initial query to %s status: %s; errMsg: %s", target, resp.Status, errMsg.Error)
}

func determineOperationType(r *http.Request) string {
	// This is a simplified version - actual implementation would depend on the service
	serviceType := os.Getenv("SERVICE_TYPE")

	switch serviceType {
	case "postgresql":
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			return "read"
		}
		return "write"
	case "kafka":
		// Kafka would have specific endpoints for produce/consume
		if strings.Contains(r.URL.Path, "/produce") {
			return "write"
		}
		return "read"
	case "redis":
		// Redis commands would need to be parsed from request
		return "unknown" // TODO: Implement Redis command parsing
	default:
		return "unknown"
	}
}
