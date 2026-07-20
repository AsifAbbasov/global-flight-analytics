# Document 63 — Stage 14.22 Trajectory Relational Integrity

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: enforce complete PostgreSQL integrity for FlightTrajectory aggregates

## 1. Correctness problem

A FlightTrajectory is stored across three tables:

```text
flight_trajectories
trajectory_segments
coverage_gaps
```

Before this increment, child `trajectory_id` values could be null, segment sequence
numbers were not unique per trajectory, a coverage gap could reference a segment
owned by another trajectory, and parent counters were not required to match the
actual child rows.

Those gaps allowed a database state that could not represent one coherent domain
aggregate even when each individual row was syntactically valid.

## 2. Repository fail-fast boundary

`SaveTrajectory` now validates the aggregate before opening a PostgreSQL
transaction. It rejects:

```text
segment_count different from len(segments)
coverage_gap_count different from len(coverage_gaps)
point_count different from the sum of segment point_count values
point_count different from an available in-memory point collection
non-contiguous segment sequence numbers
negative stored counts
segment or coverage-gap ICAO24 different from the parent ICAO24
```

This is an early diagnostic boundary. PostgreSQL remains the authoritative
integrity boundary.

## 3. Child parent identity

Migration 018 makes `trajectory_id` mandatory for both child tables.

```text
trajectory_segments.trajectory_id → NOT NULL
coverage_gaps.trajectory_id → NOT NULL
```

Deleting a parent continues to use the existing cascade behavior. A child can no
longer exist as an unowned trajectory fragment.

## 4. Segment ordering

The database now requires:

```text
sequence_number > 0
UNIQUE (trajectory_id, sequence_number)
```

The deferred aggregate verifier additionally requires the final sequence to be
contiguous from `1` through the actual segment count.

## 5. Coverage-gap segment ownership

Each previous or next segment reference is protected by a composite foreign key:

```text
(trajectory_id, previous_segment_id)
    → trajectory_segments(trajectory_id, id)

(trajectory_id, next_segment_id)
    → trajectory_segments(trajectory_id, id)
```

A segment from another FlightTrajectory can no longer be attached to the gap even
when its standalone identifier exists.

## 6. Deferred aggregate verification

Constraint triggers run at transaction completion so the repository may continue
to insert the parent first and its children afterward.

At the final database state, PostgreSQL verifies:

```text
stored segment_count = actual segment rows
stored point_count = sum of segment point_count
stored coverage_gap_count = actual coverage-gap rows
segment sequence numbers are contiguous
all segment ICAO24 values match the parent
all coverage-gap ICAO24 values match the parent
```

The same protection applies to direct SQL and to future repository code.

## 7. Legacy-data policy

Migration 018 performs a complete preflight before adding constraints. It aborts
without silently rewriting data when it finds:

```text
unowned child rows
duplicate or non-positive sequence numbers
cross-trajectory gap references
stored-count mismatches
sequence gaps
child identity mismatches
```

Repairing such rows requires an explicit evidence-backed migration rather than an
automatic guess.

## 8. Regression protection

The permanent tests cover:

```text
canonical in-memory aggregate acceptance
stored-count mismatch rejection
sequence-gap rejection
child identity rejection
point-total rejection
validation before transaction creation
PostgreSQL NOT NULL enforcement
per-trajectory sequence uniqueness
same-trajectory gap references
stored counter and contiguous sequence verification
```

PostgreSQL integration tests run when `TEST_DATABASE_URL` is configured. They skip
without falsely reporting runtime database evidence when it is absent.

## 9. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/repository/postgres/trajectory_repository.go \
  internal/repository/postgres/trajectory_relational_integrity.go \
  internal/repository/postgres/trajectory_relational_integrity_test.go \
  internal/repository/postgres/trajectory_relational_integrity_integration_test.go
go test -count=1 ./internal/repository/postgres
go test -count=1 ./internal/database/migrationaudit ./internal/database/migrator
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
git diff --check
git status --short
```

## 10. Completion statement

This increment closes the known FlightTrajectory relational-integrity debt for
child ownership, segment ordering, same-trajectory gap references, child identity,
and parent stored counters.
