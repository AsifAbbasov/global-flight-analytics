# Projection Production Review Hardening

Status: closed after Historical Projector output-lineage correction and formal reclosure

```text
REVIEW_BASELINE_COMMIT=298d3fdb2d11b1797ce3728b116702b0a978d870
ENGINEERING_CLOSURE_COMMIT=c01b6ee0affff185adeda8e7fb0e1c39681cbe8c
ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30624533886
ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91136606689
ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91136606649
ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91136606715
ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91136827987
PRIOR_FORMAL_CLOSURE_COMMIT=0f1a31f56f4baf232e978d240216068a001a184e
PRIOR_FORMAL_CLOSURE_GITHUB_ACTIONS_RUN=30626948379
PRIOR_FORMAL_CLOSURE_BACKEND_RACE_SAFETY_JOB=91144310170
PRIOR_FORMAL_CLOSURE_BACKEND_QUALITY_JOB=91144310191
PRIOR_FORMAL_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91144310201
PRIOR_FORMAL_CLOSURE_BACKEND_CONTAINER_JOB=91144541785
CORRECTIVE_ENGINEERING_COMMIT=2f352821f7ef5d1a26bbb0899bad7fc431d6363c
CORRECTIVE_ENGINEERING_GITHUB_ACTIONS_RUN=30629428359
CORRECTIVE_ENGINEERING_BACKEND_QUALITY_JOB=91152099577
CORRECTIVE_ENGINEERING_POSTGRESQL_16_INTEGRATION_JOB=91152099674
CORRECTIVE_ENGINEERING_BACKEND_RACE_SAFETY_JOB=91152099675
CORRECTIVE_ENGINEERING_BACKEND_CONTAINER_JOB=91152326899
FORMAL_CLOSURE_REOPENED_BASELINE=0f1a31f56f4baf232e978d240216068a001a184e
REOPEN_CONFIRMED_FINDING=HISTORICAL_PROJECTOR_OUTPUT_LINEAGE_BINDING
ACCEPTED_CRITICAL_FINDINGS=5
ACCEPTED_HIGH_OR_MEDIUM_FINDINGS=8
PARTIALLY_ACCEPTED_FINDINGS=3
REJECTED_STALE_FINDINGS=1
REJECTED_MECHANICAL_OR_IDIOMATIC_FINDINGS=2
SINGLE_HORIZON_PLAN=CI_CONFIRMED
IMMUTABLE_REQUEST_SNAPSHOT=CI_CONFIRMED
ROUTE_REQUEST_IDENTITY_AND_TIME_BINDING=CI_CONFIRMED
CROSS_CONTRACT_EVIDENCE_BINDING=CI_CONFIRMED
PROJECTOR_OUTPUT_POSTCONDITIONS=CI_CONFIRMED
ARRIVAL_ONLY_MUTATION_BOUNDARY=CI_CONFIRMED
UNAVAILABLE_HISTORICAL_REJECTION=CI_CONFIRMED
LIMITED_EVIDENCE_NOTICES=CI_CONFIRMED
DEPENDENCY_ERROR_CHAIN_PRESERVATION=CI_CONFIRMED
REQUEST_AND_COMPOSITION_FINGERPRINTS=CI_CONFIRMED
HISTORICAL_PROJECTOR_OUTPUT_LINEAGE_BINDING=CI_CONFIRMED
HISTORICAL_PROJECTION_PROVENANCE_RECONSTRUCTION=CI_CONFIRMED
HISTORICAL_SELECTED_NEIGHBOR_PROVENANCE=CI_CONFIRMED
PERMANENT_REVIEW_AUDIT=CI_CONFIRMED
ENGINEERING_IMPLEMENTATION=COMPLETE
ENGINEERING_DEBT=CLOSED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_PRODUCTION_FORMAL_CLOSURE=COMPLETE
PROJECTION_PRODUCTION_REVIEW_STATUS=CLOSED
```

## 1. Scope

This review covers the production orchestration boundary:

```text
apps/api/internal/projectionintelligence/projectionproduction
```

It also covers the minimum integration surfaces required to enforce one authorized
snapshot across production projection:

```text
apps/api/internal/projectionintelligence/projectionbaseline
apps/api/internal/projectionintelligence/projectioncontinuation
apps/api/internal/projectionintelligence/projectionread/composition.go
apps/api/tools/projectionproductionreviewaudit
.github/workflows/backend-ci.yml
```

The Kinematic Baseline and Historical Continuation algorithms remain unchanged. The
new entrypoints only allow production orchestration to supply an already validated
`projectionhorizon.Plan` instead of allowing each projector to rebuild a second plan.
Their previously closed reviews remain protected by their existing permanent audits.

This document does not close `projectionread` or the final end-to-end reconciliation.

## 2. Review decisions

The review correctly identified missing postconditions at dependency boundaries,
route and trajectory identity drift, inconsistent as-of boundaries, duplicate horizon
planning, an over-broad Estimated Arrival interface, mutable input slices, incomplete
fingerprints, unavailable historical result acceptance, lost dependency error chains,
and absent cross-contract regression tests.

