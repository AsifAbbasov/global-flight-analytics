# Backend Timeout Consistency Audit Closure

Status: Source hardening prepared; exact-commit Continuous Integration evidence is required after commit
Project: Global Flight Analytics
Reviewed baseline: `ba405a6dc782cacf1d26d919374e460602a9368f`
Date: 2026-08-02

## 1. Purpose

This increment closes the verified timeout-consistency gaps at production input and outbound I/O boundaries. It does not add a global database statement timeout or impose one duration on every workload. Timeout ownership stays with the boundary that knows the operation budget.

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

Formal exact-commit closure requires Backend CI success for the commit containing this document.
