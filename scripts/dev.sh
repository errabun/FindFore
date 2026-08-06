#!/usr/bin/env bash
#
# One-command local dev startup.
#
# Prereqs (first time only):
#   - Docker Desktop OR Podman running (see below)
#   - Node 22+ and Go 1.25+ installed
#   - Copy .env.example to .env and fill in real values
#
# Podman on macOS: first-time users must init and start the podman VM:
#   podman machine init
#   podman machine start
#
# Usage:
#   ./scripts/dev.sh          # start db + backend + frontend
#   ./scripts/dev.sh migrate  # run pending migrations against local db
#   ./scripts/dev.sh psql     # open a psql shell to local db
#   ./scripts/dev.sh reset    # nuke local db (destructive)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'
log()  { echo -e "${GREEN}[+]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }

require() { command -v "$1" >/dev/null 2>&1 || { warn "$1 not found"; exit 1; }; }
require go
require npm

# ------------------------------------------------------------------------------
# Detect container runtime and compose command.
# COMPOSE_CMD is set to an array so it works whether the tool is "docker compose"
# (two words) or "podman-compose" (single command).
# ------------------------------------------------------------------------------
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  RUNTIME=docker
  if docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD=(docker compose)
  else
    warn "docker compose plugin not found. Install Docker Desktop or 'brew install docker-compose'."
    exit 1
  fi
elif command -v podman >/dev/null 2>&1; then
  RUNTIME=podman
  # Podman on macOS needs a running VM
  if ! podman machine list --format '{{.Running}}' 2>/dev/null | grep -q true; then
    warn "Podman is installed but no machine is running. Start it:"
    warn "  podman machine init   # first time only"
    warn "  podman machine start"
    exit 1
  fi
  if podman compose version >/dev/null 2>&1; then
    COMPOSE_CMD=(podman compose)
  elif command -v podman-compose >/dev/null 2>&1; then
    COMPOSE_CMD=(podman-compose)
  else
    warn "No compose command found for podman. Install with: brew install podman-compose"
    exit 1
  fi
else
  warn "No container runtime found. Install one of:"
  warn "  brew install --cask docker           # Docker Desktop"
  warn "  brew install podman podman-compose   # Podman"
  exit 1
fi

log "Using $RUNTIME (${COMPOSE_CMD[*]})"

# ------------------------------------------------------------------------------
# Load .env so `go run` / migrations see the same values.
# ------------------------------------------------------------------------------
if [[ ! -f .env ]]; then
  warn ".env not found. Copy from .env.example and set JWT_SECRET + GOLF_COURSE_API_KEY:"
  warn "  cp .env.example .env"
  exit 1
fi
set -a
# shellcheck disable=SC1091
source .env
set +a

# ------------------------------------------------------------------------------
# Helpers
# ------------------------------------------------------------------------------
db_exec() {
  # Runs a psql-style command inside the db container.
  "${COMPOSE_CMD[@]}" exec -T db "$@"
}

start_db() {
  local running
  running=$("${COMPOSE_CMD[@]}" ps --status running --services 2>/dev/null || true)
  if echo "$running" | grep -qx db; then
    log "Postgres already running"
  else
    log "Starting Postgres container..."
    "${COMPOSE_CMD[@]}" up -d db
  fi

  log "Waiting for Postgres to accept connections..."
  for _ in {1..30}; do
    if db_exec pg_isready -U findfore -d findfore >/dev/null 2>&1; then
      log "Postgres is ready"
      return
    fi
    sleep 1
  done
  warn "Postgres did not become ready in 30s"
  exit 1
}

# ------------------------------------------------------------------------------
# Commands
# ------------------------------------------------------------------------------
case "${1:-up}" in
  up)
    start_db
    log "Applying migrations..."
    go run ./cmd/migrate -direction up

    log "Starting Go backend on :${PORT:-8080}"
    ( go run . & echo $! > /tmp/findfore-backend.pid )

    log "Starting Vite frontend on :3000"
    log "Open http://localhost:3000 in your browser. Ctrl+C to stop everything."

    cleanup() {
      log "Stopping backend..."
      if [[ -f /tmp/findfore-backend.pid ]]; then
        kill "$(cat /tmp/findfore-backend.pid)" 2>/dev/null || true
        rm -f /tmp/findfore-backend.pid
      fi
    }
    trap cleanup EXIT

    ( cd frontend && npm run dev )
    ;;

  migrate)
    start_db
    go run ./cmd/migrate -direction up
    ;;

  psql)
    "${COMPOSE_CMD[@]}" exec db psql -U findfore -d findfore
    ;;

  reset)
    warn "This will DELETE all local development data."
    read -r -p "Type 'yes' to continue: " CONFIRM
    [[ "$CONFIRM" == "yes" ]] || { echo "Aborted."; exit 1; }
    "${COMPOSE_CMD[@]}" down -v
    log "Local database reset. Run './scripts/dev.sh up' to recreate."
    ;;

  *)
    echo "Usage: $0 [up|migrate|psql|reset]"
    exit 1
    ;;
esac
