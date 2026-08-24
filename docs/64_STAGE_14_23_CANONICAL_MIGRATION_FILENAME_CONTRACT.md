# Document 64 — Stage 14.23 Canonical Migration Filename Contract

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: establish one canonical migration file identity parser for execution, audit, and repair verification

## 1. Correctness problem

Three migration-related subsystems interpreted migration identity independently:

- `migrator` parsed files before execution;
- `migrationaudit` parsed files while reconciling local files with database history;
- `migrationrepair` separately encoded the expected version and name of its protected historical migration.

The interpretations were not equivalent. In particular, the migrator accepted versions and names that the audit rejected. A file could therefore be executable but simultaneously reported as non-canonical, or a repair preflight could describe an expected identity using duplicated constants.

## 2. Canonical contract

The package `internal/database/migrationfile` is now the only owner of migration file identity parsing.

Canonical format:

```text
NNN_name.sql
```

Rules:

- the version contains exactly three ASCII digits;
- the separator between version and name is one underscore;
- the name is non-empty;
- the name may contain Unicode letters, Unicode digits, and underscores;
- the extension is exactly lowercase `.sql`;
- leading and trailing whitespace is rejected;
- path separators are rejected because the parser accepts a file name, not a path.

The parser returns one immutable value containing `Version`, `Name`, and the canonical `FileName`.

## 3. Subsystem integration

### 3.1 Migrator

`Runner.ListMigrations` parses every SQL file through `migrationfile.Parse`. Invalid SQL file names stop migration discovery before any migration is applied.

The former private `parseMigrationFileName` implementation is removed.

### 3.2 Migration audit

The local scanner uses the same parser. An invalid SQL file remains an audit blocker, but the reason now comes from the shared canonical contract.

The former private `parseLocalMigrationFileName` implementation is removed.

### 3.3 Migration repair

The protected historical identity is now declared as one canonical file name:

```text
010_add_reconciliation_result_identity.sql
```

`migrationrepair` derives the expected version and name from that file name through `migrationfile.MustParse`. It no longer stores an independently interpreted name constant.

`MustParse` is restricted to source-owned package constants. Runtime file discovery always uses the error-returning `Parse` function.

## 4. Preserved behavior

This increment does not change:

- migration SQL contents;
- migration checksums;
- `schema_migrations` rows;
- migration ordering for canonical files;
- audit finding severity;
- the one-time repair verifier's database checks;
- migration transaction or advisory-lock behavior.

No PostgreSQL schema migration is required.

## 5. Deliberate hardening

The canonical parser rejects ambiguous names that the old migrator accepted, including:

```text
10_short.sql
ABC_letters.sql
010_invalid-name.sql
001_name.SQL
```

It also requires ASCII digits for versions. Unicode decimal digits are not accepted as migration versions because database history, lexical ordering, operator tooling, and documentation all use ASCII version identifiers.

## 6. Regression protection

Tests protect:

- accepted canonical file names;
- all invalid version, extension, whitespace, path, and name cases;
- migrator rejection of non-canonical SQL files;
- audit invalid-file reporting through the shared parser;
- migration repair identity derivation from its canonical file name;
- source ownership, ensuring private parser implementations are not reintroduced.

