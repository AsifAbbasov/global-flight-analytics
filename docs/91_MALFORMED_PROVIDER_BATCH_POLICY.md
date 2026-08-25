# Malformed Provider Batch Policy

Status: Implemented Engineering Contract v1.0

## Scope

This increment defines one explicit item-level policy for successful
Airplanes.live and OpenSky transport responses.

## Policy

A provider batch is accounted as received items, accepted items, malformed
rejected items, and unavailable or stale rejected items. Every received
item belongs to exactly one accepted or rejected category.

A mixed batch remains usable. Valid items continue through normalization,
validation, persistence and trajectory construction. The ingestion run is
marked partial, and provider rejection evidence is retained.

An empty provider batch is a valid empty result.

A non-empty batch with zero accepted items returns the typed
`providerbatch.ErrAllItemsRejected` error. Automatic traffic mode treats
this as a recoverable invalid-response failure and may try the next
provider.

## Airplanes.live Boundary

An item is malformed when its required aircraft identity is empty, its
latitude or longitude is not finite, or its coordinates are outside valid
geographic ranges. Invalid optional telemetry remains represented through
existing availability and altitude-status semantics.

An invalid response timestamp makes every non-empty item malformed because
no trustworthy observation time can be produced.

## OpenSky Boundary

State vectors that fail validity evaluation are malformed. State vectors
with unavailable or stale provider positions are rejected as unusable.
Neither category terminates a mixed batch.

## Evidence Propagation

Provider batch evidence travels through provider orchestration, automatic
fallback and traffic ingestion. Ingestion-run records use raw provider
received counts. Provider health combines provider-level rejections with
processing invalid and duplicate counts.

## Verification

The acceptance gate includes provider accounting invariants, mixed and
fully rejected batches for both providers, partial-run policy tests,
fallback invalid-response classification, race tests, complete backend
tests, `go vet`, policy gates and clean diff validation.

## Review Closure Boundary

This document closes only the malformed item-level batch policy for
Airplanes.live and OpenSky. It does not close duration overflow protection,
Open-Meteo metric availability, the OurAirports atomic publication policy,
PostgreSQL isolated-fixture alignment, or independent Continuous Integration
verification. Complete Ingestion review closure is defined by Document 92.

---

## Canonical remediation history

### GFA-DATA-083 — successful provider batches had no explicit malformed-item accounting policy

1. **Finding / symptom.** Airplanes.live/OpenSky successful HTTP responses lacked one explicit policy for mixed valid, malformed, unavailable and stale items.
2. **Root cause.** Transport success and per-item data validity were conflated; adapters/orchestration did not share a canonical accounting invariant for every received item.
3. **Failure scenario.** One malformed item terminates a batch that contains valid observations, or malformed/unusable items are silently dropped without counts; conversely a non-empty all-invalid response can be mistaken for a valid empty response.
4. **Impact.** Valid traffic data may be unnecessarily lost, ingestion status/evidence can be inaccurate, and fallback cannot distinguish valid empty from unusable provider output.
5. **Severity rationale.** **P1 retrospective.** The policy directly controls which provider observations enter persistence and whether fallback is triggered.
6. **Existing guarantees violated.** Every received item must be accounted exactly once; mixed batches should preserve valid evidence; a non-empty zero-accepted batch must not masquerade as a legitimate empty result.
7. **Considered solutions.** Fail whole batch on first invalid item; silently skip invalid items; accept all and rely on later validation; explicit accepted/rejected categories with typed all-rejected outcome.
8. **Chosen remediation.** Account received, accepted, malformed-rejected and unavailable/stale-rejected items; continue mixed batches as partial; return typed `ErrAllItemsRejected` for non-empty batches with zero accepted items.
9. **Why this solution was selected.** It maximizes truthful usable evidence while retaining complete rejection accounting and a machine-actionable fallback signal.
10. **Rejected alternatives.** Fail-whole is too destructive for live telemetry; silent skip destroys provenance; defer-all-validation permits malformed provider evidence deeper into the pipeline.
11. **Trade-offs.** Mixed batches require partial status and richer provider-health accounting; adapters must own provider-specific required-field validity rules.
12. **Regression tests / protection.** Accounting invariants, mixed/all-rejected Airplanes.live and OpenSky tests, partial-run policy, fallback invalid-response classification, race/full backend gates.
13. **Adversarial review findings.** Empty batch is valid and distinct from non-empty all-rejected; invalid optional telemetry must remain an availability issue rather than make the whole item malformed; invalid response-level timestamp makes all non-empty items unusable because no trustworthy observation time exists.
14. **Remediation iterations.** Provider-specific validity rules were kept at adapters while common batch outcome/accounting semantics were centralized for orchestration.
15. **Residual risks and limitations.** Provider schemas may introduce new required fields or sentinel values that require adapter-specific malformed criteria updates.
16. **Operational or deployment consequences.** Ingestion runs and health evidence now expose rejected counts; mixed provider degradation can produce terminal `partial` rather than total failure.
17. **Exact evidence.** Historical implementation commit `b7bf2b762290e55a45fa8d40641435248d1aeddf` (`fix: enforce malformed provider batch policy`). A later commit `6b922cbd9df1bff3f880ad120dd883b37f658e53` extended final ingestion-review closure. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-DATA-083=CLOSED`.
19. **Prevention / future guard.** Every provider batch adapter must preserve exact received=accepted+rejected accounting, distinguish empty from all-rejected, and maintain mixed-batch/fallback regression tests.
