# Document 93 — Server and HTTP Protection Blocker Closure

Status: Implemented Engineering Contract v1.1
Project: Global Flight Analytics
Baseline: `6b922cbd9df1bff3f880ad120dd883b37f658e53`

## 1. Scope

This increment closes the two accepted release blockers from the Server and
HTTP Protection review:

```text
1. state-changing Route Intelligence POST route required authentication;
2. process liveness was incorrectly used as production readiness while
   PostgreSQL and database-backed routes could be unavailable.
```

This document does not claim that every non-blocking observation from the
original Server review was closed by the same increment. Complete classification
and closure are recorded by Document 94.

## 2. Mutation endpoint protection

The state-changing Route Intelligence route remains protected by the internal
mutation authorization middleware before the request reaches the handler.

The established protection contract remains:

```text
POST /api/v1/trajectories/:id/route-intelligence
X-Internal-API-Key: <raw high-entropy operator key>
```

The backend stores only the configured SHA-256 digest and compares the presented
credential through the existing constant-time authorization boundary.

The Backend Container gate sends the same POST request without credentials and
requires `HTTP 401`. This proves that route registration cannot bypass the
middleware.

## 3. Liveness and readiness separation

The service exposes separate contracts:

```text
GET /api/v1/health
```

This is process liveness only. It proves that the Hypertext Transfer Protocol
process can answer requests.

```text
GET /api/v1/ready
```

This is production readiness. It succeeds only when PostgreSQL is configured
and responds to a bounded ping.

A nil database pool is converted to a nil readiness function before crossing
the handler boundary. This prevents the typed-nil interface failure mode.

Failure is fail-closed:

```text
HTTP 503
SERVICE_NOT_READY
```

Public responses do not expose PostgreSQL driver errors, network addresses,
connection strings, or credentials.

## 4. Container contract

The compiled container healthcheck targets `/api/v1/ready`.

The Backend Container Continuous Integration job:

```text
creates an isolated Docker network;
starts PostgreSQL 16;
waits for pg_isready;
applies the complete production migration catalog;
starts the API with every required database-backed configuration value;
waits for Docker readiness;
verifies /health;
verifies /ready;
verifies that the mutation POST route rejects a missing key with HTTP 401.
```

## 5. Blocker closure statement

The commit containing this increment passed complete Backend Continuous
Integration.

```text
Server review release blockers: CLOSED
Open release blockers: 0
Server review full closure: tracked by Document 94
Release decision: ACCEPTABLE
```

---

## Canonical remediation history

### Existing mutation-auth finding ownership

The state-changing Route Intelligence authentication observation is **not** assigned a new canonical finding ID here. The runtime mutation authorization boundary was already introduced and registered as `GFA-SEC-039` in Document 45. Historical commit `1fc925c91117eebbb7c90c4bd6b3889548d55cb4` adds container-level regression evidence that the concrete Route Intelligence POST registration still returns `HTTP 401` without credentials; it does not introduce a second independent authorization remediation. This document therefore preserves later verification evidence without duplicating canonical ownership.

### GFA-OPS-089 — process liveness was used as production readiness while PostgreSQL could be unavailable

1. **Finding / symptom.** The container/service health contract used `/api/v1/health` to represent readiness even though that endpoint proved only that the HTTP process could answer requests.
2. **Root cause.** Process liveness and dependency readiness were modeled as one operational signal; PostgreSQL availability was not part of the container health decision.
3. **Failure scenario.** The API process is alive while PostgreSQL is missing, unavailable or not yet ready; container/platform health reports success even though database-backed routes cannot serve production work.
4. **Impact.** Deployment orchestration and operators can route traffic to an unusable instance, mask startup/database failures and misclassify service availability.
5. **Severity rationale.** **P1 retrospective.** This is a production availability correctness defect at the service admission boundary: the platform could advertise a non-functional API as ready.
6. **Existing guarantees violated.** Liveness must answer only process survival; readiness must fail closed when required production dependencies are unavailable; public readiness failure must not expose internal database details.
7. **Considered solutions.** Keep one `/health` endpoint; make `/health` database-aware; add a separate `/ready` dependency probe; rely on container startup delays instead of runtime probing.
8. **Chosen remediation.** Preserve `/health` as liveness, add `/ready` with a bounded PostgreSQL ping, return `503 SERVICE_NOT_READY` on absent/failing database, and point the compiled container healthcheck plus Backend Container gate at `/ready`.
9. **Why this solution was selected.** Separate probes preserve standard lifecycle semantics, keep liveness usable during dependency incidents and give deployment systems a fail-closed production-readiness signal.
10. **Rejected alternatives.** Making liveness dependency-aware would cause process restarts for database incidents; fixed delays cannot prove current database availability; retaining one liveness-only probe preserves the false-ready state.
11. **Trade-offs.** Readiness now performs a bounded database ping and can temporarily remove a live process from service during PostgreSQL outages; that is intentional because database-backed routes are part of the production contract.
12. **Regression tests / protection.** Readiness handler tests cover ready, absent database and failed ping; server route tests preserve liveness/readiness separation; container CI starts PostgreSQL, applies the production catalog, verifies `/ready`, and uses `/ready` for Docker health.
13. **Adversarial review findings.** A nil `*pgxpool.Pool` must be converted before crossing the function-interface boundary to avoid typed-nil behavior; database errors must map to stable public `SERVICE_NOT_READY` rather than exposing driver/network details.
14. **Remediation iterations.** The first server blocker closure added dependency-aware readiness in `1fc925c9…`; later container-readiness work in Document 95 strengthened evidence around PostgreSQL's bootstrap-to-final-process handoff.
15. **Residual risks and limitations.** Readiness proves a bounded PostgreSQL ping, not every downstream provider or external dependency. Those dependencies have their own health/evidence contracts and are not folded into this probe.
16. **Operational or deployment consequences.** Container/platform health should target `/api/v1/ready`; `/api/v1/health` remains appropriate for pure process-liveness diagnosis.
17. **Exact evidence.** Historical implementation commit `1fc925c91117eebbb7c90c4bd6b3889548d55cb4` (`fix: close server and http protection review`). Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-OPS-089=CLOSED`.
19. **Prevention / future guard.** New required production dependencies must be evaluated explicitly against readiness semantics; liveness may not be silently reused as dependency readiness, and container tests must prove the selected health endpoint against real dependency setup.
