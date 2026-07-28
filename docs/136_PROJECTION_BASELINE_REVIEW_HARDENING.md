# Projection Baseline Review Hardening

Status: implementation complete; permanent audit verification pending

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
PERMANENT_AUDIT_COMMIT=PENDING
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=PENDING
```

All confirmed engineering findings are implemented. The permanent audit must now be
committed and pass Backend Continuous Integration before this record changes from
verification pending to formally closed.

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
PROJECTION_BASELINE_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_BASELINE_REVIEW_STATUS=IMPLEMENTED_PENDING_PERMANENT_AUDIT_CI
```
