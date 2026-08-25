# Document 94 — Server Review Full Closure

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics
Baseline: `1fc925c91117eebbb7c90c4bd6b3889548d55cb4`

## 1. Purpose

This document completes classification and remediation of every observation from
the original Server and HTTP Protection review.

The closure policy follows the project Code Review Standard:

```text
fixed;
not applicable;
deliberately rejected with evidence;
deferred with owner, risk and revisit condition;
suggestion or nit that does not block merge.
```

## 2. Fixed findings

### 2.1 Process lifecycle

The command lifecycle now has one controlled `run` path and one final process
exit point.

The implementation:

```text
waits for either listener failure or context cancellation;
returns listener failures through a buffered error channel;
uses ShutdownWithTimeout with a ten-second bound;
waits for the listener to stop after shutdown;
closes PostgreSQL through a deferred cleanup path;
contains executable lifecycle tests.
```

`os.Exit` remains only in `main`, after `run` has returned and its deferred
resource cleanup has completed.

### 2.2 Request logger final status

The request logger now invokes the configured Fiber error handler before reading
the response status.

It therefore logs the status actually returned to the client rather than a
provisional status that existed before centralized error handling.

Regression tests cover both an error response and a successful response.

### 2.3 Sensitive internal error logging

The global error handler no longer records arbitrary `err.Error()` text for
internal request failures.

The log contains stable metadata:

```text
request identifier;
method;
path;
final status;
Go error type.
```

The raw error message is excluded. A regression test injects a synthetic
credential-like value and proves that it does not reach the log.

### 2.4 Historical Intelligence read interface

The read-only route registration boundary now accepts
`historicalaggregatecontract.Reader` instead of the full read/write store.

The production PostgreSQL implementation still satisfies the reader contract,
while verification stubs no longer need unrelated write methods.

### 2.5 Readiness rate-limit exclusion

`/api/v1/ready` is treated as an infrastructure route and is excluded from the
application rate limiter together with `/health` and `/version`.

A regression test repeatedly calls readiness with a one-request application
limit and proves that the response remains the dependency status rather than
becoming `HTTP 429`.

## 3. Deliberately rejected findings

### 3.1 Mechanical function-length threshold

A universal forty-line or fifty-line failure rule is rejected.

The previous database composition root was already decomposed by Document 48.
The remaining server functions are split when they mix lifecycle, route
registration, persistence ownership or independently testable policy.

This increment specifically decomposes the process lifecycle because it carried
real cleanup and failure-propagation risk. Functions are not split solely to
reduce line count.

### 3.2 Duplicate repository and service objects

Core traffic routes and Route Intelligence intentionally own separate
composition graphs.

The duplicated objects are stateless adapters and services over the same
PostgreSQL pool. They do not own independent transactions, mutable caches,
credentials or configuration that can diverge.

Route Intelligence keeps a self-contained, versioned PostgreSQL composition so
the pipeline can be verified and materialized independently. Reusing object
identity would add cross-context coupling without changing persistence
correctness.

The proposal to introduce a shared service container is therefore deliberately
rejected until a concrete shared mutable policy exists.

### 3.3 Replacing the custom rate limiter solely because of size

The local fixed-window limiter remains appropriate for the approved
single-instance MVP.

The review did not provide representative load measurements proving that the
mutex or periodic map cleanup violates a latency or throughput objective.
Replacement based only on source length is rejected.

## 4. Deferred findings

### 4.1 Trusted proxy client identity

Risk:

```text
the default limiter key uses the direct connection address;
a deployment behind a reverse proxy may require a trusted proxy contract;
blindly trusting X-Forwarded-For would permit spoofing.
```

Owner:

```text
backend deployment hardening
```

Revisit condition:

```text
before public multi-user deployment behind Render or another reverse proxy;
before horizontal scaling;
or after measured evidence that the current connection identity groups
unrelated clients.
```

Required future evidence:

```text
documented hosting proxy ranges or a platform-supported trusted proxy mode;
spoofing regression tests;
deployment smoke evidence for the selected client identity header.
```

### 4.2 Build-derived version endpoint

The hardcoded version remains a suggestion, not a correctness defect.

Owner:

```text
release engineering
```

Revisit condition:

```text
before the first tagged public release or when build provenance is exposed in
deployment diagnostics.
```

## 5. Suggestions and nits

The following observations are classified as non-blocking nits:

```text
capitalization of isolated Go error strings;
the word And in a test name;
local formatting and naming preferences without a demonstrated failure mode.
```

They may be corrected when the affected code is otherwise changed.

## 6. Verification gates

Formal full closure requires one new commit to pass:

