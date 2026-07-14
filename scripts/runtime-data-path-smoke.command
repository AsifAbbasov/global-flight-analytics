#!/bin/zsh
set -euo pipefail

PROJECT_ROOT="${PROJECT_ROOT:-/Users/asifabbasov/Documents/global-flight-analytics}"
API_DIR="$PROJECT_ROOT/apps/api"
SMOKE_PORT="${RUNTIME_SMOKE_PORT:-18080}"
SMOKE_RADIUS="${RUNTIME_SMOKE_RADIUS:-250}"
BASE_URL="http://127.0.0.1:${SMOKE_PORT}/api/v1"

# Safe non-secret defaults matching apps/api/.env.example.
# Existing shell values and values loaded by the Go commands remain authoritative.
export DATABASE_CONNECT_TIMEOUT="${DATABASE_CONNECT_TIMEOUT:-5s}"

export OURAIRPORTS_TIMEOUT="${OURAIRPORTS_TIMEOUT:-30s}"
export OURAIRPORTS_COUNTRY_CODES="${OURAIRPORTS_COUNTRY_CODES:-AZ,GE,AM,TR}"

export TRAFFIC_INGESTION_LATITUDE="${TRAFFIC_INGESTION_LATITUDE:-40.4093}"
export TRAFFIC_INGESTION_LONGITUDE="${TRAFFIC_INGESTION_LONGITUDE:-49.8671}"
export TRAFFIC_INGESTION_RADIUS="${TRAFFIC_INGESTION_RADIUS:-100}"
export AIRPLANES_LIVE_TIMEOUT="${AIRPLANES_LIVE_TIMEOUT:-10s}"
export TRAJECTORY_MAX_TIME_GAP="${TRAJECTORY_MAX_TIME_GAP:-90s}"
export TRAJECTORY_MAX_GROUND_SPEED_MPS="${TRAJECTORY_MAX_GROUND_SPEED_MPS:-420}"

export API_ALLOWED_ORIGINS="${API_ALLOWED_ORIGINS:-http://localhost:3000,http://localhost:3001}"
export API_BODY_LIMIT_BYTES="${API_BODY_LIMIT_BYTES:-1048576}"
export API_READ_TIMEOUT="${API_READ_TIMEOUT:-10s}"
export API_WRITE_TIMEOUT="${API_WRITE_TIMEOUT:-15s}"
export API_IDLE_TIMEOUT="${API_IDLE_TIMEOUT:-60s}"
export API_RATE_LIMIT_MAX="${API_RATE_LIMIT_MAX:-120}"
export API_RATE_LIMIT_WINDOW="${API_RATE_LIMIT_WINDOW:-1m}"
export OPEN_METEO_TIMEOUT="${OPEN_METEO_TIMEOUT:-10s}"

if [[ "${RUNTIME_SMOKE_ALLOW_DATABASE_WRITE:-}" != "1" ]]; then
  echo "ERROR: set RUNTIME_SMOKE_ALLOW_DATABASE_WRITE=1 to permit one bounded live ingestion cycle."
  exit 1
fi

if [[ ! -d "$PROJECT_ROOT/.git" ]]; then
  echo "ERROR: project repository was not found at $PROJECT_ROOT"
  exit 1
fi

if [[ ! -d "$API_DIR" ]]; then
  echo "ERROR: backend directory was not found at $API_DIR"
  exit 1
fi

for command_name in go curl node git; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "ERROR: required command is unavailable: $command_name"
    exit 1
  fi
done

if [[ ! -f "$API_DIR/.env" ]]; then
  echo "ERROR: backend environment file was not found at $API_DIR/.env"
  exit 1
fi

if ! grep -Eq '^[[:space:]]*DATABASE_URL=[^[:space:]].*$' "$API_DIR/.env" && [[ -z "${DATABASE_URL:-}" ]]; then
  echo "ERROR: DATABASE_URL is missing from apps/api/.env and the shell environment"
  exit 1
fi

if [[ "$(git -C "$PROJECT_ROOT" branch --show-current)" != "main" ]]; then
  echo "ERROR: runtime smoke must run from main"
  exit 1
fi

EXPECTED_PATH="scripts/runtime-data-path-smoke.command"

while IFS= read -r status_line; do
  [[ -z "$status_line" ]] && continue
  changed_file="${status_line:3}"

  if [[ "$changed_file" != "$EXPECTED_PATH" ]]; then
    echo "ERROR: runtime smoke found an unrelated worktree change: $status_line"
    exit 1
  fi
done < <(git -C "$PROJECT_ROOT" status --short --untracked-files=all)

TMP_DIR="$(mktemp -d)"
INGEST_PID=""
SERVER_PID=""

