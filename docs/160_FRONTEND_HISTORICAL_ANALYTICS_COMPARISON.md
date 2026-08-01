# Document 160 — Frontend Historical Analytics and Comparison

## Status

Implemented against exact baseline `52fb54a389016fb11d9d69bcb3a9b1437eeb3dd0`.
Formal closure still requires exact-commit Continuous Integration evidence after commit and push.

## Purpose

This increment exposes the existing production Historical Intelligence aggregate store through a typed, evidence-aware frontend workspace. It does not duplicate server-owned metric construction or invent unsupported region metrics.

## Delivered scope

- global, airport and route aggregate scopes;
- server-catalog metric filtering;
- hour, day, week and custom granularity selection;
- latest persisted aggregate loading;
- bounded aggregate history loading;
- runtime response validation;
- bucket status, coverage, confidence and limitation visibility;
- server-published previous-period comparison;
- side-by-side comparison of two compatible persisted aggregates;
- deterministic historical series and limitation ordering;
- navigation and application-shell integration.

## Production API contract

The frontend reads:

- `GET /api/v1/historical-intelligence/aggregates/latest`
- `GET /api/v1/historical-intelligence/aggregates/history`

Query parameters preserve the backend contract for metric, scope, granularity and scope identifiers. History remains bounded to twenty records in the current interface and reports when additional pages exist.

## Evidence boundaries

- unavailable buckets are not converted to zero-valued evidence;
- percentage change remains unavailable when the earlier value is zero;
- records with different metric, scope or granularity are not compared;
- region scope is not exposed because the current production metric catalog permits no metric for it;
- airport activity remains probable Route Intelligence evidence rather than filed flight-plan data;
- the workspace is research and visualization only, not operational guidance.

## Verification

The package adds eight dependency-free model tests covering catalog scope, normalization, completeness, chronological series ordering, unavailable comparisons, aggregation-aware record comparison, incompatibility rejection and deterministic ordering.

The installer must pass frontend tests, ESLint, TypeScript, production build, dependency policy, exact changed-file manifest and rollback verification before installation is considered complete.
