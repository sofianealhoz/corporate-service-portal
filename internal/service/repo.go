package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// seul endroit du projet qui écrit du SQL sur les services
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// colonnes explicites plutôt que SELECT *, pour qu'un ajout de colonne
// ne casse pas le Scan
const columns = `id, slug, title, description, tier, duration_h, price_cents, published, created_at, updated_at`

// l'ordre doit correspondre à celui de columns
func scanRow(row pgx.Row) (Service, error) {
	var s Service
	err := row.Scan(
		&s.ID, &s.Slug, &s.Title, &s.Description, &s.Tier,
		&s.DurationH, &s.PriceCents, &s.Published, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

// ListQuery rassemble les paramètres du listing. Une structure plutôt que six
// arguments positionnels, qu'on finit par intervertir.
type ListQuery struct {
	OnlyPublished bool // false pour l'administration
	Limit         int
	Offset        int
	After         *Cursor // non nil : pagination par curseur, Offset est ignoré
}

// l'ordre est total : created_at seul laisserait les ex aequo flotter d'une
// page à l'autre. Il colle à idx_services_pagination, donc rien à trier.
const listOrder = ` ORDER BY created_at DESC, id DESC`

// requêtes paramétrées : valeurs envoyées séparément du texte SQL,
// jamais de concaténation
const listByOffset = `SELECT ` + columns + `
                      FROM services
                      WHERE (NOT $1::bool OR published = TRUE)` + listOrder + `
                      LIMIT $2 OFFSET $3`

// la comparaison de n-uplets se lit comme le tri et sait utiliser l'index,
// contrairement à un created_at < x OR (created_at = x AND id < y)
const listByCursor = `SELECT ` + columns + `
                      FROM services
                      WHERE (NOT $1::bool OR published = TRUE)
                        AND (created_at, id) < ($2, $3)` + listOrder + `
                      LIMIT $4`

func (r *Repository) List(ctx context.Context, q ListQuery) ([]Service, error) {
	q.Limit, q.Offset = Page(q.Limit, q.Offset) // garde-fou, même si l'appelant a déjà borné

	var (
		rows pgx.Rows
		err  error
	)
	if q.After != nil {
		rows, err = r.pool.Query(ctx, listByCursor, q.OnlyPublished, q.After.CreatedAt, q.After.ID, q.Limit)
	} else {
		rows, err = r.pool.Query(ctx, listByOffset, q.OnlyPublished, q.Limit, q.Offset)
	}
	if err != nil {
		return nil, fmt.Errorf("liste des services : %w", err)
	}
	defer rows.Close()

	services := make([]Service, 0) // non-nil : sérialisé en [] et non null
	for rows.Next() {
		s, err := scanRow(rows)
		if err != nil {
			return nil, fmt.Errorf("lecture d'un service : %w", err)
		}
		services = append(services, s)
	}
	// sinon une liste tronquée passerait pour un résultat complet
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours des services : %w", err)
	}

	return services, nil
}

func (r *Repository) GetBySlug(ctx context.Context, slug string, includeUnpublished bool) (Service, error) {
	query := `SELECT ` + columns + `
	          FROM services
	          WHERE slug = $1 AND ($2::bool OR published = TRUE)`

	s, err := scanRow(r.pool.QueryRow(ctx, query, slug, includeUnpublished))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Service{}, ErrNotFound // pas une panne, un 404
		}
		return Service{}, fmt.Errorf("récupération du service : %w", err)
	}
	return s, nil
}

func (r *Repository) Create(ctx context.Context, in CreateInput) (Service, error) {
	if err := in.Validate(); err != nil {
		return Service{}, err
	}

	// RETURNING pour récupérer l'id et les dates sans second aller-retour
	query := `INSERT INTO services (slug, title, description, tier, duration_h, price_cents, published)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)
	          RETURNING ` + columns

	s, err := scanRow(r.pool.QueryRow(ctx, query,
		in.Slug, in.Title, in.Description, in.Tier,
		in.DurationH, in.PriceCents, in.Published,
	))
	if err != nil {
		// 23505 = violation d'unicité. On laisse la base arbitrer plutôt que de
		// faire un SELECT avant l'INSERT, qui laisserait passer un doublon.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Service{}, ErrSlugTaken
		}
		return Service{}, fmt.Errorf("création du service : %w", err)
	}
	return s, nil
}
