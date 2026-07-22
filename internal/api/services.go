package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/sofianealhoz/corporate-service-portal/internal/service"
)

// objet plutôt qu'un tableau nu, pour pouvoir ajouter des champs plus tard
type listResponse struct {
	Items []service.Service `json:"items"`
	Count int               `json:"count"`
}

func (a *API) handleListServices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// par défaut on ne montre que le publié : voir les brouillons doit être explicite
	onlyPublished := q.Get("published") != "false"

	limit := atoiOr(q.Get("limit"), 20)
	offset := atoiOr(q.Get("offset"), 0)

	// si le client coupe, le contexte est annulé et pgx interrompt la requête
	items, err := a.services.List(r.Context(), onlyPublished, limit, offset)
	if err != nil {
		log.Printf("liste des services : %v", err)
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}

	writeJSON(w, http.StatusOK, listResponse{Items: items, Count: len(items)})
}

func (a *API) handleGetService(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	s, err := a.services.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
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
