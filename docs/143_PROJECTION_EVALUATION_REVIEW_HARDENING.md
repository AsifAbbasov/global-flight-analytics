# Projection Evaluation Review Hardening

Status: closed

```text
REVIEW_BASELINE_COMMIT=61e1696b16e39f49a3850530312555c3593acfc5
ENGINEERING_CLOSURE_COMMIT=279d60543bbbb8c204fab60e442f00a56d1f3bbe
ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30619973772
ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91121986195
ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91121986123
ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91121986134
ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91122255083
ACCEPTED_CRITICAL_FINDINGS=4
ACCEPTED_HIGH_OR_MEDIUM_FINDINGS=11
ACCEPTED_DOCUMENTATION_MISMATCHES=1
ACCEPTED_TEST_GAP_GROUPS=1
PARTIALLY_ACCEPTED_FINDINGS=4
REJECTED_MECHANICAL_OR_IDIOMATIC_FINDINGS=2
REPLAY_KNOWLEDGE_CUTOFF=CI_CONFIRMED
DUPLICATE_TIMESTAMP_POLICY=CI_CONFIRMED
PROJECTION_OUTPUT_FINGERPRINT_BINDING=CI_CONFIRMED
TRUTH_ALTITUDE_STATUS_LINEAGE=CI_CONFIRMED
INTERPOLATION_PLAUSIBILITY=CI_CONFIRMED
EVALUATION_POLICY_PROVENANCE=CI_CONFIRMED
ENDPOINT_METRICS=CI_CONFIRMED
LEAD_TIME_METRICS=CI_CONFIRMED
CONFIDENCE_COMPARISON=CI_CONFIRMED
AGGREGATION_IDENTITY=CI_CONFIRMED
UNAVAILABLE_ACCURACY_ISOLATION=CI_CONFIRMED
ARRIVAL_SELECTIVE_PREDICTION_ACCOUNTING=CI_CONFIRMED
MICRO_AND_MACRO_AGGREGATION=CI_CONFIRMED
STATISTICAL_MEDIAN=CI_CONFIRMED
DERIVED_METRIC_RECOMPUTATION=CI_CONFIRMED
AGGREGATE_INPUT_FINGERPRINT=CI_CONFIRMED
ACTUAL_ARRIVAL_ICAO_VALIDATION=CI_CONFIRMED
PERMANENT_REVIEW_AUDIT=CI_CONFIRMED
ENGINEERING_IMPLEMENTATION=COMPLETE
ENGINEERING_DEBT=CLOSED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_EVALUATION_FORMAL_CLOSURE=COMPLETE
PROJECTION_EVALUATION_REVIEW_STATUS=CLOSED
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

## 18. Formal closure evidence

The engineering implementation commit is:

```text
279d60543bbbb8c204fab60e442f00a56d1f3bbe
```

The exact push-triggered Backend Continuous Integration evidence is:

```text
run: 30619973772
Backend Race Safety: 91121986123 — success
Backend Quality: 91121986134 — success
PostgreSQL 16 Integration: 91121986195 — success
Backend Container: 91122255083 — success
```

The Backend Quality job executed the permanent Projection Evaluation review audit. The
engineering debt, documentation mismatch, confirmed findings, unclassified findings,
and deferred findings recorded by this review are closed.

The formal-closure commit must pass the same four Backend Continuous Integration jobs
before the external final closure report is issued.

No statement in this document closes `projectionproduction`, `projectionread`, or the
wider repository review.

## Canonical remediation history

The following records reconcile the accepted Evaluation review into fifteen engineering findings plus one documentation mismatch and one regression-gap group. Several CI-confirmed markers (for example altitude-status lineage) are owned by the nearest root-cause record rather than duplicated as standalone findings. Severity is retrospective.

### GFA-DATA-394 — Replay truth could include evidence that was not yet available at the evaluation cutoff

1. **Finding / symptom.** Observation time alone could admit truth that had not been ingested or otherwise available when the replay was evaluated.
2. **Root cause.** `TrackPoint4D` did not carry availability time and the evaluator lacked a separate knowledge-cutoff contract.
3. **Failure scenario.** A point observed before T but first available after T is used as factual truth for an evaluation at T.
4. **Impact.** Historical evaluation benefits from future knowledge and overstates model quality.
5. **Severity rationale.** P1 retrospective because this invalidates replay truth and benchmark credibility.
6. **Existing guarantees violated.** Temporal correctness, no-future-knowledge replay, provenance.
7. **Considered solutions.** Use observation time only; infer retrieval time; require explicit point-level availability evidence.
8. **Chosen remediation.** Require `TruthAvailability{PointID, SourceName, AvailableAt}` and enforce both `ObservedAt <= EvaluatedAt` and `AvailableAt <= EvaluatedAt`.
9. **Why selected.** It separates event time from knowledge time without fabricating unavailable metadata.
10. **Rejected alternatives.** Observation-only truth leaks future knowledge; inferred retrieval time is not evidence.
11. **Trade-offs.** Missing or late availability evidence makes truth unusable.
12. **Regression tests / protection.** Missing/late availability, truth-source, altitude-status lineage and actual-arrival availability regressions.
13. **Adversarial review findings.** Geometric/barometric altitude status is fingerprinted with truth so missing/unavailable altitude cannot silently become numeric zero.
14. **Remediation iterations.** Closed in Evaluation Version Two replay-knowledge hardening.
15. **Residual risks / limitations.** Historical accuracy depends on availability evidence being captured truthfully upstream.
16. **Operational/deployment consequences.** Some historical evaluations become unavailable rather than using future knowledge.
17. **Exact evidence.** Engineering closure `279d60543bbbb8c204fab60e442f00a56d1f3bbe`, run `30619973772`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectionevaluationreviewaudit` protects point-availability evidence and cutoff semantics.

