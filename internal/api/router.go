package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sofianealhoz/corporate-service-portal/internal/service"
)

// dépendances des handlers, passées explicitement plutôt que par des globales :
// on peut en construire une autre branchée sur une base de test
type API struct {
	services *service.Repository
}

func New(services *service.Repository) *API {
	return &API{services: services}
}

func (a *API) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID) // id unique par requête, pour les logs
	r.Use(middleware.RealIP)    // vraie IP même derrière un proxy
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) // un panic devient un 500 au lieu de tuer le serveur

	r.Get("/health", handleHealth)

	// préfixe /api pour pouvoir servir le front sur / plus tard
	r.Route("/api", func(r chi.Router) {
		r.Route("/services", func(r chi.Router) {
			r.Get("/", a.handleListServices)
			r.Post("/", a.handleCreateService)
			r.Get("/{slug}", a.handleGetService)
		})
	})

	return r
}
