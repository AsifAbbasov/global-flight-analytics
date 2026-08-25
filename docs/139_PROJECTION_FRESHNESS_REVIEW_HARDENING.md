# Projection Freshness Review Hardening

Status: closed

```text
LINEAGE_AGE_INTEGRITY_COMMIT=0b47aa3231c93d573a6026651a4085d376a40583
LINEAGE_AGE_INTEGRITY_GITHUB_ACTIONS_RUN=30502639621
POLICY_DECISION_INTEGRITY_COMMIT=072d0eb349fcd0e42c1d3c0bcf54c51cefb08a19
POLICY_DECISION_INTEGRITY_GITHUB_ACTIONS_RUN=30503845277
PERMANENT_AUDIT_COMMIT=619e24878a5025decf6fe21abddba537ce195560
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30523502590
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90809046060
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90809046046
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90809046013
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90809225151
WEIGHT_POLICY_CORRECTION_COMMIT=e3e99758d6f654db12ccce32ec55ad1339fb518f
WEIGHT_POLICY_CORRECTION_GITHUB_ACTIONS_RUN=30527541240
WEIGHT_POLICY_CORRECTION_BACKEND_QUALITY_JOB=90821894564
WEIGHT_POLICY_CORRECTION_BACKEND_RACE_SAFETY_JOB=90821894465
WEIGHT_POLICY_CORRECTION_POSTGRESQL_16_INTEGRATION_JOB=90821894536
WEIGHT_POLICY_CORRECTION_BACKEND_CONTAINER_JOB=90822090361
WEIGHT_POLICY_CONSISTENCY=CLOSED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_FRESHNESS_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_FRESHNESS_ENGINEERING_DEBT=CLOSED
PROJECTION_FRESHNESS_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_FRESHNESS_REVIEW_STATUS=CLOSED
```

## 1. Scope

This record covers the dedicated review and hardening of:

```text
apps/api/internal/projectionintelligence/projectionfreshness
```

It also covers the lineage and production-fixture boundaries through:

```text
projectionneighbors
projectionpatternconfidence
projectionproduction
projectioncontinuation
projectionread
```

The review is limited to deterministic selected-neighbor age evidence, policy,
lineage, scoring, decision semantics, validation, fingerprinting, and production
contract integrity. It does not claim empirical forecast accuracy, operational
aviation suitability, calibrated probability semantics, or that historical
freshness alone proves future trajectory quality.

## 2. Findings closed

### 2.1 Pattern usability and exact lineage

Freshness now blocks historical continuation when Pattern Confidence is not usable.
The evaluator requires all of the following to agree before freshness is evaluated:

```text
Pattern Confidence source-selection fingerprint
Neighbor Selection input fingerprint
Pattern Confidence selection status
Neighbor Selection status
sorted selected trajectory identifiers
```

A Pattern Confidence result from another selection, another selection state, or
another selected-neighbor catalog cannot be reused silently.

### 2.2 Timestamp-derived age evidence

Freshness derives every candidate age from:

```text
Neighbor Selection AsOfTime - CandidateEndTime
```

The upstream `CandidateAge` field is validated by Neighbor Selection, but Freshness
does not treat the duplicated duration as its source of truth. This preserves one
canonical time calculation at the decision boundary.

The arithmetic mean uses quotient-and-remainder accumulation rather than summing all
nanoseconds first, preventing signed 64-bit overflow for valid large durations.

### 2.3 Configuration contract hardening

Configuration validation now requires:

```text
zero < maximum newest age <= maximum mean age <= maximum oldest age
zero < recent-neighbor age limit <= maximum oldest age
zero < minimum recent-neighbor count <= target recent-neighbor count
zero < minimum usable score <= complete score minimum <= one
finite positive component weights
component-weight total equal to one
```

Zero score thresholds and incoherent age ordering are rejected before an evaluator
is created.

### 2.3.1 Strictly positive component-weight correction

