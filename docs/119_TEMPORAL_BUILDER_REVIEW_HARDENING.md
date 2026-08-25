# Temporal Builder Review Hardening

## Scope

This document classifies the static review of
`apps/api/internal/features/temporalbuilder` against repository baseline
`2eb2c49c11fc1894969ef62c0f9ea0e244a3103f` and records the implemented
closure controls.

## Accepted findings

### Persisted trajectory point records are not available to production reads

The PostgreSQL trajectory write path persists the trajectory parent, segments,
and coverage gaps without materializing the in-memory `TrackPoint4D` slice.
Persisting every raw point is intentionally not introduced because the project
uses bounded trajectory summaries rather than an unbounded coordinate archive.

Temporal evidence now prefers point timestamps when they are present and falls
back to unique non-invalid persisted segment start and end timestamps when no
usable point timestamp exists. The authoritative temporal window remains the
trajectory start and end boundary.

### Zero duration was overloaded as an absence sentinel

Duration metadata is a canonical derived mirror, not an optional field. A
non-zero window with `DurationSeconds == 0` is now reported as a metadata
mismatch. A true zero-duration window remains valid because its derived duration
is also zero.

### Fractional-second policy was implicit and duplicated

`flightfeatures.TemporalDurationSeconds` is now the single implementation used
by the temporal builder and validator. The version-one contract truncates
fractional seconds toward zero; for valid non-negative windows this is equivalent
to using complete elapsed seconds.

### Context and evidence diagnostics

A nil context is rejected. Long point and segment scans check cancellation at a
bounded interval. Limitation messages include exact rejected-evidence counts,
and point-count metadata is compared with materialized point records whenever
those records are present.

## Stale findings

### Trajectory UpdatedAt after AsOfTime blocks production materialization

This was corrected before this review baseline. `AsOfTime` is an event-time
cutoff. `TrajectoryUpdatedAt` is system provenance and may occur after the event
cutoff. The validator has a regression test that accepts this relationship.
The extractor separately rejects point, segment, and coverage-gap event times
that exceed the requested cutoff.

### Temporal field count duplicates the schema

This was corrected by the flight-features schema hardening increment. The
builder aliases `flightfeatures.TemporalRequiredFeatureFieldCount`.

## Deliberately rejected recommendations

### Passing AsOfTime into the temporal builder

The extractor owns historical snapshot validation and rejects every nested
point, segment, and coverage-gap event timestamp after the cutoff before any
builder runs. Repeating that policy in each builder would reintroduce the same
cross-module duplication identified by the review.

### Persisting all trajectory points

Raw-point persistence is not required to calculate the eight boundary-derived
temporal features and would conflict with the bounded-storage architecture.
Segment-boundary fallback preserves honest production evidence without changing
other builders or manufacturing synthetic points.

### Introducing a new domain window value object in this increment

Trajectory, extraction, database, and validation boundaries intentionally apply
defense-in-depth checks. Replacing the established aggregate contract across the
entire domain would be a broad migration without a demonstrated temporal-builder
correctness benefit.

### Mechanical method-length decomposition

Line count alone is not a defect. The refactor separates point and segment
evidence evaluation where independent behavior and tests exist, without creating
one-line forwarding methods only to satisfy a numerical threshold.

### Central closed registry for limitation codes

Builder limitations are extensible evidence, while typed severity, group, path,
and validator metadata belong to `ValidationIssue`. A closed central limitation
registry would couple every provider and builder extension to the core model.

### Generic clone helper

The current clone functions are short type-specific ownership boundaries. A
shared generic helper would not remove the need to identify each mutable nested
slice and would obscure the concrete copy contract.

## Version and persistence policy

The temporal builder advances from generation one to generation two. Because
its evidence and mismatch semantics affect immutable snapshot output, the feature
processing version advances from generation seven to generation eight. The
processing version already participates in snapshot identity, so existing
snapshots remain isolated.

The validator remains generation three because its mathematical behavior is
unchanged; it now calls the shared duration policy implementation. No database
migration is required.

## Verification

The guarded installer runs targeted tests, every feature contract audit, race
tests, the complete backend test suite, `go vet`, formatting validation, and
`git diff --check` first in a detached temporary worktree and then in the real
working tree.

## Closure markers

