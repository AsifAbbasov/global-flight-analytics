# Document 69 — Stage 14.28 PostgreSQL Trajectory Repository Decomposition

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: separate the Trajectory Repository by persistence responsibility while preserving one public repository contract

## 1. Confirmed maintainability problem

The production Trajectory Repository was already behaviorally correct, but two
source files had accumulated independent responsibilities:

```text
trajectory_repository.go: 548 lines
trajectory_read_repository.go: 459 lines
```

The write source mixed repository construction, public write operations,
transaction orchestration, reconciliation attempt ownership, parent inserts,
segment inserts, coverage-gap inserts, and segment-reference inference.

The read source mixed snapshot coordination, parent selection, parent scanning,
aggregate child loading, segment reading, coverage-gap reading, and lookup
normalization.

This was a cohesion finding, not a line-count-only refactor.

## 2. Preserved public contract

The type `TrajectoryRepository` remains the production entry point. The following
public methods retain their names and signatures:

```text
SaveTrajectory
SaveReconciledTrajectory
GetLatestTrajectoryByICAO24
GetTrajectoryByID
ListTrajectorySegments
ListCoverageGaps
```

Callers, interfaces, composition roots, HTTP handlers, reconciliation workers,
feature materialization, and Projection Intelligence require no changes.

## 3. Write-side ownership

Write responsibilities are now separated into:

```text
trajectory_repository.go
  repository state and constructor only

trajectory_write_repository.go
  public write methods and transaction coordination

trajectory_reconciliation_write.go
  reconciliation attempt ownership and replacement cleanup

trajectory_parent_write.go
  direct and reconciled parent trajectory inserts

trajectory_segment_write.go
  ordered trajectory-segment persistence

trajectory_gap_write.go
  coverage-gap persistence and inferred segment references
```

Relational integrity and flight identity validation still execute before the
PostgreSQL transaction begins. Reconciled writes still verify the active attempt
under row lock before replacing the previous derived trajectory.

## 4. Read-side ownership

Read responsibilities are now separated into:

```text
trajectory_read_repository.go
  public repeatable-read snapshot coordination only

trajectory_parent_read.go
  parent selection, parent scanning, and ICAO24 normalization

trajectory_child_read.go
  aggregate child-loading coordination

trajectory_segment_read.go
  ordered segment query and mapping

trajectory_gap_read.go
  ordered coverage-gap query and mapping
```

Both public aggregate reads still use the repository-owned read-only
`REPEATABLE READ` boundary. A transaction-bound repository still reuses its
caller-owned snapshot and does not create a nested transaction.

## 5. Deliberately unchanged behavior

This increment does not change:

```text
SQL statements
transaction isolation or access mode
commit and rollback behavior
reconciliation ownership rules
relational integrity rules
trajectory identity rules
row ordering
nullable mapping
not-found semantics
public interfaces
HTTP contracts
PostgreSQL schema or migrations
```

## 6. Regression protection

Permanent tests require:

```text
constructor-only trajectory_repository.go
snapshot-coordinator-only trajectory_read_repository.go
dedicated owners for write orchestration, reconciliation, parents, segments, and gaps
dedicated owners for parent reads, child loading, segment reads, and gap reads
both public aggregate reads to retain withTrajectoryReadSnapshot
relational validation to remain before BeginTx
former monolithic SQL and mapping responsibilities not to return to coordinator files
```

The existing snapshot, reconciliation, relational-integrity, repository, and
PostgreSQL integration tests remain the behavioral evidence.

