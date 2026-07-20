# Document 67 — Stage 14.26 Airport Elevation Semantics

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: preserve the distinction between unknown airport elevation and observed sea-level elevation

## 1. Correctness problem

The airport repository previously selected `COALESCE(a.elevation_ft, 0)`.
A missing OurAirports elevation therefore became a factual zero-foot elevation.
The domain and HTTP contracts had no availability or status field, so the
incorrect value propagated into Airport profiles, Airport Intelligence,
preliminary route context, production Route Intelligence, and the frontend.

## 2. Canonical semantics

Airport elevation now has three public states:

```text
observed — a finite source value exists, including exactly zero
unknown  — the source column is NULL and no elevation is claimed
invalid  — an explicitly supplied value is non-finite
```

A non-zero legacy in-memory value remains observed for compatibility. A zero
value is observed only when availability is explicit.

## 3. PostgreSQL boundary

Both airport repository queries select nullable `elevation_ft` directly into
`pgtype.Int4`. NULL maps to unknown. Any valid integer, including zero and
negative values, is converted from international feet to metres and marked
observed. No schema migration is required because the column is already
nullable.

## 4. Propagation boundary

The semantics are preserved through:

```text
Airport domain
Airport profile HTTP API
Airport Intelligence passport
preliminary aircraft route context
production Route Intelligence contract and HTTP API
TypeScript route contracts
aircraft route-context panel
```

Public JSON uses nullable `elevation_m` plus `elevation_status`. Unknown or
invalid evidence is never serialized as a factual zero.

## 5. Regression protection

The permanent tests protect:

```text
NULL versus observed zero PostgreSQL mapping
negative and legacy non-zero elevation compatibility
non-finite value rejection
absence of COALESCE(a.elevation_ft, 0)
route catalog availability fingerprint input
nullable HTTP conversion
frontend nullable elevation formatting
```

## 6. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/domain/airport internal/repository/postgres internal/airportintelligence/passport internal/routeintelligence/airportresolver internal/routeintelligence/routecontract internal/http/dto internal/http/handlers
go test -count=1 ./internal/domain/airport ./internal/repository/postgres ./internal/airportintelligence/passport ./internal/routeintelligence/airportresolver ./internal/routeintelligence/routecontract ./internal/http/dto ./internal/http/handlers
go test -count=1 ./...
go vet ./...
```

From `apps/web`:

```bash
pnpm typecheck
pnpm lint
pnpm build
```

From the repository root:

```bash
git diff --check
git status --short
```

## 7. Completion boundary

This increment closes Airport elevation semantics. It does not close timestamp
and Unix-nanosecond consistency or large PostgreSQL repository decomposition.
