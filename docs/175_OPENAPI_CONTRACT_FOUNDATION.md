# Document 175 — OpenAPI Contract Foundation

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: source-backed contract for the stable public read API

## 1. Purpose

This increment establishes an OpenAPI 3.1 contract for the stable public read surface used by
the current frontend, operational smoke tests, and the containerized Grafana k6 baseline.

The contract is intentionally bounded. It describes eight stable GET operations:

```text
GET /api/v1/health
GET /api/v1/ready
GET /api/v1/version
GET /api/v1/regions
GET /api/v1/regions/{code}
GET /api/v1/airports
GET /api/v1/airports/{icao}
GET /api/v1/traffic/current
```

Advanced analytical, historical, projection, weather, airspace, transponder-evidence, and
mutation endpoints remain outside this first contract. They must be added in later reviewed
increments rather than represented by incomplete schemas.

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

The verifier checks the contract against:

```text
apps/api/internal/server/server.go
apps/api/internal/server/core_database_routes.go
apps/api/internal/http/response/response.go
apps/api/internal/http/dto/system.go
apps/api/internal/http/dto/region.go
apps/api/internal/http/dto/airport.go
apps/api/internal/http/dto/traffic.go
apps/api/internal/http/handlers/traffic.go
apps/api/internal/http/handlers/readiness.go
```

This protects route registration, path parameters, the optional `region` query parameter,
typed success envelopes, typed error envelopes, and canonical JSON field names.

## 4. Schemas and Boundaries

The contract includes:

- health, readiness, and version data;
- canonical regions and geographic bounds;
- airport list and airport profile records;
- nullable airport elevation with explicit status;
- current traffic telemetry with nullable altitude and explicit altitude semantics;
- not-found, rate-limit, service-unavailable, and internal-error envelopes;
- `X-Request-ID` and rate-limit response headers.

The contract is a public read contract. POST, PUT, PATCH, DELETE, TRACE, and CONNECT operations
are prohibited from this foundation.

## 5. Verification

```bash
pnpm run test:openapi-contract
pnpm run verify:openapi-contract
```

Required success markers:

```text
OPENAPI_CONTRACT_PATHS=8
OPENAPI_CONTRACT_SCHEMAS=PASS
OPENAPI_CONTRACT_ROUTE_DRIFT=PASS
OPENAPI_CONTRACT=PASS
```

The same checks run in the dedicated `OpenAPI Contract` GitHub Actions workflow and in the
repository release gate.

## 6. Playwright Dependency

The next Playwright increment must consume this contract boundary when selecting backend
fixtures and endpoint assertions. Playwright must not invent response fields that are absent
from `openapi/openapi.json`.

## 7. Evolution Rule

A route may enter this contract only when all of the following are true:

1. the Fiber route is production-reachable;
2. request parameters and response DTOs are stable;
3. success and error behavior are covered by backend tests;
4. the OpenAPI schema is source-aligned;
5. the verifier and Continuous Integration workflow pass.

This document does not claim that every backend endpoint is already described.
