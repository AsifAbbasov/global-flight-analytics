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