```text
TEMPORAL_SEGMENT_BOUNDARY_FALLBACK=ENFORCED
TEMPORAL_RAW_POINT_PERSISTENCE=NOT_REQUIRED
TEMPORAL_DURATION_METADATA_ZERO_SENTINEL=CLOSED
TEMPORAL_DURATION_ROUNDING_POLICY=CENTRALIZED
TEMPORAL_BUILDER_NIL_CONTEXT=REJECTED
TEMPORAL_BUILDER_LOOP_CANCELLATION=ENFORCED
TEMPORAL_LIMITATION_COUNTS=EXPLICIT
TEMPORAL_POINT_COUNT_MIRROR=CHECKED_WHEN_MATERIALIZED
TEMPORAL_AS_OF_SYSTEM_PROVENANCE_CONFLICT=STALE
TEMPORAL_AS_OF_EVENT_GATE=EXTRACTOR_OWNED
TEMPORAL_FIELD_COUNT_DUPLICATION=STALE
TEMPORAL_BUILDER_GENERATION=v2
TEMPORAL_BUILDER_PROCESSING_GENERATION=v8
DATABASE_MIGRATION=NOT_REQUIRED
TEMPORAL_BUILDER_REVIEW_STATUS=CLOSED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
```

## Canonical remediation history

Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, and Continuous Integration evidence. Severity labels are retrospective.

### GFA-DATA-174 — Production Temporal evidence could become empty when raw points were intentionally not persisted

1. **Finding / symptom:** production trajectory reads can legitimately have no materialized point slice even though persisted segment boundaries contain honest temporal evidence.
2. **Root cause:** Temporal Builder assumed point timestamps were the only meaningful supporting observations.
3. **Failure scenario:** bounded-storage production data is treated as if it had no temporal evidence despite valid persisted segment start/end observations.
4. **Impact:** avoidable loss of feature availability and misleading completeness for production snapshots.
5. **Severity rationale:** **P2 retrospective** because the issue degrades honest evidence availability rather than fabricating values.
6. **Existing guarantees violated:** bounded-storage architecture must still expose available persisted evidence without inventing raw points.
7. **Considered solutions:** persist all raw points, synthesize points from segments, report unavailable, or use segment-boundary timestamps as an explicit fallback.
8. **Chosen remediation:** prefer usable point timestamps when present; otherwise use unique non-invalid segment start/end timestamps while retaining the authoritative trajectory window.
9. **Why this solution was selected:** it uses evidence that actually exists in PostgreSQL and preserves the bounded-storage design.
10. **Rejected alternatives:** unbounded raw-point archive and synthetic point fabrication.
11. **Trade-offs:** fallback evidence has lower temporal granularity than real point observations.
12. **Regression tests / protection:** segment-boundary fallback tests and `temporalbuilderreviewaudit`.
13. **Adversarial review findings:** persisting all trajectory points and passing `AsOfTime` into every builder were explicitly rejected as unnecessary duplication/scope expansion.
14. **Remediation iterations:** implemented in `c0e3323328f81af8bf0b8841b1bf6756d3085d21`.
15. **Residual risks and limitations:** segment boundaries cannot reconstruct sampling detail that was never persisted; this remains visible through evidence/limitations.
16. **Operational or deployment consequences:** no migration and no raw-point storage expansion.
17. **Exact evidence:** commit `c0e3323328f81af8bf0b8841b1bf6756d3085d21`, fallback tests, permanent temporal review audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** builders must distinguish unavailable raw detail from unavailable evidence and use only persisted evidence with explicit semantics.

### GFA-DATA-175 — Zero duration overloaded valid zero and missing/incorrect duration metadata

