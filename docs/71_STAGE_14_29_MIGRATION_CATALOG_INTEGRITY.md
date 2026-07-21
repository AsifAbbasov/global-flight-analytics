# Document 71 — Stage 14.29 Migration Catalog Integrity

Status: Implementation Baseline v1.1
Project: Global Flight Analytics
Scope: restore deployability of the production migration catalog and reopen Stage 14 honestly

## 1. Confirmed blocker

The repository contained two migration files with version `016`:

```text
016_add_flight_state_observation_metadata.sql
016_data_quality_parent_integrity.sql
```

The production migrator rejects duplicate versions before applying pending migrations.
Package integration tests did not prove catalog deployability because several fixtures
executed SQL files directly instead of calling `Runner.ListMigrations`.

## 2. Canonical numbering decision

Flight State observation metadata keeps version `016`. Data Quality Parent Integrity
moves to the next available version:

```text
019_data_quality_parent_integrity.sql
```

The migration body and schema semantics are unchanged. Only its catalog identity and the
integration-test path are corrected.

## 3. Permanent regression protection

`repository_catalog_test.go` calls the production `Runner.ListMigrations` against the real
repository directory. It fails on non-canonical names, duplicate versions, missing
canonical owners, or reintroduction of the retired duplicate filename.

The permanent source audit also requires migration `019` across the Data Quality source
and integration tests, Document 60, the README status surface, and the implementation
sequence. This prevents catalog identity, tests, and documentation from drifting apart.

The cross-stack script and PostgreSQL continuous integration job now run `cmd/migrate`
against a clean PostgreSQL database, run it a second time to prove idempotency, and verify
that `schema_migrations` contains exactly one row for every SQL file in the catalog.

## 4. Status correction

The former marker `STAGE_14_COMPLETION_AUDIT=PASS` is retired because it overstated the
evidence. The current gate reports:

```text
STAGE_14_PRODUCTION_MIGRATOR=PASS
STAGE_14_CURRENT_SCOPE_AUDIT=PASS
STAGE_14_OVERALL_STATUS=REOPENED
```

Stage 14 remains reopened while the remaining correctness and maintainability debts are
worked through in separate evidence-backed increments.
