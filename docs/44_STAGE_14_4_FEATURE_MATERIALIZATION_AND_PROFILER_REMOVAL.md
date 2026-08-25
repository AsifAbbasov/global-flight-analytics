# Document 44 — Stage 14.4 Feature Materialization and Profiler Removal

Status: Remediation History v1.1
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

## 7. Canonical finding decomposition

```text
GFA-REL-037    implemented Feature Pipeline without operational materialization root
GFA-MAINT-038  isolated datasetprofiler facade without a supported consumer
```

## 8. GFA-REL-037 — Feature Pipeline without operational materialization root

### Finding / symptom

The Feature Pipeline had extraction/build/validation/store logic and PostgreSQL verification evidence but no supported executable that materialized real persisted trajectories into `flight_feature_snapshots`.

### Root cause

The analytical pipeline was built and tested as a library before an operational ownership boundary was added. Verification proved internal behavior but not that the product had a real path to run it on repository data.

### Failure scenario

A reviewer sees complete Feature Pipeline packages, tests, and database tables and assumes features are operational. In reality, no production/operational command can select a persisted trajectory, run the pipeline, and store the result. The capability exists only as code/test evidence.

### Impact

This overstates release maturity, leaves a significant package tree without a supported runtime owner, and prevents real operational idempotency/data-path verification.

### Severity rationale

**P2 retrospective.** The issue concerns release/operational correctness and truthfulness of feature availability rather than a known persisted-data corruption event.

### Existing guarantees violated

- tested library code is not a production/operational feature without a real root;
- materialization must operate on persisted evidence, not synthetic fixtures;
- repeated runs for the same evidence boundary should be deterministic/idempotent;
- unresolved `unintegrated_feature_pipeline` dispositions must not survive release.

### Considered solutions

1. remove the complete Feature Pipeline;
2. mark it permanently test-only;
3. add a real operational materializer using existing PostgreSQL and feature contracts.

### Chosen remediation

`materialize-flight-features` selects a real trajectory by UUID or ICAO24, uses an explicit/as-derived `as_of_time`, executes the existing pipeline, persists through the existing feature store, and emits a JSON report. The command is built into the backend container.

### Why selected

The pipeline already had meaningful validated responsibilities and storage ownership, so deletion would discard substantive work. A real materializer closes the reachability gap without inventing a separate service or synthetic data path.

### Rejected alternatives

A one-record synthetic command was rejected because it would satisfy import reachability without proving operational value. A separate microservice/queue was rejected as unnecessary infrastructure for a deterministic command-driven materializer.

### Trade-offs

The backend image gains another operational binary and release support surface. In return, the feature pipeline has a factual runtime owner and can be rerun deterministically.

### Regression tests / protection

Architecture reachability requires all retained Feature Pipeline packages under the `feature_materializer` root. Command/store tests protect selector validation, real trajectory loading, deterministic cutoff behavior, and idempotent storage.

### Adversarial review findings

The review required the command to use persisted trajectories rather than constructing synthetic evidence solely to make packages reachable. This prevents "runtime integration" from degenerating into an import trick.

### Remediation iterations

Stage 14.2 classified the pipeline as unintegrated. Stage 14.4 resolved that classification through a real command and simultaneously removed the one feature package that had no meaningful place in the path.

### Residual risks / limitations

An operational command does not imply scheduled/continuous materialization. Invocation cadence, production orchestration, and data volume remain separate operational concerns.

### Operational / deployment consequences

The backend container includes `/app/materialize-flight-features`. Operators can invoke it against PostgreSQL; no new schema migration is required beyond the already existing feature snapshot schema.

### Exact evidence

Implementation commit: `a1689dc71baa9b2c2b4d66febb30b86436b893c1` (`feat: add operational flight feature materialization`). Historical PR/reviewer metadata is not invented where unavailable.

### Final canonical status

**CLOSED.** The retained Feature Pipeline packages have an operational materialization root.

### Prevention / future guard

A substantial pipeline must have either a supported production/operational/research root or an explicit non-runtime disposition. Verification-only commands must not be used to simulate product integration.

## 9. GFA-MAINT-038 — Isolated `datasetprofiler` facade

### Finding / symptom

`internal/features/datasetprofiler` existed as an isolated in-memory facade with no production, operational, verifier, external test importer, or storage-query integration outside its own package.

### Root cause

An early abstraction was retained after the surrounding feature architecture evolved toward persisted snapshot materialization. No real consumer or bounded cross-snapshot data contract emerged.

### Failure scenario

The package remains indefinitely because it compiles/tests, accumulating maintenance and implying a profiler capability that no supported workflow can invoke meaningfully. A future developer may import it artificially simply to satisfy reachability governance.

### Impact

Dead/isolated code increases cognitive and review surface and makes feature inventory misleading.

### Severity rationale

**P3 retrospective.** This is confirmed maintainability/dead-code debt with no runtime correctness incident.

### Existing guarantees violated

- retained packages require a supported owner or explicit future decision;
- dead code should be removed only after importer/reachability proof;
- artificial reachability must not replace a real product/research contract.

### Considered solutions

1. keep it for possible future use;
2. import it from the new materializer with a trivial one-record path;
3. remove it and require any future profiler to begin from a bounded persisted-data contract.

### Chosen remediation and why

The package was removed. The document records a future-entry condition: a profiler may return only with bounded PostgreSQL reads and an actual operational/research command.

### Rejected alternatives

Speculative retention was rejected because there was no consumer. Artificial importing was rejected because it would game reachability without adding capability.

### Trade-offs

Reintroducing a profiler later requires deliberate implementation rather than reusing the old facade. This is preferable to carrying unowned abstraction debt.

### Regression tests / protection

Reachability/dead-code classification and expected package inventories prevent the removed package from silently returning as an unowned feature tree.

### Adversarial review findings

The review explicitly distinguished the valuable Feature Pipeline from the isolated profiler, avoiding both blanket deletion and blanket retention.

### Remediation iterations

The package was first classified in Stage 14.2 and removed only when the real feature materialization boundary clarified that it had no role.

### Residual risks / limitations

Future profiling needs may be real; the remediation intentionally does not predesign that future contract.

### Operational / deployment consequences

None; removing unused source reduces the build/review surface.

### Exact evidence

Implementation commit: `a1689dc71baa9b2c2b4d66febb30b86436b893c1`.

### Final canonical status

**CLOSED.**

### Prevention / future guard

Do not preserve isolated analytical facades merely because they might be useful later. Require a bounded data source, supported executable/research consumer, and explicit evidence semantics before adding cross-snapshot profiling code.
