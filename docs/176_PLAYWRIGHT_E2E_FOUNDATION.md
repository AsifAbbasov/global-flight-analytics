# Document 176 — Playwright End-to-End Testing Foundation

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: deterministic browser testing for the public product shell

## 1. Purpose

This increment maintains a reproducible Playwright boundary on top of the complete source-backed
OpenAPI 3.1 contract. The suite exercises the real production Next.js build and browser client
without using the public Render API, Vercel Preview, or live provider data.

## 2. Runtime Architecture

```text
Playwright Chromium
        ↓
Next.js production server on 127.0.0.1:3000
        ↓
Deterministic mock API on 127.0.0.1:8091
        ↓
OpenAPI 3.1 path, method, security, and response boundary
```

The mock API mirrors exactly the thirty-eight paths currently documented in
`openapi/openapi.json`, including the protected Route Intelligence POST and its two materialized
GET reads.

A private test control endpoint selects deterministic success and failure scenarios. It is not
part of the production API contract.

## 3. Covered User Scenarios

The original foundation contained four Chromium end-to-end scenarios:

1. the server-rendered application shell publishes a healthy initial aircraft snapshot;
2. selecting Azerbaijan updates the visible workspace and canonical shareable URL;
3. responsive mobile navigation keeps major product sections reachable;
4. a traffic API outage preserves the shell and recovers through the visible Retry path.

The larger mock surface is a transport-contract foundation, not a false claim that every route
already has a dedicated browser workflow. Route Intelligence fixture tests verify method,
credential, record, and history semantics at the mock-contract layer.

The current expanded product-journey coverage is recorded separately in
`docs/186_PLAYWRIGHT_PRODUCT_COVERAGE.md`.

## 4. Isolated Playwright Runtime

The runner installs exactly `@playwright/test@1.62.0` under
`apps/web/e2e/node_modules` with `--no-save` and `--package-lock=false`. The pnpm lockfile and
application dependency graph are not modified.

## 5. Commands

```bash
pnpm run test:playwright-e2e-contract
pnpm run verify:playwright-e2e
pnpm run run:playwright-e2e
```

## 6. Evidence

```text
PLAYWRIGHT_E2E_VERSION=1.62.0
PLAYWRIGHT_E2E_OPENAPI_PATHS=38
PLAYWRIGHT_E2E_SCENARIOS=4
PLAYWRIGHT_E2E_MOCK_API=PASS
PLAYWRIGHT_E2E_CONTRACT=PASS
```

The full browser runner additionally emits `PLAYWRIGHT_E2E_BROWSER=chromium`,
`PLAYWRIGHT_E2E_PUBLIC_NETWORK=DISABLED`, and `PLAYWRIGHT_E2E=PASS`.

## 7. Safety and Scope Boundary

The public Render deployment is never targeted by this suite. No production mutation
credential, real provider account, or external aviation service is required.

The local Route Intelligence fixture key is deterministic test data. It is not a production
secret and is accepted only by the local mock implementation. No fixture asserts a filed flight
plan or operational navigation truth; inferred route evidence retains limitations and
provenance.
