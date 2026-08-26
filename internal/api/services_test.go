package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sofianealhoz/corporate-service-portal/internal/api"
	"github.com/sofianealhoz/corporate-service-portal/internal/cache"
	"github.com/sofianealhoz/corporate-service-portal/internal/service"
	"github.com/sofianealhoz/corporate-service-portal/internal/testdb"
)

// Le détail doit masquer les brouillons comme le fait le listing, et répondre
// 404 : un 403 confirmerait au passage que le slug existe.
func TestGetServicePubliePasLesBrouillons(t *testing.T) {
	pool := testdb.Connect(t)
	repo := service.NewRepository(pool)

	seed := func(slug string, published bool) {
		t.Helper()
		_, err := repo.Create(context.Background(), service.CreateInput{
			Slug:      slug,
			Title:     "Audit de sécurité",
			Tier:      "standard",
			DurationH: 14,
			Published: published,
		})
		if err != nil {
			t.Fatalf("insertion de %s : %v", slug, err)
		}
	}
	seed("audit-publie", true)
	seed("audit-brouillon", false)

	// cache désactivé : ce test porte sur l'autorisation, pas sur Redis
	srv := httptest.NewServer(api.New(repo, cache.New("")).Router())
	defer srv.Close()

	cases := []struct {
		slug string
		want int
	}{
		{"audit-publie", http.StatusOK},
		{"audit-brouillon", http.StatusNotFound},
		{"audit-inexistant", http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.slug, func(t *testing.T) {
			res, err := http.Get(srv.URL + "/api/services/" + tc.slug)
			if err != nil {
				t.Fatalf("requête : %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != tc.want {
				t.Fatalf("statut %d attendu, obtenu %d", tc.want, res.StatusCode)
			}
		})
	}
}
