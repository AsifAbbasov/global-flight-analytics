# OpenAPI Contract Closure Route Inventory

Status: IMPLEMENTED — complete permanent inventory gate
Closure baseline: `b67005b7ff6a944cffc2f4d846aec1009ea69e53`

## Purpose

This document records the source-backed HTTP operation inventory used to keep the complete
public OpenAPI contract aligned with production registrations and authorization boundaries.

## Production HTTP surface

```text
PUBLIC_OPERATIONS=38
PUBLIC_GET_OPERATIONS=37
PUBLIC_POST_OPERATIONS=1
INTERNAL_OPERATIONS=1
TOTAL_HTTP_OPERATIONS=39
```

The public surface is mounted under `/api/v1`. The only internal operation is
`GET /internal/metrics`; it remains protected by `metricsAuthorization` and excluded from the
public OpenAPI document.

## Public mutation boundary

The only public state-changing operation is:

```text
POST /api/v1/trajectories/{id}/route-intelligence
```

Production registration preserves `mutationAuthorization` before the handler. OpenAPI declares
`InternalAPIKey` in the `X-Internal-API-Key` header and requires it only for this POST.

## Complete OpenAPI coverage

```text
SOURCE_PUBLIC_OPERATIONS=38
OPENAPI_DOCUMENTED_OPERATIONS=38
OPENAPI_MISSING_OPERATIONS=0
OPENAPI_EXTRA_OPERATIONS=0
```

The two public Route Intelligence reads remain unprotected:

```text
GET /api/v1/trajectories/{id}/route-intelligence/latest
GET /api/v1/trajectories/{id}/route-intelligence/history
```

## Permanent verifier

```text
scripts/verify-openapi-route-inventory.mjs
scripts/verify-openapi-route-inventory.test.mjs
```

The verifier discovers production routes, resolves nested Fiber groups and constant-backed
paths, classifies all public and internal operations, rejects duplicates or drift, and audits
the mutation and internal-metrics authorization boundaries.

## Completion markers

```text
OPENAPI_DOCUMENTED_OPERATIONS=38
OPENAPI_MISSING_OPERATIONS=0
OPENAPI_EXTRA_OPERATIONS=0
OPENAPI_MUTATION_AUTHORIZATION=PASS
OPENAPI_INTERNAL_METRICS_BOUNDARY=PASS
OPENAPI_ROUTE_INVENTORY=PASS
```
