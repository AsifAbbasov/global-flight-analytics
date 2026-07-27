# Document 122 — Trajectory Builder Review Hardening

Status: Review closure candidate
Baseline commit: `0b5ec52b503f1ef65c2ca5eeaba485e381710649`
Scope: `apps/api/internal/features/trajectorybuilder` and the minimum quality and validation contracts required for consistent output

## 1. Review classification

The original static review was based on stale commit `a1689dc`. Its claim that production feature materialization never loads point records is no longer current: Operational Builder hardening introduced `FeatureTrajectoryReader`, which hydrates flight-state points only for feature materialization inside the existing repeatable-read snapshot.

The following findings remained confirmed on the current baseline:

- trajectory count semantics preferred loaded slice length without a persisted-metadata fallback;
- sampling and path used different point order and eligibility rules;
- duplicate timestamps contributed zero sampling intervals;
- point evidence was not restricted to the trajectory window;
- point and segment paths bridged declared discontinuities;
- one point blocked better segment fallback;
- empty evidence claimed eleven base fields plus a zero quality score;
- coverage could report one without point or segment observation evidence;
- zero duration was handled before excluding outside-window gaps;
- zero gap-duration metadata escaped mismatch reporting;
- fractional duration policy was undocumented;
- path ratios were unconditionally clamped;
- context cancellation was absent from several large scans;
- unknown segment statuses produced one limitation per record.

## 2. Canonical evidence contract

Trajectory Builder generation two creates one canonical evidence view before calculating any field:

1. validate the trajectory start and end window;
2. reject points with missing timestamps;
3. reject points outside the authoritative trajectory window;
4. sort by observation time and deterministic point identity fields;
5. collapse duplicate timestamps to one deterministic observation;
6. share that sequence between point count, sampling and point-path calculations;
7. sort segments and coverage gaps without mutating caller-owned slices.

`AsOfTime` remains owned by Extractor snapshot validation. Builders continue to enforce their own trajectory-window membership and do not duplicate the global historical cutoff contract.

## 3. Point-count and quality reconciliation

When point records are materialized, `trajectory.point_count` is the number of unique temporally eligible canonical observations. Metadata disagreement is disclosed.

When point records are intentionally unavailable but persisted `PointCount` is positive, the persisted value is retained as a fallback count and supporting-point count. Sampling and point-path fields remain unavailable rather than being fabricated.

Extractor initial quality now derives supporting-point count exclusively from built group evidence. It no longer independently reintroduces raw trajectory metadata after group builders have canonicalized evidence. Validator and Extractor therefore reconcile against the same values.

## 4. Sampling policy

The policy is:

```text
unique-chronological-observation-instants-kahan-v1
```

Duplicate timestamps are collapsed before interval calculation. Sampling mean uses compensated summation and maximum gap uses the same unique chronological sequence.

## 5. Coverage policy

The policy is:

```text
non-invalid-segment-evidence-plus-clipped-gap-union-v1
```

Positive-duration coverage requires materialized non-invalid segment evidence. A zero-duration instant requires one canonical point. Persisted point-count metadata or absence of gaps alone is insufficient. Valid gaps are clipped to the trajectory window and merged without double counting. Zero-duration windows are evaluated only after outside-window gaps are excluded.

Gap metadata uses the shared `truncate_fractional_seconds_toward_zero` policy. Subsecond gaps still contribute their full nanosecond-resolution duration to the coverage ratio; only the integer metadata comparison is truncated.

## 6. Path policy

The policy is:

```text
continuous-parts-no-gap-or-segment-discontinuity-bridging-v1
```

Coverage gaps split point paths. Invalid point coordinates also split continuity. Segment fallback joins only exactly contiguous segment endpoints into one path part. Coordinate mismatches, invalid segments, missing temporal evidence and declared gaps split parts, so distance between disconnected segments is never called observed movement.

Path efficiency sums direct distance and observed distance within each continuous part. Ratio values are clamped only inside an explicit numerical tolerance. Larger violations become unavailable evidence with a limitation.

Both Trajectory and Geographical builders use the documented mean-Earth spherical Haversine model. They retain separate implementations because their evidence semantics differ: Geographical Builder owns envelopes and geographical movement, while Trajectory Builder owns segmented path efficiency.

## 7. Completeness semantics

Field availability is counted explicitly from zero. Empty input is unavailable. Zero segment counts may support zero status counts when an authoritative trajectory envelope exists, but segment shares are undefined for a zero denominator and are not counted as available fields. A zero quality score is accepted only when trajectory evidence exists.

## 8. Validation alignment

Validator generation five skips the geographical-to-trajectory path ratio comparison when Trajectory Builder explicitly reports:

- coverage-gap path segmentation;
- segment fallback;
- insufficient path evidence;
- zero or non-finite path distance;
- an out-of-range ratio.

This avoids treating two intentionally different evidence models as contradictory.

## 9. Version boundary

```text
trajectory-feature-builder-v2
flight-feature-validator-v5
flight-feature-processing-pipeline-v11
flight-features-v1
```

No PostgreSQL migration is required.

## 10. Closure markers

```text
TRAJECTORY_CANONICAL_EVIDENCE=ENFORCED
TRAJECTORY_POINT_WINDOW_FILTERING=ENFORCED
TRAJECTORY_POINT_CHRONOLOGICAL_ORDER=ENFORCED
TRAJECTORY_DUPLICATE_TIMESTAMP_POLICY=UNIQUE_DETERMINISTIC
TRAJECTORY_POINT_COUNT_FALLBACK=PERSISTED_METADATA_WHEN_UNMATERIALIZED
TRAJECTORY_QUALITY_SUPPORT_RECONCILIATION=GROUP_EVIDENCE_OWNED
TRAJECTORY_SAMPLING_POLICY=UNIQUE_CHRONOLOGICAL_KAHAN_V1
TRAJECTORY_COVERAGE_REQUIRES_OBSERVATION_EVIDENCE=ENFORCED
TRAJECTORY_COVERAGE_GAP_UNION=CLIPPED_AND_DEDUPLICATED
TRAJECTORY_GAP_DURATION_POLICY=TRUNCATE_FRACTIONAL_SECONDS_TOWARD_ZERO
TRAJECTORY_PATH_GAP_BRIDGING=CLOSED
TRAJECTORY_PATH_SEGMENT_FALLBACK=INDEPENDENT_PARTS
TRAJECTORY_PATH_RATIO_TOLERANCE=ENFORCED
TRAJECTORY_EMPTY_COMPLETENESS=NOT_INFLATED
TRAJECTORY_NIL_CONTEXT=REJECTED
TRAJECTORY_LOOP_CANCELLATION=ENFORCED
TRAJECTORY_BUILDER_GENERATION=v2
TRAJECTORY_VALIDATOR_GENERATION=v5
TRAJECTORY_BUILDER_PROCESSING_GENERATION=v11
DATABASE_MIGRATION=NOT_REQUIRED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
TRAJECTORY_BUILDER_REVIEW_STATUS=CLOSED
```
