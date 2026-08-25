# Projection Neighbors Review Hardening

Status: closed

```text
AUTHORITATIVE_BASELINE_COMMIT=e13a117f969e2922d09a7804fe50005d01bc2ecf
CANDIDATE_INTEGRITY_COMMIT=e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff
CONTINUATION_INTEGRITY_COMMIT=911a1b102c68af2746a13bfca48b008cf7225ff8
ROUTE_SCOPE_INTEGRITY_COMMIT=3eee05fb44484aa6e389af66520aba23d4ae277e
SELECTOR_PIPELINE_COMMIT=353d19bc97f561e1897ece1967e7304c0e10b5fb
PERMANENT_AUDIT_COMMIT=c409cc171507050625524af1a0b8b8a6f38b7a75
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30452465009
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90577613283
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90577613277
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90577613384
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90577905997
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_NEIGHBORS_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_NEIGHBORS_ENGINEERING_DEBT=CLOSED
PROJECTION_NEIGHBORS_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_NEIGHBORS_REVIEW_STATUS=CLOSED
```

## 1. Scope

This record covers the dedicated review and hardening of:

```text
apps/api/internal/projectionintelligence/projectionneighbors
```

It also covers the route-scope evidence path through:

```text
projectionread
projectionproduction
projectioncontinuation
```

The review is limited to the deterministic historical-neighbor selection contract.
It does not claim empirical prediction accuracy or operational aviation suitability.

## 2. Findings closed

### 2.1 Candidate integrity before expensive evaluation

The selector now performs source-independent eligibility and duplicate detection
before applying the expensive similarity budget. Duplicate identifiers are counted
across the whole input before truncation, so duplicate evidence cannot bypass the
guard by appearing outside the evaluation budget.

Eligible candidates are ordered by newest historical end time and then stable
trajectory identifier before `MaximumCandidateCount` is applied.

### 2.2 Canonical point and fingerprint ordering

Snapshot construction and fingerprint generation share canonical point ordering.
Equal timestamps are ordered by point identifier. Candidate fingerprints are
canonicalized independently of input order.

The selection fingerprint covers:

```text
selection contract version
similarity policy identity
route-scope fingerprint
as-of time
required continuation duration
maximum continuation gap
candidate and selection limits
similarity, distance and age policies
current trajectory snapshot
candidate trajectory snapshots
```

### 2.3 Similarity failure classification

Candidate-local non-comparability is represented as a deterministic rejection.
Systemic similarity-engine failure and malformed similarity evidence are returned
as typed errors and are not hidden as ordinary candidate rejection.

### 2.4 Continuous continuation evidence

The selector publishes structured `AnchorEvidence` and uses segmented linear
continuation search. A candidate continuation cannot cross an observation gap
larger than the configured maximum.

Unavailable duration and discontinuous continuation are distinct rejection cases.

### 2.5 Source-attested route scope

Historical candidates require explicit route-scope evidence. The PostgreSQL read
path constructs a uniform route attestation only after route-filtered candidate
loading. The read snapshot transports a defensive clone of that evidence.

Production composition validates the candidate route scope against the current
complete Route Intelligence result. Projection Continuation receives and forwards
the same route scope to its internal neighbor selector.

Cross-route candidates are rejected before anchor or similarity evaluation.

### 2.6 Selector pipeline decomposition

`Selector.Select` is a short coordinator for:

```text
prepareSelectionContext
evaluateCandidatePool
assembleSelectionResult
```

Request normalization, candidate preparation, expensive evaluation, deterministic
ranking, result assembly, limitation generation and result validation are separated
into focused files.

### 2.7 Explicit limiting semantics

Two independent conditions are now published:

```text
CandidateEvaluationTruncated
QualifiedSelectionLimited
```

`CandidateEvaluationTruncated` means the expensive evaluation budget prevented all
eligible candidates from being checked.

`QualifiedSelectionLimited` means more candidates qualified than could be returned
under `SelectionLimit`.

The deprecated `Truncated` field remains a compatibility alias for
`CandidateEvaluationTruncated` and is cross-field validated.

