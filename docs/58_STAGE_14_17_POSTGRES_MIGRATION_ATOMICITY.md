# Document 58 — Stage 14.17 PostgreSQL Migration Atomicity

Status: Reopened Amendment v1.5
Project: Global Flight Analytics
Scope: make schema mutation and migration history recording one atomic operation

## 1. Correctness problem

The previous migration runner executed a migration SQL file and then inserted its
history record through a separate database operation. A failure between those
operations could leave the schema changed while the migration remained pending.
Concurrent migration processes could also observe and attempt the same pending
migration.

## 2. Atomic migration boundary

Each pending migration is now applied through one PostgreSQL transaction:

```text
begin transaction
↓
execute migration body
↓
insert schema_migrations record
↓
commit transaction
```

If migration execution, history insertion, context handling, or commit fails,
the transaction is rolled back. A schema change cannot be committed without its
matching history record.

## 3. Existing migration envelopes

Current SQL files contain outer `BEGIN` and `COMMIT` statements. The runner
removes exactly that outer envelope before executing the body inside its own
transaction. Incomplete envelopes and nested transaction-control statements are
rejected.

The runner owns the transaction boundary. Individual migration files own only
the schema statements inside that boundary.

## 4. Interprocess serialization

`ApplyPending` acquires a PostgreSQL session advisory lock through a dedicated
pooled connection. A second compliant migration process waits instead of
evaluating and applying the same pending migration concurrently.

The lock is released on the same PostgreSQL connection with a bounded independent
context, including when the caller context has already been cancelled.

## 5. Baseline supersession

The original Document 58 implementation also serialized `Baseline` and recorded
its history writes transactionally. That reduced concurrency risk but did not
prove that the existing database schema matched the migrations being marked as
applied.

Document 59 removes the baseline operation from both the runner and command-line
interface. Migration history can no longer be manufactured without executing the
corresponding SQL.

## 6. Regression protection

The migrator package tests protect:

```text
outer transaction-envelope removal
unwrapped migration support
incomplete-envelope rejection
nested transaction-control rejection
transactional ordering of migration SQL and history insertion
advisory lock acquisition and release
ApplyPending lock usage
```

## 7. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/database/migrator/runner.go internal/database/migrator/runner_atomicity_test.go
go test -count=1 ./internal/database/migrator
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
git diff --check
git status --short
```

## 8. PostgreSQL debt status correction

The former closure statement was incorrect. The repository still contained duplicate
migration version `016`, so the production migrator could not list or apply the complete
catalog. Document 71 reopens Stage 14, assigns Data Quality Parent Integrity to version
`019`, and adds a real production-migrator gate.

Migration atomicity, advisory locking, and baseline removal remain valid. They do not by
themselves prove catalog deployability or closure of the remaining PostgreSQL correctness
and maintainability findings.

## 9. Current verification boundary

The current cross-stack script runs production `cmd/migrate` twice against a clean
PostgreSQL database and verifies that every SQL file has exactly one
`schema_migrations` row. A successful current-scope marker is evidence for the checks
actually executed, not proof that the full Stage 14 debt register is closed.

The authoritative status is:

```text
STAGE_14_CURRENT_SCOPE_AUDIT=PASS
STAGE_14_OVERALL_STATUS=REOPENED
```
