// Package config : réglages lus depuis l'environnement, jamais en dur dans le code.
package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
}

func Load() Config {
	return Config{
		Port: getenv("PORT", "8080"),
		// valeurs par défaut = celles du docker-compose de dev
		DatabaseURL: getenv("DATABASE_URL",
			"postgres://portal:portal@localhost:5432/portal?sslmode=disable"),
		RedisURL: getenv("REDIS_URL", "redis://localhost:6379"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
