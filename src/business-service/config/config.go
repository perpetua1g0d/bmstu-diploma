package config

import (
	"log"
	"os"
	"sync"
)

type Config struct {
	ServiceName      string
	Namespace        string
	InitTarget       string
	InitQuery        string
	ServiceEndpoint  string
	SignAuthEnabled  bool
	RunBenchmarks    bool
	JWTSecret        string
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
}

var (
	instance *Config
	once     sync.Once
)

func NewConfig() *Config {
	once.Do(func() {
		instance = &Config{
			ServiceName:     getEnv("SERVICE_NAME", "postgres"),
			Namespace:       getEnv("POD_NAMESPACE", "default"),
			InitTarget:      os.Getenv("INIT_TARGET_SERVICE"),
			InitQuery:       os.Getenv("INIT_SQL_QUERY"),
			ServiceEndpoint: getEnv("SERVICE_ENDPOINT", "/query"),
			SignAuthEnabled: getEnv("SIGN_AUTH_ENABLED", "false") == "true",
			RunBenchmarks:   getEnv("RUN_BENCHMARKS_ON_INIT", "false") == "true",
			JWTSecret:       getEnv("JWT_SECRET", "default-secret-256-bit"),
		}
	})
	log.Printf("service config: %v", instance)
	return instance
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
