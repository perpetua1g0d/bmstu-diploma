package main

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/config"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/handlers"
)

func main() {
	cfg := config.GetConfig()

	// Автоматический начальный запрос
	if cfg.InitTarget != "" && cfg.InitQuery != "" {
		go func() {
			for {
				time.Sleep(15 * time.Second)
				sendInitialQuery(cfg)
			}
		}()
	}

	parseServiceAccountToken()

	// Настройка HTTP-сервера
	http.HandleFunc(cfg.ServiceEndpoint, handlers.QueryHandler)
	log.Printf("Starting %s on :8080 (Auth: %v)", cfg.ServiceName, cfg.AuthEnabled)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// func getJWKs() {
// 	ca1, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
// 	if err != nil {
// 		log.Fatalf("failed to read сa1: %v", err)
// 	}
// 	log.Printf("ca1 cert: %s", string(ca1))
// }

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func getPublicKeyForJWT() (*rsa.PublicKey, error) {
	// 1. Настройка TLS
	caCert, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("error reading CA cert: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	// 2. Получение JWKS
	token, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return nil, fmt.Errorf("error reading token: %w", err)
	}

	req, _ := http.NewRequest("GET", "https://kubernetes.default.svc/openid/v1/jwks", nil)
	req.Header.Add("Authorization", "Bearer "+string(token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("JWKS request failed: %w", err)
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("JWKS parse error: %w", err)
	}

	// 3. Конвертация JWK в RSA ключ
	if len(jwks.Keys) == 0 {
		return nil, errors.New("no keys in JWKS")
	}

	// Берем первый ключ (можно добавить поиск по kid)
	key := jwks.Keys[0]

	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("invalid modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("invalid exponent: %w", err)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}, nil
}

func parseServiceAccountToken() {
	tokenBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Fatalf("failed to read token: %v", err)
	}
	log.Printf("token string: %s", string(tokenBytes))

	publicKey, err := getPublicKeyForJWT()
	if err != nil {
		panic(fmt.Sprintf("Error getting public key: %v", err))
	}

	token, err := jwt.Parse(string(tokenBytes), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		log.Fatalf("parsing jwt: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println("Valid token claims:")
		for k, v := range claims {
			fmt.Printf("%s: %v\n", k, v)
		}
	} else {
		panic("Invalid token (!token.Valid)")
	}
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
