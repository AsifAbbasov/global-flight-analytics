# Document 175 — OpenAPI Contract Foundation

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: source-backed contract for the stable public read API

## 1. Purpose

This document records the evolving OpenAPI 3.1 contract for the stable public API used by
frontend clients, deterministic browser fixtures, operational smoke tests, and reviewer tools.

The public specification now describes thirty-five stable GET operations. It combines:

- the original eight-operation system, region, airport, and current-traffic foundation;
- the ten-operation core read slice for aircraft, flights, flight states, trajectories,
  route context, and the active-aircraft metric;
- seventeen advanced intelligence reads for transponder evidence, current weather,
  analytical metrics, Airport Intelligence, Historical Intelligence, Projection
  Intelligence, Stability Intelligence, Weather Context, and Airspace Intelligence.

The only production public operations still outside this document are the final Route
Intelligence slice:

```text
POST /api/v1/trajectories/{id}/route-intelligence
GET  /api/v1/trajectories/{id}/route-intelligence/latest
GET  /api/v1/trajectories/{id}/route-intelligence/history
```

Those three operations remain separate because the POST route requires an explicit security
scheme and mutation-authorization contract.

## 2. Contract Location

```text
openapi/openapi.json
```

JSON is used because OpenAPI 3.1 accepts JSON directly and the repository can validate the
document without adding another runtime dependency. The server URL remains relative (`/`), so
the contract does not hard-code Render, Vercel, localhost, or another deployment origin.

## 3. Source Alignment

The contract is derived from production Fiber registrations, handler request parsing, HTTP
DTOs, frontend transport parsers where those parsers intentionally preserve a backend-owned
shape, and executable repository verification.

The advanced read slice adds source evidence for:

```text
apps/api/internal/http/handlers/transponder_evidence.go
apps/api/internal/http/handlers/weather.go
apps/api/internal/http/handlers/analytical_metrics.go
apps/api/internal/http/handlers/analytical_production_snapshot.go
apps/api/internal/http/handlers/airport_intelligence.go
apps/api/internal/http/handlers/historical_intelligence.go
apps/api/internal/http/handlers/projection_intelligence.go
apps/api/internal/http/handlers/stability_intelligence.go
apps/api/internal/http/handlers/weather_context.go
apps/api/internal/http/handlers/airspace_region_analytics.go
```

The contract does not publish parameters rejected by the backend. In particular:

- `area_square_kilometers` remains server-derived for traffic density;
- client-supplied coverage counters and freshness timestamps remain prohibited;
- quality-metric `limit` remains server-owned;
- historical scope identifiers remain conditional on the selected scope;
- stability history requires two to eight strictly increasing RFC 3339 timestamps;
- Airspace Intelligence windows remain whole-minute values from 60 to 3600 seconds.

## 4. Schemas and Boundaries

The specification preserves:

- typed success and error envelopes;
- `X-Request-ID` and rate-limit headers;
- nullable weather and altitude measurements rather than invented zero values;
- explicit confidence, limitations, provenance, scope guards, and evidence boundaries;
- opaque Historical Intelligence pagination cursors;
- UUID and RFC 3339 request contracts;
- research-only transponder evidence that cannot be represented as a confirmed emergency;
- source-backed analytical metrics with optional server-owned data-quality reports;
- explicit timeout, service-unavailable, not-found, and bounded-policy responses.

Typed DTO schemas use `additionalProperties: false`. `OpenObject` is the only intentionally
open schema introduced by the advanced slice and is limited to backend fields that are
explicitly strategy-specific or retained as unknown by the frontend contract.

The internal metrics endpoint remains excluded:

```text
GET /internal/metrics
```

## 5. Verification

```bash
pnpm run test:openapi-route-inventory
pnpm run verify:openapi-route-inventory
pnpm run test:openapi-contract
pnpm run verify:openapi-contract
pnpm run test:playwright-e2e-contract
pnpm run verify:playwright-e2e
```

Required success markers now include:

```text
SOURCE_PUBLIC_OPERATIONS=38
OPENAPI_DOCUMENTED_OPERATIONS=35
OPENAPI_MISSING_OPERATIONS=3
OPENAPI_EXTRA_OPERATIONS=0
OPENAPI_CONTRACT_PATHS=35
OPENAPI_CORE_READ_OPERATIONS=10
OPENAPI_ADVANCED_INTELLIGENCE_READ_OPERATIONS=17
OPENAPI_CONTRACT_SCHEMAS=PASS
OPENAPI_CONTRACT_ROUTE_DRIFT=PASS
OPENAPI_CONTRACT=PASS
```

The same static contracts run in the dedicated `OpenAPI Contract` workflow and in the full
repository release gate.

## 6. Playwright Dependency

The deterministic Playwright mock API mirrors all thirty-five documented paths. The four
existing browser scenarios remain unchanged; new fixtures provide a contract-aligned base for
later browser assertions without calling public deployments or real providers.

## 7. Evolution Rule

A route may enter the public contract only when all of the following are true:

1. the Fiber route is production-reachable;
2. parameters and response DTOs are source-backed;
3. success and error behavior are represented without inventing evidence;
4. local references resolve and operation identifiers are unique;
5. the route inventory, OpenAPI verifier, mock API, and release gate pass.

The final Route Intelligence closure must additionally define the mutation credential header,
security scheme, authorization failures, and separation between computation-triggering POST
behavior and materialized GET reads.