One critical claim was stale against the reviewed baseline. Production already called
`projectioncontinuation.ProjectApproved` and Historical Continuation already used the
approved Selection and Pattern without rerunning Neighbor Selection or Pattern
Confidence. The full candidate snapshot is still required because the selected IDs
reference observed candidate trajectories. The defect was therefore not repeated
selection; it was the absence of a single plan and complete production-level lineage
and output postconditions.

The following mechanical rules were rejected:

```text
blanket prohibition of names containing With
blanket prohibition of optional pointer evidence fields
```

`ArrivalStatusWithheld` is an explicit domain state. Optional evidence pointers
correctly distinguish a dependency that was not executed from a zero-valued result.

The description of the problem as Big Design Up Front was also rejected. The actual
failure mode was orchestration contract drift and insufficient postconditions.

## 3. One immutable production request snapshot

`Request.Clone()` now deep-clones:

```text
current trajectory points, segments, and gaps
historical candidate trajectories
candidate route scope
Route Intelligence result and nested evidence
route-history summary
```

Every dependency receives a fresh defensive clone. An interface implementation can no
longer mutate caller-owned slices or alter evidence observed by a later stage.

## 4. One authorized horizon plan

Production builds and validates one `projectionhorizon.Plan` and passes that exact
snapshot to both projection strategies.

New integration entrypoints are:

```go
projectionbaseline.ProjectWithPlan(request, plan)
projectioncontinuation.ProjectApprovedWithPlan(request, plan, evidence)
```

Both entrypoints validate that the request `AsOfTime` and `RequestedDuration` match the
supplied plan. Historical Continuation also passes the same plan to its internal
kinematic fallback. No production projection branch rebuilds a second forecast grid.

The plan is published in the production Result and must match the final projection
horizon exactly.

## 5. Route binding

A generally valid Route Intelligence contract is no longer sufficient. Production
also requires:

```text
Route.TrajectoryID == CurrentTrajectory.ID
Route flight, aircraft, ICAO24, and callsign identity == current trajectory identity
Route.Window.AsOfTime <= authorized Plan.AsOfTime
Route.GeneratedAt <= production GeneratedAt
Route.Provenance.TrajectoryUpdatedAt <= authorized Plan.AsOfTime
```

A route from another trajectory or a future analytical boundary cannot authorize
Historical Continuation or Estimated Arrival.

## 6. Authorized Historical Evidence

`AuthorizedHistoricalEvidence` publishes and validates one coherent snapshot:

```text
Plan
Route
RouteHistory
RouteScope
NeighborSelection
PatternConfidence
Freshness
RouteFrequency
```

The validation binds:

```text
Selection.CurrentTrajectoryID to the current trajectory
Selection.AsOfTime to Plan.AsOfTime
Selection.RequiredContinuationDuration to Plan.EffectiveDuration
selected neighbor IDs to supplied candidate IDs
Pattern source fingerprint and selected IDs to Selection
Freshness source fingerprints, selected IDs, and as-of time to Selection, Pattern, and Plan
RouteHistory route key and as-of time to Route and Plan
RouteFrequency route key, history fingerprint, and as-of time to RouteHistory and Plan
```

A contract may therefore be valid in isolation and still be rejected when it belongs
to another production workflow.

## 7. Projector postconditions

Both projector outputs must match the authorized request:

```text
trajectory, flight, aircraft, ICAO24, and callsign identity
exact authorized horizon
exact production GeneratedAt
expected method name, version, and DecisionClass
```

Historical strategy additionally rejects `ResultStatusUnavailable`. If the configured
failure policy permits fallback, postcondition failure selects the kinematic baseline
with an explicit reason and notice. Otherwise the original dependency error is
returned.

## 8. Estimated Arrival is an output delta

Production no longer accepts a complete replacement projection from the Estimated
Arrival dependency.

`ArrivalAdapter` invokes the existing closed `projectionarrival` estimator, compares
all fields except `Arrival`, and returns only:

```go
ArrivalOutcome{
    Status,
    Estimate,
    Notices,
}
```

Any mutation of coordinates, uncertainty, confidence, method, identity, horizon,
provenance, or generation time returns `ErrArrivalProjectionMutation`. The composer
itself attaches the cloned Estimated Arrival to the already authorized position
projection.

## 9. Limited evidence disclosure

When configuration explicitly allows limited evidence, production now publishes:

```text
historical_projection_authorized_with_limited_freshness
historical_projection_authorized_with_limited_route_frequency
```

The general authorization notice is retained, but limited evidence is no longer
hidden from consumers.

## 10. Error-chain integrity

Dependency errors now use a production sentinel and preserve the original cause with
multi-error wrapping:

```go
fmt.Errorf("%w: %w", productionSentinel, dependencyCause)
```

`errors.Is` can match both the production boundary and the underlying dependency.

## 11. Fingerprint separation

The old `InputFingerprint` mixed input data with output decisions and omitted major
inputs. Version 2 separates:

```text
InputFingerprint       = canonical request, plan, and production policy snapshot
CompositionFingerprint = input fingerprint, evidence trace, strategy, fallback,
                         full projection output, Estimated Arrival, notices, and time
```

