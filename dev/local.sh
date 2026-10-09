#!/usr/bin/env bash
# Run libra-web entirely on this machine. See dev/README.md
#
#   dev/local.sh up       start postgres + meilisearch and apply DB migrations
#   dev/local.sh seed     create a sample draft work and rebuild the search index
#   dev/local.sh reindex  rebuild the search index from everything in the local easystore
#   dev/local.sh start    run the stub services, backend and frontend dev server (ctrl-c stops all)
#   dev/local.sh down     stop the containers (data is kept)
#   dev/local.sh reset    stop the containers and delete all local data
#
# With no command: up, then start.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$ROOT/dev/docker-compose.yml"

DEV_USER="${DEV_USER:-dev1x}"
DEV_ROLE="${DEV_ROLE:-admin}"     # user | registrar | admin
DEV_FIRST="${DEV_FIRST:-Local}"
DEV_LAST="${DEV_LAST:-Developer}"
BACKEND_PORT=8085
STUB_PORT=8095

# Local S3 (see dev/docker-compose.yml). The /dev/null config files keep the AWS SDK from
# picking up any real AWS profile or credentials on this machine.
export AWS_ENDPOINT_URL=http://localhost:9000
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=local
export AWS_SECRET_ACCESS_KEY=local
export AWS_CONFIG_FILE=/dev/null
export AWS_SHARED_CREDENTIALS_FILE=/dev/null
export AWS_EC2_METADATA_DISABLED=true
export AWS_REQUEST_CHECKSUM_CALCULATION=when_required
export AWS_RESPONSE_CHECKSUM_VALIDATION=when_required
# the local postgres container does not use SSL
export PGSSLMODE=disable

compose() {
   if docker info >/dev/null 2>&1; then
      docker compose -f "$COMPOSE_FILE" "$@"
   elif command -v podman >/dev/null 2>&1; then
      if [ ! -S "${XDG_RUNTIME_DIR:-/run/user/$(id -u)}/podman/podman.sock" ]; then
         echo "podman compose needs the podman API socket. Start it with: systemctl --user start podman.socket" >&2
         exit 1
      fi
      podman compose -f "$COMPOSE_FILE" "$@"
   else
      echo "Neither a running docker daemon nor podman was found" >&2
      exit 1
   fi
}

cmd_up() {
   compose up -d --wait postgres meilisearch s3mock s3proxy
   compose run --rm migrate
}

cmd_seed() {
   (cd "$ROOT" && go run -tags service ./dev/seed -cid "$DEV_USER" -first "$DEV_FIRST" -last "$DEV_LAST" "$@")
}

cmd_start() {
   cd "$ROOT"
   if [ ! -d frontend/node_modules ]; then
      (cd frontend && npm ci)
   fi

   pids=()
   trap 'kill "${pids[@]}" 2>/dev/null; wait' EXIT INT TERM

   go run ./dev/stubsvc -port "$STUB_PORT" -first "$DEV_FIRST" -last "$DEV_LAST" &
   pids+=($!)

   go run -tags service ./backend \
      -port "$BACKEND_PORT" \
      -etdurl http://localhost:8080 \
      -jwtkey local-dev-only-key \
      -dbhost localhost -dbport 55432 -dbname libraweb -dbuser libraweb -dbpass pass \
      -deveasystoredb easystore -deveasystorebucket easystore-local \
      -index http://localhost:7700 \
      -userws "http://localhost:$STUB_PORT" \
      -depositauthurl "http://localhost:$STUB_PORT" \
      -getorcidurl "http://localhost:$STUB_PORT/orcid" \
      -orcidurl "http://localhost:$STUB_PORT/orcid-client" \
      -auditqueryurl "http://localhost:$STUB_PORT/audits" \
      -metricsqueryurl "http://localhost:$STUB_PORT/metrics" \
      -busname local -eventsrc libra-web-local -devbus \
      -devuser "$DEV_USER" -devrole "$DEV_ROLE" &
   pids+=($!)

   (cd frontend && LIBRA_SRV="http://localhost:$BACKEND_PORT" npm run dev) &
   pids+=($!)

   echo
   echo ">>> app: http://localhost:8080  (logged in as $DEV_USER, role $DEV_ROLE). ctrl-c to stop"
   echo
   wait -n
}

case "${1:-}" in
   up) cmd_up ;;
   seed) shift; cmd_seed "$@" ;;
   reindex) shift; cmd_seed -create=false "$@" ;;
   start) cmd_start ;;
   down) compose down ;;
   reset) compose down -v ;;
   "") cmd_up; cmd_start ;;
   *) sed -n '2,12p' "$0"; exit 1 ;;
esac