```text
Go formatting;
cmd/server lifecycle tests;
middleware request logger tests;
server error-log redaction test;
readiness rate-limit regression test;
complete backend tests;
targeted race tests including cmd/server;
Go vet;
project architecture and contract audit;
code review policy audit;
Stage 14 final audit;
PostgreSQL 16 Integration;
Backend Container.
```

## 7. Closure statement

When every gate in Section 6 passes on the same commit:

```text
Server and HTTP Protection review: CLOSED
Open release blockers: 0
Open required changes: 0
Unclassified original findings: 0
Deferred deployment findings at the 2573892 closure baseline: 2
Release decision: ACCEPTABLE
```

Deferred findings are classified debt with explicit owners and revisit
conditions. They do not make the review unclassified or block the current MVP.

## 8. Post-closure resolution

Document 95 resolves both findings that were deferred at this closure baseline.

```text
Trusted proxy code debt: CLOSED
Build-derived version debt: CLOSED
Current deferred Server review code findings: 0
Deployment activation inputs: verified trusted proxy ranges
```

The deployment-specific proxy range remains an environment value and is not
guessed by application source code.

---

## Canonical remediation history

The rejected mechanical observations in Section 3 and style nits in Section 5 are not assigned canonical finding IDs because the review itself records no demonstrated failure mode. The two deferred items from Section 4 are assigned canonical ownership in Document 95, where they were actually remediated.

### GFA-OPS-090 — server process lifecycle did not have one controlled shutdown/error-propagation path

1. **Finding / symptom.** Server startup/listener/shutdown behavior was not owned by one testable `run` lifecycle with deferred resource cleanup and one final process-exit point.
2. **Root cause.** Process control, listener failure and cleanup responsibilities were coupled around `main`/exit behavior rather than an ordinary return-based lifecycle.
3. **Failure scenario.** Listener failure or cancellation reaches a path that exits before deferred PostgreSQL cleanup, fails to wait for listener shutdown, or loses the original listener error.
4. **Impact.** Shutdown can become non-deterministic, resource cleanup can be skipped, and operational failures become harder to test and diagnose.
5. **Severity rationale.** **P2 retrospective.** This is a process reliability/resource-ownership defect; it can affect clean shutdown but does not directly corrupt analytical data.
6. **Existing guarantees violated.** Resource-owning code must return through deferred cleanup; graceful shutdown must be bounded; listener failures must propagate to the final process decision.
7. **Considered solutions.** Keep exit calls throughout lifecycle code; rely on Fiber defaults; move lifecycle into a single `run` function with bounded shutdown and one `os.Exit` after return.
8. **Chosen remediation.** Introduce one controlled `run` path that waits for listener failure or cancellation, uses ten-second `ShutdownWithTimeout`, waits for listener stop, defers PostgreSQL cleanup, and leaves `os.Exit` only in `main` after `run` returns.
9. **Why this solution was selected.** Ordinary return-based control makes cleanup testable and guarantees defer execution without introducing a new process supervisor abstraction.
10. **Rejected alternatives.** Scattered `os.Exit` skips defers; unbounded shutdown can hang termination; hidden framework lifecycle behavior is weaker evidence than explicit ownership.
11. **Trade-offs.** Lifecycle code is slightly more explicit and requires listener error-channel coordination.
12. **Regression tests / protection.** `cmd/server` lifecycle tests, full backend tests, targeted race coverage, Backend Container and code-review/source audits.
13. **Adversarial review findings.** Listener error propagation must not deadlock, shutdown must have a finite upper bound, and PostgreSQL cleanup must execute before the final exit code is applied.
14. **Remediation iterations.** The review rejected mechanical decomposition by line count and instead decomposed only the lifecycle responsibility with real cleanup/failure risk.
15. **Residual risks and limitations.** Operating-system forced termination can still bypass graceful cleanup; that is outside application control and must be handled by durable database semantics.
16. **Operational or deployment consequences.** Normal SIGTERM/cancellation receives a bounded graceful shutdown path and clearer failure exit behavior.
17. **Exact evidence.** Historical implementation commit `2573892ad7684f3d2646378e2021638a53173bc3` (`fix: fully close server review findings`). Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-OPS-090=CLOSED`.
19. **Prevention / future guard.** Process entrypoints must keep `os.Exit` outside resource-owning functions and retain bounded shutdown plus lifecycle/race tests.

### GFA-OBS-091 — request logging could capture a provisional status before centralized error handling

1. **Finding / symptom.** Request logs could record the response status before Fiber's configured error handler converted the returned error into the actual client response.
2. **Root cause.** Logging observed status too early in middleware control flow.
3. **Failure scenario.** A handler returns an error, the logger records the pre-error-handler status, then centralized handling emits a different final HTTP status to the client.
4. **Impact.** Operational logs disagree with client-visible behavior, degrading incident diagnosis, request metrics and auditability.
5. **Severity rationale.** **P2 retrospective.** Observability evidence becomes false for failed requests, but request execution itself remains correct.
6. **Existing guarantees violated.** Logged final status must equal the actual status published to the client.
7. **Considered solutions.** Log immediately after handler return; infer status from error type; invoke the configured error handler before reading final response status.
8. **Chosen remediation.** The request logger invokes the configured Fiber error handler first and records the response status only after centralized error handling has finalized it.
9. **Why this solution was selected.** It reuses the same response owner as production instead of duplicating status-mapping logic inside logging middleware.
10. **Rejected alternatives.** Error-type inference can drift from the real error handler; early logging preserves the mismatch.
11. **Trade-offs.** Logger middleware is more tightly coupled to the configured error-handling lifecycle, which is appropriate because it reports the final result.
12. **Regression tests / protection.** Middleware tests cover both failed and successful responses and compare logged status to the actual response.
13. **Adversarial review findings.** Successful paths must not be double-written; error handling must happen exactly once before logging final status.
14. **Remediation iterations.** The fix treats final status as output evidence, not an intermediate middleware state.
15. **Residual risks and limitations.** Logs can still be lost by external log transport; this finding concerns correctness of emitted application log content.
16. **Operational or deployment consequences.** Status-based dashboards/incident review can trust request logs to match client responses.
17. **Exact evidence.** Historical implementation commit `2573892ad7684f3d2646378e2021638a53173bc3`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-OBS-091=CLOSED`.
19. **Prevention / future guard.** Request telemetry must be emitted after the response owner finalizes status; tests must cover centralized-error paths.

