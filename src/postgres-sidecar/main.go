package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	auth_client "github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/auth/client"
	auth_config "github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/auth/config"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/config"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/handlers"
)

const talosAddress = "http://talos.talos.svc.cluster.local:80"

func main() {
	ctx := context.Background()
	cfg := config.GetConfig()

	authClient, err := createAuthClient(ctx, cfg, []string{cfg.InitTarget})
	if err != nil {
		log.Fatalf("failed to create auth client: %v", err)
	}

	if cfg.InitTarget != "" && cfg.InitQuery != "" {
		go func() {
			for {
				// tokenExchange()
				time.Sleep(5 * time.Second)
				sendInitialQuery(cfg, authClient)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	http.HandleFunc(cfg.ServiceEndpoint, handlers.NewQueryHandler(ctx, authClient))
	log.Printf("Starting %s on :8080 (Auth sign: %v, verify: %v)", cfg.ServiceName, cfg.SignAuthEnabled, cfg.VerifyAuthEnabled)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createAuthClient(ctx context.Context, cfg *config.Config, scopes []string) (*auth_client.AuthClient, error) {
	authCfg := &auth_config.Config{
		ClientID:              cfg.ServiceName,
		SignEnabled:           cfg.SignAuthEnabled,
		VerifyEnabled:         cfg.VerifyAuthEnabled,
		TokenEndpointAddress:  talosAddress + "/realms/infra2infra/protocol/openid-connect/token",
		CertsEndpointAddress:  talosAddress + "/realms/infra2infra/protocol/openid-connect/certs",
		ConfigEndpointAddress: talosAddress + "/realms/infra2infra/.well-known/openid-configuration",
		RequestTimeout:        5 * time.Second,
		ErrTokenBackoff:       1 * time.Minute,
	}

	log.Printf("auth config: %v", authCfg)

	return auth_client.NewAuthClient(ctx, authCfg, scopes)
}

func sendInitialQuery(cfg *config.Config, authClient *auth_client.AuthClient) {
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
	if cfg.SignAuthEnabled {
		token, err := authClient.Token(cfg.InitTarget)
		if err != nil {
			log.Printf("failed to issue token in auth client on scope %s: %v", cfg.InitTarget, err)
			return
		}
		req.Header.Set("X-I2I-Token", token)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 4 * time.Second}
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
			return "RO"
		}
		return "RW"
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
