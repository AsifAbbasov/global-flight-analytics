#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf '%s\n' "GRAFANA_OBSERVABILITY_PROVISION=FAIL reason=$1" >&2
  exit 1
}

command -v curl >/dev/null 2>&1 || fail 'curl is required'
command -v jq >/dev/null 2>&1 || fail 'jq is required'
command -v node >/dev/null 2>&1 || fail 'node is required'

: "${GRAFANA_INSTANCE_URL:?Set GRAFANA_INSTANCE_URL to the Grafana Cloud stack URL}"
: "${GRAFANA_STACK_ID:?Set GRAFANA_STACK_ID to the numeric Grafana Cloud stack ID}"
: "${GRAFANA_SERVICE_ACCOUNT_TOKEN:?Set GRAFANA_SERVICE_ACCOUNT_TOKEN}"
: "${GRAFANA_PROMETHEUS_DATASOURCE_UID:?Set GRAFANA_PROMETHEUS_DATASOURCE_UID}"

GRAFANA_INSTANCE_URL="${GRAFANA_INSTANCE_URL%/}"
case "$GRAFANA_INSTANCE_URL" in
  https://*) ;;
  *) fail 'GRAFANA_INSTANCE_URL must use HTTPS' ;;
esac
case "$GRAFANA_STACK_ID" in
  ''|*[!0-9]*) fail 'GRAFANA_STACK_ID must be numeric' ;;
esac
GRAFANA_NAMESPACE="stacks-$GRAFANA_STACK_ID"
GRAFANA_API_MAX_ATTEMPTS=5
GRAFANA_API_BASE_DELAY_SECONDS=2
GRAFANA_API_MAX_RETRY_AFTER_SECONDS=30

TEMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/gfa-grafana-observability.XXXXXX")"
cleanup() { rm -rf "$TEMP_ROOT"; }
trap cleanup EXIT

node scripts/render-grafana-observability.mjs "$TEMP_ROOT"

auth_header="Authorization: Bearer $GRAFANA_SERVICE_ACCOUNT_TOKEN"

is_retryable_status() {
  case "$1" in
    429|502|503|504) return 0 ;;
    *) return 1 ;;
  esac
}

retry_delay_seconds() {
  headers_file="$1"
  attempt="$2"
  retry_after="$(awk 'BEGIN { IGNORECASE=1 } /^Retry-After:/ { gsub("\r", "", $2); print $2; exit }' "$headers_file" 2>/dev/null || true)"
  if [ -n "$retry_after" ] && [ "$retry_after" -eq "$retry_after" ] 2>/dev/null && [ "$retry_after" -le "$GRAFANA_API_MAX_RETRY_AFTER_SECONDS" ]; then
    printf '%s\n' "$retry_after"
    return
  fi

  delay=$((GRAFANA_API_BASE_DELAY_SECONDS * (2 ** (attempt - 1))))
  if [ "$delay" -gt "$GRAFANA_API_MAX_RETRY_AFTER_SECONDS" ]; then
    delay="$GRAFANA_API_MAX_RETRY_AFTER_SECONDS"
  fi
  printf '%s\n' "$delay"
}

request_status() {
  method="$1"
  endpoint="$2"
  input_file="${3:-}"
  output_file="${4:-$TEMP_ROOT/response.json}"
  headers_file="$TEMP_ROOT/headers.txt"
  attempt=1

  while :; do
    args=(--silent --show-error --request "$method" --header "$auth_header" --header 'Accept: application/json' --dump-header "$headers_file" --output "$output_file" --write-out '%{http_code}')
    if [ -n "$input_file" ]; then
      args+=(--header 'Content-Type: application/json' --data-binary "@$input_file")
    fi

    set +e
    status="$(curl "${args[@]}" "$GRAFANA_INSTANCE_URL$endpoint")"
    curl_status=$?
    set -e

    if [ "$curl_status" -ne 0 ]; then
      fail "$method $endpoint transport failed with curl exit $curl_status"
    fi

    if is_retryable_status "$status" && [ "$attempt" -lt "$GRAFANA_API_MAX_ATTEMPTS" ]; then
      delay="$(retry_delay_seconds "$headers_file" "$attempt")"
      printf '%s\n' "GRAFANA_API_RETRY method=$method endpoint=$endpoint status=$status attempt=$attempt delay_seconds=$delay" >&2
      sleep "$delay"
      attempt=$((attempt + 1))
      continue
    fi

    printf '%s\n' "$status"
    return
  done
}

