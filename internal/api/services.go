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

// listResponse enveloppe la liste renvoyée.
//
// Un objet plutôt qu'un tableau nu : on peut ajouter des champs (total, page
// suivante) plus tard sans casser les clients existants.
type listResponse struct {
	Items []service.Service `json:"items"`
	Count int               `json:"count"`
}

// GET /api/services?published=true&limit=20&offset=0
func (a *API) handleListServices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Par défaut on ne montre QUE les services publiés : le comportement
	// sûr est celui par défaut. Voir les brouillons devra être un choix
	// explicite (et, plus tard, réservé aux administrateurs authentifiés).
	onlyPublished := q.Get("published") != "false"

	limit := atoiOr(q.Get("limit"), 20)
	offset := atoiOr(q.Get("offset"), 0)

	// r.Context() : si le client coupe la connexion, le contexte est annulé et
	// pgx interrompt la requête SQL. On ne travaille pas pour un client parti.
	items, err := a.services.List(r.Context(), onlyPublished, limit, offset)
	if err != nil {
		// On journalise le détail côté serveur, mais on ne renvoie qu'un message
		// générique : une erreur SQL brute révélerait la structure de la base.
		log.Printf("liste des services : %v", err)
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}

	writeJSON(w, http.StatusOK, listResponse{Items: items, Count: len(items)})
}

// GET /api/services/{slug}
func (a *API) handleGetService(w http.ResponseWriter, r *http.Request) {
	// chi.URLParam lit le segment déclaré {slug} dans la route.
	slug := chi.URLParam(r, "slug")

	s, err := a.services.GetBySlug(r.Context(), slug)
	if err != nil {
		// errors.Is compare à l'erreur métier : la couche HTTP choisit le statut
		// sans jamais avoir à connaître PostgreSQL.
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

// POST /api/services
func (a *API) handleCreateService(w http.ResponseWriter, r *http.Request) {
	var in service.CreateInput

	// Un champ inconnu dans le JSON provoque une erreur
	// au lieu d'être ignoré en silence. Ça transforme une faute de frappe du
	// client ("titel" au lieu de "title") en message clair plutôt qu'en
	// création d'un service au titre vide.
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
			// 409 Conflict : la requête est valide, mais elle entre en conflit
			// avec l'état actuel de la ressource.
			writeError(w, http.StatusConflict, "ce slug est déjà utilisé")
		default:
			// Erreur de validation -> faute du CLIENT, donc 4xx et pas 500.
			// (Create renvoie l'erreur de Validate() telle quelle.)
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	// 201 Created = une ressource a été créée (et pas un simple 200).
	writeJSON(w, http.StatusCreated, s)
}

// atoiOr convertit une chaîne en entier, avec une valeur de repli.
// Évite de répéter la gestion d'erreur pour chaque paramètre de requête.
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
