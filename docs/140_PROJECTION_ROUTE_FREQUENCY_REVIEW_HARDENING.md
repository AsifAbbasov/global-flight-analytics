# Projection Route Frequency Review Hardening

Status: closed

```text
EVIDENCE_ISOLATION_COMMIT=c6fff15f8d0c770197db40a69d54f8856044d8d2
EVIDENCE_ISOLATION_GITHUB_ACTIONS_RUN=30534243693
EVIDENCE_ISOLATION_POSTGRESQL_16_INTEGRATION_JOB=90843560950
EVIDENCE_ISOLATION_BACKEND_QUALITY_JOB=90843560991
EVIDENCE_ISOLATION_BACKEND_RACE_SAFETY_JOB=90843560999
EVIDENCE_ISOLATION_BACKEND_CONTAINER_JOB=90843853362
POLICY_DECISION_INTEGRITY_COMMIT=ee7c79bc8213dc030ce0d98f13d1065c9bb96275
POLICY_DECISION_INTEGRITY_GITHUB_ACTIONS_RUN=30544636679
POLICY_DECISION_INTEGRITY_POSTGRESQL_16_INTEGRATION_JOB=90877578926
POLICY_DECISION_INTEGRITY_BACKEND_RACE_SAFETY_JOB=90877578928
POLICY_DECISION_INTEGRITY_BACKEND_QUALITY_JOB=90877579007
POLICY_DECISION_INTEGRITY_BACKEND_CONTAINER_JOB=90877915808
PERMANENT_AUDIT_COMMIT=6f039b33c96cdb67370158b0eda5d0fc87593de5
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30548438062
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90890525039
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90890525126
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90890525150
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90890829745
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_ROUTE_FREQUENCY_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_ROUTE_FREQUENCY_ENGINEERING_DEBT=CLOSED
PROJECTION_ROUTE_FREQUENCY_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_ROUTE_FREQUENCY_REVIEW_STATUS=CLOSED
```

## 1. Scope

This record covers the dedicated code review and hardening of:

```text
apps/api/internal/projectionintelligence/projectionroutefrequency
```

It also covers the production evidence boundary through:

```text
apps/api/internal/projectionintelligence/projectionread
apps/api/internal/projectionintelligence/projectionproduction
apps/api/cmd/verify-postgres-projection-intelligence-historical-http-api
```

The review is limited to historical route evidence isolation, logical-flight
deduplication, exposure-window integrity, scoring, decision semantics, validation,
fingerprinting, production fixture correctness, and permanent regression enforcement.

It does not claim empirical route-prediction accuracy, operational aviation
suitability, authoritative schedule knowledge, or that historical route frequency
proves future aircraft intent.

## 2. Confirmed production findings corrected

### 2.1 Current trajectory and current flight leakage

Route-history evidence now excludes both:

```text
current trajectory_id
current flight_id
```

The current prediction target therefore cannot act as historical support for itself.
The correction applies before observation count, recent support, distinct UTC-day
support, and latest-observation age are calculated.

### 2.2 Logical-flight evidence deduplication

Multiple trajectories belonging to one logical flight no longer inflate route
frequency. The PostgreSQL query builds a stable evidence identity:

```text
flight_id when available
trajectory_id only when flight_id is unavailable
```

The latest route result is selected per evidence identity before aggregate counts are
computed.

### 2.3 Distinct-flight ownership

The evaluator uses `DistinctFlightCount`, not raw trajectory count, for:

```text
observation support score
minimum historical support guard
```

`ObservationCount` remains available as evidence metadata, but it does not silently
override the logical-flight contract.

### 2.4 Full and recent exposure windows

The domain configuration now publishes:

```text
HistoryWindow
RecentWindow
```

The evaluator requires exact boundaries:

```text
WindowStart       = AsOfTime - HistoryWindow
WindowEnd         = AsOfTime
RecentWindowStart = AsOfTime - RecentWindow
```

Production policy fixes the windows at:

```text
HistoryWindow = 180 days
RecentWindow  = 30 days
```

The Projection Read policy requires exact equality between data-source windows and the
Route Frequency evaluator windows.

### 2.5 Configuration integrity

Configuration validation rejects:

```text
TargetDistinctDayCount greater than TargetObservationCount
TargetRecentObservationCount greater than TargetObservationCount
RecentWindow greater than HistoryWindow
zero route-confidence threshold
zero usable-score threshold
zero complete-score threshold
zero or negative component weights
non-finite values
component weights that do not sum to one
```

