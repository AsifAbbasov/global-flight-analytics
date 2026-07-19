# Document 52 — Stage 14.12 Projection Read Snapshot Consistency

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: reproducible PostgreSQL input snapshot for one Projection Intelligence result

## 1. Problem

Projection Intelligence previously loaded its production inputs through four
independent database operations:

```text
Current Trajectory
Route
Historical Candidates
Route History
```

All operations used the same analytical `as_of_time`, but they did not share a
PostgreSQL transaction snapshot.

A concurrent ingestion, reconciliation, route materialization, or historical
backfill could therefore become visible between two reads. One projection
result could combine data from different committed database states.

The analytical time boundary prevented future observations from entering the
result, but it did not provide database snapshot consistency.

## 2. Production Decision

One projection request now performs one data-source operation:

```text
LoadSnapshot
```

The PostgreSQL implementation executes all required reads inside one
transaction configured as:

```text
isolation level: REPEATABLE READ
access mode: READ ONLY
```

This guarantees that every query in the projection input load observes one
stable PostgreSQL snapshot even when other transactions commit concurrently.

## 3. Snapshot Contents

The snapshot contains:

```text
current trajectory as of the requested analytical time
latest route result at or before the analytical time
route-scoped historical candidate trajectories
route-frequency history summary
```

A missing materialized route remains a valid analytical condition. The
snapshot contains the current trajectory and no route; the service then builds
the existing auditable unavailable-route result.

A missing route-history summary remains non-fatal and is represented as an
absent optional input.

## 4. Transaction-Scoped Trajectory Repository

Trajectory metadata, trajectory segments, coverage gaps, and flight-state
points must all use the same transaction.

The PostgreSQL trajectory repository now accepts a minimal read client that can
be either:

```text
a pgxpool.Pool
or
a pgx.Tx
```

Production snapshot loading creates a read-only trajectory repository bound to
the same `pgx.Tx` used by route, candidate, history, and point queries.

Write behavior of the existing trajectory repository is unchanged.

## 5. Service Boundary

The Projection Intelligence service no longer coordinates four storage calls.
Its data-source contract exposes only:

```text
LoadSnapshot(context, SnapshotRequest)
```

The service remains responsible for:

```text
request validation
unavailable-route domain semantics
composition policy
result error classification
```

The PostgreSQL adapter remains responsible for atomic data acquisition.

## 6. Transaction Lifecycle

Successful snapshot loading performs:

```text
BEGIN READ ONLY REPEATABLE READ
load all projection inputs
COMMIT
```

Any load error or commit failure triggers rollback cleanup.

No lock is taken on ingestion rows and no write statement is permitted by the
snapshot transaction.

## 7. Preserved Behavior

This increment does not change:

```text
projection formulas
confidence weights
historical-neighbor selection
arrival estimation
route-frequency policy
HTTP response contracts
SQL result ordering
persistence schema
migrations
provider behavior
frontend behavior
```

It changes only the consistency boundary around production reads.

## 8. Regression Gates

Automated tests require:

```text
the DataSource interface to expose one LoadSnapshot operation
Service.Get not to call independent load methods
production PostgreSQL composition to use the repeatable-read executor
transaction options to remain REPEATABLE READ and READ ONLY
successful reads to commit once
failed reads to roll back without commit
commit failures to receive rollback cleanup
snapshot clones not to share mutable slices
```

## 9. Acceptance

The increment is accepted only after:

```text
focused Projection Intelligence tests
PostgreSQL repository tests
snapshot transaction lifecycle tests
architecture regression tests
race detector
strict project architecture audit
complete Go build
go vet
complete Go test suite
frontend dependency security verification
frontend production dependency audit
ESLint
TypeScript validation
Next.js production build
backend Docker image build
git diff check
```
