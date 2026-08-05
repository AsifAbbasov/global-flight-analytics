# Document 179 — OpenAPI Advanced Intelligence Read Surface

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: seventeen source-backed advanced public GET operations
Baseline: `1c569e0cced1f4700829b02d3573b1c9502fe4ca`
Closure successor: Document 180

## 1. Purpose

This increment closes the complete advanced public read surface while keeping the protected
Route Intelligence slice separate. It adds seventeen production-reachable GET operations to
OpenAPI and expands coverage from 18 to 35 of 38 public operations.

## 2. Added operations

```text
GET /api/v1/aircraft/{icao24}/transponder-evidence/latest
GET /api/v1/weather/current
GET /api/v1/analytics/metrics/active-aircraft
GET /api/v1/analytics/metrics/traffic-density
GET /api/v1/analytics/metrics/airport-activity
GET /api/v1/analytics/metrics/coverage-score
GET /api/v1/analytics/metrics/data-freshness
GET /api/v1/airports/intelligence/ranking
GET /api/v1/airports/{icao}/intelligence/overview
GET /api/v1/airports/{icao}/intelligence/history
GET /api/v1/airports/{icao}/intelligence/trends
GET /api/v1/historical-intelligence/aggregates/latest
GET /api/v1/historical-intelligence/aggregates/history
GET /api/v1/trajectories/{id}/projection-intelligence
GET /api/v1/trajectories/{id}/stability-intelligence
GET /api/v1/trajectories/{id}/weather-context
GET /api/v1/airspace/regions/{code}/analytics
```

## 3. Request contracts

The specification derives request rules from production handlers:

- transponder evidence uses a six-hex-character ICAO24 path identifier;
- current weather requires latitude and longitude;
- analytical recent windows remain 1–180 minutes and result limits 1–5000;
- traffic density requires a region and rejects client area values;
- airport activity requires ICAO and bounds radius to 100 kilometers;
- coverage and freshness inputs remain server-owned;
- Airport Intelligence uses 1–365 completed days and ranking limits 1–200;
- Historical Intelligence requires metric, scope, and granularity, with conditional scope
  identifiers and opaque cursor pagination;
- Projection, Stability, and Weather Context preserve UUID, RFC 3339, and positive-duration
  contracts;
- Stability accepts two to eight strictly increasing as-of timestamps;
- Airspace Intelligence requires an as-of time and accepts whole-minute windows from 60 to
  3600 seconds.

## 4. Evidence and schema boundaries

The contract keeps advanced evidence explicit:

- transponder observations are evidence-only and cannot assert a confirmed emergency;
- missing current-weather measurements remain `null`;
- analytical responses preserve confidence, eligibility, scope, provenance, data quality,
  warnings, limitations, and failures;
- Airport and Historical Intelligence preserve completed-window and pagination semantics;
- Projection, Stability, Weather Context, and Airspace Intelligence preserve fingerprints,
  scope guards, uncertainty, confidence, explanations, and limitations.

All typed objects reject unknown fields. Only intentionally strategy-specific payloads use the
explicit `OpenObject` schema.

## 5. Error semantics

The new operations represent relevant 400, 404, 408, 422, 429, 500, 503, and 504 outcomes.
Shared response components preserve typed error envelopes and `X-Request-ID`.

## 6. Deterministic mock alignment

The Playwright mock API mirrors all thirty-five OpenAPI paths. New static fixtures verify:

- evidence-only transponder semantics;
- nullable current-weather fields;
- Historical Intelligence `has_more` and `next_cursor` behavior;
- projection SHA-256 fingerprints and forecast points;
- Airspace Intelligence status and temporal coverage.

The existing four browser scenarios remain unchanged; advanced route fixtures do not imply
new end-to-end product coverage.

## 7. Verification markers

```text
OPENAPI_CONTRACT_PATHS=35
OPENAPI_CORE_READ_OPERATIONS=10
OPENAPI_ADVANCED_INTELLIGENCE_READ_OPERATIONS=17
OPENAPI_DOCUMENTED_OPERATIONS=35
OPENAPI_MISSING_OPERATIONS=3
OPENAPI_EXTRA_OPERATIONS=0
PLAYWRIGHT_E2E_OPENAPI_PATHS=35
OPENAPI_ADVANCED_INTELLIGENCE_READ_SURFACE_INSTALL=PASS
```

## 8. Closure successor

Document 180 closes the previously remaining Route Intelligence slice:

```text
POST /api/v1/trajectories/{id}/route-intelligence
GET  /api/v1/trajectories/{id}/route-intelligence/latest
GET  /api/v1/trajectories/{id}/route-intelligence/history
```

Document 180 defines the mutation security scheme, internal API key header, authorization
failures, request/response DTOs, and materialized read-history pagination without weakening
public read access or exposing `/internal/metrics`.