## 7. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/database/migrationfile internal/database/migrator internal/database/migrationaudit internal/database/migrationrepair
go test -count=1 ./internal/database/migrationfile ./internal/database/migrator ./internal/database/migrationaudit ./internal/database/migrationrepair
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
git diff --check
git status --short
```

## 8. Completion boundary

This increment closes the shared migration filename parser debt. Remaining PostgreSQL correctness work is limited to separate data semantics and maintainability concerns recorded in Document 58.

## 9. Finding history, root cause, and failure scenario

### Finding

Migration execution, migration auditing, and one-time repair verification used independent migration-identity parsing rules.

### Root cause

File identity was treated as a local parsing detail in each subsystem instead of one database-evolution contract. Duplicate parsers and constants therefore drifted in accepted version width, allowed name characters, and expected protected migration identity.

### Failure scenario

```text
migration file is accepted by migrator
↓
production migration can execute or be listed as pending
↓
migrationaudit applies a different filename rule and rejects the same file
↓
operators receive contradictory deployment evidence
```

A repair verifier could also drift from the canonical filename if its version/name constants were edited independently.

### Impact

The defect weakened migration governance and could make a deployable migration simultaneously fail audit/repair checks, or make repair tooling reason about a different identity than the migrator. This is a correctness boundary for deployment evidence, even when no SQL data is changed directly.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P2 deployment-contract consistency** because the inconsistency could block or misdiagnose deployment, but did not by itself mutate valid application data.

### Existing guarantees violated

```text
one migration filename has one canonical identity across all migration tooling
execution and audit interpret the same local catalog
protected repair identities cannot drift through duplicated constants
invalid migration names fail before schema mutation
```

## 10. Considered and rejected alternatives

### Keep separate parsers but synchronize tests

Rejected because tests would still protect multiple sources of truth and require duplicated rule changes in the future.

### Make migration audit call the migrator's private parser

Rejected because it would couple audit semantics to a package-private implementation rather than establish a neutral canonical owner usable by migrator, audit, and repair.

### Relax every subsystem to the broadest historical syntax

Rejected because ambiguous/non-canonical names make lexical ordering, operator review, and database history less predictable. The project benefits from a strict repository-owned convention.

### Chosen remediation

Create `internal/database/migrationfile` as the single owner of migration filename identity and make all migration-related subsystems consume it.

## 11. Why this solution and trade-offs

A small shared value/parser removes real duplicated behavior without introducing a framework or service.

Trade-offs:

```text
+ one migration identity contract
+ execution/audit/repair cannot drift silently
+ invalid names fail before application
- previously tolerated non-canonical file names are now rejected
- future naming-rule changes require deliberate migrationfile contract review
```

The stricter naming cost is accepted because migration files are repository-controlled artifacts, not user input requiring broad compatibility.

## 12. Adversarial review and remediation iterations

### Iteration 1 — identify parser drift

Review compared migrator and audit behavior and found that a file could be executable while audit classified it as invalid.

### Iteration 2 — centralize parser ownership

Implementation commit `4c41b8588e9119f59c090c976cef494c55683e18`
(`refactor: centralize migration filename parsing`) introduced the shared owner and removed private duplicate parsers.

### Iteration 3 — repair identity challenge

The same review extended to `migrationrepair`, where duplicated version/name constants could drift. The protected identity was reduced to one canonical file name parsed by the shared contract.

### Iteration 4 — ambiguous Unicode/path challenge

The parser explicitly keeps versions ASCII-only, rejects path separators and whitespace, and limits `MustParse` to source-owned constants so runtime discovery remains error-returning and fail closed.

## 13. Residual risks and limitations

The parser validates file identity only. It does not prove migration SQL correctness, checksum integrity, version uniqueness across the catalog, or compatibility of the SQL with the runner transaction model. Those are separate migration-catalog and execution findings.

## 14. Operational/deployment consequences

A newly added non-canonical `.sql` file now blocks migration discovery/audit before schema changes occur. Operators should rename the repository artifact rather than bypass the parser.

## 15. Exact evidence

```text
implementation commit:
4c41b8588e9119f59c090c976cef494c55683e18

canonical owner:
internal/database/migrationfile

regression coverage:
migrationfile parser tests
migrator invalid-filename tests
migrationaudit shared-parser tests
migrationrepair canonical-identity tests
```

## 16. Final canonical status

```text
FINDING_GFA_DB_007_MIGRATION_FILENAME_CONTRACT=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/64_STAGE_14_23_CANONICAL_MIGRATION_FILENAME_CONTRACT.md
IMPLEMENTATION_COMMIT=4c41b8588e9119f59c090c976cef494c55683e18
```

Historical PR/reviewer identifiers are not invented when unavailable.
