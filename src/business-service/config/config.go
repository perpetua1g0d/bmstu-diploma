package config

import (
	"log"
	"os"
	"sync"
)

type Config struct {
	ServiceName     string
	Namespace       string
	PostgresService string
	InitTarget      string
	SidecarPort     string
	ServiceEndpoint string
	SignAuthEnabled bool
}

var (
	instance *Config
	once     sync.Once
)

func NewConfig() *Config {
	once.Do(func() {
		instance = &Config{
			ServiceName:     getEnv("SERVICE_NAME", "business-service"),
			Namespace:       getEnv("POD_NAMESPACE", "default"),
			PostgresService: getEnv("POSTGRES_SERVICE", ""),
			InitTarget:      getEnv("INIT_TARGET", ""),
			SidecarPort:     getEnv("SIDECAR_PORT", "8080"),
			ServiceEndpoint: getEnv("SERVICE_ENDPOINT", "/query"),
			SignAuthEnabled: getEnv("SIGN_AUTH_ENABLED", "true") == "true",
		}
	})
	log.Printf("Service config initialized: %+v", instance)
	return instance
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
