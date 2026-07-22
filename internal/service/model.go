// Package service contient le domaine « service » : la structure de données,
// ses règles de validation, et l'accès à la base (repository).
//
// Ce paquet ne connaît rien de HTTP : pas de handler, pas de JSON de requête.
// C'est volontaire : on pourrait s'en servir depuis une commande en ligne ou un
// worker sans rien changer.
package service

import (
	"errors"
	"fmt"
	"time"
)

// Erreurs métier du domaine.
//
// Pourquoi des erreurs déclarées ici plutôt que de renvoyer directement une
// erreur SQL : la couche HTTP doit pouvoir décider du code de statut
// (404 ? 409 ?) sans connaître PostgreSQL. Elle compare avec errors.Is().
var (
	ErrNotFound  = errors.New("service introuvable")
	ErrSlugTaken = errors.New("ce slug est déjà utilisé")
)

// Service = une offre du catalogue.
//
// Les tags `json:"..."` définissent le nom des champs une fois sérialisés.
// Convention retenue : snake_case dans le JSON (duration_h), CamelCase en Go
// (DurationH). La majuscule initiale est obligatoire pour qu'un champ soit
// exporté, donc visible depuis un autre paquet et sérialisable.
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

// gammes autorisées, à garder cohérentes avec la contrainte CHECK en base.
var validTiers = map[string]bool{
	"standard":   true,
	"premium":    true,
	"enterprise": true,
}

// CreateInput = les champs acceptés à la création.
//
// Structure séparée de Service : on ne veut pas qu'un client
// puisse imposer l'id ou les dates de création en les envoyant dans le JSON.
// En n'exposant que les champs modifiables, la faille est impossible par
// construction.
type CreateInput struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Tier        string `json:"tier"`
	DurationH   int    `json:"duration_h"`
	PriceCents  int    `json:"price_cents"`
	Published   bool   `json:"published"`
}

// Validate vérifie les règles métier avant tout accès à la base.
//
// La base a déjà des contraintes (CHECK, NOT NULL) : c'est le dernier rempart.
// Mais valider ici permet de renvoyer un message clair à l'utilisateur
// plutôt qu'une erreur PostgreSQL brute.
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
