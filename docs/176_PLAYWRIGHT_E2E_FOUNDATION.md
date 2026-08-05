# Document 176 — Playwright End-to-End Testing Foundation

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: deterministic browser testing for the public product shell

## 1. Purpose

This increment maintains a reproducible Playwright boundary on top of the source-backed
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
OpenAPI 3.1 path and response boundary
```

The mock API mirrors exactly the thirty-five paths currently documented in:

```text
openapi/openapi.json
```

Its deterministic fixture surface now includes:

- the original system, region, airport, and traffic foundation;
- aircraft, flights, flight states, trajectories, route context, and active-aircraft metric;
- transponder evidence and nullable current weather;
- five analytical metrics with server-owned quality semantics;
- Airport and Historical Intelligence, including opaque history pagination;
- Projection, Stability, Weather Context, and Airspace Intelligence evidence structures.

A private test control endpoint selects deterministic success and failure scenarios. It is
not part of the production API contract.

## 3. Covered User Scenarios

The foundation still contains four Chromium end-to-end scenarios:

1. the server-rendered application shell publishes a healthy initial aircraft snapshot;
2. selecting Azerbaijan updates the visible workspace and canonical shareable URL;
3. responsive mobile navigation keeps major product sections reachable;
4. a traffic API outage preserves the shell and recovers through the visible Retry path.

The larger mock surface is a transport-contract foundation, not a false claim that all thirty-
five routes already have dedicated browser workflows. Advanced product assertions remain
separate reviewed increments.

Tests use roles, accessible names, labels, and visible status text. They do not depend on
generated Tailwind classes or test-only production attributes. Continuous Integration treats
flaky success-after-retry as failure.

## 4. Isolated Playwright Runtime

The runner installs exactly:

```text
@playwright/test@1.62.0
```

Installation remains isolated under `apps/web/e2e/node_modules` with `--no-save` and
`--package-lock=false`. The pnpm lockfile and application dependency graph are not modified.
Chromium is installed explicitly by the dedicated workflow.

## 5. Commands

```bash
pnpm run test:playwright-e2e-contract
pnpm run verify:playwright-e2e
pnpm run run:playwright-e2e
```

The full browser suite remains outside the general release script because it downloads a
browser and starts two local services. Static fixture and OpenAPI alignment remain inside the
release gate.

## 6. Evidence

Required static markers include:

```text
PLAYWRIGHT_E2E_VERSION=1.62.0
PLAYWRIGHT_E2E_OPENAPI_PATHS=35
PLAYWRIGHT_E2E_SCENARIOS=4
PLAYWRIGHT_E2E_MOCK_API=PASS
PLAYWRIGHT_E2E_CONTRACT=PASS
```

The full browser runner additionally emits:

```text
PLAYWRIGHT_E2E_BROWSER=chromium
PLAYWRIGHT_E2E_PUBLIC_NETWORK=DISABLED
PLAYWRIGHT_E2E=PASS
```

Failure evidence retains trace, screenshot, video, HTML report, and per-test output.

## 7. Safety and Scope Boundary

The public Render deployment is never targeted by this suite. No production mutation route,
credential, real provider account, or external aviation service is required.

No fixture asserts confirmed emergencies, filed flight plans, navigation guidance, air
traffic control suitability, or authoritative weather truth. Nullable measurements stay
nullable; provenance and limitations remain visible; the protected Route Intelligence
mutation is still absent from the mock contract.
