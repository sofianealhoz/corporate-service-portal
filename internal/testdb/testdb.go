// Package testdb fournit une base PostgreSQL prête à l'emploi pour les tests
// d'intégration : migrations appliquées, table services vidée.
package testdb

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sofianealhoz/corporate-service-portal/internal/db"
)

// base de développement du docker-compose
const devURL = "postgres://portal:portal@localhost:5432/portal?sslmode=disable"

const suffix = "_test"

// Connect ouvre un pool sur la base de test et la rend vide.
//
// La base visée est toujours celle de l'URL suffixée par _test, jamais l'URL
// elle-même : les tests vident la table services, ils ne doivent pas pouvoir
// effacer une base de développement par une variable d'environnement mal
// pointée.
//
// Si PostgreSQL est injoignable, le test est ignoré plutôt qu'en échec :
// `go test ./...` doit rester vert sur un clone sans Docker.
func Connect(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = devURL
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("URL de base invalide : %v", err)
	}

	admin, err := open(ctx, cfg)
	if err != nil {
		SkipOrFail(t, "PostgreSQL injoignable (%v). Lancer `docker compose up -d`, "+
			"ou définir TEST_DATABASE_URL, pour exécuter les tests d'intégration.", err)
	}

	target := cfg.ConnConfig.Database
	if !strings.HasSuffix(target, suffix) {
		target += suffix
		createDatabase(ctx, t, admin, target)
	}
	admin.Close()

	cfg.ConnConfig.Database = target
	pool, err := open(ctx, cfg)
	if err != nil {
		t.Fatalf("connexion à %s : %v", target, err)
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

// SkipOrFail ignore le test, sauf si REQUIRE_INTEGRATION vaut 1. En CI les
// services sont garantis présents : un test d'intégration ignoré y serait un
// échec silencieux, exactement ce qu'une CI est censée empêcher.
func SkipOrFail(t *testing.T, format string, args ...any) {
	t.Helper()
	if os.Getenv("REQUIRE_INTEGRATION") == "1" {
		t.Fatalf(format, args...)
	}
	t.Skipf(format, args...)
}

func createDatabase(ctx context.Context, t *testing.T, admin *pgxpool.Pool, name string) {
	t.Helper()

	// CREATE DATABASE n'accepte pas IF NOT EXISTS ; 42P04 = base déjà présente
	_, err := admin.Exec(ctx, `CREATE DATABASE `+name)
	var pgErr *pgconn.PgError
	if err != nil && !(errors.As(err, &pgErr) && pgErr.Code == "42P04") {
		t.Fatalf("création de %s : %v", name, err)
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
