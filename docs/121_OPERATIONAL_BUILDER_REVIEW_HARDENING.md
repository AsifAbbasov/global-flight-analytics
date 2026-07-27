# Document 121 — Operational Builder Review Hardening

Status: implementation and formal review-closure candidate
Baseline: `1bbfd0147092baf2615f5bb0838ca12768b54846`
Scope: `apps/api/internal/features/operationalbuilder` and its production trajectory-point hydration boundary

## 1. Confirmed findings

The current operational feature path contained real production and semantic defects:

- aggregate trajectory reads restored parent metadata, segments and coverage gaps, but not the underlying flight-state points;
- `FlightState` preserved nullable telemetry availability, while `TrackPoint4D` discarded it;
- points were neither filtered to the trajectory window nor ordered deterministically before heading aggregation;
- out-of-range finite headings were normalized into evidence;
- ground and airborne shares treated every record as an available boolean observation;
- ground altitude status could erase a conflicting non-zero source value;
- barometric and geometric altitude observations could be mixed in one mean;
- heading transitions bridged unavailable or invalid measurements;
- group support used raw point-record count rather than the union of points that supplied a usable final measurement;
- nil contexts were replaced with background contexts and aggregate passes were not cancellation-aware.

## 2. Findings already resolved elsewhere or deliberately rejected

`AsOfTime` remains an extractor-owned cross-builder gate. The extractor rejects point, segment and coverage-gap timestamps after the request cutoff before any builder executes. The operational builder owns the narrower trajectory-window rule `[StartTime, EndTime]`; duplicating `AsOfTime` in every builder would recreate the distributed temporal policy criticized by the review.

PostgreSQL did not universally convert nullable telemetry into observed zero values. `FlightStateRepository` already scans nullable columns into `pgtype` values and records explicit availability flags. The confirmed loss occurred at the `FlightState` to `TrackPoint4D` conversion and is corrected there.

The complete upstream `DataQuality` object is not copied into every point. Operational aggregation independently validates the raw value, its availability and its domain status. Copying the entire validation report into trajectory points would couple the feature layer to one ingestion-validation representation without adding a missing decision input.

Means remain observation-weighted. Time weighting would imply interpolation across intervals and gaps that are not observed. The schema and policy now state the retained semantics explicitly.

Cumulative heading change remains observation-sequence dependent by definition. The corrected metric is chronological, shortest-arc and gap-aware; unavailable or invalid headings terminate a contiguous run. A frequency-normalized turn metric would be a separate schema feature rather than a silent replacement.

## 3. Production point hydration

Ordinary `TrajectoryRepository` reads deliberately retain their existing parent, segment and coverage-gap payload size. Feature materialization now uses an explicit `FeatureTrajectoryReader`. That reader opens the existing read-only `REPEATABLE READ` boundary and, inside one snapshot, loads:

1. trajectory parent metadata;
2. trajectory segments and coverage gaps through the existing private aggregate reader;
3. trajectory points reconstructed from `flight_states` inside the trajectory window.

This avoids silently adding raw point hydration to every ordinary trajectory consumer. The point query prefers stable `flight_id` identity and falls back to canonical ICAO24 plus the trajectory window when a flight identifier is unavailable. Rows are ordered by `(observed_at, id)`. Latitude and longitude remain required for a trajectory point, while velocity, heading, vertical rate and on-ground state remain nullable.

No raw-point duplication table or PostgreSQL migration is introduced. Existing `flight_states` are the authoritative telemetry source.

## 4. Telemetry availability contract

`TrackPoint4D` now carries:

```text
VelocityAvailable
HeadingAvailable
VerticalRateAvailable
OnGroundAvailable
TelemetryAvailabilityKnown
```

Legacy in-memory fixtures with `TelemetryAvailabilityKnown=false` retain their former interpretation. PostgreSQL-reconstructed points set the flag to true and preserve each SQL `NULL` as unavailable. A present numerical zero or boolean false remains an available measurement.

The canonical extractor fingerprint mirrors every new point field, preventing availability-only evidence changes from colliding under one input fingerprint.

