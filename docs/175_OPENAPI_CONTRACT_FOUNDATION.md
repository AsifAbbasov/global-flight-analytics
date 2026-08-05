# Document 175 — OpenAPI Contract Foundation

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: source-backed contract for the stable public read API

## 1. Purpose

This document records the OpenAPI 3.1 foundation and its first reviewed expansion. The
contract is consumed by the current frontend, operational smoke tests, the containerized
Grafana k6 baseline, and the deterministic Playwright mock API.

The public specification now describes eighteen stable GET operations:

```text
GET /api/v1/health
GET /api/v1/ready
GET /api/v1/version
GET /api/v1/regions
GET /api/v1/regions/{code}
GET /api/v1/airports
GET /api/v1/airports/{icao}
GET /api/v1/traffic/current
GET /api/v1/metrics/active-aircraft
GET /api/v1/aircraft
GET /api/v1/aircraft/{icao24}
GET /api/v1/flights
GET /api/v1/flights/{id}
GET /api/v1/flights/{flightID}/states
GET /api/v1/aircraft/{icao24}/latest-state
GET /api/v1/aircraft/{icao24}/trajectory
GET /api/v1/trajectories/{id}
GET /api/v1/aircraft/{icao24}/route-context
```

The original eight-operation foundation remains intact. Document 178 adds the ten-operation
core read slice for aircraft, flights, flight states, trajectories, route context, and the
active-aircraft metric.

Advanced analytical, historical, projection, weather, airspace, transponder-evidence, and
route-intelligence endpoints remain outside this reviewed slice. The protected mutation is
also still deferred until its exact OpenAPI security scheme and authorization failures are
source-derived.

## 2. Contract Location

```text
openapi/openapi.json
```

JSON is used because OpenAPI 3.1 accepts JSON directly, Node.js can parse it without adding a
runtime dependency, and the repository can validate the document in release and Continuous
Integration workflows.

The server URL is relative (`/`). The contract therefore does not hard-code Render, localhost,
Vercel, or another deployment origin.

## 3. Source Alignment

The verifier checks the contract against the production Fiber composition, typed response
envelope, DTOs, and handlers. Core evidence includes:

```text
apps/api/internal/server/server.go
apps/api/internal/server/core_database_routes.go
apps/api/internal/http/response/response.go
apps/api/internal/http/dto/system.go
apps/api/internal/http/dto/region.go
apps/api/internal/http/dto/airport.go
apps/api/internal/http/dto/traffic.go
apps/api/internal/http/dto/metrics.go
apps/api/internal/http/dto/aircraft.go
apps/api/internal/http/dto/flight.go
apps/api/internal/http/dto/flightstate.go
apps/api/internal/http/dto/trajectory.go
apps/api/internal/http/dto/route_context.go
apps/api/internal/http/handlers/traffic.go
apps/api/internal/http/handlers/metrics.go
apps/api/internal/http/handlers/trajectories.go
apps/api/internal/http/handlers/route_context.go
apps/api/internal/http/handlers/readiness.go
```

This protects route registration, path parameters, active-aircraft query bounds, typed
success and error envelopes, nullable altitude semantics, trajectory composition, and
canonical JSON field names.

## 4. Schemas and Boundaries

The contract includes:

- health, readiness, and version data;
- canonical regions and geographic bounds;
- airport list and airport profile records;
- current traffic telemetry with nullable altitude and explicit altitude semantics;
- active-aircraft metric scope, confidence, sources, limitations, and bounded window;
- aircraft list and profile records;
- flight list and profile records;
- flight-state history and latest-state records;
- trajectories, segments, coverage gaps, identity, quality, and timestamps;
- inferred route context with optional origin and destination evidence;
- bad-request, not-found, rate-limit, service-unavailable, and internal-error envelopes;
- `X-Request-ID` and rate-limit response headers.

The current contract remains public and read-only. POST, PUT, PATCH, DELETE, TRACE, and
CONNECT operations are prohibited from this slice.

## 5. Verification

```bash
pnpm run test:openapi-contract
pnpm run verify:openapi-contract
pnpm run test:openapi-route-inventory
pnpm run verify:openapi-route-inventory
```

Required success markers:

```text
OPENAPI_CONTRACT_PATHS=18
OPENAPI_CORE_READ_OPERATIONS=10
OPENAPI_CONTRACT_SCHEMAS=PASS
OPENAPI_CONTRACT_ROUTE_DRIFT=PASS
OPENAPI_CONTRACT=PASS
OPENAPI_DOCUMENTED_OPERATIONS=18
OPENAPI_MISSING_OPERATIONS=20
OPENAPI_EXTRA_OPERATIONS=0
OPENAPI_ROUTE_INVENTORY=PASS
```

The same checks run in the dedicated `OpenAPI Contract` GitHub Actions workflow and in the
repository release gate.

## 6. Playwright Dependency

The deterministic Playwright mock API mirrors all eighteen documented paths. Browser tests
still exercise the product shell and traffic recovery scenarios; the additional core fixtures
exist to prevent OpenAPI-to-mock drift and to support later selected-aircraft workflows without
inventing transport shapes.

## 7. Evolution Rule

A route may enter this contract only when all of the following are true:

1. the Fiber route is production-reachable;
2. request parameters and response DTOs are stable;
3. success and error behavior are source-backed;
4. nullable and evidence-limited semantics are preserved;
5. the OpenAPI schema is source-aligned;
6. the route inventory, contract verifier, Playwright mock contract, and release gate pass.

This document does not claim that every backend endpoint is already described. Twenty public
operations remain outside the current specification.
