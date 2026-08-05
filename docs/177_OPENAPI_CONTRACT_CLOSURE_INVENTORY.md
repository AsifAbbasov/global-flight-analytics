# OpenAPI Contract Closure Route Inventory

Status: IMPLEMENTED — permanent inventory gate active
Initial inventory baseline: `b44cc67bc6fd6076c2a0ff29174283e28fd79724`
Core read expansion baseline: `8a873c6c60add160b7f355f2a90bde6771990156`

## Purpose

This document records the source-backed HTTP operation inventory used to expand the public
OpenAPI contract without inventing routes, omitting production handlers, or crossing the
public and internal security boundary.

## Production HTTP surface

The production Fiber composition contains:

```text
PUBLIC_OPERATIONS=38
PUBLIC_GET_OPERATIONS=37
PUBLIC_POST_OPERATIONS=1
INTERNAL_OPERATIONS=1
TOTAL_HTTP_OPERATIONS=39
```

The public surface is mounted under `/api/v1` through the nested `/api` and `/v1` Fiber groups.

The internal surface contains only:

```text
GET /internal/metrics
```

The internal metrics operation is protected by `metricsAuthorization` and is intentionally
excluded from the public OpenAPI specification.

## Public mutation boundary

The only public state-changing operation is:

```text
POST /api/v1/trajectories/{id}/route-intelligence
```

Its production registration must preserve `mutationAuthorization` before the request handler.
The inventory gate validates this source contract but does not invent an OpenAPI security
scheme. The exact credential header and failure behavior must be derived from the middleware
before the mutation enters `openapi/openapi.json`.

## Current OpenAPI gap

After the core read expansion, `openapi/openapi.json` contains eighteen stable public GET
operations. Source comparison produces:

```text
SOURCE_PUBLIC_OPERATIONS=38
OPENAPI_DOCUMENTED_OPERATIONS=18
OPENAPI_MISSING_OPERATIONS=20
OPENAPI_EXTRA_OPERATIONS=0
```

The zero-extra result remains essential: every documented route exists in production source.
The remaining problem is incomplete coverage, not fabricated contract surface.

## Permanent verifier

The repository contains:

```text
scripts/verify-openapi-route-inventory.mjs
scripts/verify-openapi-route-inventory.test.mjs
```

The verifier:

- discovers production `GET`, `POST`, `PUT`, `PATCH`, and `DELETE` registrations from non-test Go files in `apps/api/internal/server`;
- resolves nested Fiber groups;
- resolves constant-backed route paths;
- converts Fiber `:parameter` syntax into OpenAPI `{parameter}` syntax;
- classifies the complete 38-operation public inventory and the one-operation internal inventory;
- compares source operations with `openapi/openapi.json`;
- rejects duplicate, missing, and unclassified source registrations;
- verifies the route-intelligence mutation authorization boundary;
- verifies the internal metrics authorization boundary;
- rejects exposure of `/internal` routes in the public OpenAPI specification.

## Current reviewed closure slice

Document 178 adds ten source-backed GET operations:

```text
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

## Test contract

The permanent inventory suite covers:

1. repository-level verifier execution;
2. the exact `37 GET + 1 POST` public method distribution;
3. the single internal metrics operation;
4. the current exact `20 missing + 0 extra` gap;
5. nested Fiber group resolution;
6. constant-backed route resolution;
7. rejection of unprotected mutation and metrics registrations.

## Continuous Integration boundary

The `OpenAPI Contract` workflow runs the route inventory and the OpenAPI verifier. Its artifact
contains the public specification and Documents 175, 177, and 178.

## Completion boundary

This increment is complete when the following markers pass:

```text
SOURCE_PUBLIC_OPERATIONS=38
SOURCE_PUBLIC_GET_OPERATIONS=37
SOURCE_PUBLIC_POST_OPERATIONS=1
SOURCE_INTERNAL_OPERATIONS=1
OPENAPI_DOCUMENTED_OPERATIONS=18
OPENAPI_MISSING_OPERATIONS=20
OPENAPI_EXTRA_OPERATIONS=0
OPENAPI_MUTATION_AUTHORIZATION=PASS
OPENAPI_INTERNAL_METRICS_BOUNDARY=PASS
OPENAPI_ROUTE_INVENTORY=PASS
OPENAPI_CONTRACT=PASS
```

Twenty public operations remain for later reviewed slices: transponder evidence, route
intelligence, weather, analytical metrics, airport intelligence, historical intelligence,
projection, stability, weather context, and airspace analytics.
