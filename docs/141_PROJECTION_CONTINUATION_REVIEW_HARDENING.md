# Projection Continuation Review Hardening

Status: closed

```text
REVIEW_BASELINE_COMMIT=a9b72001f1358af06a06f3a16212850daceef553
REVIEWED_MODULE=apps/api/internal/projectionintelligence/projectioncontinuation
CONFIRMED_BLOCKING_FINDINGS=7
PARTIALLY_ACCEPTED_RECOMMENDATIONS=5
REJECTED_NON_DEFECT_RECOMMENDATIONS=7
APPROVED_EVIDENCE_INTEGRITY=CI_CONFIRMED_COMMIT_23ecf72a0700b5a96459bc4a8618c72951a4e6aa_RUN_30573655172
INTERPOLATION_PLAUSIBILITY=CI_CONFIRMED_COMMIT_739073de31e4c1da2aa105d495bc789a294cb3c9_RUN_30576928637
ENGINEERING_CLOSURE_COMMIT=13838c4273a3be6bde63835e1d8f51af6f6daa21
ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30593549087
ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91040848886
ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91040848927
ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91040848967
ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91041042383
UNCERTAINTY_CONFIDENCE_CONSISTENCY=CI_CONFIRMED
GEODESIC_NUMERICAL_STABILITY=CI_CONFIRMED
EFFECTIVE_WEIGHTED_SUPPORT=CI_CONFIRMED
FALLBACK_ERROR_CAUSE_PRESERVATION=CI_CONFIRMED
STANDALONE_CANDIDATE_EVIDENCE_BINDING=CI_CONFIRMED
DETERMINISTIC_EQUAL_TIMESTAMP_ORDERING=CI_CONFIRMED
IRREGULAR_FORECAST_GRID_REJECTION=CI_CONFIRMED
PERMANENT_REVIEW_AUDIT=CI_CONFIRMED
ENGINEERING_IMPLEMENTATION=COMPLETE
ENGINEERING_DEBT=CLOSED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_CONTINUATION_FORMAL_CLOSURE=COMPLETE
PROJECTION_CONTINUATION_REVIEW_STATUS=CLOSED
```

## 1. Review assessment

The review correctly identified a production evidence-identity defect: Production
Composition authorized one Neighbor Selection and Pattern Confidence result while the
historical projector recomputed both dependencies. The published authorization chain
and geometric projection chain could therefore diverge for stateful implementations
or changed input.

The review also correctly identified missing binding between selected anchor metadata
and the candidate trajectory consumed by interpolation, incomplete Pattern-to-Selection
lineage validation, loss of observed source identity, unsafe interpolation, weak
uncertainty composition, disagreement-independent confidence, unstable near-antipodal
spherical means, unweighted support, and lost fallback error causes.

## 2. Approved Evidence Integrity — Continuous Integration confirmed

Production Composition invokes a compile-time distinct `ProjectApproved` contract and
passes the exact authorized Neighbor Selection and Pattern Confidence results.
Projection Continuation clones and validates that evidence, verifies request ownership,
requires exact Pattern-to-Selection fingerprint lineage, binds candidate start, end,
anchor and continuation timestamps to the consumed trajectory, and publishes
historical candidates as observed inputs with their actual source names.

```text
COMMIT=23ecf72a0700b5a96459bc4a8618c72951a4e6aa
GITHUB_ACTIONS_RUN=30573655172
```

## 3. Interpolation Plausibility — Continuous Integration confirmed

A constructor-normalized `PlausibilityPolicy` enforces maximum interpolation gap,
maximum implied horizontal speed and maximum implied vertical speed. The policy applies
to interpolated targets and exact observed endpoints. Rejected samples are
distinguished from ordinary missing samples, causing either a limited result or
fallback according to remaining support. Effective policy values are fingerprinted.

```text
COMMIT=739073de31e4c1da2aa105d495bc789a294cb3c9
GITHUB_ACTIONS_RUN=30576928637
```

## 4. Uncertainty and confidence consistency

Configured model uncertainty and weighted neighbor-disagreement uncertainty are
combined conservatively by addition:

```text
total uncertainty = configured uncertainty + disagreement uncertainty
```

This preserves both components and deliberately avoids root-sum-square because the
independence required by that formula is not established. The disagreement multiplier
must be finite and at least one, so configuration cannot suppress observed spread.

Point confidence composes four bounded factors:

