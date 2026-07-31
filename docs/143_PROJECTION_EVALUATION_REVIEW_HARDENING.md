# Projection Evaluation Review Hardening

Status: engineering implementation complete, exact Continuous Integration and formal closure pending

```text
REVIEW_BASELINE_COMMIT=61e1696b16e39f49a3850530312555c3593acfc5
ACCEPTED_CRITICAL_FINDINGS=4
ACCEPTED_HIGH_OR_MEDIUM_FINDINGS=11
ACCEPTED_DOCUMENTATION_MISMATCHES=1
ACCEPTED_TEST_GAP_GROUPS=1
PARTIALLY_ACCEPTED_FINDINGS=4
REJECTED_MECHANICAL_OR_IDIOMATIC_FINDINGS=2
REPLAY_KNOWLEDGE_CUTOFF=IMPLEMENTED_PENDING_EXACT_CI
DUPLICATE_TIMESTAMP_POLICY=IMPLEMENTED_PENDING_EXACT_CI
PROJECTION_OUTPUT_FINGERPRINT_BINDING=IMPLEMENTED_PENDING_EXACT_CI
TRUTH_ALTITUDE_STATUS_LINEAGE=IMPLEMENTED_PENDING_EXACT_CI
INTERPOLATION_PLAUSIBILITY=IMPLEMENTED_PENDING_EXACT_CI
EVALUATION_POLICY_PROVENANCE=IMPLEMENTED_PENDING_EXACT_CI
ENDPOINT_METRICS=IMPLEMENTED_PENDING_EXACT_CI
LEAD_TIME_METRICS=IMPLEMENTED_PENDING_EXACT_CI
CONFIDENCE_COMPARISON=IMPLEMENTED_PENDING_EXACT_CI
AGGREGATION_IDENTITY=IMPLEMENTED_PENDING_EXACT_CI
UNAVAILABLE_ACCURACY_ISOLATION=IMPLEMENTED_PENDING_EXACT_CI
ARRIVAL_SELECTIVE_PREDICTION_ACCOUNTING=IMPLEMENTED_PENDING_EXACT_CI
MICRO_AND_MACRO_AGGREGATION=IMPLEMENTED_PENDING_EXACT_CI
STATISTICAL_MEDIAN=IMPLEMENTED_PENDING_EXACT_CI
DERIVED_METRIC_RECOMPUTATION=IMPLEMENTED_PENDING_EXACT_CI
AGGREGATE_INPUT_FINGERPRINT=IMPLEMENTED_PENDING_EXACT_CI
ACTUAL_ARRIVAL_ICAO_VALIDATION=IMPLEMENTED_PENDING_EXACT_CI
PERMANENT_REVIEW_AUDIT=ENFORCED_PENDING_EXACT_CI
ENGINEERING_IMPLEMENTATION=COMPLETE_PENDING_EXACT_CI
OPEN_CONFIRMED_FINDINGS=0_PENDING_EXACT_CI
UNCLASSIFIED_FINDINGS=0_PENDING_EXACT_CI
DEFERRED_FINDINGS=0_PENDING_EXACT_CI
ADDITIONAL_CODE_FIXES_REQUIRED=NO_PENDING_EXACT_CI
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES
PROJECTION_EVALUATION_REVIEW_STATUS=OPEN_PENDING_EXACT_CI_AND_FORMAL_CLOSURE
```

## 1. Scope

This record covers only:

```text
apps/api/internal/projectionintelligence/projectionevaluation
```

It also covers the minimum integration surfaces required to keep the evaluation
contract usable:

```text
apps/api/internal/analytics/formulabenchmark/evaluator_test.go
apps/api/tools/projectionevaluationreviewaudit
.github/workflows/backend-ci.yml
docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md
docs/DOCUMENT_INDEX.md
```

It does not close `projectionproduction`, `projectionread`, or the final
end-to-end project reconciliation.

## 2. Review decisions