cleanup() {
  local exit_code=$?

  if [[ -n "$INGEST_PID" ]] && kill -0 "$INGEST_PID" >/dev/null 2>&1; then
    kill -TERM "$INGEST_PID" >/dev/null 2>&1 || true
    wait "$INGEST_PID" >/dev/null 2>&1 || true
  fi

  if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" >/dev/null 2>&1; then
    kill -TERM "$SERVER_PID" >/dev/null 2>&1 || true
    wait "$SERVER_PID" >/dev/null 2>&1 || true
  fi

  if [[ $exit_code -ne 0 ]]; then
    if [[ -f "$TMP_DIR/ingest.log" ]]; then
      echo "----- bounded ingestion log -----"
      cat "$TMP_DIR/ingest.log"
      echo "----- end bounded ingestion log -----"
    fi

    if [[ -f "$TMP_DIR/server.log" ]]; then
      echo "----- runtime smoke server log -----"
      cat "$TMP_DIR/server.log"
      echo "----- end runtime smoke server log -----"
    fi
  fi

  rm -rf "$TMP_DIR"
  exit $exit_code
}

trap cleanup EXIT INT TERM

validate_success_envelope() {
  local body_file="$1"

  node - "$body_file" <<'NODE'
const fs = require("fs");
const path = process.argv[2];
const raw = fs.readFileSync(path, "utf8");
let payload;

try {
  payload = JSON.parse(raw);
} catch (error) {
  console.error(`ERROR: response is not valid JSON: ${error.message}`);
  process.exit(1);
}

if (payload === null || typeof payload !== "object") {
  console.error("ERROR: response payload must be a JSON object");
  process.exit(1);
}

if (payload.success !== true) {
  console.error(`ERROR: response success must be true: ${raw}`);
  process.exit(1);
}

if (!("data" in payload)) {
  console.error(`ERROR: response data field is missing: ${raw}`);
  process.exit(1);
}
NODE
}

validate_error_envelope() {
  local body_file="$1"

  node - "$body_file" <<'NODE'
const fs = require("fs");
const path = process.argv[2];
const raw = fs.readFileSync(path, "utf8");
let payload;

try {
  payload = JSON.parse(raw);
} catch (error) {
  console.error(`ERROR: error response is not valid JSON: ${error.message}`);
  process.exit(1);
}

if (
  payload === null ||
  typeof payload !== "object" ||
  payload.success !== false ||
  payload.error === null ||
  typeof payload.error !== "object" ||
  typeof payload.error.code !== "string" ||
  payload.error.code.trim() === "" ||
  typeof payload.error.message !== "string" ||
  payload.error.message.trim() === ""
) {
  console.error(`ERROR: invalid error envelope: ${raw}`);
  process.exit(1);
}
NODE
}

validate_request_id() {
  local headers_file="$1"
  local request_path="$2"

  if ! grep -Eiq '^X-Request-ID:[[:space:]]*[^[:space:]]+' "$headers_file"; then
    echo "ERROR: $request_path did not return a non-empty X-Request-ID header"
    return 1
  fi
}

request_json() {
  local request_path="$1"
  local name="$2"
  local body_file="$TMP_DIR/${name}.json"
  local headers_file="$TMP_DIR/${name}.headers"
  local status_code

  status_code="$(curl \
    --silent \
    --show-error \
    --output "$body_file" \
    --dump-header "$headers_file" \
    --write-out "%{http_code}" \
    "$BASE_URL$request_path")"

  if [[ "$status_code" -lt 200 || "$status_code" -ge 300 ]]; then
    echo "ERROR: $request_path returned HTTP $status_code"
    cat "$body_file"
    return 1
  fi

  validate_success_envelope "$body_file"
  validate_request_id "$headers_file" "$request_path"
  echo "$body_file"
}

request_json_allow_unavailable() {
  local request_path="$1"
  local name="$2"
  local body_file="$TMP_DIR/${name}.json"
  local headers_file="$TMP_DIR/${name}.headers"
  local status_code

  status_code="$(curl \
    --silent \
    --show-error \
    --output "$body_file" \
    --dump-header "$headers_file" \
    --write-out "%{http_code}" \
    "$BASE_URL$request_path")"

  validate_request_id "$headers_file" "$request_path"

  if [[ "$status_code" -ge 200 && "$status_code" -lt 300 ]]; then
    validate_success_envelope "$body_file"
    echo "$request_path status=available http_status=$status_code"
    return 0
  fi

  if [[ "$status_code" == "404" || "$status_code" == "422" ]]; then
    validate_error_envelope "$body_file"
    echo "$request_path status=honestly_unavailable http_status=$status_code"
    return 0
  fi

  echo "ERROR: $request_path returned unexpected HTTP $status_code"
  cat "$body_file"
  return 1
}

cd "$API_DIR"

echo "[1/6] Checking OurAirports source freshness and conditional import"
go run ./cmd/import-airports | tee "$TMP_DIR/import-airports.log"

if ! grep -Eq 'retrieval_status=(downloaded|not_modified)' "$TMP_DIR/import-airports.log"; then
  echo "ERROR: OurAirports import did not report downloaded or not_modified status"
  exit 1
