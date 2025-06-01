package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	auth_client "github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/auth/client"
	auth_config "github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/auth/config"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/config"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/handlers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const idpAddress = "http://idp.idp.svc.cluster.local:80"

const benchmarksResultsFile = "/var/log/results.csv"

var (
	tokenSignedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "client_token_sign_requests_total",
		Help: "Total number of requests signed with token",
	}, []string{"scope", "result", "enabled", "service_name"})
	tokenSignDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "client_token_sign_duration_milliseconds",
		Help:    "Duration of token signing in milliseconds",
		Buckets: []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000},
	}, []string{"scope", "result", "enabled", "service_name"})
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()
	defer cfg.Close()

	authClient, err := createAuthClient(ctx, cfg, []string{cfg.InitTarget})
	if err != nil {
		log.Fatalf("failed to create auth client: %v", err)
	}

	go func() {
		time.Sleep(30 * time.Second)
		httpClient := &http.Client{
			Timeout: 5 * time.Second,
			Transport: &AuthTransport{
				authClient: authClient,
				cfg:        cfg,
				defaultRT:  http.DefaultTransport,
			},
		}
		for {
			time.Sleep(10 * time.Second)
			sendBenchmarkQuery(cfg, httpClient)
		}
	}()

	if cfg.RunBenchmarks {
		go func() {
			// runBenchmarks(cfg, authClient)
		}()
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/reload-config", cfg.RealodHandler)
	mux.HandleFunc(cfg.ServiceEndpoint, handlers.NewQueryHandler(ctx, cfg, authClient))

	log.Printf("Starting %s on :8080 (Auth sign: %v, verify: %v)", cfg.ServiceName, *cfg.SignAuthEnabled.Load(), *cfg.VerifyAuthEnabled.Load())
	log.Fatal(http.ListenAndServe(":8080", mux)) // root handler promhttp.Handler()
}

func runBenchmarks(cfg *config.Config, authClient *auth_client.AuthClient) {
	log.Printf("sleepeing before benchmarks...")
	time.Sleep(10 * time.Second)
	log.Printf("benchmarks started.")

	file, err := os.Create(benchmarksResultsFile)
	if err != nil {
		log.Fatalf("Cannot create results file: %v", err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"requests", "time_ms", "operation", "sign_enabled", "sign_disabled"})

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &AuthTransport{
			authClient: authClient,
			cfg:        cfg,
			defaultRT:  http.DefaultTransport,
		},
	}

	// requestCount := []int64{1000}
	requestCount := []int64{100, 250, 500, 750, 1000}
	rerunCount := 2
	for _, reqCount := range requestCount {
		var avgTime float64 = 0
		for _ = range rerunCount {
			wg := &sync.WaitGroup{}
			wg.Add(int(reqCount))

			start := time.Now()
			for i := 0; i < int(reqCount); i++ {
				go func() {
					defer wg.Done()
					sendBenchmarkQuery(cfg, httpClient)
				}()
			}
			wg.Wait()

			duration := time.Since(start).Milliseconds()
			avgTime += float64(duration)
		}

		avgTime = avgTime / float64(rerunCount*int(reqCount))
		log.Printf("finished %d requests, avg: %f", reqCount, avgTime)
		writer.Write([]string{
			strconv.FormatInt(reqCount, 10),
			strconv.FormatFloat(avgTime, 'f', 2, 64),
			"write",
			fmt.Sprintf("%v", *cfg.SignAuthEnabled.Load()),
			fmt.Sprintf("%v", *cfg.VerifyAuthEnabled.Load()),
			// fmt.Sprintf("sign=%v_verify=%v", cfg.SignAuthEnabled, cfg.VerifyAuthEnabled),
		})
	}

	log.Printf("benchmarks finished.")
}

func createAuthClient(ctx context.Context, cfg *config.Config, scopes []string) (*auth_client.AuthClient, error) {
	authCfg := &auth_config.Config{
		ClientID: cfg.ServiceName,
		// SignEnabled:           cfg.SignAuthEnabled,
		// VerifyEnabled:         cfg.VerifyAuthEnabled,
		TokenEndpointAddress:  idpAddress + "/realms/infra2infra/protocol/openid-connect/token",
		CertsEndpointAddress:  idpAddress + "/realms/infra2infra/protocol/openid-connect/certs",
		ConfigEndpointAddress: idpAddress + "/realms/infra2infra/.well-known/openid-configuration",
		RequestTimeout:        5 * time.Second,
		ErrTokenBackoff:       10 * time.Second,
	}

	log.Printf("auth config: %v", authCfg)

	return auth_client.NewAuthClient(ctx, authCfg, scopes)
}

func sendBenchmarkQuery(cfg *config.Config, client *http.Client) {
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
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	var errMsg map[string]any
	var respBytes []byte
	if resp != nil && resp.Body != nil {
		respBytes, _ = io.ReadAll(resp.Body)
		_ = json.Unmarshal(respBytes, &errMsg)
	}

	if err != nil {
		log.Printf("Initial query failed: %v; errMsg: %s", err, errMsg)
		return
	} else if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	// log.Printf("Initial query to %s status: %s; errMsg: %s", target, resp.Status, errMsg.Error)
}

type AuthTransport struct {
	authClient *auth_client.AuthClient
	cfg        *config.Config

	defaultRT http.RoundTripper
}

func (t *AuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var signEnabled bool
	var signResult = "ok"
	scope := t.cfg.InitTarget

	signEnabled = getSignAuth(t.cfg)
	signStart := time.Now()
	if signEnabled {
		token, err := t.authClient.Token(scope)
		if err != nil {
			log.Printf("failed to issue token in auth client on scope %s: %v", scope, err)
			signResult = "token_error"
		} else {
			r.Header.Set("X-I2I-Token", token)
		}
	}
	signDuration := float64(time.Since(signStart).Milliseconds())

	tokenSignedTotal.WithLabelValues(scope, signResult, strconv.FormatBool(signEnabled), t.cfg.ServiceName).Inc()
	tokenSignDuration.WithLabelValues(scope, signResult, strconv.FormatBool(signEnabled), t.cfg.ServiceName).Observe(signDuration)

	return t.defaultRT.RoundTrip(r)
}

func getSignAuth(cfg *config.Config) bool {
	loaded := cfg.SignAuthEnabled.Load()
	if loaded == nil {
		log.Printf("config pointer[SignAuthEnabled] is empty! Sign is enabled as fallback")
		return true
	}

	return *loaded
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