### GFA-DATA-395 — Equal-timestamp truth selection depended on input order or ambiguous content

1. **Finding / symptom.** Multiple truth points at the same observation time could produce order-dependent factual selection.
2. **Root cause.** Equal timestamps had no semantic duplicate/conflict policy beyond collection ordering.
3. **Failure scenario.** Reordering identical input changes which conflicting point is treated as truth.
4. **Impact.** Evaluation output is non-deterministic and can compare forecasts against arbitrary facts.
5. **Severity rationale.** P1 retrospective because factual truth must be deterministic.
6. **Existing guarantees violated.** Determinism, truth integrity, reproducibility.
7. **Considered solutions.** Preserve input order; choose lexical first; deduplicate identical content and reject conflicts.
8. **Chosen remediation.** Normalize UTC, order by time and point ID, deterministically deduplicate identical canonical content, and return `ErrAmbiguousTruthTimestamp` for conflicts.
9. **Why selected.** It distinguishes harmless duplicate evidence from contradictory facts.
10. **Rejected alternatives.** Input/lexical selection is deterministic only syntactically, not semantically.
11. **Trade-offs.** Conflicting equal-time datasets cannot be evaluated.
12. **Regression tests / protection.** Conflict rejection and input-order-independent identical deduplication tests.
13. **Adversarial review findings.** Canonical content includes altitude value/status and source evidence, not coordinates alone.
14. **Remediation iterations.** Closed in truth-normalization hardening.
15. **Residual risks / limitations.** Upstream duplicate IDs still require canonical point semantics.
16. **Operational/deployment consequences.** Ambiguous truth fails closed.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects duplicate/conflict policy and regression tests.

### GFA-DATA-396 — Evaluation identity did not bind the complete Projection output snapshot

