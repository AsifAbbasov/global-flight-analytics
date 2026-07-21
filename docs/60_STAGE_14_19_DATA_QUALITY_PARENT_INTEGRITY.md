# Document 60 — Stage 14.19 Data Quality Parent Integrity

Status: Implementation Baseline v1.2
Project: Global Flight Analytics
Scope: eliminate accidental orphan Data Quality Reports without losing rejected-observation evidence

## 1. Correctness problem

The previous `data_quality_reports` model allowed `flight_state_id` to be null.
That state had two different meanings:

```text
an intentionally rejected observation that was never persisted as flight_states
an accidental missing parent caused by repository or lifecycle failure
```

PostgreSQL could not distinguish those meanings. The canonical quality-report table
therefore accepted records whose relationship to a durable Flight State was not
provable.

## 2. Final persistence boundary

The two meanings are now represented by two different tables.

### Canonical persisted-state reports

`data_quality_reports` contains only reports for rows that exist in `flight_states`.

The database enforces:

```text
state_id is not null
flight_state_id is not null
state_id equals flight_state_id
flight_state_id references flight_states(id)
parent deletion cascades the derived quality report
```

A canonical report can no longer be inserted for a missing Flight State.

### Rejected-observation evidence

`rejected_flight_state_quality_reports` stores quality evidence for observations
that were intentionally rejected before canonical persistence.

This is not represented as a nullable parent inside the canonical table. The new
table records the rejected observation identity, provider context, observation
time, ingestion run when available, validation result, completeness, confidence,
score, missing fields, warnings, and evidence timestamps.

Only `validation_status = 'invalid'` is accepted in this table.

## 3. Repository behavior

`SaveFlightStateQuality` now routes intentionally invalid observations to the
rejected-evidence table.

For every non-rejected report, the insert selects the parent directly from
`flight_states` and returns `ErrDataQualityFlightStateNotPersisted` when no parent
exists. It cannot create a canonical row with a null relationship.

Reconciliation writes verify both conditions inside one PostgreSQL statement:

```text
the reconciliation task is still owned by the current attempt
the referenced Flight State still exists
```

The method preserves the existing task-transition rejection error and returns the
new parent-integrity error when the task is valid but the Flight State is absent.

## 4. Migration 019

`019_data_quality_parent_integrity.sql` performs the transition atomically:

```text
validate existing state identities
create rejected_flight_state_quality_reports
move legacy null-parent reports into the rejected-evidence table
delete those rows from data_quality_reports
make both canonical identity columns not null
replace ON DELETE SET NULL with ON DELETE CASCADE
restore the identity-equality constraint
create rejected-evidence query indexes
```

Rows without any `state_id` are not silently discarded. Migration 019 stops with
an explicit repair error so an operator can inspect them.

## 5. Deletion policy

Data Quality Reports are derived evidence. When their durable `flight_states`
parent is deliberately deleted, the canonical report is deleted with it.

The system does not preserve a report while removing the only entity that gives
that report its canonical meaning.

Rejected-observation evidence remains independent because the corresponding
observation was never admitted into `flight_states`.

## 6. Regression protection

The permanent tests protect:

```text
canonical repository inserts select an existing parent
missing parents return the explicit integrity error
invalid observations use the rejected-evidence table
migration 019 moves legacy null-parent rows
canonical identity columns become not null
canonical parent deletion uses cascade semantics
PostgreSQL rejects null and unknown canonical parents
legacy rejected evidence remains available after migration
```

PostgreSQL integration scenarios run when `TEST_DATABASE_URL` is configured.
The source-level integrity tests run in every ordinary backend test execution.

## 7. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/repository/postgres/data_quality_repository.go \
  internal/repository/postgres/data_quality_parent_integrity_test.go \
  internal/repository/postgres/data_quality_parent_integrity_integration_test.go \
  internal/repository/postgres/data_quality_association_integration_test.go
go test -count=1 ./internal/repository/postgres
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
git diff --check
git status --short
```

Optional PostgreSQL integration execution:

```bash
TEST_DATABASE_URL="$TEST_DATABASE_URL" \
  go test -count=1 ./internal/repository/postgres \
  -run 'DataQuality'
```

## 8. Completion statement

This increment closes accidental orphan creation in the canonical Data Quality
Report table while preserving intentionally rejected observation evidence through
a separate explicit persistence model.

It does not close the remaining PostgreSQL debts concerning trajectory snapshot
consistency, Ingestion Run transitions, trajectory relational constraints, shared
migration filename parsing, altitude precision and status semantics, timestamp
consistency, or repository decomposition.
