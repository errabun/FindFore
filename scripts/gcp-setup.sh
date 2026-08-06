#!/usr/bin/env bash
#
# One-time Google Cloud setup for FindFore.
# Idempotent: safe to re-run. Skips resources that already exist.
#
# Prereqs:
#   - gcloud CLI installed (`brew install --cask google-cloud-sdk`)
#   - Authenticated: `gcloud auth login`
#
# Usage:
#   ./scripts/gcp-setup.sh

set -euo pipefail

# ==============================================================================
# Config
# ==============================================================================
PROJECT_ID="${PROJECT_ID:-findfore}"
REGION="${REGION:-us-central1}"
SQL_INSTANCE="${SQL_INSTANCE:-findfore-db}"
SQL_TIER="${SQL_TIER:-db-f1-micro}"
DB_NAME="${DB_NAME:-findfore}"
DB_USER="${DB_USER:-findfore}"
AR_REPO="${AR_REPO:-findfore}"
RUN_SA_NAME="${RUN_SA_NAME:-findfore-run}"
SERVICE_NAME="${SERVICE_NAME:-findfore}"

# ==============================================================================
# Helpers
# ==============================================================================
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log()  { echo -e "${GREEN}[+]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }
die()  { echo -e "${RED}[x]${NC} $*" >&2; exit 1; }

command -v gcloud >/dev/null 2>&1 || die "gcloud CLI not found. Install with: brew install --cask google-cloud-sdk"
command -v openssl >/dev/null 2>&1 || die "openssl not found."

if ! gcloud auth list --filter=status:ACTIVE --format='value(account)' | grep -q .; then
  die "No active gcloud auth. Run: gcloud auth login"
fi

# ==============================================================================
# 1. Set active project
# ==============================================================================
log "Setting active project to $PROJECT_ID"
gcloud config set project "$PROJECT_ID" >/dev/null
PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')

# ==============================================================================
# 2. Enable APIs
# ==============================================================================
log "Enabling required APIs (this can take a minute the first time)..."
gcloud services enable \
  run.googleapis.com \
  sqladmin.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com \
  storage.googleapis.com \
  iamcredentials.googleapis.com \
  compute.googleapis.com

# ==============================================================================
# 3. Artifact Registry
# ==============================================================================
if gcloud artifacts repositories describe "$AR_REPO" --location="$REGION" >/dev/null 2>&1; then
  log "Artifact Registry repo '$AR_REPO' already exists"
else
  log "Creating Artifact Registry repo '$AR_REPO'"
  gcloud artifacts repositories create "$AR_REPO" \
    --repository-format=docker \
    --location="$REGION" \
    --description="FindFore container images"
fi

# ==============================================================================
# 4. Cloud SQL instance
# ==============================================================================
if gcloud sql instances describe "$SQL_INSTANCE" >/dev/null 2>&1; then
  log "Cloud SQL instance '$SQL_INSTANCE' already exists"
else
  log "Creating Cloud SQL instance '$SQL_INSTANCE' (~5 minutes)..."
  # ENTERPRISE edition is required for shared-core tiers (db-f1-micro / db-g1-small).
  # ENTERPRISE_PLUS only supports dedicated cores. Cloud Run reaches the DB via a
  # secure Unix socket, so the public IP isn't exposed to arbitrary clients.
  gcloud sql instances create "$SQL_INSTANCE" \
    --edition=ENTERPRISE \
    --database-version=POSTGRES_16 \
    --tier="$SQL_TIER" \
    --region="$REGION" \
    --storage-size=10GB \
    --storage-type=SSD \
    --backup \
    --backup-start-time=07:00
fi

INSTANCE_CONNECTION_NAME="${PROJECT_ID}:${REGION}:${SQL_INSTANCE}"
log "Instance connection name: ${INSTANCE_CONNECTION_NAME}"

# ==============================================================================
# 5. Database and application user
# ==============================================================================
if gcloud sql databases describe "$DB_NAME" --instance="$SQL_INSTANCE" >/dev/null 2>&1; then
  log "Database '$DB_NAME' already exists"
else
  log "Creating database '$DB_NAME'"
  gcloud sql databases create "$DB_NAME" --instance="$SQL_INSTANCE"
fi

DB_PASSWORD=$(openssl rand -base64 30 | tr -d '=+/' | cut -c1-24)
if gcloud sql users list --instance="$SQL_INSTANCE" --format='value(name)' | grep -qx "$DB_USER"; then
  log "DB user '$DB_USER' exists — rotating password"
else
  log "Creating DB user '$DB_USER'"
  gcloud sql users create "$DB_USER" --instance="$SQL_INSTANCE" --password="$DB_PASSWORD"
fi
gcloud sql users set-password "$DB_USER" --instance="$SQL_INSTANCE" --password="$DB_PASSWORD" >/dev/null

# ==============================================================================
# 6. Secret Manager
# ==============================================================================
create_or_update_secret() {
  local name="$1"
  local value="$2"
  if gcloud secrets describe "$name" >/dev/null 2>&1; then
    printf '%s' "$value" | gcloud secrets versions add "$name" --data-file=- >/dev/null
    log "Updated secret '$name'"
  else
    printf '%s' "$value" | gcloud secrets create "$name" \
      --data-file=- --replication-policy=automatic >/dev/null
    log "Created secret '$name'"
  fi
}

JWT_SECRET=$(openssl rand -base64 48 | tr -d '\n')
create_or_update_secret "findfore-jwt-secret"  "$JWT_SECRET"
create_or_update_secret "findfore-db-password" "$DB_PASSWORD"

if ! gcloud secrets describe findfore-golf-course-api-key >/dev/null 2>&1; then
  printf 'REPLACE_ME' | gcloud secrets create findfore-golf-course-api-key \
    --data-file=- --replication-policy=automatic >/dev/null
  warn "Created placeholder 'findfore-golf-course-api-key'. Update it with:"
  warn "  printf 'YOUR_REAL_KEY' | gcloud secrets versions add findfore-golf-course-api-key --data-file=-"
fi

# ==============================================================================
# 7. Cloud Run runtime service account
# ==============================================================================
RUN_SA_EMAIL="${RUN_SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
if gcloud iam service-accounts describe "$RUN_SA_EMAIL" >/dev/null 2>&1; then
  log "Service account '$RUN_SA_EMAIL' already exists"
else
  log "Creating service account '$RUN_SA_EMAIL'"
  gcloud iam service-accounts create "$RUN_SA_NAME" \
    --display-name="FindFore Cloud Run runtime"
fi

log "Granting runtime roles to $RUN_SA_EMAIL"
for role in roles/cloudsql.client roles/secretmanager.secretAccessor roles/storage.objectAdmin; do
  gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:${RUN_SA_EMAIL}" \
    --role="$role" \
    --condition=None >/dev/null
done

for secret in findfore-jwt-secret findfore-db-password findfore-golf-course-api-key; do
  gcloud secrets add-iam-policy-binding "$secret" \
    --member="serviceAccount:${RUN_SA_EMAIL}" \
    --role="roles/secretmanager.secretAccessor" \
    --condition=None >/dev/null
done

# ==============================================================================
# 8. Cloud Build permissions
# ==============================================================================
log "Granting Cloud Build permissions..."
CB_DEFAULT_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"
CB_LEGACY_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"
for sa in "$CB_DEFAULT_SA" "$CB_LEGACY_SA"; do
  for role in roles/run.admin roles/iam.serviceAccountUser roles/artifactregistry.writer roles/cloudsql.client roles/secretmanager.secretAccessor; do
    gcloud projects add-iam-policy-binding "$PROJECT_ID" \
      --member="serviceAccount:${sa}" \
      --role="$role" \
      --condition=None >/dev/null 2>&1 || true
  done
done

# ==============================================================================
# 9. Summary
# ==============================================================================
echo
echo -e "${GREEN}✓ Phase C complete!${NC}"
cat <<EOF

  Project ID:              ${PROJECT_ID}
  Region:                  ${REGION}
  Cloud SQL instance:      ${SQL_INSTANCE}
  Instance connection:     ${INSTANCE_CONNECTION_NAME}
  Database:                ${DB_NAME}
  DB user:                 ${DB_USER}
  Artifact Registry:       ${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}
  Runtime service account: ${RUN_SA_EMAIL}

Secrets in Secret Manager:
  findfore-jwt-secret            (generated)
  findfore-db-password           (generated)
  findfore-golf-course-api-key   (placeholder — replace before deploy)

Next:
  1. Add your real Golf Course API key:
       printf 'YOUR_KEY' | gcloud secrets versions add findfore-golf-course-api-key --data-file=-

  2. Run database migrations:
       ./scripts/gcp-migrate.sh

  3. Deploy the first revision:
       gcloud builds submit --config=cloudbuild.yaml

EOF
