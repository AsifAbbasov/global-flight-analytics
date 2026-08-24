# Document 60 — Stage 14.19 Data Quality Parent Integrity

Status: Remediation History v1.3
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

## 9. Finding history and root cause

### Finding

One nullable foreign-key shape represented two incompatible states: deliberately rejected
observations with no canonical `flight_states` row and accidental loss or absence of the
required canonical parent.

### Root cause

The persistence model overloaded `NULL` as both valid domain evidence and relational
failure. That made the application responsible for remembering which meaning applied,
while PostgreSQL could not enforce the distinction.

### Failure scenario

```text
quality result is treated as belonging to a canonical Flight State
↓
parent persistence fails, is skipped, or parent is later removed
↓
data_quality_reports still accepts a null/missing relationship
↓
consumer sees a canonical quality record whose durable observation cannot be proven
```

The inverse case also mattered: intentionally rejected observations still required
quality evidence even though they were never valid canonical Flight States.

### Impact

The ambiguity weakened auditability, provenance, reconciliation, deletion semantics, and
data-quality trust. Downstream analytics could not reliably distinguish `rejected before
persistence` from `canonical parent unexpectedly absent`.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1 data
integrity/provenance** because the defect allowed canonical evidence to exist without a
provable canonical parent while also risking loss of valid rejected-observation evidence.

### Existing guarantees violated

```text
canonical derived evidence has a durable canonical parent
rejected observations remain representable without pretending they were persisted
referential integrity is enforced by PostgreSQL, not caller discipline
parent deletion has explicit deterministic semantics
```

## 10. Considered and rejected alternatives

### Keep one nullable table and add an application-side discriminator

Rejected because PostgreSQL would still accept ambiguous states and every writer/reader
would need to preserve the convention correctly.

### Create a synthetic/sentinel Flight State for rejected observations

Rejected because it would manufacture a canonical entity that never passed admission and
would contaminate trajectory, analytics, and provenance semantics.

### Preserve canonical reports after parent deletion with `ON DELETE SET NULL`

Rejected because the report would outlive the only canonical entity that gives the report
its meaning, recreating the original ambiguity.

### Chosen remediation

Split persisted-state reports and rejected-observation evidence into separate tables,
make canonical parent identity non-null and foreign-key enforced, and use cascade deletion
for derived canonical evidence.

## 11. Why this solution and trade-offs

The model makes invalid states unrepresentable at the database boundary instead of
requiring application convention.

Trade-offs:

```text
+ explicit canonical vs rejected semantics
+ PostgreSQL-enforced parent integrity
+ rejected evidence remains durable
+ deterministic deletion behavior
- two persistence/query paths instead of one overloaded table
- migration must classify and move legacy nullable rows safely
- consumers needing both evidence types must intentionally query both models
```

The additional table is justified because it removes semantic ambiguity rather than
adding abstraction for its own sake.

## 12. Adversarial review and remediation iterations

### Iteration 1 — association integrity foundation

Earlier backend hardening introduced typed Data Quality associations and constrained
identity equality, but nullable parent semantics still left two meanings encoded in one
shape.

### Iteration 2 — parent-integrity challenge

Review asked whether a canonical report can still exist when no durable Flight State can
be proven. The answer was yes because `flight_state_id` remained nullable.

### Iteration 3 — explicit split

Implementation commit `0d3d1d37a65423ca6263df0816360eabf3c66235`
(`fix: enforce data quality parent integrity`) introduced the separate rejected-evidence
model and canonical non-null parent rule.

### Iteration 4 — migration adversarial case

Migration logic deliberately refuses to silently discard rows with no usable `state_id`.
That fail-closed repair boundary prevents the remediation itself from hiding unclassified
legacy corruption.

## 13. Residual risks and limitations

The remediation does not prove that upstream validation logic is correct; it guarantees
where evidence is stored and how it relates to canonical persistence. Manual database
changes can still bypass application intent if constraints are deliberately disabled.

Rejected-evidence retention policy and long-term archival volume are separate operational
concerns.

## 14. Operational and deployment consequences

Migration 019 must complete before code relying on the split model is considered fully
deployed. If the migration encounters an unclassifiable legacy row, deployment should
stop for operator repair rather than continue with silent data loss.

Canonical parent deletion now removes derived canonical quality evidence by design;
rejected evidence remains independent.

## 15. Exact evidence

```text
implementation commit:
0d3d1d37a65423ca6263df0816360eabf3c66235

migration:
019_data_quality_parent_integrity.sql

regression coverage:
internal/repository/postgres/data_quality_parent_integrity_test.go
internal/repository/postgres/data_quality_parent_integrity_integration_test.go
internal/repository/postgres/data_quality_association_integration_test.go
```

A historical pull-request number is not asserted because it is not preserved in the
currently searchable evidence. The implementation commit, migration, and regression tests
are the canonical evidence.

## 16. Final canonical status

```text
FINDING_GFA_DB_003_DATA_QUALITY_PARENT_INTEGRITY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/60_STAGE_14_19_DATA_QUALITY_PARENT_INTEGRITY.md
IMPLEMENTATION_COMMIT=0d3d1d37a65423ca6263df0816360eabf3c66235
```
