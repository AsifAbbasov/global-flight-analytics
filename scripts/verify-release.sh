#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPOSITORY_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"
cd "$REPOSITORY_ROOT"

printf '%s\n' "RELEASE_NODE_VERSION=$(node --version)"
printf '%s\n' "RELEASE_PNPM_VERSION=$(pnpm --version)"
printf '%s\n' "RELEASE_GO_VERSION=$(go version)"
printf '%s\n' "RELEASE_COMMIT=$(git rev-parse HEAD)"

bash scripts/verify-recruiter-quickstart.sh
pnpm run test:release-contract
pnpm run verify:release-contract
pnpm run test:backend-operations-contract
pnpm run verify:backend-operations-contract
pnpm install --frozen-lockfile
pnpm run test:web-dependency-policy
pnpm run verify:web-dependencies
pnpm --dir apps/web test
pnpm --dir apps/web lint
pnpm --dir apps/web typecheck
NEXT_PUBLIC_API_BASE_URL=http://127.0.0.1:8080 pnpm --dir apps/web build

unformatted_files="$(cd apps/api && gofmt -l .)"
if [ -n "$unformatted_files" ]; then
  printf '%s\n' 'The following Go files are not formatted:' >&2
  printf '%s\n' "$unformatted_files" >&2
  exit 1
fi

(
  cd apps/api
  go test -count=1 ./...
  go vet ./...
  go run ./tools/projectaudit -mode all -strict
  go run ./tools/codereviewaudit -strict
  go run ./tools/backendcontextownershipaudit -strict
  go run ./tools/backendtimeoutconsistencyaudit -strict
  go run ./tools/backendobservabilityaudit -strict
  go run ./tools/stage14finalaudit -strict
  go run ./tools/analyticalcorefinalaudit -strict
)

docker compose config >/dev/null
git diff --check

printf '%s\n' 'RELEASE_RECRUITER_QUICKSTART=PASS'
printf '%s\n' 'RELEASE_PORTFOLIO_CONTRACT_TESTS=PASS'
printf '%s\n' 'RELEASE_BACKEND_OPERATIONS=PASS'
printf '%s\n' 'RELEASE_FRONTEND=PASS'
printf '%s\n' 'RELEASE_BACKEND=PASS'
printf '%s\n' 'RELEASE_COMPOSE=PASS'
printf '%s\n' 'RELEASE_VERIFICATION=PASS'
