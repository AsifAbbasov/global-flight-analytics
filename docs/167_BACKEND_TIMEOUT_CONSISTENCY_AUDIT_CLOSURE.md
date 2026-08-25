# Backend Timeout Consistency Audit Closure

Status: CLOSED — canonical remediation history reconciled  
Project: Global Flight Analytics  
Reviewed baseline: `ba405a6dc782cacf1d26d919374e460602a9368f`  
Implementation and permanent audit: `76e8744513bc1ad7d185692ea9ad3222bc961e0c`  
Exact historical Backend CI: run `30724291596` — SUCCESS  
Date: 2026-08-02

## 1. Purpose

This increment closes the verified timeout-consistency gaps at production input and outbound I/O boundaries. It does not add a global database statement timeout or impose one duration on every workload. Timeout ownership stays with the boundary that knows the operation budget.

The original review produced two independent correctness findings. They are canonicalized below as `GFA-OPS-439` and `GFA-REL-440`; the other inspected timeout boundaries were already correct and do not receive synthetic finding IDs.

## 2. Confirmed findings

### Finding A — HTTP service work had no request-owned deadline

Application handlers passed `fiber.Ctx.Context()` into services and repositories. That object represents the fasthttp transport context; it was not the request-operation context configured by this application. Consequently PostgreSQL reads, analytical work and external-provider calls could receive no application-owned deadline even though socket read and write timeouts existed.

Classification: Required change.

Correction:

- add `API_REQUEST_TIMEOUT` with a positive default;
- create a request-timeout middleware using `context.WithTimeout`;
- store the bounded context through `fiber.Ctx.SetUserContext`;
- make every service-bearing HTTP handler call `fiber.Ctx.UserContext`;
- preserve a shorter upstream deadline automatically;
- return HTTP 504 after an application request deadline and HTTP 408 after caller cancellation.

### Finding B — custom provider HTTP clients could be unbounded or mutated after validation

OpenSky required a non-nil `http.Client` but retained caller-owned timeout state and allowed a zero timeout. Open-Meteo preserved injected clients by contract, but an injected client with `Timeout == 0` had no independent request budget.

Classification: Required change.

Correction:

- OpenSky clones the supplied client and applies a fifteen-second fallback when its timeout is non-positive;
- caller mutation cannot change the OpenSky timeout contract after construction;
- Open-Meteo preserves its established injected-client identity contract;
- Open-Meteo owns a separate positive request deadline derived from explicit configuration, the injected client timeout, or a ten-second safe fallback;
- every Open-Meteo request uses `context.WithTimeout` without mutating the injected client.

## 3. Boundaries already correct

The audit confirmed that the following boundaries already own positive limits or caller-cancelable waits:

- Airplanes.live HTTP requests;
- OurAirports CSV downloads;
- PostgreSQL pool connection and ping;
- migration execution;
- historical materialization operations;
- readiness database probes;
- server graceful shutdown;
- ingest daemon retry waits and bounded exponential backoff;
- production API smoke requests.

These are closure evidence, not additional findings.

## 4. Deliberate database policy

A single global PostgreSQL `statement_timeout` is not introduced. HTTP operations now carry an application request deadline into repositories. CLI migrations and materialization commands already own longer operation-specific contexts. This avoids applying an interactive HTTP budget to legitimate administrative workloads while still preventing unbounded public request execution.

## 5. Permanent audit gate

The permanent tool is:

```text
apps/api/tools/backendtimeoutconsistencyaudit
```

It scans production Go files under `cmd`, `internal` and `tools` and rejects:

- package-level `http.Get`, `http.Head`, `http.Post` and `http.PostForm` calls;
- `http.NewRequest` without a caller context;
- `http.DefaultClient` and `http.DefaultTransport` ownership;
- `http.Client` literals without an explicit timeout;
- literal non-positive HTTP client timeouts;
- `fiber.Ctx.Context()` use inside application HTTP handlers.

The gate accepts `http.NewRequestWithContext`, bounded client literals and `fiber.Ctx.UserContext` propagation. Test files, vendored code and testdata are excluded.

Permanent commands:

```bash
go test -count=1 ./tools/backendtimeoutconsistencyaudit
go run ./tools/backendtimeoutconsistencyaudit -strict
```

Backend Quality and the local release verifier run the strict audit.

## 6. Regression evidence

The increment adds tests for:

- invalid request-timeout configuration;
- request deadline installation;
- preservation of an earlier parent deadline;
- HTTP 504 after deadline expiration;
- OpenSky fallback bounding of unbounded custom clients;
- OpenSky ownership of a cloned bounded client;
- Open-Meteo preservation of injected client identity;
- Open-Meteo request-deadline fallback without caller mutation;
- audit detection of every prohibited HTTP and handler-context pattern.

