package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sofianealhoz/corporate-service-portal/internal/service"
)

// API porte les dépendances des handlers (aujourd'hui le repository des
// services ; demain le cache Redis, le service d'authentification...).
//
// Pourquoi une structure plutôt que des variables globales : chaque handler
// est une MÉTHODE sur *API, donc il accède à ses dépendances via le receveur.
// Dans un test, on construit un API avec un repository branché sur une base de
// test — sans toucher au code des handlers.
//
// Parallèle : en Django, une vue importe les modèles directement (état global).
// Ici tout est passé explicitement — plus verbeux, mais on voit toujours d'où
// vient chaque dépendance, et rien n'est caché.
type API struct {
	services *service.Repository
}

// New construit l'API avec ses dépendances.
func New(services *service.Repository) *API {
	return &API{services: services}
}

// Router construit le routeur HTTP et y branche les routes.
//
// Un routeur fait une seule chose : regarder la méthode (GET/POST/...) et le
// chemin d'une requête entrante, puis appeler le bon handler.  (≈ urls.py)
//
// Pourquoi chi plutôt que net/http seul : il apporte les middlewares et les
// paramètres d'URL (/services/{slug}) tout en restant 100 % compatible —
// un handler chi EST un http.HandlerFunc ordinaire. Rien de magique n'est caché,
// contrairement à un framework plus lourd type Gin.
func (a *API) Router() *chi.Mux {
	r := chi.NewRouter()

	// Traversés par CHAQUE requête, dans cet ordre, avant d'atteindre le handler
	// — puis en sens inverse pour la réponse.  (≈ app.use() d'Express)
	r.Use(middleware.RequestID) // identifiant unique par requête (traçabilité des logs)
	r.Use(middleware.RealIP)    // vraie IP client même derrière un proxy
	r.Use(middleware.Logger)    // journalise méthode, chemin, statut, durée
	r.Use(middleware.Recoverer) // un panic -> 500, au lieu de tuer le serveur

	r.Get("/health", handleHealth)

	// Les routes de l'API sont regroupées sous /api : le préfixe permettra
	// plus tard de servir le front React sur "/" sans collision de chemins.
	r.Route("/api", func(r chi.Router) {
		r.Route("/services", func(r chi.Router) {
			r.Get("/", a.handleListServices)
			r.Post("/", a.handleCreateService)
			r.Get("/{slug}", a.handleGetService)
		})
	})

	return r
}
