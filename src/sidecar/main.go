package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

type Config struct {
	AuthEnabled bool `json:"auth_enabled"`
}

type OIDCConfig struct {
	Issuer        string `json:"issuer"`
	JWKSURI       string `json:"jwks_uri"`
	TokenEndpoint string `json:"token_endpoint"`
}

var (
	serviceConfig Config
	oidcConfig    OIDCConfig
)

func init() {
	// Load configuration (simplified for demo)
	serviceConfig.AuthEnabled = os.Getenv("AUTH_ENABLED") == "true"

	// Mock OIDC configuration
	oidcConfig = OIDCConfig{
		Issuer:        "http://talos.default.svc.cluster.local",
		JWKSURI:       "http://talos.default.svc.cluster.local/protocol/openid-connect/certs",
		TokenEndpoint: "http://talos.default.svc.cluster.local/protocol/openid-connect/token",
	}
}

func main() {
	r := mux.NewRouter()

	// Admin endpoints
	r.HandleFunc("/admin/config", handleConfig).Methods("GET", "POST")
	r.HandleFunc("/admin/auth/toggle", handleAuthToggle).Methods("POST")

	// OIDC discovery endpoints
	r.HandleFunc("/.well-known/openid-configuration", handleOIDCConfiguration).Methods("GET")
	r.HandleFunc("/protocol/openid-connect/certs", handleJWKS).Methods("GET")
	r.HandleFunc("/protocol/openid-connect/token", handleToken).Methods("POST")

	// Proxy endpoints
	r.PathPrefix("/").HandlerFunc(handleProxy)

	port := os.Getenv("SIDECAR_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Sidecar server started on port %s, auth enabled: %v", port, serviceConfig.AuthEnabled)
	log.Fatal(http.ListenAndServe(":"+port, r))
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

func handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		json.NewEncoder(w).Encode(serviceConfig)
	case "POST":
		var newConfig Config
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		serviceConfig = newConfig
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAuthToggle(w http.ResponseWriter, r *http.Request) {
	serviceConfig.AuthEnabled = !serviceConfig.AuthEnabled
	json.NewEncoder(w).Encode(map[string]bool{"auth_enabled": serviceConfig.AuthEnabled})
}

func handleOIDCConfiguration(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(oidcConfig)
}

func handleJWKS(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement actual JWKS endpoint
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"keys":[]}`))
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement actual token endpoint
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"access_token":"dummy","token_type":"bearer"}`))
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
	// Determine operation type based on request
	operationType := determineOperationType(r)
	log.Printf("Incoming request: %s %s, operation: %s", r.Method, r.URL.Path, operationType)

	if serviceConfig.AuthEnabled {
		// Check authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		// TODO: Validate token signature with OIDC provider
		// For now just log it
		log.Printf("Token received: %s", token)

		// Получаем токен сервисного аккаунта
		// k8sToken, err := getK8SToken()
		// if err != nil {
		// 	log.Printf("Error getting k8s token: %v", err)
		// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
		// 	return
		// }

		// // Используем токен для запроса к idP
		// req, _ := http.NewRequest("POST", oidcConfig.TokenEndpoint, nil)
		// req.Header.Set("Authorization", "Bearer "+k8sToken)

		// TODO: Check token claims for required permissions based on operationType
	}

	// TODO: Forward request to actual service
	// For now just respond with dummy data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "Request would be forwarded to service",
		"operation":    operationType,
		"auth_enabled": fmt.Sprintf("%v", serviceConfig.AuthEnabled),
	})
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
