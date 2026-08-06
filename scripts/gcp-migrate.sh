#!/usr/bin/env bash
#
# Run FindFore database migrations against Cloud SQL via the Cloud SQL Auth Proxy.
# Uses the same golang-migrate CLI defined in cmd/migrate.
#
# Prereqs:
#   - Phase C setup complete (./scripts/gcp-setup.sh)
#   - cloud-sql-proxy installed:
#       brew install cloud-sql-proxy
#     or download: https://cloud.google.com/sql/docs/postgres/sql-proxy
#
# Usage:
#   ./scripts/gcp-migrate.sh              # migrate up
#   ./scripts/gcp-migrate.sh down 1       # roll back one migration

set -euo pipefail

PROJECT_ID="${PROJECT_ID:-findfore}"
REGION="${REGION:-us-central1}"
SQL_INSTANCE="${SQL_INSTANCE:-findfore-db}"
DB_NAME="${DB_NAME:-findfore}"
DB_USER="${DB_USER:-findfore}"
LOCAL_PORT="${LOCAL_PORT:-5433}"

DIRECTION="${1:-up}"
STEPS="${2:-0}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'
log()  { echo -e "${GREEN}[+]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }

command -v cloud-sql-proxy >/dev/null 2>&1 || {
  warn "cloud-sql-proxy not found. Install with:"
  warn "  brew install cloud-sql-proxy"
  exit 1
}
command -v go >/dev/null 2>&1 || { warn "go not found"; exit 1; }

INSTANCE_CONNECTION_NAME="${PROJECT_ID}:${REGION}:${SQL_INSTANCE}"

log "Fetching DB password from Secret Manager..."
DB_PASSWORD=$(gcloud secrets versions access latest --secret=findfore-db-password --project="$PROJECT_ID")

log "Starting Cloud SQL Auth Proxy on 127.0.0.1:${LOCAL_PORT}"
cloud-sql-proxy --port "$LOCAL_PORT" "$INSTANCE_CONNECTION_NAME" &
PROXY_PID=$!
trap 'kill $PROXY_PID 2>/dev/null || true' EXIT

# Wait for proxy to be listening
for i in {1..30}; do
  if nc -z 127.0.0.1 "$LOCAL_PORT" 2>/dev/null; then break; fi
  sleep 0.5
done

log "Running migrations (${DIRECTION}${STEPS:+ steps=$STEPS})"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@127.0.0.1:${LOCAL_PORT}/${DB_NAME}?sslmode=disable" \
JWT_SECRET="not-used-for-migrations" \
go run ./cmd/migrate -direction="$DIRECTION" ${STEPS:+-steps="$STEPS"}

log "Migrations complete"
