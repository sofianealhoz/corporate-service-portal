// Package db gère la connexion à PostgreSQL et l'application des migrations.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect ouvre un POOL de connexions et vérifie que la base répond.
//
// Pourquoi un pool et pas une connexion unique : ouvrir une connexion PostgreSQL
// coûte cher (poignée de main réseau + authentification). Le pool en garde
// plusieurs ouvertes et les prête aux requêtes, ce qui permet de servir des
// requêtes en parallèle sans payer ce coût à chaque fois.
//
// Parallèle : Django fait ça derrière ton dos (CONN_MAX_AGE), en Go on le déclare.
func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("URL de base invalide : %w", err)
	}

	// Bornes du pool : sans limite, un pic de trafic ouvrirait des centaines de
	// connexions et saturerait PostgreSQL (qui a lui-même une limite, ~100 par défaut).
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("création du pool impossible : %w", err)
	}

	// NewWithConfig n'ouvre pas forcément de connexion tout de suite :
	// on force un aller-retour pour échouer MAINTENANT si la base est injoignable,
	// plutôt qu'à la première requête d'un utilisateur.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("base injoignable : %w", err)
	}

	return pool, nil
}
