package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// writeJSON sérialise `v` en JSON et l'écrit dans la réponse.
//
// ⚠️ Ordre imposé par net/http : headers -> code de statut -> corps.
// Inverser fait perdre silencieusement le code de statut (Go envoie 200 d'office).
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Ici les headers sont déjà partis : impossible de changer la réponse.
		// On ne peut que tracer.
		log.Printf("écriture JSON impossible : %v", err)
	}
}

// errorResponse = format d'erreur unique pour toute l'API.
// Un client qui sait lire une erreur sait les lire toutes.
type errorResponse struct {
	Error string `json:"error"`
}

// writeError renvoie une erreur au format JSON.
//
// Règle : le message doit être utile au client SANS révéler l'intérieur du
// système (pas de message d'erreur PostgreSQL brut, qui divulguerait la
// structure de la base à un attaquant).
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
