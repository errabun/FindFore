#!/usr/bin/env bash
#
# Cloud SQL cost hygiene for FindFore.
#
# Sets up Cloud Scheduler jobs that stop/start the Cloud SQL instance so you
# only pay compute during weekday work hours. Storage + public IP still bill
# while stopped (~$11/mo). Cloud Run already scales to zero — leave it alone.
#
# Default schedule (America/Denver = Mountain Time):
#   Start  Mon–Fri 08:00  → activationPolicy ALWAYS
#   Stop   Mon–Fri 20:00  → activationPolicy NEVER
#   Stop   Sat–Sun 08:00  → safety net if you left it running over the weekend
#
# Usage:
#   ./scripts/gcp-sql-schedule.sh setup     # create SA + scheduler jobs (idempotent)
#   ./scripts/gcp-sql-schedule.sh start     # start Cloud SQL now (evening/weekend work)
#   ./scripts/gcp-sql-schedule.sh stop      # stop Cloud SQL now
#   ./scripts/gcp-sql-schedule.sh status    # show instance state + scheduled jobs
#   ./scripts/gcp-sql-schedule.sh teardown  # delete the scheduler jobs (keeps SA)

set -euo pipefail

PROJECT_ID="${PROJECT_ID:-findfore}"
REGION="${REGION:-us-central1}"
SQL_INSTANCE="${SQL_INSTANCE:-findfore-db}"
SCHEDULER_REGION="${SCHEDULER_REGION:-us-central1}"
TIME_ZONE="${TIME_ZONE:-America/Denver}"
SA_NAME="${SA_NAME:-findfore-sql-scheduler}"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

START_CRON="${START_CRON:-0 8 * * 1-5}"   # Mon–Fri 08:00
STOP_CRON="${STOP_CRON:-0 20 * * 1-5}"    # Mon–Fri 20:00
WEEKEND_STOP_CRON="${WEEKEND_STOP_CRON:-0 8 * * 0,6}"  # Sat–Sun 08:00

API_URL="https://sqladmin.googleapis.com/v1/projects/${PROJECT_ID}/instances/${SQL_INSTANCE}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'
log()  { echo -e "${GREEN}[+]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }

command -v gcloud >/dev/null 2>&1 || { warn "gcloud not found"; exit 1; }

gcloud config set project "$PROJECT_ID" >/dev/null

patch_activation() {
  local policy="$1"
  log "Setting ${SQL_INSTANCE} activationPolicy=${policy}"
  gcloud sql instances patch "$SQL_INSTANCE" \
    --activation-policy="$policy" \
    --quiet
}

create_or_update_job() {
  local name="$1"
  local schedule="$2"
  local policy="$3"
  local description="$4"
  local body

  body=$(printf '{"settings":{"activationPolicy":"%s"}}' "$policy")

  if gcloud scheduler jobs describe "$name" --location="$SCHEDULER_REGION" >/dev/null 2>&1; then
    log "Updating scheduler job '$name' ($schedule → $policy)"
    gcloud scheduler jobs update http "$name" \
      --location="$SCHEDULER_REGION" \
      --schedule="$schedule" \
      --time-zone="$TIME_ZONE" \
      --uri="$API_URL" \
      --http-method=PUT \
      --headers="Content-Type=application/json" \
      --message-body="$body" \
      --oauth-service-account-email="$SA_EMAIL" \
      --oauth-token-scope="https://www.googleapis.com/auth/cloud-platform" \
      --description="$description" \
      --quiet
  else
    log "Creating scheduler job '$name' ($schedule → $policy)"
    gcloud scheduler jobs create http "$name" \
      --location="$SCHEDULER_REGION" \
      --schedule="$schedule" \
      --time-zone="$TIME_ZONE" \
      --uri="$API_URL" \
      --http-method=PUT \
      --headers="Content-Type=application/json" \
      --message-body="$body" \
      --oauth-service-account-email="$SA_EMAIL" \
      --oauth-token-scope="https://www.googleapis.com/auth/cloud-platform" \
      --description="$description"
  fi
}

setup() {
  log "Enabling Cloud Scheduler API..."
  gcloud services enable cloudscheduler.googleapis.com sqladmin.googleapis.com

  if gcloud iam service-accounts describe "$SA_EMAIL" >/dev/null 2>&1; then
    log "Service account $SA_EMAIL already exists"
  else
    log "Creating service account $SA_EMAIL"
    gcloud iam service-accounts create "$SA_NAME" \
      --display-name="FindFore Cloud SQL start/stop scheduler"
  fi

  log "Granting roles/cloudsql.editor to $SA_EMAIL"
  gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/cloudsql.editor" \
    --condition=None >/dev/null

  create_or_update_job \
    "findfore-sql-start-weekday" \
    "$START_CRON" \
    "ALWAYS" \
    "Start FindFore Cloud SQL weekday mornings"

  create_or_update_job \
    "findfore-sql-stop-weekday" \
    "$STOP_CRON" \
    "NEVER" \
    "Stop FindFore Cloud SQL weekday evenings"

  create_or_update_job \
    "findfore-sql-stop-weekend" \
    "$WEEKEND_STOP_CRON" \
    "NEVER" \
    "Stop FindFore Cloud SQL on weekends if left running"

  echo
  log "Schedule installed (${TIME_ZONE}):"
  echo "  Start  Mon–Fri ${START_CRON}  → ALWAYS"
  echo "  Stop   Mon–Fri ${STOP_CRON} → NEVER"
  echo "  Stop   Sat–Sun ${WEEKEND_STOP_CRON} → NEVER"
  echo
  log "Manual overrides when you need the DB outside that window:"
  echo "  ./scripts/gcp-sql-schedule.sh start"
  echo "  ./scripts/gcp-sql-schedule.sh stop"
  echo "  ./scripts/gcp-sql-schedule.sh status"
}

status() {
  log "Cloud SQL instance:"
  gcloud sql instances describe "$SQL_INSTANCE" \
    --format='table(name,state,settings.activationPolicy,settings.tier)'

  echo
  log "Scheduler jobs:"
  gcloud scheduler jobs list --location="$SCHEDULER_REGION" \
    --filter='name:findfore-sql-' \
    --format='table(name.basename(),schedule,timeZone,state,httpTarget.uri)' \
    2>/dev/null || warn "No FindFore scheduler jobs found (run setup first)"
}

teardown() {
  warn "Deleting FindFore SQL scheduler jobs..."
  for name in findfore-sql-start-weekday findfore-sql-stop-weekday findfore-sql-stop-weekend; do
    if gcloud scheduler jobs describe "$name" --location="$SCHEDULER_REGION" >/dev/null 2>&1; then
      gcloud scheduler jobs delete "$name" --location="$SCHEDULER_REGION" --quiet
      log "Deleted $name"
    fi
  done
  warn "Service account $SA_EMAIL left in place (harmless). Delete manually if desired."
}

case "${1:-}" in
  setup)    setup ;;
  start)    patch_activation ALWAYS ;;
  stop)     patch_activation NEVER ;;
  status)   status ;;
  teardown) teardown ;;
  *)
    echo "Usage: $0 {setup|start|stop|status|teardown}"
    exit 1
    ;;
esac
