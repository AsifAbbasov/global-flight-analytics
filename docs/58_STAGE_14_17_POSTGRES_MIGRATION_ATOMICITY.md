# Document 58 — Stage 14.17 PostgreSQL Migration Atomicity

Status: Reopened Amendment v1.6
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

The authoritative status at the time of this amendment was:

```text
STAGE_14_CURRENT_SCOPE_AUDIT=PASS
STAGE_14_OVERALL_STATUS=REOPENED
```

## 10. Stage 14.30 current-scope closure

Migration 020 closes the confirmed Ingestion Run evidence, Route/Historical timestamp
mirror, and cancelled-context repository rollback correctness gaps. This does not close
the remaining maintainability and Clean Code register; Stage 14 remained reopened at
that historical point.

## 11. Finding history and failure scenario

### Finding

Migration execution and migration-history persistence were two independent durable
operations, and concurrent runners were not serialized around pending-migration
evaluation.

### Root cause

The runner treated `execute migration SQL` and `record migration version` as sequential
application steps instead of one database invariant. Concurrency ownership was also left
to process timing rather than PostgreSQL.

### Failure scenario

```text
runner reads migration N as pending
↓
runner executes schema SQL for N
↓
process exits before schema_migrations insert
↓
schema changed, migration N still appears pending
↓
next startup may attempt N again against an already-mutated schema
```

A second scenario was two compliant processes evaluating the same pending migration at
the same time before either had recorded completion.

### Impact

The defect could make migration history cease to be trustworthy, cause duplicate or
partially repeated schema changes, and move failure from the migration boundary into
later application startup or runtime behavior.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1
correctness/integrity** because the defect could invalidate the database evolution
source of truth and prevent reliable deployment recovery.

### Existing guarantees violated

```text
migration history reflects committed schema state
one migration is applied once by compliant runners
failed migration work does not become partially durable
schema evolution remains reproducible from repository migrations
```

## 12. Considered and rejected alternatives

### Keep separate SQL and history writes with compensating cleanup

Rejected because compensation after a process crash is not atomic and cannot prove which
schema statements became durable before the failure.

### Add only a process-local mutex

Rejected because it protects one process only and does not serialize multiple application
instances or deployment jobs.

### Move migrations to a new external service

Rejected because no measured need justified another service. PostgreSQL already provides
the transaction and advisory-lock primitives needed to enforce the invariant.

### Chosen remediation

Use one PostgreSQL transaction for migration SQL plus history insertion, and use a
PostgreSQL advisory lock to serialize compliant migration runners.

## 13. Why this solution and trade-offs

The database owns the schema state and the history table, so the invariant is enforced at
the narrowest durable boundary instead of in orchestration code.

Trade-offs:

```text
+ schema/history atomicity
+ cross-process serialization
+ deterministic failure rollback
- migration runner must own transaction-envelope parsing
- a competing compliant runner may wait on the advisory lock
- future SQL that cannot run inside the transaction requires explicit design review
```

The waiting cost is acceptable because migrations are deployment-time operations rather
than a latency-sensitive request path.

## 14. Adversarial review and remediation iterations

### Iteration 1 — atomic migration execution

Implementation commit `07c0907eb4b739ca2ba12259600df537254a1075`
(`fix: make postgres migrations atomic`) coupled schema SQL and history insertion.

### Iteration 2 — serialization and envelope hardening

The final runner additionally protects concurrent compliant processes with a PostgreSQL
advisory lock and rejects unsupported/incomplete transaction envelopes.

### Iteration 3 — challenge to the meaning of migration history

Adversarial review found that making the former `Baseline` operation atomic did not make
its assertion truthful: it could still record migrations that had never executed.
Document 59 therefore removed baseline entirely.

### Iteration 4 — challenge to catalog deployability

Later review found duplicate migration version `016`. Stage 14 was explicitly reopened
instead of treating the atomic runner as proof that the full catalog was deployable.
This is preserved as evidence that closure status was corrected when a broader failure
scenario was discovered.

## 15. Residual risks and limitations

This finding does not protect against:

```text
manual out-of-band database changes
future migration statements incompatible with the runner-owned transaction model
corrupt or duplicated migration catalogs not detected by catalog validation
non-compliant external tools that bypass the repository migration runner
```

The duplicate-version/catalog class is handled by later migration-catalog validation and
must not be conflated with this atomicity finding.

## 16. Operational consequences

Deployment migration processes are serialized. Operators must treat advisory-lock waits
as expected coordination rather than bypassing the lock. Recovery from an untrustworthy
legacy schema must use an explicit reconciliation path rather than fabricated history.

## 17. Exact evidence

```text
implementation commit:
07c0907eb4b739ca2ba12259600df537254a1075

related remediation:
Document 59 — unsafe migration baseline removal

permanent regression coverage:
internal/database/migrator/runner_atomicity_test.go
migration-envelope tests
advisory-lock acquisition/release tests
production migration catalog verification in later Stage 14 gates
```

A historical pull-request number for the original direct remediation is not asserted
here because it is not preserved in the currently searchable evidence. The commit SHA is
the canonical implementation evidence.

## 18. Final canonical status

```text
FINDING_GFA_DB_001_MIGRATION_ATOMICITY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/58_STAGE_14_17_POSTGRES_MIGRATION_ATOMICITY.md
IMPLEMENTATION_COMMIT=07c0907eb4b739ca2ba12259600df537254a1075
```

`CLOSED` applies to the migration-atomicity finding only. Historical statements that
Stage 14 as a whole was reopened remain valid for their point in time and are not
silently rewritten into broader closure claims.
