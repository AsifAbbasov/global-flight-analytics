# Document 84 — Provider HTTP Resilience Hardening

Status: Implemented
Project: Global Flight Analytics
Scope: external provider error preservation, response body limits, fallback classification, and non-destructive observation

## 1. Problem

Provider response observation previously participated in the primary data-plane result. For Airplanes.live and Open-Meteo, an observation failure could replace a real HTTP 429 or 500 error. OpenSky could reject an otherwise valid response when its observer failed. This made automatic fallback dependent on an auxiliary telemetry path.

External JSON and CSV response bodies were also decoded without an explicit byte limit. HTTP timeouts bound duration but do not bound decoded response size.

## 2. Error preservation contract

The provider HTTP status remains the primary error classification.

For failed HTTP responses:

```text
provider status error + observer error = errors.Join
```

This preserves `errors.Is` and `errors.As` matching for rate limiting, provider server failure, authorization failure, and other typed provider outcomes.

For successful responses, observation is best-effort. An auxiliary health or budget evidence failure does not discard a valid provider payload.

Transport and response parsing failures remain visible as primary failures. Observation failures may be joined to those failures but cannot replace them.

## 3. Response size contract

A shared integration helper now enforces response limits before decoding:

```text
Airplanes.live state response: 8 MiB
OpenSky state response: 16 MiB
Open-Meteo weather response: 1 MiB
OurAirports CSV response: 32 MiB
```

The helper rejects:

- a declared `Content-Length` above the limit;
- a streamed response that exceeds the limit even when `Content-Length` is absent or incorrect.

The canonical typed error is:

```text
integrations/common.ErrProviderResponseTooLarge
```

No partial provider payload is published after the limit is exceeded.

## 4. Fallback guarantee

Automatic traffic fallback continues to recognize the original provider failure when an observer failure is joined to it. A primary Airplanes.live server failure still permits OpenSky fallback.

## 5. Verification

The permanent regression coverage includes:

```text
exact-limit response acceptance
declared oversized response rejection
streamed oversized response rejection
Airplanes.live server classification with observer failure
OpenSky server classification with observer failure
OpenSky successful payload with observer failure
Open-Meteo server classification with observer failure
oversized response wiring for all four provider adapters
fallback after a joined provider and observer failure
targeted race tests
full backend tests
Go static analysis
git diff validation
```

## 6. Completion boundary

This document closes the provider observer error replacement and unbounded response body findings. It does not close daemon retry scheduling, publication reservation lifecycle, or full per-attempt fallback evidence. Those remain separate orchestration contracts.

---

## Canonical remediation history

### GFA-REL-064 — auxiliary provider observation could replace the primary data-plane result

