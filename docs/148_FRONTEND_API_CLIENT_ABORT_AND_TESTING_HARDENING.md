# Frontend API Client Abort and Testing Hardening

Status: CLOSED
Project: Global Flight Analytics
Reviewed baseline: `44d41d6ece32b8196a5613c8d4839d7ebe972cac`
Engineering commit: `ee2bb13e60c29ae9ecdcb7736d4fe39561e3b28d`
Exact Frontend CI run: `30705944365`
Date: 2026-08-01
Canonical reconciliation: 2026-08-25

## 1. Purpose

This increment corrected one concrete frontend API-client cancellation defect and
added the first automated frontend contract-test boundary without introducing a new
test framework or production dependency.

## 2. Canonical finding

### GFA-CONTRACT-437 — Frontend API client could overwrite caller-cancellation ownership with a later local timeout

1. **Finding / symptom**

   `requestAPIData` used one mutable `timedOut` boolean to classify request aborts.
   When caller cancellation happened first but fetch rejection settled only after the
   local request timeout callback fired, the later timeout callback changed
   `timedOut` to `true`. The catch path then reported `The API request timed out.`
   instead of preserving the caller-owned `AbortError`.

2. **Root cause**

   Abort state recorded only whether the local timer had ever fired, not which owner
   won the abort race. Caller cancellation and the API client's own deadline shared
   one `AbortController` without immutable first-cause ownership.

3. **Failure scenario**

   The caller aborts the request. The shared request controller becomes aborted, but
   the underlying or mocked fetch promise has not rejected yet. The local timeout
   callback fires afterward and marks the mutable `timedOut` flag. Fetch then rejects
   with `AbortError`; the catch path sees `timedOut === true` and misclassifies the
   already caller-cancelled request as a locally timed-out request.

4. **Impact**

   The frontend can expose the wrong failure semantics to query consumers and UI
   recovery logic. Caller cancellation becomes indistinguishable from an API-owned
   deadline, which can produce misleading timeout messaging and incorrect higher-level
   handling. The reviewed evidence does not show server-side data corruption or a
   security-boundary failure.

5. **Severity rationale**

   **P2 retrospective.** This is a deterministic request-lifecycle contract defect in
   the shared API client. It affects all consumers using caller cancellation and local
   deadlines, but its demonstrated impact is incorrect cancellation/error semantics
   rather than irreversible data loss or authorization bypass.

6. **Existing guarantees violated**

   - caller-owned cancellation must remain observable as caller cancellation;
   - locally owned timeout semantics must be attributed only to the local deadline;
   - a later event must not overwrite the first abort owner;
   - shared API-client error classification must be deterministic.

7. **Considered solutions**

   - retain the mutable `timedOut` flag and rely on fetch settling quickly;
   - remove local timeouts when a caller signal exists;
   - create separate controllers and race their promises externally;
   - record one immutable first abort cause on the existing shared controller.

8. **Chosen remediation**

   Introduce `RequestAbortCause = 'caller' | 'timeout'` and one `abortRequest(cause)`
   helper. The helper records a cause only if the shared request controller has not
   already been aborted, then aborts it. Caller cancellation invokes
   `abortRequest('caller')`; the local timer invokes `abortRequest('timeout')`.

9. **Why this solution was selected**

   It preserves the existing request structure, makes first-cause ownership explicit,
   prevents later events from rewriting history, requires no production dependency,
   and can be regression-tested deterministically.

10. **Rejected alternatives**

    - depending on incidental promise settlement timing;
    - treating every abort as a timeout;
    - removing caller cancellation or removing the bounded local deadline;
    - adding a test framework solely for this API-client defect.

11. **Trade-offs**

    The client now maintains one small piece of explicit abort-cause state. This is
    more code than a boolean but it encodes the actual semantic distinction and avoids
    timing-dependent attribution.

12. **Regression tests / protection**

    The engineering commit adds a dependency-free Node test harness and six contract
    tests covering:

    1. successful envelope parsing and search-parameter preservation;
    2. fail-fast invalid timeout validation;
    3. caller cancellation preserved after the timeout deadline passes;
    4. local deadline reported as timeout;
    5. non-JSON response rejection;
    6. production API base URL requirement.

    Frontend CI runs `pnpm --filter web test` after TypeScript validation and before
    the production build.

