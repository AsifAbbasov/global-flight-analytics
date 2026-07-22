#!/bin/bash
set -euo pipefail

REPOSITORY_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
API_ROOT="$REPOSITORY_ROOT/apps/api"

export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.26.5+auto}"

cd "$API_ROOT"
go test -count=1 ./tools/codereviewaudit
go run ./tools/codereviewaudit -strict

echo 'CODE_REVIEW_POLICY=PASS'
