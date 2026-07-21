package db

import (
	"context"
	"embed"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// go:embed embarque les fichiers .sql DANS le binaire compilé.
// Conséquence : on déploie UN SEUL fichier exécutable, sans avoir à copier
// le dossier migrations/ à côté. Les fichiers doivent être sous ce paquet,
// d'où internal/db/migrations/ plutôt qu'un dossier à la racine.
//
//go:embed migrations/*.sql
var migrationFS embed.FS

// Migrate applique, dans l'ordre, les migrations pas encore appliquées.
//
// Principe (le même que Django ou Knex) : une table de suivi mémorise les
// versions déjà passées, donc relancer l'application ne rejoue rien.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	// 1. La table qui mémorise ce qui a déjà été appliqué.
	//    (≈ la table django_migrations de Django)
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("création de schema_migrations : %w", err)
	}

	// 2. Lister les fichiers embarqués et les TRIER : l'ordre d'application
	//    doit être déterministe (0001 avant 0002), jamais l'ordre du système de fichiers.
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
	sort.Strings(names)

	// 3. Appliquer celles qui manquent.
	for _, name := range names {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, name,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("vérification de %s : %w", name, err)
		}
		if exists {
			continue // déjà appliquée
		}

		content, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("lecture de %s : %w", name, err)
		}

		// TRANSACTION : le SQL de la migration ET l'enregistrement de sa version
		// réussissent ensemble, ou échouent ensemble. Sans ça, une panne au mauvais
		// moment laisserait la base à moitié migrée sans qu'on le sache.
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
