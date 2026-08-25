# Document 42 — Stage 14.2 Dead Code Classification and Removal

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: confirmed dead package removal and explicit ownership of all remaining non-runtime analytical packages

## 1. Purpose

Stage 14.2 removes only code proven to have no runtime, operational,
verification, test, or build-tagged external importer.

Successful compilation is not considered proof that every package is needed.
Likewise, lack of production reachability is not considered sufficient proof
that a package is safe to delete.

## 2. Removed Packages

The following early Analytical Core foundation packages are removed:

```text
internal/analytics/query
internal/analytics/window
```

`analytics/query` was a minimal string-based metric request with only a time
range. Production analytical reads use the validated and bounded
`analytics/metricquery` contracts.

`analytics/window` was a generic in-memory interval wrapper. Production metric
queries and Historical Intelligence use context-specific time-window
contracts, including `analytics/metricquery` and
`historicalintelligence/historicalwindow`.

Before deletion, the installer verifies:

```text
the expected package file set
no ordinary Go importer
no test importer
no external test importer
no build-tagged source import literal
no production runtime reachability
```

Any failed proof stops the installation before deletion.

## 3. Mandatory Non-Runtime Classification

Every remaining analytical package that is not reachable from a production or
operational root must have an explicit policy.

Supported dispositions:

```text
offline_research
planned_production_integration
unintegrated_feature_pipeline
offline_evaluation
```

Unknown non-runtime packages fail strict project audit.

This prevents new tested-but-unused package trees from silently accumulating.

## 4. Current Classification

### Offline research

```text
analytics/researchbenchmark
analytics/researchdataset
```

These packages remain intentionally excluded from production runtime.

### Planned production integration

```text
analytics/transponderalert
airportintelligence/history
airportintelligence/overview
airportintelligence/passport
airportintelligence/ranking
airportintelligence/statistics
airportintelligence/trends
```

These packages must receive a real read path or be removed before release.

### Unintegrated feature pipeline

```text
features/aircraftprovider
features/datasetprofiler
features/extractor
features/extractorcomposition
features/featurepipeline
features/featurestore
features/flightfeatures
features/geographicalbuilder
features/operationalbuilder
features/temporalbuilder
features/trajectorybuilder
features/validator
```

The pipeline has PostgreSQL verification evidence but no operational command.
The next decision is binary: add an operational materialization root or remove
the complete pipeline.

### Offline evaluation

```text
projectionintelligence/projectionevaluation
```

This package must remain outside live forecast generation because it consumes
later truth. It requires a real offline benchmark entrypoint before any
calibration claim.

## 5. Release Rule

No package with disposition `planned_production_integration` or
`unintegrated_feature_pipeline` may remain unresolved at final release.

Tests and verifier commands alone do not qualify a package as a production
feature.

<!-- STAGE-14-3-AIRPORT-INTELLIGENCE-PRODUCTION:DOCUMENT-42 -->

## Airport Intelligence Resolution

Airport Intelligence was resolved through production integration rather than deletion. All six original domain packages are now reachable through the `cmd/server` dependency graph and no longer require non-runtime disposition entries.

<!-- STAGE-14-4-FEATURE-MATERIALIZATION:DOCUMENT-42 -->

## Feature Pipeline Resolution

The Feature Pipeline was resolved through an operational materialization command. Eleven packages now belong to the `feature_materializer` runtime root. `features/datasetprofiler` was removed as a confirmed isolated facade rather than being artificially imported.

<!-- STAGE-14-5-MUTATION-ENDPOINT-PROTECTION:DOCUMENT-42 -->

## Mutation Security Resolution

The architecture audit now covers mutation route authorization in addition to reachability, duplicate vocabulary, and Go-to-TypeScript contract alignment. A newly added mutation route cannot pass strict audit unless its first route middleware is `mutationAuthorization`.

<!-- STAGE-14-6-FORMULA-BENCHMARK:DOCUMENT-42 -->

## Offline Projection Evaluation Resolution

`projectionevaluation` is now reachable from `benchmark-projection-formulas` as an explicit verification root while remaining excluded from production runtime. `formulabenchmark` receives the same offline-evaluation disposition and cannot enter the backend Docker image.