13. **Adversarial review findings**

    The critical ordering is not simply “caller cancellation vs timeout callback”; it
    is a three-event sequence: caller abort first, timeout callback second, fetch
    rejection last. The regression test intentionally delays the rejection after the
    caller-owned abort so the old misclassification is deterministic rather than
    dependent on browser scheduling luck.

14. **Remediation iterations**

    The defect and the test boundary were corrected in one engineering increment,
    `ee2bb13e60c29ae9ecdcb7736d4fe39561e3b28d`. Canonical reconciliation later
    recovered the exact historical Frontend CI run and replaced the old
    `exact-commit Continuous Integration closure pending` status with the source-backed
    closed state.

15. **Residual risks / limitations**

    The contract tests exercise the shared API client in the dependency-free Node
    harness, not a browser end-to-end environment. Browser navigation, React Query
    cancellation orchestration and product-level recovery surfaces require their own
    tests and are not claimed by this document.

16. **Operational / deployment consequences**

    No new production dependency, service, environment variable, database change or
    deployment component is introduced. The only production behavior change is
    deterministic classification of an already existing cancellation/timeout race.

17. **Exact evidence**

    Baseline:

    ```text
    44d41d6ece32b8196a5613c8d4839d7ebe972cac
    ```

    Engineering owner:

    ```text
    ee2bb13e60c29ae9ecdcb7736d4fe39561e3b28d
    fix: harden frontend API cancellation and testing
    ```

    Source diff evidence in `apps/web/lib/api/client.ts`:

    ```text
    mutable `timedOut` boolean -> explicit RequestAbortCause
    caller abort -> abortRequest('caller')
    local deadline -> abortRequest('timeout')
    later abort attempt -> ignored once requestController.signal.aborted
    timeout error mapping -> only when abortCause === 'timeout'
    ```

    Exact GitHub Actions evidence:

    ```text
    SHA: ee2bb13e60c29ae9ecdcb7736d4fe39561e3b28d
    Frontend CI: 30705944365 — SUCCESS
    Frontend Quality — SUCCESS
      dependency security policy — SUCCESS
      production dependency audit — SUCCESS
      ESLint — SUCCESS
      TypeScript validation — SUCCESS
      frontend contract tests — SUCCESS
      production frontend build — SUCCESS

    Backend CI: 30705944373 — SUCCESS
    ```

18. **Final canonical status**

    **CLOSED.** The abort-owner race is corrected, deterministic regression coverage
    is installed, and exact-commit Frontend CI is green.

19. **Prevention / future guard**

    Keep the frontend contract tests in Frontend CI; preserve first-abort-cause
    semantics when extending the shared API client; and add focused regression cases
    whenever a new independently owned deadline or cancellation source is introduced.

## 3. Correction

The API client records the first abort cause only:

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
client into an isolated `.test-dist` directory and executes the six contract tests
listed in the canonical record above.

The temporary compilation directory is cleaned by the package script and ignored by
Git as a defensive fallback.

## 5. Continuous Integration closure

Frontend Continuous Integration runs:

```text
pnpm --filter web test
```

Exact engineering-commit evidence is no longer pending:

```text
Frontend CI run 30705944365 — SUCCESS
Frontend Quality — SUCCESS
Run frontend contract tests — SUCCESS
Build production frontend — SUCCESS
```

## 6. Scope boundary

This increment deliberately did not add:

- Playwright, Cypress, Vitest, Jest, or React Testing Library;
- browser end-to-end automation;
- Prometheus or OpenTelemetry;
- an API gateway;
- Redis, a message queue, or database sharding;
- a new production dependency or lockfile change.

Those items require separate evidence of need and are not consequences of this
API-client correctness defect.

## 7. Formal closure

```text
Canonical finding: GFA-CONTRACT-437
Implementation: ee2bb13e60c29ae9ecdcb7736d4fe39561e3b28d
Exact Frontend CI: 30705944365 — SUCCESS
Open findings in this document: 0
Deferred findings in this document: 0
Unclassified findings in this document: 0
Status: CLOSED
```
