package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sofianealhoz/corporate-service-portal/internal/api"
	"github.com/sofianealhoz/corporate-service-portal/internal/cache"
	"github.com/sofianealhoz/corporate-service-portal/internal/service"
	"github.com/sofianealhoz/corporate-service-portal/internal/testdb"
)

// miroir de la réponse du listing, qui n'est pas exportée par le paquet api
type listBody struct {
	Items      []service.Service `json:"items"`
	Count      int               `json:"count"`
	NextCursor string            `json:"next_cursor"`
}

// Monte le routeur complet sur une vraie base. Cache désactivé : ces tests
// portent sur le contrat HTTP, pas sur Redis.
func newServer(t *testing.T) (*httptest.Server, *service.Repository) {
	t.Helper()

	repo := service.NewRepository(testdb.Connect(t))
	srv := httptest.NewServer(api.New(repo, cache.New("")).Router())
	t.Cleanup(srv.Close)
	return srv, repo
}

// insère n services publiés et rend leurs slugs
func seed(t *testing.T, repo *service.Repository, n int) map[string]bool {
	t.Helper()

	slugs := map[string]bool{}
	for i := range n {
		slug := fmt.Sprintf("service-%d", i)
		if _, err := repo.Create(context.Background(), service.CreateInput{
			Slug:      slug,
			Title:     fmt.Sprintf("Service %d", i),
			Tier:      "standard",
			DurationH: 4,
			Published: true,
		}); err != nil {
			t.Fatalf("insertion de %s : %v", slug, err)
		}
		slugs[slug] = true
	}
	return slugs
}

func TestListingPagine(t *testing.T) {
	srv, repo := newServer(t)

	const total = 5
	attendus := seed(t, repo, total)

	// on parcourt tout le catalogue deux par deux et on vérifie que la
	// pagination rend chaque service exactement une fois
	const pageSize = 2
	vus := map[string]bool{}
	for offset := 0; offset < total; offset += pageSize {
		body := get[listBody](t, srv.URL+fmt.Sprintf("/api/services?limit=%d&offset=%d", pageSize, offset))

		voulu := min(pageSize, total-offset)
		if body.Count != voulu || len(body.Items) != voulu {
			t.Fatalf("offset %d : %d services attendus, obtenu %d", offset, voulu, body.Count)
		}

		for _, s := range body.Items {
			if vus[s.Slug] {
				t.Fatalf("%s apparaît sur deux pages", s.Slug)
			}
			vus[s.Slug] = true
		}
	}

	if len(vus) != len(attendus) {
		t.Fatalf("%d services insérés, %d rendus par la pagination", len(attendus), len(vus))
	}
	for slug := range attendus {
		if !vus[slug] {
			t.Fatalf("%s n'est apparu sur aucune page", slug)
		}
	}
}

func TestCreationPuisRelecture(t *testing.T) {
	srv, _ := newServer(t)

	in := service.CreateInput{
		Slug:        "migration-cloud",
		Title:       "Migration cloud",
		Description: "Reprise d'un parc existant.",
		Tier:        "enterprise",
		DurationH:   40,
		PriceCents:  1250000,
		Published:   true,
	}

	payload, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("sérialisation : %v", err)
	}

	res, err := http.Post(srv.URL+"/api/services", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("création : %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("201 attendu à la création, obtenu %d", res.StatusCode)
	}

	var cree service.Service
	if err := json.NewDecoder(res.Body).Decode(&cree); err != nil {
		t.Fatalf("réponse de création illisible : %v", err)
	}
	if cree.ID == 0 || cree.CreatedAt.IsZero() {
		t.Fatalf("la création doit renvoyer l'id et les dates, obtenu %+v", cree)
	}

	relu := get[service.Service](t, srv.URL+"/api/services/"+in.Slug)
	if relu != cree {
		t.Fatalf("la relecture diffère de la création :\n création %+v\n relecture %+v", cree, relu)
	}
}

func get[T any](t *testing.T, url string) T {
	t.Helper()

	res, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s : %v", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s : 200 attendu, obtenu %d", url, res.StatusCode)
	}

	var body T
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("GET %s : réponse illisible : %v", url, err)
	}
	return body
}

// Même garantie que la pagination par offset, mais en reprenant après la
// dernière ligne rendue plutôt qu'en sautant des lignes.
func TestListingParCurseur(t *testing.T) {
	srv, repo := newServer(t)

	const total = 5
	const pageSize = 2
	attendus := seed(t, repo, total)

	vus := map[string]bool{}
	cursor := ""
	for pages := 0; ; pages++ {
		if pages > total {
			t.Fatal("le parcours par curseur ne s'arrête pas")
		}

		url := fmt.Sprintf("%s/api/services?limit=%d", srv.URL, pageSize)
		if cursor != "" {
			url += "&after=" + cursor
		}
		body := get[listBody](t, url)

		for _, s := range body.Items {
			if vus[s.Slug] {
				t.Fatalf("%s apparaît sur deux pages", s.Slug)
			}
			vus[s.Slug] = true
		}

		// pas de curseur suivant : c'était la dernière page
		if body.NextCursor == "" {
			break
		}
		cursor = body.NextCursor
	}

	if len(vus) != len(attendus) {
		t.Fatalf("%d services insérés, %d rendus par le curseur", len(attendus), len(vus))
	}
	for slug := range attendus {
		if !vus[slug] {
			t.Fatalf("%s n'est apparu sur aucune page", slug)
		}
	}
}

func TestCurseurInvalide(t *testing.T) {
	srv, _ := newServer(t)

	res, err := http.Get(srv.URL + "/api/services?after=pas-un-curseur")
	if err != nil {
		t.Fatalf("requête : %v", err)
	}
	defer res.Body.Close()

	// faute du client, pas panne du serveur
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("400 attendu, obtenu %d", res.StatusCode)
	}
}
