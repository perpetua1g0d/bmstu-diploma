package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/perpetua1g0d/bmstu-diploma/business-service/config"
	auth_signer "github.com/perpetua1g0d/bmstu-diploma/src/auth-client/pkg/signer"
)

type Service struct {
	cfg         *config.Config
	signer      *auth_signer.TokenSigner
	httpClient  *http.Client
	authEnabled *atomic.Bool
	benchmark   *BenchmarkState
}

type BenchmarkState struct {
	mu            sync.Mutex
	running       bool
	wg            sync.WaitGroup
	results       *BenchmarkResults
	startTime     time.Time
	counter       int64
	totalRequests int
	concurrency   int
	queryType     string
	delay         time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
}

type BenchmarkResults struct {
	mu             sync.Mutex
	TotalRequests  int64
	SuccessCount   int64
	ErrorCount     int64
	TotalDuration  time.Duration
	MinDuration    time.Duration
	MaxDuration    time.Duration
	SumDurations   time.Duration
	StatusCounters map[int]int64
}

func NewService(cfg *config.Config, signer *auth_signer.TokenSigner) *Service {
	authTransport := auth_signer.NewAuthTransport(signer, cfg.InitTarget)

	s := &Service{
		cfg:    cfg,
		signer: signer,
		httpClient: &http.Client{
			Timeout:   5 * time.Second,
			Transport: authTransport,
		},
		authEnabled: &atomic.Bool{},
		benchmark: &BenchmarkState{
			results: &BenchmarkResults{
				StatusCounters: make(map[int]int64),
				MinDuration:    time.Hour,
			},
		},
	}
	s.authEnabled.Store(cfg.SignAuthEnabled)

	return s
}

func (s *Service) sendBenchmarkQuery(queryType string) (int, error) {
	var sql string
	var params []interface{}

	switch queryType {
	case "heavy":
		sql = `WITH heavy_cte AS (SELECT generate_series(1,1000000) AS data)
               SELECT COUNT(*), AVG(data) FROM heavy_cte`
	default: // "light"
		sql = `INSERT INTO log (message) VALUES ($1)`
		params = []interface{}{fmt.Sprintf("Benchmark at %s", time.Now())}
	}

	target := fmt.Sprintf("http://%s:%s%s",
		s.cfg.PostgresService,
		s.cfg.SidecarPort,
		s.cfg.ServiceEndpoint,
	)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"sql":    sql,
		"params": params,
	})

	req, err := http.NewRequest("POST", target, bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return resp.StatusCode, nil
}