## 7. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/repository/postgres/trajectory_*.go
go test -count=1 ./internal/repository/postgres
go test -count=1 ./...
go vet ./...
```

From `apps/web`:

```bash
pnpm typecheck
pnpm lint
pnpm build
```

From the repository root:

```bash
git diff --check
git status --short
```

## 8. Completion boundary

This increment closes the final known PostgreSQL maintainability debt recorded
in Document 58. Stage 14 still requires one final repository-wide audit before
its correction programme can be declared complete.

## 9. Finding history, root cause, and failure scenario

### Finding

The Trajectory Repository remained behaviorally correct but concentrated unrelated read
and write responsibilities into two large coordination files, increasing the cost and
risk of future changes.

### Root cause

Repository evolution added reconciliation, relational integrity, snapshot consistency,
parent/child persistence, and mapping incrementally while preserving one public type. The
public contract stayed cohesive, but internal ownership boundaries did not evolve at the
same pace.

### Failure scenario

This finding was maintainability-oriented rather than an observed production corruption
scenario. The credible failure path was:

```text
future change targets one responsibility
↓
engineer edits a large file containing transaction, SQL, mapping, and reconciliation logic
↓
unrelated invariant is accidentally changed or duplicated
↓
regression escapes local reasoning because ownership is unclear
```

### Impact

High cognitive load and mixed responsibilities make correctness-sensitive PostgreSQL code
harder to review, test, and modify safely. The risk grows as trajectory features evolve,
even when the current implementation is already correct.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P3
maintainability/cohesion**. The stage itself states that behavior was already correct; this
was not a data-loss or availability incident.

### Existing guarantees at risk

```text
transaction ownership remains obvious
snapshot consistency is not mixed with mapping concerns
reconciliation ownership is localized
SQL/mapping responsibilities have clear source owners
public repository contract remains stable while internals evolve
```

## 10. Considered and rejected alternatives

### Split by arbitrary line-count threshold

Rejected because file length alone does not define architecture. The stage explicitly
uses persistence responsibility as the split criterion.

### Introduce interfaces for every internal helper

Rejected because the code is one package and no alternate implementations were needed.
Interfaces would add ceremony without improving ownership.

### Create separate public repositories for every table

Rejected because callers need one FlightTrajectory aggregate contract and should not
coordinate parent/segment/gap persistence themselves.

### Introduce a dependency-injection container or service locator

Rejected because the problem was source cohesion, not runtime construction complexity.

### Chosen remediation

Keep one public `TrajectoryRepository` contract and decompose implementation files by
actual persistence responsibility inside the same package.

## 11. Why this solution and trade-offs

The solution reduces cognitive surface while preserving all proven runtime behavior and
without creating new abstraction layers.

Trade-offs:

```text
+ smaller responsibility-focused files
+ easier targeted review and ownership tests
+ public API and SQL behavior remain unchanged
+ no new runtime indirection
- more source files to navigate
- same-package private coupling still exists where responsibilities legitimately interact
```

The additional file count is acceptable because it maps directly to distinct persistence
responsibilities rather than artificial architectural layers.

## 12. Adversarial review and remediation iterations

### Iteration 1 — classify the problem correctly

Review did not treat line count alone as a defect. It identified mixed construction,
transaction, reconciliation, parent/child SQL, snapshot, and mapping responsibilities.

### Iteration 2 — behavior-preserving decomposition

Implementation commit `24aa9a41abf4b5048e207c72d6aa4f93ab86319a`
(`refactor: decompose postgres trajectory repository`) moved responsibilities into focused
files while retaining the same public type and contracts.

### Iteration 3 — architecture regression guard

Ownership tests prevent SQL/mapping/orchestration responsibilities from accumulating back
inside coordinator files and require snapshot/integrity boundaries to remain in place.

### Iteration 4 — avoid abstraction overreaction

The remediation deliberately stopped at same-package source ownership. No repository
microservice, generic persistence framework, or interface hierarchy was introduced because
none was required to solve the finding.

## 13. Residual risks and limitations

Decomposition does not guarantee future maintainability automatically. New responsibilities
can still accumulate, and cross-file coupling should be reviewed when actual evidence
appears. File count should not become a metric target in its own right.

## 14. Operational/deployment consequences

None. No SQL, schema, runtime configuration, public API, or deployment behavior changes.
This is an internal maintainability remediation protected by existing behavioral tests.

## 15. Exact evidence

```text
implementation commit:
24aa9a41abf4b5048e207c72d6aa4f93ab86319a

regression evidence:
trajectory repository architecture/ownership tests
existing snapshot, relational-integrity, reconciliation, and PostgreSQL integration tests
```

## 16. Final canonical status

```text
FINDING_GFA_MAINT_012_TRAJECTORY_REPOSITORY_COHESION=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/69_STAGE_14_28_POSTGRES_TRAJECTORY_REPOSITORY_DECOMPOSITION.md
IMPLEMENTATION_COMMIT=24aa9a41abf4b5048e207c72d6aa4f93ab86319a
```

Historical PR/reviewer identifiers are not invented when unavailable.