The original closure document stated that every component weight is finite and
strictly positive, while the implementation rejected only negative weights. That
allowed a zero weight to disable one of the four canonical Freshness components while
keeping the remaining weights normalized to one.

The correction changes both configuration and result validation to reject every
non-finite or non-positive component weight. Regression tests cover zero values for
newest age, mean age, oldest age, and recent support, including a coordinated Result
mutation whose policy, component catalog, weighted score, decision, usability, and
limitations otherwise remain internally consistent.

The permanent audit requires the `<= 0` guards and forbids the former `< 0`
contracts. Corrective commit `e3e99758d6f654db12ccce32ec55ad1339fb518f` completed exact Continuous
Integration run `30527541240`, so the code, tests, permanent audit, and formal
weight-policy contract are now aligned.

### 2.4 Semantic freshness fingerprint

The input fingerprint binds:

```text
Freshness fingerprint version
Neighbor Selection input fingerprint and status
Pattern Confidence input fingerprint
Pattern Confidence source-selection fingerprint
Pattern Confidence selection status
Pattern Confidence status and usability
normalized AsOfTime
all normalized policy thresholds and weights
sorted selected trajectory identifiers
candidate end timestamps
timestamp-derived candidate ages
```

Upstream status or usability changes therefore produce a different Freshness input
identity even when selected trajectory identifiers remain unchanged.

### 2.5 Complete hard-violation reporting

The policy evaluator accumulates every simultaneous blocking reason rather than
returning only the first matching switch branch. The result can report all applicable
violations, including:

```text
historical neighbors unavailable
Pattern Confidence unusable
newest selected neighbor too old
mean selected-neighbor age too old
oldest selected neighbor too old
insufficient recent-neighbor support
Freshness score below the usable minimum
```

Limitations are normalized, sorted, and unique.

### 2.6 Policy and upstream-state snapshots

Every Result publishes the normalized policy that created it and the upstream state
that authorized its calculation:

```text
age thresholds
recent-support thresholds
score thresholds
component weights
Neighbor Selection status
Pattern Confidence status
Pattern Confidence usability
source Neighbor Selection fingerprint
source Pattern Confidence fingerprint
```

The snapshot makes the result independently inspectable without relying on mutable
runtime evaluator configuration.

### 2.7 Component and decision reconstruction

`Result.Validate()` reconstructs and verifies:

```text
policy validity
upstream status and usability consistency
neighbor counts and ordered age aggregates
canonical four-component catalog
component scores from aggregate evidence
component weights from the policy snapshot
weighted total score
Decision
Usable
exact required limitations
sorted unique trajectory identifiers
sorted unique limitations
source lineage fingerprints
```

A caller cannot make a coordinated mutation to component scores, total score,
policy, decision, usability, or limitations and still pass validation.

### 2.8 Production fixture contract integrity

Production tests no longer hand-assemble nominal Freshness results. Allowed, limited,
and blocked fixtures are generated through the real evaluator with valid Neighbor
Selection and Pattern Confidence inputs.

A dedicated fixture-contract test validates policy snapshots, source fingerprints,
and the final Freshness result so future contract changes cannot leave stale test
objects that bypass production semantics.

## 3. Deliberately retained and rejected recommendations

The following contracts were reviewed and deliberately retained:

```text
four fixed components as a versioned domain schema
float64 arithmetic with explicit finite checks and comparison tolerance
idiomatic New(Config) (*Evaluator, error) constructor
Usable compatibility fields with cross-field validation
blocked domain result for unusable upstream pattern evidence
small local numeric and normalization helpers
```

The following recommendations were rejected as unsupported or outdated mechanical
findings:

```text
mandatory integer basis points
mandatory dynamic component registry
constructor criticism based only on nil plus error
claim that candidate age is not validated upstream
claim that Freshness remains duplicated inside Pattern Confidence
blanket removal of every compatibility usability field
blanket rejection of small local helpers
```

