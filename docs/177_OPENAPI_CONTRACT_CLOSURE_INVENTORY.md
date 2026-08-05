# OpenAPI Contract Closure Route Inventory

Status: IMPLEMENTED — inventory gate only
Baseline commit: `b44cc67bc6fd6076c2a0ff29174283e28fd79724`

## Purpose

This document records the source-backed HTTP operation inventory required before the public OpenAPI contract can be expanded from its eight-operation foundation to the complete production public surface.

This increment does not claim that the OpenAPI closure is complete. It establishes the permanent evidence needed to perform that closure without inventing routes, omitting production handlers, or crossing the public and internal security boundary.

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

The internal metrics operation is protected by `metricsAuthorization` and is intentionally excluded from the public OpenAPI specification.

## Public mutation boundary

The only public state-changing operation is:

```text
POST /api/v1/trajectories/{id}/route-intelligence
```

Its production registration must preserve `mutationAuthorization` before the request handler. The inventory gate validates this source contract but deliberately does not invent an OpenAPI security scheme. The exact credential header and failure behavior must be derived from the middleware before the mutation is added to `openapi/openapi.json`.

## Current OpenAPI gap

At this baseline, `openapi/openapi.json` contains eight stable public GET operations. Source comparison produces:

```text
SOURCE_PUBLIC_OPERATIONS=38
OPENAPI_DOCUMENTED_OPERATIONS=8
OPENAPI_MISSING_OPERATIONS=30
OPENAPI_EXTRA_OPERATIONS=0
```

The zero-extra result is important: the foundation specification contains no route that is absent from production source. The remaining problem is incomplete coverage, not fabricated contract surface.

## Permanent verifier

The repository now contains:

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

## Test contract

The seven-test suite covers:

1. repository-level verifier execution;
2. the exact `37 GET + 1 POST` public method distribution;
3. the single internal metrics operation;
4. the exact `30 missing + 0 extra` foundation gap;
5. nested Fiber group resolution;
6. constant-backed route resolution;
7. rejection of unprotected mutation and metrics registrations.

## Continuous Integration boundary

The `OpenAPI Contract` workflow runs both the route inventory and the existing OpenAPI foundation verifier. The workflow artifact includes the public specification, Document 175, and this inventory document.

## Completion boundary

This increment is complete when the following markers pass:

```text
SOURCE_PUBLIC_OPERATIONS=38
SOURCE_PUBLIC_GET_OPERATIONS=37
SOURCE_PUBLIC_POST_OPERATIONS=1
SOURCE_INTERNAL_OPERATIONS=1
OPENAPI_DOCUMENTED_OPERATIONS=8
OPENAPI_MISSING_OPERATIONS=30
OPENAPI_EXTRA_OPERATIONS=0
OPENAPI_MUTATION_AUTHORIZATION=PASS
OPENAPI_INTERNAL_METRICS_BOUNDARY=PASS
OPENAPI_ROUTE_INVENTORY=PASS
OPENAPI_CONTRACT=PASS
```

The next increment is the actual expansion of `openapi/openapi.json`. Until that work is implemented and verified, the public OpenAPI contract remains a deliberately limited foundation.
