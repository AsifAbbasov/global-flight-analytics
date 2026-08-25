# Document 43 — Stage 14.3 Airport Intelligence Production Integration

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: PostgreSQL composition, read-only HTTP integration, and runtime reachability of the complete Airport Intelligence domain

## 1. Purpose

Stage 14.3 converts the previously isolated Airport Intelligence domain foundation into a production read path.

The domain packages are retained because they implement distinct validated responsibilities:

```text
passport
statistics
ranking
overview
history
trends
```

They are now composed by `airportintelligence/airportproduction`.

## 2. Production Data Boundary

The production reader uses existing project tables:

```text
airports
airport_profiles
airport_statistics
route_statistics
route_predictions
```

No official airport operations feed is claimed.

Daily airport observations are built only from completed Coordinated Universal Time dates present in `airport_statistics`. Active-route context is derived from `route_statistics`. Active-aircraft context is derived from available `route_predictions`.

## 3. Read-Only HTTP Routes

```text
GET /api/v1/airports/intelligence/ranking
GET /api/v1/airports/:icao/intelligence/overview
GET /api/v1/airports/:icao/intelligence/history
GET /api/v1/airports/:icao/intelligence/trends
```

Supported query parameters are `days`, `as_of_time`, and ranking-only `limit`.

## 4. Window Semantics

The current partial Coordinated Universal Time day is excluded. This prevents partial-day values from being ranked against complete daily windows.

## 5. Statistics Semantics

For a multi-day window:

```text
arrivals = sum of daily arrivals
departures = sum of daily departures
active aircraft = peak available daily active-aircraft count
active routes = peak available daily active-route count
observed samples = number of dates with airport statistics
expected samples = requested completed-day count
```

## 6. Ranking Semantics

Airport Activity Score remains relative to airports with observations in the same requested window. It is not an absolute worldwide airport classification.

## 7. Security Boundary

All Stage 14.3 routes are read-only and expose open research data. User authentication is not required for these routes. Administrative materialization and mutation routes must be protected before deployment.

## 8. Runtime Completion Gate

Stage 14.3 is complete only when strict architecture audit reports every Airport Intelligence package as reachable from `cmd/server`.

## 9. Known Limitations

```text
No official airport operations feed
No guaranteed complete route coverage
No guaranteed complete active-aircraft coverage
No current partial-day ranking
No causal explanations
No forecasting
No universal calibration of project-derived ranking weights
```

## 10. Canonical finding record — GFA-REL-036

### Finding / symptom

The six Airport Intelligence domain packages were implemented and tested but remained classified as `planned_production_integration`: no production server read path made the capability reachable to the product.

### Root cause

Domain foundations were developed before the final production composition/HTTP boundary. Testability and domain completeness were therefore ahead of runtime integration.

### Failure scenario

The repository contains sophisticated passport/statistics/ranking/history/trends code and tests, yet `cmd/server` cannot reach it. Project documentation or portfolio review can interpret the code as a shipped capability even though no production endpoint exists.

### Impact

The gap creates false release/product capability evidence and leaves maintained code without a supported runtime consumer. It also prevents real HTTP/PostgreSQL integration from exercising the domain against production data contracts.

### Severity rationale

**P2 retrospective.** This is release/product correctness and production reachability, not evidence of corrupted persisted data.

### Existing guarantees violated

- production capability claims require a real runtime root;
- planned integration must be resolved before final release;
- read-only analytical domains require explicit production composition and bounded data/claim semantics.

### Considered solutions

1. delete Airport Intelligence as unused code;
2. keep it as permanently offline/test-only code;
3. integrate the complete validated domain through a read-only production composition and HTTP surface.

### Chosen remediation

The project retained all six packages and composed them through `airportproduction`, backed by existing PostgreSQL tables, with four read-only endpoints and strict runtime-reachability verification.

### Why this solution was selected

The packages represented distinct validated domain responsibilities and had a plausible product read path. Deletion would discard real functionality; keeping them unintegrated would preserve the release-truth defect.

### Rejected alternatives

Deletion was rejected because the domain was not obsolete. Artificial imports solely to satisfy reachability were rejected because they would not create meaningful product behavior. A new service/microservice was unnecessary; the modular monolith composition root already fit the requirement.

### Trade-offs

Production integration increases the supported API surface and requires PostgreSQL/HTTP contract maintenance. In exchange, code ownership becomes truthful: the domain is either reachable and supported or explicitly removed.

### Regression tests / protection

Strict architecture audit requires all Airport Intelligence packages to remain reachable from `cmd/server`. HTTP/service/repository tests protect the production read path and query semantics.

### Adversarial review findings

The integration preserves claim boundaries: existing project tables do not become an "official airport operations feed," incomplete current UTC days are excluded, and ranking remains relative to observed airports rather than universal. This prevents production reachability from inflating scientific/data-source claims.

### Remediation iterations

Stage 14.2 first classified Airport Intelligence as planned production integration. Stage 14.3 resolved that disposition through real server composition rather than a synthetic reachability import.

### Residual risks / limitations

Coverage depends on project-derived airport/route/route-prediction data and can remain incomplete. Runtime reachability does not imply worldwide coverage or calibrated absolute airport ranking.

### Operational / deployment consequences

The backend exposes new read-only routes and must have access to the existing PostgreSQL tables. No schema migration or authentication change is introduced.

### Exact evidence

Implementation commit: `bb9f3510fd9fead1a80edb688c1ab125b8fbdb1b` (`feat: integrate airport intelligence production api`). Historical PR/reviewer evidence is not reconstructed where unavailable.

### Final canonical status

**CLOSED.** The Airport Intelligence packages are production-reachable and no longer carry the planned-integration disposition.

### Prevention / future guard

Implemented analytical domains must not be described as production capabilities until a named runtime root, real data boundary, behavioral integration tests, and explicit claim limitations exist. Synthetic imports are not acceptable reachability evidence.
