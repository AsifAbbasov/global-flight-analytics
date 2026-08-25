# Projection Baseline Review Hardening

Status: closed

## Scope

This review hardens and classifies the findings for:

```text
apps/api/internal/projectionintelligence/projectionbaseline
apps/api/internal/projectionintelligence/projectionread
apps/api/internal/analytics/trajectoryeligibility
```

The purpose is to restore historical replay trust, make projection decisions
reproducible, reject invalid collaborator output, and preserve a conservative
research-only model boundary.

## Accepted findings

- Historical `as_of` snapshots filtered future points but retained aggregate quality
  calculated from segments that could include later evidence.
- Segments and coverage gaps were not consistently isolated by completed event time.
- The former future-point regression test did not change aggregate quality and could
  not detect aggregate leakage.
- PostgreSQL trajectory hydration filtered points without recomputing cutoff-safe
  trajectory quality.
- Unavailable results omitted the evidence required to reproduce denial decisions.
- Source identity, on-ground state, selected altitude reference, observation-age
  policy, physical policy constants, horizontal fallback policy, and eligibility
  policy identity were incomplete fingerprint inputs.
- Confidence did not account for age of the latest observation.
- Kinematics lacked conservative physical bounds.
- Altitude selection did not publish whether geometric or barometric altitude was
  used, while vertical-rate reference remained source-limited.
- Conflicting observations at an identical latest timestamp were selected by lexical
  ordering rather than rejected as ambiguous.
- Allowed on-ground input reused the airborne propagation model.
- The horizontal-only result branch was blocked by default altitude eligibility.
- Custom collaborator evaluations were insufficiently validated.
- Nil `Baseline` use returned an unrelated dependency-construction error.
- Permanent regression enforcement for Projection Baseline contracts was absent.

## Corrected contracts

- `buildCutoffSnapshot` creates a defensive cutoff view containing only observations,
  segments, and gaps whose event-time evidence is complete at or before `as_of`.
- Trajectory quality is recomputed from completed segments. Projection fails closed
  when completed quality evidence does not cover the latest included point.
- PostgreSQL hydration applies the same completed-segment and completed-gap boundary
  and recomputes trajectory quality before returning the trajectory.
- Unavailable and successful results publish deterministic input fingerprints,
  input references, latest observed time, and eligibility policy evidence.
- Projection Baseline advances to `short-horizon-kinematic-baseline-v3`.
- Baseline input fingerprinting advances to
  `projection-baseline-input-fingerprint-v4` and consumes the canonical horizon plan
  fingerprint, which already binds plan version, truncation evidence, and every
  forecast timestamp.
- Confidence combines cutoff-safe trajectory quality, latest-observation age, and
  forecast-horizon decay. Non-finite calculations are rejected instead of clamped.
- Conservative bounds protect ground speed, heading, vertical rate, altitude, and
  allowed on-ground motion.
- Altitude selection is typed and publishes geometric, barometric, or unavailable
  reference evidence. Vertical-rate provenance states the source limitation.
- Conflicting latest observations with the same timestamp return an unavailable
  result with `projection_latest_observation_ambiguous`.
- Allowed on-ground observations use a stationary conservative model and always
  produce a limited result.
- Horizontal-only fallback is an explicit policy and is allowed only when missing
  altitude is the sole eligibility denial reason.
- Projection eligibility output must contain exactly one projection decision, valid
  allowed/denied reason semantics, and known unique reason codes.
- Default eligibility policy name, version, configuration fingerprint, and policy
  inputs are included in provenance and baseline fingerprints. Custom evaluators get
  an explicit unversioned type-derived identity unless they publish a valid identity.
- Targeted, race, full-suite, vet, PostgreSQL, container, and Continuous Integration
  verification protect the engineering changes.

## Qualified or rejected findings

- Coverage-gap `CreatedAt` is not compared with `as_of`. `CreatedAt` represents
  processing or materialization time, while `StartTime` and `EndTime` represent the
  event interval that can influence historical analytics. Using processing time as
  event truth would incorrectly hide valid replay evidence. Completed `EndTime` is
  therefore the authoritative cutoff boundary.
- Horizon version, truncation status, truncation reason, and every forecast timestamp
  were already covered by `plan.Fingerprint`; duplicating those fields in the baseline
  fingerprint would create two competing horizon identities.
- Result status and limitations are outputs derived from the fingerprinted inputs and
  policies. They are not input fingerprint material.
- Source-file length alone is not a defect. Responsibilities are now separated into
  cutoff, confidence, kinematics, altitude, eligibility, latest-observation,
  explanations, fingerprint, and provenance units without forcing an arbitrary line
  limit.
