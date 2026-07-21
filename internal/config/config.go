// Package config centralise la lecture des réglages de l'application.
// Principe : le code ne contient JAMAIS de valeur d'environnement en dur
// (port, mot de passe, URL de base). Tout vient de variables d'environnement,
// ce qui permet de faire tourner le même binaire en local et en production.
package config

import "os"

// Config regroupe tous les réglages lus au démarrage.
type Config struct {
	Port        string // port d'écoute HTTP
	DatabaseURL string // connexion PostgreSQL (servira au M2)
	RedisURL    string // connexion Redis (servira au M4)
}

// Load lit la configuration depuis l'environnement.
// Chaque réglage a une valeur par défaut, donc l'app démarre sans rien configurer.
func Load() Config {
	return Config{
		Port: getenv("PORT", "8080"),
		// Par défaut : les identifiants du docker-compose de développement.
		// En production, DATABASE_URL est fournie par l'environnement.
		DatabaseURL: getenv("DATABASE_URL",
			"postgres://portal:portal@localhost:5432/portal?sslmode=disable"),
		RedisURL: getenv("REDIS_URL", "redis://localhost:6379"),
	}
}

// getenv renvoie la variable d'environnement `key`, ou `fallback` si elle est vide.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
