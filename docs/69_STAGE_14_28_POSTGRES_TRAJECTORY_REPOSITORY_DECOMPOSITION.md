# Document 69 — Stage 14.28 PostgreSQL Trajectory Repository Decomposition

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: separate the Trajectory Repository by persistence responsibility while preserving one public repository contract

## 1. Confirmed maintainability problem

The production Trajectory Repository was already behaviorally correct, but two
source files had accumulated independent responsibilities:

```text
trajectory_repository.go: 548 lines
trajectory_read_repository.go: 459 lines
```

The write source mixed repository construction, public write operations,
transaction orchestration, reconciliation attempt ownership, parent inserts,
segment inserts, coverage-gap inserts, and segment-reference inference.

The read source mixed snapshot coordination, parent selection, parent scanning,
aggregate child loading, segment reading, coverage-gap reading, and lookup
normalization.

This was a cohesion finding, not a line-count-only refactor.

## 2. Preserved public contract

The type `TrajectoryRepository` remains the production entry point. The following
public methods retain their names and signatures:

```text
SaveTrajectory
SaveReconciledTrajectory
GetLatestTrajectoryByICAO24
GetTrajectoryByID
ListTrajectorySegments
ListCoverageGaps
```

Callers, interfaces, composition roots, HTTP handlers, reconciliation workers,
feature materialization, and Projection Intelligence require no changes.

## 3. Write-side ownership

Write responsibilities are now separated into:

```text
trajectory_repository.go
  repository state and constructor only

trajectory_write_repository.go
  public write methods and transaction coordination

trajectory_reconciliation_write.go
  reconciliation attempt ownership and replacement cleanup

trajectory_parent_write.go
  direct and reconciled parent trajectory inserts

trajectory_segment_write.go
  ordered trajectory-segment persistence

trajectory_gap_write.go
  coverage-gap persistence and inferred segment references
```

Relational integrity and flight identity validation still execute before the
PostgreSQL transaction begins. Reconciled writes still verify the active attempt
under row lock before replacing the previous derived trajectory.

## 4. Read-side ownership

Read responsibilities are now separated into:

```text
trajectory_read_repository.go
  public repeatable-read snapshot coordination only

trajectory_parent_read.go
  parent selection, parent scanning, and ICAO24 normalization

trajectory_child_read.go
  aggregate child-loading coordination

trajectory_segment_read.go
  ordered segment query and mapping

trajectory_gap_read.go
  ordered coverage-gap query and mapping
```

Both public aggregate reads still use the repository-owned read-only
`REPEATABLE READ` boundary. A transaction-bound repository still reuses its
caller-owned snapshot and does not create a nested transaction.

## 5. Deliberately unchanged behavior

This increment does not change:

```text
SQL statements
transaction isolation or access mode
commit and rollback behavior
reconciliation ownership rules
relational integrity rules
trajectory identity rules
row ordering
nullable mapping
not-found semantics
public interfaces
HTTP contracts
PostgreSQL schema or migrations
```

## 6. Regression protection

Permanent tests require:

```text
constructor-only trajectory_repository.go
snapshot-coordinator-only trajectory_read_repository.go
dedicated owners for write orchestration, reconciliation, parents, segments, and gaps
dedicated owners for parent reads, child loading, segment reads, and gap reads
both public aggregate reads to retain withTrajectoryReadSnapshot
relational validation to remain before BeginTx
former monolithic SQL and mapping responsibilities not to return to coordinator files
```

The existing snapshot, reconciliation, relational-integrity, repository, and
PostgreSQL integration tests remain the behavioral evidence.

## 7. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/repository/postgres/trajectory_*.go
go test -count=1 ./internal/repository/postgres
go test -count=1 ./...
go vet ./...
```

From `apps/web`:

```bash
pnpm typecheck
pnpm lint
pnpm build
```

From the repository root:

```bash
git diff --check
git status --short
```

## 8. Completion boundary

This increment closes the final known PostgreSQL maintainability debt recorded
in Document 58. Stage 14 still requires one final repository-wide audit before
its correction programme can be declared complete.