api() {
  method="$1"
  endpoint="$2"
  input_file="${3:-}"
  output_file="${4:-$TEMP_ROOT/response.json}"
  status="$(request_status "$method" "$endpoint" "$input_file" "$output_file")"
  case "$status" in
    2??) ;;
    *)
      cat "$output_file" >&2 || true
      fail "$method $endpoint returned HTTP $status"
      ;;
  esac
}

upsert_resource() {
  kind="$1"; collection="$2"; name="$3"; source="$4"
  current="$TEMP_ROOT/${kind}-current.json"
  status="$(request_status GET "$collection/$name" '' "$current")"
  if [ "$status" = '404' ]; then
    api POST "$collection" "$source" "$TEMP_ROOT/${kind}-created.json"
  elif [ "$status" = '200' ]; then
    resource_version="$(jq -er '.metadata.resourceVersion' "$current")"
    update="$TEMP_ROOT/${kind}-update.json"
    jq --arg version "$resource_version" '.metadata.resourceVersion = $version' "$source" > "$update"
    api PUT "$collection/$name" "$update" "$TEMP_ROOT/${kind}-updated.json"
  else
    cat "$current" >&2 || true
    fail "$kind lookup returned HTTP $status"
  fi
}

folder_collection="/apis/folder.grafana.app/v1/namespaces/$GRAFANA_NAMESPACE/folders"
dashboard_collection="/apis/dashboard.grafana.app/v1/namespaces/$GRAFANA_NAMESPACE/dashboards"

upsert_resource folder "$folder_collection" 'gfa-observability' "$TEMP_ROOT/folder.json"
upsert_resource dashboard "$dashboard_collection" 'gfa-production-slo' "$TEMP_ROOT/dashboard.json"

api PUT '/api/v1/provisioning/folder/gfa-observability/rule-groups/global-flight-analytics-production-slo' "$TEMP_ROOT/alert-rules.json" "$TEMP_ROOT/alerts-response.json"

api GET "$dashboard_collection/gfa-production-slo" '' "$TEMP_ROOT/dashboard-evidence.json"
jq -e '.metadata.name == "gfa-production-slo" and .spec.title == "Global Flight Analytics — Production SLO" and (.spec.panels | length >= 12)' "$TEMP_ROOT/dashboard-evidence.json" >/dev/null

api GET '/api/v1/provisioning/folder/gfa-observability/rule-groups/global-flight-analytics-production-slo' '' "$TEMP_ROOT/alerts-evidence.json"
jq -e '.title == "global-flight-analytics-production-slo" and (.rules | length == 9)' "$TEMP_ROOT/alerts-evidence.json" >/dev/null

api GET '/api/v1/provisioning/policies' '' "$TEMP_ROOT/policy-evidence.json"
receiver="$(jq -er '.receiver | select(type == "string" and length > 0)' "$TEMP_ROOT/policy-evidence.json")"
if [ -n "${GRAFANA_EXPECTED_RECEIVER:-}" ] && [ "$receiver" != "$GRAFANA_EXPECTED_RECEIVER" ]; then
  fail "notification policy receiver '$receiver' does not match GRAFANA_EXPECTED_RECEIVER"
fi

printf '%s\n' "GRAFANA_NAMESPACE=$GRAFANA_NAMESPACE"
printf '%s\n' 'GRAFANA_FOLDER=PASS'
printf '%s\n' 'GRAFANA_SLO_DASHBOARD=PASS'
printf '%s\n' 'GRAFANA_ALERT_RULES=PASS'
printf '%s\n' "GRAFANA_NOTIFICATION_POLICY_RECEIVER=$receiver"
printf '%s\n' 'GRAFANA_NOTIFICATION_POLICY=PASS'
printf '%s\n' 'GRAFANA_OBSERVABILITY_PROVISION=PASS'
