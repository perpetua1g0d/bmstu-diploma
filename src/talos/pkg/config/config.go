package config

type Config struct {
	Address string
	Issuer  string
}

func Load() *Config {
	return &Config{
		Address: ":8080",
		Issuer:  "http://talos.talos.svc.cluster.local",
	}
}
