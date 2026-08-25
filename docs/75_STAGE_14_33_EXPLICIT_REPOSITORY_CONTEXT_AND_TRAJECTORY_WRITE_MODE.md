# Document 75 — Stage 14.33 Explicit Repository Context and Trajectory Write Mode

Status: Implemented current-scope baseline
Project: Global Flight Analytics
Scope: remove invented caller contexts from PostgreSQL repository operations and replace the Trajectory write empty-string mode sentinel with an explicit typed request

## 1. Confirmed defects

The Stage 14 maintainability register identified two related write-boundary ambiguities:

```text
selected PostgreSQL repository methods accepted a nil context and silently replaced it with context.Background()
TrajectoryRepository.saveTrajectory selected live versus reconciled behavior from an empty reconciliation task identifier
```

Both behaviors hid caller intent. Cancellation and deadlines belong to the caller, while live and reconciled persistence are distinct write modes that must be represented explicitly.

## 2. Caller-owned context contract

The shared PostgreSQL repository helper now rejects a nil context with:

```text
ErrRepositoryContextRequired
```

The contract is enforced by database-reaching paths for:

```text
Airport import writes
Airport keyset page reads
Airport ICAO reads
Flight State writes
Trajectory writes
```

Empty Airport Import and Flight State batches retain their existing no-operation behavior and return before opening a database transaction. Real database work never invents a background context.

The independent rollback helper from Document 72 remains intentionally different. Rollback is cleanup after caller cancellation and therefore continues to own a fresh bounded background context.

## 3. Explicit Trajectory write mode

Trajectory persistence now uses an internal typed request:

```text
trajectoryWriteModeLive
trajectoryWriteModeReconciled
trajectoryWriteRequest
```

`SaveTrajectory` constructs a live request. `SaveReconciledTrajectory` constructs a reconciled request only after normalizing the task identifier and validating the positive attempt count.

The coordinator branches on the explicit mode. It no longer interprets an empty task identifier as live persistence.

## 4. Validation behavior

The request validator rejects:

```text
an unknown zero-value or unsupported write mode
live mode carrying reconciliation metadata
reconciled mode without a normalized task identifier
reconciled mode without a positive attempt count
```

Trajectory relational-integrity and persisted-flight-identity validation still execute before transaction creation.

## 5. Preserved behavior

This increment does not change:

```text
public Trajectory Repository interfaces
live trajectory SQL
reconciled trajectory SQL
reconciliation ownership checks
replacement of an existing reconciled trajectory
segment and coverage-gap writes
transaction atomicity
independent rollback context
Airport pagination ordering
Flight State persistence mapping
PostgreSQL schema or migration history
```

## 6. Regression protection

Permanent tests and the Stage 14 source audit verify:

```text
repository operations do not contain ctx = context.Background()
all selected database-reaching methods call requireRepositoryContext
Trajectory writes use live and reconciled typed requests
Trajectory coordinator branching uses request.mode
empty reconciliation identifiers are not used as a mode switch
invalid and mixed write modes fail before persistence
Document 75 and the Stage 14.33 marker remain registered
```

## 7. Acceptance evidence

The installer requires, before modifying the real repository:

```text
targeted repository and audit tests
strict Stage 14 source audit
strict Backend Final Correctness Audit
go list ./...
all command builds
go vet ./...
complete go test ./...
```

The unified current-scope audit then runs backend, PostgreSQL, vulnerability, frontend, and container gates.

Successful installation prints:

```text
STAGE_14_33_EXPLICIT_CONTEXT_AND_WRITE_MODE=PASS
STAGE_14_CURRENT_SCOPE_AUDIT=PASS
STAGE_14_OVERALL_STATUS=REOPENED
INSTALLATION_COMPLETE=PASS
```

## 8. Completion boundary

This increment closes the confirmed nil-context fallback and implicit Trajectory write-mode findings for the named PostgreSQL repository paths.

Stage 14 remains reopened for nullable-helper and synthetic-source semantics, PostgreSQL query type ownership, migration-repair generalization, evidence-backed trajectory index profiling, and the final closure audit.

## 9. Canonical finding decomposition

```text
GFA-DB-021  Repository caller-context substitution
GFA-DB-022  Implicit Trajectory write-mode sentinel
```

## 10. GFA-DB-021 — Repository caller-context substitution

### Finding / symptom

Selected database-reaching repository methods treated `nil` as permission to create `context.Background()` and continue database work.

### Root cause

The repository attempted to be tolerant of incomplete callers instead of treating cancellation/deadline ownership as part of the API contract. That convenience obscured whether a caller intentionally supplied a lifecycle boundary.

### Failure scenario

A caller accidentally passes `nil`. Instead of failing immediately, the repository performs PostgreSQL work with no caller deadline or cancellation. Shutdown, request cancellation, or higher-level timeout policy can no longer control that operation.

### Impact

The primary risk is unbounded or detached database work, inconsistent cancellation semantics, and harder operational diagnosis. It can also make resource usage survive beyond the lifecycle of the request or job that initiated it.

