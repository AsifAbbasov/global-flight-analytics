# Document 66 — Stage 14.25 Traffic Altitude Status Semantics

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: preserve typed altitude meaning from PostgreSQL current-traffic reads through the public HTTP contract and frontend presentation

## 1. Correctness problem

The former current-traffic query reduced geometric and barometric altitude into one
number through this shape:

```text
geometric altitude unless it equals zero
otherwise barometric altitude
otherwise zero
```

That logic introduced two correctness failures:

1. a legitimate observed geometric altitude of zero was treated as missing;
2. unknown, unavailable, or invalid altitude evidence was published as the
   numeric value zero.

The HTTP and frontend contracts could not distinguish an observed zero, an
aircraft on the ground, and a missing altitude.

## 2. Typed current-traffic altitude contract

Each current-traffic item now carries:

```text
altitude_m       nullable numeric value
altitude_status  observed | ground | unknown | unavailable | invalid
altitude_source  geometric | barometric | ground | none
```

A missing altitude is represented by `null`, never by a fabricated zero.

## 3. Selection policy

The current-traffic altitude resolver uses this deterministic order:

```text
on_ground = true
    → altitude 0, status ground, source ground

usable observed geometric altitude
    → preserve the value, including observed zero

usable observed barometric altitude
    → use barometric fallback

no usable observed altitude
    → null value and the strongest available status
```

For absent evidence, invalid dominates unknown, and unknown dominates
unavailable. A ground altitude status without independent `on_ground` evidence is
invalid.

## 4. PostgreSQL read behavior

Both unbounded and bounded current-traffic queries load the geometric and
barometric values and statuses independently. They no longer use `NULLIF`,
`COALESCE`, or zero as altitude-selection logic.

The repository scans nullable values and delegates selection to the domain
resolver. This keeps SQL responsible for data retrieval and the domain layer
responsible for meaning.

## 5. Public HTTP behavior

The endpoint remains:

```text
GET /api/v1/traffic/current
GET /api/v1/traffic/current?region=<code>
```

Existing fields remain available. The altitude value is now nullable and two
explicit fields are added:

```json
{
  "altitude_m": null,
  "altitude_status": "unknown",
  "altitude_source": "none"
}
```

An observed zero remains numeric:

```json
{
  "altitude_m": 0,
  "altitude_status": "observed",
  "altitude_source": "geometric"
}
```

## 6. Frontend behavior

The current traffic map and aircraft detail panel render altitude meaning
directly:

```text
Ground (0 m)
0 m (geometric)
2400 m (barometric)
Unknown
Unavailable
Invalid altitude evidence
```

The frontend no longer renders missing altitude as `0 m`.

## 7. Regression protection

Permanent tests protect:

```text
observed geometric zero preservation
barometric fallback
ground representation
unknown and unavailable nullability
invalid evidence handling
both current-traffic query variants
HTTP DTO propagation
removal of zero-sentinel SQL
frontend type checking, linting, and production build
```

The PostgreSQL integration test is enabled when `TEST_DATABASE_URL` is set and
uses an isolated temporary schema.

## 8. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/domain/traffic/altitude.go internal/domain/traffic/altitude_test.go internal/domain/traffic/model.go internal/repository/postgres/traffic_repository.go internal/repository/postgres/traffic_altitude_semantics_ownership_test.go internal/repository/postgres/traffic_altitude_semantics_integration_test.go internal/http/dto/traffic.go internal/http/handlers/traffic.go internal/http/handlers/traffic_altitude_semantics_test.go
go test -count=1 ./internal/domain/traffic ./internal/repository/postgres ./internal/http/handlers
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
pnpm --dir apps/web typecheck
pnpm --dir apps/web lint
pnpm --dir apps/web build
git diff --check
git status --short
```

## 9. Completion boundary

This increment closes Traffic altitude-status semantics.

It does not close:

```text
Airport elevation semantics
timestamp and Unix-nanosecond consistency
large PostgreSQL repository decomposition
```