- A public `SnapshotBuilder` interface was not required. The package-private cutoff
  builder provides the necessary deterministic boundary without creating an unused
  abstraction or widening the public API.
- Cross-package consolidation of confidence and hashing helpers is not forced by this
  module review. Projection strategies currently have distinct evidence semantics;
  premature shared abstractions could erase those differences.
- `New(Config) (*Baseline, error)` returning `nil` with an error on failed construction
  is idiomatic Go and is not a domain null-state defect.
- Test names containing `And` or `With` do not alter production behavior or violate a
  meaningful Go contract.
- Nullable altitude and vertical-uncertainty pointers are retained because absence is
  an intentional published contract state for horizontal-only projections.
- `float64` is retained for geodesy and confidence. There are no monetary calculations,
  and deterministic formulas and fingerprints provide reproducibility without an
  arbitrary decimal quantization policy.

## Permanent verification

`apps/api/tools/projectionbaselinereviewaudit` verifies:

- cutoff-safe points, segments, gaps, and quality recomputation;
- PostgreSQL cutoff alignment;
- complete input and policy fingerprint evidence;
- unavailable-result provenance;
- observation-age confidence decay;
- physical kinematic bounds;
- typed altitude reference and provenance;
- deterministic rejection of conflicting latest observations;
- validated eligibility collaborator output;
- explicit horizontal fallback policy;
- stationary on-ground behavior;
- regression test presence;
- Stage 9 completion evidence;
- this review record and the documentation index;
- Backend Continuous Integration enforcement.

## Engineering evidence

```text
AUTHORITATIVE_BASELINE_COMMIT=b4da27772fad838bf2a237ff9989621bfae6d5f2
CUTOFF_INTEGRITY_COMMIT=0f2c1b2c6f91f104b8e0880e85dc8144fed6a910
CUTOFF_INTEGRITY_GITHUB_ACTIONS_RUN=30404866760
KINEMATIC_CONFIDENCE_COMMIT=af9c377193c21c048721e9cc28bf885d6ad276ec
KINEMATIC_CONFIDENCE_GITHUB_ACTIONS_RUN=30406050920
COLLABORATION_INTEGRITY_COMMIT=560e4ed15cabbf0042110e00363a3a7c4d0c0d2e
COLLABORATION_INTEGRITY_GITHUB_ACTIONS_RUN=30407620031
COLLABORATION_INTEGRITY_BACKEND_RACE_SAFETY_JOB=90436489162
COLLABORATION_INTEGRITY_POSTGRESQL_16_INTEGRATION_JOB=90436489212
COLLABORATION_INTEGRITY_BACKEND_QUALITY_JOB=90436489267
COLLABORATION_INTEGRITY_BACKEND_CONTAINER_JOB=90436728619
PERMANENT_AUDIT_COMMIT=51476c427f77b5a7375cd30b6f9a81d446c1c3f2
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30408617024
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90439610654
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90439610660
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90439610677
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90439793759
```

All confirmed engineering findings are implemented. The permanent audit is committed,
enforced by Backend Continuous Integration, and verified by successful targeted, race,
full-suite, vet, PostgreSQL integration, and container jobs. No Projection Baseline
review item remains open, unclassified, or deferred.