1. **Finding / symptom.** Evaluations could share identity when upstream projection input fingerprints matched even though actual forecast outputs differed.
2. **Root cause.** Evaluation fingerprinting trusted upstream input identity instead of hashing the complete result being evaluated.
3. **Failure scenario.** Two methods produce different coordinates/uncertainty/confidence from the same source evidence yet collide in evaluation identity.
4. **Impact.** Distinct predictions become indistinguishable in benchmark and aggregate evidence.
5. **Severity rationale.** P1 retrospective because fingerprint collision corrupts analytical provenance.
6. **Existing guarantees violated.** Semantic identity and reproducibility.
7. **Considered solutions.** Keep upstream fingerprint; hash selected metrics; hash canonical projection snapshot.
8. **Chosen remediation.** Bind schema/status, identities, method/decision class, grid, all forecast values, uncertainty, confidence reasons, arrival, limitations, provenance and `GeneratedAt` into a projection snapshot fingerprint.
9. **Why selected.** It identifies the exact output under evaluation rather than merely its upstream inputs.
10. **Rejected alternatives.** Partial metric hashing leaves collision paths.
11. **Trade-offs.** Any output-relevant change intentionally changes evaluation identity.
12. **Regression tests / protection.** Projection-output and altitude-status fingerprint mutation regressions.
13. **Adversarial review findings.** Optional altitude presence/status is part of identity so unavailable and zero are not conflated.
14. **Remediation iterations.** Closed in Version Two canonical snapshot hardening.
15. **Residual risks / limitations.** Snapshot schema changes require explicit fingerprint versioning.
16. **Operational/deployment consequences.** Evaluation fingerprints change by design.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`, run `30619973772`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects projection snapshot coverage and mutation tests.

### GFA-DATA-397 — Truth interpolation lacked physical rate bounds

1. **Finding / symptom.** Truth could be interpolated across segments implying impossible horizontal or vertical motion.
2. **Root cause.** Maximum interpolation gap existed without matching physical speed/rate constraints.
3. **Failure scenario.** Corrupt truth points close in time create an interpolated factual position that no plausible aircraft movement supports.
4. **Impact.** Forecast error is measured against fabricated physical truth.
5. **Severity rationale.** P1 retrospective because bad truth directly invalidates accuracy metrics.
6. **Existing guarantees violated.** Physical plausibility and benchmark integrity.
7. **Considered solutions.** Gap-only guard; clamp movement; reject segments above explicit ground/vertical limits.
8. **Chosen remediation.** Add `MaximumTruthGroundSpeedMPS` and `MaximumTruthVerticalRateMPS` and reject implausible segments.
9. **Why selected.** It keeps interpolation conservative and auditable.
10. **Rejected alternatives.** Clamping manufactures truth; gap-only checks allow high-rate corruption.
11. **Trade-offs.** Some otherwise temporally close points cannot be interpolated.
12. **Regression tests / protection.** Impossible-movement rejection regressions.
13. **Adversarial review findings.** Rejection is reported explicitly as `implausible_truth_interpolation_rejected`.
14. **Remediation iterations.** Closed in physical truth hardening.
15. **Residual risks / limitations.** Static research limits do not encode aircraft-specific performance envelopes.
16. **Operational/deployment consequences.** Suspect truth segments are excluded.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects both rate bounds and tests.

### GFA-DATA-398 — Evaluation results omitted the effective policy required to interpret their denominators

1. **Finding / symptom.** Error ratios and eligibility decisions could be published without an immutable snapshot of the thresholds that created them.
2. **Root cause.** Runtime evaluator configuration was not first-class result provenance.
3. **Failure scenario.** Two evaluations with different interpolation/error/minimum-count policies appear comparable without exposing the policy difference.
4. **Impact.** Consumers cannot independently reconstruct or group evaluation semantics.
5. **Severity rationale.** P1 retrospective because policy changes alter both eligibility and metric meaning.
6. **Existing guarantees violated.** Provenance completeness and reproducibility.
7. **Considered solutions.** Rely on deployment config; publish version only; publish full normalized policy plus fingerprint.
8. **Chosen remediation.** Publish immutable `projection-replay-evaluation-policy-v2` values and a policy fingerprint.
9. **Why selected.** It makes denominator and plausibility semantics self-contained.
10. **Rejected alternatives.** Deployment config is mutable external state; version-only identity cannot reconstruct thresholds.
11. **Trade-offs.** Results carry more policy metadata and aggregation becomes stricter.
12. **Regression tests / protection.** Policy publication and mutation regressions.
13. **Adversarial review findings.** Policy fingerprint is also part of aggregation identity, preventing silent cross-policy merges.
14. **Remediation iterations.** Closed in Version Two policy-provenance hardening.
15. **Residual risks / limitations.** Policy values remain engineering thresholds rather than statistical calibration claims.
16. **Operational/deployment consequences.** Results from different policies separate cleanly.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects policy snapshot/fingerprint fields.

### GFA-DATA-399 — Evaluation omitted explicit endpoint accuracy metrics

1. **Finding / symptom.** Aggregate path accuracy could hide how the projection performed at its final forecast endpoint.
2. **Root cause.** Metrics focused on point collections without a dedicated terminal-point contract.
3. **Failure scenario.** A method performs well early but poorly at the terminal horizon and still appears strong in aggregate means.
4. **Impact.** A key forecast-quality characteristic is invisible.
5. **Severity rationale.** P2 retrospective because the existing metrics were not false but materially incomplete.
6. **Existing guarantees violated.** Complete analytical observability of forecast behavior.
7. **Considered solutions.** Infer endpoint from last bucket; keep only aggregate error; publish explicit endpoint horizontal/altitude error.
8. **Chosen remediation.** Add endpoint horizontal and altitude metrics and validate their derivation.
9. **Why selected.** It makes terminal performance explicit and independently checkable.
10. **Rejected alternatives.** Aggregate means and bucket inference do not provide one canonical endpoint metric.
11. **Trade-offs.** Result schema expands.
12. **Regression tests / protection.** Endpoint metric publication and tamper-rejection tests.
13. **Adversarial review findings.** Endpoint availability follows the same truth/altitude evidence rules as other metrics.
14. **Remediation iterations.** Added in Evaluation Version Two.
15. **Residual risks / limitations.** A single endpoint metric does not replace full trajectory error distributions.
16. **Operational/deployment consequences.** Additional analytics only.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit and validator reconstruction protect endpoint metrics.

### GFA-DATA-400 — Evaluation omitted lead-time-specific accuracy behavior

1. **Finding / symptom.** Overall accuracy did not show how error changes as forecast lead time increases.
2. **Root cause.** Points were aggregated without a canonical lead-time bucket dimension.
3. **Failure scenario.** Strong near-term performance masks rapid degradation at longer horizons.
4. **Impact.** Model comparison lacks a basic horizon-sensitivity view.
5. **Severity rationale.** P2 retrospective because interpretation was materially incomplete rather than mathematically false.
6. **Existing guarantees violated.** Decision-useful evaluation coverage.
7. **Considered solutions.** One global average; per-point raw export; fixed policy-defined lead-time buckets.
8. **Chosen remediation.** Publish lead-time buckets using the policy's `LeadTimeBucketSize` and validate their aggregates.
9. **Why selected.** It provides deterministic comparable horizon slices without unbounded raw-output expansion.
10. **Rejected alternatives.** Global averages hide degradation; raw-only output shifts canonical aggregation to every consumer.
11. **Trade-offs.** Bucket size becomes part of policy and grouping identity.
12. **Regression tests / protection.** Lead-time publication and aggregate reconstruction tests.
13. **Adversarial review findings.** Lead-time buckets complement rather than replace point micro and trajectory macro metrics.
14. **Remediation iterations.** Added in Version Two.
15. **Residual risks / limitations.** Bucket boundaries are an engineering choice.
16. **Operational/deployment consequences.** More detailed evaluation payloads.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Policy snapshot and audit protect bucket semantics.

### GFA-DATA-401 — Evaluation lacked explicit confidence-versus-accuracy comparison metrics

1. **Finding / symptom.** Projection confidence was published but not compared systematically with realized accuracy.
2. **Root cause.** Evaluation treated spatial error and confidence as separate outputs.
3. **Failure scenario.** A method remains overconfident while ordinary position-error summaries do not expose the mismatch.
4. **Impact.** Consumers cannot assess whether confidence tracks outcome quality even as an engineering diagnostic.
5. **Severity rationale.** P2 retrospective because the missing comparison limits interpretation rather than fabricating truth.
6. **Existing guarantees violated.** Confidence accountability and evaluation completeness.
7. **Considered solutions.** Ignore confidence; publish raw confidence only; add bounded normalized accuracy, confidence gap and calibration RMSE diagnostics.
8. **Chosen remediation.** Add mean point confidence, normalized horizontal accuracy, absolute confidence gap and confidence-calibration RMSE.
9. **Why selected.** It exposes disagreement between claimed confidence and observed outcome while avoiding a scientific calibration claim.
10. **Rejected alternatives.** Raw confidence alone provides no realized comparison.
11. **Trade-offs.** Normalized accuracy depends on the published maximum-error policy.
12. **Regression tests / protection.** Confidence metric publication and recomputation tests.
13. **Adversarial review findings.** Documentation explicitly calls this a bounded engineering diagnostic, not probability calibration.
14. **Remediation iterations.** Added in Evaluation Version Two.
15. **Residual risks / limitations.** The diagnostic is policy-normalized and should not be interpreted as a calibrated probability score.
16. **Operational/deployment consequences.** Additional benchmark metrics only.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Validator and audit protect confidence-comparison arithmetic.

### GFA-DATA-402 — Evaluation validation trusted derived metrics instead of recomputing their mathematics

1. **Finding / symptom.** Structurally valid results could contain forged or inconsistent error, coverage, endpoint, confidence, lead-time or arrival-derived values.
2. **Root cause.** Validation emphasized range/shape checks without reconstructing all published arithmetic.
3. **Failure scenario.** A caller mutates error ratios or aggregate values while keeping them finite and in range.
4. **Impact.** Invalid benchmark results can be accepted and aggregated.
5. **Severity rationale.** P1 retrospective because this compromises the integrity of the evaluation output itself.
6. **Existing guarantees violated.** Derived-value integrity and trust-boundary validation.
7. **Considered solutions.** Range checks; spot-check selected metrics; fully recompute deterministic derived values.
8. **Chosen remediation.** `Result.Validate()` recomputes the complete documented derived metric set.
9. **Why selected.** Deterministic arithmetic should have one independently checkable definition.
10. **Rejected alternatives.** Range checks cannot detect coordinated plausible tampering.
11. **Trade-offs.** Validation performs additional computation and must evolve with schema semantics.
12. **Regression tests / protection.** Derived-metric tampering tests across position, endpoint, confidence, lead time and arrival fields.
13. **Adversarial review findings.** Partial and complete statuses remain accuracy-eligible only when the minimum truth policy is satisfied.
14. **Remediation iterations.** Closed in Version Two validation hardening.
15. **Residual risks / limitations.** Non-deterministic external claims cannot be recomputed and therefore require provenance instead.
16. **Operational/deployment consequences.** Forged/inconsistent results fail closed before aggregation.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects recomputation coverage and tamper regressions.

### GFA-DATA-403 — Evaluation aggregates could merge results with incompatible method, decision, horizon or policy identity

1. **Finding / symptom.** Aggregation grouping was too weak to separate semantically different evaluation results.
2. **Root cause.** Group keys omitted one or more of method version, DecisionClass, horizon grid and evaluation policy identity.
3. **Failure scenario.** Results from different decision semantics or thresholds are averaged into one benchmark row.
4. **Impact.** Aggregate metrics become analytically meaningless despite individually valid inputs.
5. **Severity rationale.** P1 retrospective because incompatible evidence is combined into a false summary.
6. **Existing guarantees violated.** Homogeneous aggregation and semantic identity.
7. **Considered solutions.** Group by method name only; group by selected fields; require complete documented identity tuple.
8. **Chosen remediation.** Group by method name/version, DecisionClass, horizon duration, forecast step and evaluation policy version/fingerprint.
9. **Why selected.** Every grouping field corresponds to output or metric semantics.
10. **Rejected alternatives.** Partial grouping leaves collision paths across policy/horizon changes.
11. **Trade-offs.** More benchmark groups are produced instead of broad pooled summaries.
12. **Regression tests / protection.** DecisionClass and policy aggregation-separation tests.
13. **Adversarial review findings.** `GeneratedAt` is intentionally not a grouping semantic.
14. **Remediation iterations.** Closed in aggregation identity hardening.
15. **Residual risks / limitations.** New semantic dimensions must be added when future schema changes affect comparability.
16. **Operational/deployment consequences.** Historical benchmark grouping may become more granular.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects the aggregation identity tuple.

### GFA-DATA-404 — Point micro-averages were the only visible weighting interpretation

1. **Finding / symptom.** Long trajectories could dominate available accuracy summaries when only point-level micro-averages were exposed.
2. **Root cause.** The evaluator had one legitimate weighting view but no complementary trajectory-level view.
3. **Failure scenario.** One very long trajectory contributes most points and masks weak performance across many shorter trajectories.
4. **Impact.** Aggregate interpretation is biased toward observation-rich trajectories.
5. **Severity rationale.** P2 retrospective because micro-averaging itself is valid, but presenting it alone is materially incomplete.
6. **Existing guarantees violated.** Balanced evaluation interpretation.
7. **Considered solutions.** Replace micro with macro; keep micro only; publish both micro and trajectory macro plus lead-time buckets.
8. **Chosen remediation.** Retain point micro-average and add trajectory macro-average and lead-time metrics.
9. **Why selected.** Both weighting questions are valid and should be explicit rather than one silently replacing the other.
10. **Rejected alternatives.** Declaring micro inherently wrong was rejected; macro-only would lose point-weighted behavior.
11. **Trade-offs.** More aggregate fields must be understood by consumers.
12. **Regression tests / protection.** Micro/macro aggregation regressions.
13. **Adversarial review findings.** This finding preserves the partially accepted nature of the original recommendation.
14. **Remediation iterations.** Closed by complementary metrics rather than deleting existing semantics.
15. **Residual risks / limitations.** Neither weighting alone captures every portfolio-level comparison question.
16. **Operational/deployment consequences.** Additional benchmark summaries only.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects coexistence of micro, macro and lead-time views.

### GFA-DATA-405 — Unavailable evaluations contaminated accuracy distributions

1. **Finding / symptom.** `StatusUnavailable` results could contribute points/errors/coverage/confidence or matched-arrival values to accuracy summaries.
2. **Root cause.** Availability accounting and accuracy eligibility were not cleanly separated.
3. **Failure scenario.** A method with unavailable truth contributes default/partial values that shift error distributions.
4. **Impact.** Aggregate accuracy can improve or degrade for reasons unrelated to evaluated predictions.
5. **Severity rationale.** P1 retrospective because denominators and distributions become false.
6. **Existing guarantees violated.** Denominator integrity and missing-data semantics.
7. **Considered solutions.** Drop unavailable results entirely; include them as zero accuracy; count availability but exclude them from accuracy distributions.
8. **Chosen remediation.** Keep unavailable results in evaluation/availability counts while excluding all accuracy-derived fields.
9. **Why selected.** It preserves visibility of coverage without fabricating error values.
10. **Rejected alternatives.** Dropping hides availability failure; zero scoring invents a metric for unevaluated truth.
11. **Trade-offs.** Availability and accuracy denominators differ explicitly.
12. **Regression tests / protection.** Unavailable accuracy-isolation regressions.
13. **Adversarial review findings.** Partial/complete results remain eligible when they satisfy the published minimum truth policy.
14. **Remediation iterations.** Closed in aggregation eligibility hardening.
15. **Residual risks / limitations.** Consumers must inspect both availability and accuracy metrics for a complete picture.
16. **Operational/deployment consequences.** Aggregate distributions change to exclude unevaluable cases.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects unavailable isolation and counter semantics.

### GFA-DATA-406 — Arrival evaluation could reward selective prediction and accepted weak actual-arrival identifiers

1. **Finding / symptom.** Arrival MAE on matched predictions alone could make a method look accurate by omitting difficult arrivals, while actual airport evidence accepted arbitrary four-character strings.
2. **Root cause.** Prediction availability/mismatch accounting and airport-identifier validation were incomplete at the arrival truth boundary.
3. **Failure scenario.** A method predicts only easy cases, or malformed punctuation-based airport evidence is treated as a valid location indicator.
4. **Impact.** Arrival performance and airport-match accuracy are overstated or evaluated against invalid truth identifiers.
5. **Severity rationale.** P1 retrospective because benchmark denominators and truth identity are both affected.
6. **Existing guarantees violated.** Selective-prediction transparency and arrival truth integrity.
7. **Considered solutions.** Assign artificial errors to missing predictions; ignore missing predictions; publish explicit availability/mismatch counters and validate canonical ICAO evidence.
8. **Chosen remediation.** Add actual/predicted/matched/missing/unexpected/mismatch counters, recall and airport accuracy, while normalizing and enforcing `^[A-Z0-9]{4}$` for actual arrival ICAO.
9. **Why selected.** It exposes prediction coverage without inventing time errors and strengthens truth identity.
10. **Rejected alternatives.** Artificial errors mix availability with timing accuracy; ignoring omissions preserves selection bias.
11. **Trade-offs.** Arrival evaluation has multiple complementary denominators instead of one simple MAE.
12. **Regression tests / protection.** Arrival recall, airport mismatch and strict ICAO validation regressions.
13. **Adversarial review findings.** Time MAE remains defined only where a matched airport/time comparison exists.
14. **Remediation iterations.** Closed in selective-prediction and truth-validation hardening.
15. **Residual risks / limitations.** ICAO-format validity does not prove the airport identifier is present in an external authoritative registry.
16. **Operational/deployment consequences.** Benchmark reports expose omitted/unexpected predictions explicitly.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects arrival counters, recall/mismatch semantics and identifier validation.

### GFA-DATA-407 — Fields named Median used non-standard statistical semantics

1. **Finding / symptom.** Even-count `Median...` values did not use the conventional mean of the two central sorted values.
2. **Root cause.** A different percentile-selection rule was reused under a median field name.
3. **Failure scenario.** For an even sample set, published median differs from the conventional statistical median expected by consumers.
4. **Impact.** Metrics are mislabeled and cross-tool comparisons are misleading.
5. **Severity rationale.** P2 retrospective because the arithmetic is deterministic but semantically wrong for the published name.
6. **Existing guarantees violated.** Metric-definition correctness.
7. **Considered solutions.** Rename field; document unusual semantics; implement conventional median and retain nearest-rank P95 separately.
8. **Chosen remediation.** Use middle value for odd counts and mean of two central values for even counts.
9. **Why selected.** It restores standard meaning without changing the public field name.
10. **Rejected alternatives.** Keeping unusual semantics under `Median` would preserve ambiguity.
11. **Trade-offs.** Historical even-count values may change.
12. **Regression tests / protection.** Conventional even-count median regression; P95 nearest-rank remains separately tested.
13. **Adversarial review findings.** The correction is limited to median semantics and does not impose one percentile convention on P95.
14. **Remediation iterations.** Closed in statistical semantics hardening.
15. **Residual risks / limitations.** Consumers comparing older outputs should account for the corrected version boundary.
16. **Operational/deployment consequences.** Some aggregate median values change by design.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects conventional median tests.

### GFA-DATA-408 — Aggregate input identity was contaminated by result metadata and did not explicitly preserve multiplicity

1. **Finding / symptom.** Aggregate fingerprinting could vary with `GeneratedAt` or fail to distinguish repeated identical evaluation inputs.
2. **Root cause.** Result metadata and semantic aggregate inputs were not cleanly separated in fingerprint construction.
3. **Failure scenario.** Re-running the same input set at a different generation time changes identity, or adding a duplicate evaluation does not alter the aggregate input identity correctly.
4. **Impact.** Aggregate reproducibility and idempotent comparison are weakened.
5. **Severity rationale.** P1 retrospective because aggregate identity no longer represents only its actual semantic inputs.
6. **Existing guarantees violated.** Deterministic input fingerprinting.
7. **Considered solutions.** Hash entire result including metadata; set-hash unique inputs; hash sorted input fingerprints with multiplicity.
8. **Chosen remediation.** Use `projection-replay-aggregate-fingerprint-v2` over sorted evaluation input fingerprints while preserving multiplicity; exclude `GeneratedAt`.
9. **Why selected.** It is order-independent, multiplicity-sensitive and metadata-independent.
10. **Rejected alternatives.** Set semantics lose duplicate inputs; generation time is output metadata, not input identity.
11. **Trade-offs.** Aggregate fingerprints intentionally change from the prior version.
12. **Regression tests / protection.** GeneratedAt-independence and multiplicity/order regressions.
13. **Adversarial review findings.** Generation time remains available as result metadata without contaminating identity.
14. **Remediation iterations.** Closed in aggregate-fingerprint Version Two.
15. **Residual risks / limitations.** Input evaluation fingerprints must themselves remain canonical.
16. **Operational/deployment consequences.** More stable rerun identity; no migration claimed by this review.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects aggregate fingerprint version and regression semantics.

### GFA-GOV-409 — Projection Evaluation documentation drifted from the implemented review contract

1. **Finding / symptom.** Stage/Documentation Index material did not fully describe the hardened replay-knowledge, metric, aggregation and review-closure contract.
2. **Root cause.** Engineering changes advanced faster than the authoritative narrative and navigation layer.
3. **Failure scenario.** A maintainer reads stage documentation and reconstructs obsolete evaluation semantics despite correct runtime behavior.
4. **Impact.** Governance evidence and future review decisions can diverge from implementation.
5. **Severity rationale.** P2 retrospective because this is a documentation/governance mismatch around critical analytical behavior.
6. **Existing guarantees violated.** Documentation truth and evidence-based closure.
7. **Considered solutions.** Leave code as authority only; add scattered comments; align Stage 9, this review record and Documentation Index.
8. **Chosen remediation.** Update authoritative documentation and protect the markers with the permanent Evaluation review audit.
9. **Why selected.** Repository policy requires code, tests, CI and documentation to agree before closure.
10. **Rejected alternatives.** Code-only truth makes retrospective review non-reproducible.
11. **Trade-offs.** Semantic changes require coordinated documentation updates.
12. **Regression tests / protection.** Strict source/documentation audit.
13. **Adversarial review findings.** This finding does not claim to close Projection Production, Projection Read or repository-wide reconciliation.
14. **Remediation iterations.** Closed with the Version Two documentation alignment.
15. **Residual risks / limitations.** External presentations outside the repository are not covered by this record.
16. **Operational/deployment consequences.** None at runtime.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`; Backend Quality in run `30619973772` executed the strict audit.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectionevaluationreviewaudit` protects stage/index/review markers.

### GFA-TEST-410 — Projection Evaluation regressions did not cover the accepted integrity boundaries

1. **Finding / symptom.** The review identified missing focused protection for projection-output identity, truth availability/order, physical interpolation, policy/metric reconstruction, aggregation separation, unavailable isolation and arrival accounting.
2. **Root cause.** Existing tests covered nominal evaluation behavior without one adversarial matrix matching the hardened contract.
3. **Failure scenario.** A coordinated mutation or future refactor reintroduces a benchmark-integrity defect while broad tests remain green.
4. **Impact.** Critical evaluation guarantees can regress without invalidating CI.
5. **Severity rationale.** P2 retrospective because this is a protection gap around multiple P1 analytical findings.
6. **Existing guarantees violated.** Regression durability and closure evidence.
7. **Considered solutions.** Manual review; rely on formular benchmark tests; add focused behavioral regressions plus a permanent strict audit.
8. **Chosen remediation.** Add the documented regression suite and wire `projectionevaluationreviewaudit -strict` into Backend Continuous Integration.
9. **Why selected.** It makes the exact accepted failure modes executable and permanently enforced.
10. **Rejected alternatives.** Manual review is non-reproducible; generic tests do not prove cross-field invariants.
11. **Trade-offs.** Contract evolution requires coordinated test/audit/documentation changes.
12. **Regression tests / protection.** The complete matrix listed in `Regression coverage` plus the strict audit.
13. **Adversarial review findings.** Mechanical `With` naming and optional-pointer recommendations remain outside the finding because they were explicitly rejected.
14. **Remediation iterations.** Completed with engineering closure and CI wiring.
15. **Residual risks / limitations.** Static audit checks complement, not replace, runtime behavior tests.
16. **Operational/deployment consequences.** CI is stricter; no runtime behavior added by this finding itself.
17. **Exact evidence.** `279d60543bbbb8c204fab60e442f00a56d1f3bbe`; run `30619973772`; jobs `91121986123`, `91121986134`, `91121986195`, `91122255083`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent Evaluation review audit remains required in Backend CI.