## 3. Deliberately retained contracts

The following were reviewed and deliberately retained:

```text
exact float64 comparison for deterministic ranking tie-breakers
idiomatic New(Config) (*Selector, error) constructor
public compatibility alias Result.Truncated
producer-owned similarity implementation behind the consumer-facing selector contract
```

No product requirement justified arbitrary coordinate or score quantization.

## 4. Permanent regression coverage

Permanent tests cover:

```text
eligibility before expensive budget
whole-input duplicate detection
canonical equal-timestamp ordering
systemic similarity failure propagation
similarity evidence validation
continuous continuation gaps
large linear anchor search
source-attested route scope
cross-route rejection before similarity
route-scope fingerprint identity
read, production and continuation propagation
candidate-evaluation truncation
qualified-selection limiting
cross-field result validation
```

## 5. Permanent audit gate

The source audit is:

```text
apps/api/tools/projectionneighborsreviewaudit
```

Backend Continuous Integration executes:

```text
go run ./tools/projectionneighborsreviewaudit -strict
```

The gate protects the hardened contracts, tests, documentation and workflow wiring.

## 6. Closure evidence

The permanent audit commit completed every Backend Continuous Integration job:

```text
PERMANENT_AUDIT_COMMIT=c409cc171507050625524af1a0b8b8a6f38b7a75
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30452465009
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90577613283
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90577613277
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90577613384
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90577905997
```

Engineering implementation and formal closure documentation are complete. No
confirmed finding remains open, unclassified or deferred. The permanent audit gate
remains mandatory in Backend Continuous Integration.

## Canonical remediation history

The following records retrospectively normalize the accepted review findings into the
repository-wide nineteen-field standard. Severity values are retrospective engineering
classifications. Intermediate implementation commits are cited only where the source
record names them; the only exact historical CI run claimed here is the permanent-audit
run explicitly recorded above.

### GFA-DATA-342 — Candidate eligibility and duplicate checks occurred after the expensive evaluation budget

