# Projection Continuation Review Hardening

Status: open

```text
REVIEW_BASELINE_COMMIT=a9b72001f1358af06a06f3a16212850daceef553
REVIEWED_MODULE=apps/api/internal/projectionintelligence/projectioncontinuation
CONFIRMED_BLOCKING_FINDINGS=7
PARTIALLY_ACCEPTED_RECOMMENDATIONS=5
REJECTED_NON_DEFECT_RECOMMENDATIONS=7
APPROVED_EVIDENCE_INTEGRITY=IMPLEMENTED_PENDING_EXACT_CI
INTERPOLATION_PLAUSIBILITY=OPEN
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

## 3. Confirmed work still open

```text
bounded interpolation gaps
maximum implied horizontal speed
maximum implied vertical rate
confidence penalty from neighbor disagreement
uncertainty composition semantics
near-antipodal spherical-mean epsilon
explicit status semantics when confidence reaches zero
```

These mathematical-policy items require separate regression and calibration
evidence and are not mixed into the evidence-identity patch.

## 4. Recommendations not accepted as defects

### 4.1 Continuation fingerprint collision

Rejected. Continuation fingerprints the complete Horizon Plan fingerprint. The
Horizon Plan fingerprint already includes version, policy name, requested and
effective duration, truncation state and reason, and every forecast timestamp.

### 4.2 Mandatory basis-point score conversion

Rejected. These are bounded non-monetary analytical scores. `float64`, finite-value
validation and explicit comparison tolerance are appropriate.

### 4.3 Constructor returning nil and error

Rejected. `New(Config) (*Baseline, error)` is idiomatic Go.

### 4.4 Optional altitude through pointer presence

Rejected. `*float64` represents explicit value absence in the current Go contract.

### 4.5 Test-name wording rules

Rejected. Descriptive Go test names containing words such as `And` are not a
correctness or maintainability violation.

### 4.6 Function length alone

Partially rejected. Decomposition is justified by responsibility and testability,
not line count alone. The package is already split into focused source units.

### 4.7 Quadrature as the only correct uncertainty composition

Not accepted without qualification. Root-sum-square requires documented
independence assumptions for its components.
