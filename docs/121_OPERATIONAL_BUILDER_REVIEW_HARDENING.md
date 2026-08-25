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

## 10. Canonical remediation history

Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, and Continuous Integration evidence. Historical exact-commit GitHub Actions evidence for the original implementation commit is not reconstructed here; current permanent audits provide repository-state regression evidence. Severity labels are retrospective.

### GFA-DB-186 — Production feature materialization did not hydrate operational point evidence
1. **Finding / symptom:** aggregate trajectory reads restored parent, segments and gaps but omitted flight-state points needed for Operational features.
2. **Root cause:** ordinary lightweight trajectory read semantics were reused for a feature-materialization use case with stronger evidence needs.
3. **Failure scenario:** production Operational Builder receives no points while the authoritative `flight_states` table contains usable telemetry.
4. **Impact:** false unavailability and divergence between in-memory/test feature behavior and production materialization.
5. **Severity rationale:** **P1 retrospective** because production analytical output could omit existing evidence systematically.
6. **Existing guarantees violated:** production reachability of implemented Operational features and repeatable-read snapshot consistency.
7. **Considered solutions:** hydrate points in every trajectory read, add a raw-point duplication table, or create an explicit feature-only reader.
8. **Chosen remediation:** `FeatureTrajectoryReader` hydrates `flight_states` inside the existing read-only `REPEATABLE READ` snapshot.
9. **Why selected:** preserves lightweight ordinary reads and reuses the authoritative telemetry table without duplicate storage.
10. **Rejected alternatives:** global read expansion and a new raw-point persistence table.
11. **Trade-offs:** feature materialization performs additional point queries by design.
12. **Regression tests / protection:** PostgreSQL feature-reader integration tests and `operationalbuilderreviewaudit`.
13. **Adversarial review findings:** raw-point duplication and builder-level `AsOfTime` were rejected as unnecessary architecture expansion/duplication.
14. **Remediation iterations:** `0b5ec52b503f1ef65c2ca5eeaba485e381710649`.
15. **Residual risks and limitations:** reconstruction depends on retained `flight_states`; bounded historical retention outside this contract limits future re-materialization.
16. **Operational or deployment consequences:** no migration; feature materialization performs point hydration only on its dedicated path.
17. **Exact evidence:** implementation commit, PostgreSQL reader tests, permanent operational review audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every production feature must have explicit evidence-loading reachability tests against the real persistence boundary.

### GFA-DATA-187 — `FlightState` telemetry availability was lost when converted to `TrackPoint4D`
1. **Finding / symptom:** nullable velocity, heading, vertical-rate and on-ground availability flags did not survive the domain conversion.
2. **Root cause:** `TrackPoint4D` stored values but not availability semantics.
3. **Failure scenario:** SQL `NULL` becomes an apparently observed numerical zero or boolean false in operational aggregation.
4. **Impact:** fabricated measurements and corrupted analytical evidence.
5. **Severity rationale:** **P1 retrospective** because missing telemetry could be published as real observations.
6. **Existing guarantees violated:** end-to-end nullable telemetry integrity and legitimate-zero semantics.
7. **Considered solutions:** copy full upstream DataQuality, sentinel values, or explicit per-signal availability flags.
8. **Chosen remediation:** add availability flags plus `TelemetryAvailabilityKnown`, preserve them in conversion and fingerprint identity.
9. **Why selected:** retains exactly the decision inputs needed by feature aggregation without coupling to the full ingestion validation object.
10. **Rejected alternatives:** sentinel zeros and copying complete `DataQuality` into every point.
11. **Trade-offs:** point contract is wider and compatibility behavior for legacy in-memory fixtures is explicit.
12. **Regression tests / protection:** conversion, nullable PostgreSQL, legitimate zero/false, fingerprint mirror tests.
13. **Adversarial review findings:** review confirmed the loss at `FlightState → TrackPoint4D`, not universally in PostgreSQL scanning.
14. **Remediation iterations:** `0b5ec52b503f1ef65c2ca5eeaba485e381710649`.
15. **Residual risks and limitations:** legacy points with unknown availability retain compatibility semantics intentionally.
16. **Operational or deployment consequences:** no DB migration; reconstructed points now distinguish absent from zero/false.
17. **Exact evidence:** implementation commit and availability/fingerprint regression tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** nullable telemetry additions must update domain conversion, availability methods, canonical fingerprint mirrors, and integration fixtures together.