fi

echo "[2/6] Building bounded runtime binaries"
go build -o "$TMP_DIR/ingest" ./cmd/ingest
go build -o "$TMP_DIR/server" ./cmd/server

echo "[3/6] Running one bounded live ingestion cycle"
TRAFFIC_INGESTION_RADIUS="$SMOKE_RADIUS" "$TMP_DIR/ingest" >"$TMP_DIR/ingest.log" 2>&1 &
INGEST_PID=$!

CYCLE_COMPLETED=0

for attempt in {1..240}; do
  if grep -q 'ingestion_run_id=' "$TMP_DIR/ingest.log" 2>/dev/null; then
    CYCLE_COMPLETED=1
    break
  fi

  if ! kill -0 "$INGEST_PID" >/dev/null 2>&1; then
    echo "ERROR: ingest process exited before completing its first cycle"
    exit 1
  fi

  sleep 0.5
done

if (( CYCLE_COMPLETED != 1 )); then
  echo "ERROR: first ingestion cycle did not complete within 120 seconds"
  exit 1
fi

kill -TERM "$INGEST_PID" >/dev/null 2>&1 || true
wait "$INGEST_PID" >/dev/null 2>&1 || true
INGEST_PID=""

cat "$TMP_DIR/ingest.log"

LOADED_COUNT="$(sed -nE 's/.*loaded=([0-9]+).*/\1/p' "$TMP_DIR/ingest.log" | tail -n 1)"
STORED_COUNT="$(sed -nE 's/.*stored=([0-9]+).*/\1/p' "$TMP_DIR/ingest.log" | tail -n 1)"
TRAJECTORY_COUNT="$(sed -nE 's/.*trajectories=([0-9]+).*/\1/p' "$TMP_DIR/ingest.log" | tail -n 1)"

if [[ -z "$LOADED_COUNT" || -z "$STORED_COUNT" || -z "$TRAJECTORY_COUNT" ]]; then
  echo "ERROR: bounded ingestion output did not contain loaded, stored and trajectories counters"
  exit 1
fi

if (( LOADED_COUNT <= 0 )); then
  echo "ERROR: provider returned zero aircraft within radius $SMOKE_RADIUS. Rerun later or set RUNTIME_SMOKE_RADIUS to a supported larger radius."
  exit 1
fi

if (( STORED_COUNT <= 0 )); then
  echo "ERROR: live ingestion did not persist any flight states"
  exit 1
fi

echo "[4/6] Starting the real API server on port $SMOKE_PORT"
API_PORT="$SMOKE_PORT" "$TMP_DIR/server" >"$TMP_DIR/server.log" 2>&1 &
SERVER_PID=$!

SERVER_READY=0

for attempt in {1..60}; do
  if curl --silent --fail "$BASE_URL/health" >/dev/null 2>&1; then
    SERVER_READY=1
    break
  fi

  if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
    echo "ERROR: API server exited before becoming ready"
    exit 1
  fi

  sleep 0.5
done

if (( SERVER_READY != 1 )); then
  echo "ERROR: API server did not become ready within 30 seconds"
  exit 1
fi

echo "[5/6] Verifying system and collection HTTP endpoints"
request_json "/health" "health" >/dev/null
request_json "/version" "version" >/dev/null
request_json "/regions" "regions" >/dev/null
TRAFFIC_FILE="$(request_json "/traffic/current" "traffic-current")"
request_json "/aircraft" "aircraft" >/dev/null
request_json "/airports" "airports" >/dev/null

ICAO24="$(node - "$TRAFFIC_FILE" <<'NODE'
const fs = require("fs");
const payload = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const items = Array.isArray(payload.data) ? payload.data : [];
const first = items.find(
  (item) =>
    item &&
    typeof item.icao24 === "string" &&
    item.icao24.trim() !== ""
);
process.stdout.write(first ? first.icao24.trim() : "");
NODE
)"

if [[ -z "$ICAO24" ]]; then
  echo "ERROR: current traffic endpoint returned no aircraft identifier after live ingestion"
  exit 1
fi

echo "[6/6] Verifying aircraft-specific runtime endpoints for $ICAO24"
request_json "/aircraft/$ICAO24/latest-state" "aircraft-latest-state" >/dev/null
request_json_allow_unavailable "/aircraft/$ICAO24/trajectory" "aircraft-trajectory"
request_json_allow_unavailable "/aircraft/$ICAO24/route-context" "aircraft-route-context"

echo "RUNTIME_DATA_PATH_VERIFICATION=passed"
echo "ourairports_status=verified"
echo "provider_live_request=verified"
echo "loaded_states=$LOADED_COUNT"
echo "stored_states=$STORED_COUNT"
echo "built_trajectories=$TRAJECTORY_COUNT"
echo "verified_icao24=$ICAO24"
echo "http_server=verified"
echo "request_id_headers=verified"