The request fingerprint includes the full candidate pool even when Neighbor Selection
fails. Different failing candidate inputs can no longer collapse to the same identity.

`Result.Validate()` recomputes the Composition Fingerprint, so post-composition point
or evidence mutation is rejected.

## 12. Regression coverage

The focused hardening suite includes:

```text
TestComposeUsesOneAuthorizedHorizonPlan
TestComposeRejectsRouteFromAnotherTrajectory
TestComposeRejectsFutureRouteEvidence
TestComposeBindsSelectionPatternAndFreshness
TestComposeBindsRouteHistoryAndFrequencyToPlan
TestComposeRejectsHistoricalProjectionPostconditionDrift
TestComposeRejectsUnavailableHistoricalProjection
TestComposeDefensivelyClonesDependencyInputs
TestComposePreservesUnderlyingDependencyError
TestComposePublishesLimitedEvidenceNotice
TestRequestFingerprintCoversCandidatesOnFallback
TestCompositionFingerprintBindsPublishedProjection
TestArrivalAdapterRejectsPositionProjectionMutation
```

The fakes now record received requests, plans, evidence, candidates, route history,
and arrival input instead of ignoring dependency arguments.

## 13. Permanent audit

The strict audit is:

```text
go run ./tools/projectionproductionreviewaudit -strict
```

It checks the single-plan entrypoints, immutable snapshot, evidence binding,
projector postconditions, arrival-only adapter, fingerprint versions, regression
tests, read composition adapter, documentation, and Backend Continuous Integration
wiring.

## 14. Prior formal closure evidence and reopen

The first engineering implementation commit was:

```text
c01b6ee0affff185adeda8e7fb0e1c39681cbe8c
```

Its exact push-triggered Backend Continuous Integration evidence was:

```text
run: 30624533886
Backend Race Safety: 91136606649 — success
PostgreSQL 16 Integration: 91136606689 — success
Backend Quality: 91136606715 — success
Backend Container: 91136827987 — success
```

The first formal-closure commit was:

```text
0f1a31f56f4baf232e978d240216068a001a184e
```

Its exact push-triggered Backend Continuous Integration evidence was:

```text
run: 30626948379
Backend Race Safety: 91144310170 — success
Backend Quality: 91144310191 — success
PostgreSQL 16 Integration: 91144310201 — success
Backend Container: 91144541785 — success
```

Those commits and runs remain valid evidence for all previously closed findings. The
formal verdict was subsequently reopened because the production dependency contract
still accepted a Historical Projector result without a typed and independently
validated receipt binding that result to the authorized Plan, Selection, Pattern, and
selected-neighbor provenance.

## 15. Historical Projector output-lineage correction

Production now uses `HistoricalProjectionAdapter` instead of injecting
`projectioncontinuation.Baseline` directly into the composer. The adapter calls the
existing plan-aware projector and then asks the concrete Historical Continuation
source to independently validate the returned projection against its immutable
continuation policy.

The validator reconstructs:

```text
continuation input fingerprint from current trajectory, Plan, Selection, Pattern, and policy
canonical current-endpoint provenance
canonical Selection and Pattern provenance
exact historical_neighbor:<selected-id> provenance inputs
latest input observation boundary
```

The adapter returns a typed `HistoricalProjectionOutcome` with an
`ApprovedProjectionLineage` receipt containing:

```text
Plan fingerprint
Selection fingerprint
Pattern fingerprint
sorted selected trajectory identifiers
validated projection input fingerprint
```

The composer compares every receipt field with the already authorized production
evidence and independently requires the projection provenance input names to contain
exactly the current endpoint, Selection, Pattern, and selected historical neighbors.
A substituted projector with valid trajectory identity, horizon, method, status, and
generation time but foreign lineage now produces controlled kinematic fallback or a
typed error according to the configured dependency-failure policy.

Focused corrective regression coverage includes:

```text
TestValidateApprovedProjectionLineageReconstructsFingerprintAndInputs
TestValidateApprovedProjectionLineageRejectsForeignFingerprint
TestValidateApprovedProjectionLineageRejectsForeignNeighborProvenance
TestComposeRejectsHistoricalProjectionLineageDrift
TestComposeReturnsHistoricalProjectionLineageErrorWhenConfigured
```

The permanent audit now requires the adapter, independent source validator, typed
lineage receipt, malicious lineage tests, runtime wiring, exact prior closure evidence,
and the exact corrective engineering Continuous Integration evidence.

## 16. Corrective engineering and formal reclosure evidence

The corrective output-lineage engineering commit is:

```text
2f352821f7ef5d1a26bbb0899bad7fc431d6363c
```

Its exact push-triggered Backend Continuous Integration evidence is:

```text
run: 30629428359
Backend Quality: 91152099577 — success
PostgreSQL 16 Integration: 91152099674 — success
Backend Race Safety: 91152099675 — success
Backend Container: 91152326899 — success
```

The run completed successfully on the exact corrective commit. The permanent Projection
Production review audit passed with the sealed Historical Projector adapter, independent
continuation-fingerprint reconstruction, exact selected-neighbor provenance binding,
and malicious output-lineage regression tests enabled.

