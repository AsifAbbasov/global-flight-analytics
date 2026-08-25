# Projection Arrival Review Hardening

Status: closed

```text
REVIEW_BASELINE_COMMIT=d179e6529c2ce75ac1519d49e72015622617cbc0
REVIEW_MODULE=projectionarrival
ACCEPTED_CORRECTNESS_FINDINGS=10
ACCEPTED_TEST_GAP_GROUPS=1
PARTIALLY_ACCEPTED_REFACTORING_FINDINGS=3
REJECTED_MECHANICAL_OR_IDIOMATIC_FINDINGS=5
ENGINEERING_CLOSURE_COMMIT=65311c066aebbc278b63e2d25558f79f57584ca3
ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30614617800
ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91104833141
ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91104833127
ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91104833181
ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91105067522
DIRECTIONAL_CLOSING_SPEED=CI_CONFIRMED
MAXIMUM_PHYSICAL_GROUND_SPEED=CI_CONFIRMED
LOW_SPEED_SAMPLE_PRESERVATION=CI_CONFIRMED
RADIAL_RADIUS_ENTRY_UNCERTAINTY=CI_CONFIRMED
COMPLETE_ARRIVAL_INTERVAL_BOUND=CI_CONFIRMED
CURRENT_TRAJECTORY_IDENTITY=CI_CONFIRMED
CURRENT_ENDPOINT_PROVENANCE=CI_CONFIRMED
POSITION_SAMPLE_FINGERPRINT_LINEAGE=CI_CONFIRMED
DURATION_CEILING_AND_OVERFLOW_GUARD=CI_CONFIRMED
CONFIDENCE_REASON_RECONSTRUCTION=CI_CONFIRMED
ARRIVAL_DURATION_POLICY_COHERENCE=CI_CONFIRMED
PERMANENT_REVIEW_AUDIT=CI_CONFIRMED
ENGINEERING_IMPLEMENTATION=COMPLETE
ENGINEERING_DEBT=CLOSED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_ARRIVAL_FORMAL_CLOSURE=COMPLETE
PROJECTION_ARRIVAL_REVIEW_STATUS=CLOSED
```

## Review decision

The original review correctly identified that extrapolated Estimated Arrival used the
magnitude of ground speed without the direction of movement. A trajectory moving away
from the destination could therefore receive an Estimated Arrival. The review also
correctly identified missing physical speed limits, optimistic deletion of slow
segments, an unbounded LatestTime, tangential interval compression, conditional
trajectory identity, incomplete sample lineage, implicit duration truncation, and
confidence reasons that did not reconstruct the final score.

The following findings were accepted as correctness or contract defects:

```text
direction-independent extrapolation
missing maximum physical ground speed
slow and receding sample deletion
LatestTime outside the configured duration bound
radius-entry uncertainty divided by full ground speed
empty current trajectory identifier bypass
missing current endpoint input reference
missing exact used-sample fingerprint lineage
implicit float-to-duration truncation and overflow risk
confidence contributions inconsistent with final score
minimum interval and maximum duration policy incoherence
missing focused regressions
```

The following findings were not classified as correctness defects:

```text
idiomatic nil,error constructor return
optional Arrival pointer used by the shared projection contract
mechanical prohibition of the word With
small local clamp duplication without a proven shared contract
immutable nested contract traversal labelled as Law of Demeter failure
```

Function decomposition and typed unavailable reasons were accepted as supporting
refactoring because they reduce the risk of the correctness changes. Function length
alone was not treated as a severity-bearing defect.

## Directional speed contract

Version Two computes every profile sample from two separate quantities:

```text
ground_speed = segment_distance / segment_duration
closing_speed = (previous_distance_to_destination - current_distance_to_destination) / segment_duration
```

Ground speed is used only for physical plausibility. Signed closing speed is used for
Estimated Arrival. Positive values approach the airport, zero values do not close the
remaining distance, and negative values move away.

No slow or negative closing-speed sample is silently deleted. The latest bounded
sample set is preserved in full. Extrapolation is withheld unless both the expected
closing speed and the conservative closing speed meet the configured minimum.