1. **Finding / symptom.** Health/budget observation failures could replace real HTTP status errors or reject otherwise valid provider payloads.
2. **Root cause.** Auxiliary observation participated in the primary provider return path instead of being subordinate evidence collection.
3. **Failure scenario.** A provider returns HTTP 429/500 while the observer also fails; fallback sees the observer error rather than the typed provider error. On a successful OpenSky payload, observer failure could discard usable traffic data.
4. **Impact.** Automatic fallback classification becomes incorrect, valid observations can be lost, and provider error evidence becomes misleading.
5. **Severity rationale.** **P1 retrospective.** The defect can directly suppress valid source data and change provider-selection behavior on production ingestion paths.
6. **Existing guarantees violated.** Provider transport/status classification is authoritative; auxiliary telemetry must not alter successful source data or obscure typed provider failures.
7. **Considered solutions.** Ignore observer errors entirely; let observer errors replace provider outcomes; wrap errors opaquely; join observer error to failed primary outcomes while making successful observation best-effort.
8. **Chosen remediation.** Preserve the provider/transport/parsing result as primary, use `errors.Join` on failed paths, and make observation failures non-destructive on successful responses.
9. **Why this solution was selected.** It retains diagnostic evidence without breaking `errors.Is`/`errors.As` classification or discarding valid payloads.
10. **Rejected alternatives.** Dropping observer errors loses diagnostics; opaque wrapping can break classification; keeping observer errors authoritative preserves the original defect.
11. **Trade-offs.** A successful payload may be accepted without complete auxiliary health evidence, which is preferable to inventing a data-plane failure.
12. **Regression tests / protection.** Provider-specific status-plus-observer tests, successful OpenSky observer-failure test, fallback-after-joined-error test, race tests and full backend verification.
13. **Adversarial review findings.** Joined errors must preserve typed provider classification; success must remain success even when observation fails; transport/parsing failure must remain visible.
14. **Remediation iterations.** The final contract distinguishes failed primary outcomes from successful primary outcomes rather than applying one blanket observer policy.
15. **Residual risks and limitations.** Auxiliary evidence can still be absent for a successful response; downstream observability must represent that absence rather than infer health.
16. **Operational or deployment consequences.** Provider health/budget telemetry is best-effort relative to source payload acceptance; operators must diagnose joined errors without assuming the observer caused the provider failure.
17. **Exact evidence.** Historical implementation commit `57a67488e1717f1109eab3a850e09d4525ca444d` (`fix: harden provider HTTP resilience`). Historical pull-request/reviewer evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-REL-064=CLOSED`.
19. **Prevention / future guard.** New provider adapters must keep auxiliary observation outside the authoritative data-plane result and retain typed-error/fallback regression coverage.

### GFA-REL-065 — provider response bodies were not explicitly bounded

1. **Finding / symptom.** External JSON and CSV bodies were decoded without an explicit maximum byte count.
2. **Root cause.** HTTP timeout ownership bounded elapsed time but was incorrectly treated as sufficient protection for response size and memory consumption.
3. **Failure scenario.** A provider or intermediary streams an unexpectedly large body within the timeout or reports a false/missing `Content-Length`, causing excessive memory use before decoding fails.
4. **Impact.** A single external response can create avoidable memory pressure or process instability and may consume resources before application validation runs.
5. **Severity rationale.** **P1 retrospective.** The untrusted external data boundary allowed unbounded resource consumption on production provider paths.
6. **Existing guarantees violated.** External input must be bounded before decoding; time limits and byte limits are independent resilience controls.
7. **Considered solutions.** Trust provider `Content-Length`; rely only on timeout; use decoder-specific limits; centralize a bounded-body helper that checks declared and streamed size.
8. **Chosen remediation.** Introduce a shared bounded response reader with provider-specific limits and typed `integrations/common.ErrProviderResponseTooLarge`.
9. **Why this solution was selected.** It covers both honest oversized declarations and streamed bodies with absent/incorrect length while keeping limits visible per provider.
10. **Rejected alternatives.** `Content-Length` alone is forgeable/optional; timeout alone does not bound bytes; duplicated adapter-specific readers would drift.
11. **Trade-offs.** Hard limits can reject legitimately larger future provider payloads and therefore require explicit policy changes when provider contracts evolve.
12. **Regression tests / protection.** Exact-limit acceptance, declared oversized rejection, streamed oversized rejection and wiring tests across Airplanes.live, OpenSky, Open-Meteo and OurAirports.
13. **Adversarial review findings.** The guard must reject a stream that exceeds the limit even with no declared length and must not publish partial provider data after overflow.
14. **Remediation iterations.** Centralization replaced per-adapter implicit behavior so all four external adapters share one size-enforcement primitive.
15. **Residual risks and limitations.** Limits protect response-body memory, not every possible downstream allocation after a valid bounded payload is decoded.
16. **Operational or deployment consequences.** Provider payload growth beyond the documented limits becomes a visible typed failure requiring reviewed limit changes rather than silent memory expansion.
17. **Exact evidence.** Historical implementation commit `57a67488e1717f1109eab3a850e09d4525ca444d` (`fix: harden provider HTTP resilience`). Historical pull-request/reviewer evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-REL-065=CLOSED`.
19. **Prevention / future guard.** Every new external HTTP adapter must declare and test a bounded body policy before decoding, including streamed-overflow behavior.
