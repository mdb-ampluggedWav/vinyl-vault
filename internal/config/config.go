package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
	Environment string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "host:localhost user=postgres password=postgres dbname=vinylvault port=5432 sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