The Config field `MinimumGroundSpeedMPS` is retained for source compatibility but is
now interpreted by the arrival algorithm as the minimum usable positive closing
speed. `MaximumGroundSpeedMPS` is a separate physical segment limit. Zero is
normalized to the production-safe default of 400 metres per second so existing
external Config literals remain source compatible.

## Arrival interval contract

Within-horizon radius crossing uses radial closing speed for conversion of horizontal
position uncertainty into time uncertainty. Full ground speed is not used for this
conversion, preventing a nearly tangential path from publishing an artificially narrow
interval.

Extrapolated arrival builds three durations with explicit nanosecond ceiling:

```text
earliest = (remaining_distance - position_uncertainty) / optimistic_closing_speed
estimated = remaining_distance / expected_closing_speed
latest = (remaining_distance + position_uncertainty) / conservative_closing_speed
```

Every conversion rejects non-finite, negative, or overflowing values. Earliest,
estimated, and latest durations must each fit the configured maximum. The complete
interval is checked again after minimum-interval expansion, so `LatestTime` cannot
escape the maximum bound.

Config validation now requires:

```text
0 < MinimumGroundSpeedMPS < MaximumGroundSpeedMPS
0 < MinimumArrivalInterval <= MaximumEstimatedArrivalDuration
```

## Evidence identity and provenance

Current trajectory identity is mandatory and must exactly match both the projection
and route trajectory identifiers. The current endpoint is chosen canonically at the
projection as-of time.

The fingerprint now includes the complete canonical sequence of position samples
actually consumed by the arrival calculation:

```text
source class
source identifier
source name
projection sequence
UTC timestamp
latitude
longitude
horizontal uncertainty
```

Changing the identity of the used current endpoint changes both available and
speed-withheld arrival fingerprints even when aggregate speed and interval values are
unchanged.

When the current endpoint participates in the calculation, provenance publishes the
observed input `current_trajectory_arrival_endpoint` with its actual source and
observation time. The directional profile is separately classified as estimated.

## Confidence contract

Arrival confidence retains the existing weighted components but applies the exact
extrapolation-retention factor to each additive component before publication:

```text
projection contribution = projection score * projection weight * retention
destination contribution = destination score * destination weight * retention
speed contribution = directional speed stability * speed weight * retention
```

The three contributions sum to the final arrival confidence score. A fourth zero-value
reason explains that retention is already included and therefore must not be
subtracted a second time.

## Versioning

```text
METHOD_VERSION=estimated-arrival-boundary-v2
FINGERPRINT_VERSION=estimated-arrival-boundary-fingerprint-v2
UNAVAILABLE_FINGERPRINT_VERSION=estimated-arrival-unavailable-fingerprint-v2
POSITION_SAMPLE_FINGERPRINT_VERSION=estimated-arrival-position-samples-v1
DEFAULT_MAXIMUM_GROUND_SPEED_MPS=400
DURATION_ROUNDING_POLICY=CEILING_TO_NANOSECOND
```

## Regression evidence

The permanent regression suite covers:

```text
aircraft moving away from the destination
physically impossible ground speed
preservation of slow and receding closing samples
empty current trajectory identifier
current endpoint identity fingerprint drift
observed current endpoint provenance
radial uncertainty for a nearly tangential radius crossing
minimum interval pushing LatestTime beyond the maximum duration
duration ceiling and overflow rejection
confidence reason sum equal to final score
maximum ground speed validation
minimum interval versus maximum duration validation
```

## Permanent audit

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionarrivalreviewaudit
```

Backend Continuous Integration executes:

```text
go run ./tools/projectionarrivalreviewaudit -strict
```

The audit protects Version Two identities, directional speed semantics, physical speed
limits, complete interval bounds, duration ceiling, strict trajectory identity,
position-sample lineage, provenance, confidence reconstruction, regression tests,
documentation markers, and Continuous Integration wiring.

## Exact engineering-closure Continuous Integration

The engineering hardening commit passed the exact Backend Continuous Integration run
and all four mandatory jobs:

```text
COMMIT=65311c066aebbc278b63e2d25558f79f57584ca3
GITHUB_ACTIONS_RUN=30614617800
POSTGRESQL_16_INTEGRATION_JOB=91104833141
BACKEND_RACE_SAFETY_JOB=91104833127
BACKEND_QUALITY_JOB=91104833181
BACKEND_CONTAINER_JOB=91105067522
```

## Formal closure

All accepted correctness findings and the focused regression gap group are implemented
and protected by the permanent strict audit. Exact Continuous Integration evidence for
the engineering closure is recorded above. There are no open, unclassified or deferred
findings in the reviewed module scope.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_ARRIVAL_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_ARRIVAL_ENGINEERING_DEBT=CLOSED
PROJECTION_ARRIVAL_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_ARRIVAL_FORMAL_CLOSURE=COMPLETE
PROJECTION_ARRIVAL_REVIEW_STATUS=CLOSED
```

