#!/usr/bin/env bash
# Démarre tout ce qu'il faut pour une démonstration : base, cache, catalogue
# rempli, API et front. Ctrl+C arrête l'ensemble.
#
# Se décale automatiquement si un port est déjà pris sur la machine.
set -euo pipefail
cd "$(dirname "$0")/.."

libre() { ! (exec 3<>"/dev/tcp/127.0.0.1/$1") 2>/dev/null; }

choisir() {
  local port=$1
  while ! libre "$port"; do port=$((port + 1)); done
  echo "$port"
}

export POSTGRES_PORT=${POSTGRES_PORT:-$(choisir 5432)}
export REDIS_PORT=${REDIS_PORT:-$(choisir 6379)}
API_PORT=${PORT:-$(choisir 8080)}

export DATABASE_URL="postgres://portal:portal@localhost:${POSTGRES_PORT}/portal?sslmode=disable"
export REDIS_URL="redis://localhost:${REDIS_PORT}"
export PORT="$API_PORT"
export API_PROXY_TARGET="http://localhost:${API_PORT}"

echo "PostgreSQL :${POSTGRES_PORT}  Redis :${REDIS_PORT}  API :${API_PORT}"
docker compose up -d

echo "attente de la base..."
until docker compose exec -T db pg_isready -U portal -d portal >/dev/null 2>&1; do sleep 1; done

go run ./cmd/seed -reset
# le listing est caché 30 s : on vide le cache pour ne pas servir l'ancien
docker compose exec -T redis redis-cli FLUSHALL >/dev/null

[ -d web/node_modules ] || npm --prefix web ci

echo
echo "  front : http://localhost:5173"
echo
trap 'kill 0' EXIT INT TERM
go run ./cmd/api &
npm --prefix web run dev &
wait
