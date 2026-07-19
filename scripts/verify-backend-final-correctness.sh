#!/bin/bash
set -euo pipefail

REPOSITORY_ROOT="$(
  cd "$(dirname "$0")/.." &&
    pwd
)"
API_ROOT="$REPOSITORY_ROOT/apps/api"

cd "$API_ROOT"

echo '$ go test ./tools/backendfinalaudit -count=1'
go test \
  ./tools/backendfinalaudit \
  -count=1

echo '$ go run ./tools/backendfinalaudit -strict'
go run \
  ./tools/backendfinalaudit \
  -strict

echo '$ go run ./tools/projectaudit -mode all -strict'
go run \
  ./tools/projectaudit \
  -mode all \
  -strict

echo '$ go test final correctness boundary packages'
go test \
  ./internal/projectionintelligence/projectionread \
  ./internal/historicalintelligence/historicalaggregate \
  ./internal/historicalintelligence/historicalaggregatecontract \
  ./internal/http/historicalcursor \
  ./internal/http/dto \
  ./internal/http/handlers \
  ./internal/server \
  ./internal/services/weather \
  ./internal/services/traffic/validator \
  ./internal/integrations/openmeteo \
  ./internal/integrations/opensky \
  ./internal/integrations/airplaneslive \
  ./internal/airspaceintelligence/airspaceproduction \
  ./internal/orchestration/weatherprovider \
  ./internal/orchestration/providerresponse \
  ./internal/orchestration/providerbudget \
  ./internal/orchestration/ingestionorchestrator \
  ./internal/repository/postgres \
  ./internal/security/internalapikey \
  ./internal/database/... \
  ./cmd/verify-postgres-projection-intelligence-http-api \
  ./cmd/verify-postgres-historical-http-api \
  ./cmd/verify-postgres-historical-aggregate-store \
  ./cmd/verify-postgres-historical-materialization-replay \
  ./cmd/verify-postgres-weather-context-http-api \
  ./cmd/verify-postgres-stability-intelligence-http-api \
  ./cmd/verify-postgres-airspace-region-analytics-http-api \
  ./cmd/verify-source-constraints \
  ./cmd/verify-opensky-rest-compatibility \
  -count=1

echo '$ go test -race final corrected boundaries'
go test \
  -race \
  ./internal/projectionintelligence/projectionread \
  ./internal/historicalintelligence/historicalaggregate \
  ./internal/historicalintelligence/historicalaggregatecontract \
  ./internal/http/historicalcursor \
  ./internal/server \
  ./internal/services/weather \
  ./internal/services/traffic/validator \
  ./internal/integrations/openmeteo \
  ./internal/integrations/opensky \
  ./internal/integrations/airplaneslive \
  ./internal/airspaceintelligence/airspaceproduction \
  ./internal/orchestration/weatherprovider \
  ./internal/orchestration/providerresponse \
  ./internal/orchestration/providerbudget \
  ./internal/orchestration/ingestionorchestrator \
  ./internal/repository/postgres \
  ./internal/security/internalapikey \
  -count=1

echo '$ go list ./...'
go list ./... >/dev/null

echo '$ go build ./cmd/...'
go build ./cmd/...

echo '$ go vet ./...'
go vet ./...

echo '$ go test ./...'
go test ./...

echo 'BACKEND_FINAL_CORRECTNESS_AUDIT=PASS'