### GFA-DATA-188 — Operational points lacked trajectory-window filtering and deterministic ordering
1. **Finding / symptom:** aggregation could consume points outside `[StartTime, EndTime]` and depend on input ordering.
2. **Root cause:** point collection did not establish one canonical evidence sequence before aggregation.
3. **Failure scenario:** stale/outside-window records influence metrics or permutation changes cumulative heading output.
4. **Impact:** temporally invalid and non-reproducible Operational features.
5. **Severity rationale:** **P1 retrospective** because durable feature values could depend on record ordering rather than evidence.
6. **Existing guarantees violated:** trajectory-window ownership and deterministic replay.
7. **Considered solutions:** rely on SQL ordering/window only, sort per metric, or canonicalize once in the builder.
8. **Chosen remediation:** filter to trajectory window and deterministic `(observed_at,id)` order before aggregation.
9. **Why selected:** defense in depth protects in-memory and PostgreSQL callers identically.
10. **Rejected alternatives:** duplicating global `AsOfTime`; Extractor remains historical-cutoff owner.
11. **Trade-offs:** out-of-window points are explicitly omitted and diagnosed.
12. **Regression tests / protection:** permutation/window tests and audit.
13. **Adversarial review findings:** observation weighting retained; time weighting would imply unobserved interpolation.
14. **Remediation iterations:** implementation commit `0b5ec52b...`.
15. **Residual risks and limitations:** observation frequency still influences observation-weighted metrics by design.
16. **Operational or deployment consequences:** corrected semantics isolated by processing v10.
17. **Exact evidence:** `0b5ec52b503f1ef65c2ca5eeaba485e381710649`, ordering/window tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every sequence-dependent operational metric must consume the same canonical point ordering.

### GFA-DATA-189 — Out-of-range finite headings were normalized into apparently valid evidence
1. **Finding / symptom:** finite values outside `[0,360)` could be normalized rather than rejected.
2. **Root cause:** normalization was used where domain validity should have been checked first.
3. **Failure scenario:** malformed upstream heading `725°` becomes an ordinary valid angle and contributes to turn metrics.
4. **Impact:** bad source data is hidden and analytical heading metrics are corrupted.
5. **Severity rationale:** **P1 retrospective** because invalid evidence was silently transformed into valid-looking evidence.
6. **Existing guarantees violated:** raw evidence validity must fail closed before aggregation.
7. **Considered solutions:** modulo-normalize, clamp, or reject invalid headings.
8. **Chosen remediation:** require finite heading in `[0,360)`; invalid observations are omitted with limitations.
9. **Why selected:** preserves evidence honesty rather than guessing source intent.
10. **Rejected alternatives:** modulo normalization and clamping.
11. **Trade-offs:** malformed headings reduce availability instead of contributing a guessed value.
12. **Regression tests / protection:** invalid-heading boundary tests and review audit.
13. **Adversarial review findings:** cumulative heading remains sequence-dependent by design but only over valid contiguous observations.
14. **Remediation iterations:** `0b5ec52b...`.
15. **Residual risks and limitations:** valid but inaccurate upstream headings cannot be detected without additional source-quality evidence.
16. **Operational or deployment consequences:** none beyond stricter evidence exclusion.
17. **Exact evidence:** implementation commit and heading validity tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** normalize representation only after validating the source domain range.

### GFA-DATA-190 — Ground/airborne share denominator treated unavailable booleans as observed false
1. **Finding / symptom:** all point records could enter ground/airborne share denominator even when `on_ground` was unavailable.
2. **Root cause:** boolean value and boolean availability were not separated in aggregation.
3. **Failure scenario:** SQL-null on-ground state becomes a false/airborne observation and biases both shares.
4. **Impact:** fabricated operational state proportions.
5. **Severity rationale:** **P1 retrospective** because missing evidence was counted as a real categorical observation.
6. **Existing guarantees violated:** nullable telemetry integrity and honest denominator semantics.
7. **Considered solutions:** count all records, impute unknown as airborne, or use only explicitly available on-ground observations.
8. **Chosen remediation:** denominator contains only points with `OnGroundAvailable`; zero/false remains valid when available.
9. **Why selected:** mathematically honest and consistent with telemetry availability contract.
10. **Rejected alternatives:** imputation and raw-record denominator.
11. **Trade-offs:** denominator may be smaller than total point count and can be unavailable.
12. **Regression tests / protection:** null/false/true denominator tests and Validator reconciliation.
13. **Adversarial review findings:** full DataQuality copy was unnecessary because availability is the required decision input.
14. **Remediation iterations:** `0b5ec52b...`.
15. **Residual risks and limitations:** source-specific on-ground accuracy remains upstream.
16. **Operational or deployment consequences:** corrected share semantics under v10.
17. **Exact evidence:** implementation commit, operational share tests, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** categorical ratios require explicit availability-aware denominators.

