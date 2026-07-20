# Document 62 — Stage 14.21 Ingestion Run Terminal Integrity

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: make completed ingestion runs terminal and immutable

## 1. Correctness problem

The previous PostgreSQL repository finalized an ingestion run with an update filtered
only by the run identifier. A second completion call could overwrite a successful run
with failed status, replace the original counters, change the finish time, and replace
the original error evidence.

The table constraints validated status names and non-negative counters but did not
define a lifecycle boundary. Direct SQL could also mutate an already completed run.

## 2. Canonical lifecycle

An ingestion run now has exactly two lifecycle classes:

```text
running
↓ one accepted finalization
success | failed | partial
↓
immutable terminal record
```

A running row must have `finished_at IS NULL`. A terminal row must have a non-null
finish time. A terminal row cannot be changed into another terminal status and its
record counts, finish time, error evidence, source metadata, or other columns cannot
be rewritten.

## 3. Repository transition guard

`IngestionRunRepository.markFinished` updates only a row whose current status is
`running`. The operation returns one explicit outcome:

```text
updated
transition_rejected
not_found
```

`transition_rejected` maps to `ErrIngestionRunTransitionRejected`. It is different
from `ErrIngestionRunNotFound`, allowing callers and diagnostics to distinguish a
missing identifier from a duplicate or conflicting completion attempt.

The update and outcome classification use one PostgreSQL statement. A competing
completion cannot successfully overwrite the winner.

## 4. Database lifecycle constraint

Migration `017_ingestion_run_terminal_integrity.sql` adds
`ingestion_runs_lifecycle_check`:

```text
running  => finished_at is null
terminal => finished_at is not null
```

The migration performs a preflight query. If legacy rows violate that shape, the
migration stops instead of silently fabricating finish timestamps or changing status.

## 5. Terminal immutability trigger

The migration installs `ingestion_runs_terminal_immutability`. The trigger compares
the complete old and new rows. When the old status is terminal, any real change is
rejected with PostgreSQL check-violation code `23514`.

A byte-for-byte no-op update is harmless and remains allowed. A legitimate transition
from `running` to one terminal state remains allowed.

This database boundary protects direct SQL, maintenance commands, future repository
implementations, and accidental code paths that bypass the current repository guard.

## 6. Preserved behavior

This increment preserves:

```text
CreateRunning behavior
MarkSuccess arguments and successful result
MarkFailed arguments and successful result
success, failed, and partial status vocabulary
non-negative counter constraints
started_at <= finished_at constraint
nullable failure message
```

No public HTTP contract changes are introduced.

## 7. Regression protection

Always-running source and migration tests verify:

```text
repository update requires current running status
transition-rejected error is present
single-statement outcome classification is present
lifecycle check constraint is present
terminal trigger is present
terminal mutation uses PostgreSQL code 23514
```

When `TEST_DATABASE_URL` is available, PostgreSQL integration tests additionally
verify:

```text
first finalization succeeds
second conflicting finalization is rejected
original terminal status and counters remain unchanged
direct SQL mutation of a terminal row is rejected
missing run and rejected transition remain distinct
running rows cannot have a finish time
terminal rows cannot omit a finish time
```

## 8. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/repository/postgres/ingestionrun_repository.go \
  internal/repository/postgres/ingestionrun_terminal_integrity_test.go \
  internal/repository/postgres/ingestionrun_terminal_integrity_integration_test.go
go test -count=1 ./internal/repository/postgres
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
git diff --check
git status --short
```

## 9. Completion statement

Stage 14.21 closes Ingestion Run transition integrity. Completion is now a one-way
state transition and the resulting PostgreSQL row is durable terminal evidence rather
than mutable operational state.
