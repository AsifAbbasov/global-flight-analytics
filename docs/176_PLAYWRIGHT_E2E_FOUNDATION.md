# Document 176 — Playwright End-to-End Testing Foundation

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: deterministic browser testing for the public product shell

## 1. Purpose

This increment establishes a reproducible Playwright end-to-end boundary on top of the
source-backed OpenAPI 3.1 contract introduced by Document 175.

The suite exercises the real production Next.js build and browser client. It does not use
the public Render API, Vercel Preview, or live provider data.

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

The mock API implements exactly the eight paths currently documented in:

```text
openapi/openapi.json
```

A private test control endpoint selects deterministic success and failure scenarios. It is
not part of the production API contract.

## 3. Covered User Scenarios

The foundation contains four Chromium end-to-end scenarios:

1. the server-rendered application shell publishes a healthy initial aircraft snapshot;
2. selecting Azerbaijan updates both the visible workspace and canonical shareable URL;
3. the responsive mobile navigation keeps major product sections reachable;
4. a traffic API outage preserves the product shell and recovers through the visible Retry path.

Tests use roles, accessible names, form labels, and visible status text. They do not depend
on generated Tailwind class names or test-only attributes in production components.

The URL-state scenario begins with a deliberately non-canonical query and waits for the client
to replace it with the canonical workspace URL before interacting. This is the deterministic
hydration boundary; the test no longer races React hydration.

Continuous Integration sets `failOnFlakyTests: true` through the `CI` environment. A scenario
that passes only after retry is therefore a failed workflow, not accepted evidence.

## 4. Isolated Playwright Runtime

The runner installs exactly:

```text
@playwright/test@1.62.0
```

Installation occurs under:

```text
apps/web/e2e/node_modules
```

The command uses `--no-save` and `--package-lock=false`. The repository pnpm lockfile and the
application dependency graph are not modified.

Chromium is installed explicitly. Continuous Integration uses:

```text
playwright install --with-deps chromium
```

## 5. Commands

Static and source-alignment verification:

```bash
pnpm run test:playwright-e2e-contract
pnpm run verify:playwright-e2e
```

Full browser suite:

```bash
pnpm run run:playwright-e2e
```

The full browser suite is intentionally not part of the general release script because it
downloads a browser runtime and starts two local services. It runs in the dedicated
`Playwright E2E` GitHub Actions workflow.

## 6. Evidence

Failure evidence retains:

- Playwright trace;
- screenshot;
- video;
- HTML report;
- per-test output directory.

GitHub Actions uploads the report and test-results directories for fourteen days.

Required success markers:

```text
PLAYWRIGHT_E2E_VERSION=1.62.0
PLAYWRIGHT_E2E_BROWSER=chromium
PLAYWRIGHT_E2E_PUBLIC_NETWORK=DISABLED
PLAYWRIGHT_E2E=PASS
```

## 7. Safety and Scope Boundary

The public Render deployment is never targeted by this suite. No production mutation route,
credential, real provider account, or external aviation service is required.

This foundation covers the application shell and live traffic recovery boundary. Airport,
historical, projection, weather, airspace, and selected-aircraft workflows should be added
in separate reviewed increments after their fixtures and product assertions are stable.
