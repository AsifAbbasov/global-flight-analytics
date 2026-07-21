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
