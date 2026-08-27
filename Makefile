# Une seule chaîne d'outils pour les deux applications du dépôt : on ne devrait
# pas avoir à savoir laquelle est en Go et laquelle est en TypeScript pour la
# construire ou la tester.

SHELL := /bin/bash
.DEFAULT_GOAL := help

.PHONY: help up down dev build test test-go check-web lint clean

help: ## Affiche les commandes disponibles
	@grep -hE '^[a-z-]+:.*?## ' $(MAKEFILE_LIST) \
		| sed -e 's/:.*## /\t/' \
		| sort \
		| awk -F'\t' '{ printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }'

up: ## Démarre PostgreSQL et Redis
	docker compose up -d

down: ## Arrête PostgreSQL et Redis, données conservées
	docker compose down

# cible fichier : npm ci ne tourne que si le verrou a bougé
web/node_modules: web/package-lock.json
	npm ci --prefix web
	@touch web/node_modules

dev: up web/node_modules ## Démarre base, cache, API et front, Ctrl+C arrête tout
	@echo "API sur :8080, front sur :5173"
	@trap 'kill 0' EXIT INT TERM; \
		go run ./cmd/api & \
		npm --prefix web run dev & \
		wait

build: web/node_modules ## Compile l'API dans bin/ et le front dans web/dist/
	go build -o bin/api ./cmd/api
	npm --prefix web run build

test: test-go check-web ## Vérifie les deux applications

test-go: ## Tests Go, unitaires et d'intégration si les services tournent
	go test ./...

check-web: web/node_modules ## Contrôle du front : types puis lint
	npm --prefix web run typecheck
	npm --prefix web run lint

lint: check-web ## Analyse statique des deux applications
	go vet ./...
	@test -z "$$(gofmt -l cmd internal)" || { echo "gofmt requis :"; gofmt -l cmd internal; exit 1; }

clean: ## Supprime les artefacts de build
	rm -rf bin web/dist
	go clean -testcache