### Severity rationale

**P2 retrospective.** The defect weakens production lifecycle and cancellation guarantees but is not evidence of silent persisted-data corruption by itself.

### Existing guarantees violated

- database work must inherit caller-owned cancellation/deadline intent;
- repositories must fail fast on missing required lifecycle inputs;
- background contexts are reserved for narrowly documented cleanup paths.

### Considered solutions

1. retain `nil` fallback;
2. generate a default timeout internally;
3. reject `nil` at database-reaching boundaries and leave timeout ownership to callers.

### Chosen remediation and why

`requireRepositoryContext` rejects `nil` with `ErrRepositoryContextRequired`. Real database work no longer invents a caller lifecycle. Empty no-op batches may return before needing a context because they do not reach PostgreSQL.

### Rejected alternatives

A default background timeout was rejected because the repository still would be inventing policy that belongs to the caller. Silent fallback was rejected because it preserves the ambiguity.

### Trade-offs

Callers must now satisfy an explicit contract and previously tolerated misuse becomes an error. This is an intentional compatibility tightening at an internal persistence boundary.

### Regression tests

Source audits forbid `ctx = context.Background()` on the selected production paths and require direct calls to `requireRepositoryContext`.

### Adversarial review and remediation iterations

The later global review in Document 79 found the same class of defect still present in the migrator. That post-closure finding demonstrates that Stage 14.33 intentionally covered named repository boundaries rather than proving a repository-wide/migrator-wide universal invariant. The later review expands the policy without rewriting this document's historical scope.

### Residual risk / limitations

Other packages can still misuse contexts unless they are covered by their own contracts or broader audits. Context correctness is a cross-cutting property that requires repeated structural review.

### Operational / deployment consequences

No schema change. Misconfigured callers fail earlier and more visibly instead of issuing detached database work.

### Exact evidence

Implementation commit: `211c774bb4820b6607bdbb6bd4e9cf17f1bc697b` (`refactor: enforce repository context and trajectory write mode`). Historical PR/reviewer metadata is not fabricated when unavailable.

### Final canonical status

**CLOSED for the named repository paths.** The later migrator finding is separately tracked in Document 79.

## 11. GFA-DB-022 — Implicit Trajectory write-mode sentinel

### Finding / symptom

The internal Trajectory write coordinator inferred live versus reconciled persistence from whether a reconciliation task identifier was empty.

### Root cause

Two distinct operation modes were encoded indirectly in the presence/absence of data rather than represented as an explicit type. This made an empty string both a value state and a control-flow signal.

### Failure scenario

A future call path can omit or normalize away reconciliation metadata and accidentally enter live persistence, or attach reconciliation metadata to a live write without the coordinator having a typed mode contract that rejects the combination.

### Impact

The ambiguity increases the risk of applying the wrong persistence workflow, ownership checks, or replacement semantics to a Trajectory write.

### Severity rationale

**P2 retrospective.** The design permits mode confusion at a correctness-sensitive persistence boundary, but no historical incorrect write is asserted by the evidence available.

### Existing guarantees violated

- live and reconciled writes are semantically distinct operations;
- operation mode should be explicit and validated;
- impossible mode/metadata combinations must fail before SQL.

### Considered solutions

1. keep empty-string sentinel and document it;
2. split into entirely separate repository implementations;
3. introduce an internal typed mode/request while preserving public methods and shared persistence machinery.

### Chosen remediation and why

`trajectoryWriteRequest` carries an explicit `trajectoryWriteMode`. Public methods construct only valid live or reconciled requests, and validation rejects unsupported or mixed combinations before persistence.

### Rejected alternatives

Documentation alone does not make illegal states unrepresentable. Separate repository implementations would duplicate substantial write logic and create a larger maintenance surface than the problem requires.

### Trade-offs

The coordinator gains an internal request type and validator. In return, control flow is explicit and malformed combinations are testable.

### Regression tests

Tests require typed live/reconciled requests, branch on `request.mode`, reject empty-id sentinels as control flow, and verify invalid mixed modes fail before persistence.

### Adversarial review and remediation iterations

The design preserves all existing SQL and reconciliation ownership while changing only the mode representation. This behavior-preserving constraint prevents a maintainability fix from silently altering persistence semantics.

### Residual risk / limitations

The type is package-internal; future new write modes still require explicit validation and tests. The current solution does not attempt to generalize to a plugin-style write-mode framework because no such need exists.

### Operational / deployment consequences

None. No schema or public API change.

### Exact evidence

Implementation commit: `211c774bb4820b6607bdbb6bd4e9cf17f1bc697b`.

### Final canonical status

**CLOSED.**

## 12. Prevention / future guard

Database-reaching code must not silently invent lifecycle policy for callers. Cleanup exceptions require a documented bounded rationale. Distinct persistence modes must be represented explicitly rather than encoded through empty strings, nil values, magic numbers, or other incidental data sentinels.
