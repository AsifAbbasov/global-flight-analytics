# Document 79 — Post-Closure Migrator Context Hardening

Status: Implemented v1.1
Project: Global Flight Analytics
Scope: remove remaining nil caller-context substitution from the PostgreSQL migrator without reopening Stage 14

## 1. Confirmed residual defect

The Stage 14.33 repository-context contract covered named PostgreSQL repository paths. A later
global review found two remaining substitutions in the migration runner: ApplyPending and
withMigrationLock accepted nil and replaced it with context.Background(). That hid caller intent
and made the migration API inconsistent with the hardened repository boundary.

## 2. Corrected contract

The migrator now exposes ErrMigrationContextRequired and validates caller context at:

```text
EnsureSchemaMigrations
Status
ApplyPending
ensureSchemaMigrations
applyMigrationAtomically
withMigrationLock
appliedMigrations
appliedMigrationsWith
```

A nil context fails before pool access, SQL execution, transaction creation, or advisory-lock
acquisition.

## 3. Intentional cleanup contexts

The bounded contexts inside releaseMigrationLock, destroyLockedConnection, and
rollbackMigrationTransaction remain intentional. They are cleanup contexts used after the caller
may already be cancelled; they do not replace an input context for normal database work.

## 4. Permanent protection

The permanent gate now parses every production Go file in the migrator package with the standard
`go/parser` and `go/ast` packages. It does not depend on comments, whitespace, formatting, or one
exact assignment string.

The syntax-tree policy:

- rejects `context.Background()` outside the three named cleanup functions;
- rejects `context.TODO()` everywhere in production migrator code;
- rejects `context.WithoutCancel()` so database work cannot silently detach from caller cancellation;
- recognizes renamed and dot-imported `context` packages;
- rejects storing `context.Background` or `context.TODO` as function values;
- rejects reassignment of caller-owned `context.Context` parameters;
- scans helper functions and additional production files in the migrator package;
- requires every database-reaching migrator boundary to call `requireMigrationContext` directly;
- permits cleanup contexts only as the exact bounded expression
  `context.WithTimeout(context.Background(), migrationLockReleaseTimeout)`.

Behavioral nil-context tests remain in place. The syntax-tree audit adds structural regression
protection without changing runtime migration behavior.

## 5. Status decision

Stage 14 remains closed. This post-closure corrective increment does not alter application
features, database schema, migration history, provider behavior, HTTP contracts, frontend
behavior, or analytical formulas.

Successful verification emits:

```text
MIGRATOR_CONTEXT_AST_AUDIT=PASS
POST_CLOSURE_MIGRATOR_CONTEXT_HARDENING=PASS
STAGE_14_OVERALL_STATUS=CLOSED
```
