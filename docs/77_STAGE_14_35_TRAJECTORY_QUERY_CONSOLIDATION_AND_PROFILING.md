# Document 77 — Stage 14.35 Trajectory Query Consolidation and Profiling

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: consolidate Trajectory read SQL and row mapping, preserve caller context, and prove index eligibility with PostgreSQL execution plans

## 1. Problems closed

The Trajectory read surface still repeated the complete `flight_trajectories` select list and the same eighteen-field row scan in parent and analytical repositories. Segment and coverage-gap reads owned large inline scanners, and several read boundaries silently replaced a missing caller context with `context.Background()`.

The most frequent Trajectory query shapes also lacked permanent plan evidence. The latest-by-ICAO24 query ordered by `end_time`, `start_time`, and `created_at`, while the original index ordered the time columns differently. Analytical end-time reads had no complete order-preserving index.

## 2. Canonical query ownership

`trajectory_read_queries.go` now owns:

```text
flightTrajectorySelectColumns
latestTrajectoryByICAO24Query
trajectoryByIDQuery
trajectoriesByEndTimeQuery
trajectoriesByIDsQuery
trajectorySegmentsByTrajectoryIDQuery
coverageGapsByTrajectoryIDQuery
```

Repository coordinators reference these constants and no longer carry copied SQL column lists.

The identifier-list query uses `unnest(... ) WITH ORDINALITY`, joins the UUID column to a typed UUID value, and preserves caller order without casting the indexed UUID column to text.

## 3. Canonical row mapping

Dedicated scanner files own the database-to-domain mapping:

```text
trajectory_row_scan.go
trajectory_segment_row_scan.go
trajectory_gap_row_scan.go
```

Both single-row and multi-row parent reads use `scanFlightTrajectory`. Segment and coverage-gap read coordinators delegate to their corresponding scanners.

## 4. Caller-owned context

The following read boundaries now use `requireRepositoryContext` and reject a nil caller context:

```text
GetLatestTrajectoryByICAO24
GetTrajectoryByID
withTrajectoryReadSnapshot
ListTrajectoriesByEndTime
ListTrajectoriesByIDs
ListTrajectorySegments
ListCoverageGaps
```

The independent bounded background context remains only in rollback cleanup, where cancellation of the caller must not prevent transaction cleanup.

## 5. Index decisions

Migration `021_trajectory_query_profiles.sql` adds:

```text
flight_trajectories_icao24_latest_idx
    (icao24, end_time DESC, start_time DESC, created_at DESC)

flight_trajectories_end_time_order_idx
    (end_time DESC, start_time DESC, created_at DESC)
```

It also removes the older non-unique `trajectory_segments_trajectory_sequence_idx`. Migration 018 already owns a unique index on the same `(trajectory_id, sequence_number)` key, so retaining both forced every segment write to maintain duplicate structures.

No new coverage-gap index is added. The existing `coverage_gaps_trajectory_time_idx` supports the equality predicate and can be scanned backward for ascending gap time. This decision is verified rather than guessed.

## 6. PostgreSQL execution-plan evidence

`TestTrajectoryQueryProfilesUseExpectedIndexes` runs the production query constants through:

```text
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
```

The permanent profiling gate verifies index eligibility for:

```text
latest trajectory by ICAO24
analytical trajectories by end time
trajectory segments by parent
coverage gaps by parent
```

The test requires both planning and execution timing evidence. `scripts/profile-stage-14-trajectory-queries.sh` provides a focused repeatable entry point.

## 7. Regression protection

Permanent tests protect:

```text
one canonical query owner
one canonical parent scanner
separate segment and gap scanners
absence of copied SQL and inline Scan loops in coordinators
caller-owned context on every Trajectory read boundary
query/index ordering agreement
migration ownership
EXPLAIN ANALYZE index evidence
```

## 8. Acceptance boundary

The safe installer runs all backend gates in a temporary shadow copy before changing the real repository. After application, the unified Stage 14 script runs production migrations, PostgreSQL integration and query profiling, vulnerability analysis, frontend checks, and container checks.

Successful completion is represented by:

```text
STAGE_14_35_TRAJECTORY_QUERY_PROFILING=PASS
STAGE_14_CURRENT_SCOPE_AUDIT=PASS
STAGE_14_OVERALL_STATUS=REOPENED
```

