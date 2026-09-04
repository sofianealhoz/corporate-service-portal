// Remplit le catalogue avec un jeu de services de démonstration.
// Idempotent : un slug déjà présent est ignoré, on peut le relancer.
package main

import (
	"context"
	"errors"
	"flag"
	"log"

	"github.com/sofianealhoz/corporate-service-portal/internal/config"
	"github.com/sofianealhoz/corporate-service-portal/internal/db"
	"github.com/sofianealhoz/corporate-service-portal/internal/service"
)

var catalogue = []service.CreateInput{
	{
		Slug: "audit-securite", Title: "Audit de sécurité",
		Description: "Revue complète de votre infrastructure et de vos applications, suivie d'un rapport priorisé des vulnérabilités.",
		Tier:        "premium", DurationH: 40, PriceCents: 1250000, Published: true,
	},
	{
		Slug: "migration-cloud", Title: "Migration vers le cloud",
		Description: "Reprise d'un parc existant vers une infrastructure managée, sans interruption de service.",
		Tier:        "enterprise", DurationH: 120, PriceCents: 4800000, Published: true,
	},
	{
		Slug: "refonte-site-vitrine", Title: "Refonte de site vitrine",
		Description: "Conception et développement d'un site rapide, accessible et administrable par vos équipes.",
		Tier:        "standard", DurationH: 60, PriceCents: 890000, Published: true,
	},
	{
		Slug: "integration-api", Title: "Intégration d'API",
		Description: "Connexion de vos outils métier entre eux : CRM, facturation, messagerie, avec supervision des échanges.",
		Tier:        "premium", DurationH: 35, PriceCents: 1150000, Published: true,
	},
	{
		Slug: "formation-equipe", Title: "Formation des équipes",
		Description: "Trois jours de montée en compétence sur vos outils, en présentiel, exercices sur vos propres données.",
		Tier:        "standard", DurationH: 21, PriceCents: 450000, Published: true,
	},
	{
		Slug: "supervision-24-7", Title: "Supervision 24/7",
		Description: "Surveillance continue de vos services, alertes qualifiées et intervention d'astreinte.",
		Tier:        "enterprise", DurationH: 200, PriceCents: 6200000, Published: true,
	},
	{
		Slug: "optimisation-base-donnees", Title: "Optimisation de base de données",
		Description: "Analyse des requêtes lentes, indexation et plan de montée en charge documenté.",
		Tier:        "premium", DurationH: 28, PriceCents: 980000, Published: true,
	},
	{
		Slug: "conformite-rgpd", Title: "Mise en conformité RGPD",
		Description: "Cartographie des traitements, registre, politique de conservation et plan de mise en conformité.",
		Tier:        "premium", DurationH: 45, PriceCents: 1380000, Published: true,
	},
	{
		Slug: "tableau-de-bord", Title: "Tableau de bord décisionnel",
		Description: "Agrégation de vos sources de données en indicateurs suivis au quotidien par la direction.",
		Tier:        "standard", DurationH: 50, PriceCents: 760000, Published: true,
	},
	{
		Slug: "reprise-de-donnees", Title: "Reprise de données",
		Description: "Extraction, nettoyage et import de votre historique vers un nouvel outil, avec contrôle de cohérence.",
		Tier:        "standard", DurationH: 30, PriceCents: 520000, Published: true,
	},
	{
		Slug: "plan-reprise-activite", Title: "Plan de reprise d'activité",
		Description: "Sauvegardes testées, procédure de bascule documentée et exercice de restauration annuel.",
		Tier:        "enterprise", DurationH: 80, PriceCents: 3100000, Published: true,
	},
	// non publiés : servent à montrer que la vitrine filtre les brouillons
	{
		Slug: "accompagnement-ia", Title: "Accompagnement IA",
		Description: "Cadrage des cas d'usage, prototype et mise en production d'un premier assistant métier.",
		Tier:        "enterprise", DurationH: 90, PriceCents: 3900000, Published: false,
	},
	{
		Slug: "audit-accessibilite", Title: "Audit d'accessibilité",
		Description: "Vérification RGAA de vos parcours principaux et corrections priorisées.",
		Tier:        "standard", DurationH: 18, PriceCents: 390000, Published: false,
	},
}

func main() {
	reset := flag.Bool("reset", false, "vide le catalogue avant d'insérer, pour repartir d'un état connu")
	flag.Parse()

	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("base de données : %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrations : %v", err)
	}

	repo := service.NewRepository(pool)

	if *reset {
		if err := repo.DeleteAll(ctx); err != nil {
			log.Fatalf("%v", err)
		}
		log.Println("catalogue vidé")
	}

	var ajoutes, existants int

	for _, in := range catalogue {
		switch _, err := repo.Create(ctx, in); {
		case err == nil:
			ajoutes++
		case errors.Is(err, service.ErrSlugTaken):
			existants++
		default:
			log.Fatalf("insertion de %s : %v", in.Slug, err)
		}
	}

	log.Printf("catalogue prêt : %d ajoutés, %d déjà présents", ajoutes, existants)
}
