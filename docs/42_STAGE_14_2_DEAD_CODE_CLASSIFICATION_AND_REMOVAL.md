# Document 42 — Stage 14.2 Dead Code Classification and Removal

Status: Implementation Baseline v1.0
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
