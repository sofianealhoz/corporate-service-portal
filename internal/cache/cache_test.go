package cache_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sofianealhoz/corporate-service-portal/internal/cache"
	"github.com/sofianealhoz/corporate-service-portal/internal/testdb"
)

type payload struct {
	Items []string `json:"items"`
}

// Le chemin de dégradation ne dépend d'aucun service : il doit être vérifié
// partout, y compris sur un clone sans Docker.
func TestCacheIndisponibleNeCassePas(t *testing.T) {
	cas := map[string]string{
		"désactivé":   "",
		"URL cassée":  "pas-une-url",
		"injoignable": "redis://127.0.0.1:1",
	}

	for nom, url := range cas {
		t.Run(nom, func(t *testing.T) {
			c := cache.New(url)
			defer c.Close()

			var dest payload
			// une lecture qui échoue doit dire « je n'ai rien », pas paniquer
			if c.Get(context.Background(), "k", "f", &dest) {
				t.Fatal("un cache indisponible ne doit jamais annoncer un succès")
			}

			// écriture et invalidation doivent être des non-opérations
			c.Set(context.Background(), "k", "f", payload{Items: []string{"a"}}, time.Second)
			c.Delete(context.Background(), "k")
		})
	}
}

func TestCacheAllerRetour(t *testing.T) {
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		url = "redis://localhost:6379"
	}

	c := cache.New(url)
	defer c.Close()

	const key, field = "test:cache", "published=true&limit=20&offset=0"
	c.Delete(context.Background(), key)

	écrit := payload{Items: []string{"audit", "migration"}}
	c.Set(context.Background(), key, field, écrit, 30*time.Second)

	var relu payload
	if !c.Get(context.Background(), key, field, &relu) {
		testdb.SkipOrFail(t, "Redis injoignable sur %s. Lancer `docker compose up -d` "+
			"pour exécuter ce test.", url)
	}
	if len(relu.Items) != len(écrit.Items) || relu.Items[0] != écrit.Items[0] {
		t.Fatalf("relu %+v, écrit %+v", relu, écrit)
	}

	// un champ jamais écrit est un défaut de cache, pas une erreur
	if c.Get(context.Background(), key, "published=false&limit=20&offset=0", &relu) {
		t.Fatal("un champ absent ne doit pas être servi")
	}

	c.Delete(context.Background(), key)
	if c.Get(context.Background(), key, field, &relu) {
		t.Fatal("l'invalidation doit vider l'entrée")
	}
}
