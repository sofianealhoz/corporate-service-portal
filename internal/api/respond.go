package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// ordre imposé par net/http : headers, statut, puis corps
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("écriture JSON impossible : %v", err)
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

// message utile au client, sans détail interne (pas d'erreur SQL brute)
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
