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

## Canonical remediation history

The following records retrospectively normalize the accepted review findings into the
repository-wide nineteen-field standard. Severity values are retrospective engineering
classifications. Historical implementation ownership follows the named remediation
waves above; exact historical CI is claimed only for the permanent-audit run explicitly
recorded by this source.

### GFA-DATA-351 — Pattern Confidence fingerprint omitted decision-relevant selected-neighbor evidence

1. **Finding / symptom.** Pattern Confidence identity could rely on aggregate or selection identity without binding all selected-neighbor evidence used by the score.
2. **Root cause.** Fingerprinting was narrower than the semantic inputs consumed by confidence evaluation.
3. **Failure scenario.** Similarity fingerprints, anchor distances, continuation vectors, limitations, or policy weights change while the Pattern Confidence input identity remains unchanged.
4. **Impact.** Distinct confidence computations can collide in provenance, caches, or downstream lineage checks.
5. **Severity rationale.** P1 retrospective because this fingerprint authorizes downstream projection evidence.
6. **Existing guarantees violated.** Semantic identity, reproducible provenance, and change-sensitive lineage.
7. **Considered solutions.** Hash only upstream selection fingerprint; hash result outputs; hash every decision-relevant normalized input and policy value.
8. **Chosen remediation.** Bind selected trajectory IDs, similarity scores/fingerprints, anchor evidence, continuation samples/policy, limitations, and normalized weights into the input fingerprint.
9. **Why selected.** Identity now changes exactly when evidence capable of changing the decision changes.
10. **Rejected alternatives.** Upstream aggregate identity was too coarse; output hashing would conflate inputs with derived results.
11. **Trade-offs.** Fingerprint construction is larger and must evolve when semantic inputs evolve.
12. **Regression tests / protection.** Semantic fingerprint mutation and candidate-order invariance tests are permanent.
13. **Adversarial review findings.** Candidate age is intentionally excluded because freshness is owned by `projectionfreshness`.
14. **Remediation iterations.** Fingerprint scope expanded through contract, distribution, and continuation-agreement hardening.
15. **Residual risks / limitations.** New evidence fields must be added deliberately to the fingerprint when they become decision-relevant.
16. **Operational/deployment consequences.** Fingerprint versions changed; no database migration.
17. **Exact evidence.** `6e6ac17cfcfca688d57829adfe2468346db6db1a`, `5873ae911b40197ee45eea30e7558aa04af78064`; permanent audit `cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42`, run `30497703314`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectionpatternconfidencereviewaudit` and fingerprint mutation tests protect semantic identity.

### GFA-CONTRACT-352 — Pattern Confidence configuration admitted zero-information or internally incoherent policy states

1. **Finding / symptom.** Configuration validation did not fully reject weak or contradictory neighbor, score, weight, distribution, and continuation-policy values.
2. **Root cause.** Individual fields were accepted without complete cross-field and strict-positive domain validation.
3. **Failure scenario.** A zero usable-score threshold, zero component weight, one-neighbor pattern, invalid divergence ordering, or conflicting legacy alias reaches evaluation.
4. **Impact.** High-confidence or usable states can become reachable under policy that carries insufficient information or ambiguous configuration intent.
5. **Severity rationale.** P1 retrospective because invalid policy can directly authorize a projection confidence result.
6. **Existing guarantees violated.** Constructor-time policy validity and deterministic compatibility normalization.
7. **Considered solutions.** Clamp/normalize every value; allow zero as disabled components; reject invalid or conflicting policy while preserving only unambiguous aliases.
8. **Chosen remediation.** Require at least two neighbors, positive thresholds and weights summing to one, valid distribution and continuation policies, and reject conflicting legacy/canonical aliases.
9. **Why selected.** Fail-fast validation prevents ambiguous policy from entering evidence computation.
10. **Rejected alternatives.** Silent clamping hides caller mistakes; zero weights undermine the fixed component contract.
11. **Trade-offs.** Some formerly accepted configurations now fail construction and legacy aliases require unambiguous migration.
12. **Regression tests / protection.** Config migration, zero/invalid thresholds, weights, sample counts, and divergence ordering are tested.
13. **Adversarial review findings.** Fixed components and `float64` remain intentional; the defect was invalid policy acceptance, not their existence.
14. **Remediation iterations.** Contract validation was tightened first, then expanded for similarity-distribution and continuation-agreement policy.
15. **Residual risks / limitations.** Threshold values are engineering policy and do not imply empirical calibration.
16. **Operational/deployment consequences.** No migration; invalid startup/configuration fails earlier and more explicitly.
17. **Exact evidence.** `6e6ac17cfcfca688d57829adfe2468346db6db1a`, `f73534feb275c5e109fa12fcfd9df5b69c56c03a`, `5873ae911b40197ee45eea30e7558aa04af78064`; permanent audit run `30497703314`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Constructor validation and strict review audit protect policy coherence.

### GFA-DATA-353 — Mean similarity alone hid weak neighbors and unstable similarity distributions

1. **Finding / symptom.** Pattern support could appear strong from a high arithmetic mean even when one selected neighbor was weak or similarity scores were highly dispersed.
2. **Root cause.** Confidence evidence summarized similarity primarily through its mean.
3. **Failure scenario.** A set with one very poor neighbor and several strong neighbors passes because the mean remains above threshold.
4. **Impact.** Pattern confidence overstates consistency of the selected historical evidence.
5. **Severity rationale.** P1 retrospective because weak or unstable evidence can authorize downstream historical projection.
6. **Existing guarantees violated.** Evidence-quality transparency and conservative confidence semantics.
7. **Considered solutions.** Keep mean only; use minimum only; publish mean, minimum, and standard deviation with explicit policy thresholds.
8. **Chosen remediation.** Add minimum similarity and similarity standard deviation, validate them, and include a similarity-consistency component.
9. **Why selected.** It exposes both weakest-link and distribution-dispersion information without discarding the useful mean.
10. **Rejected alternatives.** Mean-only hides tails; minimum-only discards distribution shape.
11. **Trade-offs.** Result schema and policy are richer and require compatibility handling.
12. **Regression tests / protection.** Similarity-floor and dispersion-rejection tests are permanent.
13. **Adversarial review findings.** A fixed versioned component schema was retained so validation can reconstruct the result exactly.
14. **Remediation iterations.** Distribution evidence replaced the former freshness component during the second hardening wave.
15. **Residual risks / limitations.** Standard deviation is descriptive engineering evidence, not a calibrated probability model.
16. **Operational/deployment consequences.** No migration; some formerly usable patterns become unavailable or lower confidence.
17. **Exact evidence.** `f73534feb275c5e109fa12fcfd9df5b69c56c03a`; permanent audit `cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42`, run `30497703314`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit requires minimum/dispersion evidence and corresponding regressions.

### GFA-CONTRACT-354 — Pattern Confidence duplicated freshness ownership by scoring candidate age

1. **Finding / symptom.** Pattern Confidence included candidate age even though Projection Freshness is the dedicated owner of age-based evidence and policy.
2. **Root cause.** Early confidence composition mixed similarity-pattern quality with recency quality.
3. **Failure scenario.** The same historical age evidence is penalized once inside Pattern Confidence and again inside `projectionfreshness`.
4. **Impact.** Confidence semantics become double-counted, harder to explain, and policy ownership becomes ambiguous.
5. **Severity rationale.** P1 retrospective because duplicated decision evidence can materially alter whether historical projection is usable.
6. **Existing guarantees violated.** Single-owner domain semantics and reconstructible confidence composition.
7. **Considered solutions.** Keep freshness in both modules; remove the dedicated freshness module; isolate age entirely to `projectionfreshness`.
8. **Chosen remediation.** Replace the Pattern Confidence freshness component with `similarity_consistency`; keep new evaluator `MeanCandidateAgeSeconds` at zero.
9. **Why selected.** Each module now has one clear responsibility and downstream decisions can identify which policy applied the age penalty.
10. **Rejected alternatives.** Double scoring is opaque; collapsing Freshness into Pattern Confidence destroys its independent lineage and policy contract.
11. **Trade-offs.** Historical compatibility fields remain but no longer authorize new age-based Pattern Confidence behavior.
12. **Regression tests / protection.** Freshness-isolation and compatibility tests protect the ownership boundary.
13. **Adversarial review findings.** Candidate age is deliberately excluded from the Pattern Confidence fingerprint for the same ownership reason.
14. **Remediation iterations.** Freshness was removed during distribution-evidence hardening and downstream fixtures were updated.
15. **Residual risks / limitations.** Legacy fields remain source-compatible and therefore must continue to be validated as non-authorizing.
16. **Operational/deployment consequences.** No migration; confidence scores changed because component semantics changed.
17. **Exact evidence.** `f73534feb275c5e109fa12fcfd9df5b69c56c03a`; permanent audit run `30497703314`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Review audit requires freshness isolation and forbids reintroducing candidate-age scoring into new Pattern Confidence results.

### GFA-DATA-355 — Pattern Confidence did not verify that selected neighbors agreed on future continuation

1. **Finding / symptom.** Similar prefixes could receive high confidence even when their observed continuations diverged sharply or moved in opposing directions.
2. **Root cause.** Confidence compared historical prefixes but lacked explicit continuation-shape agreement evidence over the requested horizon.
3. **Failure scenario.** Two neighbors look similar at the anchor and later take opposite paths; the pattern is still treated as coherent.
4. **Impact.** Historical continuation can be authorized despite contradictory future evidence.
5. **Severity rationale.** P1 retrospective because continuation agreement is central to using neighbors as predictive evidence.
6. **Existing guarantees violated.** Conservative historical-pattern evidence and continuation plausibility.
7. **Considered solutions.** Trust similarity score; compare only final endpoints; sample complete continuations and measure pairwise spread/divergence.
8. **Chosen remediation.** Sample continuation vectors across the required horizon, compute pairwise spread/divergence, and fail closed when evidence is missing or exceeds policy.
9. **Why selected.** Intermediate divergence and opposing futures remain visible even when paths later reconverge.
10. **Rejected alternatives.** Prefix similarity does not prove future agreement; endpoint-only checks miss intermediate route divergence.
11. **Trade-offs.** Evaluation needs candidate trajectory data and additional O(neighbor-pairs × samples) computation within bounded policy limits.
12. **Regression tests / protection.** Opposing continuation, intermediate divergence, missing candidate, interpolation, pair-count, spread, and divergence consistency are tested.
13. **Adversarial review findings.** The contract explicitly does not claim empirical forecast calibration; it only rejects contradictory observed continuation evidence.
14. **Remediation iterations.** A fifth `continuation_agreement` component and policy snapshot were added in the continuation-agreement wave.
15. **Residual risks / limitations.** Agreement among historical neighbors does not prove the current flight will follow them.
16. **Operational/deployment consequences.** No migration; evaluation requires actual candidate trajectories and may return unavailable more often.
17. **Exact evidence.** `5873ae911b40197ee45eea30e7558aa04af78064`; permanent audit `cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42`, run `30497703314`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent review audit requires continuation evidence, pairwise reconstruction, and divergence regressions.

### GFA-CONTRACT-356 — Production consumers could fall back to a continuation-unaware Pattern Confidence evaluator

1. **Finding / symptom.** Production and Continuation accepted a broad evaluator interface and could runtime-fallback to legacy `Evaluate(selection)` without candidate continuations.
2. **Root cause.** Continuation awareness was optional through type assertion rather than mandatory in the compile-time consumer contract.
3. **Failure scenario.** A custom evaluator implements only the legacy method and production silently authorizes confidence without continuation agreement.
4. **Impact.** The hardened evidence requirement can be bypassed by dependency substitution while all types still compile.
5. **Severity rationale.** P1 retrospective because an incomplete collaborator can bypass a mandatory authorization input.
6. **Existing guarantees violated.** Compile-time dependency contract integrity and fail-closed continuation evidence.
7. **Considered solutions.** Keep runtime capability detection; reject legacy implementations at runtime; make `EvaluateWithContinuations` the required interface method.
8. **Chosen remediation.** Require the continuation-aware method in both production consumer interfaces and remove fallback calls.
9. **Why selected.** Invalid collaborators fail at compile time instead of silently weakening production semantics.
10. **Rejected alternatives.** Runtime fallback preserves precisely the bypass the hardening was intended to close.
11. **Trade-offs.** Custom implementations must update to the stronger interface; the concrete legacy method remains only non-authorizing compatibility API.
12. **Regression tests / protection.** Compile-time interface assertions, mandatory consumer tests, and historical candidate propagation are permanent.
13. **Adversarial review findings.** The legacy concrete method was deliberately retained for source compatibility but cannot satisfy production consumer interfaces alone.
14. **Remediation iterations.** Continuation support was introduced with optional detection, then strengthened to a mandatory interface in a separate follow-up.
15. **Residual risks / limitations.** A malicious collaborator satisfying the interface can still return invalid data, which is separately constrained by result validation.
16. **Operational/deployment consequences.** No migration; downstream custom evaluator implementations require code changes to compile.
17. **Exact evidence.** Optional phase `5873ae911b40197ee45eea30e7558aa04af78064`; mandatory interface `e31fcb5bbbb76093305e8b2c137c793a85dc6795`; permanent audit run `30497703314`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Compile-time interfaces and strict audit forbid continuation-unaware production fallback.

### GFA-DATA-357 — Pattern Confidence validation did not independently reconstruct the published decision

1. **Finding / symptom.** A manually assembled result could retain internally inconsistent policy, components, score, status, usability, confidence level, or continuation evidence.
2. **Root cause.** Validation checked structure more weakly than the evaluator's full decision function.
3. **Failure scenario.** A custom collaborator mutates component scores and total score together, or publishes high confidence without valid continuation agreement, and the result still appears structurally valid.
4. **Impact.** Downstream modules can accept forged or stale Pattern Confidence evidence.
5. **Severity rationale.** P1 retrospective because Pattern Confidence is an authorization input for later projection stages.
6. **Existing guarantees violated.** Cross-field integrity, collaborator distrust, and reproducible decision semantics.
7. **Considered solutions.** Trust evaluator construction; validate only ranges; reconstruct the entire canonical policy/component/decision state in `Result.Validate()`.
8. **Chosen remediation.** Recompute policy validity, counts, similarity aggregates, five-component catalog, weighted score, status/usability/level, limitations, IDs, and continuation spread/divergence consistency.
9. **Why selected.** Validation becomes an independent semantic verifier rather than a shape check.
10. **Rejected alternatives.** Range-only validation cannot detect coordinated mutations; trusting producers breaks the consumer-side integrity boundary.
11. **Trade-offs.** Validation duplicates deterministic formulas intentionally and must evolve with versioned semantics.
12. **Regression tests / protection.** Policy/component mutation, weighted score, continuation pair counts, spread/divergence, limitations, and decision reconstruction are tested.
13. **Adversarial review findings.** Fixed versioned components were retained specifically because they make independent reconstruction possible.
14. **Remediation iterations.** Validation strengthened across the contract, distribution, and continuation waves as new evidence became authoritative.
15. **Residual risks / limitations.** Validator correctness remains part of the trusted computing base and is therefore protected by the permanent source audit.
16. **Operational/deployment consequences.** No migration; malformed custom or stale fixtures now fail validation.
17. **Exact evidence.** `6e6ac17cfcfca688d57829adfe2468346db6db1a`, `f73534feb275c5e109fa12fcfd9df5b69c56c03a`, `5873ae911b40197ee45eea30e7558aa04af78064`; permanent audit `cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42`, run `30497703314`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Strict `Result.Validate()` plus `projectionpatternconfidencereviewaudit` permanently protect reconstruction semantics.
