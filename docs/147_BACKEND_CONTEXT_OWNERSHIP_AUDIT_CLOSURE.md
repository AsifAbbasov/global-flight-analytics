# Backend Context Ownership Audit Closure

Status: CLOSED
Project: Global Flight Analytics
Remediation-wave parent: `4942db6fbdd792e2b6b62a4bd965f423fdd2c558`
Pre-closure reviewed baseline: `3b09d4f642cdc534067ec1fd819593fcee123069`
Closure engineering commit: `44d41d6ece32b8196a5613c8d4839d7ebe972cac`
Closure Backend CI run: `30703281257`
Date: 2026-08-01
Canonical reconciliation: 2026-08-25

## 1. Purpose

This document closes the repository-wide caller-context ownership audit and records
its remediation history at canonical finding granularity.

The audited invariant is narrow and evidence-based:

```text
A reusable Go function or method that accepts context.Context must not silently
replace that caller-owned context with context.Background() or context.TODO().
```

This is not a mechanical ban on every root context. Process composition roots,
signal-owned roots, and independently bounded terminal bookkeeping remain valid
when their ownership is explicit and they do not overwrite a caller parameter.

## 2. Evidence-honesty correction of the historical count

An earlier revision of this document stated:

```text
Production Go files: 931
Context parameters: 387
Caller-context replacements: 24
```

The first two values are retained as historical sweep evidence. The value `24` is
**not retained as the canonical final replacement count**.

Repository reconstruction shows that the remediation wave from
`4942db6fbdd792e2b6b62a4bd965f423fdd2c558` through
`44d41d6ece32b8196a5613c8d4839d7ebe972cac` contains a broader sequence of direct
caller-context substitutions than that historical summary value describes. The
engineering history includes grouped fixes across workers, ingestion, traffic,
weather, analytics, Airspace Intelligence, transponder evidence, OurAirports,
Route Intelligence, Stability Intelligence, verification adapters and migration
audit code.

The repository does not preserve a pre-remediation AST report that binds an exact
finding total to one exact starting tree. Therefore this canonical record does not
replace `24` with a guessed number. The historical `24` is treated as a stale or
intermediate sweep count, not as authoritative final cardinality.

What is source-proven is stronger and more useful than an uncertain total:

- the affected production surfaces are directly visible in the remediation commits;
- the final closure commit installs the permanent AST audit;
- the exact closure SHA passes that audit in Backend Quality;
- zero deliberate direct caller-parameter replacements are retained or allowlisted.

## 3. Canonical finding

### GFA-OPS-436 — Backend-wide caller-context ownership was not enforced consistently

1. **Finding / symptom**

   Reusable backend functions and methods across multiple architectural boundaries
   silently replaced a missing caller-owned `context.Context` with
   `context.Background()`. The same failure pattern independently appeared in worker,
   ingestion, traffic, analytical, integration, Route Intelligence, Stability,
   Weather, verification and migration-audit surfaces.

2. **Root cause**

   Caller-context ownership existed as a local convention rather than one
   repository-wide executable contract. Individual packages could normalize a nil
   context to an independent root without a permanent cross-repository guard that
   rejected the pattern.

3. **Failure scenario**

   A caller passes a nil, cancelled or deadline-owned context into a reusable
   operation. A nil context is silently replaced with a fresh background root. The
   delegated database, HTTP, repository, evaluator or worker operation can then run
   independently of the caller lifecycle instead of failing fast or observing caller
   cancellation.

4. **Impact**

   Cancellation and deadline ownership can be lost, shutdown can become less
   deterministic, external calls or database work can outlive the initiating
   operation, and error attribution can diverge from the real caller lifecycle. No
   repository evidence reviewed for this reconciliation demonstrates data corruption
   caused by this defect.

5. **Severity rationale**

   **P2 retrospective.** The defect is cross-cutting operational correctness and
   resource-lifecycle risk. It can cause work to continue after the owning request or
   worker lifecycle has ended, but the reviewed evidence does not establish a P1 data
   integrity, security-boundary or irreversible production-loss event.

6. **Existing guarantees violated**

   - caller-owned cancellation and deadline propagation;
   - bounded background work;
   - deterministic worker and request shutdown semantics;
   - fail-fast dependency boundaries;
   - the code-review standard's context-ownership rule.

7. **Considered solutions**

   - retain nil-to-background normalization as an accepted convenience;
   - document package-by-package conventions without enforcement;
   - route all contexts through one helper that may manufacture a root;
   - reject missing caller contexts at reusable boundaries and add a permanent
     repository-wide AST audit.

