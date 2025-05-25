func runBenchmarks(cfg *config.Config, authClient *auth_client.AuthClient) {
	file, err := os.Create(benchmarksResultsFile)
	if err != nil {
		log.Fatalf("Cannot create results file: %v", err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"requests", "time_ms", "operation", "sign_enabled", "sign_disabled"})

	requestCount := []int64{100, 250, 500, 750, 1000}
	rerunCount := 10
	for _, reqCount := range requestCount {
		var avgTime float64 = 0
		for _ = range rerunCount {
			wg := &sync.WaitGroup{}
			wg.Add(int(reqCount))

			start := time.Now()
			for i := 0; i < int(reqCount); i++ {
				go func() {
					defer wg.Done()
					sendBenchmarkQuery(cfg, authClient)
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
			fmt.Sprintf("%v", cfg.SignAuthEnabled),
			fmt.Sprintf("%v", cfg.VerifyAuthEnabled),
		})
	}
}