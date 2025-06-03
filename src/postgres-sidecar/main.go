package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	auth_verifier "github.com/perpetua1g0d/bmstu-diploma/auth-client/pkg/verifier"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/config"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/handlers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	dbSizeBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "postgres_db_size_bytes",
		Help: "Size of PostgreSQL database in bytes",
	}, []string{"database"})

	dbIdleConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "postgres_idle_connections",
		Help: "Number of idle connections to the PostgreSQL database",
	}, []string{"database"})

	dbOpenConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "postgres_open_connections",
		Help: "Number of opened connections to the PostgreSQL database",
	}, []string{"database"})
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()

	verifier, err := auth_verifier.NewVerifier(ctx, cfg.ServiceName, cfg.VerifyAuthEnabled)
	if err != nil {
		log.Fatalf("failed to create verifier: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/reload_config", verifier.NewRealodHandler())

	mux.Handle("/metrics", promhttp.Handler())

	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDB,
	))
	if err != nil {
		log.Fatalf("failed to connect to local db: %v", err)
	}
	defer db.Close()

	go collectDBMetrics(db, cfg.PostgresDB)

	mux.HandleFunc(cfg.ServiceEndpoint, handlers.NewQueryHandler(ctx, cfg, db, verifier))

	log.Printf("Starting %s on :8080 (verify: %v)", cfg.ServiceName, cfg.VerifyAuthEnabled)
	log.Fatal(http.ListenAndServe(":8080", mux)) // root handler promhttp.Handler()
}

func collectDBMetrics(db *sql.DB, dbName string) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		updateDBMetrics(db, dbName)
	}
}

func updateDBMetrics(db *sql.DB, dbName string) {
	var size int64
	if err := db.QueryRow("SELECT pg_database_size($1)", dbName).Scan(&size); err != nil {
		log.Printf("failed to collect pg_database_size in %s: %v", dbName, err)
	}

	dbSizeBytes.WithLabelValues(dbName).Set(float64(size))
	dbIdleConnections.WithLabelValues(dbName).Set(float64(db.Stats().Idle))
	dbOpenConnections.WithLabelValues(dbName).Set(float64(db.Stats().OpenConnections))
}
