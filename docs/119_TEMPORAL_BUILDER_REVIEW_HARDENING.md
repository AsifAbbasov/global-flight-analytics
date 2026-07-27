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
