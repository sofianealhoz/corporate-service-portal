package api

import "net/http"

// healthResponse est le corps JSON renvoyé par /health.
type healthResponse struct {
	Status string `json:"status"`
}

// handleHealth répond 200 avec {"status":"ok"}.
// Sert aux health checks : c'est ce que Docker, Kubernetes ou un load balancer
// appellent pour savoir s'il faut envoyer du trafic à cette instance.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
