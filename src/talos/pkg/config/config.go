package config

import "time"

type Config struct {
	Address  string
	Issuer   string
	TokenTTL time.Duration
}

func Load() *Config {
	return &Config{
		Address:  ":8080",
		Issuer:   "http://talos.talos.svc.cluster.local",
		TokenTTL: time.Hour,
	}
}
