// Package testdb fournit une base PostgreSQL prête à l'emploi pour les tests
// d'intégration : migrations appliquées, table services vidée.
package testdb

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sofianealhoz/corporate-service-portal/internal/db"
)

// base de développement du docker-compose
const devURL = "postgres://portal:portal@localhost:5432/portal?sslmode=disable"

// base dérivée pour les tests, pour ne pas toucher aux données de développement
const testDatabase = "portal_test"

// Connect ouvre un pool sur la base de test. Si PostgreSQL est injoignable, le
// test est ignoré plutôt qu'en échec : `go test ./...` doit rester vert sur un
// clone sans Docker.
func Connect(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	// une URL explicite est utilisée telle quelle, sans créer de base
	url, own := os.Getenv("TEST_DATABASE_URL"), false
	if url == "" {
		url, own = devURL, true
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("URL de base invalide : %v", err)
	}

	pool, err := open(ctx, cfg)
	if err != nil {
		t.Skipf("PostgreSQL injoignable (%v). Lancer `docker compose up -d`, "+
			"ou définir TEST_DATABASE_URL, pour exécuter les tests d'intégration.", err)
	}

	if own {
		createDatabase(ctx, t, pool)
		pool.Close()

		cfg.ConnConfig.Database = testDatabase
		if pool, err = open(ctx, cfg); err != nil {
			t.Fatalf("connexion à %s : %v", testDatabase, err)
		}
	}
	t.Cleanup(pool.Close)

	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrations : %v", err)
	}
	// isolation entre tests : chacun repart d'une table vide
	if _, err := pool.Exec(ctx, `TRUNCATE services RESTART IDENTITY`); err != nil {
		t.Fatalf("nettoyage de la table services : %v", err)
	}

	return pool
}

func createDatabase(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	// CREATE DATABASE n'accepte pas IF NOT EXISTS ; 42P04 = base déjà présente
	_, err := pool.Exec(ctx, `CREATE DATABASE `+testDatabase)
	var pgErr *pgconn.PgError
	if err != nil && !(errors.As(err, &pgErr) && pgErr.Code == "42P04") {
		t.Fatalf("création de %s : %v", testDatabase, err)
	}
}

func open(ctx context.Context, cfg *pgxpool.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(ctx, cfg.Copy())
	if err != nil {
		return nil, err
	}

	// Ping : NewWithConfig n'ouvre rien, sans lui l'absence de base ne serait
	// détectée qu'à la première requête
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
