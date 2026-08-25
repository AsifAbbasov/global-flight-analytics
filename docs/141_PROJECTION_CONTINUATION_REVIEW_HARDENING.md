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

## Canonical remediation history

The following seven records normalize the `CONFIRMED_BLOCKING_FINDINGS=7` review scope into the repository-wide nineteen-field standard. Severity is retrospective. Partially accepted hardening items such as deterministic equal-timestamp ordering and irregular-grid rejection remain supporting protection unless they own an independent documented defect.

### GFA-DATA-376 — Production authorization evidence could diverge from Continuation execution evidence

1. **Finding / symptom.** Production Composition authorized one Neighbor Selection and Pattern Confidence result while Projection Continuation could recompute those collaborators.
2. **Root cause.** Authorization and geometric execution did not consume one immutable approved evidence bundle.
3. **Failure scenario.** A stateful collaborator or changed input produces a different selection during projection than the result previously authorized by production composition.
4. **Impact.** Published continuation geometry can be supported by a different evidence chain from the one that authorized historical continuation.
5. **Severity rationale.** P1 retrospective because this is an evidence-identity violation on a production analytical path.
6. **Existing guarantees violated.** Deterministic lineage, authorization integrity, reproducible provenance.
7. **Considered solutions.** Recompute and compare; trust recomputation; pass the exact approved collaborator results into the projector.
8. **Chosen remediation.** Introduce the distinct `ProjectApproved` contract and pass the exact authorized Neighbor Selection and Pattern Confidence results.
9. **Why selected.** It removes the second source of truth instead of trying to prove two executions happened to agree.
10. **Rejected alternatives.** Recompute-and-compare retained duplicate execution ownership; trust-only behavior left stateful divergence possible.
11. **Trade-offs.** Production composition and projector contracts are more explicit and tightly versioned.
12. **Regression tests / protection.** Approved-evidence integration tests and the permanent Projection Continuation review audit.
13. **Adversarial review findings.** The defect is about identity of approved evidence, not merely deterministic sorting.
14. **Remediation iterations.** Closed in the approved-evidence integrity wave and retained through engineering closure.
15. **Residual risks / limitations.** Correctness still depends on upstream approved results themselves satisfying their own validators.
16. **Operational/deployment consequences.** No migration; production composition now supplies exact approved evidence rather than requesting recomputation.
17. **Exact evidence.** `23ecf72a0700b5a96459bc4a8618c72951a4e6aa`, run `30573655172`; closure `13838c4273a3be6bde63835e1d8f51af6f6daa21`, run `30593549087`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectioncontinuationreviewaudit` protects the approved-evidence contract and lineage markers.

### GFA-DATA-377 — Selected neighbor metadata was not fully bound to the candidate trajectory consumed by interpolation

1. **Finding / symptom.** Selected anchor metadata, Pattern-to-Selection lineage, and historical source identity could disagree with the trajectory actually used by interpolation.
2. **Root cause.** Structural result validation did not prove candidate start/end/anchor/continuation timestamps and source identity against the consumed trajectory.
3. **Failure scenario.** A valid-looking selector result points at one anchor while interpolation consumes a different candidate trajectory or source.
4. **Impact.** Continuation output can be geometrically derived from evidence different from its published neighbor provenance.
5. **Severity rationale.** P1 retrospective because candidate evidence determines the projected path.
6. **Existing guarantees violated.** Evidence binding, provenance truth, selection-to-projection lineage.
7. **Considered solutions.** Trust selector metadata; compare only trajectory ID; validate the complete candidate evidence tuple and Pattern lineage.
8. **Chosen remediation.** Clone and validate approved evidence, require exact Pattern-to-Selection fingerprint lineage, bind candidate timestamps to the consumed trajectory, and publish actual observed source names.
9. **Why selected.** It validates the semantic object used by interpolation rather than a partial identifier mirror.
10. **Rejected alternatives.** ID-only checks missed timestamp/source drift; trusting selector output left the projector boundary porous.
11. **Trade-offs.** Standalone and approved projector paths perform stricter candidate validation.
12. **Regression tests / protection.** Candidate-binding, lineage-mismatch, source-identity and standalone-project regressions.
13. **Adversarial review findings.** Standalone `Project` was included so the production-only fix could not leave an alternate bypass.
14. **Remediation iterations.** Approved-path binding was followed by standalone candidate-evidence protection.
15. **Residual risks / limitations.** Upstream trajectory source metadata must itself be truthful.
16. **Operational/deployment consequences.** Inconsistent candidate evidence now fails closed or falls back instead of being projected.
17. **Exact evidence.** `23ecf72a0700b5a96459bc4a8618c72951a4e6aa`; engineering closure `13838c4273a3be6bde63835e1d8f51af6f6daa21`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit requires approved and standalone candidate evidence binding and exact lineage.

### GFA-DATA-378 — Historical interpolation lacked explicit physical plausibility bounds

1. **Finding / symptom.** Interpolation could use long temporal gaps or imply physically implausible horizontal or vertical motion.
2. **Root cause.** Sample availability was checked without one versioned plausibility policy for gap and implied rates.
3. **Failure scenario.** Sparse or corrupt historical points are interpolated into a plausible-looking future continuation.
4. **Impact.** Numerically valid but physically misleading projection coordinates can be published.
5. **Severity rationale.** P1 retrospective because the defect directly changes continuation geometry.
6. **Existing guarantees violated.** Conservative modeling, evidence integrity, fail-closed physical bounds.
7. **Considered solutions.** Trust upstream trajectories; clamp implied motion; reject samples exceeding explicit gap/speed policies.
8. **Chosen remediation.** Add constructor-normalized maximum interpolation gap, maximum implied horizontal speed, and maximum implied vertical speed.
9. **Why selected.** Rejection preserves source truth and distinguishes implausible evidence from ordinary missing evidence.
10. **Rejected alternatives.** Clamping manufactures motion; upstream-only validation does not protect historical/custom paths.
11. **Trade-offs.** Some sparse candidates lose support or trigger fallback.
12. **Regression tests / protection.** Gap, horizontal-speed, vertical-speed, endpoint and fallback regressions.
13. **Adversarial review findings.** Exact observed endpoints are subject to the same plausibility policy rather than receiving a hidden exemption.
14. **Remediation iterations.** Implemented in the dedicated interpolation-plausibility wave.
15. **Residual risks / limitations.** Static research bounds are conservative guards, not operational aircraft envelopes.
16. **Operational/deployment consequences.** More suspect historical samples are rejected; no persistence migration.
17. **Exact evidence.** `739073de31e4c1da2aa105d495bc789a294cb3c9`, run `30576928637`; closure run `30593549087`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects plausibility policy identity and regression coverage.

### GFA-DATA-379 — Neighbor disagreement did not coherently increase uncertainty and reduce confidence

1. **Finding / symptom.** Historical-neighbor spread could be weakly represented in uncertainty while point confidence remained insufficiently coupled to the same disagreement evidence.
2. **Root cause.** Model uncertainty, observed disagreement, support, and confidence were composed by separate optimistic rules.
3. **Failure scenario.** Widely disagreeing historical continuations produce a high-confidence point with understated uncertainty.
4. **Impact.** Published uncertainty and confidence tell inconsistent stories about the same evidence.
5. **Severity rationale.** P1 retrospective because both values are public decision evidence consumed downstream.
6. **Existing guarantees violated.** Conservative uncertainty, confidence explainability, cross-field consistency.
7. **Considered solutions.** RSS uncertainty; max-only uncertainty; additive conservative uncertainty with explicit agreement factor.
8. **Chosen remediation.** Add configured and disagreement uncertainty, require disagreement multiplier at least one, and include neighbor agreement directly in point confidence.
9. **Why selected.** Additive composition does not assume statistical independence and makes greater spread monotonically reduce confidence.
10. **Rejected alternatives.** RSS lacked an independence model; disagreement-independent confidence remained contradictory.
11. **Trade-offs.** Confidence becomes more conservative when historical candidates disagree.
12. **Regression tests / protection.** Disagreement, uncertainty, confidence-factor, zero-confidence and configuration regressions.
13. **Adversarial review findings.** Quadrature was explicitly rejected as the only valid formula because independence was not established.
14. **Remediation iterations.** Finalized in engineering closure after plausibility and evidence-identity hardening.
15. **Residual risks / limitations.** The uncertainty model remains a research heuristic, not calibrated probabilistic uncertainty.
16. **Operational/deployment consequences.** High-disagreement outputs may become limited or lower-confidence.
17. **Exact evidence.** Engineering closure `13838c4273a3be6bde63835e1d8f51af6f6daa21`, run `30593549087`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Strict audit protects additive uncertainty, disagreement penalty and confidence-reason semantics.

### GFA-DATA-380 — Raw sample count overstated support when neighbor weights were concentrated

1. **Finding / symptom.** Support used raw contributing-neighbor count even when one or few similarity weights dominated the combination.
2. **Root cause.** Cardinality was treated as equivalent to independent effective evidence.
3. **Failure scenario.** Ten nominal samples with almost all weight on one neighbor receive near-full support.
4. **Impact.** Confidence can be overstated despite low effective diversity of contributing evidence.
5. **Severity rationale.** P1 retrospective because support directly contributes to published continuation confidence.
6. **Existing guarantees violated.** Evidence-weighted confidence and conservative support semantics.
7. **Considered solutions.** Raw count; entropy-based support; Kish effective sample size.
8. **Chosen remediation.** Use Kish effective sample size and divide by authorized neighbor count for the support ratio.
9. **Why selected.** It is deterministic, bounded, weight-aware, and preserves full support for balanced weights.
10. **Rejected alternatives.** Raw count ignored concentration; more complex diversity metrics were unnecessary for the documented contract.
11. **Trade-offs.** Highly concentrated selections receive lower support even with many nominal samples.
12. **Regression tests / protection.** Balanced and dominated-weight effective-support tests.
13. **Adversarial review findings.** The denominator remains authorized-neighbor count, keeping the score tied to the approved selection.
14. **Remediation iterations.** Closed in the engineering-closure wave.
15. **Residual risks / limitations.** Effective sample size measures weight concentration, not semantic correlation among trajectories.
16. **Operational/deployment consequences.** Some historical continuations publish lower confidence.
17. **Exact evidence.** `13838c4273a3be6bde63835e1d8f51af6f6daa21`, run `30593549087`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects the Kish support formula and associated regressions.

### GFA-DATA-381 — Near-antipodal spherical means could resolve to an arbitrary direction from floating-point residue

1. **Finding / symptom.** Equal or nearly antipodal historical samples could produce a tiny mean vector that normalized into an arbitrary geographic direction.
2. **Root cause.** The spherical mean normalized any non-zero floating-point vector without a stability threshold.
3. **Failure scenario.** Opposing candidate positions numerically cancel and residual rounding determines the published point.
4. **Impact.** Projection coordinates become non-physical and non-robust around a known geometric degeneracy.
5. **Severity rationale.** P2 retrospective because the case is narrower than general interpolation defects but directly affects output geometry.
6. **Existing guarantees violated.** Numerical stability and deterministic geodesic semantics.
7. **Considered solutions.** Normalize every non-zero vector; choose one sample deterministically; reject unstable means and use fallback.
8. **Chosen remediation.** Require vector norm above `weightedMeanVectorNormEpsilon`; otherwise reject the combination and invoke deterministic fallback.
9. **Why selected.** It avoids inventing a direction when the evidence has no stable spherical mean.
10. **Rejected alternatives.** Arbitrary sample choice hides disagreement; zero-only checks remain vulnerable to near-zero residue.
11. **Trade-offs.** Degenerate combinations cannot produce a weighted spherical point.
12. **Regression tests / protection.** Antipodal and near-antipodal regressions plus fallback verification.
13. **Adversarial review findings.** The fallback path is preferred over pretending numerical residue is evidence.
14. **Remediation iterations.** Finalized in engineering closure.
15. **Residual risks / limitations.** The epsilon is a numerical policy and must remain versioned with method semantics.
16. **Operational/deployment consequences.** Degenerate samples fall back rather than emitting unstable coordinates.
17. **Exact evidence.** `13838c4273a3be6bde63835e1d8f51af6f6daa21`, run `30593549087`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Strict audit protects the norm threshold and regression tests.

### GFA-OPS-382 — Continuation fallback discarded the underlying projector error cause

1. **Finding / symptom.** Fallback failures could preserve only the strategy-level error and lose the underlying projector cause.
2. **Root cause.** Error wrapping did not retain both layers of failure through Go error-chain semantics.
3. **Failure scenario.** Operators see `ErrFallbackProjectionFailed` but cannot determine or match the original cause with `errors.Is`.
4. **Impact.** Incident diagnosis and typed recovery behavior are weakened on already-degraded projection paths.
5. **Severity rationale.** P2 retrospective because execution fails closed but causality and diagnosability are materially degraded.
6. **Existing guarantees violated.** Typed error semantics, failure provenance, operability.
7. **Considered solutions.** Flatten error text; retain only strategy error; wrap both strategy and underlying projector causes.
8. **Chosen remediation.** Preserve both layers with `%w`-compatible wrapping so callers can match `ErrFallbackProjectionFailed` and the original cause.
9. **Why selected.** It keeps the public fallback classification without destroying causal evidence.
10. **Rejected alternatives.** String concatenation is not machine-inspectable; single-layer wrapping loses information.
11. **Trade-offs.** Error chains are slightly richer and callers should use `errors.Is` rather than text matching.
12. **Regression tests / protection.** Fallback cause-preservation regression and permanent audit.
13. **Adversarial review findings.** This is distinct from ordinary domain fallback: the issue was loss of the failure cause after fallback itself failed.
14. **Remediation iterations.** Closed in engineering closure alongside standalone candidate binding and deterministic ordering guards.
15. **Residual risks / limitations.** External logging layers must still avoid flattening the error chain into misleading text.
16. **Operational/deployment consequences.** Better diagnosis only; no availability expansion or migration.
17. **Exact evidence.** `13838c4273a3be6bde63835e1d8f51af6f6daa21`, run `30593549087`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit requires fallback cause preservation and its focused regression.