1. **Finding / symptom.** Source-independent eligibility and duplicate detection could be applied too late, after the finite similarity-evaluation budget had already shaped the candidate pool.
2. **Root cause.** Cheap integrity filtering and expensive similarity evaluation were not separated into explicit pipeline phases.
3. **Failure scenario.** Invalid or duplicated candidates consume the budget while a valid historical candidate outside the truncated prefix is never evaluated.
4. **Impact.** Neighbor selection can become incomplete or biased by malformed input ordering rather than domain evidence.
5. **Severity rationale.** P1 retrospective because it can change which historical evidence authorizes a projection.
6. **Existing guarantees violated.** Deterministic candidate eligibility, bounded-compute fairness, evidence integrity.
7. **Considered solutions.** Increase the budget; reject only duplicates encountered inside the budget; pre-classify the entire input before expensive work.
8. **Chosen remediation.** Perform source-independent eligibility and whole-input duplicate detection before `MaximumCandidateCount` is applied.
9. **Why selected.** It preserves the compute bound without allowing invalid input to consume scarce evaluation slots.
10. **Rejected alternatives.** A larger budget only moves the failure threshold; inside-budget duplicate checks still allow duplicates outside the prefix to bypass the guard.
11. **Trade-offs.** The selector performs a full cheap preparation pass before similarity evaluation.
12. **Regression tests / protection.** Tests cover eligibility-before-budget and duplicate identifiers appearing outside the evaluation prefix.
13. **Adversarial review findings.** Duplicate detection was intentionally defined over the whole request, not merely the evaluated subset.
14. **Remediation iterations.** Candidate preparation was extracted and made authoritative before expensive evaluation.
15. **Residual risks / limitations.** Correctness still depends on truthful candidate identifiers and upstream trajectory content.
16. **Operational/deployment consequences.** No migration; some requests reject more candidates earlier and spend fewer similarity calls.
17. **Exact evidence.** `e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff`; permanent audit `c409cc171507050625524af1a0b8b8a6f38b7a75`, run `30452465009`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectionneighborsreviewaudit` and regression tests require pre-budget eligibility and whole-input duplicate detection.

### GFA-DATA-343 — Candidate-budget truncation was input-order sensitive rather than recency deterministic

1. **Finding / symptom.** Applying the candidate budget before a canonical domain order could make the evaluated subset depend on caller input ordering.
2. **Root cause.** No explicit newest-evidence-first ordering contract preceded truncation.
3. **Failure scenario.** The same candidate set in a different request order evaluates a different subset under the same `MaximumCandidateCount`.
4. **Impact.** Equivalent evidence sets can produce different selected neighbors and fingerprints.
5. **Severity rationale.** P2 retrospective because the defect is deterministic-correctness loss under bounded evaluation rather than unrestricted data corruption.
6. **Existing guarantees violated.** Order invariance and reproducible historical selection.
7. **Considered solutions.** Preserve caller order; sort by identifier only; sort by recency with a stable identifier tie-break.
8. **Chosen remediation.** Order eligible candidates by newest historical end time, then trajectory identifier, before truncation.
9. **Why selected.** Recency is a meaningful historical evidence priority and the identifier provides deterministic ties.
10. **Rejected alternatives.** Caller order is not semantic; identifier-only ordering ignores evidence freshness.
11. **Trade-offs.** The budget intentionally favors newer qualifying history.
12. **Regression tests / protection.** Candidate-order and budget tests verify stable selection under reordered inputs.
13. **Adversarial review findings.** Ranking after similarity remains separate from pre-budget recency ordering.
14. **Remediation iterations.** Ordering moved into candidate preparation before the expensive phase.
15. **Residual risks / limitations.** Recency priority is policy, not a claim that newer history is always more predictive.
16. **Operational/deployment consequences.** No migration; bounded requests may evaluate a different, now deterministic subset.
17. **Exact evidence.** `e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff`; permanent audit run `30452465009`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent tests retain newest-first plus identifier tie-break behavior.

### GFA-CONTRACT-344 — Selection limit could exceed the maximum candidate evaluation budget

1. **Finding / symptom.** Configuration could request more selected neighbors than the evaluator was permitted to inspect.
2. **Root cause.** `SelectionLimit` and `MaximumCandidateCount` were validated independently without a cross-field constraint.
3. **Failure scenario.** A nominally valid configuration makes `StatusComplete` mathematically unreachable because fewer candidates may be checked than must be selected.
4. **Impact.** Result status and completeness semantics become incoherent by construction.
5. **Severity rationale.** P2 retrospective because it creates impossible policy states but does not itself fabricate observations.
6. **Existing guarantees violated.** Constructor-time policy validity and reachable result semantics.
7. **Considered solutions.** Clamp the selection limit; raise the candidate budget silently; reject incoherent configuration.
8. **Chosen remediation.** Reject configuration when `SelectionLimit > MaximumCandidateCount`.
9. **Why selected.** Invalid policy is surfaced explicitly rather than silently rewritten.
10. **Rejected alternatives.** Clamping or auto-expansion hides caller mistakes and changes requested resource policy.
11. **Trade-offs.** Some previously accepted configurations now fail construction.
12. **Regression tests / protection.** Config validation tests cover the cross-field relationship.
13. **Adversarial review findings.** The relationship is treated as a contract invariant, not merely a tuning recommendation.
14. **Remediation iterations.** A dedicated typed configuration error was added during candidate-integrity hardening.
15. **Residual risks / limitations.** Other target relationships still require their own explicit validation when introduced.
16. **Operational/deployment consequences.** No migration; invalid deployments fail fast at construction/configuration time.
17. **Exact evidence.** `e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff`; permanent audit run `30452465009`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Constructor validation and permanent audit protect the budget/selection relationship.

### GFA-DATA-345 — Snapshot and selection fingerprints were not fully canonical under equivalent input ordering

1. **Finding / symptom.** Equal-timestamp points and same-ID candidate snapshots could retain caller-order sensitivity in snapshot/fingerprint construction.
2. **Root cause.** Snapshot ordering and fingerprint ordering did not share one complete canonical comparator and candidate sort key.
3. **Failure scenario.** Equivalent trajectories or candidate lists arrive in a different in-memory order and produce a different input fingerprint.
4. **Impact.** Cache identity, provenance comparison, and replay reproducibility can diverge without a semantic evidence change.
5. **Severity rationale.** P1 retrospective because fingerprints are evidence identities used across Projection authorization and provenance.
6. **Existing guarantees violated.** Semantic fingerprint stability and input-order invariance.
7. **Considered solutions.** Preserve stable sort input order; order only by timestamp/ID; use one canonical point comparator plus content-derived candidate sort key.
8. **Chosen remediation.** Share canonical point ordering, bit-stable numeric tie-breaks, and canonical candidate fingerprint sort keys across snapshots and selection hashing.
9. **Why selected.** It binds identity to evidence content rather than incidental slice order.
10. **Rejected alternatives.** Stable-sort preservation still depends on caller order when keys collide.
11. **Trade-offs.** Fingerprint construction performs additional deterministic sorting/hashing work.
12. **Regression tests / protection.** Equal-timestamp and candidate-order invariance tests protect canonical fingerprints.
13. **Adversarial review findings.** Exact `float64` ranking comparisons were retained; arbitrary quantization was rejected as unrelated.
14. **Remediation iterations.** A shared `canonicalPointLess` and trajectory fingerprint sort key were introduced during candidate hardening.
15. **Residual risks / limitations.** Fingerprint stability assumes deterministic serialization of every newly added evidence field.
16. **Operational/deployment consequences.** Fingerprint version advanced; no database migration.
17. **Exact evidence.** `e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff`; permanent audit run `30452465009`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit checks canonical equal-timestamp ordering and fingerprint invariance coverage.

### GFA-OPS-346 — Systemic similarity failures were indistinguishable from ordinary candidate non-comparability

1. **Finding / symptom.** Similarity-engine or malformed-evidence failures could be treated like a normal candidate rejection.
2. **Root cause.** Candidate-local domain non-comparability and subsystem failure did not have separate typed error paths.
3. **Failure scenario.** The similarity subsystem fails or returns invalid evidence and the selector merely reports fewer qualifying neighbors.
4. **Impact.** Operational failures can be silently converted into domain conclusions, masking degraded computation.
5. **Severity rationale.** P1 retrospective because a failed dependency can otherwise masquerade as valid negative evidence.
6. **Existing guarantees violated.** Fail-closed dependency handling, observability, and honest result semantics.
7. **Considered solutions.** Reject every similarity error as candidate-local; fail the whole request on every error; classify candidate-local versus systemic/malformed failure.
8. **Chosen remediation.** Keep deterministic non-comparability as a rejection while returning typed errors for systemic engine failure and invalid similarity evidence.
9. **Why selected.** It preserves useful partial evaluation without hiding infrastructure or contract failures.
10. **Rejected alternatives.** Treating every error identically either hides faults or makes one incomparable candidate abort valid selection.
11. **Trade-offs.** The selector and tests maintain a richer error taxonomy.
12. **Regression tests / protection.** Systemic failure propagation and malformed similarity evidence validation are permanent regressions.
13. **Adversarial review findings.** Candidate-local inability to compare remains a domain rejection by design.
14. **Remediation iterations.** Error classification was consolidated into the candidate evaluation phase.
15. **Residual risks / limitations.** Upstream similarity implementations must return errors that can be classified truthfully.
16. **Operational/deployment consequences.** No migration; some former unavailable/limited results become explicit request errors.
17. **Exact evidence.** Candidate-integrity and selector-pipeline waves `e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff`, `353d19bc97f561e1897ece1967e7304c0e10b5fb`; permanent audit run `30452465009`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent tests require systemic failure propagation and similarity-evidence validation.

### GFA-DATA-347 — Neighbor continuation could bridge unobserved temporal gaps

1. **Finding / symptom.** Historical continuation sufficiency could be inferred across a large observation gap rather than continuous observed evidence.
2. **Root cause.** Anchor search considered duration availability without a maximum allowed gap between continuation observations.
3. **Failure scenario.** A candidate has points before and after a long outage and is treated as providing a continuous future continuation.
4. **Impact.** Projection can learn from synthetic continuity unsupported by observations.
5. **Severity rationale.** P1 retrospective because discontinuous historical evidence can directly authorize or shape a future path.
6. **Existing guarantees violated.** Observed-evidence continuity and historical neighbor integrity.
7. **Considered solutions.** Ignore gaps; reject any candidate containing any gap; segment candidate evidence by a configured maximum observation gap.
8. **Chosen remediation.** Add a maximum continuation gap policy, segmented linear anchor search, and structured `AnchorEvidence` with distinct unavailable/discontinuous outcomes.
9. **Why selected.** It rejects only the unsupported continuation while retaining valid continuous segments.
10. **Rejected alternatives.** Global rejection is unnecessarily destructive; ignoring gaps manufactures evidence.
11. **Trade-offs.** The policy requires an explicit/default gap threshold and additional scan state.
12. **Regression tests / protection.** Discontinuous continuation, continuous evidence, large trajectory search, and fingerprinted gap policy are covered.
13. **Adversarial review findings.** Duration-unavailable and duration-present-but-discontinuous cases remain separate rejection reasons.
14. **Remediation iterations.** Anchor search was replaced with segmented continuation-aware logic.
15. **Residual risks / limitations.** The configured gap threshold is an engineering policy, not empirical proof of path continuity.
16. **Operational/deployment consequences.** No migration; some previously qualifying candidates are rejected as discontinuous.
17. **Exact evidence.** `911a1b102c68af2746a13bfca48b008cf7225ff8`; permanent audit `c409cc171507050625524af1a0b8b8a6f38b7a75`, run `30452465009`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Review audit requires continuation-gap policy, structured evidence, and discontinuity regressions.

### GFA-DATA-348 — Historical neighbors lacked source-attested route-scope ownership

1. **Finding / symptom.** Candidate trajectories could reach neighbor evaluation without evidence proving they belonged to the current origin-destination route scope.
2. **Root cause.** Route filtering existed upstream but route identity was not transported and validated as a first-class selection input.
3. **Failure scenario.** A stale, custom, or incorrectly composed candidate list injects a cross-route trajectory that is geometrically similar near the current endpoint.
4. **Impact.** A projection can be supported by history from the wrong route while appearing internally valid.
5. **Severity rationale.** P1 retrospective because route identity is a primary semantic boundary for historical projection evidence.
6. **Existing guarantees violated.** Provenance ownership, route-scope isolation, and cross-module contract integrity.
7. **Considered solutions.** Trust SQL filtering implicitly; infer route from trajectory geometry; publish and validate explicit route-scope evidence end to end.
8. **Chosen remediation.** Introduce fingerprinted `RouteScope`/candidate route evidence, propagate it through read/production/continuation, and reject cross-route candidates before similarity evaluation.
9. **Why selected.** The consuming selector can independently verify the evidence boundary instead of trusting transport history.
10. **Rejected alternatives.** Implicit query trust is lost after data crosses package boundaries; geometry is not authoritative route identity.
11. **Trade-offs.** Requests carry additional route provenance and fingerprints.
12. **Regression tests / protection.** Route-scope identity, propagation, missing/mismatched evidence, and pre-similarity cross-route rejection are tested.
13. **Adversarial review findings.** Uniform route scope is constructed only after route-filtered PostgreSQL loading; explicit scope remains available for independently attested candidates.
14. **Remediation iterations.** Route scope was added to neighbors and then threaded through Projection Read, Production, and Continuation.
15. **Residual risks / limitations.** Correctness depends on the upstream route attestation itself being derived from truthful Route Intelligence evidence.
16. **Operational/deployment consequences.** No migration; custom integrations must provide valid route-scope evidence.
17. **Exact evidence.** `3eee05fb44484aa6e389af66520aba23d4ae277e`; permanent audit run `30452465009`, PostgreSQL job `90577613384`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit checks route-scope fingerprints and propagation across read, production, and continuation.

### GFA-MAINT-349 — Neighbor selection mixed preparation, evaluation, ranking, assembly, and validation in one coordinator

1. **Finding / symptom.** The selection path concentrated unrelated responsibilities, making integrity changes difficult to reason about and regression-prone.
2. **Root cause.** Candidate preparation, expensive evaluation, ranking, limitations, result construction, and validation evolved inside one broad control flow.
3. **Failure scenario.** A future budget, error-classification, or result-semantic change accidentally changes ordering or validation in another stage.
4. **Impact.** Maintenance risk around a high-integrity projection boundary increases even when current output is correct.
5. **Severity rationale.** P3 retrospective because this is a maintainability/prevention defect, not an independent evidence-corruption bug.
6. **Existing guarantees violated.** Reviewability and separation of correctness-critical stages.
7. **Considered solutions.** Keep the monolith with comments; introduce an abstract framework; split only concrete pipeline stages.
8. **Chosen remediation.** Reduce `Selector.Select` to orchestration over preparation, candidate evaluation, and result assembly helpers.
9. **Why selected.** It makes existing invariants locally inspectable without widening the public API.
10. **Rejected alternatives.** Comments do not enforce ownership; a generic framework would over-abstract one concrete pipeline.
11. **Trade-offs.** More package-private files and helper types must be navigated.
12. **Regression tests / protection.** Existing behavior tests plus permanent source audit cover the decomposed stages.
13. **Adversarial review findings.** Function length alone was not assigned severity; the finding is responsibility coupling at an integrity boundary.
14. **Remediation iterations.** The pipeline was split after candidate, continuation, and route-scope hardening established stable stage boundaries.
15. **Residual risks / limitations.** Decomposition does not remove the need for end-to-end result validation.
16. **Operational/deployment consequences.** Refactor only; no runtime policy or migration consequence intended.
17. **Exact evidence.** `353d19bc97f561e1897ece1967e7304c0e10b5fb`; permanent audit run `30452465009`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Review audit and focused package structure keep stage ownership explicit.

### GFA-CONTRACT-350 — One truncation flag conflated evaluation-budget truncation with qualified-result limiting

1. **Finding / symptom.** A single `Truncated` state could not distinguish unevaluated eligible candidates from evaluated candidates omitted only because `SelectionLimit` was full.
2. **Root cause.** Compute-budget truncation and output-cardinality limiting were modeled as one boolean concept.
3. **Failure scenario.** A consumer sees `Truncated=true` but cannot tell whether evidence was never evaluated or whether all candidates were evaluated and only the returned list was capped.
4. **Impact.** Completeness, limitations, and downstream confidence semantics can be interpreted incorrectly.
5. **Severity rationale.** P2 retrospective because this corrupts contract interpretation rather than the underlying observations.
6. **Existing guarantees violated.** Explicit limitation semantics and reconstructible result status.
7. **Considered solutions.** Keep one flag plus message text; derive semantics from counts downstream; publish two explicit states with cross-field validation.
8. **Chosen remediation.** Add `CandidateEvaluationTruncated` and `QualifiedSelectionLimited`; retain deprecated `Truncated` only as a validated alias of the first.
9. **Why selected.** The result states exactly which information was lost and remains backward compatible.
10. **Rejected alternatives.** Message parsing is not a contract; downstream count inference duplicates domain rules.
11. **Trade-offs.** Result schema gains fields and compatibility validation.
12. **Regression tests / protection.** Tests cover both limiting modes, status reconstruction, counts, and deprecated alias consistency.
13. **Adversarial review findings.** The compatibility alias was deliberately retained rather than mechanically removed.
14. **Remediation iterations.** Split semantics were introduced while decomposing result assembly.
15. **Residual risks / limitations.** Future limiting modes must receive distinct semantics rather than reuse these fields.
16. **Operational/deployment consequences.** No migration; downstream consumers can migrate from `Truncated` to the explicit fields.
17. **Exact evidence.** `353d19bc97f561e1897ece1967e7304c0e10b5fb`; permanent audit `c409cc171507050625524af1a0b8b8a6f38b7a75`, run `30452465009`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `Result.Validate()` and permanent audit reconstruct both limiting states from authoritative counts.