## 5. Operational aggregation policies

```text
Operational Builder version:
  operational-feature-builder-v2

Processing Pipeline version:
  flight-feature-processing-pipeline-v10

Aggregation policy:
  observation-weighted-kahan-v1

Altitude source policy:
  single-source-prefer-barometric-v1

Heading policy:
  chronological-shortest-arc-contiguous-valid-runs-v1
```

Altitude aggregation selects one source for the full trajectory: any usable barometric series wins; geometric altitude is used only when no usable barometric series exists. Source mixing is disclosed and excluded.

Heading evidence must be finite and lie in `[0, 360)`. Invalid and unavailable headings break continuity and cannot create a direct transition between surrounding observations.

Ground and airborne shares use only points with explicit on-ground availability. Their denominator therefore cannot contain SQL-null values fabricated as airborne observations.

`SupportingPointCount` is the union count of temporally eligible points that contribute at least one measurement to the final selected operational evidence. Exact per-signal omission counts are carried in typed limitations; the version-one schema is not expanded with eleven additional diagnostic counters.

## 6. Numerical and cancellation policy

Means and cumulative heading change use compensated summation. A non-finite aggregate is not published and produces an explicit limitation. Long collection and aggregation loops periodically observe `context.Context` cancellation. Nil contexts are rejected.

## 7. Versioning and persistence

Operational output semantics change, so the builder advances from generation one to generation two and the processing pipeline advances from generation nine to generation ten. The schema remains `flight-features-v1`; no persisted field was added to the feature payload. Existing snapshot identity includes processing generation, preventing generation-ten output from colliding with generation-nine output.

No PostgreSQL migration is required.

## 8. Permanent regression gates

Tests and the strict review audit require:

- production trajectory reads to hydrate points inside the same repeatable-read snapshot;
- SQL-null operational values to remain unavailable while legitimate zero and false values remain available;
- TrackPoint and canonical fingerprint availability mirrors;
- trajectory-window filtering and deterministic ordering;
- permutation-stable operational output;
- strict heading range and gap termination;
- single-source altitude aggregation;
- explicit ground-state share denominator;
- strict nil-context and mid-loop cancellation;
- compensated finite aggregation;
- processing-generation alignment across all active tests and audits.

## 9. Closure record

```text
OPERATIONAL_PRODUCTION_POINT_HYDRATION=ENFORCED
OPERATIONAL_NULL_TELEMETRY_SEMANTICS=PRESERVED
OPERATIONAL_TRACK_POINT_AVAILABILITY=ENFORCED
OPERATIONAL_POINT_WINDOW_FILTERING=ENFORCED
OPERATIONAL_POINT_CHRONOLOGICAL_ORDER=ENFORCED
OPERATIONAL_POINT_PERMUTATION_STABILITY=ENFORCED
OPERATIONAL_INVALID_HEADING_NORMALIZATION=CLOSED
OPERATIONAL_HEADING_GAP_BRIDGING=CLOSED
OPERATIONAL_GROUND_SHARE_DENOMINATOR=EXPLICIT_AVAILABLE_STATES
OPERATIONAL_GROUND_ALTITUDE_CONFLICT=REJECTED
OPERATIONAL_ALTITUDE_SOURCE_MIXING=CLOSED
OPERATIONAL_SUPPORTING_POINT_COUNT=USABLE_UNION
OPERATIONAL_AGGREGATION_POLICY=OBSERVATION_WEIGHTED_KAHAN_V1
OPERATIONAL_NIL_CONTEXT=REJECTED
OPERATIONAL_LOOP_CANCELLATION=ENFORCED
OPERATIONAL_BUILDER_GENERATION=v2
OPERATIONAL_BUILDER_PROCESSING_GENERATION=v10
DATABASE_MIGRATION=NOT_REQUIRED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
OPERATIONAL_BUILDER_REVIEW_STATUS=CLOSED
```

Formal closure still requires successful local installation validation and exact-commit GitHub Actions evidence for Backend Quality, Backend Race Safety, PostgreSQL 16 Integration and Backend Container.
