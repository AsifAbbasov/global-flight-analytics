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

## 6. Canonical finding record — GFA-DB-030

### Finding / symptom

After Stage 14 closure, a global review found that normal migrator work still accepted a nil caller context in `ApplyPending` and `withMigrationLock` and silently substituted `context.Background()`.

### Root cause

The earlier context-hardening stages were scoped to selected repository boundaries and later Trajectory reads. The migrator retained an older convenience pattern and was not included in those package-local structural guards.

### Failure scenario

A migration caller passes nil. Instead of failing at the API boundary, the migrator can acquire a connection/advisory lock and execute migration work without caller cancellation or deadline ownership. A shutdown or orchestration timeout therefore cannot reliably govern the operation.

### Impact

The defect weakens lifecycle control around one of the highest-impact database operations in the project. Detached migration work can hold locks/connections or continue schema changes beyond the initiating caller's intended lifetime.

### Severity rationale

**P2 retrospective.** This is a production lifecycle/operational-control defect on migration execution, but the evidence does not show that it caused incorrect committed schema state. Migration atomicity and advisory locking remained separate valid guarantees.

### Existing guarantees violated

- normal database work must require caller-owned context;
- migration execution must not invent request/job lifecycle;
- background contexts are allowed only for bounded cleanup after caller cancellation;
- package-wide policy should not depend on a small list of string patterns.

### Considered solutions

1. preserve nil fallback because migrations are administrative;
2. add checks only to the two discovered functions;
3. define a migrator-wide required-context contract and a structural AST policy that scans all production files while permitting only named cleanup expressions.

### Chosen remediation and why

The migrator exposes `ErrMigrationContextRequired` and validates every database-reaching boundary. A second remediation iteration adds a Go AST audit that detects background/TODO/WithoutCancel usage, aliases/import variants, function-value storage, and caller-context reassignment across production migrator files.

### Rejected alternatives

Administrative status was not accepted as justification for detached work; migrations still need cancellation and orchestration ownership. Fixing only two string occurrences was rejected because the same pattern could return under a helper, alias, different formatting, or a newly added production file.

### Trade-offs

The AST audit is stricter and more maintenance-sensitive than a behavioral test alone. New legitimate cleanup helpers must be intentionally added to the narrow policy rather than using arbitrary background contexts. This friction is intentional at a migration boundary.

### Regression tests

Behavioral nil-context tests protect public/helper behavior. The AST gate protects structural context ownership across all production migrator files and recognizes syntactic variations that source-string tests would miss.

### Adversarial review and remediation iterations

The remediation chronology is directly visible in commit history:

```text
1c4a7bb992056e6b2c1d1394424643f913d31b00
fix: reject nil migrator contexts
↓
review of regression strength
↓
384f526474282a8ae63250fa36d8182eb342f772
test: enforce migrator context AST policy
```

The second commit is important evidence: the team did not stop at the literal bug fix; it hardened the guard against equivalent reintroductions.

### Residual risk / limitations

AST policy proves syntactic context ownership rules inside the migrator package. It cannot prove that every caller chooses an appropriate deadline or that external PostgreSQL/network cancellation is instantaneous. Cleanup can still fail if PostgreSQL is unavailable beyond its bounded timeout.

### Operational / deployment consequences

No schema migration is introduced. Callers that previously relied on nil fallback now fail immediately and must provide a context. Legitimate lock-release/rollback cleanup remains independently bounded so cancellation does not block cleanup.

### Exact evidence

Runtime fix: `1c4a7bb992056e6b2c1d1394424643f913d31b00` (`fix: reject nil migrator contexts`).

Structural regression hardening: `384f526474282a8ae63250fa36d8182eb342f772` (`test: enforce migrator context AST policy`).

Historical PR/reviewer identifiers are not asserted where repository evidence does not reliably recover them.

### Final canonical status

**CLOSED.** Stage 14 remains closed because this is a later, separately tracked residual finding whose remediation does not invalidate the explicitly scoped closure evidence in Document 78.

## 7. Why Stage 14 was not reopened

The existence of a later defect does not mean a prior closure audit claimed universal absence of all future defects. Document 78 explicitly limits closure to its executed repository boundaries. Document 79 therefore preserves both truths:

- Stage 14 satisfied its declared final closure gate;
- a later broader review found a residual context-policy gap outside the earlier scoped repository checks.

Reopening the historical stage would erase that distinction. Treating the new defect as an independent post-closure finding preserves more accurate engineering history.

## 8. Rejected status alternatives

Two status treatments were considered and rejected:

- **silently amend Document 78 as if the defect had been known before closure** — rejected because it falsifies chronology;
- **reopen all of Stage 14 automatically for any later defect** — rejected because closure was explicitly scoped and the new finding had an isolated remediation/evidence boundary.

The chosen treatment is a separately registered post-closure finding with its own fix and guard.

## 9. Prevention / future guard

Context ownership reviews should operate package-wide or repository-wide where practical rather than rely only on named methods. Structural policies should permit `context.Background()` only for explicitly bounded cleanup, reject lifecycle-detaching constructs in normal database work, and require new database-reaching entry points to opt into the same guard automatically.
