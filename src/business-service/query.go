package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/perpetua1g0d/bmstu-diploma/business-service/config"
)

func sendDBQuery(cfg *config.Config, client *http.Client) {
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
		log.Printf("Request failed: %v; errMsg: %s", err, errMsg)
		return
	} else if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	// log.Printf("Initial query to %s status: %s; errMsg: %s", target, resp.Status, errMsg.Error)
}

// func runBenchmarks(cfg *config.Config, authClient *auth_client.AuthClient) {
// 	log.Printf("sleepeing before benchmarks...")
// 	time.Sleep(10 * time.Second)
// 	log.Printf("benchmarks started.")

// 	file, err := os.Create(benchmarksResultsFile)
// 	if err != nil {
// 		log.Fatalf("Cannot create results file: %v", err)
// 	}
// 	defer file.Close()
// 	writer := csv.NewWriter(file)
// 	defer writer.Flush()
// 	writer.Write([]string{"requests", "time_ms", "operation", "sign_enabled", "sign_disabled"})

// 	httpClient := &http.Client{
// 		Timeout: 5 * time.Second,
// 		Transport: &AuthTransport{
// 			authClient: authClient,
// 			cfg:        cfg,
// 			defaultRT:  http.DefaultTransport,
// 		},
// 	}

// 	// requestCount := []int64{1000}
// 	requestCount := []int64{100, 250, 500, 750, 1000}
// 	rerunCount := 2
// 	for _, reqCount := range requestCount {
// 		var avgTime float64 = 0
// 		for _ = range rerunCount {
// 			wg := &sync.WaitGroup{}
// 			wg.Add(int(reqCount))

// 			start := time.Now()
// 			for i := 0; i < int(reqCount); i++ {
// 				go func() {
// 					defer wg.Done()
// 					sendBenchmarkQuery(cfg, httpClient)
// 				}()
// 			}
// 			wg.Wait()

// 			duration := time.Since(start).Milliseconds()
// 			avgTime += float64(duration)
// 		}

// 		avgTime = avgTime / float64(rerunCount*int(reqCount))
// 		log.Printf("finished %d requests, avg: %f", reqCount, avgTime)
// 		writer.Write([]string{
// 			strconv.FormatInt(reqCount, 10),
// 			strconv.FormatFloat(avgTime, 'f', 2, 64),
// 			"write",
// 			fmt.Sprintf("%v", *cfg.SignAuthEnabled.Load()),
// 			fmt.Sprintf("%v", *cfg.VerifyAuthEnabled.Load()),
// 			// fmt.Sprintf("sign=%v_verify=%v", cfg.SignAuthEnabled, cfg.VerifyAuthEnabled),
// 		})
// 	}

// 	log.Printf("benchmarks finished.")
// }
