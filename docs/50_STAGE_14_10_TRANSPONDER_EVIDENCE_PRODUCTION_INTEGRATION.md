# Document 50 — Stage 14.10 Transponder Evidence Production Integration

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: read-only production exposure of observed special transponder code evidence

## 1. Purpose

The existing `internal/analytics/transponderalert` package already converted
persisted flight-state observations into bounded evidence for special
transponder codes.

Before this stage, that package had no production entry point and was formally
classified as planned integration.

This stage connects it to the database-backed server through a read-only API.

## 2. Production Endpoint

```text
GET /api/v1/aircraft/:icao24/transponder-evidence/latest
```

The endpoint reads the latest persisted flight state for the aircraft.

It returns evidence only when the latest state contains one of the supported
special transponder codes.

The endpoint does not trigger ingestion, mutation, notification, dispatch, or
operational action.

## 3. Safety Semantics

Every successful response states:

```text
evidence_only = true
confirmed_emergency = false
operational_alert = false
maximum_claim_strength = observed_transponder_code_only
```

The response describes an observed transmitted code.

It does not claim that the platform confirmed:

```text
an emergency
unlawful interference
radio communication failure
an incident cause
an operational directive
```

## 4. Evidence Model

The response includes:

```text
aircraft ICAO24 and optional callsign
observed transponder code
bounded classification and label
evidence strength
observation count
Special Purpose Indicator observation status
first and last observation times
as-of time
data age and freshness threshold
limited or degraded assessment confidence
source names and deterministic fingerprint
explicit limitations
```

No arbitrary numerical probability is produced.

Confidence is qualitative because one latest persisted external observation is
not sufficient to justify a calibrated incident probability.

## 5. Freshness

The default freshness threshold is five minutes.

This threshold is a product data-timeliness boundary, not an aviation safety
standard and not a scientific constant.

Evidence older than the threshold is still returned because it remains valid
historical evidence of the persisted observation, but it is marked:

```text
freshness = stale
confidence = degraded
```

A stale response includes an additional limitation explaining that the code
may no longer represent the aircraft's current transmitted value.

## 6. Dependency Boundary

The production service depends on a narrow interface:

```text
GetLatestByICAO24(context, ICAO24) FlightState
```

The analytics package and HTTP layer do not import PostgreSQL or `pgx`.

Concrete PostgreSQL repository selection remains inside the server composition
root.

## 7. Runtime Integration

The server composition root now owns a dedicated bounded-context pair:

```text
transponder_evidence_database_composition.go
transponder_evidence_database_routes.go
```

Composition creates:

```text
PostgreSQL Flight State repository
Flight State domain service
Transponder Evidence service
HTTP handler
```

The route file registers only the read-only HTTP path.

## 8. Reachability Governance

`internal/analytics/transponderalert` is now reachable from `cmd/server`.

Its former `planned_production_integration` policy is removed.

Architecture tests require the production package to remain outside the
non-runtime allowlist.

## 9. Intentionally Excluded Behavior

This stage does not add:

```text
push notifications
email or messaging alerts
automatic emergency declarations
incident escalation
continuous background monitoring
new database tables
new migrations
external paid feeds
first-party sensor claims
```

It also does not infer the cause behind an observed code.

## 10. Acceptance

The increment is accepted only after:

```text
focused analytics service tests
DTO and handler tests
read-only route topology test
HTTP and analytics dependency-boundary tests
race detector
strict project reachability audit
complete Go build
complete Go test suite
go vet
frontend dependency security verification
frontend production dependency audit
ESLint
TypeScript validation
Next.js production build
backend Docker image build
git diff check
```