These checks prevent unreachable targets and nominally allowed zero-information
decisions.

### 2.6 Complete blocking evidence

The evaluator accumulates every applicable hard violation rather than returning only
the first matching reason. A blocked result can simultaneously explain:

```text
route confidence below minimum
distinct-flight support below minimum
distinct-day support below minimum
recent support below minimum
latest historical observation too old
aggregate score below minimum
```

Limitations are normalized, deduplicated, and deterministically ordered.

### 2.7 Weighted-score reconstruction

`Result.Validate()` independently reconstructs:

```text
sum(Component.Score multiplied by Component.Weight)
```

The reconstructed score must match the published aggregate score within the explicit
comparison tolerance. Component weights must be strictly positive and sum to one.

This validation exposed and corrected a stale production fixture:

```text
old published score: 0.85
actual weighted score: 0.94
```

### 2.8 Semantic fingerprint

The Route Frequency fingerprint binds:

```text
contract and fingerprint versions
Route Intelligence schema version and status
trajectory and flight identifiers
resolved route availability and canonical route key
same-airport state
route confidence
Route Intelligence provenance fingerprint
route-history fingerprint
full and recent window boundaries
historical support counts and latest observation
all policy thresholds
all component weights
```

Decision-relevant mutations therefore change the fingerprint.

## 3. Exact evidence already closed

The evidence-isolation commit completed all four Backend Continuous Integration jobs:

```text
COMMIT=c6fff15f8d0c770197db40a69d54f8856044d8d2
GITHUB_ACTIONS_RUN=30534243693
POSTGRESQL_16_INTEGRATION_JOB=90843560950
BACKEND_QUALITY_JOB=90843560991
BACKEND_RACE_SAFETY_JOB=90843560999
BACKEND_CONTAINER_JOB=90843853362
```

This exact evidence closes current-target leakage and multiple-trajectory inflation at
the committed baseline.

## 4. Policy and decision integrity evidence closed

The policy and decision integrity implementation completed all four exact Backend
Continuous Integration jobs:

```text
POLICY_DECISION_INTEGRITY_COMMIT=ee7c79bc8213dc030ce0d98f13d1065c9bb96275
POLICY_DECISION_INTEGRITY_GITHUB_ACTIONS_RUN=30544636679
POLICY_DECISION_INTEGRITY_POSTGRESQL_16_INTEGRATION_JOB=90877578926
POLICY_DECISION_INTEGRITY_BACKEND_RACE_SAFETY_JOB=90877578928
POLICY_DECISION_INTEGRITY_BACKEND_QUALITY_JOB=90877579007
POLICY_DECISION_INTEGRITY_BACKEND_CONTAINER_JOB=90877915808
```

This exact evidence closes exposure-window, policy-target, decision-reporting,
weighted-score, and semantic-fingerprint integrity at the committed baseline.

## 5. Deliberately retained and rejected recommendations

The following contracts are deliberately retained:

```text
fixed five-component versioned schema
float64 normalized scores with finite checks and explicit tolerance
idiomatic New(Config) (*Evaluator, error)
Usable compatibility field with cross-field validation
count-based support inside a strictly fixed exposure window
small local fingerprint and normalization helpers
```

The following mechanical recommendations are rejected:

```text
mandatory basis-point conversion for non-monetary scores
mandatory dynamic component registry
constructor criticism based only on nil plus error
blanket removal of every compatibility field
function-size criticism without a correctness or maintenance defect
renaming tests merely because a name contains the word And
```

A route-frequency rate per day is not required while the full exposure window is
strictly fixed, validated, and fingerprinted.

## 6. Permanent regression coverage

Permanent tests cover:

```text
incoherent target relationships
zero score thresholds
zero component weights
full-window mismatch
recent-window mismatch
all simultaneous blocking reasons
weighted-score mutation
current trajectory exclusion
current flight exclusion
logical-flight deduplication
canonical evidence identity
UTC distinct-day calculation
recent-window boundary
production policy window alignment
production fixture weighted-score consistency
```

## 7. Permanent audit gate

The source audit introduced by the current increment is:

```text
apps/api/tools/projectionroutefrequencyreviewaudit
```

Backend Continuous Integration executes:

```text
go run ./tools/projectionroutefrequencyreviewaudit -strict
```

The gate protects the production implementation, SQL evidence boundary, configuration,
fingerprint, regression tests, production fixture, Stage 9 markers, this authoritative
review record, Documentation Index registration, and workflow wiring.

