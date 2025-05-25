package config

import (
	"log"
	"os"
	"sync"
)

type Config struct {
	ServiceName       string
	Namespace         string
	InitTarget        string
	InitQuery         string
	ServiceEndpoint   string
	SignAuthEnabled   bool
	VerifyAuthEnabled bool
	JWTSecret         string
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

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{
			ServiceName:       getEnv("SERVICE_NAME", "postgres"),
			Namespace:         getEnv("POD_NAMESPACE", "default"),
			InitTarget:        os.Getenv("INIT_TARGET_SERVICE"),
			InitQuery:         os.Getenv("INIT_SQL_QUERY"),
			ServiceEndpoint:   getEnv("SERVICE_ENDPOINT", "/query"),
			SignAuthEnabled:   getEnv("SIGN_AUTH_ENABLED", "false") == "true",
			VerifyAuthEnabled: getEnv("VERIFY_AUTH_ENABLED", "false") == "true",
			JWTSecret:         getEnv("JWT_SECRET", "default-secret-256-bit"),
			PostgresHost:      getEnv("POSTGRES_HOST", "not_found_env_db_host"),
			PostgresPort:      getEnv("POSTGRES_PORT", "5432"),
			PostgresUser:      getEnv("POSTGRES_USER", "not_found_env_db_user"),
			PostgresPassword:  getEnv("POSTGRES_PASSWORD", ""),
			PostgresDB:        getEnv("POSTGRES_DB", "not_found_postgres_db"),
		}
	})
	log.Printf("sidecar config: %v", instance)
	return instance
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