8. **Chosen remediation**

   Reusable operations reject nil caller contexts before dependency work, forward
   non-nil contexts unchanged, preserve cancellation/deadline errors, add focused
   zero-dependency-call regression tests, and enforce direct parameter replacement
   with `apps/api/tools/backendcontextownershipaudit -strict` in Backend Quality.

9. **Why this solution was selected**

   It preserves ownership rather than hiding missing ownership, fails close at the
   boundary where the contract is known, requires no runtime infrastructure, and
   converts a repeated review class into executable CI policy.

10. **Rejected alternatives**

    - silently manufacturing `context.Background()` for reusable operations;
    - allowlisting selected packages without a demonstrated independent-root owner;
    - relying only on grep or reviewer memory;
    - introducing a framework-level context abstraction solely to solve this class.

11. **Trade-offs**

    Callers that previously relied on undocumented nil normalization now receive an
    explicit error and must own a real context. This is intentionally stricter and
    can expose invalid call sites earlier.

12. **Regression tests / protection**

    Focused tests prove rejection before downstream work for the remediated surfaces.
    The permanent AST audit scans production Go sources under `cmd`, `internal` and
    `tools`, excluding tests, `testdata` and vendored sources. Backend Quality runs:

    ```text
    go run ./tools/backendcontextownershipaudit -strict
    ```

13. **Adversarial review findings**

    Canonical reconstruction found two documentation/audit nuances that must remain
    explicit:

    - the old `24` total is not trustworthy as final full-wave cardinality and is no
      longer asserted;
    - the AST audit detects direct assignment of a named `context.Context` parameter
      through `context.Background()` or `context.TODO()`, but it is not a general
      interprocedural proof against every helper that could manufacture a root.

    The historical Route Store helper that returned `context.Background()` for nil
    input demonstrates why code review and contract tests remain necessary in
    addition to the AST gate.

14. **Remediation iterations**

    The correction was deliberately incremental. The wave includes targeted worker,
    ingestion, traffic, weather, analytics, integration and intelligence commits;
    follow-up commits tightened shutdown/cancellation semantics; the closure commit
    removed the final directly detected residuals and installed the permanent audit.

15. **Residual risks / limitations**

    The AST audit is intentionally narrow. It does not claim interprocedural proof of
    every helper-mediated root, independently created process root,
    `context.WithoutCancel`, or semantically incorrect timeout choice. Those remain
    governed by code review, focused tests and `docs/82_CODE_REVIEW_STANDARD.md`.

16. **Operational / deployment consequences**

    No new service, database migration, environment variable, dependency or
    deployment component is required. Runtime behavior changes only for invalid
    missing-context calls, which now fail fast instead of detaching work from caller
    ownership.

17. **Exact evidence**

    Remediation-wave parent:

    ```text
    4942db6fbdd792e2b6b62a4bd965f423fdd2c558
    ```

    Representative and grouped engineering commits include:

    ```text
    0fab05ead6f18fa6d910c4934aaf24c6b3735b8e  fix: require background worker contexts
    698bab48567e9e962da26c3aeaa984717942ca2c  fix: require traffic ingestion contexts
    4cb15eb8b392923b2f22fa2643ecaffaf370e98a  fix: require traffic application contexts
    a446c3518162fb7044fb6818e6eb5f8be8550aa8  fix: require weather service contexts
    0d8112896666d58b9a1b61419fdaa6496ef091df  fix: require traffic query contexts
    0a3ed287709ac734f2e08f2fecd8fb2613250289  fix: require route store contexts
    37f88266e12c9421ecca5ff6635f39841b2b68ab  fix: require ingestion orchestrator contexts
    4a660620c7f7b3f7978893b093c662a6ba7ae582  fix: require request coalescing contexts
    7f8695fd440c78b4e73f39acac6a25ceab7d6dd6  fix: require provider fallback contexts
    afbf7b8d0949f85ae3b9315ea628337b33665dcc  fix: require flight feature materializer contexts
    be587cc7446ef48121486ebe621c62810328cd7e  fix: require healthcheck contexts
    01be0d39dcc80a87fca773d9f24694213a207452  fix: require traffic fallback provider contexts
    7c479083f5bd26496e029a1323c85c5a1b2ee93b  fix: require analytics executor contexts
    4273de899d409e1fba371899dddb6056b53d78d1  fix: require metric execution contexts
    3be8685a36ef9fb49eb5d1f8221934e060a81812  fix: require metric query contexts
    ebeda1bf089239630690150c0f876a2c712607a4  fix: require analytical scope contexts
    291642215fbbe33e3fdc8075210e4d8359ef08e1  fix: require runtime data source contexts
    104105235edeb8a6c60f619867bf202db4b8f65d  fix: require route intelligence contexts
    3b09d4f642cdc534067ec1fd819593fcee123069  fix: require stability and weather contexts
    44d41d6ece32b8196a5613c8d4839d7ebe972cac  fix: close backend context ownership audit
    ```

    Important remediation follow-ups that strengthen semantics without being counted
    as new direct replacement roots include:

    ```text
    33d308fcfac63103a7f07eead8c5976f8f7bfbe0  fix: handle reconciliation worker shutdown
    26db51f5748512bda896cfc88ee7fb65aef92b50  fix: preserve ingest wait context ownership
    a2faf3481f9db94d87fbfe0f778fc026b95c0e4b  fix: require reconciliation batch contexts
    8d1366ecc6ae3528dfdf68cda4a8c46c296dd098  fix: require airport HTTP validator contexts
    ```

    Historical grouped Backend CI evidence recorded by the original document:

    ```text
    3be8685a36ef9fb49eb5d1f8221934e060a81812  Backend CI 30698836704
    ebeda1bf089239630690150c0f876a2c712607a4  Backend CI 30699622039
    291642215fbbe33e3fdc8075210e4d8359ef08e1  Backend CI 30700233742
    104105235edeb8a6c60f619867bf202db4b8f65d  Backend CI 30700868209
    3b09d4f642cdc534067ec1fd819593fcee123069  Backend CI 30702233472
    ```

    Exact closure evidence recovered from GitHub Actions:

    ```text
    SHA: 44d41d6ece32b8196a5613c8d4839d7ebe972cac
    Backend CI: 30703281257 — SUCCESS
    Backend Race Safety — SUCCESS
    Backend Quality — SUCCESS
      Run backend context ownership audit — SUCCESS
    PostgreSQL 16 Integration — SUCCESS
    Backend Container — SUCCESS
    ```

    Earlier canonical findings that overlap local portions of the same policy include
    `GFA-DB-021`, `GFA-DB-028`, `GFA-DB-030`, `GFA-OPS-172`, `GFA-OPS-177`,
    `GFA-OPS-185`, `GFA-OPS-195` and `GFA-OPS-204`. They remain valid local records;
    `GFA-OPS-436` records the later repository-wide policy defect rather than
    duplicating those histories.

