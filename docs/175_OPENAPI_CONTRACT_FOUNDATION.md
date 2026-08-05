# Document 175 — OpenAPI Contract Foundation

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: source-backed contract for the complete stable public API

## 1. Purpose

This document records the complete OpenAPI 3.1 contract for the stable public API used by
frontend clients, deterministic browser fixtures, operational smoke tests, and reviewer tools.

The specification now documents all thirty-eight production public operations:

- thirty-seven source-backed GET operations;
- one protected Route Intelligence POST operation.

The final Route Intelligence slice adds:

```text
POST /api/v1/trajectories/{id}/route-intelligence
GET  /api/v1/trajectories/{id}/route-intelligence/latest
GET  /api/v1/trajectories/{id}/route-intelligence/history
```

The internal metrics endpoint remains outside the public contract:

```text
GET /internal/metrics
```

## 2. Contract Location

```text
openapi/openapi.json
```

JSON is used because OpenAPI 3.1 accepts JSON directly and the repository can validate the
document without adding another runtime dependency. The server URL remains relative (`/`), so
the contract does not hard-code Render, Vercel, localhost, or another deployment origin.

## 3. Source Alignment

The contract is derived from production Fiber registrations, handler request parsing, HTTP
DTOs, authorization middleware, internal API-key rules, frontend transport parsers where those
parsers intentionally preserve a backend-owned shape, and executable repository verification.

Route Intelligence evidence is anchored in:

```text
apps/api/internal/server/route_intelligence_database_routes.go
apps/api/internal/http/handlers/route_intelligence.go
apps/api/internal/http/dto/route_intelligence.go
apps/api/internal/middleware/mutation_authorization.go
apps/api/internal/security/internalapikey/key.go
apps/api/internal/routeintelligence/routestore/contracts.go
apps/api/internal/routeintelligence/routecontract/model.go
```

## 4. Security and Mutation Boundary

The mutation operation declares the `InternalAPIKey` OpenAPI security scheme:

```text
header: X-Internal-API-Key
type: apiKey
```

The server stores only a SHA-256 digest. Clients send the raw key, which must be between 32 and
256 characters. Missing or invalid credentials return `401 MUTATION_AUTHENTICATION_REQUIRED`. Missing
server-side mutation-authentication configuration returns
`503 MUTATION_AUTHENTICATION_UNAVAILABLE` before the handler runs.

The two materialized GET operations remain publicly readable and do not inherit the mutation
credential requirement.

## 5. Schemas and Boundaries

The specification preserves:

- typed success and error envelopes;
- `X-Request-ID` and rate-limit headers;
- UUID and RFC 3339 request contracts;
- nullable airport elevation rather than invented zero values;
- explicit confidence, evidence, limitations, provenance, and route status;
- origin and destination as optional inferred endpoints;
- history limits from 1 to 100, defaulting to 20;
- an exclusive `before_as_of_time` RFC 3339 pagination cursor;
- conflict, timeout, not-found, service-unavailable, and authorization outcomes.

Typed DTO schemas use `additionalProperties: false`. The Route Intelligence POST has no request
body because the production handler derives the computation request solely from the trajectory
UUID.

## 6. Verification

```bash
pnpm run test:openapi-route-inventory
pnpm run verify:openapi-route-inventory
pnpm run test:openapi-contract
pnpm run verify:openapi-contract
pnpm run test:playwright-e2e-contract
pnpm run verify:playwright-e2e
```

Required success markers include:

```text
SOURCE_PUBLIC_OPERATIONS=38
OPENAPI_DOCUMENTED_OPERATIONS=38
OPENAPI_MISSING_OPERATIONS=0
OPENAPI_EXTRA_OPERATIONS=0
OPENAPI_CONTRACT_PATHS=38
OPENAPI_PUBLIC_READ_OPERATIONS=37
OPENAPI_PROTECTED_MUTATION_OPERATIONS=1
OPENAPI_ROUTE_INTELLIGENCE_OPERATIONS=3
OPENAPI_CONTRACT_SCHEMAS=PASS
OPENAPI_CONTRACT_SECURITY=PASS
OPENAPI_CONTRACT_ROUTE_DRIFT=PASS
OPENAPI_CONTRACT=PASS
```

## 7. Playwright Dependency

The deterministic Playwright mock API mirrors all thirty-eight documented paths. The local
Route Intelligence mutation fixture requires a deterministic test-only key and never uses a
production credential. Existing browser scenarios remain unchanged.

## 8. Evolution Rule

Any future public route change must update production source, OpenAPI, route inventory,
deterministic fixtures, contract tests, documentation, and the release gate in one reviewed
increment. Mutation routes must also define an explicit security scheme and source-backed
authorization failures.
