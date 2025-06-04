package config

import (
	"log"
	"os"
	"sync"
)

type Config struct {
	ServiceName       string
	Namespace         string
	ServiceEndpoint   string
	VerifyAuthEnabled bool
	RunBenchmarks     bool
	PostgresHost      string
	PostgresPort      string
	PostgresUser      string
	PostgresPassword  string
	PostgresDB        string
}

var (
	instance *Config
	once     sync.Once
)

func NewConfig() *Config {
	once.Do(func() {
		instance = &Config{
			ServiceName:       getEnv("SERVICE_NAME", "postgres"),
			Namespace:         getEnv("POD_NAMESPACE", "default"),
			ServiceEndpoint:   getEnv("SERVICE_ENDPOINT", "/query"),
			VerifyAuthEnabled: getEnv("VERIFY_AUTH_ENABLED", "false") == "true",
			RunBenchmarks:     getEnv("RUN_BENCHMARKS_ON_INIT", "false") == "true",
			PostgresHost:      getEnv("POSTGRES_HOST", "not_found_env_db_host"),
			PostgresPort:      getEnv("POSTGRES_PORT", "5432"),
			PostgresUser:      getEnv("POSTGRES_USER", "not_found_env_db_user"),
			PostgresPassword:  getEnv("POSTGRES_PASSWORD", ""),
			PostgresDB:        getEnv("POSTGRES_DB", "not_found_postgres_db"),
		}
	})
	log.Printf("sidecar config: %+v", instance)
	return instance
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
