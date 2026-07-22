package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository est le seul endroit du projet qui écrit du SQL sur les services.
// Tout le reste du code passe par ses méthodes.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository construit le repository à partir du pool de connexions.
// La dépendance est injectée plutôt que récupérée d'une variable globale :
// on peut ainsi en créer un autre, branché sur une base de test.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// colonnes listées explicitement (jamais SELECT *) : si quelqu'un ajoute une
// colonne demain, le code continue de fonctionner au lieu de casser au Scan.
const columns = `id, slug, title, description, tier, duration_h, price_cents, published, created_at, updated_at`

// scanRow lit une ligne SQL vers une struct Service.
// L'ordre des champs doit correspondre exactement à celui de `columns`.
func scanRow(row pgx.Row) (Service, error) {
	var s Service
	err := row.Scan(
		&s.ID, &s.Slug, &s.Title, &s.Description, &s.Tier,
		&s.DurationH, &s.PriceCents, &s.Published, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

// List renvoie les services, les plus récents d'abord.
//
// onlyPublished=true correspond au site public ; false à l'espace d'administration.
// limit/offset = pagination : sans limite, une table qui grossit finirait par
// renvoyer des dizaines de milliers de lignes en une seule réponse.
func (r *Repository) List(ctx context.Context, onlyPublished bool, limit, offset int) ([]Service, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // garde-fou : un client ne peut pas réclamer 1 million de lignes
	}

	// Requête paramétrée : les valeurs sont envoyées séparément du
	// texte SQL, donc elles ne peuvent pas être interprétées comme du code :
	// c'est ce qui rend l'injection SQL impossible. Ne jamais construire une
	// requête par concaténation de chaînes.
	query := `SELECT ` + columns + `
	          FROM services
	          WHERE (NOT $1::bool OR published = TRUE)
	          ORDER BY created_at DESC
	          LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, onlyPublished, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("liste des services : %w", err)
	}
	defer rows.Close() // libère la connexion vers le pool, même en cas d'erreur

	// Slice initialisée non-nil : sérialisée en `[]` et non `null` quand elle est vide.
	services := make([]Service, 0)
	for rows.Next() {
		s, err := scanRow(rows)
		if err != nil {
			return nil, fmt.Errorf("lecture d'un service : %w", err)
		}
		services = append(services, s)
	}
	// Une erreur peut survenir pendant l'itération, par exemple une connexion coupée.
	// L'oublier ferait passer une liste tronquée pour un résultat complet.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours des services : %w", err)
	}

	return services, nil
}

// GetBySlug renvoie un service par son slug, ou ErrNotFound.
func (r *Repository) GetBySlug(ctx context.Context, slug string) (Service, error) {
	query := `SELECT ` + columns + ` FROM services WHERE slug = $1`

	s, err := scanRow(r.pool.QueryRow(ctx, query, slug))
	if err != nil {
		// pgx.ErrNoRows = « aucune ligne » : ce n'est pas une panne, c'est un 404.
		// On le traduit en erreur métier pour que la couche HTTP décide du statut.
		if errors.Is(err, pgx.ErrNoRows) {
			return Service{}, ErrNotFound
		}
		return Service{}, fmt.Errorf("récupération du service : %w", err)
	}
	return s, nil
}

// Create insère un service et renvoie la ligne créée.
func (r *Repository) Create(ctx context.Context, in CreateInput) (Service, error) {
	if err := in.Validate(); err != nil {
		return Service{}, err
	}

	// RETURNING : PostgreSQL renvoie la ligne insérée dans la même requête.
	// On récupère ainsi l'id et les dates générés par la base sans second aller-retour.
	query := `INSERT INTO services (slug, title, description, tier, duration_h, price_cents, published)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)
	          RETURNING ` + columns

	s, err := scanRow(r.pool.QueryRow(ctx, query,
		in.Slug, in.Title, in.Description, in.Tier,
		in.DurationH, in.PriceCents, in.Published,
	))
	if err != nil {
		// 23505 = violation de contrainte d'unicité (ici : slug déjà pris).
		// On laisse la base arbitrer plutôt que de faire un SELECT avant l'INSERT :
		// entre les deux, une autre requête pourrait insérer le même slug
		// (situation de compétition), alors que la contrainte est fiable.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Service{}, ErrSlugTaken
		}
		return Service{}, fmt.Errorf("création du service : %w", err)
	}
	return s, nil
}
