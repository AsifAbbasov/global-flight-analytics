# Document 85 — Ingestion Retry Scheduling and Fallback Evidence Hardening

Status: Implemented
Project: Global Flight Analytics
Scope: traffic ingestion retry timing, local denial semantics, fallback attempt evidence, and OpenSky polling ownership

## 1. Failure mode

The traffic ingestion daemon previously waited one fixed interval after every cycle. Provider-directed retry times were visible in errors but did not affect scheduling, repeated failures did not receive bounded backoff, and locally denied requests could still create failed ingestion-run rows even though no external HTTP operation occurred.

Fallback decisions also recorded only selected providers and one trigger reason. A recoverable primary failure followed by a non-recoverable secondary failure could terminate without a complete ordered attempt record.

## 2. Retry scheduling contract

The daemon now calculates the next delay from:

```text
normal interval
bounded exponential failure backoff
provider-directed RetryAt
```

The first failure waits the normal interval. Repeated consecutive failures double the delay up to `TRAFFIC_INGESTION_MAX_BACKOFF`. A later provider-directed retry time always wins over the local backoff.

A successful cycle resets the consecutive-failure count.

## 3. External-request evidence

Errors that represent a local policy decision expose:

```text
ExternalRequestAttempted() bool
RetryAtTime() time.Time
```

The contract is implemented by:

```text
ingestionorchestrator.AccessDeniedError
providerfallback.NoProviderAvailableError
opensky.PollingTooSoonError
```

Traffic ingestion does not create an `ingestion_runs` row when the entire provider chain was rejected locally before an HTTP attempt. Real provider requests that fail continue to create durable failed-run evidence.

## 4. Ordered fallback evidence

Every fallback decision now retains ordered attempt evidence:

```text
provider
outcome
reason
retry_at
error_class
request_attempted
```

Supported outcomes include success, denied, failed, and terminal failure. Mixed paths such as a primary server error followed by a secondary authorization error are recorded before the original terminal error is returned.

Provider Decision Collector copies attempt slices at both input and output boundaries so callers cannot mutate stored evidence.

## 5. OpenSky polling ownership

OpenSky now returns a typed polling-cooldown error with an exact retry time. A polling reservation can be released when request preparation fails before the HTTP transport is invoked.

The first unauthorized response body is explicitly closed before the authenticated retry starts.

## 6. Configuration

```text
TRAFFIC_INGESTION_MAX_BACKOFF=2m
```

The configured maximum must be at least the normal ingestion interval.

## 7. Verification

The permanent verification path includes:

```text
configuration tests
ingest daemon retry-policy tests
local-denial ingestion tests
fallback selector evidence tests
mixed fallback terminal-path tests
provider decision copy-boundary tests
OpenSky polling reservation tests
targeted race detector
full backend test suite
Go static analysis
working-tree diff validation
```

## 8. Completion boundary

This increment closes retry scheduling, false failed-run creation, incomplete fallback-chain evidence, OpenSky local polling retry metadata, polling-slot release before transport, and delayed first-401 body closure.

OurAirports publication reservation and commit semantics remain the final separate ingestion-layer increment.

---

## Canonical remediation history

### GFA-OPS-066 — provider retry evidence did not control ingestion scheduling

