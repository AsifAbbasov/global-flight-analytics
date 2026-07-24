# Document 100 — Query and Architecture Consolidation

Status: Implemented Engineering Increment v1.0
Project: Global Flight Analytics
Baseline: `a31fd8ce3fb6f42a9c90a5153f902c37e7b0f111`

## 1. Purpose

This increment closes the accepted Analytical Core findings concerning zero
query reference time, non-canonical UUID identifiers, inconsistent Metric IDs,
stored legacy calculator state, exposed executor internals, and the concrete
executor dependency owned by the metric service.

## 2. Query contracts

`metricquery.RecentRequest.Normalize` now rejects a zero reference time with
`ErrReferenceTimeRequired`.

`NormalizeTrajectoryIDs` now:

```text
trims input;
parses each UUID;
converts it to the canonical lowercase hyphenated representation;
deduplicates by the canonical representation;
returns only canonical identifiers.
```

Equivalent UUID spellings can no longer reach the repository as distinct IDs.

## 3. Canonical Metric IDs

The `metrics` package is the authoritative owner of these identifiers:

```text
traffic.active_aircraft
traffic.traffic_density
traffic.airport_activity
traffic.coverage_score
traffic.data_freshness
```

The `metricexecution` package exposes compatibility aliases to the same
constants and uses those aliases for every execution result.

## 4. Executor boundary

The legacy `calculator` constructor argument remains accepted to avoid an
unnecessary breaking change, but it is no longer retained by `Executor`.

The following internal dependency getters are removed:

```text
Executor.Calculator
Executor.ScopeGuard
Executor.ConfidenceEvaluator
Service.Executor
```

Confidence evaluation is exposed as the narrow behavior
`Executor.EvaluateConfidence`, rather than exposing the evaluator object.

`metricexecution.Service` now depends on an unexported behavioral interface
containing only:

```text
FilterTrajectories
EvaluateConfidence
```

## 5. Legacy package classification

`analytics/calculator` and `analytics/registry` remain compiled and tested as
compatibility foundation packages. They are not composed by the production
server, not stored by the runtime executor, and not exposed through the metric
service.

A later breaking-version cleanup may remove their compatibility constructor
type completely. That deletion is not required for the current runtime
architecture.

## 6. Permanent regression coverage

The increment adds tests for:

```text
zero reference-time rejection;
canonical UUID output and canonical deduplication;
canonical Metric IDs and execution aliases;
absence of stored calculator state;
absence of internal dependency getter methods;
narrow executor behavior.
```

## 7. Verification

Before changing `main`, the installer validates all patch anchors and symbol
usage, applies the patch in a detached temporary Git worktree, and runs:

```text
complete backend compilation;
targeted tests;
complete backend tests.
```

The working tree then runs targeted compilation, targeted tests, race tests,
the complete backend test suite, Go vet, architecture audits, static contracts,
documentation checks, and whitespace validation.

## 8. Remaining Analytical Core review scope

```text
replace request-parameter Coverage Score and Data Freshness endpoints with
server-owned production snapshots or explicitly separated calculator routes;

perform the final Analytical Core closure audit and evidence register.
```
<!-- ANALYTICAL-CORE-REVIEW-CLOSURE:QUERY-ARCHITECTURE -->

## 9. Post-closure resolution

Document 102 completes the original Analytical Core finding register.

```text
ANALYTICAL_CORE_REVIEW_STATUS=CLOSED
Open Query and Architecture findings: 0
```

The compatibility calculator and registry packages remain classified exactly as
recorded in Section 5. They do not form a production runtime architecture.
