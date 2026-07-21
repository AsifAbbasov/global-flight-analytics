# Document 79 — Post-Closure Migrator Context Hardening

Status: Implemented v1.0
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

The repository now includes unit and source-contract tests plus a Stage 14 source-audit rule that
rejects any return of `ctx = context.Background()` in migrator/runner.go while requiring bounded
independent cleanup contexts.

## 5. Status decision

Stage 14 remains closed. This post-closure corrective increment does not alter application
features, database schema, migration history, provider behavior, HTTP contracts, frontend
behavior, or analytical formulas.

Successful verification emits:

```text
POST_CLOSURE_MIGRATOR_CONTEXT_HARDENING=PASS
STAGE_14_OVERALL_STATUS=CLOSED
```