## 6. Canonical finding record — GFA-REL-035

### Finding / symptom

The repository contained both confirmed obsolete analytical foundations and a larger set of tested packages that were not reachable from production/operational roots but had no explicit lifecycle disposition.

### Root cause

Before Stage 14.1/14.2, package existence, test coverage, and production reachability were not tracked as separate facts. This made two opposite mistakes possible: retaining obsolete foundations because they compiled, or deleting legitimate offline/planned code because the server did not import it.

### Failure scenario

A tested analytical package tree remains in the repository for months without any executable path and is nevertheless treated as a product capability; or cleanup deletes a valid offline evaluation/research package because it is not in the server graph. In both cases repository structure ceases to reflect product/release truth.

### Impact

Unowned code increases maintenance and review surface, creates misleading feature claims, and makes release readiness ambiguous. Unsafe deletion can also remove legitimate offline/research tooling.

### Severity rationale

**P2 retrospective.** This is release/architecture correctness rather than runtime data corruption. It directly affects whether the repository truthfully represents supported capabilities and whether cleanup decisions are safe.

### Existing guarantees violated

- code must have explicit runtime, operational, verification, research, planned, or obsolete ownership;
- compilation/tests alone do not establish product reachability;
- deletion requires importer/reachability evidence rather than naming or line-count heuristics;
- unresolved planned/unintegrated packages must not silently pass final release.

### Considered solutions

1. delete every package not reachable from `cmd/server`;
2. keep every compiling/tested package indefinitely;
3. prove actual import/reachability state, delete only confirmed dead packages, and require explicit dispositions for the rest.

### Chosen remediation and why

`analytics/query` and `analytics/window` were removed only after proving no ordinary/test/external/build-tagged importers and no production reachability. Every other non-runtime analytical package received a mandatory disposition with release rules. Later stages either integrated, removed, or preserved packages according to that disposition.

### Rejected alternatives

Server-only deletion was rejected because offline evaluation, research, ingestion, reconciliation, and materialization roots are legitimate. Keeping all code was rejected because it leaves obsolete and unfinished package trees unowned.

### Trade-offs

The project carries a disposition registry and must update it as packages move into/out of runtime. This governance overhead is accepted in exchange for evidence-based cleanup and truthful release claims.

### Regression tests / protection

Strict project audit fails on unknown non-runtime analytical packages. Deletion installers verify exact file sets, importer classes, build-tagged imports, and runtime reachability before removing code. Final release cannot retain unresolved `planned_production_integration` or `unintegrated_feature_pipeline` entries.

### Adversarial review findings

The review explicitly challenged the simplistic rule "not runtime reachable = dead." That challenge led to separate categories for offline research/evaluation, planned production integration, and genuinely obsolete code.

### Remediation iterations

```text
Stage 14.1: establish factual go-list reachability and review classifications
↓
Stage 14.2: remove only proven obsolete analytics/query and analytics/window
↓
Stage 14.3/14.4/14.10/14.6: resolve Airport Intelligence, Feature Pipeline, Transponder Evidence, and offline Projection Evaluation through real roots or removal
```

### Residual risks / limitations

A disposition can become stale if a package changes ownership without updating the audit policy. Static reachability also does not prove route/job behavioral correctness.

### Operational / deployment consequences

No schema change. Release governance becomes stricter: unresolved planned/unintegrated package trees are blockers rather than passive source inventory.

### Exact evidence

Implementation commit: `8bcc73ad1281d468fc17dc9f0628d54f79d7e2b0` (`refactor: classify and remove obsolete analytical foundations`). Foundation reachability evidence: `fc6c3dbafa302d061653587163457d72f08c7a77`. Historical PR/reviewer details are not invented when unavailable.

### Final canonical status

**CLOSED for the Stage 14.2 unowned/dead-code classification finding.** Later package-specific integration findings have their own canonical documents.

### Prevention / future guard

New packages must either be reachable from a named runtime/verification root or receive an explicit non-runtime disposition with a release decision. Removal must be preceded by importer and reachability proof; neither folder names nor successful compilation are sufficient evidence.
