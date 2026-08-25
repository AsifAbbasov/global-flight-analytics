# Health-Aware Traffic Provider Selection

Status: Implemented Engineering Contract v1.0

## Scope

This increment closes the health-aware primary and fallback selection
boundary for automatic traffic ingestion. The configured provider order
remains authoritative when health evidence is equal, unknown, or
unavailable. Stronger health evidence may change the attempt order before
an external request is made.

## Previous Failure Mode

Automatic mode always attempted Airplanes.live before OpenSky. Provider
health was collected and printed, but it was not used by the selection
path. A provider already classified as unavailable was therefore attempted
again before a provider with healthy evidence.

## Implemented Ordering Policy

Each cycle reads a fresh snapshot for every configured traffic provider.

The stable priority is:

1. healthy;
2. degraded or unknown;
3. unavailable.

Equal priorities preserve configuration order. The policy does not
permanently disable a provider and does not create a separate circuit
breaker. If the first health-preferred provider fails with a recoverable
error, the remaining configured provider is still attempted.

An unavailable configured primary receives a bounded recovery probe after two
minutes without an attempt. This prevents a previously failing primary from
being permanently starved by a healthy fallback. A successful recovery resets
its consecutive failure evidence through the existing health collector.

Snapshot lookup failure is fail-open. The affected provider receives an
unknown status and configured order is preserved unless another provider
has strictly stronger evidence.

## Decision Evidence

Fallback decisions now record:

- whether health evidence was evaluated;
- whether attempt order changed;
- the ordering reason;
- configured primary health status;
- selected provider health status.

`PrimaryProvider` continues to mean the configured primary provider, even
when a healthier secondary provider is attempted first. Selecting that
secondary provider is recorded as `fallback_selected`, not
`primary_selected`.

## Production Wiring

The traffic provider factory accepts an optional health snapshot source.
Existing isolated callers remain compatible and use unknown health
evidence. Production `cmd/ingest` passes the existing provider health
collector, so automatic mode is health-aware without introducing another
health store.

Provider health history remains process-local. The selection decision is
therefore operational evidence for the current daemon process, not a
cross-process global availability claim.

## Verification

The acceptance gate includes:

- healthy secondary preference over unavailable primary;
- configured order preservation for equal evidence;
- bounded configured-primary recovery probing;
- fail-open behavior when snapshots cannot be read;
- existing fallback compatibility tests;
- provider decision evidence tests;
- targeted race detector;
- complete backend tests;
- `go vet`;
- existing code review policy gates;
- clean Git diff validation.

## Remaining Review Boundary

After this increment, the remaining known Ingestion, Provider Adapters and
Orchestration review boundary is the explicit malformed-item policy for an
otherwise successful provider batch.

---

## Canonical remediation history

### GFA-OPS-082 — provider health evidence was collected but ignored by automatic selection

1. **Finding / symptom.** Automatic mode always attempted the configured primary first even when current health evidence classified it unavailable and a secondary provider healthy.
2. **Root cause.** Health collection/printing and provider selection were separate subsystems with no explicit decision-policy connection.
3. **Failure scenario.** The primary is already known unavailable; every cycle still spends an attempt/failure on it before reaching the healthy fallback.
4. **Impact.** Avoidable latency, request waste, noisy failures and slower recovery during provider incidents.
5. **Severity rationale.** **P2 retrospective.** It materially degrades ingestion reliability but does not by itself corrupt accepted source observations.
6. **Existing guarantees violated.** Available health evidence should inform attempt order while configured provider order remains authoritative when evidence is equal, unknown or unavailable.
7. **Considered solutions.** Keep static order; permanently disable unhealthy providers; introduce a circuit breaker service; stable health-aware ranking with bounded primary recovery probes.
8. **Chosen remediation.** Rank healthy before degraded/unknown before unavailable, preserve configured order for equal priority, fail open on snapshot lookup failure, and periodically probe the configured primary.
9. **Why this solution was selected.** It uses existing health evidence without adding infrastructure or allowing a transient status to permanently remove the configured primary.
10. **Rejected alternatives.** Static order ignores evidence; permanent disable can starve recovery; a new circuit-breaker subsystem adds complexity beyond the measured need.
11. **Trade-offs.** Health history remains process-local, so ordering is daemon-local operational evidence rather than a cross-instance global health truth.
12. **Regression tests / protection.** Healthy-secondary preference, equal-evidence order preservation, recovery probing, fail-open lookup behavior, fallback compatibility, decision evidence and race/full backend tests.
13. **Adversarial review findings.** An unavailable primary must eventually receive a recovery probe; equal/unknown evidence must not reorder configuration arbitrarily; snapshot-read failure must not block all providers.
14. **Remediation iterations.** Bounded two-minute primary probing was added to avoid starvation introduced by naive always-healthiest-first ranking.
15. **Residual risks and limitations.** Process-local health can differ between daemon instances and is not represented as a global provider availability claim.
16. **Operational or deployment consequences.** Production `cmd/ingest` passes the existing health collector into selection; no new datastore or service is deployed.
17. **Exact evidence.** Historical implementation commit `a9896ade17f6a36b80a5cef6abb8ffd9a5687cc1` (`fix: make traffic provider selection health aware`). Historical pull-request/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-OPS-082=CLOSED`.
19. **Prevention / future guard.** Future provider-ordering policies must preserve configuration-order fallback, bounded recovery probing and explicit decision evidence whenever health can reorder attempts.
