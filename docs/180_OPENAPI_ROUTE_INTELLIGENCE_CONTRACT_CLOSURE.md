# Document 180 — OpenAPI Route Intelligence Contract Closure

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: final three-operation OpenAPI closure with mutation security
Baseline: `b67005b7ff6a944cffc2f4d846aec1009ea69e53`

## 1. Purpose

This increment closes the final OpenAPI gap by documenting the protected Route Intelligence
computation route and its two materialized read routes. Coverage increases from 35 to all 38
production public operations.

## 2. Added operations

```text
POST /api/v1/trajectories/{id}/route-intelligence
GET  /api/v1/trajectories/{id}/route-intelligence/latest
GET  /api/v1/trajectories/{id}/route-intelligence/history
```

The POST triggers bounded computation and persistence using the trajectory UUID. It accepts no
request body. The GET routes read already materialized records and remain publicly accessible.

## 3. Mutation authorization

OpenAPI adds:

```text
security scheme: InternalAPIKey
type: apiKey
location: header
name: X-Internal-API-Key
```

The POST operation alone declares this security requirement. The contract represents:

- `401 MUTATION_AUTHENTICATION_REQUIRED` for missing or invalid credentials;
- `503 MUTATION_AUTHENTICATION_UNAVAILABLE` when server-side mutation authentication is not
  configured;
- downstream Route Intelligence service-unavailable behavior under the same status boundary.

The server stores only a SHA-256 digest and validates raw keys with bounded length and
constant-time digest comparison.

## 4. Request and pagination contracts

All three routes require a trajectory UUID. History accepts:

```text
limit: integer, 1..100, default 20
before_as_of_time: optional RFC 3339 timestamp, exclusive cursor
```

The cursor is derived from the oldest returned record when another page exists.

## 5. Response model

The typed record includes:

- record identity, input fingerprint, and storage time;
- route schema version and availability status;
- trajectory, flight, aircraft, ICAO24, and callsign identity;
- observation window;
- optional inferred origin and destination;
- endpoint airport, distance, confidence, evidence, and limitations;
- route summary, overall confidence, provenance, and generation time.

History returns records, `has_more`, and an optional `next_before_as_of_time` cursor. Typed Route
Intelligence schemas reject unknown properties.

## 6. Error behavior

The operations preserve source-backed 400, 401, 404, 408, 409, 429, 500, 503, and 504
outcomes where applicable. The POST conflict response represents previously stored evidence
that conflicts with the computed result.

## 7. Deterministic mock alignment

The local Playwright mock adds all three paths, a complete evidence-bounded record fixture,
history pagination, and test-only mutation-key handling. Missing or invalid local credentials
return the same `401` code as production middleware. No production secret is embedded.

## 8. Verification markers

```text
OPENAPI_CONTRACT_PATHS=38
OPENAPI_PUBLIC_READ_OPERATIONS=37
OPENAPI_PROTECTED_MUTATION_OPERATIONS=1
OPENAPI_ROUTE_INTELLIGENCE_OPERATIONS=3
OPENAPI_DOCUMENTED_OPERATIONS=38
OPENAPI_MISSING_OPERATIONS=0
OPENAPI_EXTRA_OPERATIONS=0
OPENAPI_CONTRACT_SECURITY=PASS
PLAYWRIGHT_E2E_OPENAPI_PATHS=38
OPENAPI_ROUTE_INTELLIGENCE_CONTRACT_CLOSURE_INSTALL=PASS
```

## 9. Completion state

The public OpenAPI route inventory is closed at 38 of 38 operations. Future route changes are
handled as ordinary contract evolution and must preserve the same source, security, fixture,
documentation, and release-gate alignment.
