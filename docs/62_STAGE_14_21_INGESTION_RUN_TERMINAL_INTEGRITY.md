# Document 62 — Stage 14.21 Ingestion Run Terminal Integrity

Status: Remediation History v1.3
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

## 10. Final integration-fixture parity

The unified Stage 14 PostgreSQL run exposed an outdated direct-SQL fixture in
`reconciliation_result_identity_integration_test.go`. It inserted a terminal `success`
ingestion run without `finished_at`, which correctly violated
`ingestion_runs_lifecycle_check`.

The fixture now supplies a non-null finish time equal to its start time. The final source
audit scans every PostgreSQL integration test for direct terminal `ingestion_runs`
inserts and rejects any statement that omits `finished_at`. Production lifecycle
semantics are unchanged; the fixture now obeys the contract already defined here.

The same final audit distinguishes complete `FlightStateRepository` fixtures from
purpose-built minimal schemas used only by migration, metric, or traffic-query tests.
Only fixtures that instantiate the complete repository must reproduce every current
Flight State evidence column. This avoids both stale repository fixtures and false
positives against intentionally narrow test tables.

## 11. Stage 14.30 evidence amendment

Document 72 strengthens the terminal lifecycle with processed-count and error-message
status constraints. Repository finalization now rejects impossible counters and invalid
error evidence before SQL, while migration 020 enforces the same contract for direct
PostgreSQL writes. `partial` and `failed` require a non-empty explanation; `running`
and `success` require a null error message.

## 12. Finding history, root cause, and failure scenario

### Finding

A completed ingestion run was still mutable operational state. Repository finalization
updated by identifier rather than by an explicit `running -> terminal` lifecycle guard,
and PostgreSQL did not independently protect terminal rows.

### Root cause

Status values existed, but the system had no canonical state-machine invariant at either
the repository transition or database boundary. As a result, `success`, `failed`, and
`partial` were labels rather than terminal evidence states.

### Failure scenario

```text
run R is finalized success with counters/evidence A
↓
a retry, duplicate worker, or competing caller finalizes R again
↓
second update writes failed/partial with counters/evidence B
↓
historical ingestion evidence now describes the retry instead of the accepted completion
```

Direct SQL could produce the same class of mutation without going through the repository.

### Impact

The defect could corrupt operational history, freshness/accounting evidence, data-quality
analysis, reconciliation reasoning, and any audit that trusts completed ingestion runs as
immutable evidence.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1
evidence/data integrity** because accepted ingestion results could be rewritten after
completion without an obvious failure.

### Existing guarantees violated

```text
one ingestion run has one accepted terminal outcome
completed evidence is immutable
missing identifiers are distinguishable from rejected transitions
direct SQL cannot bypass lifecycle integrity
running and terminal timestamps match lifecycle state
```

## 13. Considered and rejected alternatives

### Rely on idempotent callers

Rejected because duplicate or competing completion is exactly the failure mode. A durable
invariant cannot depend on every current and future caller behaving perfectly.

### Add only `WHERE status = 'running'` in repository SQL

Rejected as incomplete because direct SQL, maintenance code, or future repositories could
still mutate a terminal row.

### Make the whole table append-only

Rejected because the legitimate `running -> terminal` transition requires one controlled
update and would otherwise force a more complex event-store model not justified by the
scope.

### Chosen remediation

Use a single-statement repository transition guard plus PostgreSQL lifecycle constraints
and terminal-row immutability trigger.

## 14. Why this solution and trade-offs

The application provides typed transition outcomes while PostgreSQL protects the durable
truth against all writers.

Trade-offs:

```text
+ deterministic first-writer-wins finalization
+ immutable terminal evidence
+ direct-SQL protection
+ explicit duplicate/conflict diagnostics
- trigger and lifecycle constraints add schema complexity
- fixture/direct-SQL writers must obey the lifecycle contract
- corrective edits to terminal evidence require explicit repair tooling rather than ad hoc UPDATE
```

The stricter repair cost is intentional because silent mutation of historical evidence is
more dangerous than requiring a reviewed repair.

## 15. Adversarial review and remediation iterations

### Iteration 1 — terminal lifecycle enforcement

Implementation commit `b3603311d86f23c66bc945c8a61471142ccbec63`
(`fix: enforce ingestion run terminal integrity`) introduced repository transition
classification, lifecycle checks, and terminal immutability.

### Iteration 2 — integration fixture challenge

The unified Stage 14 integration run exposed a fixture that inserted `success` without a
finish time. The fixture was corrected and source auditing was strengthened so tests
cannot silently model an illegal terminal state.

### Iteration 3 — deeper terminal evidence review

Document 72 later found that terminal integrity also required processed-count and
error-message/status relationships. The lifecycle finding remained valid but was
strengthened rather than falsely treated as proof of every terminal-field invariant.

## 16. Residual risks and limitations

This remediation does not prove that the counters themselves reflect upstream reality;
it guarantees lifecycle consistency and immutability once accepted. Administrator-level
manual changes can still bypass constraints if database protections are intentionally
disabled.

A legitimate correction of bad historical terminal evidence requires an explicit repair
procedure with its own audit trail.

## 17. Operational/deployment consequences

Migration 017 must precede code relying on terminal immutability. Legacy lifecycle
violations stop migration for explicit operator repair. Monitoring/operations should
interpret `ErrIngestionRunTransitionRejected` as duplicate/conflicting completion, not as
a missing run.

## 18. Exact evidence and canonical status

```text
implementation commit:
b3603311d86f23c66bc945c8a61471142ccbec63

migration:
017_ingestion_run_terminal_integrity.sql

regression coverage:
internal/repository/postgres/ingestionrun_terminal_integrity_test.go
internal/repository/postgres/ingestionrun_terminal_integrity_integration_test.go

canonical status:
FINDING_GFA_DB_005_INGESTION_RUN_TERMINAL_INTEGRITY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/62_STAGE_14_21_INGESTION_RUN_TERMINAL_INTEGRITY.md
```

Historical pull-request/reviewer identifiers are not asserted when they are not present in
the currently recoverable repository evidence.