The fixed component catalog is intentional because `Result.Validate()` reconstructs
a stable, versioned Freshness schema. Floating-point values remain appropriate for
normalized ratios when all inputs are finite and equality-sensitive checks use an
explicit comparison tolerance.

## 4. Permanent regression coverage

Permanent tests cover:

```text
unusable Pattern Confidence blocking
source-selection fingerprint mismatch
selection-status lineage mismatch
upstream status fingerprint mutation
timestamp-derived age consistency
signed 64-bit mean-duration overflow avoidance
ordered age-threshold validation
positive score-threshold validation
all simultaneous hard violations
weighted-score mutation rejection
policy snapshot mutation rejection
coordinated component-and-score mutation rejection
decision-and-limitation mutation rejection
upstream status and usability mutation rejection
source Pattern Confidence fingerprint mutation
valid limited-decision reconstruction
production fixture evaluator generation
production fixture contract drift
```

## 5. Permanent audit gate

The source audit is:

```text
apps/api/tools/projectionfreshnessreviewaudit
```

Backend Continuous Integration executes:

```text
go run ./tools/projectionfreshnessreviewaudit -strict
```

The gate protects the hardened implementation, regression tests, production fixture
contracts, workflow wiring, Stage 9 closure markers, this authoritative record, and
the Documentation Index entry.

## 6. Closure evidence

The permanent audit commit completed every Backend Continuous Integration job:

```text
PERMANENT_AUDIT_COMMIT=619e24878a5025decf6fe21abddba537ce195560
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30523502590
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90809046060
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90809046046
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90809046013
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90809225151
```

The `Backend Quality` job executed and passed the dedicated step:

```text
Run projection freshness review audit
```

The strictly positive component-weight correction completed every required Backend
Continuous Integration job:

```text
WEIGHT_POLICY_CORRECTION_COMMIT=e3e99758d6f654db12ccce32ec55ad1339fb518f
WEIGHT_POLICY_CORRECTION_GITHUB_ACTIONS_RUN=30527541240
WEIGHT_POLICY_CORRECTION_BACKEND_QUALITY_JOB=90821894564
WEIGHT_POLICY_CORRECTION_BACKEND_RACE_SAFETY_JOB=90821894465
WEIGHT_POLICY_CORRECTION_POSTGRESQL_16_INTEGRATION_JOB=90821894536
WEIGHT_POLICY_CORRECTION_BACKEND_CONTAINER_JOB=90822090361
```

The corrective `Backend Quality` job executed and passed the dedicated step:

```text
Run projection freshness review audit
```

Engineering implementation and formal closure documentation are complete. No
confirmed finding remains open, unclassified, or deferred. The permanent audit gate
remains mandatory in Backend Continuous Integration.

## Canonical remediation history

The following records retrospectively normalize the accepted review findings into the
repository-wide nineteen-field standard. Severity values are retrospective engineering
classifications. Exact historical CI evidence is claimed only where this source records
the corresponding run or job identifiers.

### GFA-DATA-358 — Freshness accepted Pattern Confidence without exact selection lineage and usability agreement

