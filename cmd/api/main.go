// Commande `api` : point d'entrée du serveur HTTP.
//
// Ce fichier ne contient AUCUNE logique métier. Son unique rôle est d'assembler
// les morceaux (config -> routeur -> serveur) et de gérer le cycle de vie du
// processus (démarrage, arrêt propre). Toute la logique vit dans internal/.
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
	"github.com/sofianealhoz/corporate-service-portal/internal/config"
	"github.com/sofianealhoz/corporate-service-portal/internal/db"
	"github.com/sofianealhoz/corporate-service-portal/internal/service"
)

func main() {
	cfg := config.Load()

	// Contexte de démarrage : sert de parent aux opérations d'initialisation.
	ctx := context.Background()

	// 1. Connexion à PostgreSQL (échoue tout de suite si la base est injoignable).
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("base de données : %v", err)
	}
	defer pool.Close()
	log.Println("connecté à PostgreSQL")

	// 2. Migrations : la structure de la base est mise à jour au démarrage.
	//    Choix assumé : simple et suffisant pour un service unique. À plus grande
	//    échelle (plusieurs instances qui démarrent ensemble), on sortirait ça
	//    dans une étape de déploiement dédiée.
	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrations : %v", err)
	}
	log.Println("migrations à jour")

	// 3. Assemblage : le repository reçoit le pool, l'API reçoit le repository.
	//    Chaque couche déclare ce dont elle a besoin, rien n'est global.
	services := service.NewRepository(pool)
	handlers := api.New(services)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handlers.Router(),
		// Garde-fou : un client qui ouvre une connexion puis n'envoie jamais
		// ses headers immobiliserait une ressource indéfiniment (attaque Slowloris).
		ReadHeaderTimeout: 5 * time.Second,
	}

	// ListenAndServe est BLOQUANT : on le lance dans une goroutine (un fil
	// d'exécution léger géré par Go) pour que main puisse continuer et se mettre
	// en attente du signal d'arrêt juste en dessous.
	go func() {
		log.Printf("API démarrée sur http://localhost:%s", cfg.Port)
		// À l'arrêt volontaire, ListenAndServe renvoie ErrServerClosed :
		// ce n'est pas une erreur, on ne doit donc pas paniquer dessus.
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("erreur serveur : %v", err)
		}
	}()

	// On attend un signal d'arrêt du système : Ctrl+C (Interrupt) ou SIGTERM
	// (ce que Docker/Kubernetes envoient pour demander l'arrêt d'un conteneur).
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop // bloque ici jusqu'à réception du signal

	// Arrêt PROPRE (graceful shutdown) : on cesse d'accepter de nouvelles
	// requêtes, mais on laisse jusqu'à 10 s aux requêtes en cours pour finir.
	// Sans ça, un déploiement couperait des requêtes utilisateur en plein vol.
	log.Println("arrêt en cours...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("arrêt forcé : %v", err)
	}
	log.Println("serveur arrêté proprement")
}
