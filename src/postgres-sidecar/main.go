package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
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

func tokenExchange() {
	k8sToken, err := getK8SToken()
	if err != nil {
		log.Fatalf("failed to get k8s token: %v", err)
	}

	certs, _ := getTalosCerts()
	log.Printf("got talos certs: %v", certs)
	token := getTalosToken(k8sToken)
	log.Printf("got talos token: %s", token)

	claims, err := verifyToken(token, certs)
	if err != nil {
		log.Fatalf("verify error: %v", err)
	}

	log.Printf("got claims: %v", claims)
}

func verifyToken(rawToken string, certs jose.JSONWebKeySet) (map[string]any, error) {
	token, err := jwt.ParseSigned(rawToken)
	if err != nil {
		log.Printf("failed to parse token: %v", err)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	var claims map[string]interface{}
	for _, header := range token.Headers {
		keys := certs.Key(header.KeyID)
		if len(keys) == 0 {
			continue
		}

		for _, key := range keys {
			if err := token.Claims(key.Public(), &claims); err == nil {
				return claims, nil
			}
		}
	}

	log.Printf("no certificate found to parse token. certs: %v, tokenHeaders: %v", certs, token.Headers)
	return nil, fmt.Errorf("no certificate found to parse token")
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
		}
		req.Header.Set("X-I2I-Token", token)
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

// Функция для получения Kubernetes SA токена
func getK8SToken() (string, error) {
	tokenPath := filepath.Join("/var/run/secrets/kubernetes.io/serviceaccount", "token")
	token, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("failed to read k8s token: %v", err)
	}
	return string(token), nil
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