Stage 14 remains reopened only for the independent final closure audit in Document 78.

## 9. Canonical finding decomposition

```text
GFA-MAINT-027  duplicated Trajectory read SQL and row-mapping ownership
GFA-DB-028     Trajectory read caller-context substitution
GFA-PERF-029   query/index ordering mismatch and missing execution-plan evidence
```

## 10. GFA-MAINT-027 — Duplicated Trajectory read SQL and row-mapping ownership

### Finding / symptom

Parent and analytical repositories repeated the same Trajectory select-list shape and eighteen-field row scan, while segment and gap reads retained large inline scanners.

### Root cause

Read surfaces evolved around different use cases before a shared query/mapping owner existed. Schema/semantic additions were therefore copied into multiple SQL/scanner contours.

### Failure scenario

A future field or nullable semantic is added to one query/scanner but omitted from another. Different repository paths then map the same persisted Trajectory shape differently or scan columns in mismatched order.

### Impact

The immediate issue is maintainability and semantic-drift risk. Repeated SQL also makes query/index review harder because one logical query shape can exist in several source locations.

### Severity rationale

**P3 retrospective.** No historical incorrect row mapping is asserted. The problem is duplicated ownership that increases the probability and review cost of future drift.

### Existing guarantees violated

- one persisted row shape should have one canonical column/scanner owner;
- repository coordinators should orchestrate, not duplicate mapping logic;
- query review should operate on production constants rather than copied variants.

### Considered solutions

1. retain duplicate queries/scanners and synchronize via tests;
2. generate all query/mapping code;
3. centralize query constants and package-local scanners while preserving repository methods.

### Chosen remediation and why

`trajectory_read_queries.go` owns SQL shapes and dedicated scanner files own parent, segment, and gap mapping. Existing repositories delegate to them without new public abstractions.

### Rejected alternatives

Tests alone would detect some drift but retain duplicate ownership. Code generation was unnecessary for the current stable shapes and would add tooling/maintenance overhead.

### Trade-offs

Queries and scanners become shared dependencies inside the package. That coupling is intentional because they represent one canonical storage contract.

### Regression tests

Source tests forbid copied select lists and inline scan loops in coordinators and require the canonical owners.

### Adversarial review and remediation iterations

This work follows the write-side decomposition in Documents 69 and 73. The later read consolidation avoids a one-sided architecture where writes have clear responsibility boundaries but reads still duplicate storage contracts.

### Residual risk / limitations

Centralized files can themselves grow. Future decomposition should be driven by distinct storage shapes or semantic responsibilities, not file-count targets.

### Operational / deployment consequences

None by itself; public repository and HTTP contracts are unchanged.

### Exact evidence and final status

Implementation commit: `f414f6638f8ba5fbe61321e55a21ff3ac91a4986` (`refactor: consolidate trajectory queries and profile indexes`). Historical PR/reviewer metadata is not invented where unavailable. **Canonical finding status: CLOSED.**

## 11. GFA-DB-028 — Trajectory read caller-context substitution

### Finding / symptom

Several Trajectory read boundaries silently replaced a missing caller context with `context.Background()`.

### Root cause

Stage 14.33 hardened selected repository/write paths, but Trajectory analytical/read helpers retained earlier tolerant context behavior and were not yet covered by the same ownership rule.

### Failure scenario

A caller passes nil to a Trajectory read. The repository proceeds without caller cancellation or deadline ownership, allowing a database read/snapshot to outlive the request/job that initiated it.

### Impact

Detached reads can consume PostgreSQL connections and snapshot resources beyond the intended lifecycle and make timeout/shutdown behavior inconsistent across repository methods.

### Severity rationale

**P2 retrospective.** This is a production lifecycle/resource-control defect without evidence of silent data corruption.

### Existing guarantees violated

- all database-reaching reads must preserve caller lifecycle intent;
- context policy must be consistent across repository read/write surfaces;
- cleanup exceptions must not leak into normal work.

### Considered solutions

The same alternatives as GFA-DB-021 were considered: silent fallback, repository-invented timeout, or explicit required caller context.

### Chosen remediation and why

All listed Trajectory read boundaries call `requireRepositoryContext`. Only the documented rollback cleanup path may create an independent bounded context.

### Rejected alternatives

Silent fallback and repository-owned timeout were rejected because they continue to hide caller intent.

### Trade-offs

