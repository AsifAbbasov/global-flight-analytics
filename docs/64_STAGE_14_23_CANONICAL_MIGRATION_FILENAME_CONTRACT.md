# Document 64 — Stage 14.23 Canonical Migration Filename Contract

Status: Implemented
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