18. **Final canonical status**

    **CLOSED.** Direct caller-parameter replacement through
    `context.Background()`/`context.TODO()` is source-closed for the audited tree,
    exact closure CI is green, and the repository-wide prevention gate is installed.

19. **Prevention / future guard**

    Keep `backendcontextownershipaudit -strict` in Backend Quality; require explicit
    ownership for every independent root; reject reusable nil-context normalization;
    and review helper-mediated context creation separately because the AST gate is not
    an interprocedural ownership proof.

## 4. Closure classification

No direct caller-context replacement is deliberately retained, deferred, allowlisted,
or classified as a false positive.

The corrected surfaces include ingestion fallback, analytical execution, metric
execution, metric query, analytical scope, Airspace Intelligence, transponder
evidence, OurAirports access, Route Intelligence, traffic route context,
Stability Intelligence, Weather Context, migration inspection, and verification
adapters.

The last four direct findings closed by `44d41d6ece32b8196a5613c8d4839d7ebe972cac`
are:

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

## 5. Required contract

Every audited operation now follows these rules:

1. A nil caller context is rejected before database, HTTP, repository, evaluator,
   or delegated service work.
2. A non-nil caller context is forwarded without replacement.
3. Existing cancellation and deadline errors remain observable.
4. Regression tests prove zero dependency calls on nil-context rejection.
5. No package receives an exception for assigning `context.Background()` or
   `context.TODO()` to a context parameter.

## 6. Permanent audit gate

The permanent tool is:

```text
apps/api/tools/backendcontextownershipaudit
```

The tool uses the Go abstract syntax tree rather than textual matching. It scans
production `.go` files under `apps/api/cmd`, `apps/api/internal`, and
`apps/api/tools`. Test files, `testdata`, and vendored sources are excluded.

For each named standard-library `context.Context` parameter, the audit rejects
direct assignments whose corresponding expression contains `context.Background()`
or `context.TODO()`. Import aliases and nested expressions are covered.

## 7. Formal closure

```text
Historical summary count `24`: superseded as non-canonical final cardinality
Exact reconstructed original cardinality: not asserted without preserved pre-remediation AST output
Deliberately retained direct replacements: 0
Deferred direct replacements: 0
Unclassified direct replacements: 0
Permanent repository-wide direct-assignment gate: installed
Exact closure Backend CI: 30703281257 — SUCCESS
Canonical finding: GFA-OPS-436 — CLOSED
```

The repository-wide caller-context ownership finding is formally closed. The audit's
known interprocedural limitation remains explicit rather than being misrepresented as
a proof about every possible independent context root.
