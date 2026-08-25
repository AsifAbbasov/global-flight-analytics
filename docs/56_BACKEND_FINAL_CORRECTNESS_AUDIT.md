# Document 56 — Backend Final Correctness Audit

Status: Remediation History / Reopened Amendment v1.4
Project: Global Flight Analytics
Scope: permanent reproducible backend correctness gate before Stage 15

## 1. Purpose

Stage 14 corrected four high-risk backend boundaries:

```text
Projection Intelligence read snapshot consistency
nullable telemetry integrity
Historical Intelligence pagination integrity
Weather production composition
```

Passing tests once is not enough. The repository needs a permanent,
reproducible gate that prevents these defects from returning while later work
focuses on frontend visual design.

This document defines that gate.

## 2. Added Verification Assets

The repository contains:

```text
apps/api/tools/backendfinalaudit
scripts/verify-backend-final-correctness.sh
```

`backendfinalaudit` performs source-level invariant checks. The shell verifier composes those checks with architecture audits, focused tests, race detection, complete Go compilation/static analysis and the backend test suite.

## 3. Projection Snapshot Consistency Gate

The audit requires a single `LoadSnapshot` service boundary, `REPEATABLE READ`, `READ ONLY`, one transaction-scoped trajectory/data client, commit on success and bounded independent rollback cleanup on failure.

## 4. Nullable Telemetry Integrity Gate

The audit requires trajectory point queries to preserve nullable telemetry, prohibit zero/false fabrication, filter incomplete required telemetry and preserve explicit zero values. Scanners retain PostgreSQL nullability and completeness validation.

## 5. Historical Pagination Integrity Gate

The audit requires the full `(WindowEnd, WindowStart, AsOfTime, ID)` cursor, a matching mixed-direction PostgreSQL keyset predicate, opaque versioned HTTP transport, strict decoding and absence of legacy single-field cursor names.

## 6. Weather Composition Boundary Gate

The audit requires narrow Weather coordination, provider-owned external/governance composition, application-owned repository/service/handler construction, route-only HTTP registration, and preserved timeout/route behavior.

## 7. Existing Architecture and Security Gates

The final verification script also runs `tools/projectaudit -mode all -strict`, preserving checks for shared confidence vocabulary, Go/TypeScript trajectory contracts, mutation authorization, formula benchmark isolation and analytical production reachability.

## 8. Runtime Verification Coverage

Focused build/test coverage includes Projection, Historical Intelligence/Aggregate, Historical replay/materialization, Weather, Stability, Airspace, source-constraint and OpenSky compatibility verifiers. Database-dependent commands are compiled and package-tested without falsely claiming an external database connection during offline source audit.

## 9. Race Detection

Race coverage includes Projection read, Historical store/cursor/HTTP, server composition, Weather orchestration, provider budget/response state, PostgreSQL repositories and internal mutation authorization.

## 10. Reproducible Command

From repository root:

```bash
scripts/verify-backend-final-correctness.sh
```

Successful completion ends with:

```text
BACKEND_FINAL_CORRECTNESS_AUDIT=PASS
```

## 11. Non-Goals

This audit does not claim production load evidence, external PostgreSQL availability, deployed Render/Neon health, frontend visual completion or browser-level end-to-end coverage.

## 12. Acceptance

Acceptance requires backendfinalaudit tests/strict scan, projectaudit strict scan, focused corrected-boundary tests, race detection, all command builds, vet, complete Go tests, frontend dependency/lint/type/build checks, backend container build and diff validation.

<!-- STAGE-14-16-END-TO-END-TELEMETRY-AVAILABILITY:SCOPE-AMENDMENT -->

## 13. Stage 14.16 scope amendment

The original nullable telemetry section protected the Projection PostgreSQL read boundary. Stage 14.16 expands the permanent audit to provider mapping, Flight State availability, PostgreSQL writes, general/reconciliation reads, Traffic, Airspace, validator behavior and Projection eligibility.

## 14. Stage 14 current-scope amendment

The command in this document remains the permanent backend-specific correctness gate. Document 70 later composes it with PostgreSQL, security, frontend and container checks through `scripts/verify-stage-14-completion.sh`.

Document 71 proved that a previous broader completion claim was too strong because the production migration catalog had not been exercised through `cmd/migrate`. Therefore `BACKEND_FINAL_CORRECTNESS_AUDIT=PASS` is necessary evidence for its defined backend boundaries, but never proof that every Stage 14 or release concern is closed.

## 15. Canonical remediation history

### Finding / symptom

Several high-risk backend defects had been repaired independently, but their invariants were distributed across package tests and stage-specific checks. A later change could regress one boundary while unrelated tests still passed, and a one-time successful remediation run could be mistaken for durable closure evidence.