The review correctly identified fingerprint collisions, input-order-dependent truth,
insufficient replay-knowledge provenance, physically unbounded interpolation,
aggregation identity collisions, unavailable-result contamination, selective arrival
prediction bias, an incorrectly named median, and missing endpoint, lead-time, and
confidence-comparison metrics.

The review was only partially accepted where it treated micro-averaging as inherently
incorrect. Point micro-averages are retained because they answer a valid question, but
they are now accompanied by trajectory macro averages and lead-time buckets.

The following mechanical rules were rejected:

```text
blanket prohibition of names containing With
blanket prohibition of optional pointer fields
```

`WithinHorizontalUncertainty` and `WithinVerticalUncertainty` are precise domain
predicates. Optional altitude and arrival fields preserve the distinction between
missing, non-applicable, false, and numeric zero.

## 3. Strict replay knowledge cutoff

`TrackPoint4D` does not contain ingestion or retrieval time. The evaluator therefore
requires explicit point-level evidence:

```go
TruthAvailability{
    PointID,
    SourceName,
    AvailableAt,
}
```

A trajectory point can be used only when both conditions hold:

```text
ObservedAt <= EvaluatedAt
AvailableAt <= EvaluatedAt
```

Missing, duplicate, malformed, or late availability evidence cannot silently become
truth. The published truth evidence summary declares:

```text
KnowledgeCutoffMode=point_availability_evidence
```

Actual arrival evidence also contains `AvailableAt` and must have been available by
the evaluation cutoff.

## 4. Deterministic truth normalization

Truth points are normalized in UTC and ordered by observation time and point
identifier. Equal timestamps follow an explicit policy:

```text
identical canonical truth content -> deterministic deduplication
conflicting canonical truth content -> ErrAmbiguousTruthTimestamp
```

Input order can no longer change the selected factual point.

The canonical truth fingerprint includes:

```text
trajectory identity
point identity
ObservedAt
AvailableAt
evidence source
latitude and longitude
geometric altitude and status
barometric altitude and status
truth source name
cutoff exclusion counts
```

## 5. Canonical projection snapshot

The evaluation fingerprint no longer relies only on an upstream projection input
fingerprint. A complete projection snapshot fingerprint binds:

```text
schema and result status
trajectory and flight identity
method name, version, and DecisionClass
as-of time, end time, and forecast step
all forecast coordinates and optional altitude
horizontal and vertical uncertainty
point confidence and reasons
result confidence and reasons
Estimated Arrival and its confidence
limitations and explanations
scope guard
provenance inputs and retrieval timestamps
GeneratedAt
```

Different forecast outputs therefore cannot share an evaluation fingerprint merely
because they were produced from the same upstream evidence.

## 6. Physical truth interpolation

Interpolation remains bounded by `MaximumInterpolationGap` and now also enforces:

```text
MaximumTruthGroundSpeedMPS
MaximumTruthVerticalRateMPS
```

An interpolation segment exceeding either configured physical rate is rejected and
reported through `implausible_truth_interpolation_rejected`.

## 7. Published evaluation policy

Each result publishes an immutable policy snapshot and policy fingerprint:

```text
projection-replay-evaluation-policy-v2
MaximumInterpolationGap
MaximumTruthGroundSpeedMPS
MaximumTruthVerticalRateMPS
MinimumEvaluatedPointCount
MaximumHorizontalErrorM
MaximumAltitudeErrorM
LeadTimeBucketSize
```

Consumers can now reconstruct the denominators used by horizontal and altitude error
ratios.

## 8. Evaluation metrics

The version 2 result adds explicit:

```text
endpoint horizontal and altitude error
lead-time buckets
mean point confidence
normalized horizontal accuracy
absolute confidence gap
confidence calibration root mean square error
truth evidence provenance
projection and truth snapshot fingerprints
```

Normalized horizontal accuracy is the bounded score:

```text
clamp(1 - HorizontalErrorM / MaximumHorizontalErrorM, 0, 1)
```

The comparison is intentionally a bounded engineering diagnostic, not a claim of
scientific probability calibration.

