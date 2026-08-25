# Document 52 — Stage 14.12 Projection Read Snapshot Consistency

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: reproducible PostgreSQL input snapshot for one Projection Intelligence result

## 1. Problem

Projection Intelligence previously loaded its production inputs through four
independent database operations:

```text
Current Trajectory
Route
Historical Candidates
Route History
```

All operations used the same analytical `as_of_time`, but they did not share a
PostgreSQL transaction snapshot.

A concurrent ingestion, reconciliation, route materialization, or historical
backfill could therefore become visible between two reads. One projection
result could combine data from different committed database states.

The analytical time boundary prevented future observations from entering the
result, but it did not provide database snapshot consistency.

## 2. Production Decision

One projection request now performs one data-source operation:

```text
LoadSnapshot
```

The PostgreSQL implementation executes all required reads inside one
transaction configured as:

```text
isolation level: REPEATABLE READ
access mode: READ ONLY
```

This guarantees that every query in the projection input load observes one
stable PostgreSQL snapshot even when other transactions commit concurrently.

## 3. Snapshot Contents

The snapshot contains:

```text
current trajectory as of the requested analytical time
latest route result at or before the analytical time
route-scoped historical candidate trajectories
route-frequency history summary
```

A missing materialized route remains a valid analytical condition. The
snapshot contains the current trajectory and no route; the service then builds
the existing auditable unavailable-route result.

A missing route-history summary remains non-fatal and is represented as an
absent optional input.

## 4. Transaction-Scoped Trajectory Repository

Trajectory metadata, trajectory segments, coverage gaps, and flight-state
points must all use the same transaction.

The PostgreSQL trajectory repository now accepts a minimal read client that can
be either:

```text
a pgxpool.Pool
or
a pgx.Tx
```

Production snapshot loading creates a read-only trajectory repository bound to
the same `pgx.Tx` used by route, candidate, history, and point queries.

Write behavior of the existing trajectory repository is unchanged.

## 5. Service Boundary

The Projection Intelligence service no longer coordinates four storage calls.
Its data-source contract exposes only:

```text
LoadSnapshot(context, SnapshotRequest)
```

The service remains responsible for:

```text
request validation
unavailable-route domain semantics
composition policy
result error classification
```

The PostgreSQL adapter remains responsible for atomic data acquisition.

## 6. Transaction Lifecycle

Successful snapshot loading performs:

```text
BEGIN READ ONLY REPEATABLE READ
load all projection inputs
COMMIT
```

Any load error or commit failure triggers rollback cleanup.

No lock is taken on ingestion rows and no write statement is permitted by the
snapshot transaction.

## 7. Preserved Behavior

This increment does not change:

```text
projection formulas
confidence weights
historical-neighbor selection
arrival estimation
route-frequency policy
HTTP response contracts
SQL result ordering
persistence schema
migrations
provider behavior
frontend behavior
```

It changes only the consistency boundary around production reads.

## 8. Regression Gates

Automated tests require:

```text
the DataSource interface to expose one LoadSnapshot operation
Service.Get not to call independent load methods
production PostgreSQL composition to use the repeatable-read executor
transaction options to remain REPEATABLE READ and READ ONLY
successful reads to commit once
failed reads to roll back without commit
commit failures to receive rollback cleanup
snapshot clones not to share mutable slices
```

## 9. Acceptance

The increment is accepted only after:

```text
focused Projection Intelligence tests
PostgreSQL repository tests
snapshot transaction lifecycle tests
architecture regression tests
race detector
strict project architecture audit
complete Go build
go vet
complete Go test suite
frontend dependency security verification
frontend production dependency audit
ESLint
TypeScript validation
Next.js production build
backend Docker image build
git diff check
```

## 10. Canonical finding record — GFA-DB-050

### Finding / symptom

One Projection Intelligence result read four related PostgreSQL input groups through independent operations. They shared an analytical `as_of_time` but not one database snapshot.

### Root cause