All confirmed Projection Production findings are therefore resolved in the formal
reclosure record. The formal-reclosure commit that contains this record must itself pass
Backend Quality, Backend Race Safety, PostgreSQL 16 Integration, and Backend Container
before the external final closure verdict is issued.

No statement in this document closes `projectionread` or the final end-to-end project
reconciliation.

## Canonical remediation history

The following thirteen records reconcile the accepted Production review and its later formal re-open/re-close without erasing earlier valid evidence. The stale repeated-selection claim remains rejected. The later lineage correction is recorded as a real reopened defect with its own corrective evidence. Severity is retrospective.

### GFA-DATA-411 — Production projection strategies could build different horizon plans for one request

1. **Finding / symptom.** Baseline and Historical Continuation could independently rebuild forecast grids for the same production request.
2. **Root cause.** Horizon planning was owned by individual strategies instead of the production orchestration snapshot.
3. **Failure scenario.** Equivalent request policy produces two different effective/truncated plan snapshots across strategy or fallback paths.
4. **Impact.** Strategy choice can change the authorized forecast grid independently of request evidence.
5. **Severity rationale.** P1 retrospective because the horizon is a core identity and output boundary.
6. **Existing guarantees violated.** Single-snapshot orchestration and deterministic forecast identity.
7. **Considered solutions.** Let strategies rebuild and compare; cache planner output; build one validated plan and inject it everywhere.
8. **Chosen remediation.** Production creates one `projectionhorizon.Plan`, validates it, publishes it, and passes it through `ProjectWithPlan` / `ProjectApprovedWithPlan` including fallback.
9. **Why selected.** One immutable owner removes plan drift rather than detecting it after execution.
10. **Rejected alternatives.** Independent recomputation retains duplicate semantic ownership.
11. **Trade-offs.** Projector integration APIs become explicitly plan-aware.
12. **Regression tests / protection.** `TestComposeUsesOneAuthorizedHorizonPlan` and strict Production review audit.
13. **Adversarial review findings.** The previously reviewed Baseline/Continuation algorithms are not reopened by adding plan-aware entrypoints.
14. **Remediation iterations.** Closed in the initial Production engineering wave.
15. **Residual risks / limitations.** Correctness depends on the Horizon Plan validator remaining authoritative.
16. **Operational/deployment consequences.** No migration; all production branches share one exact grid.
17. **Exact evidence.** `c01b6ee0affff185adeda8e7fb0e1c39681cbe8c`, run `30624533886`; prior closure `0f1a31f56f4baf232e978d240216068a001a184e`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectionproductionreviewaudit` protects single-plan entrypoints and wiring.

### GFA-DATA-412 — Production dependencies could mutate caller-owned request evidence

1. **Finding / symptom.** Current trajectory, historical candidates, route scope, Route evidence or history slices could be shared by reference across dependency calls.
2. **Root cause.** The orchestration boundary lacked one complete defensive-clone policy.
3. **Failure scenario.** A dependency mutates candidate points or route evidence and a later stage observes altered input.
4. **Impact.** Output becomes dependency-order-sensitive and provenance no longer represents the original request.
5. **Severity rationale.** P1 retrospective because mutable aliasing can silently alter production analytical evidence.
6. **Existing guarantees violated.** Immutable request snapshot and deterministic dependency isolation.
7. **Considered solutions.** Trust interfaces; shallow clone; deep clone each dependency input.
8. **Chosen remediation.** `Request.Clone()` deep-clones nested trajectory points/segments/gaps, candidates, route scope/result and route history; each dependency receives a fresh clone.
9. **Why selected.** It makes mutation safety local to the orchestration boundary.
10. **Rejected alternatives.** Interface trust cannot enforce immutability; shallow copies retain nested aliasing.
11. **Trade-offs.** Additional allocation/copy cost is accepted for bounded production requests.
12. **Regression tests / protection.** `TestComposeDefensivelyClonesDependencyInputs` and argument-recording fakes.
13. **Adversarial review findings.** The fix protects against both accidental and adversarial interface implementations.
14. **Remediation iterations.** Closed in the initial Production hardening.
15. **Residual risks / limitations.** Newly added mutable fields must be included in future clone updates.
16. **Operational/deployment consequences.** Small memory/CPU overhead; no persistence change.
17. **Exact evidence.** `c01b6ee0affff185adeda8e7fb0e1c39681cbe8c`, run `30624533886`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit and defensive-clone regressions protect snapshot ownership.

### GFA-DATA-413 — Route evidence was not fully bound to current trajectory identity and authorized time

1. **Finding / symptom.** A generally valid Route Intelligence result could belong to another trajectory or future analytical boundary and still reach Production.
2. **Root cause.** Production trusted isolated Route validity without workflow-specific postconditions.
3. **Failure scenario.** A route for a different flight/aircraft or route evidence updated after Projection `AsOfTime` authorizes historical continuation or ETA.
4. **Impact.** Projection strategy and arrival can be driven by foreign/future route evidence.
5. **Severity rationale.** P1 retrospective because route evidence directly authorizes downstream behavior.
6. **Existing guarantees violated.** Entity ownership and temporal lineage.
7. **Considered solutions.** Trust Route validator; compare only trajectory ID; bind complete identity and time tuple.
8. **Chosen remediation.** Require route trajectory/flight/aircraft/ICAO24/callsign equality and bound Route as-of/generated/provenance update times to the authorized plan/production time.
9. **Why selected.** It validates workflow ownership, not only standalone structural correctness.
10. **Rejected alternatives.** Trajectory-only checks miss identity mirrors and future evidence.
11. **Trade-offs.** Previously reusable but weakly owned Route objects may be rejected.
12. **Regression tests / protection.** Another-trajectory and future-route-evidence tests.
13. **Adversarial review findings.** A valid contract can still be invalid for a specific production workflow.
14. **Remediation iterations.** Closed in initial Production hardening.
15. **Residual risks / limitations.** Upstream route provenance must expose truthful update times.
16. **Operational/deployment consequences.** Foreign/future routes fail closed or select configured fallback.
17. **Exact evidence.** `c01b6ee0affff185adeda8e7fb0e1c39681cbe8c`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects route identity/time postconditions.

### GFA-DATA-414 — Independently valid historical contracts were not bound into one authorized production evidence graph

1. **Finding / symptom.** Selection, Pattern Confidence, Freshness, Route History and Route Frequency could each validate independently while belonging to different requests or plans.
2. **Root cause.** Cross-contract lineage was implicit rather than represented by one `AuthorizedHistoricalEvidence` snapshot.
3. **Failure scenario.** Freshness from one selection or Route Frequency from another history object authorizes the current workflow.
4. **Impact.** Historical continuation can be approved from semantically mixed evidence.
5. **Severity rationale.** P1 retrospective because authorization integrity depends on the complete evidence graph.
6. **Existing guarantees violated.** Cross-contract lineage and single-snapshot authorization.
7. **Considered solutions.** Trust individual validators; compare selected IDs ad hoc; validate one typed evidence aggregate.
8. **Chosen remediation.** Publish/validate `AuthorizedHistoricalEvidence` binding Plan, Route, RouteHistory, RouteScope, Selection, Pattern, Freshness and RouteFrequency fingerprints/status/times/IDs.
9. **Why selected.** One typed aggregate centralizes the exact production relationship.
10. **Rejected alternatives.** Ad hoc pairwise checks are easier to omit and harder to audit.
11. **Trade-offs.** Production composition has a stronger, larger validation contract.
12. **Regression tests / protection.** Selection/Pattern/Freshness and RouteHistory/Frequency binding tests.
13. **Adversarial review findings.** The stale claim that Selection/Pattern were rerun was rejected; the real defect was production-level binding.
14. **Remediation iterations.** Closed in the initial Production evidence-integrity wave.
15. **Residual risks / limitations.** New authorization dependencies must be added explicitly to the evidence graph.
16. **Operational/deployment consequences.** Mixed evidence fails closed or falls back according to policy.
17. **Exact evidence.** `c01b6ee0affff185adeda8e7fb0e1c39681cbe8c`, run `30624533886`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects the typed evidence graph and focused regressions.

### GFA-DATA-415 — Projector outputs lacked complete production postconditions

1. **Finding / symptom.** A dependency could return a structurally valid projection with drifted identity, horizon, generation time, method or decision class.
2. **Root cause.** Production treated projector success as sufficient without validating the result against the authorized request.
3. **Failure scenario.** A substituted projector returns a result for the right type but wrong trajectory/grid/method and Production publishes it.
4. **Impact.** Dependency boundaries become a path around orchestration authorization.
5. **Severity rationale.** P1 retrospective because the final published projection can be foreign to the request.
6. **Existing guarantees violated.** Trust-boundary postconditions and result ownership.
7. **Considered solutions.** Trust concrete implementations; validate only `Result.Validate()`; enforce production-specific result postconditions.
8. **Chosen remediation.** Require exact identities, horizon, production `GeneratedAt`, method/version and DecisionClass; Historical strategy also rejects unavailable status.
9. **Why selected.** It separates standalone result validity from workflow authorization.
10. **Rejected alternatives.** General validation cannot prove request ownership.
11. **Trade-offs.** Custom projectors must satisfy stricter postconditions.
12. **Regression tests / protection.** Historical projection postcondition-drift and unavailable-result tests.
13. **Adversarial review findings.** Failure policy explicitly decides controlled fallback versus typed error.
14. **Remediation iterations.** Initial postconditions were later strengthened by the lineage receipt correction.
15. **Residual risks / limitations.** Generic fields cannot prove internal derivation; that residual gap became the reopened lineage finding below.
16. **Operational/deployment consequences.** Drifted results fail closed/fallback.
17. **Exact evidence.** Initial `c01b6ee0affff185adeda8e7fb0e1c39681cbe8c`; later lineage correction `2f352821f7ef5d1a26bbb0899bad7fc431d6363c`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects both generic postconditions and the stronger lineage receipt.

### GFA-CONTRACT-416 — Estimated Arrival dependency could replace the entire authorized position projection

1. **Finding / symptom.** The arrival interface accepted/returned a complete projection object even though it should only contribute the Arrival delta.
2. **Root cause.** Dependency capability exceeded its semantic responsibility.
3. **Failure scenario.** Arrival code mutates coordinates, uncertainty, confidence, method, horizon or provenance while attaching ETA.
4. **Impact.** A downstream enrichment dependency can overwrite an already authorized projection.
5. **Severity rationale.** P1 retrospective because the interface permits silent corruption of core projection output.
6. **Existing guarantees violated.** Least-capability interfaces and projection immutability.
7. **Considered solutions.** Trust arrival implementation; compare before/after full projection; expose an arrival-only outcome contract and verify no mutation.
8. **Chosen remediation.** `ArrivalAdapter` validates that every non-Arrival field is unchanged and returns only `ArrivalOutcome`; composer attaches the cloned estimate.
9. **Why selected.** The interface now matches the only output Arrival is authorized to change.
10. **Rejected alternatives.** Trust-only behavior leaves an over-broad mutation surface.
11. **Trade-offs.** Adapter code performs an explicit projection comparison.
12. **Regression tests / protection.** `TestArrivalAdapterRejectsPositionProjectionMutation`.
13. **Adversarial review findings.** Optional Arrival presence remains a valid shared-contract state and was not treated as a defect.
14. **Remediation iterations.** Closed in initial Production hardening.
15. **Residual risks / limitations.** Future permitted deltas require explicit contract expansion rather than mutation-by-convention.
16. **Operational/deployment consequences.** Unauthorized arrival mutations now fail with `ErrArrivalProjectionMutation`.
17. **Exact evidence.** `c01b6ee0affff185adeda8e7fb0e1c39681cbe8c`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects arrival-only adapter semantics.

### GFA-DATA-417 — Production accepted unavailable Historical Continuation as a successful strategy result

1. **Finding / symptom.** Historical projector success could still carry `ResultStatusUnavailable` and be treated as an ordinary production result.
2. **Root cause.** Transport-level success and analytical availability were conflated.
3. **Failure scenario.** Historical strategy returns a validated but unavailable domain result and Production publishes it instead of applying configured fallback/error policy.
4. **Impact.** Strategy authorization semantics are violated and consumers receive a result that should not have been selected.
5. **Severity rationale.** P1 retrospective because final strategy selection is wrong despite typed domain evidence.
6. **Existing guarantees violated.** Availability-aware orchestration.
7. **Considered solutions.** Publish unavailable; convert to error universally; reject at production postcondition and follow configured failure policy.
8. **Chosen remediation.** Historical strategy postconditions reject unavailable status and trigger controlled fallback or typed error.
9. **Why selected.** It preserves domain availability while keeping policy ownership in Production.
10. **Rejected alternatives.** Universal error loses configured fallback semantics; publishing unavailable misstates strategy success.
11. **Trade-offs.** Some dependency-success calls now result in fallback.
12. **Regression tests / protection.** `TestComposeRejectsUnavailableHistoricalProjection`.
13. **Adversarial review findings.** Limited historical evidence remains separately configurable and is not treated as unavailable.
14. **Remediation iterations.** Closed in initial Production hardening.
15. **Residual risks / limitations.** Strategy-specific statuses must remain explicit as future methods are added.
16. **Operational/deployment consequences.** More predictable fallback behavior.
17. **Exact evidence.** `c01b6ee0affff185adeda8e7fb0e1c39681cbe8c`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects unavailable rejection and failure-policy paths.

### GFA-DATA-418 — Authorized limited historical evidence was not disclosed explicitly in production output

1. **Finding / symptom.** Production could allow limited Freshness or Route Frequency evidence while publishing only a general historical authorization notice.
2. **Root cause.** Authorization status did not expose which supporting evidence remained limited.
3. **Failure scenario.** Consumer sees historical strategy selected but cannot tell that freshness or route-frequency support was degraded.
4. **Impact.** Evidence quality is hidden even though policy intentionally accepted it.
5. **Severity rationale.** P2 retrospective because output remains valid under policy but provenance/explanation is incomplete.
6. **Existing guarantees violated.** Limitation disclosure and explainability.
7. **Considered solutions.** General notice only; reject all limited evidence; publish specific limited-evidence notices.
8. **Chosen remediation.** Add dedicated notices for limited Freshness and limited Route Frequency while retaining general authorization notice.
9. **Why selected.** It preserves configured capability without hiding degraded support.
10. **Rejected alternatives.** Blanket rejection removes a supported policy mode; generic notice is insufficiently specific.
11. **Trade-offs.** More notices are published for limited-but-authorized results.
12. **Regression tests / protection.** `TestComposePublishesLimitedEvidenceNotice`.
13. **Adversarial review findings.** Limited and unavailable remain distinct domain states.
14. **Remediation iterations.** Closed in initial Production hardening.
15. **Residual risks / limitations.** Consumers must interpret notice codes in addition to top-level strategy.
16. **Operational/deployment consequences.** Metadata becomes more explicit.
17. **Exact evidence.** `c01b6ee0affff185adeda8e7fb0e1c39681cbe8c`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects limited-evidence notice codes and tests.

### GFA-OPS-419 — Production dependency failures lost their underlying error cause

1. **Finding / symptom.** Dependency errors could be reclassified at Production while losing machine-inspectable original causality.
2. **Root cause.** Error wrapping did not preserve both production sentinel and dependency cause.
3. **Failure scenario.** Caller can detect a production-stage failure but cannot match the original dependency error with `errors.Is`.
4. **Impact.** Diagnosis and typed recovery are weakened.
5. **Severity rationale.** P2 retrospective because failures remain fail-closed but causality is operationally important.
6. **Existing guarantees violated.** Error-chain integrity and diagnosability.
7. **Considered solutions.** Concatenate text; expose raw cause only; multi-wrap production sentinel and cause.
8. **Chosen remediation.** Use `fmt.Errorf("%w: %w", productionSentinel, dependencyCause)`.
9. **Why selected.** Both orchestration stage and root cause remain machine-matchable.
10. **Rejected alternatives.** Text flattening is fragile; cause-only errors lose stage ownership.
11. **Trade-offs.** Callers should use Go error-chain semantics instead of string matching.
12. **Regression tests / protection.** `TestComposePreservesUnderlyingDependencyError`.
13. **Adversarial review findings.** This applies to dependency failure, not ordinary domain limitation notices.
14. **Remediation iterations.** Closed in initial Production hardening.
15. **Residual risks / limitations.** External logging layers must preserve, not flatten, the chain.
16. **Operational/deployment consequences.** Improved incident diagnostics only.
17. **Exact evidence.** `c01b6ee0affff185adeda8e7fb0e1c39681cbe8c`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Strict audit protects multi-wrap semantics and regression coverage.

### GFA-DATA-420 — Production input and composed-output identity were conflated and incomplete

1. **Finding / symptom.** One `InputFingerprint` mixed request evidence with output decisions while omitting major inputs such as the full candidate pool.
2. **Root cause.** Semantic request identity and final composition identity were owned by one under-specified fingerprint.
3. **Failure scenario.** Different failing candidate pools collapse to one identity, or output mutations are not reflected distinctly from request identity.
4. **Impact.** Production replay, comparison and provenance cannot distinguish materially different inputs/results.
5. **Severity rationale.** P1 retrospective because deterministic identity is foundational to the research result.
6. **Existing guarantees violated.** Input/output identity separation and reproducibility.
7. **Considered solutions.** Expand one mixed fingerprint; hash result only; separate canonical request fingerprint from composition fingerprint.
8. **Chosen remediation.** Version Two defines `InputFingerprint` over request/plan/policy and `CompositionFingerprint` over input identity, trace, strategy/fallback, full projection, Arrival, notices and time.
9. **Why selected.** It assigns one clear semantic purpose to each identity.
10. **Rejected alternatives.** One mixed hash obscures whether input or output changed.
11. **Trade-offs.** Two versioned fingerprints must be maintained.
12. **Regression tests / protection.** Candidate-pool-on-fallback and published-projection mutation tests; validator recomputes composition fingerprint.
13. **Adversarial review findings.** Full candidate pool is included even when Neighbor Selection fails.
14. **Remediation iterations.** Closed in initial Production fingerprint hardening.
15. **Residual risks / limitations.** Newly decision-relevant request/output fields require coordinated version updates.
16. **Operational/deployment consequences.** Production fingerprints change by design.
17. **Exact evidence.** `c01b6ee0affff185adeda8e7fb0e1c39681cbe8c`, run `30624533886`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects both fingerprint versions and mutation tests.

### GFA-DATA-421 — Historical projector output was not cryptographically/semantically bound to the already authorized evidence lineage

1. **Finding / symptom.** After the first formal closure, Production still accepted a Historical projector result that satisfied generic postconditions without a typed receipt proving derivation from the authorized Plan, Selection, Pattern and selected neighbors.
2. **Root cause.** Output postconditions validated surface result fields but not the projector's internal approved evidence lineage.
3. **Failure scenario.** A substituted projector returns correct identity/horizon/method/status with a foreign continuation fingerprint and Production accepts it.
4. **Impact.** Final historical projection can be generated from evidence different from the authorization chain while appearing valid.
5. **Severity rationale.** P1 retrospective because this reopened the formal Production closure on a core lineage guarantee.
6. **Existing guarantees violated.** End-to-end authorization-to-output lineage and evidence provenance.
7. **Considered solutions.** Trust concrete projector; compare generic result fields; independently reconstruct source lineage and return a typed receipt.
8. **Chosen remediation.** Introduce `HistoricalProjectionAdapter` and `ApprovedProjectionLineage` receipt with exact Plan/Selection/Pattern/selected IDs/validated projection-input fingerprint.
9. **Why selected.** Production receives independently validated proof tied to the same evidence it already authorized.
10. **Rejected alternatives.** Generic postconditions cannot prove derivation; concrete-type trust is not an interface guarantee.
11. **Trade-offs.** Historical projector integration becomes more explicit and concrete-source validation is required.
12. **Regression tests / protection.** Foreign-fingerprint and compose-lineage-drift malicious tests.
13. **Adversarial review findings.** Earlier closure evidence remains valid for earlier findings; only this unresolved guarantee reopened the verdict.
14. **Remediation iterations.** First formal closure `0f1a31f...` was reopened; corrective engineering `2f352821...` sealed the lineage boundary.
15. **Residual risks / limitations.** Future historical projector implementations must provide equivalent independently validated lineage receipts.
16. **Operational/deployment consequences.** Foreign-lineage results now fallback or error according to dependency policy.
17. **Exact evidence.** Reopened baseline `0f1a31f56f4baf232e978d240216068a001a184e`; correction `2f352821f7ef5d1a26bbb0899bad7fc431d6363c`, run `30629428359`, jobs `91152099577`, `91152099674`, `91152099675`, `91152326899`.
18. **Final canonical status.** CLOSED after formal reclosure.
19. **Prevention / future guard.** Production review audit requires adapter, receipt, independent validator and malicious lineage tests.

### GFA-DATA-422 — Historical projection provenance could not be independently reconstructed against Continuation policy

1. **Finding / symptom.** Production could inspect published provenance but lacked a source-owned reconstruction proving it matched canonical Continuation inputs and policy.
2. **Root cause.** Provenance was consumed as output metadata rather than re-derived from the concrete historical projector's immutable semantics.
3. **Failure scenario.** A projector supplies plausible provenance labels while its input fingerprint reflects a different current endpoint or Selection/Pattern chain.
4. **Impact.** Provenance text can disagree with the actual semantic continuation identity.
5. **Severity rationale.** P1 retrospective because source provenance is part of research reproducibility.
6. **Existing guarantees violated.** Independently verifiable provenance and input identity.
7. **Considered solutions.** Trust result provenance; re-hash at Production using duplicated policy; ask Continuation source to reconstruct its own canonical fingerprint/provenance.
8. **Chosen remediation.** Concrete Continuation validator reconstructs continuation input fingerprint, current endpoint, Selection/Pattern provenance and latest observation boundary from immutable policy.
9. **Why selected.** Semantic policy remains owned by the source module instead of being duplicated in Production.
10. **Rejected alternatives.** Production-side duplicate fingerprint logic would drift from Continuation semantics.
11. **Trade-offs.** Adapter depends on a source-specific lineage validation capability.
12. **Regression tests / protection.** `TestValidateApprovedProjectionLineageReconstructsFingerprintAndInputs` and foreign fingerprint tests.
13. **Adversarial review findings.** This is a distinct sub-boundary of the reopened lineage defect, not a rewrite of prior generic postconditions.
14. **Remediation iterations.** Added in corrective engineering `2f352821...`.
15. **Residual risks / limitations.** Source validator must advance with future Continuation fingerprint versions.
16. **Operational/deployment consequences.** Invalid provenance lineage is rejected before publication.
17. **Exact evidence.** `2f352821f7ef5d1a26bbb0899bad7fc431d6363c`, run `30629428359`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects independent reconstruction and version wiring.

### GFA-DATA-423 — Historical selected-neighbor provenance was not required to match the exact authorized neighbor set

1. **Finding / symptom.** A historical projection could expose generic Selection/Pattern provenance without proving exact `historical_neighbor:<selected-id>` inputs for the authorized selected set.
2. **Root cause.** Selected-neighbor membership was validated upstream but not independently required in final projector provenance.
3. **Failure scenario.** Projection uses or claims a different historical neighbor set while generic lineage fingerprints remain superficially plausible.
4. **Impact.** Consumers cannot prove which observed historical trajectories actually supported the final continuation.
5. **Severity rationale.** P1 retrospective because neighbor evidence directly determines historical geometry.
6. **Existing guarantees violated.** Exact contributing-evidence provenance.
7. **Considered solutions.** Trust Selection IDs; publish count only; require exact sorted selected IDs in receipt and exact per-neighbor provenance names.
8. **Chosen remediation.** Receipt contains sorted selected trajectory IDs and composer requires exactly the current endpoint, Selection, Pattern and each selected historical-neighbor provenance input.
9. **Why selected.** It binds final published evidence to the exact authorized contributing set.
10. **Rejected alternatives.** Counts/fingerprints alone do not expose missing/substituted neighbor provenance.
11. **Trade-offs.** Provenance validation is stricter and grows linearly with the bounded selected set.
12. **Regression tests / protection.** Foreign-neighbor provenance and compose-lineage-drift tests.
13. **Adversarial review findings.** The full candidate pool remains an input-fingerprint concern; final provenance requires only the selected contributing set.
14. **Remediation iterations.** Closed in corrective lineage engineering.
15. **Residual risks / limitations.** Accuracy of provenance still depends on canonical neighbor IDs upstream.
16. **Operational/deployment consequences.** Missing/substituted selected-neighbor provenance triggers fallback/error.
17. **Exact evidence.** `2f352821f7ef5d1a26bbb0899bad7fc431d6363c`, run `30629428359`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Production audit protects selected-neighbor receipt/provenance equality and malicious regressions.
