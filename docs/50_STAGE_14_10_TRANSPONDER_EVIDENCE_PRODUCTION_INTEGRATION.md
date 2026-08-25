# Document 50 — Stage 14.10 Transponder Evidence Production Integration

Status: Remediation History v1.1
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

## 11. Canonical finding record — GFA-REL-045

### Finding / symptom

`internal/analytics/transponderalert` contained validated logic for interpreting supported special transponder codes, but it remained `planned_production_integration` with no server/operational entry point.

### Root cause

The evidence-analysis package was implemented before a production composition and HTTP claim boundary was defined. Internal analytical capability therefore existed ahead of supported product reachability.

### Failure scenario

The repository appears to contain a transponder alert/evidence capability because the package and tests exist, while no production route can invoke it. Alternatively, a rushed integration could expose the package as an "emergency alert" and overstate what one external observation proves.

### Impact

The first risk is release/product truth: implemented code is mistaken for shipped capability. The second is claim safety: special transponder codes can be presented as confirmed incidents instead of bounded observed evidence.

### Severity rationale

**P2 retrospective.** The primary defect is production reachability/release correctness with a safety-sensitive claim boundary. No evidence shows that GFA issued a false operational emergency alert, so it is not classified as P1.

### Existing guarantees violated

- planned production packages must be integrated or removed before release;
- production evidence must remain bounded to what the persisted observation supports;
- GFA is not an operational ATC/safety alerting system;
- analytics/HTTP layers should not depend directly on PostgreSQL implementation details.

### Considered solutions

1. delete the package as unused;
2. expose it as an active alert/notification subsystem;
3. integrate it as a read-only evidence endpoint with explicit non-operational claim limits.

### Chosen remediation

The package is wired through the server composition root to `GET /api/v1/aircraft/:icao24/transponder-evidence/latest`, backed by the existing Flight State service. Responses explicitly mark `evidence_only=true`, `confirmed_emergency=false`, `operational_alert=false`, include freshness/provenance/limitations, and avoid arbitrary incident probabilities.

### Why selected

The underlying analysis had real product value as research evidence, so deletion was unnecessary. Read-only bounded exposure fits the platform mission without introducing monitoring, escalation, or safety-critical behavior.

### Rejected alternatives

Deletion was rejected because the package was not obsolete. Automatic alerts/incident declarations were rejected because one persisted external transponder observation cannot establish cause or operational truth. A new service/microservice was unnecessary for one read-only modular-monolith boundary.

### Trade-offs

The API supports an additional evidence surface and must preserve careful language/freshness semantics. It intentionally provides less dramatic behavior than an "alert" product: no push, escalation, or operational action.

### Regression tests / protection

Architecture reachability requires the package from `cmd/server`; route topology protects read-only GET; service/DTO/handler tests preserve evidence/freshness/claim semantics; dependency-boundary tests keep PostgreSQL out of analytics/HTTP layers.

### Adversarial review findings

The remediation explicitly constrained the maximum claim to `observed_transponder_code_only`, degraded stale evidence rather than deleting history, and refused to infer emergency cause. These safeguards address the high-risk semantic shortcut of equating a code with a confirmed incident.

### Remediation iterations

Stage 14.2 classified the package as planned integration. Stage 14.10 resolved that disposition with a real read-only runtime path and explicit safety semantics rather than synthetic reachability.

### Residual risks / limitations

Provider observations may be stale, incomplete, incorrect, or spoofed. GFA does not confirm emergencies, identify causes, or provide operational instructions. Five minutes is a product freshness boundary, not a safety standard.

### Operational / deployment consequences

One read-only endpoint and server composition path are added; no migration, background monitoring, paid feed, or notification infrastructure is introduced.

### Exact evidence

Implementation commit: `19b187d848e993d13a72b0c3c4f212db8c7577fb` (`feat: integrate transponder evidence api`). Historical PR/reviewer metadata is not invented where unavailable.

### Final canonical status

**CLOSED.** The package is production-reachable as bounded read-only evidence and no longer carries the planned-integration disposition.

### Prevention / future guard

Safety-adjacent aviation evidence must state the maximum supported claim and explicit non-claims. Production reachability must be real, but integration must not silently escalate research evidence into operational alerting or causal conclusions.
