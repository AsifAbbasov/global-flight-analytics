# Backend Context Ownership Audit Closure

Status: Formal source closure; Continuous Integration closure is attached to the commit containing this document
Project: Global Flight Analytics
Reviewed baseline: `3b09d4f642cdc534067ec1fd819593fcee123069`
Baseline GitHub Actions run: `30702233472`
Date: 2026-08-01

## 1. Purpose

This document closes the repository-wide caller-context ownership audit.

The audited invariant is narrow and evidence-based:

```text
A reusable Go function or method that accepts context.Context must not silently
replace that caller-owned context with context.Background() or context.TODO().
```

This is not a mechanical ban on every root context. Process composition roots,
signal-owned roots, and independently bounded terminal bookkeeping remain valid
when their ownership is explicit and they do not overwrite a caller parameter.

## 2. Original Evidence

The original broad source sweep inspected:

```text
Production Go files: 931
Context parameters: 387
Caller-context replacements: 24
```

The twenty-four findings were concrete assignments of a caller parameter through
`context.Background()`. They were not inferred from naming or from the mere
presence of the `context` package.

## 3. Closure Classification

All twenty-four caller-context replacements are closed as **fixed**.

No caller-context replacement is deliberately retained, deferred, allowlisted,
or classified as a false positive.

The corrected surfaces include ingestion fallback, analytical execution, metric
execution, metric query, analytical scope, Airspace Intelligence, transponder
evidence, OurAirports access, Route Intelligence, traffic route context,
Stability Intelligence, Weather Context, migration inspection, and verification
adapters.

The last four findings closed by the change containing this document are:

```text
apps/api/cmd/verify-postgres-projection-intelligence-historical-http-api/assertions.go
  - verifyHistoricalService
  - verifyHistoricalEndpoint

apps/api/cmd/verify-postgres-stability-intelligence-http-api/main.go
  - runtimeStabilityReader.Get

apps/api/internal/database/migrationaudit/postgres.go
  - PostgresStateLoader.Load
```

Verification-only reachability limits production exposure but does not justify
losing caller cancellation ownership. `PostgresStateLoader.Load` performs
PostgreSQL reads and therefore must also reject a missing caller context.

## 4. Required Contract

Every audited operation now follows these rules:

1. A nil caller context is rejected before database, HTTP, repository, evaluator,
   or delegated service work.
2. A non-nil caller context is forwarded without replacement.
3. Existing cancellation and deadline errors remain observable.
4. Regression tests prove zero dependency calls on nil-context rejection.
5. No package receives an exception for assigning `context.Background()` or
   `context.TODO()` to a context parameter.

## 5. Permanent Audit Gate

The permanent tool is:

```text
apps/api/tools/backendcontextownershipaudit
```

The tool uses the Go abstract syntax tree rather than textual matching. It scans
production `.go` files under `apps/api/cmd`, `apps/api/internal`, and
`apps/api/tools`. Test files, `testdata`, and vendored sources are excluded.

For each named standard-library `context.Context` parameter, the audit rejects
assignments whose corresponding expression contains `context.Background()` or
`context.TODO()`. Import aliases and nested expressions are covered.

The permanent Backend Quality command is:

```text
go run ./tools/backendcontextownershipaudit -strict
```

## 6. Regression and Verification Evidence

Focused tests prove:

- zero Projection Intelligence service calls for nil verification context;
- no Fiber request construction for nil verification context;
- zero Stability Intelligence service calls for nil adapter context;
- migration audit rejection before PostgreSQL pool access;
- abstract-syntax-tree detection of `Background`, `TODO`, aliases, nested
  expressions, and nested function scopes;
- repository-wide zero-finding enforcement.

The installer runs targeted tests, focused race tests, the permanent audit, all
backend tests, `go vet`, architecture gates, code-review policy gates, and static
documentation checks in an isolated worktree before changing the real repository.

## 7. Engineering and Continuous Integration Evidence

Recent grouped closure evidence:

```text
3be8685a36ef9fb49eb5d1f8221934e060a81812  Backend CI 30698836704
ebeda1bf089239630690150c0f876a2c712607a4  Backend CI 30699622039
291642215fbbe33e3fdc8075210e4d8359ef08e1  Backend CI 30700233742
104105235edeb8a6c60f619867bf202db4b8f65d  Backend CI 30700868209
3b09d4f642cdc534067ec1fd819593fcee123069  Backend CI 30702233472
```

Earlier focused commits closed worker, ingestion, orchestration, provider,
materialization, executor, and metric-execution ownership.

The commit containing this document is Continuous Integration-closed only after
Backend Quality, PostgreSQL 16 Integration, Backend Race Safety, and Backend
Container all succeed. Its GitHub Actions record is authoritative; this document
does not fabricate a future run identifier.

## 8. Scope Boundary

This closure proves zero caller-parameter replacement through
`context.Background()` or `context.TODO()` in the scanned production Go tree.

It does not claim that every independent root context is automatically correct.
New roots, detached cleanup, timeout selection, and `context.WithoutCancel`
remain governed by `docs/82_CODE_REVIEW_STANDARD.md`.

## 9. Formal Closure

```text
Original caller-context replacements: 24
Fixed caller-context replacements: 24
Deliberately retained replacements: 0
Deferred replacements: 0
Unclassified replacements: 0
Permanent repository-wide gate: installed
```

The caller-context ownership finding set is formally source-closed with no open,
unclassified, or deferred findings.
