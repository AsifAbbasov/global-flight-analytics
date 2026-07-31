# Projection Production Review Hardening

Status: reopened — Historical Projector output-lineage correction implemented, exact Continuous Integration and formal reclosure pending

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
HISTORICAL_PROJECTOR_OUTPUT_LINEAGE_BINDING=IMPLEMENTED_PENDING_EXACT_CI
HISTORICAL_PROJECTION_PROVENANCE_RECONSTRUCTION=IMPLEMENTED_PENDING_EXACT_CI
HISTORICAL_SELECTED_NEIGHBOR_PROVENANCE=IMPLEMENTED_PENDING_EXACT_CI
PERMANENT_REVIEW_AUDIT=UPDATED_PENDING_EXACT_CI
ENGINEERING_IMPLEMENTATION=CORRECTIVE_IMPLEMENTATION_COMPLETE_PENDING_EXACT_CI
ENGINEERING_DEBT=OPEN_PENDING_EXACT_CI_AND_FORMAL_RECLOSURE
OPEN_CONFIRMED_FINDINGS=0_PENDING_EXACT_CI
UNCLASSIFIED_FINDINGS=0_PENDING_EXACT_CI
DEFERRED_FINDINGS=0_PENDING_EXACT_CI
ADDITIONAL_CODE_FIXES_REQUIRED=NO_PENDING_EXACT_CI
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES
PROJECTION_PRODUCTION_FORMAL_CLOSURE=REOPENED_PENDING_EXACT_CI_AND_FORMAL_RECLOSURE
PROJECTION_PRODUCTION_REVIEW_STATUS=OPEN_PENDING_EXACT_CI_AND_FORMAL_RECLOSURE
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
lineage receipt, malicious lineage tests, runtime wiring, reopened documentation, and
exact prior closure evidence. Formal reclosure remains blocked until this corrective
engineering commit and the later formal-reclosure commit each pass the exact required
Backend Continuous Integration jobs.

No statement in this document closes `projectionread` or the final end-to-end project
reconciliation.
