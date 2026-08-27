// Package service : domaine du catalogue. Pas de HTTP ici.
package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// erreurs métier, pour que la couche HTTP choisisse son code de statut
// sans connaître PostgreSQL
var (
	ErrNotFound  = errors.New("service introuvable")
	ErrSlugTaken = errors.New("ce slug est déjà utilisé")
)

type Service struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Tier        string    `json:"tier"`
	DurationH   int       `json:"duration_h"`
	PriceCents  int       `json:"price_cents"`
	Published   bool      `json:"published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// à garder cohérent avec le CHECK en base
var validTiers = map[string]bool{
	"standard":   true,
	"premium":    true,
	"enterprise": true,
}

const (
	defaultLimit = 20
	maxLimit     = 100
)

// Page borne les paramètres de pagination. Exporté pour que la clé de cache
// soit bâtie sur les valeurs réellement envoyées à PostgreSQL : sans cela,
// limit=0 et limit=20 donnent le même résultat sous deux entrées différentes.
func Page(limit, offset int) (int, int) {
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	if offset < 0 {
		offset = 0 // un OFFSET négatif est une erreur SQL, pas un 500
	}
	return limit, offset
}

// séparé de Service : un client ne doit pas pouvoir imposer l'id ni les dates
type CreateInput struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Tier        string `json:"tier"`
	DurationH   int    `json:"duration_h"`
	PriceCents  int    `json:"price_cents"`
	Published   bool   `json:"published"`
}

func (in CreateInput) Validate() error {
	if in.Slug == "" {
		return fmt.Errorf("le slug est obligatoire")
	}
	if in.Title == "" {
		return fmt.Errorf("le titre est obligatoire")
	}
	if !validTiers[in.Tier] {
		return fmt.Errorf("gamme invalide : attendu standard, premium ou enterprise")
	}
	if in.DurationH <= 0 {
		return fmt.Errorf("la durée doit être supérieure à 0")
	}
	if in.PriceCents < 0 {
		return fmt.Errorf("le prix ne peut pas être négatif")
	}
	return nil
}

// Cursor repère la dernière ligne rendue, pour reprendre juste après elle.
// created_at seul ne suffit pas : il n'est pas unique.
type Cursor struct {
	CreatedAt time.Time
	ID        int64
}

// Encodé en base64 pour que le client le traite comme opaque et ne se mette pas
// à le fabriquer lui-même : le format reste libre d'évoluer.
func EncodeCursor(s Service) string {
	raw := s.CreatedAt.UTC().Format(time.RFC3339Nano) + "|" + strconv.FormatInt(s.ID, 10)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(v string) (Cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(v)
	if err != nil {
		return Cursor{}, fmt.Errorf("curseur illisible")
	}

	at, id, ok := strings.Cut(string(raw), "|")
	if !ok {
		return Cursor{}, fmt.Errorf("curseur mal formé")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, at)
	if err != nil {
		return Cursor{}, fmt.Errorf("date de curseur invalide")
	}
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return Cursor{}, fmt.Errorf("identifiant de curseur invalide")
	}

	return Cursor{CreatedAt: createdAt, ID: n}, nil
}
