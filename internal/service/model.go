// Package service : domaine du catalogue. Pas de HTTP ici.
package service

import (
	"errors"
	"fmt"
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
