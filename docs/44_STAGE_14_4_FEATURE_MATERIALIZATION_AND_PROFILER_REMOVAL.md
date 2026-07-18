# Document 44 — Stage 14.4 Feature Materialization and Profiler Removal

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: operational feature materialization, production reachability, and deletion of an unused dataset profiler

## 1. Decision

The Flight Feature Pipeline is retained because it already has validated
extraction contracts, feature-group builders, aircraft enrichment, validation,
idempotent PostgreSQL storage, and the `flight_feature_snapshots` migration.

It is connected through a real operational command instead of being treated
as complete merely because a synthetic verifier passes.

The disconnected `datasetprofiler` package is removed. It had no production,
operational, verifier, test-importer, or storage-query integration outside its
own package. Reaching it artificially from a one-record command would not
create a meaningful product capability.

## 2. Operational Command

```text
materialize-flight-features
```

Exactly one selector is required:

```text
--trajectory-id <UUID>
--icao24 <six hexadecimal characters>
```

Optional evidence cutoff:

```text
--as-of-time <RFC 3339 timestamp>
```

When omitted, the command uses the persisted trajectory end time. Repeated
execution for the same trajectory is therefore deterministic and idempotent.

## 3. Real Data Path

```text
PostgreSQL flight_trajectories, trajectory_segments, coverage_gaps
PostgreSQL aircraft metadata
↓
Feature Extractor Composition
↓
Temporal, Geographical, Operational, Trajectory, Aircraft groups
↓
Feature Validator
↓
PostgreSQL flight_feature_snapshots
↓
JSON materialization report
```

The command never creates a synthetic trajectory.

## 4. Container Boundary

The command is compiled into the backend container as:

```text
/app/materialize-flight-features
```

## 5. Removed Package

```text
internal/features/datasetprofiler
```

A future cross-snapshot profiler must begin with a bounded PostgreSQL read
contract and an actual operational or research command. It must not return as
an isolated in-memory facade.

## 6. Completion Gate

```text
features: total=11 runtime=11 feature_materializer=11
```

No Feature Pipeline package may remain in the non-runtime allowlist.