## 7. Configuration

Default public request budget:

```text
API_REQUEST_TIMEOUT=12s
```

The value is present in the environment example, Docker Compose and Render Blueprint. Provider and administrative operation timeouts remain separately configurable.

## 8. Closure state

```text
HTTP_REQUEST_DEADLINE_PROPAGATION=CLOSED
HANDLER_USER_CONTEXT_PROPAGATION=CLOSED
OPENSKY_CUSTOM_CLIENT_TIMEOUT=CLOSED
OPEN_METEO_CUSTOM_CLIENT_TIMEOUT=CLOSED
TIMEOUT_AUDIT_GATE=INSTALLED
DEFERRED_TIMEOUT_FINDINGS=0
UNCLASSIFIED_TIMEOUT_FINDINGS=0
```

Exact historical closure evidence exists for implementation commit `76e8744513bc1ad7d185692ea9ad3222bc961e0c`: Backend CI run `30724291596` completed successfully, including Backend Quality job `91433041929`, Backend Race Safety `91433041892`, PostgreSQL 16 Integration `91433041870`, and Backend Container `91433122150`.

---

## 9. Canonical remediation record — GFA-OPS-439

### 1. Finding / symptom

Public HTTP service work had no application-owned request deadline.

### 2. Root cause

The HTTP layer forwarded `fiber.Ctx.Context()` as if it were the application's operation context. Socket read/write limits existed, but the application had no separately configured request-work budget installed into Fiber user context and propagated through service/repository calls.

### 3. Failure scenario

A public request could enter PostgreSQL, analytical work, or an outbound provider call and remain active beyond the intended interactive request budget. The transport-level socket limits did not provide a single application-owned cancellation deadline for that downstream work.

### 4. Impact

Slow or stalled public operations could retain application resources longer than intended, amplify tail latency, and reduce availability under concurrent load.

### 5. Severity rationale

**P2 retrospective.** The defect affected bounded execution and service availability. Repository evidence does not establish a data-integrity failure or an incident requiring P1 classification, and no historical severity label was recorded at remediation time.

### 6. Existing guarantees violated

- public request work must be bounded by an explicit operation budget;
- downstream services and repositories must receive the request-owned cancellation context;
- timeout ownership must belong to the boundary that knows the interactive budget.

### 7. Considered solutions

- rely on server socket read/write timeouts;
- introduce one global PostgreSQL `statement_timeout`;
- add ad-hoc timeouts independently in handlers/services;
- install one application request deadline at the HTTP boundary and propagate it through `UserContext`.

### 8. Chosen remediation

Add positive `API_REQUEST_TIMEOUT` configuration, install `context.WithTimeout` middleware, store the bounded context with `fiber.Ctx.SetUserContext`, require service-bearing handlers to use `fiber.Ctx.UserContext`, preserve a shorter upstream deadline, and map application deadline/caller cancellation to the documented HTTP responses.

### 9. Why this solution was selected

It establishes one explicit ownership point for the public interactive budget while preserving ordinary Go context propagation and allowing longer administrative/CLI operations to retain their own budgets.

### 10. Rejected alternatives

- socket timeouts were rejected as an operation-budget substitute;
- a global database statement timeout was rejected because it would incorrectly impose the interactive budget on migrations/materialization;
- scattered handler-specific deadlines were rejected because they would duplicate policy and invite drift.

### 11. Trade-offs

The default `12s` request budget is policy that may need operational tuning. Downstream code must remain context-aware for cancellation to be effective, and administrative commands intentionally keep separate timeout policies.

### 12. Regression tests / protection

Tests cover invalid request-timeout configuration, deadline installation, preservation of an earlier parent deadline, and HTTP 504 behavior after application deadline expiration. The permanent timeout-consistency audit rejects `fiber.Ctx.Context()` in application handlers.

### 13. Adversarial review findings

The review explicitly distinguished transport socket limits from application request deadlines and rejected applying one database timeout to every workload. It also verified already-bounded administrative/provider paths instead of creating findings for them.

### 14. Remediation iterations

The accepted remediation and permanent audit were committed together in `76e8744513bc1ad7d185692ea9ad3222bc961e0c`.

### 15. Residual risks and limitations

The middleware cannot force cancellation inside code that ignores the propagated context. Timeout values remain operational policy rather than a proof that every request completes within the budget.

### 16. Operational or deployment consequences

Deployments must carry a valid positive `API_REQUEST_TIMEOUT`; the documented default is `12s`. Public timeout behavior is now observable as an application deadline rather than depending only on transport timing.

### 17. Exact evidence