```text
pattern confidence
× effective weighted support
× neighbor agreement
× horizon retention
```

The agreement factor is the configured component divided by total uncertainty. Greater
horizontal or vertical disagreement therefore increases uncertainty and decreases
confidence in the same result. Confidence reasons publish every factor separately.

`MaximumConfidenceLoss` must be less than one. The production policy therefore cannot
force the terminal point to zero solely through horizon decay. Independently, the
result builder treats any zero-confidence point as concrete limited-status evidence
with `historical_continuation_confidence_none`.

## 5. Effective weighted support

Raw sample count is replaced by Kish effective sample size:

```text
effective sample size = (sum of weights squared) / sum of squared weights
effective support ratio = effective sample size / authorized neighbor count
```

Balanced weights preserve full support. A collection dominated by one similarity
weight receives less support even when the raw sample count is unchanged.

## 6. Geodesic numerical stability

The normalized spherical mean vector must have a norm greater than
`weightedMeanVectorNormEpsilon`. Equal or nearly antipodal samples no longer produce an
arbitrary direction from floating-point residue; the combination is rejected and the
existing deterministic fallback path is used.

## 7. Candidate evidence and fallback causality

Candidate evidence binding applies to both `ProjectApproved` and standalone `Project`.
A structurally valid but trajectory-inconsistent selector result cannot silently direct
interpolation to a different anchor.

Fallback wrapping uses `%w` for both the strategy error and underlying projector error.
Callers can therefore use `errors.Is` for `ErrFallbackProjectionFailed` and the original
cause.

Trajectory snapshots use a canonical equal-timestamp tie-break based on point ID and
bit-stable numeric fields instead of preserving input order. Equivalent point sets
therefore produce the same interpolation order. An irregular forecast grid is rejected
by complete Horizon Plan validation before projection.

## 8. Fingerprint collision assessment

The alleged truncated-plan collision is not present in the hardened implementation.
Projection Continuation fingerprints the complete Horizon Plan fingerprint, which
includes version, policy name, requested and effective duration, truncation state and
reason, and every forecast timestamp. A dedicated regression test proves that exact and
truncated requests with the same effective duration produce different continuation
fingerprints.

Method and continuation/fallback fingerprint versions advanced to version 3 because
the uncertainty, support and confidence semantics changed.

## 9. Recommendations not accepted as defects

Mandatory basis-point conversion was rejected because these are bounded non-monetary
analytical scores with finite-value validation and explicit comparison tolerance.
`New(Config) (*Baseline, error)` remains idiomatic Go. Optional altitude remains an
explicit pointer-presence contract. Test names containing conjunctions are not defects.
Function length alone is not a correctness finding, and quadrature is not treated as
the only valid uncertainty formula without an independence model.

## 10. Permanent review enforcement

Permanent regression enforcement is implemented in:

```text
apps/api/tools/projectioncontinuationreviewaudit
```

Backend Continuous Integration executes the audit in strict mode. It protects version
identity, configuration boundaries, additive uncertainty composition, effective
weighted support, disagreement confidence penalty, zero-confidence status semantics,
near-antipodal rejection, standalone candidate binding, fallback cause preservation,
regression tests, exact engineering-closure evidence and formal review status.

## 11. Exact engineering-closure Continuous Integration

The engineering closure commit passed the exact Backend Continuous Integration run and
all four mandatory jobs:

```text
COMMIT=13838c4273a3be6bde63835e1d8f51af6f6daa21
GITHUB_ACTIONS_RUN=30593549087
POSTGRESQL_16_INTEGRATION_JOB=91040848886
BACKEND_RACE_SAFETY_JOB=91040848927
BACKEND_QUALITY_JOB=91040848967
BACKEND_CONTAINER_JOB=91041042383
```

## 12. Formal closure

All confirmed findings are implemented and protected by regression tests and the
permanent strict audit. Exact Continuous Integration evidence for the engineering
closure is recorded above. There are no open, unclassified or deferred findings in the
review scope.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_CONTINUATION_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_CONTINUATION_ENGINEERING_DEBT=CLOSED
PROJECTION_CONTINUATION_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_CONTINUATION_FORMAL_CLOSURE=COMPLETE
PROJECTION_CONTINUATION_REVIEW_STATUS=CLOSED
```

The formal-closure commit must itself pass the same four Backend Continuous Integration
jobs before an external final report is issued. That final run is a release gate for the
closure record, not an additional engineering finding.
