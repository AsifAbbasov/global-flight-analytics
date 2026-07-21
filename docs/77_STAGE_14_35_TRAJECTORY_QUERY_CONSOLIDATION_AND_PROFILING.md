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
