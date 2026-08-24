# Document 59 — Stage 14.18 PostgreSQL Baseline Removal

Status: Remediation History v1.1
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

## 8. Finding history and root cause

### Finding

The migration CLI could mark all repository migrations as applied without proving that
any corresponding schema or data transformation existed in the target database.

### Root cause

The baseline operation treated migration history as an administrative bookkeeping table
rather than evidence derived from executed migration SQL. Atomicity and locking solved
how the write occurred, but not whether the claim being written was true.

### Failure scenario

```text
legacy or partially-created database exists
↓
operator runs migrate --baseline
↓
all local versions are recorded as applied
↓
one or more required constraints/tables/transforms are actually absent
↓
normal migrator now skips those versions
↓
runtime fails later against an unproven schema
```

### Impact

The defect could permanently desynchronize migration history from the real schema and
convert an explicit deployment-time problem into a delayed runtime correctness failure.
It also undermined the ability to reconstruct or audit database state from repository
migrations.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1
correctness/deployment integrity** because the operation could create authoritative but
false migration evidence and prevent normal remediation from running.

### Existing guarantees violated

```text
schema_migrations is evidence of executed repository migrations
clean bootstrap and deployed history are reproducible
operators cannot assert schema state without verification
migration history is safer to fail closed than to fabricate
```

## 9. Considered and rejected alternatives

### Keep baseline but add a warning/confirmation prompt

Rejected because a stronger warning does not prove schema equivalence and still permits
the false state to be committed.

### Keep baseline and rely on transaction/advisory-lock protection

Rejected because atomic false evidence is still false evidence. This was the adversarial
review finding that followed the initial migration atomicity hardening.

### Build full schema introspection and migration reconciliation immediately

Rejected because the project had no current requirement to adopt unmanaged legacy
schemas, and such a framework would add substantial complexity without a present user or
operational need.

### Chosen remediation

Remove the unsafe baseline capability. Require normal migration execution for supported
databases and require a future explicit reconciliation design if legacy adoption ever
becomes a real requirement.

## 10. Why this solution and trade-offs

The chosen solution follows fail-closed and simplification-first principles: delete an
unsafe capability instead of surrounding it with a larger framework.

Trade-offs:

```text
+ migration history remains tied to executed SQL
+ less migration code and fewer dangerous operator modes
+ no hidden schema-introspection heuristic
- convenient adoption of an unmanaged legacy database is no longer supported
- recovery from missing/untrusted history may require backup restore or clean rebuild
- a future legitimate legacy-adoption need requires a separately reviewed reconciliation tool
```

The loss of convenience is intentional because correctness evidence has higher value than
a shortcut that cannot establish equivalence.

## 11. Adversarial review and remediation iterations

### Iteration 1 — migration atomicity

Document 58 made migration/history writes atomic and serialized. That initially also
made baseline history writes safer mechanically.

### Iteration 2 — semantic challenge

Review then challenged the stronger question: **what proves that a baseline history row
corresponds to real schema state?** The answer was nothing. Transactional execution could
not repair that semantic defect.

### Iteration 3 — removal

Implementation commit `3dafcf8ad08a8a4b270456cc2a023e8f4d0ffd53`
(`fix: remove unsafe migration baseline`) deleted the unsupported capability rather than
adding a speculative reconciliation subsystem.

## 12. Residual risks and limitations

The remediation does not automatically repair databases whose migration history was
already fabricated or lost. It also does not detect manual out-of-band DDL by itself.
Those states require explicit operator investigation or later catalog/schema verification.

The project intentionally has no generic legacy-schema adoption workflow today.

## 13. Operational and deployment consequences

Operators may use normal pending migration execution and status inspection only. A
database with application tables but untrusted history must not be forced into a
`current` state by metadata writes. Supported recovery remains restore-with-history or
clean bootstrap until a dedicated reconciliation procedure is designed.

## 14. Exact evidence

```text
implementation commit:
3dafcf8ad08a8a4b270456cc2a023e8f4d0ffd53

related design history:
Document 58 — migration atomicity and baseline supersession

permanent regression coverage:
cmd/migrate/baseline_removal_test.go
internal/database/migrator/baseline_removal_test.go
```

A historical pull-request number is not asserted because it is not preserved in the
currently searchable evidence. The implementation commit and regression tests are the
canonical evidence for this remediation.

## 15. Final canonical status

```text
FINDING_GFA_DB_002_UNSAFE_MIGRATION_BASELINE=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/59_STAGE_14_18_POSTGRES_BASELINE_REMOVAL.md
IMPLEMENTATION_COMMIT=3dafcf8ad08a8a4b270456cc2a023e8f4d0ffd53
```