1. **Finding / symptom:** `DurationSeconds == 0` could mean either a legitimate zero-duration window or an incorrect/missing mirror for a non-zero window.
2. **Root cause:** zero was treated as an absence sentinel for a field that is actually a required derived mirror.
3. **Failure scenario:** a positive trajectory duration carries zero duration metadata and passes without a mismatch.
4. **Impact:** inconsistent temporal metadata can become durable analytical evidence.
5. **Severity rationale:** **P2 retrospective** because the mismatch can skew derived temporal interpretation but is directly detectable from authoritative timestamps.
6. **Existing guarantees violated:** derived duration mirror must equal the canonical window duration.
7. **Considered solutions:** make duration optional, reserve zero as missing, or validate zero against authoritative start/end timestamps.
8. **Chosen remediation:** true zero-duration windows remain valid; non-zero windows with zero mirror report a mismatch.
9. **Why this solution was selected:** preserves legitimate zero while eliminating sentinel ambiguity.
10. **Rejected alternatives:** optionalizing a required mirror or inventing a separate missing state without schema need.
11. **Trade-offs:** stricter validation exposes previously tolerated malformed metadata.
12. **Regression tests / protection:** zero/non-zero duration mismatch tests and temporal review audit.
13. **Adversarial review findings:** no new domain window type was required to enforce this invariant.
14. **Remediation iterations:** `c0e3323328f81af8bf0b8841b1bf6756d3085d21`.
15. **Residual risks and limitations:** correctness still depends on authoritative start/end timestamps themselves being valid, which is guarded at upstream boundaries.
16. **Operational or deployment consequences:** no migration; malformed snapshots are reported more honestly.
17. **Exact evidence:** implementation commit `c0e3323328f81af8bf0b8841b1bf6756d3085d21` and duration metadata tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** required numeric mirrors must not use valid numeric values as implicit absence sentinels.

### GFA-CONTRACT-176 — Fractional-second duration policy was duplicated and implicit

1. **Finding / symptom:** builder and validator independently derived integer duration semantics without one canonical rounding/truncation policy.
2. **Root cause:** the version-one temporal duration contract was encoded in implementation details rather than a shared function.
3. **Failure scenario:** builder and validator evolve differently around sub-second windows and disagree on otherwise identical evidence.
4. **Impact:** cross-component validation drift and non-reproducible temporal semantics.
5. **Severity rationale:** **P2 retrospective** because divergent mathematical policy can invalidate or misclassify snapshots.
6. **Existing guarantees violated:** one deterministic formula per versioned analytical field.
7. **Considered solutions:** duplicate tests, round to nearest, preserve fractional float seconds, or centralize the existing complete-seconds policy.
8. **Chosen remediation:** `flightfeatures.TemporalDurationSeconds` is the single implementation; version one truncates fractional seconds toward zero.
9. **Why this solution was selected:** preserves established output semantics while making them explicit and shared.
10. **Rejected alternatives:** silently changing the numerical contract during a correctness hardening increment.
11. **Trade-offs:** integer seconds intentionally lose sub-second duration detail for this field.
12. **Regression tests / protection:** fractional-duration tests in builder/validator and permanent review audit.
13. **Adversarial review findings:** validator generation remained unchanged because the mathematical rule itself was centralized, not changed.
14. **Remediation iterations:** `c0e3323328f81af8bf0b8841b1bf6756d3085d21`.
15. **Residual risks and limitations:** future higher-resolution metrics need a new explicit contract rather than reinterpretation of this field.
16. **Operational or deployment consequences:** no migration; deterministic semantics across builder and validator.
17. **Exact evidence:** implementation commit `c0e3323328f81af8bf0b8841b1bf6756d3085d21`, shared duration tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** shared analytical formulas must have one package-level owner and versioned tests across all consumers.

### GFA-OPS-177 — Temporal Builder accepted nil context and long scans were not cancellation-aware

1. **Finding / symptom:** nil context could be accepted and long point/segment scans did not consistently observe cancellation.
2. **Root cause:** lifecycle checks were omitted from a pure-computation builder because it was assumed to be cheap.
3. **Failure scenario:** large evidence sets continue processing after request cancellation or run without an owned caller context.
4. **Impact:** wasted CPU and inconsistent cancellation behavior in materialization paths.
5. **Severity rationale:** **P2 retrospective** as an operational lifecycle defect.
6. **Existing guarantees violated:** bounded work and explicit context ownership through analytical stages.
7. **Considered solutions:** ignore context, check every iteration, or reject nil and poll cancellation at a bounded interval.
8. **Chosen remediation:** nil contexts are rejected and long loops periodically check cancellation.
9. **Why this solution was selected:** predictable responsiveness without unnecessary per-element overhead.
10. **Rejected alternatives:** background fallback and cancellation checks only before/after the full scan.
11. **Trade-offs:** small bookkeeping overhead in long scans.
12. **Regression tests / protection:** nil-context and mid-scan cancellation tests, temporal review audit.
13. **Adversarial review findings:** a generic clone/helper refactor was rejected because it did not strengthen lifecycle invariants.
14. **Remediation iterations:** `c0e3323328f81af8bf0b8841b1bf6756d3085d21`.
15. **Residual risks and limitations:** cancellation latency is bounded by polling interval, not instantaneous.
16. **Operational or deployment consequences:** more responsive cancellation under large evidence sets.
17. **Exact evidence:** commit `c0e3323328f81af8bf0b8841b1bf6756d3085d21`, context/cancellation tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** any builder loop over potentially large evidence collections must own explicit cancellation policy.