### GFA-DATA-191 — Ground altitude status could erase conflicting non-zero altitude evidence
1. **Finding / symptom:** a ground status could cause a conflicting non-zero altitude measurement to be treated as ordinary ground-zero evidence.
2. **Root cause:** status interpretation overrode the raw value without checking consistency.
3. **Failure scenario:** source reports `on_ground`/ground altitude status with materially non-zero altitude and aggregation silently substitutes/accepts zero semantics.
4. **Impact:** contradictory altitude evidence is hidden.
5. **Severity rationale:** **P1 retrospective** because conflicting source evidence could be silently rewritten.
6. **Existing guarantees violated:** contradictory evidence must be rejected/disclosed, not normalized away.
7. **Considered solutions:** trust status, trust value, clamp to zero, or reject the conflicting observation.
8. **Chosen remediation:** conflicting ground/non-zero altitude evidence is excluded with an explicit limitation.
9. **Why selected:** fail-closed evidence semantics avoid choosing an arbitrary source signal.
10. **Rejected alternatives:** forced zero and status/value precedence guesses.
11. **Trade-offs:** fewer altitude samples when source signals conflict.
12. **Regression tests / protection:** ground-altitude conflict tests and audit.
13. **Adversarial review findings:** altitude policy is intentionally explicit and versioned.
14. **Remediation iterations:** `0b5ec52b...`.
15. **Residual risks and limitations:** small sensor noise thresholds remain governed by existing altitude/status contracts.
16. **Operational or deployment consequences:** corrected evidence under processing v10.
17. **Exact evidence:** implementation commit and altitude conflict tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** status/value pairs must define and test contradiction handling explicitly.

### GFA-DATA-192 — Barometric and geometric altitude observations could be mixed into one mean
1. **Finding / symptom:** altitude aggregation could combine two different measurement sources in one statistic.
2. **Root cause:** source selection occurred per observation rather than for the aggregate series.
3. **Failure scenario:** part of a flight contributes barometric altitude and another part geometric altitude, yielding a mean with no single measurement semantics.
4. **Impact:** analytically ambiguous altitude features.
5. **Severity rationale:** **P1 retrospective** because a published metric could mix incomparable evidence sources silently.
6. **Existing guarantees violated:** one feature must have one declared measurement-source policy.
7. **Considered solutions:** mix all available, expose two new metrics, interpolate, or select one source for the complete trajectory.
8. **Chosen remediation:** `single-source-prefer-barometric-v1`: any usable barometric series wins; geometric is fallback only when barometric is wholly unusable.
9. **Why selected:** preserves existing schema while giving the metric one reproducible source meaning.
10. **Rejected alternatives:** silent source mixing and schema expansion without product need.
11. **Trade-offs:** some geometric observations are ignored when any usable barometric series exists.
12. **Regression tests / protection:** mixed-source, fallback and source-policy tests.
13. **Adversarial review findings:** policy remains observation-weighted; time interpolation was rejected.
14. **Remediation iterations:** `0b5ec52b...`.
15. **Residual risks and limitations:** barometric/geometric biases remain inherent to source measurements.
16. **Operational or deployment consequences:** no migration; processing v10 isolates semantics.
17. **Exact evidence:** implementation commit, altitude source tests, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** multi-source measurements require a named deterministic source-selection policy.

