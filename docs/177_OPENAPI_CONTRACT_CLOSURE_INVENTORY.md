# OpenAPI Contract Closure Route Inventory

Status: IMPLEMENTED — permanent inventory gate
Current expansion baseline: `1c569e0cced1f4700829b02d3573b1c9502fe4ca`

## Purpose

This document records the source-backed HTTP operation inventory used to close the public
OpenAPI contract without inventing routes, omitting production handlers, or crossing the
public/internal security boundary.

## Production HTTP surface

```text
PUBLIC_OPERATIONS=38
PUBLIC_GET_OPERATIONS=37
PUBLIC_POST_OPERATIONS=1
INTERNAL_OPERATIONS=1
TOTAL_HTTP_OPERATIONS=39
```

The public surface is mounted under `/api/v1`. The only internal operation is:

```text
GET /internal/metrics
```

It remains protected by `metricsAuthorization` and excluded from the public OpenAPI document.

## Public mutation boundary

The only public state-changing operation is:

```text
POST /api/v1/trajectories/{id}/route-intelligence
```

Production registration must preserve `mutationAuthorization` before the handler. The final
OpenAPI slice must derive the exact credential and error semantics from that middleware.

## Current OpenAPI gap

After the core and advanced read expansions:

```text
SOURCE_PUBLIC_OPERATIONS=38
OPENAPI_DOCUMENTED_OPERATIONS=35
OPENAPI_MISSING_OPERATIONS=3
OPENAPI_EXTRA_OPERATIONS=0
```

The exact remaining gap is:

```text
POST /api/v1/trajectories/{id}/route-intelligence
GET  /api/v1/trajectories/{id}/route-intelligence/latest
GET  /api/v1/trajectories/{id}/route-intelligence/history
```

The verifier rejects any other three-operation gap, so the number alone cannot conceal drift.

## Permanent verifier

```text
scripts/verify-openapi-route-inventory.mjs
scripts/verify-openapi-route-inventory.test.mjs
```

The verifier:

- discovers production GET, POST, PUT, PATCH, and DELETE registrations;
- resolves nested Fiber groups and constant-backed paths;
- converts Fiber parameters into OpenAPI parameters;
- classifies all 38 public operations and the internal metrics operation;
- compares production source and OpenAPI operations;
- rejects duplicate, missing, extra, or unclassified registrations;
- verifies mutation and internal metrics authorization boundaries;
- requires the remaining gap to be exactly the Route Intelligence slice.

## Completion boundary

This inventory remains open until all 38 public operations are documented and the final
mutation security contract passes. The current advanced-read increment is complete when:

```text
OPENAPI_DOCUMENTED_OPERATIONS=35
OPENAPI_MISSING_OPERATIONS=3
OPENAPI_EXTRA_OPERATIONS=0
OPENAPI_MUTATION_AUTHORIZATION=PASS
OPENAPI_INTERNAL_METRICS_BOUNDARY=PASS
OPENAPI_ROUTE_INVENTORY=PASS
```
