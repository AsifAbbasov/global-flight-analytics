# Projection Continuation Review Hardening

Status: open

```text
REVIEW_BASELINE_COMMIT=a9b72001f1358af06a06f3a16212850daceef553
REVIEWED_MODULE=apps/api/internal/projectionintelligence/projectioncontinuation
CONFIRMED_BLOCKING_FINDINGS=7
PARTIALLY_ACCEPTED_RECOMMENDATIONS=5
REJECTED_NON_DEFECT_RECOMMENDATIONS=7
APPROVED_EVIDENCE_INTEGRITY=CI_CONFIRMED_COMMIT_23ecf72a0700b5a96459bc4a8618c72951a4e6aa_RUN_30573655172
INTERPOLATION_PLAUSIBILITY=IMPLEMENTED_PENDING_EXACT_CI
UNCERTAINTY_CONFIDENCE_CONSISTENCY=OPEN
GEODESIC_NUMERICAL_STABILITY=OPEN
PERMANENT_REVIEW_AUDIT=OPEN
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES
PROJECTION_CONTINUATION_REVIEW_STATUS=OPEN
```

## 1. Review assessment

The review correctly identifies a production evidence-identity defect: Production
Composition authorizes one Neighbor Selection and Pattern Confidence result, while
the historical projector previously recomputed both dependencies. That allowed the
published authorization chain and the geometric projection chain to diverge when an
implementation was stateful or its input changed.

The review is also correct that the continuation boundary did not bind selected
anchor metadata to the actual candidate trajectory consumed by interpolation, did
not require Pattern Confidence `SourceSelectionFingerprint` equality, and discarded
the observed source identity of historical trajectory evidence.

## 2. Approved Evidence Integrity increment

Production Composition now invokes a compile-time distinct `ProjectApproved`
contract. It passes the exact authorized Neighbor Selection and Pattern Confidence
results to Projection Continuation.

Projection Continuation clones and validates the approved evidence, verifies
request ownership, requires exact Pattern-to-Selection fingerprint lineage, binds
candidate start/end and anchor/continuation timestamps to the consumed trajectory,
and publishes historical candidate inputs as observed evidence with their real
source.

The standalone `Project` entrypoint remains available for direct module use and
performs its own selection and pattern evaluation. Production cannot accidentally
call that path because its interface requires `ProjectApproved`.

## 3. Interpolation Plausibility increment

Projection Continuation now uses a constructor-normalized `PlausibilityPolicy`
with explicit maximum interpolation gap, maximum implied horizontal speed, and
maximum implied vertical speed.

The policy is applied to both interpolated targets and exact observed endpoints.
A giant jump cannot bypass validation merely because the forecast time exactly
matches the right-hand point.

Rejected samples are distinguished from ordinary missing samples. If filtering
removes required point support, the method falls back with
`historical_continuation_plausibility_support_insufficient`. If sufficient support
remains, the output is marked limited and publishes
`historical_continuation_plausibility_filtered`.

Effective policy values are included in the continuation fingerprint.

## 4. Confirmed work still open

```text
confidence penalty from neighbor disagreement
uncertainty composition semantics
near-antipodal spherical-mean epsilon
explicit status semantics when confidence reaches zero
```

These remaining mathematical-policy items require separate regression and
calibration evidence.

## 5. Recommendations not accepted as defects

### 5.1 Continuation fingerprint collision

Rejected. Continuation fingerprints the complete Horizon Plan fingerprint. The
Horizon Plan fingerprint already includes version, policy name, requested and
effective duration, truncation state and reason, and every forecast timestamp.

### 5.2 Mandatory basis-point score conversion

Rejected. These are bounded non-monetary analytical scores. `float64`, finite-value
validation and explicit comparison tolerance are appropriate.

### 5.3 Constructor returning nil and error

Rejected. `New(Config) (*Baseline, error)` is idiomatic Go.

### 5.4 Optional altitude through pointer presence

Rejected. `*float64` represents explicit value absence in the current Go contract.

### 5.5 Test-name wording rules

Rejected. Descriptive Go test names containing words such as `And` are not a
correctness or maintainability violation.

### 5.6 Function length alone

Partially rejected. Decomposition is justified by responsibility and testability,
not line count alone. The package is already split into focused source units.

### 5.7 Quadrature as the only correct uncertainty composition

Not accepted without qualification. Root-sum-square requires documented
independence assumptions for its components.
