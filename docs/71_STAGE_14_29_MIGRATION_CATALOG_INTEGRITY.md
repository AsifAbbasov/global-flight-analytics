# Document 71 — Stage 14.29 Migration Catalog Integrity

Status: Remediation History v1.2
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

## 5. Finding history, root cause, and failure scenario

### Finding

The real repository migration catalog contained duplicate version `016`, so the production
migrator could not enumerate and deploy the supposedly completed catalog.

### Root cause

Data Quality Parent Integrity was added as another `016` while existing package tests
proved individual SQL behavior by executing files directly. The closure gate therefore
validated migration contents without exercising the same catalog-discovery path used by
production `cmd/migrate`.

### Failure scenario

```text
Stage 14 source/package checks pass
↓
repository is treated as completion candidate
↓
production migrator calls Runner.ListMigrations on the full directory
↓
two files claim version 016
↓
migrator rejects catalog before application
↓
a clean production database cannot be deployed from the repository baseline
```

### Impact

The defect invalidated deployability evidence and therefore invalidated the earlier Stage
14 closure claim. A repository whose full migration catalog cannot be applied from clean
state is not a reproducible production baseline even if individual migration tests pass.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1
deployment/correctness** because the defect blocked clean production migration and proved
a prior formal closure statement false.

### Existing guarantees violated

```text
migration versions are unique across the real repository catalog
production and test paths discover migrations through compatible semantics
clean database bootstrap is reproducible
closure claims include production-migrator evidence
one applied history row corresponds to one canonical migration file
```

## 6. Considered and rejected alternatives

### Allow duplicate migration versions in the migrator

Rejected because version identity is the ordering and history key. Permitting duplicates
would make application order/history ambiguous rather than repair the catalog.

### Keep both version 016 files and change only tests

Rejected because the production catalog itself was invalid; stronger tests without a
catalog correction would correctly continue to fail.

### Renumber the older Flight State observation migration

Rejected because that migration already owned the earlier historical catalog position.
The later Data Quality migration was the conflicting addition and moved to the next
available version instead.

### Renumber dynamically at runtime

Rejected because migration identity must be a committed repository artifact. Runtime
renumbering would make checksums, history, documentation, and operator evidence unstable.

### Chosen remediation

Keep Flight State observation metadata at `016`, move Data Quality Parent Integrity to
`019`, update all references, and add permanent tests that exercise the production
migration catalog and real migrator twice on a clean PostgreSQL database.

## 7. Why this solution and trade-offs

The remediation repairs the source-of-truth catalog and closes the test-evidence gap at
the same time.

Trade-offs:

```text
+ production catalog becomes deployable and uniquely ordered
+ clean bootstrap and idempotent second run become permanent evidence
+ documentation/tests share the same migration identity
- references to the Data Quality migration must be updated consistently
- earlier closure evidence is explicitly invalidated rather than preserved as a success claim
```

Invalidating the old closure marker is a feature of the process, not a documentation
failure: the project prefers corrected evidence over status continuity.

## 8. Adversarial review and remediation iterations

### Iteration 1 — early cross-stack completion attempt

Commit `eb37e03c6793314e446cdb048ae9584e38f2567c`
(`fix: complete stage 14 final audit`) established an early unified audit and produced a
completion candidate.

### Iteration 2 — production-catalog challenge

A later review exercised the real migration discovery path and found duplicate version
`016`. This demonstrated that package-level migration tests had not proven repository
catalog deployability.

### Iteration 3 — catalog repair and reopening

Implementation commit `4ef16aaa53e5b749e841a4b3226516c65da1bd06`
(`fix: validate migration catalog integrity`) renumbered the later migration, added real
catalog validation, and changed the authoritative Stage 14 status back to `REOPENED`.

### Iteration 4 — independent final closure

Only after subsequent Stage 14 findings were remediated did commit
`202c00cabb352b50a6d3a2a6002ad3401c1ad23e`
(`chore: close Stage 14 after final audit`) establish the later final closure described by
Document 78.

## 9. Residual risks and limitations

Catalog integrity proves canonical naming/version uniqueness and real migrator
applicability. It does not by itself prove each migration's business semantics, absence of
manual production drift, or successful migration against arbitrary unsupported legacy
databases.

A migration may still be individually wrong even when the catalog is structurally valid;
that is why clean PostgreSQL execution and package integration remain separate gates.

## 10. Operational/deployment consequences

The production migration command must be part of release evidence. Clean-database
application, a second idempotent application, and applied-count verification are required
before a migration-catalog closure claim. Duplicate versions are blockers, not warnings.

## 11. Exact evidence

```text
early invalidated completion candidate:
eb37e03c6793314e446cdb048ae9584e38f2567c

implementation/reopening commit:
4ef16aaa53e5b749e841a4b3226516c65da1bd06

later independent final closure:
202c00cabb352b50a6d3a2a6002ad3401c1ad23e

permanent evidence:
repository_catalog_test.go
production cmd/migrate clean-database execution
second idempotent migration execution
schema_migrations applied-count verification
```

## 12. Final canonical status

```text
FINDING_GFA_DB_013_MIGRATION_CATALOG_INTEGRITY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/71_STAGE_14_29_MIGRATION_CATALOG_INTEGRITY.md
IMPLEMENTATION_COMMIT=4ef16aaa53e5b749e841a4b3226516c65da1bd06
```

The finding is closed. The historical reopening remains canonical evidence that the first
Stage 14 closure attempt was not accepted once stronger production evidence contradicted
it.
