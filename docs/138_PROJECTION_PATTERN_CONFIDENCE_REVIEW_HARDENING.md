# Projection Pattern Confidence Review Hardening

Status: closed

```text
FIRST_CONTRACT_HARDENING_COMMIT=6e6ac17cfcfca688d57829adfe2468346db6db1a
DISTRIBUTION_INTEGRITY_COMMIT=f73534feb275c5e109fa12fcfd9df5b69c56c03a
CONTINUATION_AGREEMENT_COMMIT=5873ae911b40197ee45eea30e7558aa04af78064
MANDATORY_CONTINUATION_INTERFACE_COMMIT=e31fcb5bbbb76093305e8b2c137c793a85dc6795
PERMANENT_AUDIT_COMMIT=cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30497703314
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90730254967
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90730255221
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90730255044
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90730452053
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_PATTERN_CONFIDENCE_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_PATTERN_CONFIDENCE_ENGINEERING_DEBT=CLOSED
PROJECTION_PATTERN_CONFIDENCE_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_PATTERN_CONFIDENCE_REVIEW_STATUS=CLOSED
```

## 1. Scope

This record covers the dedicated review and hardening of:

```text
apps/api/internal/projectionintelligence/projectionpatternconfidence
```

It also covers the mandatory continuation-aware dependency boundary through:

```text
projectionproduction
projectioncontinuation
```

The review is limited to deterministic historical-pattern confidence evidence,
policy, validation, fingerprinting, and production dependency contracts. It does
not claim empirical forecast accuracy, operational aviation suitability, or
calibrated probability semantics.

## 2. Findings closed

### 2.1 Semantic evidence fingerprint

The input fingerprint now binds the actual selected-neighbor evidence rather than
only aggregate or selection identity. It includes normalized trajectory identifiers,
similarity scores, similarity fingerprints, anchor distances, continuation sample
vectors, continuation-policy values, limitations, and all normalized component
weights.

Candidate age is deliberately absent because freshness belongs exclusively to
`projectionfreshness`.

### 2.2 Configuration contract hardening

Configuration validation now requires:

```text
minimum neighbor count of at least two
strictly positive minimum usable score
finite positive component weights
component-weight total equal to one
valid similarity floor
valid maximum similarity standard deviation
valid continuation sample count
ordered positive continuation divergence policy
```

Legacy aliases are normalized only when they do not conflict with canonical values.

### 2.3 Similarity distribution evidence

Pattern confidence publishes and validates:

```text
mean similarity score
minimum similarity score
similarity standard deviation
mean anchor distance
```

A weak individual neighbor or excessive score dispersion can block usability even
when the arithmetic mean remains high.

### 2.4 Freshness separation

Pattern confidence no longer scores candidate age. The former freshness component
was replaced by `similarity_consistency`. New evaluator results keep
`MeanCandidateAgeSeconds` at zero, and freshness decisions remain owned by
`projectionfreshness`.

### 2.5 Continuation agreement

The evaluator samples each selected historical continuation over the required
horizon, converts positions into anchor-relative displacement vectors, and computes
pairwise spread and divergence.

The contract detects opposing futures and intermediate route divergence even when
trajectories later reconverge. Missing continuation evidence or divergence above the
configured maximum makes the result unavailable.

### 2.6 Mandatory continuation-aware production interface

Both production consumers require:

```go
type PatternConfidenceEvaluator interface {
    EvaluateWithContinuations(
        selection projectionneighbors.Result,
        candidates []trajectory.FlightTrajectory,
    ) (projectionpatternconfidence.Result, error)
}
```

Runtime type assertions and fallback calls to `Evaluate(selection)` were removed.
A dependency that does not accept the actual historical candidates cannot satisfy
the production interface.

The concrete legacy `Evaluator.Evaluate` method remains only for source compatibility
and deliberately returns non-authorizing continuation-unknown evidence.

### 2.7 Result cross-field reconstruction

`Result.Validate()` reconstructs and verifies:

```text
policy validity
neighbor and continuation counts
aggregate similarity measurements
canonical five-component catalog
component scores and weights
weighted total score
usable decision
status
confidence level
required limitations
sorted unique trajectory identifiers
sorted unique limitations
continuation spread and divergence consistency
```

A manually assembled high-confidence result cannot bypass missing continuation
agreement, policy thresholds, component arithmetic, or decision semantics.

## 3. Deliberately retained and rejected recommendations

The following contracts were reviewed and deliberately retained:

```text
five fixed components as a versioned domain schema
float64 arithmetic with explicit finite checks and comparison tolerance
idiomatic New(Config) (*Evaluator, error) constructor
small local numeric helpers
legacy concrete Evaluate method as non-authorizing compatibility API
public deprecated aliases with conflict validation
```

The following recommendations were rejected as unsupported mechanical findings:

```text
mandatory integer basis points
mandatory dynamic component registry
constructor criticism based only on nil plus error
repository-wide prohibition of test names containing And
blanket rejection of local numeric helpers
```

The fixed component catalog is intentional because `Result.Validate()` must
reconstruct a stable, versioned confidence schema. A dynamic registry would weaken
that contract without a product requirement for runtime-extensible components.

## 4. Permanent regression coverage

Permanent tests cover:

```text
complete continuation-aware confidence
legacy evaluation cannot authorize
opposing continuation rejection
intermediate divergence rejection
semantic fingerprint mutation
candidate-order invariance
missing candidate rejection
interpolated continuation samples
similarity floor rejection
similarity dispersion rejection
freshness isolation
configuration migration and validation
policy and component mutation rejection
continuation pair-count validation
spread and divergence consistency
mandatory consumer interfaces
historical candidate propagation
```

## 5. Permanent audit gate

The source audit is:

```text
apps/api/tools/projectionpatternconfidencereviewaudit
```

Backend Continuous Integration executes:

```text
go run ./tools/projectionpatternconfidencereviewaudit -strict
```

The gate protects the hardened implementation, mandatory continuation-aware
interfaces, regression tests, workflow wiring, Stage 9 closure markers, this
authoritative record, and the Documentation Index entry.

## 6. Closure evidence

The permanent audit commit completed every Backend Continuous Integration job:

```text
PERMANENT_AUDIT_COMMIT=cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30497703314
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90730254967
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90730255221
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90730255044
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90730452053
```

The `Backend Quality` job executed and passed the dedicated step:

```text
Run projection pattern confidence review audit
```

Engineering implementation and formal closure documentation are complete. No
confirmed finding remains open, unclassified, or deferred. The permanent audit gate
remains mandatory in Backend Continuous Integration.
