package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/config"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/handlers"
)

func main() {
	cfg := config.GetConfig()

	// Автоматический начальный запрос
	if cfg.InitTarget != "" && cfg.InitQuery != "" {
		go func() {
			time.Sleep(5 * time.Second)
			sendInitialQuery(cfg)
		}()
	}

	// Настройка HTTP-сервера
	http.HandleFunc(cfg.ServiceEndpoint, handlers.QueryHandler)
	log.Printf("Starting %s on :8080 (Auth: %v)", cfg.ServiceName, cfg.AuthEnabled)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sendInitialQuery(cfg *config.Config) {
	target := fmt.Sprintf("http://%s.%s.svc.cluster.local:8080%s",
		cfg.InitTarget,
		cfg.Namespace,
		cfg.ServiceEndpoint,
	)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"sql":    cfg.InitQuery,
		"params": []interface{}{},
	})

	req, _ := http.NewRequest("POST", target, bytes.NewBuffer(reqBody))
	if cfg.AuthEnabled {
		req.Header.Set("X-I2I-Token", os.Getenv("TOKEN_JWT"))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Initial query failed: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Initial query to %s status: %s", target, resp.Status)
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