### Root cause

The project had accumulated multiple correctness gates organically as each defect was fixed, without one permanent source-level and behavioral verifier explicitly owning the corrected Stage 14 backend invariants.

### Failure scenario

```text
high-risk defect is fixed and local tests pass
↓
project moves to later stages
↓
future refactor changes SQL/transaction/cursor/composition behavior
↓
no single closure command checks all corrected invariants together
↓
regression may merge while historical documentation still says the defect was closed
```

### Impact

The gap threatened durability of remediation evidence and could allow closed correctness findings to reappear silently. It was primarily release/governance risk, but it protected P1/P2 data and consistency invariants.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P2 governance/reliability**. The gap did not itself corrupt data, but it weakened the permanent protection around several previously confirmed P1 correctness defects.

### Existing guarantees violated

```text
closure must be reproducible rather than one-time
high-risk corrected invariants require permanent regression ownership
one audit marker must have a precisely bounded claim
source-level guards and behavioral tests must complement rather than replace one another
```

### Considered solutions

1. Rely only on existing package tests.
2. Duplicate every protected test into one giant new suite.
3. Add a lightweight source-invariant audit and compose existing focused tests/audits/race/build checks through one reproducible command.

### Chosen remediation

Option 3: `backendfinalaudit` owns source-level invariants while `verify-backend-final-correctness.sh` composes existing high-value behavioral and architecture verification without cloning their implementation.

### Why this solution was selected

It creates one durable closure command while avoiding duplicated test logic and a new testing framework. Each underlying subsystem keeps its natural tests; the final audit verifies that the critical protections still exist and are executed together.

### Rejected alternatives

Package tests alone were rejected because no single test suite owned cross-package source invariants or proved that all relevant checks remained wired into closure. A monolithic duplicated suite was rejected because it would create a second, drifting implementation of the same tests. Treating the audit as full production validation was rejected because external deployment/load evidence belongs to separate scopes.

### Trade-offs

```text
+ one reproducible backend closure command
+ permanent source guards around fragile invariants
+ existing tests remain the behavioral owners
+ audit claim is explicitly scoped
- source audits require maintenance when legitimate architecture changes occur
- passing the backend audit still does not prove deployed infrastructure health or full release readiness
```

### Regression tests / protection

`backendfinalaudit` unit tests, strict repository scan, projectaudit, focused package tests, race tests, complete build/vet/test, frontend dependency/tooling checks and backend container construction form the permanent composed gate.

### Adversarial review findings

The most important later challenge was Document 71: a broader Stage 14 completion claim had passed the backend final audit while the production migration catalog still contained a blocker outside this audit's scope. The correct response was not to invalidate the backend audit, but to narrow its canonical claim and introduce a broader Stage 14 production-migrator gate.

Historical PR/reviewer evidence for the original July audit addition is unavailable; reconstruction is limited to repository source, tests, commits and subsequent documented reopening evidence.

### Remediation iterations

1. Commit `483815bdd60251e16960ec480cadd3bb93ee7f28` added the backend final correctness audit.
2. Stage 14.16 expanded nullable-telemetry protection end to end.
3. Documents 70–71 later amended the scope: backend audit remains necessary but is no longer treated as sufficient evidence for total Stage 14 closure.

### Residual risks and limitations

Static source audits can become stale if implementation structure changes legitimately. The command also does not exercise deployed cloud infrastructure, production load or every external integration. Those are intentionally separate acceptance layers.

### Operational or deployment consequences

The audit becomes a pre-merge/release evidence requirement for backend correctness-sensitive changes. A failing invariant should block closure until either the behavior is restored or the audit is deliberately updated with reviewed evidence for a legitimate contract change.

### Exact evidence

```text
implementation commit:
483815bdd60251e16960ec480cadd3bb93ee7f28

primary command:
scripts/verify-backend-final-correctness.sh

source audit:
apps/api/tools/backendfinalaudit

scope correction / broader gate:
Documents 70 and 71
```

### Final canonical status

```text
FINDING_GFA_GOV_054_PERMANENT_BACKEND_CORRECTNESS_GATE=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/56_BACKEND_FINAL_CORRECTNESS_AUDIT.md
IMPLEMENTATION_COMMIT=483815bdd60251e16960ec480cadd3bb93ee7f28
CANONICAL_CLAIM=backend-specific corrected invariants only
```

### Prevention / future guard

Every high-risk remediation family must have a permanent executable owner. Closure markers must state exactly what they prove and must never be promoted from subsystem evidence to stage/release evidence without an explicit broader gate. When a later audit discovers an out-of-scope failure, amend the claim instead of retroactively pretending the earlier check covered it.
