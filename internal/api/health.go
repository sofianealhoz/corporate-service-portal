package api

import "net/http"

type healthResponse struct {
	Status string `json:"status"`
}

// appelé par Docker ou un load balancer pour savoir si l'instance répond
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
