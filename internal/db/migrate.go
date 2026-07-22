package db

import (
	"context"
	"embed"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// les .sql sont embarqués dans le binaire : un seul fichier à déployer
//
//go:embed migrations/*.sql
var migrationFS embed.FS

// applique les migrations pas encore passées, dans l'ordre
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("création de schema_migrations : %w", err)
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("lecture du dossier migrations : %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names) // ordre déterministe, pas celui du système de fichiers

	for _, name := range names {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, name,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("vérification de %s : %w", name, err)
		}
		if exists {
			continue
		}

		content, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("lecture de %s : %w", name, err)
		}

		// transaction : le SQL et l'enregistrement de la version passent ensemble
		// ou pas du tout, sinon la base reste à moitié migrée
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("ouverture de transaction pour %s : %w", name, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("exécution de %s : %w", name, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("enregistrement de %s : %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit de %s : %w", name, err)
		}

		fmt.Printf("migration appliquée : %s\n", name)
	}

	return nil
}