1. **Finding / symptom.** The daemon slept one fixed interval after every cycle even when provider errors supplied a later retry time, and repeated failures did not receive bounded backoff.
2. **Root cause.** Retry metadata was treated as diagnostic text rather than scheduling input; failure count had no explicit policy owner.
3. **Failure scenario.** A provider returns a cooldown or repeated transient failures and the daemon wakes again too early on every cycle.
4. **Impact.** Avoidable provider pressure, wasted requests, noisy failures and possible rate-limit amplification.
5. **Severity rationale.** **P2 retrospective.** This is an operational/reliability defect with potential provider-access consequences, but not direct data fabrication.
6. **Existing guarantees violated.** Provider-directed retry evidence must constrain request timing; repeated failures must not form a tight fixed-frequency loop.
7. **Considered solutions.** Fixed delay; provider `RetryAt` only; local exponential backoff only; bounded composition of normal interval, backoff and provider retry.
8. **Chosen remediation.** Compute the next delay from normal interval, bounded exponential backoff and provider `RetryAt`, with provider-directed later retry winning.
9. **Why this solution was selected.** It respects external cooldown evidence while remaining bounded and deterministic when providers do not supply retry metadata.
10. **Rejected alternatives.** Fixed delay ignores evidence; provider-only retry gives no policy for repeated failures without metadata; unlimited exponential backoff can starve recovery.
11. **Trade-offs.** Local backoff policy introduces configuration and may delay recovery slightly after repeated transient errors.
12. **Regression tests / protection.** Config validation and daemon retry-policy tests, plus full/race backend gates.
13. **Adversarial review findings.** Success must reset the consecutive-failure count; configured max backoff must not be shorter than the normal interval; later provider retry must dominate local backoff.
14. **Remediation iterations.** Retry policy was elevated from error metadata to an explicit scheduling algorithm.
15. **Residual risks and limitations.** A provider can still supply overly conservative retry guidance; policy does not predict provider recovery beyond available evidence.
16. **Operational or deployment consequences.** `TRAFFIC_INGESTION_MAX_BACKOFF` becomes an operator-owned runtime setting.
17. **Exact evidence.** Historical implementation commit `bd291eaa758a30329abb10ffb15542c70d05e82e` (`fix: harden ingestion retry and fallback evidence`). Historical pull-request/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-OPS-066=CLOSED`.
19. **Prevention / future guard.** Any new provider cooldown/error type must expose scheduling metadata through the same retry-policy path and be covered by delay-ordering tests.

### GFA-DATA-067 — local policy denial could create false failed provider-run evidence

1. **Finding / symptom.** An ingestion run could be recorded as a failed external provider attempt when every provider was rejected locally before HTTP transport.
2. **Root cause.** Run persistence did not distinguish local admission failure from an actual external request attempt.
3. **Failure scenario.** Budget or polling policy denies a request locally, but a failed `ingestion_runs` row is still stored as though the provider had been contacted and failed.
4. **Impact.** Historical provider reliability evidence becomes false and operational diagnostics overstate external failures.
5. **Severity rationale.** **P2 retrospective.** The defect fabricates operational provenance/status evidence but does not fabricate aircraft observations.
6. **Existing guarantees violated.** Durable ingestion history must truthfully represent whether an external request was attempted.
7. **Considered solutions.** Never create runs for failures; always create runs; infer from error strings; expose typed `ExternalRequestAttempted()` evidence.
8. **Chosen remediation.** Local-denial error contracts expose request-attempt state; no failed run is created when the full provider chain is denied before transport.
9. **Why this solution was selected.** Typed evidence keeps orchestration independent from error text and preserves real failed-run records when HTTP was actually attempted.
10. **Rejected alternatives.** String inspection is brittle; suppressing all failed runs loses real failure evidence; always persisting preserves the false-evidence bug.
11. **Trade-offs.** Some locally denied cycles intentionally leave no ingestion-run row, so decision evidence must be read from provider/fallback policy evidence instead.
12. **Regression tests / protection.** Local-denial ingestion tests and request-attempt metadata tests for the typed denial errors.
13. **Adversarial review findings.** Mixed provider chains must report `request_attempted=true` if any external attempt occurred, even when the terminal error comes from a later local denial.
14. **Remediation iterations.** Request-attempt semantics were added to multiple error types so the rule applies across orchestration, no-provider and OpenSky polling denial.
15. **Residual risks and limitations.** The contract depends on future denial error types implementing request-attempt semantics correctly.
16. **Operational or deployment consequences.** Provider failure dashboards must not equate a locally denied cycle with a failed external request.
17. **Exact evidence.** Historical implementation commit `bd291eaa758a30329abb10ffb15542c70d05e82e`. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-DATA-067=CLOSED`.
19. **Prevention / future guard.** Admission-policy errors must preserve explicit external-request-attempt evidence and regression tests must cover zero-attempt and mixed-attempt chains.

### GFA-DATA-068 — fallback history did not preserve the complete ordered attempt chain