### GFA-SEC-092 — raw internal error messages could be written to server logs

1. **Finding / symptom.** The global error handler logged arbitrary `err.Error()` text for internal request failures.
2. **Root cause.** Diagnostic convenience treated untrusted/internal error strings as safe log payloads without a stable redaction boundary.
3. **Failure scenario.** An internal error contains a credential-like token, database/network address or other sensitive runtime detail and the global handler writes the raw text to logs.
4. **Impact.** Secrets or sensitive infrastructure details can leak into persistent/third-party logging systems.
5. **Severity rationale.** **P1 retrospective.** This is a direct sensitive-information disclosure path at a centralized production error boundary.
6. **Existing guarantees violated.** Public/internal failure handling must not expose arbitrary error details; logs should use stable metadata and controlled classifications.
7. **Considered solutions.** Keep raw errors; redact common token patterns; log only request/status plus Go error type; maintain an allow-list of safe typed details.
8. **Chosen remediation.** Remove arbitrary error text from the global handler and log stable request ID, method, path, final status and Go error type.
9. **Why this solution was selected.** Metadata/type retains enough diagnostic classification without attempting brittle pattern-based secret detection.
10. **Rejected alternatives.** Pattern redaction is incomplete and future secret formats can bypass it; raw logging preserves the disclosure risk.
11. **Trade-offs.** Some detailed root-cause text is no longer available in generic request logs and must be diagnosed through safe typed/context-specific telemetry.
12. **Regression tests / protection.** A synthetic credential-like error is injected and asserted absent from logs; server error-path tests remain in Backend CI.
13. **Adversarial review findings.** The protection must apply to arbitrary internal errors, not only known credential formats; logging the concrete Go error type is accepted as non-secret classification evidence.
14. **Remediation iterations.** Stable metadata replaced content redaction rather than growing a fragile list of prohibited substrings.
15. **Residual risks and limitations.** Other component-specific logs must independently avoid raw secrets; this fix owns the centralized server error handler only.
16. **Operational or deployment consequences.** Generic request logs are safer for centralized collection; deeper debugging relies on structured subsystem evidence rather than raw exception strings.
17. **Exact evidence.** Historical implementation commit `2573892ad7684f3d2646378e2021638a53173bc3`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-SEC-092=CLOSED`.
19. **Prevention / future guard.** Central error/logging middleware must use stable structured fields; regression tests must inject secret-like text and prove it is not emitted.

### GFA-ARCH-093 — read-only Historical HTTP registration depended on a read/write store contract

1. **Finding / symptom.** A read-only Historical Intelligence route boundary accepted the full read/write aggregate store, forcing unrelated write capability into HTTP composition and test stubs.
2. **Root cause.** Interface ownership followed a broad implementation type rather than the behavior actually required by the route.
3. **Failure scenario.** Read-only handlers/tests gain unnecessary write methods or future write-side changes create coupling/recompilation pressure at the HTTP boundary.
4. **Impact.** Broader dependency surface, weaker least-capability architecture and harder isolated verification; no immediate data corruption was demonstrated.
5. **Severity rationale.** **P3 retrospective.** This is a maintainability/dependency-boundary finding with behavior already correct.
6. **Existing guarantees violated.** Read-only composition should depend on the narrow reader behavior it consumes.
7. **Considered solutions.** Keep full store; create a duplicate HTTP repository abstraction; use the existing `historicalaggregatecontract.Reader` narrow interface.
8. **Chosen remediation.** Route registration now accepts `historicalaggregatecontract.Reader`; PostgreSQL still satisfies it and test stubs no longer implement unrelated writes.
9. **Why this solution was selected.** It improves least-capability ownership without adding an adapter or changing persistence behavior.
10. **Rejected alternatives.** Keeping the broad store preserves unnecessary coupling; an extra wrapper interface duplicates an already correct domain contract.
11. **Trade-offs.** Additional narrow interfaces can fragment APIs if created mechanically, so this rule is applied where a concrete read-only boundary exists.
12. **Regression tests / protection.** HTTP/server compile tests and existing interface conformance through the production PostgreSQL implementation.
13. **Adversarial review findings.** Interface narrowing must not duplicate semantics or create a new service-container abstraction solely for aesthetic reasons.
14. **Remediation iterations.** The same review deliberately rejected forcing shared service-object identity across independent stateless composition graphs.
15. **Residual risks and limitations.** Other boundaries may still use intentionally broader contracts where behavior actually requires them.
16. **Operational or deployment consequences.** None; runtime behavior and PostgreSQL implementation remain unchanged.
17. **Exact evidence.** Historical implementation commit `2573892ad7684f3d2646378e2021638a53173bc3`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-ARCH-093=CLOSED`.
19. **Prevention / future guard.** New read-only HTTP composition should prefer existing narrow reader contracts when they accurately represent required behavior; do not create mechanical one-method interfaces without a real boundary benefit.

