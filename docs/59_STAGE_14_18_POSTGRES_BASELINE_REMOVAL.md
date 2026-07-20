# Document 59 — Stage 14.18 PostgreSQL Baseline Removal

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: remove the ability to manufacture migration history without executing migration SQL

## 1. Correctness problem

The former migration command exposed a `baseline` mode that inserted every local
migration into `schema_migrations` without executing its SQL. The operation did
not prove that the existing PostgreSQL schema contained the tables, columns,
constraints, indexes, functions, and data transformations represented by those
migration files.

A successful baseline could therefore create false evidence:

```text
schema_migrations says applied
↓
required schema object may be absent or incompatible
↓
normal migration execution skips the missing change
↓
runtime failure appears later and outside the migration boundary
```

Transactional history writes and advisory locking cannot make that assertion
correct. They only make the unsupported assertion atomic and serialized.

## 2. Decision

Global Flight Analytics does not currently need to adopt an unmanaged legacy
schema. The repository already owns the full migration sequence, clean database
bootstrap is supported, and deployed databases already maintain migration
history.

The safe decision is removal rather than a larger schema-introspection framework.
The following surfaces are removed:

```text
Runner.Baseline
migrate --baseline
baseline-specific logging and history insertion
```

## 3. Preserved behavior

The change does not modify:

```text
migrate --status
normal pending migration execution
schema_migrations table shape
existing schema_migrations rows
migration checksums
per-migration PostgreSQL transactions
PostgreSQL advisory lock serialization
Docker and deployment migration startup
```

No database migration is required because this increment removes an operational
code path; it does not mutate the schema.

## 4. Operational recovery rule

If a future database is discovered with application tables but without trustworthy
migration history, operators must not fabricate history. Recovery requires an
explicit, separately reviewed reconciliation procedure that compares the real
schema and data state with the repository migration sequence.

Until such a procedure exists, the supported recovery paths are:

```text
restore a database backup with trustworthy migration history
or
create a clean database and apply the repository migrations normally
```

## 5. Regression protection

Source-level regression tests fail if either the command or runner reintroduces
the removed baseline surface. The tests protect against:

```text
registration of a baseline command-line flag
calls to Runner.Baseline
reintroduction of the Runner.Baseline method
baseline-specific migration history insertion
```

## 6. Acceptance commands

From `apps/api`:

```bash
gofmt -w cmd/migrate/main.go cmd/migrate/baseline_removal_test.go internal/database/migrator/baseline_removal_test.go
go test -count=1 ./cmd/migrate ./internal/database/migrator
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
git diff --check
git status --short
```

## 7. Completion statement

The unsafe PostgreSQL migration baseline capability is removed. Migration history
can now be created only by the normal migration execution path, where schema SQL
and its history record commit atomically under the migration advisory lock.