### GFA-DATA-178 — Temporal limitations did not report exact rejected-evidence counts

1. **Finding / symptom:** diagnostics could state that evidence was rejected without preserving exact counts of rejected point/segment observations.
2. **Root cause:** limitation text prioritized categorical reason over quantitative evidence accounting.
3. **Failure scenario:** two snapshots with materially different evidence loss present indistinguishable diagnostic messages.
4. **Impact:** weaker explainability and harder audit/debug of completeness reductions.
5. **Severity rationale:** **P3 retrospective** because output values remain protected; the gap is diagnostic precision.
6. **Existing guarantees violated:** explainable evidence limitations should quantify material exclusions where counts are known.
7. **Considered solutions:** add schema counters, generic codes only, or encode exact counts in typed limitation messages.
8. **Chosen remediation:** limitation messages include exact rejected-evidence counts without expanding the feature schema.
9. **Why this solution was selected:** improves evidence traceability with minimal schema churn.
10. **Rejected alternatives:** adding many permanent diagnostic fields for data already available during validation.
11. **Trade-offs:** consumers parsing free-form messages must not treat wording as a stable machine API.
12. **Regression tests / protection:** exact limitation-count assertions and temporal audit.
13. **Adversarial review findings:** a closed global limitation-code registry was rejected to preserve extensible builder evidence.
14. **Remediation iterations:** `c0e3323328f81af8bf0b8841b1bf6756d3085d21`.
15. **Residual risks and limitations:** counts improve diagnostics but do not replace structured per-record provenance.
16. **Operational or deployment consequences:** better debugging; no persistence migration.
17. **Exact evidence:** implementation commit `c0e3323328f81af8bf0b8841b1bf6756d3085d21` and limitation tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** when a builder discards countable evidence, regression tests should require the limitation to expose that count or an equivalent structured measure.

### GFA-DATA-179 — Materialized point-count metadata was not reconciled with actual point evidence

1. **Finding / symptom:** when point records were present, persisted trajectory point-count metadata could disagree without an explicit mismatch signal.
2. **Root cause:** metadata and materialized evidence were consumed independently.
3. **Failure scenario:** a trajectory declares one point count while the actual materialized point slice contains another number; quality/support diagnostics trust both inconsistently.
4. **Impact:** misleading evidence accounting and possible completeness/support disagreement.
5. **Severity rationale:** **P2 retrospective** because the mismatch concerns analytical evidence integrity, though it is observable and does not directly fabricate coordinates.
6. **Existing guarantees violated:** persisted mirrors must reconcile with materialized evidence when both are available.
7. **Considered solutions:** always trust metadata, always replace it with slice length, or compare and disclose disagreement.
8. **Chosen remediation:** compare point-count metadata with materialized records whenever records are present and report mismatch explicitly.
9. **Why this solution was selected:** preserves both evidence sources without silently overwriting either.
10. **Rejected alternatives:** persisting all raw points solely to make counts authoritative and silently normalizing metadata to the observed slice length.
11. **Trade-offs:** legacy/incomplete fixtures may surface new limitations.
12. **Regression tests / protection:** point-count reconciliation tests and temporal review audit.
13. **Adversarial review findings:** the builder remains bounded-storage compatible; absent point slices are handled by segment fallback rather than treated as mismatch.
14. **Remediation iterations:** `c0e3323328f81af8bf0b8841b1bf6756d3085d21`.
15. **Residual risks and limitations:** mismatch identifies inconsistency but does not determine which upstream source was wrong.
16. **Operational or deployment consequences:** improved production diagnostics without schema or database changes.
17. **Exact evidence:** commit `c0e3323328f81af8bf0b8841b1bf6756d3085d21`, point-count tests, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** whenever persisted summary metadata and detailed evidence coexist, the consumer must define an explicit reconciliation policy.