### GFA-OPS-094 — application rate limiting could throttle the readiness probe

1. **Finding / symptom.** `/api/v1/ready` could pass through the application rate limiter instead of remaining an infrastructure dependency-status endpoint.
2. **Root cause.** The new readiness route was not initially classified with `/health` and `/version` in infrastructure-route rate-limit exclusions.
3. **Failure scenario.** A platform executes frequent readiness checks while the application limiter is configured tightly; readiness begins returning HTTP 429 even though PostgreSQL is healthy.
4. **Impact.** Deployment orchestration can mark healthy instances unavailable/restart them due to application traffic policy rather than dependency state.
5. **Severity rationale.** **P1 retrospective.** This can directly destabilize production service availability and makes the readiness signal semantically false.
6. **Existing guarantees violated.** Infrastructure liveness/readiness endpoints must report lifecycle/dependency state independently of end-user application rate limits.
7. **Considered solutions.** Increase rate-limit quota; exempt only liveness; classify readiness as infrastructure and bypass application limiter; add a separate infrastructure listener.
8. **Chosen remediation.** Exclude `/api/v1/ready` from the application rate limiter together with `/health` and `/version`.
9. **Why this solution was selected.** It preserves one server while keeping infrastructure probes semantically independent from application request policy.
10. **Rejected alternatives.** Larger quotas can still be exhausted and couple readiness to traffic; a second listener is unnecessary complexity for the MVP.
11. **Trade-offs.** Readiness is intentionally not protected by the generic application limiter; it remains a cheap bounded PostgreSQL ping endpoint.
12. **Regression tests / protection.** A one-request limiter test repeatedly calls readiness and proves the result remains dependency status instead of 429.
13. **Adversarial review findings.** Exemption must be path-specific and not accidentally bypass rate limiting for business endpoints; readiness work itself remains bounded by its database-ping timeout.
14. **Remediation iterations.** Readiness was first added as a separate dependency probe, then full server review classified it correctly as infrastructure traffic.
15. **Residual risks and limitations.** Volumetric protection at an upstream proxy/platform is a separate deployment concern and may still apply globally.
16. **Operational or deployment consequences.** Platform health checks cannot consume or be blocked by application limiter buckets.
17. **Exact evidence.** Historical implementation commit `2573892ad7684f3d2646378e2021638a53173bc3`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-OPS-094=CLOSED`.
19. **Prevention / future guard.** Infrastructure endpoints must be explicitly classified in middleware policy and tested under intentionally tiny application limits.
