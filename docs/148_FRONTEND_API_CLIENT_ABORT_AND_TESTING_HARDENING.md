# Frontend API Client Abort and Testing Hardening

Status: Implementation prepared; exact-commit Continuous Integration closure pending
Project: Global Flight Analytics
Reviewed baseline: `44d41d6ece32b8196a5613c8d4839d7ebe972cac`
Date: 2026-08-01

## 1. Purpose

This increment corrects one concrete frontend API-client cancellation defect and
adds the first automated frontend contract test boundary without introducing a new
test framework or production dependency.

## 2. Confirmed defect

`requestAPIData` previously used one mutable `timedOut` boolean. When the caller
cancelled first but the mocked or real fetch operation settled after the local
timeout callback, the timeout callback changed `timedOut` to true even though the
request controller was already aborted by the caller. The catch block then reported
`The API request timed out.` instead of preserving the caller-owned `AbortError`.

The regression is deterministic when caller cancellation happens first, timeout
expires second, and fetch rejection settles last.

## 3. Correction

The API client now records the first abort cause only:

```text
caller cancellation -> caller
local request deadline -> timeout
```

A later abort attempt cannot overwrite the original cause after the shared request
controller is already aborted. Caller cancellation remains observable as
`AbortError`; only a deadline owned by `requestAPIData` becomes
`APIRequestError("The API request timed out.")`.

## 4. Frontend test boundary

The increment adds a dependency-free Node test harness that compiles only the API
client into an isolated `.test-dist` directory and executes six contract tests:

1. successful envelope parsing and search-parameter preservation;
2. fail-fast invalid timeout validation;
3. caller cancellation preserved after the timeout deadline passes;
4. local deadline reported as timeout;
5. non-JSON response rejection;
6. production API base URL requirement.

The temporary compilation directory is cleaned by the package script and ignored by
Git as a defensive fallback.

## 5. Continuous Integration

Frontend Continuous Integration now runs:

```text
pnpm --filter web test
```

The test step runs after TypeScript validation and before the production build.
Existing dependency policy, production dependency audit, lint, type checking, and
build gates remain unchanged.

## 6. Scope boundary

This increment deliberately does not add:

- Playwright, Cypress, Vitest, Jest, or React Testing Library;
- browser end-to-end automation;
- Prometheus or OpenTelemetry;
- an API gateway;
- Redis, a message queue, or database sharding;
- a new production dependency or lockfile change.

Those items require separate evidence of need and are not consequences of this
API-client correctness defect.

## 7. Closure requirements

Formal closure requires all installer markers, the six frontend contract tests,
ESLint, TypeScript validation, production frontend build, dependency policy gates,
and the exact post-commit Frontend Continuous Integration run to pass.
