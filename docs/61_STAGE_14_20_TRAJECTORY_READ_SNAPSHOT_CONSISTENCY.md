# Document 61 — Stage 14.20 FlightTrajectory Read Snapshot Consistency

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: keep every aggregate FlightTrajectory read inside one PostgreSQL snapshot

## 1. Correctness problem

A FlightTrajectory is stored as one parent record plus child trajectory segments and
coverage gaps. The previous production repository loaded these parts through separate
pool queries:

```text
read flight_trajectories parent
↓
read trajectory_segments
↓
read coverage_gaps
```

Under PostgreSQL `READ COMMITTED`, another transaction could commit between those
queries. The returned aggregate could therefore combine a parent from one database
state with children from a later state.

Document 52 already established a caller-owned repeatable-read transaction for the
Projection Intelligence workflow. It did not make the core production trajectory
repository safe when used directly by HTTP handlers, route context, ingestion
continuation, or another service.

## 2. Repository-owned snapshot boundary

Both public aggregate read operations now enter the same boundary:

```text
GetLatestTrajectoryByICAO24
GetTrajectoryByID
↓
begin PostgreSQL transaction
isolation = REPEATABLE READ
access mode = READ ONLY
↓
read parent
read segments
read coverage gaps
↓
commit
```

The snapshot is owned by `TrajectoryRepository` whenever it was created from a
PostgreSQL pool. This makes consistency independent of the calling service and prevents
future composition roots from accidentally bypassing the protection.

## 3. Caller-owned transaction compatibility

`NewTrajectoryReadRepository` still supports binding the repository to an existing
`pgx` transaction. A transaction-bound repository has no pool ownership and therefore
reuses the caller snapshot without opening a nested transaction.

This preserves the Projection Intelligence transaction boundary and supports other
workflows that need to combine trajectory reads with additional evidence inside one
larger snapshot.

When `NewTrajectoryReadRepository` receives a pool, it records that pool as the snapshot
owner, so its public aggregate reads receive the same repository-owned protection as
`NewTrajectoryRepository`.

## 4. Failure behavior

The repository returns without a partial aggregate when:

```text
the transaction cannot start
the parent query fails
a child query fails
the caller context is cancelled
the read-only transaction cannot commit
```

An uncommitted transaction is rolled back through an independent bounded cleanup
context. Operation errors preserve their existing domain and repository semantics.

## 5. Concurrency evidence

The PostgreSQL integration test establishes a snapshot, reads one row, commits a second
row through another pooled connection, and reads again through the snapshot repository.
The second read still sees the original row count while a read after transaction commit
sees both rows.

This proves that the repository uses one repeatable-read snapshot rather than merely
issuing sequential queries on the same pool.

## 6. Regression protection

Permanent tests protect:

```text
both public aggregate reads enter withTrajectoryReadSnapshot
pool-backed reads use REPEATABLE READ
pool-backed reads use READ ONLY
parent and child queries receive the transaction-bound repository
transactions commit only after the aggregate is loaded
caller-owned transactions do not create nested transactions
pool-backed NewTrajectoryReadRepository retains snapshot ownership
concurrent commits do not change an active snapshot
```

The integration test is activated when `TEST_DATABASE_URL` is available and otherwise
skips without weakening the static architecture checks.

## 7. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/repository/postgres/trajectory_read_repository.go \
  internal/repository/postgres/trajectory_read_client.go \
  internal/repository/postgres/trajectory_read_snapshot.go \
  internal/repository/postgres/trajectory_read_snapshot_test.go \
  internal/repository/postgres/trajectory_read_snapshot_integration_test.go
go test -count=1 ./internal/repository/postgres
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
git diff --check
git status --short
```

## 8. Completion boundary

This increment closes production FlightTrajectory aggregate snapshot consistency.
It does not close:

```text
Ingestion Run transition integrity
trajectory relational constraints
shared migration filename parser
altitude precision policy
Traffic altitude-status semantics
timestamp and Unix-nanosecond consistency
large PostgreSQL repository decomposition
```