```text
PROJECTION_BASELINE_METHOD_VERSION=short-horizon-kinematic-baseline-v3
PROJECTION_BASELINE_INPUT_FINGERPRINT_VERSION=projection-baseline-input-fingerprint-v4
PROJECTION_BASELINE_CUTOFF_ISOLATION=ENFORCED
PROJECTION_BASELINE_QUALITY_RECOMPUTATION=ENFORCED
PROJECTION_BASELINE_POSTGRES_CUTOFF_ALIGNMENT=ENFORCED
PROJECTION_BASELINE_UNAVAILABLE_PROVENANCE=ENFORCED
PROJECTION_BASELINE_OBSERVATION_AGE_CONFIDENCE=ENFORCED
PROJECTION_BASELINE_PHYSICAL_BOUNDS=ENFORCED
PROJECTION_BASELINE_ALTITUDE_REFERENCE=EXPLICIT
PROJECTION_BASELINE_VERTICAL_RATE_REFERENCE_LIMITATION=EXPLICIT
PROJECTION_BASELINE_LATEST_OBSERVATION_AMBIGUITY=REJECTED
PROJECTION_BASELINE_ON_GROUND_MODEL=STATIONARY_LIMITED
PROJECTION_BASELINE_HORIZONTAL_FALLBACK=EXPLICIT_POLICY
PROJECTION_BASELINE_ELIGIBILITY_OUTPUT_VALIDATION=ENFORCED
PROJECTION_BASELINE_ELIGIBILITY_POLICY_PROVENANCE=ENFORCED
PROJECTION_BASELINE_HORIZON_IDENTITY=CANONICAL_PLAN_FINGERPRINT
PROJECTION_BASELINE_FAILED_CONSTRUCTOR_NIL_RESULT=IDIOMATIC_GO_RETAINED
PROJECTION_BASELINE_OPTIONAL_ALTITUDE_POINTER=CONTRACT_SEMANTICS_RETAINED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_BASELINE_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_BASELINE_ENGINEERING_DEBT=CLOSED
PROJECTION_BASELINE_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_BASELINE_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

The following records retrospectively normalize the accepted review findings into the
repository-wide nineteen-field finding standard. Severity values are retrospective
engineering classifications; the historical review did not record severity labels per
accepted bullet. The original document records grouped engineering commits rather than
one commit per bullet, so exact evidence below is attributed to the smallest named
remediation wave supported by the source instead of inventing finer historical
ownership.

### GFA-DATA-328 — Cutoff snapshots retained post-`as_of` aggregate quality evidence

1. **Finding / symptom.** Point filtering respected `as_of`, but segment, gap, and aggregate quality evidence could still include later event-time information.
2. **Root cause.** Snapshot construction filtered points independently from the evidence used to derive trajectory quality.
3. **Failure scenario.** Historical replay at time T consumes a quality score influenced by a segment or gap completed after T.
4. **Impact.** A projection can be admitted, denied, or scored using knowledge unavailable at the replay cutoff.
5. **Severity rationale.** P1 retrospective because this violates historical replay truth and can materially change projection output.
6. **Existing guarantees violated.** Historical `as_of` isolation, temporal correctness, reproducible provenance.
7. **Considered solutions.** Filter points only; filter all event-time evidence; discard quality entirely for historical replay.
8. **Chosen remediation.** Build one cutoff snapshot containing only points, segments, and gaps completed by `as_of`, then recompute quality from that evidence.
9. **Why selected.** It preserves useful quality while maintaining one coherent temporal evidence boundary.
10. **Rejected alternatives.** Point-only filtering remained leaky; discarding all quality unnecessarily destroyed valid evidence.
11. **Trade-offs.** Some historical snapshots become unavailable when completed quality cannot cover the latest included observation.
12. **Regression tests / protection.** Cutoff tests vary future segment/gap evidence and verify recomputed quality and fail-closed behavior.
13. **Adversarial review findings.** `CoverageGap.CreatedAt` was deliberately rejected as event-time truth; completed `EndTime` remains authoritative.
14. **Remediation iterations.** The former point-only snapshot was replaced by `buildCutoffSnapshot` and cutoff-safe quality recomputation.
15. **Residual risks / limitations.** Correctness still depends on upstream segment and gap event timestamps being truthful.
16. **Operational/deployment consequences.** No migration; some replay requests may return unavailable instead of an optimistic projection.
17. **Exact evidence.** `0f2c1b2c6f91f104b8e0880e85dc8144fed6a910`; Backend CI run `30404866760`; permanent audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2`, run `30408617024`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectionbaselinereviewaudit` permanently checks cutoff-safe point/segment/gap and quality semantics.

### GFA-DB-329 — PostgreSQL Projection hydration returned cutoff-unsafe trajectory quality

1. **Finding / symptom.** PostgreSQL hydration filtered future observations but could return aggregate trajectory quality derived from later evidence.
2. **Root cause.** Persistence hydration and Projection Baseline snapshot semantics were not aligned on completed segment/gap evidence and quality recomputation.
3. **Failure scenario.** A database-backed replay and an equivalent in-memory cutoff view produce different quality evidence at the same `as_of`.
4. **Impact.** Production read paths could bypass the temporal guarantee enforced by the domain builder.
5. **Severity rationale.** P1 retrospective because the production PostgreSQL path is an authoritative Projection input source.
6. **Existing guarantees violated.** Storage/domain semantic parity and historical snapshot correctness.
7. **Considered solutions.** Trust persisted aggregate quality; recompute only in Baseline; align PostgreSQL hydration with the same cutoff policy.
8. **Chosen remediation.** PostgreSQL hydration applies completed segment/gap cutoff and recomputes trajectory quality before returning the trajectory.
9. **Why selected.** It prevents transport-path differences from changing the semantic snapshot.
10. **Rejected alternatives.** Trusting stale aggregate fields retained leakage; relying only on a later caller made repository output misleading.
11. **Trade-offs.** Additional read-side computation is accepted for correctness.
12. **Regression tests / protection.** PostgreSQL integration verifies cutoff-aligned evidence and recomputed quality.
13. **Adversarial review findings.** The repository boundary was treated as part of the defect, not merely a Baseline implementation detail.
14. **Remediation iterations.** Repository cutoff logic was aligned after the domain cutoff defect was established.
15. **Residual risks / limitations.** Other consumers must still honor the semantics of the hydrated cutoff-safe trajectory.
16. **Operational/deployment consequences.** No schema migration; historical Projection reads may expose less optimistic quality.
17. **Exact evidence.** Cutoff integrity commit `0f2c1b2c6f91f104b8e0880e85dc8144fed6a910`, run `30404866760`; permanent PostgreSQL/audit evidence in `51476c427f77b5a7375cd30b6f9a81d446c1c3f2`, run `30408617024`, PostgreSQL job `90439610677`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit checks PostgreSQL cutoff alignment and CI keeps PostgreSQL integration mandatory.

### GFA-TEST-330 — Future-evidence regression could not detect aggregate quality leakage

1. **Finding / symptom.** The previous future-point regression changed point visibility without changing the aggregate quality source.
2. **Root cause.** The fixture did not include future segment evidence capable of altering the quality score.
3. **Failure scenario.** A temporal regression reintroduces aggregate leakage while the future-point test remains green.
4. **Impact.** The most important replay-cutoff invariant could regress undetected.
5. **Severity rationale.** P2 retrospective because this is a protection gap around a P1 temporal defect.
6. **Existing guarantees violated.** Regression-test effectiveness and evidence-based closure.
7. **Considered solutions.** Keep structural point assertions; add a quality-changing future segment; rely only on source audit.
8. **Chosen remediation.** Make the fixture contain segment evidence and assert cutoff-safe quality behavior.
9. **Why selected.** It tests the semantic failure mode rather than an incidental implementation detail.
10. **Rejected alternatives.** Point-count-only assertions could not prove aggregate isolation; audit-only protection lacked executable behavior.
11. **Trade-offs.** Fixtures are slightly more elaborate because they model real segment evidence.
12. **Regression tests / protection.** Dedicated Baseline cutoff regression plus permanent source audit.
13. **Adversarial review findings.** The review explicitly identified the former test as incapable of detecting the leak.
14. **Remediation iterations.** Baseline test trajectory gained authoritative segment evidence during cutoff hardening.
15. **Residual risks / limitations.** Future tests still need to vary the value that owns the asserted semantic outcome.
16. **Operational/deployment consequences.** Test-only behavior; no runtime deployment effect.
17. **Exact evidence.** `0f2c1b2c6f91f104b8e0880e85dc8144fed6a910`, including Baseline fixture changes; run `30404866760`; audit run `30408617024`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent Baseline review audit requires the regression coverage to remain present.

### GFA-DATA-331 — Unavailable Projection Baseline results lacked reproducible denial provenance

1. **Finding / symptom.** Unavailable results could omit the evidence needed to explain why Projection Baseline denied output.
2. **Root cause.** Provenance construction was richer on successful paths than on fail-closed paths.
3. **Failure scenario.** Two unavailable results look equivalent even though different trajectory, cutoff, or policy evidence caused denial.
4. **Impact.** Historical analysis cannot independently reconstruct a denial decision.
5. **Severity rationale.** P1 retrospective because unavailable evidence is part of the research result, not merely an error message.
6. **Existing guarantees violated.** Provenance completeness, reproducibility, explainability.
7. **Considered solutions.** Keep limitation text only; attach generic source data; publish deterministic unavailable input references and fingerprint evidence.
8. **Chosen remediation.** Unavailable results publish deterministic input fingerprints, input references, latest observation time, policy evidence, and provenance.
9. **Why selected.** It makes denial a first-class inspectable analytical result.
10. **Rejected alternatives.** Limitation text alone was not enough to bind the decision to exact inputs.
11. **Trade-offs.** Unavailable payloads carry more metadata.
12. **Regression tests / protection.** Unavailable-result provenance is validated by focused tests and strict review audit.
13. **Adversarial review findings.** The remediation preserves the distinction between unavailable analytical output and runtime failure.
14. **Remediation iterations.** Provenance was added to fail-closed cutoff and later collaborator/eligibility paths.
15. **Residual risks / limitations.** Provenance can only be as precise as the upstream source metadata available to Baseline.
16. **Operational/deployment consequences.** Response metadata changes only; no migration.
17. **Exact evidence.** Cutoff integrity `0f2c1b2c6f91f104b8e0880e85dc8144fed6a910`; collaboration integrity `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`; permanent audit run `30408617024`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit requires unavailable-result provenance and canonical input evidence.

### GFA-DATA-332 — Projection Baseline fingerprint omitted output-affecting evidence and policy identity

1. **Finding / symptom.** Source identity, ground state, altitude reference, observation-age policy, physical constants, fallback policy, and eligibility policy were incompletely represented in input identity.
2. **Root cause.** Fingerprinting evolved incrementally and did not yet bind the complete effective decision policy.
3. **Failure scenario.** Two materially different Baseline decisions can share an input fingerprint after a policy or evidence mutation.
4. **Impact.** Stored or compared results can appear identical when their analytical semantics differ.
5. **Severity rationale.** P1 retrospective because fingerprint collision breaks provenance and reproducibility.
6. **Existing guarantees violated.** Deterministic identity and semantic provenance.
7. **Considered solutions.** Hash result outputs; duplicate Horizon fields locally; fingerprint all effective inputs/policies while consuming canonical Horizon identity.
8. **Chosen remediation.** Advance Baseline fingerprint v4 and bind complete effective evidence/policy identity plus the canonical Horizon Plan fingerprint.
9. **Why selected.** It keeps one owner per semantic identity and avoids competing Horizon encodings.
10. **Rejected alternatives.** Result status/limitations remain derived outputs; duplicating Horizon fields would create drift-prone identity mirrors.
11. **Trade-offs.** Any decision-relevant policy change intentionally invalidates prior Baseline input identity.
12. **Regression tests / protection.** Fingerprint mutation tests and permanent audit cover evidence and policy fields.
13. **Adversarial review findings.** Horizon version/truncation/timestamps were correctly classified as already owned by `plan.Fingerprint` rather than new defects.
14. **Remediation iterations.** Fingerprint ownership was extended across kinematic and collaboration hardening waves.
15. **Residual risks / limitations.** Custom evaluators without a published identity receive an explicit unversioned type-derived identity, which is less strong than a versioned provider contract.
16. **Operational/deployment consequences.** Historical Baseline fingerprints change by design; no database migration is required by this review.
17. **Exact evidence.** Kinematic/confidence `af9c377193c21c048721e9cc28bf885d6ad276ec`; collaboration integrity `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`; permanent audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Fingerprint-version assertions and audit require effective policy/evidence identity to remain bound.

### GFA-DATA-333 — Baseline confidence ignored latest-observation age

1. **Finding / symptom.** Confidence could remain high even when the latest admissible observation was old at projection `as_of`.
2. **Root cause.** Confidence combined trajectory quality and horizon decay without an observation-freshness component.
3. **Failure scenario.** A stale but historically high-quality trajectory receives confidence comparable to a fresh trajectory.
4. **Impact.** Confidence overstates evidence quality and weakens downstream decision semantics.
5. **Severity rationale.** P1 retrospective because confidence is published analytical evidence used by downstream Projection components.
6. **Existing guarantees violated.** Evidence-aware confidence and conservative research semantics.
7. **Considered solutions.** Ignore freshness; hard-reject after a fixed age only; incorporate bounded age decay alongside quality and horizon decay.
8. **Chosen remediation.** Confidence explicitly combines cutoff-safe quality, latest-observation age, and forecast-horizon retention.
9. **Why selected.** It preserves usable stale evidence while reducing confidence continuously and transparently.
10. **Rejected alternatives.** Ignoring age remained optimistic; a single hard cutoff discarded useful graded evidence.
11. **Trade-offs.** Confidence now depends on an additional versioned policy.
12. **Regression tests / protection.** Observation-age decay and non-finite computation tests.
13. **Adversarial review findings.** Non-finite values fail closed rather than being silently clamped into a plausible score.
14. **Remediation iterations.** Added during the kinematic/confidence hardening wave.
15. **Residual risks / limitations.** Age decay remains a research policy, not empirically calibrated aviation probability.
16. **Operational/deployment consequences.** Confidence values may decrease for stale historical inputs.
17. **Exact evidence.** `af9c377193c21c048721e9cc28bf885d6ad276ec`, run `30406050920`; audit run `30408617024`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects observation-age confidence semantics and policy fingerprinting.

### GFA-DATA-334 — Projection Baseline kinematics lacked conservative physical bounds

1. **Finding / symptom.** Ground speed, vertical rate, altitude, heading, or allowed ground motion could reach projection math without one conservative plausibility policy.
2. **Root cause.** Early Baseline logic validated shape/availability more strongly than physical plausibility.
3. **Failure scenario.** Corrupt but finite telemetry creates physically implausible projected positions or altitudes.
4. **Impact.** Research output can be numerically valid yet physically misleading.
5. **Severity rationale.** P1 retrospective because invalid kinematics directly alter forecast coordinates.
6. **Existing guarantees violated.** Conservative model boundary and input integrity.
7. **Considered solutions.** Clamp values; trust upstream validation; reject outside explicit conservative bounds.
8. **Chosen remediation.** Introduce explicit bounded kinematic policy covering motion and altitude inputs.
9. **Why selected.** Fail-closed rejection preserves evidence honesty better than silent normalization.
10. **Rejected alternatives.** Clamping manufactures plausible telemetry; upstream-only validation does not protect custom or historical data paths.
11. **Trade-offs.** Some extreme observations become unavailable rather than projected.
12. **Regression tests / protection.** Physical-bound tests cover speed, heading, vertical rate, altitude, and on-ground movement.
13. **Adversarial review findings.** Limits are deliberately conservative research guards, not operational aircraft envelopes.
14. **Remediation iterations.** Centralized in the kinematic policy during Baseline v2/v3 hardening.
15. **Residual risks / limitations.** Static bounds cannot capture every legitimate aircraft/phase-of-flight regime.
16. **Operational/deployment consequences.** More invalid observations fail closed; no infrastructure change.
17. **Exact evidence.** `af9c377193c21c048721e9cc28bf885d6ad276ec`, run `30406050920`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects policy constants and physical-bound regression coverage.

### GFA-DATA-335 — Altitude and vertical-rate evidence lacked explicit reference semantics

1. **Finding / symptom.** Published Baseline output did not make clear whether geometric or barometric altitude was selected, while vertical-rate provenance was source-limited.
2. **Root cause.** Altitude availability was represented as value plus boolean rather than a typed reference selection.
3. **Failure scenario.** Two projections with equal numeric altitude but different measurement references appear semantically identical.
4. **Impact.** Consumers cannot reconstruct altitude evidence or understand a vertical-rate reference limitation.
5. **Severity rationale.** P1 retrospective because reference identity changes the meaning of analytical altitude evidence.
6. **Existing guarantees violated.** Provenance specificity and measurement semantics.
7. **Considered solutions.** Keep boolean availability; force one altitude source; publish typed geometric/barometric/unavailable reference.
8. **Chosen remediation.** Introduce typed altitude selection and explicit provenance limitation for the selected reference and vertical-rate source boundary.
9. **Why selected.** It preserves available evidence without pretending heterogeneous altitude references are equivalent.
10. **Rejected alternatives.** Forcing one source reduces coverage; boolean-only output erases meaning.
11. **Trade-offs.** More provenance and fingerprint fields must remain versioned.
12. **Regression tests / protection.** Reference-selection and fingerprint/provenance tests.
13. **Adversarial review findings.** Nullable altitude remains a deliberate contract state for horizontal-only results.
14. **Remediation iterations.** Added as part of kinematic/confidence hardening.
15. **Residual risks / limitations.** Vertical-rate reference remains explicitly limited by upstream source semantics.
16. **Operational/deployment consequences.** Metadata becomes more explicit; no migration.
17. **Exact evidence.** `af9c377193c21c048721e9cc28bf885d6ad276ec`, run `30406050920`; permanent audit run `30408617024`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit requires typed altitude reference and provenance wording to remain aligned.

### GFA-DATA-336 — Conflicting latest observations at the same timestamp were selected lexically

1. **Finding / symptom.** Multiple latest observations with the same timestamp but conflicting content could be resolved by lexical ordering.
2. **Root cause.** Deterministic ordering was mistaken for semantic disambiguation.
3. **Failure scenario.** Two contradictory observations at the identical latest instant cause one arbitrary candidate to drive Projection.
4. **Impact.** Forecast output depends on identifier ordering rather than trustworthy evidence.
5. **Severity rationale.** P1 retrospective because the selected latest point anchors the complete Baseline trajectory.
6. **Existing guarantees violated.** Evidence integrity and deterministic semantic selection.
7. **Considered solutions.** Keep lexical tie-break; merge conflicting values; reject ambiguous latest evidence.
8. **Chosen remediation.** Detect conflicting equal-latest timestamps and return unavailable with `projection_latest_observation_ambiguous`.
9. **Why selected.** Ambiguity is evidence that must be disclosed, not guessed away.
10. **Rejected alternatives.** Lexical choice was deterministic but untruthful; merging could fabricate a non-observed state.
11. **Trade-offs.** Ambiguous datasets lose an otherwise computable projection.
12. **Regression tests / protection.** Equal-timestamp conflict tests plus audit markers.
13. **Adversarial review findings.** Deterministic ordering remains useful only after semantic ambiguity checks.
14. **Remediation iterations.** Added in collaboration-integrity hardening.
15. **Residual risks / limitations.** Exact duplicate observations may still be normalized according to their own canonical equality rules.
16. **Operational/deployment consequences.** More explicit unavailable results for corrupt/ambiguous evidence.
17. **Exact evidence.** `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`, run `30407620031`; permanent audit run `30408617024`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit and regression tests require ambiguity rejection before latest-observation use.

### GFA-DATA-337 — Allowed on-ground evidence reused the airborne propagation model

1. **Finding / symptom.** When on-ground input was permitted, Baseline could still apply the same motion propagation used for airborne observations.
2. **Root cause.** Eligibility allowance and physical propagation policy were coupled through one generic kinematic path.
3. **Failure scenario.** A ground observation with heading/speed evidence produces an airborne-like moving forecast.
4. **Impact.** The model can imply unjustified future movement from weak ground evidence.
5. **Severity rationale.** P1 retrospective because it changes forecast coordinates and status semantics.
6. **Existing guarantees violated.** Conservative modeling and evidence-to-model consistency.
7. **Considered solutions.** Reject all ground input; propagate reported motion; use stationary conservative ground model with limited status.
8. **Chosen remediation.** Allowed on-ground observations use a stationary model and always publish limited status.
9. **Why selected.** It retains explicit support for ground observations without over-interpreting them.
10. **Rejected alternatives.** Blanket rejection removed configured capability; airborne propagation overstated intent.
11. **Trade-offs.** Ground projections provide less motion information by design.
12. **Regression tests / protection.** On-ground stationary/limited tests and review audit.
13. **Adversarial review findings.** The policy remains research-only and does not infer taxi intent.
14. **Remediation iterations.** Corrected in collaboration-integrity hardening.
15. **Residual risks / limitations.** Stationary behavior is conservative, not a prediction that the aircraft will actually remain stopped.
16. **Operational/deployment consequences.** Ground-result semantics change; no migration.
17. **Exact evidence.** `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`, run `30407620031`; audit run `30408617024`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit checks stationary on-ground behavior and limited status.

### GFA-CONTRACT-338 — Default eligibility made the horizontal-only Projection branch unreachable

1. **Finding / symptom.** Baseline exposed horizontal-only output, but default eligibility denied missing altitude before that fallback could be reached.
2. **Root cause.** Eligibility and fallback policies were designed independently and expressed contradictory defaults.
3. **Failure scenario.** Valid horizontal kinematics with unavailable altitude are denied even though the published Baseline contract claims a horizontal-only mode.
4. **Impact.** A documented result mode is dead under production defaults and coverage is unnecessarily lost.
5. **Severity rationale.** P2 retrospective because the contradiction is a production contract/reachability defect rather than fabricated data.
6. **Existing guarantees violated.** Configuration coherence and reachable documented behavior.
7. **Considered solutions.** Remove horizontal-only mode; loosen all eligibility; allow fallback only when missing altitude is the sole denial reason.
8. **Chosen remediation.** Add explicit horizontal fallback policy constrained to altitude-only denial.
9. **Why selected.** It restores the intended mode without bypassing unrelated safety/integrity denials.
10. **Rejected alternatives.** Broad eligibility relaxation would admit invalid evidence; removing fallback reduced useful horizontal coverage.
11. **Trade-offs.** Policy interaction becomes an explicit contract that must be versioned and fingerprinted.
12. **Regression tests / protection.** Default-policy horizontal fallback and denial-reason tests.
13. **Adversarial review findings.** Nullable altitude is therefore retained intentionally rather than treated as a design defect.
14. **Remediation iterations.** Added in collaboration-integrity hardening after the unreachable branch was identified.
15. **Residual risks / limitations.** Horizontal-only output remains limited and carries no vertical prediction guarantee.
16. **Operational/deployment consequences.** Some previously unavailable cases become limited horizontal projections.
17. **Exact evidence.** `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`, run `30407620031`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Review audit checks explicit horizontal fallback policy and eligibility alignment.

### GFA-DATA-339 — Projection Baseline trusted malformed custom eligibility output

1. **Finding / symptom.** A custom eligibility collaborator could return internally inconsistent or unknown decision/reason evidence that Baseline accepted.
2. **Root cause.** Interface conformance was treated as sufficient; semantic postconditions were incomplete.
3. **Failure scenario.** A custom evaluator returns both allowed and denied semantics, duplicate/unknown reasons, or otherwise malformed evidence and Projection proceeds.
4. **Impact.** Extensibility becomes a path around Baseline integrity guarantees.
5. **Severity rationale.** P1 retrospective because collaborator output directly authorizes or denies projection.
6. **Existing guarantees violated.** Trust-boundary validation and deterministic policy evidence.
7. **Considered solutions.** Trust collaborators; restrict to default evaluator only; validate exact decision/reason postconditions for every evaluator.
8. **Chosen remediation.** Require exactly one valid decision, coherent allowed/denied reasons, and known unique reason codes.
9. **Why selected.** It preserves extensibility while making the interface a real semantic boundary.
10. **Rejected alternatives.** Trust-only behavior was unsafe; banning custom evaluators unnecessarily narrowed the architecture.
11. **Trade-offs.** Custom implementations must satisfy stricter postconditions and identity evidence.
12. **Regression tests / protection.** Malformed collaborator-output tests and strict review audit.
13. **Adversarial review findings.** Custom evaluator identity is explicit; versioned identity is preferred but not fabricated when unavailable.
14. **Remediation iterations.** Added in collaboration-integrity hardening.
15. **Residual risks / limitations.** Semantic correctness inside an otherwise valid custom evaluator remains its implementation responsibility.
16. **Operational/deployment consequences.** Invalid plugins/collaborators now fail closed.
17. **Exact evidence.** `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`, run `30407620031`; audit run `30408617024`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects collaborator postconditions and policy-identity provenance.

### GFA-OPS-340 — Nil Projection Baseline returned an unrelated construction error

1. **Finding / symptom.** Calling a nil Baseline returned `ErrHorizonPlannerRequired`, falsely blaming a dependency rather than the receiver lifecycle.
2. **Root cause.** A reused construction-time error escaped into a runtime nil-receiver guard.
3. **Failure scenario.** Operators/tests diagnose a missing planner even though the actual fault is a nil Baseline instance.
4. **Impact.** Failure classification and incident diagnosis are misleading.
5. **Severity rationale.** P2 retrospective because the failure is fail-closed but misclassified at an operational boundary.
6. **Existing guarantees violated.** Typed error semantics and diagnosability.
7. **Considered solutions.** Keep reused error; panic; introduce a typed Baseline-unavailable error.
8. **Chosen remediation.** Return `ErrBaselineUnavailable` for nil receiver use.
9. **Why selected.** It preserves fail-closed behavior and accurately identifies the lifecycle fault.
10. **Rejected alternatives.** Panic is unnecessary; dependency error was factually wrong.
11. **Trade-offs.** Callers may need to recognize the new typed error.
12. **Regression tests / protection.** Nil-receiver error classification test and source audit.
13. **Adversarial review findings.** `New(Config) (*Baseline, error)` returning nil with a construction error remains idiomatic and is explicitly not this finding.
14. **Remediation iterations.** Error classification corrected during kinematic/collaboration hardening.
15. **Residual risks / limitations.** Other lifecycle misuse must retain similarly precise typed errors.
16. **Operational/deployment consequences.** Better diagnosis only; no behavioral availability expansion.
17. **Exact evidence.** `af9c377193c21c048721e9cc28bf885d6ad276ec` contains the typed Baseline-unavailable correction; run `30406050920`; permanent audit run `30408617024`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit distinguishes nil-receiver lifecycle semantics from constructor dependency validation.

### GFA-GOV-341 — Projection Baseline remediation lacked permanent regression and CI enforcement

1. **Finding / symptom.** The review corrections initially had no single permanent strict gate proving the hardened contracts remained wired into required CI.
2. **Root cause.** Individual tests and engineering commits existed before one consolidated source audit owned closure invariants.
3. **Failure scenario.** A future change deletes or weakens a critical Baseline guard while ordinary tests no longer exercise the exact remediation contract.
4. **Impact.** Closed findings can silently regress while documentation continues to claim closure.
5. **Severity rationale.** P2 retrospective because this is a governance/protection defect around multiple P1 analytical guarantees.
6. **Existing guarantees violated.** Evidence-based closure and permanent CI protection.
7. **Considered solutions.** Rely on normal tests; manual review checklist; add one strict package-specific audit to Backend CI.
8. **Chosen remediation.** Add `apps/api/tools/projectionbaselinereviewaudit` and execute it in Backend Continuous Integration.
9. **Why selected.** It makes critical source, test, documentation, and workflow markers fail closed together.
10. **Rejected alternatives.** Manual-only review is non-reproducible; generic tests did not prove all closure invariants.
11. **Trade-offs.** Intentional contract changes require coordinated audit/documentation updates.
12. **Regression tests / protection.** The audit itself is the permanent protection and runs alongside race, PostgreSQL, full-suite, vet, and container verification.
13. **Adversarial review findings.** The gate protects concrete invariants rather than arbitrary source-file length or naming style.
14. **Remediation iterations.** Engineering fixes were followed by a dedicated permanent-audit commit.
15. **Residual risks / limitations.** Static audit markers supplement but do not replace behavioral tests.
16. **Operational/deployment consequences.** CI becomes stricter; no runtime deployment change.
17. **Exact evidence.** `51476c427f77b5a7375cd30b6f9a81d446c1c3f2`; Backend CI run `30408617024`; Quality `90439610654`, Race `90439610660`, PostgreSQL `90439610677`, Container `90439793759`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** The permanent strict audit remains mandatory in Backend CI and binds this review record to implementation/test evidence.