1. **Finding / symptom.** Fallback decisions stored selected providers and one trigger reason but could omit intermediate/terminal attempts.
2. **Root cause.** Decision evidence modeled a summary outcome rather than an ordered immutable attempt sequence.
3. **Failure scenario.** Primary fails recoverably, secondary fails non-recoverably, and the final record lacks enough information to reconstruct both attempts and why fallback stopped.
4. **Impact.** Provider behavior, outage diagnosis and retry decisions cannot be audited from durable decision evidence.
5. **Severity rationale.** **P2 retrospective.** Evidence integrity and diagnosability are materially affected, without direct corruption of accepted observations.
6. **Existing guarantees violated.** Provider selection/fallback must be explainable and attempt order must be reconstructable.
7. **Considered solutions.** Store only final provider; store first trigger; log attempts only; persist structured ordered attempt evidence.
8. **Chosen remediation.** Record provider, outcome, reason, retry time, error class and request-attempt state for every ordered attempt; copy slices at collector boundaries.
9. **Why this solution was selected.** It makes fallback reproducible and prevents callers from mutating recorded history after submission.
10. **Rejected alternatives.** Logs are not durable decision records; one-summary fields cannot represent mixed failure chains; shared mutable slices undermine evidence integrity.
11. **Trade-offs.** Decision records are larger and require version-conscious evolution when new outcome metadata is added.
12. **Regression tests / protection.** Fallback selector evidence, mixed terminal-path and collector copy-boundary tests.
13. **Adversarial review findings.** A terminal secondary failure must still preserve the primary attempt; caller mutation after collection must not alter stored evidence.
14. **Remediation iterations.** The model evolved from summary reason/provider fields to a structured ordered attempt list.
15. **Residual risks and limitations.** Attempt records explain application decisions, not undocumented behavior inside the remote provider.
16. **Operational or deployment consequences.** Observability consumers can reconstruct fallback chains without parsing logs.
17. **Exact evidence.** Historical implementation commit `bd291eaa758a30329abb10ffb15542c70d05e82e`. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-DATA-068=CLOSED`.
19. **Prevention / future guard.** Future fallback paths must append one immutable evidence item per evaluated provider before returning or selecting the next provider.

### GFA-REL-069 — OpenSky polling reservation and unauthorized-response lifecycle were incomplete

1. **Finding / symptom.** Polling cooldown lacked exact typed retry metadata, a reservation could remain consumed when request preparation failed before transport, and the first unauthorized response body was not guaranteed closed before authenticated retry.
2. **Root cause.** Polling admission and HTTP response cleanup were spread across request preparation/retry paths without one explicit lifecycle contract.
3. **Failure scenario.** Local preparation fails after acquiring a polling slot, delaying a later valid request; or an unauthorized first response remains open while the authenticated retry starts.
4. **Impact.** Avoidable local denial, resource leakage/connection reuse degradation and inaccurate retry scheduling.
5. **Severity rationale.** **P2 retrospective.** This affects provider reliability/resource lifecycle but is bounded to one integration path and does not directly corrupt stored observations.
6. **Existing guarantees violated.** Reservations must represent actual transport opportunities; local failure before transport must be releasable; every HTTP response body must be closed before retry; cooldown errors must carry actionable retry evidence.
7. **Considered solutions.** Keep reservations consumed on any failure; manually infer cooldown; defer body close; explicit typed cooldown plus release-before-transport and immediate response cleanup.
8. **Chosen remediation.** Add typed `PollingTooSoonError` with exact retry time, release polling reservations on pre-transport preparation failure, and close the first 401 body before retry.
9. **Why this solution was selected.** It aligns rate policy with actual request attempts and makes HTTP resource ownership deterministic.
10. **Rejected alternatives.** Consuming unused slots reduces availability; delayed close risks connection/resource accumulation; string retry metadata is not machine-actionable.
11. **Trade-offs.** Reservation lifecycle gains additional branches and tests; callers must preserve typed error behavior.
12. **Regression tests / protection.** OpenSky polling reservation tests, retry scheduling tests and full/race verification.
13. **Adversarial review findings.** Reservation release is valid only before transport; after transport begins the attempt must remain accounted. Unauthorized retry cannot begin with the first body still owned.
14. **Remediation iterations.** The final change coupled retry metadata, reservation release and body cleanup because they share the same request-attempt ownership boundary.
15. **Residual risks and limitations.** Remote server connection behavior remains outside application control after the client correctly closes bodies.
16. **Operational or deployment consequences.** OpenSky cooldown now participates directly in daemon scheduling and unused local reservations no longer suppress future attempts.
17. **Exact evidence.** Historical implementation commit `bd291eaa758a30329abb10ffb15542c70d05e82e`. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-REL-069=CLOSED`.
19. **Prevention / future guard.** Provider admission reservations must have explicit acquire/use/release ownership and every retry path must prove prior response cleanup.