- reviewed baseline: `ba405a6dc782cacf1d26d919374e460602a9368f`;
- implementation/audit: `76e8744513bc1ad7d185692ea9ad3222bc961e0c`;
- Backend CI: run `30724291596` — SUCCESS;
- Backend Quality: job `91433041929` — SUCCESS;
- Backend Race Safety: job `91433041892` — SUCCESS;
- PostgreSQL 16 Integration: job `91433041870` — SUCCESS;
- Backend Container: job `91433122150` — SUCCESS.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

`apps/api/tools/backendtimeoutconsistencyaudit` remains in Backend Quality and rejects direct Fiber transport-context use at service-bearing HTTP boundaries plus other prohibited unbounded HTTP patterns.

---

## 10. Canonical remediation record — GFA-REL-440

### 1. Finding / symptom

Custom provider HTTP clients could remain unbounded or have their validated timeout contract changed by caller mutation.

### 2. Root cause

OpenSky accepted a caller-owned `http.Client` without requiring a positive timeout and retained that mutable object. Open-Meteo intentionally preserved injected-client identity, but a client with `Timeout == 0` had no independent request deadline owned by the provider boundary.

### 3. Failure scenario

A custom provider client could execute an outbound request without a positive time budget, or OpenSky behavior could change after construction if the caller mutated the retained client's timeout. A stalled provider call could therefore outlive the intended provider-operation budget.

### 4. Impact

Provider requests could retain connections/goroutines and delay request or ingestion completion, degrading reliability and availability when an upstream service stalled.

### 5. Severity rationale

**P2 retrospective.** The failure mode is bounded-execution/reliability risk at external I/O boundaries. No historical severity was recorded and repository evidence does not show corruption of persisted analytical evidence.

### 6. Existing guarantees violated

- every outbound provider operation must have a positive bounded execution budget;
- validated constructor policy must not be silently changed through shared mutable client state;
- injected-client compatibility must not remove request-level timeout ownership.

### 7. Considered solutions

- require callers to configure every `http.Client` correctly;
- mutate injected clients during construction;
- clone the OpenSky client and own its timeout while preserving Open-Meteo client identity with a request-scoped deadline;
- use a single shared global default client.

### 8. Chosen remediation

OpenSky clones the supplied client and applies a `15s` fallback when its timeout is non-positive. Open-Meteo preserves its injected-client identity contract but computes a positive provider request deadline from explicit configuration, client timeout, or a `10s` fallback and applies it with `context.WithTimeout` per request.

### 9. Why this solution was selected

It closes unbounded I/O without breaking Open-Meteo's established injected-client identity contract and prevents post-construction caller mutation from changing OpenSky's validated timeout policy.

### 10. Rejected alternatives

- trusting callers was rejected because zero-timeout clients were explicitly allowed by the old boundary;
- mutating the Open-Meteo injected client was rejected because identity/ownership was an established contract;
- global default-client ownership was rejected by the timeout audit policy.

### 11. Trade-offs

OpenSky no longer shares later client-timeout mutations with its caller because it owns a clone. Open-Meteo retains shared client identity, so non-timeout mutable client properties remain subject to its existing injection contract; request deadlines are owned independently.

### 12. Regression tests / protection

Tests cover OpenSky fallback bounding, cloned-client ownership, Open-Meteo injected-client identity, and Open-Meteo request-deadline fallback without caller mutation. The permanent audit rejects unbounded client literals and default-client ownership patterns.

### 13. Adversarial review findings

The review deliberately used different remediation mechanics for OpenSky and Open-Meteo because their client-ownership contracts differ. It avoided a mechanical 'clone every client' rule that would have broken the Open-Meteo injection contract.

### 14. Remediation iterations

The provider timeout corrections and permanent timeout-consistency audit landed in `76e8744513bc1ad7d185692ea9ad3222bc961e0c`.

### 15. Residual risks and limitations

Timeouts bound elapsed request execution but do not guarantee provider success. Provider-specific budgets remain configuration/policy and must be revisited if upstream behavior changes materially.

### 16. Operational or deployment consequences

OpenSky receives a safe `15s` fallback for non-positive custom-client timeouts. Open-Meteo receives a positive request deadline with a `10s` fallback when neither explicit configuration nor client timeout supplies one.

### 17. Exact evidence

- reviewed baseline: `ba405a6dc782cacf1d26d919374e460602a9368f`;
- implementation/audit: `76e8744513bc1ad7d185692ea9ad3222bc961e0c`;
- Backend CI: run `30724291596` — SUCCESS;
- Backend Quality: job `91433041929` — SUCCESS;
- Backend Race Safety: job `91433041892` — SUCCESS;
- PostgreSQL 16 Integration: job `91433041870` — SUCCESS;
- Backend Container: job `91433122150` — SUCCESS.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

The strict backend timeout-consistency audit remains part of Backend Quality and rejects package-level default HTTP operations, requests without caller context, default client/transport ownership, client literals without explicit timeout, and literal non-positive client timeout values.