Misconfigured callers now fail immediately. This tightens an internal contract consistently with Stage 14.33.

### Regression tests

Source/audit tests require caller context on every Trajectory read boundary and forbid background-context substitution.

### Adversarial review and remediation iterations

This is a genuine continuation of the earlier context audit, showing the first pass was scoped rather than universal. Document 79 later finds the same residual pattern in the migrator and further broadens the guard.

### Residual risk / limitations

Context correctness still requires package-wide/global audits for new code. A local list of protected methods can become stale if new database-reaching entry points are added without architecture updates.

### Operational / deployment consequences

No schema change. Detached read behavior becomes a visible caller error.

### Exact evidence and final status

Implementation commit: `f414f6638f8ba5fbe61321e55a21ff3ac91a4986`. **Canonical finding status: CLOSED for the Trajectory read boundaries named here.**

## 12. GFA-PERF-029 — Query/index ordering mismatch and missing execution-plan evidence

### Finding / symptom

Hot Trajectory query ordering did not fully match the available index order, the analytical end-time query lacked a complete order-preserving index, and the repository had no permanent production-query `EXPLAIN ANALYZE` gate proving index eligibility.

### Root cause

Indexes had been added incrementally around earlier query shapes. Query evolution changed sort requirements without a corresponding evidence-backed index review. A redundant segment index also remained after migration 018 introduced a unique index on the same key.

### Failure scenario

As Trajectory volume grows, PostgreSQL may perform extra sorting or scan more data for latest/end-time queries. Duplicate indexes force additional write maintenance. Without permanent plan evidence, a superficially plausible index change can be merged without proving the intended production query can use it.

### Impact

The risk is avoidable read latency, write amplification, and speculative index proliferation. It affects scalability rather than current analytical semantics.

### Severity rationale

**P2 retrospective.** The problem concerns production query efficiency and evidence quality on an important historical/trajectory path, but no measured production outage or latency SLO breach is claimed.

### Existing guarantees violated

- indexes should be shaped from actual production query predicates/order;
- redundant indexes should not impose write cost without read value;
- performance changes require executable plan evidence rather than assumption.

### Considered solutions

1. add several speculative indexes;
2. keep existing indexes and rely on planner heuristics;
3. align only the proven hot query indexes, remove the duplicate segment index, and add permanent `EXPLAIN (ANALYZE, BUFFERS)` tests.

### Chosen remediation and why

Migration 021 adds exactly the latest-by-ICAO24 and end-time order indexes, removes the duplicate segment index, and deliberately adds no coverage-gap index because the existing one is proven sufficient. Production query constants are profiled directly.

### Rejected alternatives

Speculative indexing was rejected because every index has write/storage cost. Retaining a duplicate segment index was rejected because the unique index already owns the same key order. Adding a coverage-gap index was rejected after plan evidence showed the existing index is usable.

### Trade-offs

`EXPLAIN ANALYZE` integration tests require a real PostgreSQL environment and representative fixture shape, making the gate heavier than a source-level assertion. The cost is accepted because planner behavior cannot be proved by string inspection alone.

### Regression tests

`TestTrajectoryQueryProfilesUseExpectedIndexes`, migration ownership checks, query/index order tests, and the focused profiling script protect the decisions.

### Adversarial review and remediation iterations

This stage explicitly adopts the simplification principle later formalized in GFA engineering guidance: measure query behavior, remove redundant work, add only evidence-backed indexes, and avoid infrastructure/cache responses to unproven bottlenecks.

### Residual risk / limitations

An index being eligible in integration tests does not guarantee the same planner choice at every production cardinality/statistics state. Production telemetry and periodic plan review remain necessary as data volume grows.

### Operational / deployment consequences

Migration 021 adds two indexes and removes one redundant index. Deployments must account for index-build cost on larger future datasets; at the historical project scale this was part of the accepted migration path.

### Exact evidence and final status

Implementation commit: `f414f6638f8ba5fbe61321e55a21ff3ac91a4986`. **Canonical finding status: CLOSED for the recorded Stage 14 query/index backlog.**

## 13. Prevention / future guard

Trajectory read changes must update one canonical SQL/scanner owner, require caller contexts on database work, and justify index changes with the actual production query plus PostgreSQL plan evidence. Redundant indexes and speculative caches/infrastructure are not acceptable substitutes for measured query work.