The corresponding Backend Quality step is:

```text
Run projection route frequency review audit
```

The permanent-audit increment completed all four exact Backend Continuous Integration
jobs:

```text
PERMANENT_AUDIT_COMMIT=6f039b33c96cdb67370158b0eda5d0fc87593de5
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30548438062
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90890525039
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90890525126
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90890525150
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90890829745
```

## 8. Formal closure

All confirmed findings are corrected and permanently protected. Exact Continuous
Integration is recorded for evidence isolation, policy and decision integrity, and the
permanent audit. No confirmed production-code defect is deferred, unclassified, or
left open.

The commit that introduces this formal-closure record must itself complete the same
four Backend Continuous Integration jobs before the external
`FORMAL_MODULE_CLOSURE=PASS` verdict is declared. That post-commit verification does
not require another source or documentation change when all four jobs succeed.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_ROUTE_FREQUENCY_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_ROUTE_FREQUENCY_ENGINEERING_DEBT=CLOSED
PROJECTION_ROUTE_FREQUENCY_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_ROUTE_FREQUENCY_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

The following records retrospectively normalize the accepted review findings into the
repository-wide nineteen-field standard. Severity values are retrospective engineering
classifications. Exact historical CI evidence follows the implementation waves and run
identifiers already recorded above.

### GFA-DATA-368 — Current target trajectory and flight leaked into their own historical Route Frequency evidence