The service-level data-source contract exposed separate reads for current trajectory, route, historical candidates, and route history. Analytical time filtering was mistaken for sufficient consistency even though PostgreSQL visibility could advance between calls.

### Failure scenario

```text
read current trajectory at committed DB state A
↓
concurrent ingestion/reconciliation/materialization commits state B
↓
read route/history/candidates at state B
↓
compose one projection from evidence that never coexisted in one committed snapshot
```

The projection can be temporally bounded yet internally inconsistent.

### Impact

Projection outputs, confidence, provenance, neighbor selection, route-frequency context, or ETA can be derived from mutually inconsistent evidence. The result remains syntactically valid, making the inconsistency difficult to detect downstream.

### Severity rationale

**P1 retrospective.** This is a production analytical consistency defect: one persisted/read result could combine different committed database states while appearing valid.

### Existing guarantees violated

- one projection result should be reproducible from one committed evidence snapshot;
- analytical `as_of_time` and database visibility are distinct guarantees;
- trajectory parent/segments/gaps/points and route/history evidence used together must share transactional visibility;
- data acquisition atomicity belongs to the PostgreSQL adapter, not ad hoc service sequencing.

### Considered solutions

1. keep independent reads and rely on close timing;
2. use explicit table/row locks across ingestion/materialization;
3. load all inputs in one read-only `REPEATABLE READ` transaction through a single `LoadSnapshot` operation.

### Chosen remediation

The Projection data source exposes one `LoadSnapshot`. PostgreSQL starts a `REPEATABLE READ`, `READ ONLY` transaction and binds the Trajectory repository plus route/candidate/history reads to the same `pgx.Tx` before commit.

### Why selected

Repeatable-read provides a stable view without blocking writers or changing analytical formulas. A single adapter-owned operation prevents future service composition from accidentally splitting the snapshot again.

### Rejected alternatives

Close-in-time reads do not prove consistent visibility. Row/table locking was rejected because the operation is read-only and should not block ingestion/backfill just to obtain a stable snapshot. Moving transaction coordination into the service would leak persistence mechanics across the application boundary.

### Trade-offs

Each projection input load now holds a read-only transaction for the duration of all required queries. That consumes one transactional connection/snapshot slightly longer than separate calls, but avoids writer blocking and inconsistent evidence.

### Regression tests / protection

Tests require one `LoadSnapshot` service call, exact `REPEATABLE READ`/`READ ONLY` options, transaction-bound Trajectory reads, commit/rollback behavior, commit-failure cleanup, and clone isolation. Later Document 61 extends snapshot ownership into direct Trajectory aggregate reads.

### Adversarial review findings

The review distinguished analytical time from database snapshot time. It also required trajectory children/points to use the same transaction rather than starting a transaction only for route/history queries while leaving trajectory reads pool-backed.

### Remediation iterations

```text
Stage 14.12: Projection workflow owns one repeatable-read snapshot
↓
Stage 14.20 / Document 61: core Trajectory aggregate repository gains its own snapshot guarantee for callers outside Projection
```

The later hardening supplements rather than invalidates this workflow-level boundary.

### Residual risks / limitations

Repeatable-read ensures database-state consistency, not semantic correctness of upstream observations or formulas. Long-running snapshots can retain MVCC state; production query duration still matters.

### Operational / deployment consequences

No migration/API change. Projection reads use one read-only transaction and one connection for the snapshot duration; PostgreSQL must support the configured isolation level (standard PostgreSQL behavior).

### Exact evidence

Implementation commit: `4f5920a25e6a5ba8e5a3f5db82fee8e7a90a5649` (`fix: enforce consistent projection read snapshot`). Historical PR/reviewer metadata is not invented where unavailable.

### Final canonical status

**CLOSED for Projection input snapshot consistency.** Direct Trajectory aggregate consistency outside this workflow is separately owned by GFA-DB-004 / Document 61.

### Prevention / future guard

Any analytical result composed from multiple related database reads must explicitly decide whether one committed snapshot is required. If yes, transaction ownership belongs in the persistence adapter and the application contract should expose one atomic load rather than several independently callable reads.