1. **Finding / symptom.** Freshness could consume Pattern Confidence evidence without proving it belonged to the same Neighbor Selection state and selected trajectory catalog.
2. **Root cause.** Upstream result compatibility was inferred from partial identity rather than one complete lineage contract.
3. **Failure scenario.** A Pattern Confidence result from another selection, another status, or another neighbor set is paired with the current selection and used to authorize historical continuation.
4. **Impact.** Freshness can approve or reject continuation using evidence that was never computed for the selected neighbors.
5. **Severity rationale.** P1 retrospective because mismatched authorization evidence can directly change whether a projection path is allowed.
6. **Existing guarantees violated.** Exact provenance lineage, collaborator integrity, and fail-closed historical authorization.
7. **Considered solutions.** Trust caller composition; compare only trajectory IDs; require fingerprints, statuses, usability, and sorted selected IDs to agree.
8. **Chosen remediation.** Validate Pattern source-selection fingerprint, Neighbor Selection fingerprint/status, Pattern status/usability, and selected trajectory identity before Freshness evaluation.
9. **Why selected.** It makes the consumer independently verify the full evidence chain rather than trusting orchestration history.
10. **Rejected alternatives.** ID-only checks miss changed selection state and policy identity; caller trust is not a durable contract.
11. **Trade-offs.** Freshness carries and validates more upstream lineage fields.
12. **Regression tests / protection.** Tests cover unusable Pattern Confidence, source-selection mismatch, selection-status mismatch, and upstream status mutation.
13. **Adversarial review findings.** Unusable Pattern Confidence remains a blocked domain result rather than an infrastructure exception by design.
14. **Remediation iterations.** Exact lineage and age ownership were established in the lineage-integrity wave, then included in policy snapshots/fingerprints.
15. **Residual risks / limitations.** Correctness still depends on upstream fingerprints truthfully binding their own evidence.
16. **Operational/deployment consequences.** No migration; stale or mismatched composed results now fail closed.
17. **Exact evidence.** `0b47aa3231c93d573a6026651a4085d376a40583`, run `30502639621`; permanent audit `619e24878a5025decf6fe21abddba537ce195560`, run `30523502590`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectionfreshnessreviewaudit` and lineage-mismatch regressions require exact upstream ownership.

### GFA-DATA-359 — Freshness trusted duplicated candidate-age evidence instead of canonical timestamps

1. **Finding / symptom.** Candidate age could be treated as an authoritative copied duration rather than reconstructed at the Freshness decision boundary.
2. **Root cause.** Age existed both as upstream metadata and as a value derivable from `AsOfTime - CandidateEndTime`.
3. **Failure scenario.** A stale or inconsistent copied age disagrees with the selected candidate timestamp and changes recent-support or age-threshold decisions.
4. **Impact.** Freshness scoring and blocking semantics can be driven by duplicated, contradictory evidence.
5. **Severity rationale.** P1 retrospective because age is a direct authorization input.
6. **Existing guarantees violated.** Single source of temporal truth and deterministic replay semantics.
7. **Considered solutions.** Trust upstream `CandidateAge`; cross-check and still use it; derive age only from canonical timestamps.
8. **Chosen remediation.** Recompute every age from Selection `AsOfTime` and candidate `EndTime`; keep upstream age only under its upstream validation contract.
9. **Why selected.** Timestamp evidence is canonical and removes competing duration ownership.
10. **Rejected alternatives.** Copied durations can drift even when identifiers remain unchanged.
11. **Trade-offs.** Freshness repeats a cheap deterministic subtraction at evaluation time.
12. **Regression tests / protection.** Timestamp-derived age consistency tests cover the decision boundary.
13. **Adversarial review findings.** The review explicitly rejected the claim that upstream age was wholly unvalidated; the defect was treating duplication as decision truth.
14. **Remediation iterations.** Age derivation moved into Freshness metrics during lineage hardening.
15. **Residual risks / limitations.** Candidate `EndTime` itself must remain truthful upstream evidence.
16. **Operational/deployment consequences.** No migration; inconsistent copied ages no longer control Freshness.
17. **Exact evidence.** `0b47aa3231c93d573a6026651a4085d376a40583`, run `30502639621`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit requires timestamp-derived age semantics and the regression test.

### GFA-DATA-360 — Mean selected-neighbor age accumulation could overflow signed 64-bit duration arithmetic

1. **Finding / symptom.** Summing valid large `time.Duration` nanosecond values before division could overflow `int64`.
2. **Root cause.** Mean-age computation used additive duration accumulation without overflow-safe decomposition.
3. **Failure scenario.** Several individually valid large ages overflow the sum and yield a negative or corrupted mean.
4. **Impact.** Freshness score and maximum-mean-age blocking can become mathematically false.
5. **Severity rationale.** P2 retrospective because the defect requires extreme valid durations but corrupts a decision metric when triggered.
6. **Existing guarantees violated.** Finite deterministic arithmetic and fail-closed temporal scoring.
7. **Considered solutions.** Restrict maximum age further; use arbitrary precision; compute quotient/remainder accumulation without forming the full sum.
8. **Chosen remediation.** Use quotient-and-remainder mean accumulation that stays inside signed duration range.
9. **Why selected.** It preserves the existing duration domain without extra dependencies or arbitrary policy narrowing.
10. **Rejected alternatives.** Narrower thresholds would change product policy; big integers are unnecessary for a bounded mean.
11. **Trade-offs.** Mean computation is slightly less obvious than sum-then-divide.
12. **Regression tests / protection.** Signed-64-bit mean-duration overflow avoidance is covered permanently.
13. **Adversarial review findings.** The fix protects valid large durations, not invalid negative ages, which are rejected separately.
14. **Remediation iterations.** Overflow-safe arithmetic was added with timestamp-derived metrics.
15. **Residual risks / limitations.** Other newly introduced duration aggregates require their own overflow review.
16. **Operational/deployment consequences.** No migration or API change.
17. **Exact evidence.** `0b47aa3231c93d573a6026651a4085d376a40583`, run `30502639621`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Regression coverage keeps the quotient/remainder arithmetic contract visible.

### GFA-CONTRACT-361 — Freshness configuration allowed zero or internally incoherent age, support, and score thresholds

1. **Finding / symptom.** Freshness policy could admit threshold combinations that were zero-information, reversed, or impossible to interpret consistently.
2. **Root cause.** Field-level validation did not fully enforce ordered age thresholds, support relationships, and strictly positive score policy.
3. **Failure scenario.** Newest age exceeds mean age, recent limit exceeds oldest age, minimum support exceeds target, or a zero usable score allows evidence with no positive score requirement.
4. **Impact.** Freshness decisions become unreachable, contradictory, or too permissive by construction.
5. **Severity rationale.** P1 retrospective because invalid policy can directly authorize or block historical continuation.
6. **Existing guarantees violated.** Constructor-time policy validity and reconstructible decision semantics.
7. **Considered solutions.** Clamp/reorder values; accept zero as a disabled guard; reject incoherent policy explicitly.
8. **Chosen remediation.** Enforce ordered positive age thresholds, bounded recent window, valid support counts, and positive ordered score thresholds before evaluator construction.
9. **Why selected.** Policy mistakes fail explicitly instead of being silently normalized into different semantics.
10. **Rejected alternatives.** Clamping hides configuration defects; zero guard values weaken the intended policy without an explicit disabled mode.
11. **Trade-offs.** Some formerly accepted configurations now fail fast.
12. **Regression tests / protection.** Ordered age-threshold and positive score-threshold tests are permanent.
13. **Adversarial review findings.** Fixed component count and float arithmetic were not classified as defects.
14. **Remediation iterations.** Validation was decomposed into age, support, score, and weight policy checks in the lineage/policy waves.
15. **Residual risks / limitations.** Threshold values remain engineering policy, not empirical calibration.
16. **Operational/deployment consequences.** No migration; invalid startup/configuration fails earlier.
17. **Exact evidence.** `0b47aa3231c93d573a6026651a4085d376a40583`, run `30502639621`; policy wave `072d0eb349fcd0e42c1d3c0bcf54c51cefb08a19`, run `30503845277`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Config validation and strict audit protect threshold relationships.

### GFA-CONTRACT-362 — Zero component weights passed validation despite the strictly-positive Freshness component contract

1. **Finding / symptom.** The documented contract required every component weight to be strictly positive, but implementation rejected only negative values.
2. **Root cause.** Weight validation used `< 0` rather than `<= 0` in configuration and result semantics.
3. **Failure scenario.** One Freshness component is silently disabled with weight zero while the remaining weights still sum to one.
4. **Impact.** A versioned four-component score can claim to use all four components while one contributes nothing.
5. **Severity rationale.** P1 retrospective because disabling a decision component can materially change authorization while preserving apparently valid policy metadata.
6. **Existing guarantees violated.** Fixed component schema, documentation/runtime consistency, and score reconstruction integrity.
7. **Considered solutions.** Permit zero as an explicit disabled mode; remove zero-weight components dynamically; require strictly positive weights.
8. **Chosen remediation.** Reject non-finite and non-positive component weights in both config and result validation.
9. **Why selected.** It matches the declared fixed four-component contract and prevents hidden policy disabling.
10. **Rejected alternatives.** Dynamic omission changes schema identity; a disabled mode was not part of the product contract.
11. **Trade-offs.** Policies that intentionally used zero weight must be redesigned rather than silently accepted.
12. **Regression tests / protection.** Zero-value tests cover every component plus a coordinated mutated Result.
13. **Adversarial review findings.** This remained a separate finding because the original closure record and implementation were demonstrably inconsistent after broader policy hardening.
14. **Remediation iterations.** A dedicated post-audit corrective wave aligned code, tests, audit, and documentation.
15. **Residual risks / limitations.** Positive weights can still be arbitrarily small; that is explicit policy rather than hidden disablement.
16. **Operational/deployment consequences.** No migration; zero-weight config now fails validation.
17. **Exact evidence.** `e3e99758d6f654db12ccce32ec55ad1339fb518f`, exact run `30527541240`; quality `90821894564`, race `90821894465`, PostgreSQL `90821894536`, container `90822090361`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit explicitly requires `<= 0` guards and forbids the former `< 0` contract.

### GFA-DATA-363 — Freshness fingerprint omitted decision-relevant upstream state, age evidence, or policy identity

1. **Finding / symptom.** Freshness input identity was not guaranteed to change for every upstream state, timestamp-derived age, or normalized policy change capable of changing the result.
2. **Root cause.** Fingerprinting lagged behind the full decision surface.
3. **Failure scenario.** Pattern usability/status, selection status, age timestamps, thresholds, or weights change while a stale fingerprint is reused.
4. **Impact.** Distinct Freshness decisions can collide in provenance or downstream lineage checks.
5. **Severity rationale.** P1 retrospective because the fingerprint is a durable evidence identity consumed across projection composition.
6. **Existing guarantees violated.** Semantic fingerprinting, reproducibility, and change-sensitive provenance.
7. **Considered solutions.** Hash only selected IDs; hash result outputs; hash every normalized decision input and upstream identity.
8. **Chosen remediation.** Bind selection/pattern fingerprints and statuses, Pattern usability, AsOf, policy, IDs, candidate end times, and timestamp-derived ages.
9. **Why selected.** It ties identity to the actual evidence and policy that determine Freshness.
10. **Rejected alternatives.** IDs alone miss changed upstream state; output hashing conflates inputs and derived decisions.
11. **Trade-offs.** Fingerprint construction and version maintenance are broader.
12. **Regression tests / protection.** Upstream-status and source-Pattern-fingerprint mutation tests protect the semantic identity.
13. **Adversarial review findings.** Candidate age is fingerprinted as derived timestamp evidence rather than trusted copied duration.
14. **Remediation iterations.** Fingerprint expansion followed lineage/policy normalization.
15. **Residual risks / limitations.** Future decision inputs must be explicitly added to the fingerprint contract.
16. **Operational/deployment consequences.** Fingerprint identity changes; no database migration.
17. **Exact evidence.** `072d0eb349fcd0e42c1d3c0bcf54c51cefb08a19`, run `30503845277`; permanent audit `619e24878a5025decf6fe21abddba537ce195560`, run `30523502590`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Strict review audit and fingerprint-mutation tests protect decision-input coverage.

### GFA-DATA-364 — Freshness reported only one blocking reason when multiple hard violations coexisted

1. **Finding / symptom.** Policy evaluation could stop at the first failed guard instead of preserving all simultaneously true denial evidence.
2. **Root cause.** Decision control flow favored early-return/single-branch classification over complete limitation accumulation.
3. **Failure scenario.** Historical evidence is both too old and insufficiently recent, but only one reason is published.
4. **Impact.** Operators and downstream consumers receive incomplete denial provenance and can misdiagnose which evidence must change.
5. **Severity rationale.** P2 retrospective because the decision remains conservative but its explanation is incomplete.
6. **Existing guarantees violated.** Complete limitation evidence and reproducible decision diagnostics.
7. **Considered solutions.** Keep first reason; publish generic blocked; accumulate all hard violations deterministically.
8. **Chosen remediation.** Evaluate every hard guard, normalize/deduplicate/sort limitations, then derive blocked status.
9. **Why selected.** One result now explains the full state without changing the conservative decision boundary.
10. **Rejected alternatives.** Generic or first-only messages are insufficient for audit and operations.
11. **Trade-offs.** Evaluation constructs a larger limitation set.
12. **Regression tests / protection.** A dedicated all-simultaneous-hard-violations test protects this behavior.
13. **Adversarial review findings.** Limited-but-usable reasons remain a separate second decision phase after hard blockers.
14. **Remediation iterations.** Policy evaluation was extracted into a complete decision function.
15. **Residual risks / limitations.** Human-readable messages remain explanatory metadata; codes are the stable contract.
16. **Operational/deployment consequences.** No migration; blocked results may contain more limitation entries.
17. **Exact evidence.** `072d0eb349fcd0e42c1d3c0bcf54c51cefb08a19`, run `30503845277`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Result validation and regression coverage require exact required limitation sets.

### GFA-DATA-365 — Freshness results lacked immutable policy and upstream-state snapshots

1. **Finding / symptom.** A published score/decision could not be independently inspected without the evaluator's mutable runtime configuration and upstream state.
2. **Root cause.** Result contracts exposed derived values but not the normalized policy and authorization state that produced them.
3. **Failure scenario.** Configuration changes after evaluation and an archived Freshness result no longer reveals which thresholds, weights, statuses, or fingerprints governed it.
4. **Impact.** Replay, audit, and provenance reconstruction become ambiguous.
5. **Severity rationale.** P1 retrospective because durable analytical evidence must remain self-describing after runtime state changes.
6. **Existing guarantees violated.** Provenance completeness and independent result inspectability.
7. **Considered solutions.** Depend on external config logs; retain only a policy hash; publish immutable policy and upstream snapshots in the result.
8. **Chosen remediation.** Store normalized thresholds, weights, upstream statuses/usability, and source fingerprints with each result.
9. **Why selected.** Consumers can validate and explain the result without reconstructing hidden runtime state.
10. **Rejected alternatives.** Hash-only evidence proves identity but does not expose semantics; external logs are not part of the result contract.
11. **Trade-offs.** Result schema is larger and versioned policy fields must be maintained.
12. **Regression tests / protection.** Policy snapshot, upstream status/usability, and source-fingerprint mutation tests are permanent.
13. **Adversarial review findings.** Compatibility usability fields were retained only with cross-field validation.
14. **Remediation iterations.** Policy/upstream snapshots were added alongside semantic fingerprinting and decision reconstruction.
15. **Residual risks / limitations.** Snapshot completeness depends on correctly classifying all future decision-relevant policy fields.
16. **Operational/deployment consequences.** No migration in this review; public/domain result payload grows.
17. **Exact evidence.** `072d0eb349fcd0e42c1d3c0bcf54c51cefb08a19`, run `30503845277`; audit run `30523502590`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `Result.Validate()` and review audit require policy and upstream-state snapshots.

### GFA-DATA-366 — Freshness validation did not independently reconstruct component and decision semantics

1. **Finding / symptom.** Structurally plausible Freshness results could contain coordinated mutations to components, score, decision, usability, limitations, or lineage.
2. **Root cause.** Validation did not fully recompute the evaluator's mathematical and policy decision function.
3. **Failure scenario.** A custom producer changes component scores and total together, or marks blocked evidence usable, while range checks still pass.
4. **Impact.** Forged or stale Freshness evidence can enter later Projection stages.
5. **Severity rationale.** P1 retrospective because Freshness is an authorization boundary for historical continuation.
6. **Existing guarantees violated.** Consumer-side distrust, cross-field integrity, and reproducible decision semantics.
7. **Considered solutions.** Trust the evaluator; validate only ranges/catalog shape; reconstruct the full canonical result from published evidence.
8. **Chosen remediation.** Recompute policy validity, counts/ages, four components, weights, total score, decision/usability, exact limitations, IDs, and source lineage.
9. **Why selected.** Validation becomes an independent semantic verifier rather than a serializer check.
10. **Rejected alternatives.** Range-only validation cannot detect coordinated mutations.
11. **Trade-offs.** Deterministic formulas are intentionally duplicated in validation and must evolve with versioned semantics.
12. **Regression tests / protection.** Weighted-score, policy, coordinated component/score, decision/limitation, status/usability, and source-fingerprint mutation tests are permanent.
13. **Adversarial review findings.** Fixed four-component schema is retained partly because it enables exact reconstruction.
14. **Remediation iterations.** Validation hardened after policy snapshots and complete decision logic were established.
15. **Residual risks / limitations.** Validator code is part of the trusted computing base and therefore protected by the source audit.
16. **Operational/deployment consequences.** No migration; stale/custom malformed results now fail validation.
17. **Exact evidence.** `072d0eb349fcd0e42c1d3c0bcf54c51cefb08a19`, run `30503845277`; permanent audit `619e24878a5025decf6fe21abddba537ce195560`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Strict `Result.Validate()` and permanent audit protect reconstruction semantics.

### GFA-TEST-367 — Production Freshness fixtures bypassed the real evaluator contract

1. **Finding / symptom.** Production tests could hand-assemble nominal Freshness results that no longer matched evaluator policy, lineage, or validation semantics.
2. **Root cause.** Fixtures duplicated domain result construction instead of using the production evaluator.
3. **Failure scenario.** The Freshness contract evolves, production fixture literals stay stale, and integration tests continue exercising impossible result states.
4. **Impact.** Tests can report green while production composition is incompatible with the actual Freshness contract.
5. **Severity rationale.** P2 retrospective because this is a regression-protection gap capable of masking production contract drift.
6. **Existing guarantees violated.** Test realism and contract parity between unit/integration fixtures and runtime producers.
7. **Considered solutions.** Keep literals synchronized manually; add fixture-only validators; generate fixtures through the real evaluator and validate them.
8. **Chosen remediation.** Build allowed/limited/blocked production fixtures through real Neighbor Selection, Pattern Confidence, and Freshness evaluators; add fixture-contract validation.
9. **Why selected.** Tests exercise the same construction path and invariants as production evidence.
10. **Rejected alternatives.** Manual fixture maintenance is precisely the drift source.
11. **Trade-offs.** Fixtures require more setup and can intentionally fail when upstream contracts change.
12. **Regression tests / protection.** Dedicated evaluator-generation and production fixture-contract drift tests are permanent.
13. **Adversarial review findings.** Fixture failures are treated as contract evidence, not proof of forecast quality.
14. **Remediation iterations.** Production fixture construction was replaced after result validation became strict enough to expose stale literals.
15. **Residual risks / limitations.** Generated fixtures still depend on the correctness of the real evaluator and upstream fixture inputs.
16. **Operational/deployment consequences.** Test-only construction change; no runtime migration.
17. **Exact evidence.** Policy/decision integrity `072d0eb349fcd0e42c1d3c0bcf54c51cefb08a19`; permanent audit `619e24878a5025decf6fe21abddba537ce195560`, exact run `30523502590`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Dedicated fixture-contract test and Backend CI audit prevent silent fixture drift.