1. **Finding / symptom.** Route-history evidence could include the trajectory or logical flight currently being predicted.
2. **Root cause.** Historical route queries did not exclude both current `trajectory_id` and current `flight_id` before aggregation.
3. **Failure scenario.** The current flight's already observed route result is counted as historical support for predicting that same route.
4. **Impact.** Route frequency, recent support, distinct-day support, and freshness can be self-inflated.
5. **Severity rationale.** P1 retrospective because target leakage directly fabricates predictive support.
6. **Existing guarantees violated.** Historical evidence isolation and no-target-leakage semantics.
7. **Considered solutions.** Exclude only trajectory ID; exclude only flight ID; exclude both identities before all aggregate calculations.
8. **Chosen remediation.** Filter the current trajectory and current logical flight at the SQL/read evidence boundary before counts are derived.
9. **Why selected.** Either identity alone is insufficient when one flight can own multiple trajectories.
10. **Rejected alternatives.** Downstream score correction still leaves contaminated counts and provenance.
11. **Trade-offs.** Historical support can decrease for flights whose current evidence had previously been counted.
12. **Regression tests / protection.** Tests cover current trajectory and current flight exclusion.
13. **Adversarial review findings.** Exclusion happens before observation, recency, day, and latest-observation calculations, not only before final scoring.
14. **Remediation iterations.** Target isolation and logical-flight evidence ownership were corrected in the evidence-isolation wave.
15. **Residual risks / limitations.** Correctness depends on current trajectory/flight identifiers being present and truthful at the read boundary.
16. **Operational/deployment consequences.** No migration; some route-frequency scores become lower and more conservative.
17. **Exact evidence.** `c6fff15f8d0c770197db40a69d54f8856044d8d2`, run `30534243693`; PostgreSQL `90843560950`, quality `90843560991`, race `90843560999`, container `90843853362`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectionroutefrequencyreviewaudit` and target-exclusion regressions protect the historical boundary.

### GFA-DATA-369 — Multiple trajectories from one logical flight inflated Route Frequency support

1. **Finding / symptom.** Several trajectories belonging to the same flight could be counted as independent route-history observations.
2. **Root cause.** Historical evidence identity was trajectory-centric rather than logical-flight-centric when `flight_id` existed.
3. **Failure scenario.** Segmentation or trajectory reconstruction yields three trajectories for one flight, and all three increase route-frequency support.
4. **Impact.** Routes with fragmented data appear more frequent than routes represented by one trajectory per flight.
5. **Severity rationale.** P1 retrospective because evidence duplication directly biases the predictive support denominator/numerator.
6. **Existing guarantees violated.** One logical flight equals one historical route observation.
7. **Considered solutions.** Count every trajectory; deduplicate only identical trajectory IDs; define canonical evidence identity as flight ID with trajectory fallback.
8. **Chosen remediation.** Use `flight_id` when available, otherwise `trajectory_id`, and retain the latest route result per evidence identity.
9. **Why selected.** It preserves flights without a flight ID while preventing multi-trajectory inflation when ownership is known.
10. **Rejected alternatives.** Trajectory-only dedup does not solve logical duplication.
11. **Trade-offs.** Historical observation count can decrease relative to raw trajectory count.
12. **Regression tests / protection.** Logical-flight deduplication and canonical evidence identity tests are permanent.
13. **Adversarial review findings.** The latest result is selected before aggregation so one evidence identity cannot contribute multiple route results.
14. **Remediation iterations.** SQL evidence isolation introduced canonical logical-flight identity.
15. **Residual risks / limitations.** Flights lacking `flight_id` necessarily fall back to trajectory identity.
16. **Operational/deployment consequences.** No migration; historical counts become semantically stricter.
17. **Exact evidence.** `c6fff15f8d0c770197db40a69d54f8856044d8d2`, exact run `30534243693`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** PostgreSQL integration and review audit require canonical evidence identity.

### GFA-DATA-370 — Route Frequency support score and minimum guard were owned by raw observations instead of distinct flights

1. **Finding / symptom.** Even after richer evidence metadata existed, scoring/guard semantics could remain tied to raw observation or trajectory count.
2. **Root cause.** Aggregate field ownership was not aligned with the logical-flight contract.
3. **Failure scenario.** `ObservationCount` is high because of repeated trajectory records while `DistinctFlightCount` is low, yet support is treated as sufficient.
4. **Impact.** The evaluator can authorize weak logical support through duplicate technical observations.
5. **Severity rationale.** P1 retrospective because support count is a direct usable/blocking input.
6. **Existing guarantees violated.** Logical-flight evidence ownership and conservative support semantics.
7. **Considered solutions.** Keep raw count for scoring; use minimum of both counts; make distinct-flight count authoritative and preserve raw count only as metadata.
8. **Chosen remediation.** Use `DistinctFlightCount` for observation-support scoring and the minimum historical-support guard.
9. **Why selected.** It matches the semantic unit established by evidence deduplication.
10. **Rejected alternatives.** Raw count remains susceptible to fragmentation; mixed ownership makes policy hard to reconstruct.
11. **Trade-offs.** Raw observation metadata may exceed the score's authoritative support count.
12. **Regression tests / protection.** Distinct-flight support and logical-flight dedup tests cover the ownership rule.
13. **Adversarial review findings.** `ObservationCount` is deliberately retained as evidence metadata, not silently removed.
14. **Remediation iterations.** Evaluator ownership was aligned after SQL deduplication established the canonical evidence unit.
15. **Residual risks / limitations.** Distinct flight identity quality depends on upstream `flight_id` assignment.
16. **Operational/deployment consequences.** No migration; support can become more conservative.
17. **Exact evidence.** Evidence isolation `c6fff15f8d0c770197db40a69d54f8856044d8d2`; policy/decision integrity `ee7c79bc8213dc030ce0d98f13d1065c9bb96275`, run `30544636679`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Result validation and audit require distinct-flight-owned support semantics.

### GFA-DATA-371 — Route Frequency did not bind one exact full and recent exposure-window policy end to end

1. **Finding / symptom.** Data-source history windows and evaluator scoring windows could differ without a contract failure.
2. **Root cause.** Full and recent exposure windows were implicit or independently configured across read and evaluation layers.
3. **Failure scenario.** The reader supplies 90 days while the evaluator assumes 180 days, causing counts to be interpreted against the wrong exposure period.
4. **Impact.** Count-based support, recent support, freshness, and route-frequency score lose comparable meaning.
5. **Severity rationale.** P1 retrospective because the same counts acquire different semantics under different exposure windows.
6. **Existing guarantees violated.** Exposure-denominator integrity and read/evaluator policy alignment.
7. **Considered solutions.** Convert all counts to rates; trust production defaults; publish explicit History/Recent windows and require exact boundary equality.
8. **Chosen remediation.** Add `HistoryWindow`/`RecentWindow`, exact derived boundaries, production 180d/30d policy, and exact Projection Read/evaluator window equality.
9. **Why selected.** Fixed, fingerprinted exposure windows make count-based evidence interpretable without inventing a new rate model.
10. **Rejected alternatives.** Rate-per-day conversion was not required while exposure is strictly fixed; trusting defaults leaves custom composition unsafe.
11. **Trade-offs.** Read and evaluator configs are more tightly coupled by an explicit domain contract.
12. **Regression tests / protection.** Full-window mismatch, recent-window mismatch, recent boundary, and production policy alignment tests are permanent.
13. **Adversarial review findings.** Count-based support was deliberately retained because the exposure window is now exact, validated, and fingerprinted.
14. **Remediation iterations.** Window evidence was added in the policy/decision-integrity wave after target leakage was closed.
15. **Residual risks / limitations.** Fixed windows are policy choices, not proof of optimal statistical exposure.
16. **Operational/deployment consequences.** No migration; mismatched custom reader/evaluator configuration fails closed.
17. **Exact evidence.** `ee7c79bc8213dc030ce0d98f13d1065c9bb96275`, run `30544636679`; quality `90877579007`, PostgreSQL `90877578926`, race `90877578928`, container `90877915808`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Projection Read contract checks and review audit require exact window alignment.

### GFA-CONTRACT-372 — Route Frequency configuration admitted incoherent targets, zero thresholds, or disabled component weights

1. **Finding / symptom.** Policy values could be internally unreachable or allow zero-information scoring/authorization.
2. **Root cause.** Cross-field target relationships, positive thresholds, window ordering, and strictly positive weight rules were incomplete.
3. **Failure scenario.** Target recent/day count exceeds total target, recent window exceeds history, a usable threshold is zero, or a component has zero/non-finite weight.
4. **Impact.** Result status and score semantics become impossible, overly permissive, or silently component-disabled.
5. **Severity rationale.** P1 retrospective because invalid policy can directly change route-frequency authorization.
6. **Existing guarantees violated.** Constructor-time policy coherence and fixed-component score semantics.
7. **Considered solutions.** Clamp relationships; allow zero to disable policy; reject all incoherent/non-positive/non-finite values.
8. **Chosen remediation.** Validate target ordering, window ordering, positive thresholds, positive finite weights, and unit total.
9. **Why selected.** Invalid policy is explicit rather than silently rewritten.
10. **Rejected alternatives.** Clamping hides configuration defects; zero weights conflict with the fixed five-component contract.
11. **Trade-offs.** Some previously accepted configurations fail fast.
12. **Regression tests / protection.** Incoherent targets, zero thresholds, zero weights, and window mismatch tests are permanent.
13. **Adversarial review findings.** Mandatory basis-point conversion and dynamic component registries were rejected as unrelated mechanical changes.
14. **Remediation iterations.** Config and policy snapshots were hardened in the policy/decision wave.
15. **Residual risks / limitations.** Threshold choices remain engineering policy, not calibrated probabilities.
16. **Operational/deployment consequences.** No migration; invalid configuration now fails before evaluation.
17. **Exact evidence.** `ee7c79bc8213dc030ce0d98f13d1065c9bb96275`, exact run `30544636679`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Constructor validation, Result validation, and strict audit protect policy coherence.

### GFA-DATA-373 — Route Frequency reported only the first blocking reason instead of complete simultaneous denial evidence

1. **Finding / symptom.** A blocked Route Frequency result could expose only one failed guard while several were simultaneously true.
2. **Root cause.** Decision logic used first-match/early-return semantics instead of complete violation accumulation.
3. **Failure scenario.** Route confidence is low, support is insufficient, and history is stale, but only one limitation is published.
4. **Impact.** Denial provenance is incomplete and operational diagnosis is misleading.
5. **Severity rationale.** P2 retrospective because the result remains conservative but its explanation and audit evidence are incomplete.
6. **Existing guarantees violated.** Complete limitations and independently explainable decisions.
7. **Considered solutions.** Preserve first reason; generic blocked reason; accumulate all hard violations and normalize them.
8. **Chosen remediation.** Evaluate all blockers, deduplicate/sort limitations, and derive blocked state from the complete set.
9. **Why selected.** One result explains every evidence deficit without weakening policy.
10. **Rejected alternatives.** First-only and generic messages cannot support precise remediation or audit.
11. **Trade-offs.** Blocked results may contain more limitation records.
12. **Regression tests / protection.** All-simultaneous-blocking-reasons regression is permanent.
13. **Adversarial review findings.** Limitation codes, not prose ordering, are treated as stable semantics.
14. **Remediation iterations.** Decision reporting was rebuilt in the policy/decision-integrity wave.
15. **Residual risks / limitations.** Human-readable messages can evolve while codes remain stable.
16. **Operational/deployment consequences.** No migration; richer blocked-result diagnostics.
17. **Exact evidence.** `ee7c79bc8213dc030ce0d98f13d1065c9bb96275`, run `30544636679`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Result validation and regression tests require the complete limitation set.

### GFA-DATA-374 — Route Frequency validation did not independently reconstruct the weighted score

1. **Finding / symptom.** A published aggregate score could disagree with the component scores and weights while still appearing structurally valid.
2. **Root cause.** Result validation did not recompute the weighted sum as an independent integrity check.
3. **Failure scenario.** A stale fixture publishes `0.85` while its actual components evaluate to `0.94`, or a custom producer mutates score without changing components.
4. **Impact.** Downstream decisions can trust a mathematically inconsistent score.
5. **Severity rationale.** P1 retrospective because score thresholds directly determine usable/blocked Route Frequency evidence.
6. **Existing guarantees violated.** Mathematical integrity and consumer-side result validation.
7. **Considered solutions.** Trust producer score; check only [0,1] range; recompute weighted score using canonical component weights and tolerance.
8. **Chosen remediation.** `Result.Validate()` independently reconstructs `sum(score × weight)` and requires strictly positive unit-sum weights.
9. **Why selected.** Coordinated result integrity is checked at the consumer boundary rather than assumed from construction.
10. **Rejected alternatives.** Range checks cannot detect mathematically forged or stale scores.
11. **Trade-offs.** Validation intentionally duplicates deterministic scoring arithmetic.
12. **Regression tests / protection.** Weighted-score mutation and production fixture weighted-score consistency tests are permanent.
13. **Adversarial review findings.** The stale 0.85 fixture is closure evidence for this finding, not a separate production defect ID.
14. **Remediation iterations.** Score reconstruction exposed the stale fixture, which was corrected to the actual 0.94 value.
15. **Residual risks / limitations.** Validation depends on the versioned component catalog remaining authoritative.
16. **Operational/deployment consequences.** No migration; malformed custom results and stale fixtures fail validation.
17. **Exact evidence.** `ee7c79bc8213dc030ce0d98f13d1065c9bb96275`, exact run `30544636679`; permanent audit `6f039b33c96cdb67370158b0eda5d0fc87593de5`, run `30548438062`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Weighted-score mutation regression and permanent audit protect the arithmetic contract.

### GFA-DATA-375 — Route Frequency fingerprint omitted decision-relevant route, history, exposure, or policy evidence

1. **Finding / symptom.** Distinct Route Frequency computations could share identity if route state, history evidence, window boundaries, support counts, thresholds, or weights changed outside the fingerprint.
2. **Root cause.** Fingerprint scope did not fully mirror the evaluator's decision surface.
3. **Failure scenario.** Route provenance or exposure window changes but downstream lineage sees the same input identity.
4. **Impact.** Provenance collisions can hide materially different historical support and policy decisions.
5. **Severity rationale.** P1 retrospective because Route Frequency fingerprint is part of the projection authorization chain.
6. **Existing guarantees violated.** Semantic identity, reproducibility, and change-sensitive provenance.
7. **Considered solutions.** Hash only canonical route key; trust route-history fingerprint alone; bind every normalized decision-relevant upstream and policy field.
8. **Chosen remediation.** Fingerprint contract/version, Route Intelligence schema/status/identity/confidence/provenance, trajectory/flight identity, route-history fingerprint, exposure windows, support/latest evidence, thresholds, and weights.
9. **Why selected.** Identity now tracks every published input capable of changing the decision.
10. **Rejected alternatives.** Route key alone misses evidence/policy changes; upstream hash alone cannot bind local policy.
11. **Trade-offs.** Fingerprint input is larger and version evolution requires explicit maintenance.
12. **Regression tests / protection.** Semantic mutation tests and permanent review audit protect decision-input coverage.
13. **Adversarial review findings.** `GeneratedAt`-style output metadata is not used as input identity unless it changes decision evidence.
14. **Remediation iterations.** Fingerprinting was expanded with explicit windows and policy snapshots in the policy/decision wave.
15. **Residual risks / limitations.** Newly introduced decision fields require corresponding fingerprint updates.
16. **Operational/deployment consequences.** Fingerprint identity changes; no migration required by this review.
17. **Exact evidence.** `ee7c79bc8213dc030ce0d98f13d1065c9bb96275`, run `30544636679`; permanent audit `6f039b33c96cdb67370158b0eda5d0fc87593de5`, run `30548438062`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Strict `projectionroutefrequencyreviewaudit` protects semantic fingerprint coverage.