## 9. Derived-value integrity

`Result.Validate()` now recomputes and verifies:

```text
horizontal geodesic error
horizontal error ratio
horizontal uncertainty coverage
altitude absolute error
altitude error ratio
vertical uncertainty coverage
normalized accuracy
confidence gap
position aggregates
endpoint metrics
confidence aggregates
lead-time aggregates
arrival signed and absolute error
arrival interval width and coverage
```

A structurally plausible but arithmetically forged evaluation result is rejected
before aggregation.

## 10. Aggregation identity and weighting

Aggregation groups now include:

```text
method name
method version
DecisionClass
projection horizon duration
forecast step
evaluation policy version
evaluation policy fingerprint
```

Results with different decision semantics, horizon grids, or evaluation policies
cannot be merged.

Available accuracy summaries publish both:

```text
point micro-average
trajectory macro-average
lead-time bucket metrics
```

This keeps point-level performance while preventing long trajectories from being the
only visible weighting interpretation.

## 11. Unavailable evaluation isolation

`StatusUnavailable` remains visible in evaluation and availability counters, but its
points, errors, uncertainty coverage, endpoint error, confidence comparison, and
matched arrival accuracy do not enter accuracy distributions.

Partial and complete evaluations remain accuracy-eligible because both contain enough
truth under the published minimum policy.

## 12. Arrival availability and selective prediction

Arrival Mean Absolute Error and interval coverage remain defined only for matched
airport/time comparisons. Missing predictions are not assigned an artificial time
error. Instead the aggregate publishes:

```text
ActualArrivalTruthCount
ArrivalPredictionCount
MatchedArrivalCount
MissingArrivalPredictionCount
UnexpectedArrivalPredictionCount
ArrivalAirportMismatchCount
ArrivalPredictionRecall
ArrivalAirportAccuracy
ArrivalEvaluationCount
```

A method can no longer look accurate merely by omitting difficult arrival forecasts.

## 13. Statistical semantics

Fields named `Median...` now use the conventional median:

```text
odd count  -> middle sorted value
even count -> mean of the two central sorted values
```

P95 continues to use nearest-rank semantics.

## 14. Aggregate fingerprint

`AggregateResult.InputFingerprint` depends on:

```text
projection-replay-aggregate-fingerprint-v2
sorted evaluation input fingerprints
input multiplicity
```

`GeneratedAt` remains result metadata and no longer changes the input fingerprint.

## 15. Arrival identifier validation

Actual arrival ICAO evidence must match:

```text
^[A-Z0-9]{4}$
```

Whitespace is removed and letters are normalized to uppercase before validation.
Arbitrary four-character punctuation is rejected.

## 16. Regression coverage

The permanent regression suite covers:

```text
projection-output fingerprint mutation
altitude-status fingerprint mutation
duplicate timestamp conflict rejection
input-order-independent identical deduplication
missing availability evidence
late availability cutoff
impossible movement rejection
endpoint, lead-time, confidence, and policy publication
derived-metric tampering
conventional even-count median
DecisionClass and policy aggregation separation
unavailable accuracy isolation
arrival recall and airport mismatch accounting
aggregate fingerprint independence from GeneratedAt
strict actual-arrival ICAO validation
```

## 17. Permanent enforcement

The strict audit is:

```text
go run ./tools/projectionevaluationreviewaudit -strict
```

Backend Continuous Integration executes this audit after the Projection Arrival review
audit. The audit protects version markers, knowledge-cutoff evidence, canonical
fingerprints, interpolation plausibility, grouping identity, unavailable isolation,
selective arrival accounting, regression tests, documentation, and workflow wiring.

## 18. Remaining process work

The engineering implementation is complete locally, but the review remains open until:

```text
engineering commit is created and pushed
exact Backend Continuous Integration run succeeds
all four required jobs succeed
exact run and job identifiers are recorded
formal closure documentation is committed
formal-closure commit passes exact Backend Continuous Integration
```

No statement in this document closes the wider repository review.
