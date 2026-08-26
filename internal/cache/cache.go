// Package cache : cache Redis du listing.
//
// Optionnel par construction. Toute erreur est journalisée puis ignorée : un
// cache indisponible fait perdre une optimisation, il ne doit pas faire tomber
// l'API, qui répond alors depuis PostgreSQL.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// une commande Redis qui traîne ne doit pas retarder la réponse : au delà de ce
// délai on abandonne le cache et on va en base
const opTimeout = 200 * time.Millisecond

type Cache struct {
	client *redis.Client // nil = cache désactivé
}

// New ne se connecte pas : go-redis ouvre la connexion à la première commande,
// donc une URL vide ou invalide donne un cache désactivé, jamais une erreur
// fatale au démarrage.
func New(url string) *Cache {
	if url == "" {
		log.Println("cache : désactivé, aucune URL Redis")
		return &Cache{}
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		log.Printf("cache : désactivé, URL Redis invalide : %v", err)
		return &Cache{}
	}

	return &Cache{client: redis.NewClient(opts)}
}

func (c *Cache) Enabled() bool { return c.client != nil }

func (c *Cache) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}

// Get remplit dest depuis le champ d'un hash et signale si le cache a servi.
// false couvre indifféremment l'absence, la panne et le JSON illisible :
// l'appelant n'a qu'un chemin de repli à écrire.
func (c *Cache) Get(ctx context.Context, key, field string, dest any) bool {
	if c.client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	raw, err := c.client.HGet(ctx, key, field).Bytes()
	if err != nil {
		// redis.Nil = absent, c'est le cas normal d'un cache froid
		if !errors.Is(err, redis.Nil) {
			log.Printf("cache : lecture de %s/%s impossible, repli sur la base : %v", key, field, err)
		}
		return false
	}

	if err := json.Unmarshal(raw, dest); err != nil {
		log.Printf("cache : entrée %s/%s illisible, repli sur la base : %v", key, field, err)
		return false
	}
	return true
}

// Set écrit le champ et borne la durée de vie du hash entier.
func (c *Cache) Set(ctx context.Context, key, field string, v any, ttl time.Duration) {
	if c.client == nil {
		return
	}

	raw, err := json.Marshal(v)
	if err != nil {
		log.Printf("cache : sérialisation de %s/%s impossible : %v", key, field, err)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	// ExpireNX et non Expire : sans le NX, chaque nouveau champ repousserait
	// l'expiration du hash et les champs les plus anciens vivraient au delà du
	// TTL. Avec NX, le hash meurt au plus tard ttl après sa première écriture.
	_, err = c.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		p.HSet(ctx, key, field, raw)
		p.ExpireNX(ctx, key, ttl)
		return nil
	})
	if err != nil {
		log.Printf("cache : écriture de %s/%s impossible : %v", key, field, err)
	}
}

// Delete vide une entrée. Un hash unique par famille de clés rend
// l'invalidation atomique et en une commande, sans SCAN.
func (c *Cache) Delete(ctx context.Context, key string) {
	if c.client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	if err := c.client.Del(ctx, key).Err(); err != nil {
		// pire cas : le listing reste périmé jusqu'à expiration du TTL
		log.Printf("cache : invalidation de %s impossible : %v", key, err)
	}
}
