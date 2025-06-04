package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type BenchmarkRequest struct {
	Requests    int           `json:"requests"`
	Concurrency int           `json:"concurrency"`
	QueryType   string        `json:"query_type"`
	Delay       time.Duration `json:"delay"`
}

type BenchmarkResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

const resultsCSVFile = "/var/log/results.csv"

func (s *Service) startBenchmark(req BenchmarkRequest) error {
	s.benchmark.mu.Lock()
	defer s.benchmark.mu.Unlock()

	if s.benchmark.running {
		return fmt.Errorf("benchmark already running")
	}

	s.benchmark.running = true
	s.benchmark.startTime = time.Now()
	s.benchmark.totalRequests = req.Requests
	s.benchmark.concurrency = req.Concurrency
	s.benchmark.queryType = req.QueryType
	s.benchmark.delay = req.Delay
	atomic.StoreInt64(&s.benchmark.counter, 0)
	s.benchmark.results = &BenchmarkResults{
		StatusCounters: make(map[int]int64),
		MinDuration:    time.Hour,
	}
	s.benchmark.ctx, s.benchmark.cancel = context.WithCancel(context.Background())
	s.benchmark.wg.Add(req.Concurrency)

	for i := 0; i < req.Concurrency; i++ {
		go s.worker()
	}

	go func() {
		s.benchmark.wg.Wait()
		s.stopBenchmark()
		log.Printf("Benchmark succefully finished, results in %s", resultsCSVFile)
	}()

	return nil
}

func (s *Service) worker() {
	defer s.benchmark.wg.Done()

	for {
		if s.benchmark.ctx.Err() != nil {
			return
		}

		current := atomic.AddInt64(&s.benchmark.counter, 1)
		if current > int64(s.benchmark.totalRequests) {
			return
		}

		start := time.Now()
		status, err := s.sendBenchmarkQuery(s.benchmark.queryType)
		duration := time.Since(start)

		s.benchmark.results.mu.Lock()
		if err != nil || status < 200 || status >= 300 {
			s.benchmark.results.ErrorCount++
		} else {
			s.benchmark.results.SuccessCount++
		}
		if duration < s.benchmark.results.MinDuration {
			s.benchmark.results.MinDuration = duration
		}
		if duration > s.benchmark.results.MaxDuration {
			s.benchmark.results.MaxDuration = duration
		}
		s.benchmark.results.SumDurations += duration
		s.benchmark.results.StatusCounters[status]++
		s.benchmark.results.mu.Unlock()

		select {
		case <-time.After(s.benchmark.delay):
		case <-s.benchmark.ctx.Done():
			return
		}
	}
}

func (s *Service) stopBenchmark() {
	s.benchmark.mu.Lock()
	defer s.benchmark.mu.Unlock()

	if !s.benchmark.running {
		return
	}

	s.benchmark.cancel()
	s.benchmark.wg.Wait()
	s.benchmark.running = false

	s.writeResultsToCSV()
}

func (s *Service) writeResultsToCSV() {
	file, err := os.OpenFile(
		resultsCSVFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Printf("Error opening results file: %v", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	duration := time.Since(s.benchmark.startTime)
	avgDuration := time.Duration(0)
	if s.benchmark.results.SuccessCount > 0 {
		avgDuration = time.Duration(
			int64(s.benchmark.results.SumDurations) / s.benchmark.results.SuccessCount,
		)
	}
	rps := float64(s.benchmark.totalRequests) / duration.Seconds()

	record := []string{
		s.benchmark.startTime.Format(time.RFC3339),
		fmt.Sprintf("%d", s.benchmark.totalRequests),
		fmt.Sprintf("%d", s.benchmark.concurrency),
		s.benchmark.queryType,
		fmt.Sprintf("%d", s.benchmark.results.SuccessCount),
		fmt.Sprintf("%d", s.benchmark.results.ErrorCount),
		fmt.Sprintf("%.2f", rps),
		avgDuration.String(),
		s.benchmark.results.MinDuration.String(),
		s.benchmark.results.MaxDuration.String(),
	}

	if err := writer.Write(record); err != nil {
		log.Printf("Error writing to CSV: %v", err)
	}
}

func (s *Service) benchmarkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BenchmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.startBenchmark(req); err != nil {
		respondError(w, fmt.Sprintf("Start failed: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(BenchmarkResponse{
		Status:  "started",
		Message: fmt.Sprintf("running %d requests", req.Requests),
	})
	log.Printf("Benchmarking started")
}

func (s *Service) stopBenchmarkHandler(w http.ResponseWriter, r *http.Request) {
	if !s.benchmark.running {
		json.NewEncoder(w).Encode(BenchmarkResponse{
			Status:  "idle",
			Message: "Исследование не проводится.",
		})
		return
	}

	s.stopBenchmark()
	json.NewEncoder(w).Encode(BenchmarkResponse{
		Status:  "stopped",
		Message: "Исследование закончилось.",
	})
	log.Printf("Benchmarking stopped")
}

func respondError(w http.ResponseWriter, message string, code int) {
	log.Printf("error: %s, code: %d", message, code)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(BenchmarkResponse{
		Status:  "error",
		Message: message,
	})
}