### GFA-DATA-193 — Heading change bridged unavailable or invalid observations
1. **Finding / symptom:** cumulative heading change could connect valid headings across a missing/invalid measurement gap.
2. **Root cause:** unavailable observations were skipped without terminating the contiguous heading run.
3. **Failure scenario:** headings before and after an evidence gap create a turn transition that was never observed continuously.
4. **Impact:** fabricated turn activity.
5. **Severity rationale:** **P1 retrospective** because the metric could claim movement across absent evidence.
6. **Existing guarantees violated:** sequence-derived movement metrics require contiguous valid observations.
7. **Considered solutions:** bridge gaps, interpolate, reset continuity, or drop metric entirely.
8. **Chosen remediation:** invalid/unavailable heading terminates the run; shortest-arc change is computed only inside contiguous valid runs.
9. **Why selected:** honest evidence with no invented interpolation.
10. **Rejected alternatives:** gap bridging and frequency-normalized replacement metric.
11. **Trade-offs:** cumulative heading change can decrease when evidence has gaps, reflecting lower observation support.
12. **Regression tests / protection:** gap-termination, shortest-arc, invalid-heading tests.
13. **Adversarial review findings:** observation-sequence dependency is retained as the metric's intended meaning.
14. **Remediation iterations:** `0b5ec52b...`.
15. **Residual risks and limitations:** uneven sampling still affects an observation-sequence metric by design.
16. **Operational or deployment consequences:** processing v10 output change only.
17. **Exact evidence:** implementation commit and heading continuity tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** sequence transitions must specify what breaks continuity and test gaps explicitly.

### GFA-DATA-194 — Operational `SupportingPointCount` counted raw records instead of usable contributing observations
1. **Finding / symptom:** group support reflected input record count, including points that contributed no final measurement.
2. **Root cause:** evidence support was measured before availability/validity/source-selection filtering.
3. **Failure scenario:** many unusable points make a sparse feature group appear strongly supported.
4. **Impact:** overstated evidence strength and misleading completeness context.
5. **Severity rationale:** **P2 retrospective** because values remain protected but their support evidence is overstated.
6. **Existing guarantees violated:** support count must represent observations that actually contribute to published evidence.
7. **Considered solutions:** raw count, per-signal counters in schema, minimum count, or union of points contributing any final selected measurement.
8. **Chosen remediation:** usable-union support count plus typed per-signal omission limitations.
9. **Why selected:** honest group-level support without expanding schema by many diagnostic counters.
10. **Rejected alternatives:** raw record count and eleven new permanent diagnostic fields.
11. **Trade-offs:** one group count cannot express per-signal support distribution; limitations provide that context.
12. **Regression tests / protection:** usable-union support tests and audit.
13. **Adversarial review findings:** diagnostic schema expansion rejected as unnecessary.
14. **Remediation iterations:** `0b5ec52b...`.
15. **Residual risks and limitations:** consumers needing exact per-signal counts must inspect limitations or future dedicated diagnostics.
16. **Operational or deployment consequences:** corrected evidence metadata only.
17. **Exact evidence:** implementation commit and supporting-count tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** support metrics must be computed after final evidence eligibility decisions.

### GFA-OPS-195 — Operational aggregation substituted nil contexts and lacked complete cancellation/finite-aggregation guards
1. **Finding / symptom:** nil contexts became background contexts and long scans did not consistently stop on cancellation; aggregate arithmetic also needed an explicit non-finite fail-closed rule.
2. **Root cause:** lifecycle/numerical guards were not applied uniformly across collection and aggregation passes.
3. **Failure scenario:** cancelled feature materialization continues CPU work or a non-finite aggregate is published.
4. **Impact:** wasted resources and potentially invalid numeric output.
5. **Severity rationale:** **P2 retrospective** because guards protect reliability and mathematical fail-closed behavior; upstream invalid inputs are separately validated.
6. **Existing guarantees violated:** explicit context ownership, bounded cancellation, finite analytical outputs.
7. **Considered solutions:** background fallback, entry-only checks, ordinary summation, or strict context + periodic checks + compensated finite aggregation.
8. **Chosen remediation:** reject nil, poll cancellation, use Kahan sums, and convert non-finite aggregate results into explicit limitations rather than publication.
9. **Why selected:** aligns builder behavior with the repository-wide lifecycle/numerical policy.
10. **Rejected alternatives:** implicit background work and unchecked ordinary aggregation.
11. **Trade-offs:** small polling/compensation overhead.
12. **Regression tests / protection:** nil/cancel/non-finite/compensated aggregation tests and audit.
13. **Adversarial review findings:** no generic abstraction was required; guards remain close to operational semantics.
14. **Remediation iterations:** `0b5ec52b503f1ef65c2ca5eeaba485e381710649`.
15. **Residual risks and limitations:** observation weighting remains deliberate and can reflect sampling density.
16. **Operational or deployment consequences:** more predictable cancellation and finite output guarantees.
17. **Exact evidence:** implementation commit, context/numeric tests, `operationalbuilderreviewaudit`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** large aggregation passes must combine explicit lifecycle checks with finite-output assertions.
