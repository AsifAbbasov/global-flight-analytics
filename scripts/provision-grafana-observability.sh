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

TEMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/gfa-grafana-observability.XXXXXX")"
cleanup() { rm -rf "$TEMP_ROOT"; }
trap cleanup EXIT

node scripts/render-grafana-observability.mjs "$TEMP_ROOT"

auth_header="Authorization: Bearer $GRAFANA_SERVICE_ACCOUNT_TOKEN"
api() {
  method="$1"; endpoint="$2"; input_file="${3:-}"; output_file="${4:-$TEMP_ROOT/response.json}"
  args=(--fail --silent --show-error --request "$method" --header "$auth_header" --header 'Accept: application/json')
  if [ -n "$input_file" ]; then
    args+=(--header 'Content-Type: application/json' --data-binary "@$input_file")
  fi
  curl "${args[@]}" --output "$output_file" "$GRAFANA_INSTANCE_URL$endpoint"
}

upsert_resource() {
  kind="$1"; collection="$2"; name="$3"; source="$4"
  current="$TEMP_ROOT/${kind}-current.json"
  status="$(curl --silent --show-error --output "$current" --write-out '%{http_code}' --header "$auth_header" --header 'Accept: application/json' "$GRAFANA_INSTANCE_URL$collection/$name")"
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
