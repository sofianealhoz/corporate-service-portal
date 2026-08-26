// Point d'entrée : assemble config, base et routeur, gère le cycle de vie.
// Aucune logique métier ici, elle vit dans internal/.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sofianealhoz/corporate-service-portal/internal/api"
	"github.com/sofianealhoz/corporate-service-portal/internal/cache"
	"github.com/sofianealhoz/corporate-service-portal/internal/config"
	"github.com/sofianealhoz/corporate-service-portal/internal/db"
	"github.com/sofianealhoz/corporate-service-portal/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("base de données : %v", err)
	}
	defer pool.Close()
	log.Println("connecté à PostgreSQL")

	// suffisant pour une instance unique ; à sortir dans une étape de
	// déploiement si plusieurs instances démarrent en même temps
	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrations : %v", err)
	}
	log.Println("migrations à jour")

	// pas de vérification bloquante : l'API doit démarrer même sans Redis
	redis := cache.New(cfg.RedisURL)
	defer redis.Close()

	services := service.NewRepository(pool)
	handlers := api.New(services, redis)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handlers.Router(),
		// évite qu'une connexion ouverte sans headers immobilise une ressource
		ReadHeaderTimeout: 5 * time.Second,
	}

	// bloquant, donc lancé à part pour attendre le signal d'arrêt en dessous
	go func() {
		log.Printf("API démarrée sur http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("erreur serveur : %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// on laisse 10 s aux requêtes en cours plutôt que de les couper
	log.Println("arrêt en cours...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("arrêt forcé : %v", err)
	}
	log.Println("serveur arrêté proprement")
}
