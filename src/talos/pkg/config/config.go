package config

type Config struct {
	Address string
}

func Load() *Config {
	return &Config{
		Address: ":8080",
	}
}