The formal-closure commit must itself pass the same four Backend Continuous Integration
jobs before an external final report is issued. That final run is a release gate for the
closure record, not an additional engineering finding.

## Canonical remediation history

The review records ten accepted correctness findings plus one focused regression-gap group. The entries below preserve that ownership rather than turning supporting refactoring or rejected idiomatic recommendations into defect IDs. Severity is retrospective.

### GFA-DATA-383 — Estimated Arrival extrapolation ignored the direction of motion

1. **Finding / symptom.** Extrapolated arrival used ground-speed magnitude without signed closing direction.
2. **Root cause.** Segment speed and destination-closing speed were represented by one unsigned quantity.
3. **Failure scenario.** An aircraft moving away from the destination receives a finite Estimated Arrival.
4. **Impact.** The system publishes an arrival prediction unsupported by actual movement toward the airport.
5. **Severity rationale.** P1 retrospective because the defect directly creates false analytical output.
6. **Existing guarantees violated.** Physical-model integrity and evidence-consistent arrival semantics.
7. **Considered solutions.** Use ground speed; use bearing threshold; compute signed radial closing speed.
8. **Chosen remediation.** Separate physical ground speed from signed destination closing speed and use only positive closing speed for extrapolation.
9. **Why selected.** It represents the actual quantity required by the arrival equation.
10. **Rejected alternatives.** Magnitude-only speed remains direction-blind; coarse bearing thresholds lose continuous radial information.
11. **Trade-offs.** Receding or non-closing trajectories intentionally receive no extrapolated arrival.
12. **Regression tests / protection.** Moving-away and positive-closing-speed regressions plus strict review audit.
13. **Adversarial review findings.** Ground speed remains useful for physical plausibility but no longer authorizes arrival by itself.
14. **Remediation iterations.** Closed in the Version Two directional-speed hardening.
15. **Residual risks / limitations.** Closing speed remains a short-horizon research extrapolation, not a flight-plan intent model.
16. **Operational/deployment consequences.** Some previously predicted arrivals become unavailable; no migration.
17. **Exact evidence.** Engineering closure `65311c066aebbc278b63e2d25558f79f57584ca3`, run `30614617800`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectionarrivalreviewaudit` protects directional closing-speed semantics and regressions.

### GFA-DATA-384 — Arrival profile lacked a maximum physical ground-speed bound

1. **Finding / symptom.** Finite but physically implausible segment speeds could enter arrival estimation.
2. **Root cause.** Only minimum usable speed semantics existed; no upper physical plausibility policy bounded source motion.
3. **Failure scenario.** Corrupt coordinates over a short interval imply extreme speed and produce an unrealistically early arrival.
4. **Impact.** Invalid source geometry can dominate ETA arithmetic.
5. **Severity rationale.** P1 retrospective because implausible telemetry directly changes arrival time.
6. **Existing guarantees violated.** Conservative physical bounds and fail-closed source validation.
7. **Considered solutions.** Clamp speed; trust upstream trajectory validation; reject above an explicit maximum.
8. **Chosen remediation.** Add `MaximumGroundSpeedMPS` with a production-safe default and enforce `minimum < maximum`.
9. **Why selected.** Rejection preserves evidence honesty and works at the Arrival boundary even for alternate inputs.
10. **Rejected alternatives.** Clamping manufactures plausible motion; upstream-only trust leaves this module unprotected.
11. **Trade-offs.** Extreme samples are rejected rather than contributing weakly.
12. **Regression tests / protection.** Maximum-speed validation and impossible-motion regressions.
13. **Adversarial review findings.** The bound is a conservative research guard, not an aircraft certification envelope.
14. **Remediation iterations.** Added with directional-speed Version Two.
15. **Residual risks / limitations.** A static maximum cannot encode all aircraft/phase-specific envelopes.
16. **Operational/deployment consequences.** Suspect profiles may withhold arrival.
17. **Exact evidence.** `65311c066aebbc278b63e2d25558f79f57584ca3`, run `30614617800`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects the maximum-speed policy and tests.

### GFA-DATA-385 — Slow and receding arrival-profile samples were silently discarded

1. **Finding / symptom.** Low-speed or negative-closing samples could be removed, leaving an optimistically selected subset.
2. **Root cause.** Eligibility filtering deleted weak/receding evidence before the estimator judged the complete bounded sample set.
3. **Failure scenario.** A mixed profile containing approach and receding segments appears steadily approaching after adverse samples are dropped.
4. **Impact.** Estimated Arrival becomes selection-biased and overconfident.
5. **Severity rationale.** P1 retrospective because evidence deletion changes whether an ETA is published.
6. **Existing guarantees violated.** Evidence preservation and conservative analytical selection.
7. **Considered solutions.** Drop weak samples; average absolute speed; preserve all bounded samples and withhold extrapolation when closing speed is inadequate.
8. **Chosen remediation.** Preserve slow and receding samples and require expected and conservative closing speeds to satisfy the minimum.
9. **Why selected.** Adverse evidence remains visible instead of being optimized away.
10. **Rejected alternatives.** Filtering created survivorship bias; absolute-speed averaging erased direction.
11. **Trade-offs.** More profiles fail closed when directional evidence is weak.
12. **Regression tests / protection.** Slow/receding sample-preservation and speed-withheld regressions.
13. **Adversarial review findings.** `MinimumGroundSpeedMPS` is retained for source compatibility but interpreted as minimum usable positive closing speed in this algorithm.
14. **Remediation iterations.** Closed in Version Two profile hardening.
15. **Residual risks / limitations.** Historical noise can still widen or suppress estimates by design.
16. **Operational/deployment consequences.** Fewer optimistic ETAs; no schema change.
17. **Exact evidence.** `65311c066aebbc278b63e2d25558f79f57584ca3`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Strict audit protects sample preservation and directional minimum semantics.

### GFA-DATA-386 — Arrival interval expansion could escape the configured maximum-duration contract

1. **Finding / symptom.** `LatestTime` could end beyond the configured maximum after uncertainty/minimum-interval expansion, and minimum interval could be incoherent with maximum duration.
2. **Root cause.** Duration bounds were checked before every later interval transformation was complete and configuration allowed contradictory limits.
3. **Failure scenario.** An initially valid ETA is widened until its latest bound exceeds the maximum estimated-arrival duration.
4. **Impact.** Published result violates the evaluator's own horizon contract.
5. **Severity rationale.** P1 retrospective because a public time interval can contradict configured policy.
6. **Existing guarantees violated.** Policy coherence and complete result-bound validation.
7. **Considered solutions.** Clamp `LatestTime`; validate only estimated time; require coherent config and revalidate the complete interval after expansion.
8. **Chosen remediation.** Require `MinimumArrivalInterval <= MaximumEstimatedArrivalDuration` and recheck earliest/estimated/latest after minimum-interval expansion.
9. **Why selected.** It prevents impossible policy states and avoids silently clipping published uncertainty.
10. **Rejected alternatives.** Clamping hides policy contradiction; checking only midpoint does not constrain the interval.
11. **Trade-offs.** Some wide-uncertainty arrivals become unavailable.
12. **Regression tests / protection.** Minimum-interval overflow and policy-coherence regressions.
13. **Adversarial review findings.** Latest bound is part of the contract, not auxiliary display metadata.
14. **Remediation iterations.** Closed during interval-contract hardening.
15. **Residual risks / limitations.** Maximum duration remains a policy choice, not empirical arrival certainty.
16. **Operational/deployment consequences.** Out-of-policy intervals fail closed.
17. **Exact evidence.** `65311c066aebbc278b63e2d25558f79f57584ca3`, run `30614617800`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects complete interval bounds and config ordering.

### GFA-DATA-387 — Radius-entry uncertainty used total ground speed instead of radial closing speed

1. **Finding / symptom.** Horizontal position uncertainty was converted to time uncertainty by full ground speed for within-horizon radius crossing.
2. **Root cause.** Tangential motion and radial progress toward the airport were conflated.
3. **Failure scenario.** A fast nearly tangential trajectory receives an artificially narrow arrival interval despite little radial approach.
4. **Impact.** Arrival uncertainty is understated.
5. **Severity rationale.** P1 retrospective because the published interval materially misrepresents geometric evidence.
6. **Existing guarantees violated.** Geometry-consistent uncertainty and conservative ETA bounds.
7. **Considered solutions.** Full speed; projected radial speed; fixed time margin.
8. **Chosen remediation.** Use radial closing speed for radius-entry uncertainty conversion.
9. **Why selected.** It matches the distance component relevant to crossing the destination radius.
10. **Rejected alternatives.** Full speed overstates useful progress; fixed margins ignore actual geometry.
11. **Trade-offs.** Tangential paths receive wider intervals.
12. **Regression tests / protection.** Nearly tangential radius-crossing regression.
13. **Adversarial review findings.** This is separate from extrapolated directional-speed logic because it affects within-horizon uncertainty conversion.
14. **Remediation iterations.** Closed in arrival interval hardening.
15. **Residual risks / limitations.** Position uncertainty remains modeled as a scalar horizontal radius.
16. **Operational/deployment consequences.** Some arrival windows widen conservatively.
17. **Exact evidence.** `65311c066aebbc278b63e2d25558f79f57584ca3`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit requires radial uncertainty semantics and regression coverage.

### GFA-DATA-388 — Estimated Arrival allowed an empty or inconsistent current trajectory identity

1. **Finding / symptom.** Empty current trajectory ID could bypass exact identity reconciliation with Projection and Route evidence.
2. **Root cause.** Trajectory identity was treated as conditionally optional at the Arrival boundary.
3. **Failure scenario.** Arrival combines projection/route evidence for one trajectory with an unowned or differently identified current trajectory.
4. **Impact.** ETA provenance can attach to the wrong analytical target.
5. **Severity rationale.** P1 retrospective because target identity is foundational to all arrival evidence.
6. **Existing guarantees violated.** Entity ownership and cross-module lineage.
7. **Considered solutions.** Allow empty as wildcard; fill from projection; require exact non-empty identity across inputs.
8. **Chosen remediation.** Require current trajectory identity and exact match with both Projection and Route trajectory identifiers.
9. **Why selected.** It prevents fallback identity fabrication and cross-target mixing.
10. **Rejected alternatives.** Wildcards weaken lineage; copying identity from another input hides missing evidence.
11. **Trade-offs.** Legacy incomplete inputs now fail validation.
12. **Regression tests / protection.** Empty and mismatch trajectory-ID regressions.
13. **Adversarial review findings.** Optional Arrival output in the shared contract remains valid; the defect was optional target identity, not optional result presence.
14. **Remediation iterations.** Closed in evidence-identity hardening.
15. **Residual risks / limitations.** Correctness depends on upstream IDs being canonical and truthful.
16. **Operational/deployment consequences.** Invalid mixed-target arrival requests fail closed.
17. **Exact evidence.** `65311c066aebbc278b63e2d25558f79f57584ca3`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects mandatory trajectory identity reconciliation.

### GFA-DATA-389 — Arrival provenance omitted the observed current endpoint used by the calculation

1. **Finding / symptom.** The current trajectory endpoint could participate in ETA calculation without an explicit observed input reference.
2. **Root cause.** Directional profile provenance was published more readily than the concrete current endpoint evidence.
3. **Failure scenario.** Consumers can inspect estimated profile evidence but cannot identify the exact observed endpoint anchoring it.
4. **Impact.** Arrival output is not independently traceable to all contributing observations.
5. **Severity rationale.** P1 retrospective because provenance completeness is a core research guarantee.
6. **Existing guarantees violated.** Source attribution and reproducible input lineage.
7. **Considered solutions.** Publish aggregate profile only; publish endpoint ID only; add a full observed input reference with source and observation time.
8. **Chosen remediation.** Publish `current_trajectory_arrival_endpoint` as an observed input with actual source and timestamp.
9. **Why selected.** It exposes the exact observation rather than a derived summary.
10. **Rejected alternatives.** Aggregate-only provenance cannot reconstruct the anchor.
11. **Trade-offs.** Provenance payload becomes slightly richer.
12. **Regression tests / protection.** Observed current-endpoint provenance regression.
13. **Adversarial review findings.** The directional profile remains separately classified as estimated, preserving observed/derived distinction.
14. **Remediation iterations.** Closed in evidence/provenance hardening.
15. **Residual risks / limitations.** Upstream source naming still determines provenance precision.
16. **Operational/deployment consequences.** Metadata-only expansion; no migration.
17. **Exact evidence.** `65311c066aebbc278b63e2d25558f79f57584ca3`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit requires the observed endpoint input reference.

### GFA-DATA-390 — Arrival fingerprint did not bind the exact sequence of position samples consumed

1. **Finding / symptom.** Different used endpoint/sample identities could produce the same Arrival input identity when aggregate speed/interval values matched.
2. **Root cause.** Fingerprinting emphasized derived values rather than the complete canonical sample lineage.
3. **Failure scenario.** Replacing the current endpoint with another observation that yields equal aggregates leaves the arrival fingerprint unchanged.
4. **Impact.** Distinct evidence sets become indistinguishable in stored or compared analytical results.
5. **Severity rationale.** P1 retrospective because fingerprint collisions break reproducibility and provenance.
6. **Existing guarantees violated.** Semantic identity and exact input lineage.
7. **Considered solutions.** Hash aggregates; hash endpoint ID; hash the full canonical used-sample sequence.
8. **Chosen remediation.** Add `estimated-arrival-position-samples-v1` binding source class/ID/name, sequence, UTC time, coordinates and horizontal uncertainty.
9. **Why selected.** It identifies the evidence actually consumed by the algorithm.
10. **Rejected alternatives.** Aggregate or endpoint-only identity misses multi-sample substitutions.
11. **Trade-offs.** Fingerprints intentionally change when any used sample evidence changes.
12. **Regression tests / protection.** Current-endpoint identity drift and sample-lineage fingerprint regressions.
13. **Adversarial review findings.** Available and speed-withheld paths both receive evidence-sensitive identity.
14. **Remediation iterations.** Closed with Version Two fingerprint hardening.
15. **Residual risks / limitations.** Fingerprint fidelity depends on the canonical sample representation remaining versioned.
16. **Operational/deployment consequences.** Historical fingerprints change by design; no schema migration required by this review.
17. **Exact evidence.** `65311c066aebbc278b63e2d25558f79f57584ca3`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects position-sample fingerprint version and mutation regressions.

### GFA-DATA-391 — Float-to-duration conversion could truncate ETA and overflow time.Duration

1. **Finding / symptom.** Floating ETA durations could be truncated implicitly and converted without explicit overflow protection.
2. **Root cause.** Numerical seconds/nanoseconds crossed the duration boundary without a named rounding and range policy.
3. **Failure scenario.** A large finite duration wraps/overflows or a fractional duration is rounded earlier than the conservative bound expects.
4. **Impact.** Arrival timestamps can be incorrect or violate interval ordering.
5. **Severity rationale.** P1 retrospective because time arithmetic corruption directly affects published ETA.
6. **Existing guarantees violated.** Numeric safety and deterministic temporal policy.
7. **Considered solutions.** Truncate; round-to-nearest; ceiling with finite/range checks.
8. **Chosen remediation.** Use explicit nanosecond ceiling and reject non-finite, negative or overflowing values.
9. **Why selected.** Ceiling preserves conservative temporal bounds and makes conversion policy explicit.
10. **Rejected alternatives.** Truncation can understate time; unchecked casts can overflow.
11. **Trade-offs.** ETA may be rounded upward by less than one nanosecond-equivalent conversion unit.
12. **Regression tests / protection.** Duration-ceiling and overflow regressions.
13. **Adversarial review findings.** Earliest, estimated and latest durations are all independently bounded.
14. **Remediation iterations.** Closed in Version Two duration hardening.
15. **Residual risks / limitations.** Upstream distance/speed arithmetic still requires finite-value guards, which remain enforced.
16. **Operational/deployment consequences.** Invalid extreme durations fail closed.
17. **Exact evidence.** `65311c066aebbc278b63e2d25558f79f57584ca3`, run `30614617800`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects `CEILING_TO_NANOSECOND` and overflow regressions.

### GFA-DATA-392 — Arrival confidence reasons did not reconstruct the published confidence score

1. **Finding / symptom.** Published reason contributions could disagree with final confidence because extrapolation retention was applied outside the additive reason values.
2. **Root cause.** Score composition and explanation composition were separate calculations.
3. **Failure scenario.** Consumers sum projection/destination/speed reason contributions and obtain a value different from result confidence.
4. **Impact.** Confidence becomes non-auditable and explanations can misstate the decision mathematics.
5. **Severity rationale.** P1 retrospective because confidence is public analytical evidence.
6. **Existing guarantees violated.** Explainability and derived-value integrity.
7. **Considered solutions.** Keep reasons descriptive; add a separate retention subtraction; apply retention identically to each additive contribution.
8. **Chosen remediation.** Multiply each weighted component by the exact retention factor before publishing it; reasons then sum to final score.
9. **Why selected.** One arithmetic representation owns both the score and its explanation.
10. **Rejected alternatives.** Descriptive-only reasons are not reconstructable; double-applying retention would understate confidence.
11. **Trade-offs.** Reason semantics are stricter and version-sensitive.
12. **Regression tests / protection.** Confidence-reason sum equality regression.
13. **Adversarial review findings.** A fourth zero-value reason explicitly states that retention is already included and must not be subtracted twice.
14. **Remediation iterations.** Closed in confidence-contract hardening.
15. **Residual risks / limitations.** The component weights themselves remain a research policy rather than calibrated probability weights.
16. **Operational/deployment consequences.** Explanation values change to match the existing intended score semantics.
17. **Exact evidence.** `65311c066aebbc278b63e2d25558f79f57584ca3`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects reason reconstruction and its regression.

### GFA-TEST-393 — Projection Arrival regression coverage did not protect the accepted correctness invariants

1. **Finding / symptom.** Focused regressions were missing for directional motion, physical speed, sample preservation, interval bounds, identity, provenance, fingerprint, duration safety and confidence reconstruction.
2. **Root cause.** Existing tests exercised nominal Arrival behavior without one suite mirroring the review's cross-field invariants.
3. **Failure scenario.** A future refactor reintroduces one of the accepted correctness defects while ordinary happy-path tests remain green.
4. **Impact.** Closure claims can silently drift from executable protection.
5. **Severity rationale.** P2 retrospective because this is a protection gap around multiple P1 defects.
6. **Existing guarantees violated.** Regression durability and evidence-based closure.
7. **Considered solutions.** Manual review; broad integration tests only; focused invariant regressions plus strict source audit.
8. **Chosen remediation.** Add the dedicated regression matrix recorded in this review and wire `projectionarrivalreviewaudit -strict` into Backend CI.
9. **Why selected.** It tests the exact failure modes and keeps documentation/workflow markers synchronized.
10. **Rejected alternatives.** Manual-only review is non-reproducible; broad tests do not prove each invariant.
11. **Trade-offs.** Contract changes require coordinated test/audit updates.
12. **Regression tests / protection.** The permanent suite listed in `Regression evidence` and strict audit.
13. **Adversarial review findings.** Mechanical naming/style recommendations remain outside this test-gap finding.
14. **Remediation iterations.** Completed with engineering closure and permanent audit wiring.
15. **Residual risks / limitations.** Static audit markers supplement but do not replace behavioral regression tests.
16. **Operational/deployment consequences.** CI becomes stricter; no runtime behavior added by this finding itself.
17. **Exact evidence.** `65311c066aebbc278b63e2d25558f79f57584ca3`, Backend CI run `30614617800`, jobs `91104833141`, `91104833127`, `91104833181`, `91105067522`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectionarrivalreviewaudit` remains mandatory in Backend Continuous Integration.
