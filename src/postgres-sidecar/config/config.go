package config

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	ServiceName       string
	Namespace         string
	InitTarget        string
	InitQuery         string
	ServiceEndpoint   string
	SignAuthEnabled   atomic.Pointer[bool]
	VerifyAuthEnabled atomic.Pointer[bool]
	RunBenchmarks     bool
	JWTSecret         string
	PostgresHost      string
	PostgresPort      string
	PostgresUser      string
	PostgresPassword  string
	PostgresDB        string
	watcher           *fsnotify.Watcher
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
			InitTarget:        os.Getenv("INIT_TARGET_SERVICE"),
			InitQuery:         os.Getenv("INIT_SQL_QUERY"),
			ServiceEndpoint:   getEnv("SERVICE_ENDPOINT", "/query"),
			SignAuthEnabled:   atomic.Pointer[bool]{},
			VerifyAuthEnabled: atomic.Pointer[bool]{},
			RunBenchmarks:     getEnv("RUN_BENCHMARKS_ON_INIT", "false") == "true",
			JWTSecret:         getEnv("JWT_SECRET", "default-secret-256-bit"),
			PostgresHost:      getEnv("POSTGRES_HOST", "not_found_env_db_host"),
			PostgresPort:      getEnv("POSTGRES_PORT", "5432"),
			PostgresUser:      getEnv("POSTGRES_USER", "not_found_env_db_user"),
			PostgresPassword:  getEnv("POSTGRES_PASSWORD", ""),
			PostgresDB:        getEnv("POSTGRES_DB", "not_found_postgres_db"),
		}

		currentSignEnabled := getEnv("SIGN_AUTH_ENABLED", "false") == "true"
		currentVerifyEnabled := getEnv("VERIFY_AUTH_ENABLED", "false") == "true"
		instance.SignAuthEnabled.Store(&currentSignEnabled)
		instance.VerifyAuthEnabled.Store(&currentVerifyEnabled)

		registerAuthWatchers(instance)
	})
	log.Printf("sidecar config: %v", instance)
	return instance
}

func (c *Config) updateAuthConfig() {
	signContent, _ := os.ReadFile("/etc/auth-config/SIGN_AUTH_ENABLED")
	verifyContent, _ := os.ReadFile("/etc/auth-config/VERIFY_AUTH_ENABLED")

	newSign := strings.TrimSpace(string(signContent)) == "true"
	newVerify := strings.TrimSpace(string(verifyContent)) == "true"

	instance.SignAuthEnabled.Store(&newSign)
	instance.VerifyAuthEnabled.Store(&newVerify)
	log.Printf("Настройки обновлены: SIGN=%v, VERIFY=%v", newSign, newVerify)
}

func (c *Config) RealodHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received config reload request")

	var data struct {
		Sign   bool `json:"sign"`
		Verify bool `json:"verify"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		log.Printf("JSON decode error: %v", err)
		return
	}

	// атомарно обновляем значения
	instance.SignAuthEnabled.Store(&data.Sign)
	instance.VerifyAuthEnabled.Store(&data.Verify)

	log.Printf("Settings updated via HTTP: SIGN=%v, VERIFY=%v", data.Sign, data.Verify)
	w.WriteHeader(http.StatusOK)
}

func (c *Config) Close() {
	if c.watcher != nil {
		c.watcher.Close()
	}
}

func registerAuthWatchers(cfg *Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	cfg.watcher = watcher

	// Добавляем файлы для отслеживания
	files := []string{
		"/etc/auth-config/SIGN_AUTH_ENABLED",
		"/etc/auth-config/VERIFY_AUTH_ENABLED",
	}

	for _, file := range files {
		if err := watcher.Add(file); err != nil {
			log.Printf("Failed to watch %s: %v", file, err)
		} else {
			log.Printf("Watching file for changes: %s", file)
		}
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("File modified: %s", event.Name)
					cfg.updateAuthConfig()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
