package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/sofianealhoz/corporate-service-portal/internal/service"
)

// objet plutôt qu'un tableau nu, pour pouvoir ajouter des champs plus tard
type listResponse struct {
	Items []service.Service `json:"items"`
	Count int               `json:"count"`
}

const (
	listCacheKey = "services:list"
	listCacheTTL = 30 * time.Second
)

// La clé doit contenir tout ce qui change le résultat. En oublier un revient à
// servir la page d'une autre requête, ce qui est pire que pas de cache.
func listCacheField(onlyPublished bool, limit, offset int) string {
	return fmt.Sprintf("published=%t&limit=%d&offset=%d", onlyPublished, limit, offset)
}

func (a *API) handleListServices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// par défaut on ne montre que le publié : voir les brouillons doit être explicite
	onlyPublished := q.Get("published") != "false"

	limit, offset := service.Page(atoiOr(q.Get("limit"), 20), atoiOr(q.Get("offset"), 0))
	field := listCacheField(onlyPublished, limit, offset)

	var resp listResponse
	if a.cache.Get(r.Context(), listCacheKey, field, &resp) {
		writeJSON(w, http.StatusOK, resp)
		return
	}

	// si le client coupe, le contexte est annulé et pgx interrompt la requête
	items, err := a.services.List(r.Context(), onlyPublished, limit, offset)
	if err != nil {
		log.Printf("liste des services : %v", err)
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}

	resp = listResponse{Items: items, Count: len(items)}
	a.cache.Set(r.Context(), listCacheKey, field, resp, listCacheTTL)

	writeJSON(w, http.StatusOK, resp)
}

func (a *API) handleGetService(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// pas encore d'authentification : personne ne voit les brouillons. Un
	// contrôle administrateur se brancherait ici et passerait true.
	const includeUnpublished = false

	s, err := a.services.GetBySlug(r.Context(), slug, includeUnpublished)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			// 404 et non 403 : un refus confirmerait que le slug existe
			writeError(w, http.StatusNotFound, "service introuvable")
			return
		}
		log.Printf("récupération du service %q : %v", slug, err)
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}

	writeJSON(w, http.StatusOK, s)
}

func (a *API) handleCreateService(w http.ResponseWriter, r *http.Request) {
	var in service.CreateInput

	// un champ inconnu lève une erreur au lieu d'être ignoré en silence
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "JSON invalide : "+err.Error())
		return
	}

	s, err := a.services.Create(r.Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlugTaken):
			writeError(w, http.StatusConflict, "ce slug est déjà utilisé")
		default:
			// erreur de validation : faute du client, donc 400
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	// le listing vient de changer : on le vide plutôt que de deviner quelles
	// pages sont touchées
	a.cache.Delete(r.Context(), listCacheKey)

	writeJSON(w, http.StatusCreated, s)
}

func atoiOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}
