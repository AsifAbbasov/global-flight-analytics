# Document 61 — Stage 14.20 FlightTrajectory Read Snapshot Consistency

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: keep every aggregate FlightTrajectory read inside one PostgreSQL snapshot

## 1. Correctness problem

A FlightTrajectory is stored as one parent record plus child trajectory segments and
coverage gaps. The previous production repository loaded these parts through separate
pool queries:

```text
read flight_trajectories parent
↓
read trajectory_segments
↓
read coverage_gaps
```

Under PostgreSQL `READ COMMITTED`, another transaction could commit between those
queries. The returned aggregate could therefore combine a parent from one database
state with children from a later state.

Document 52 already established a caller-owned repeatable-read transaction for the
Projection Intelligence workflow. It did not make the core production trajectory
repository safe when used directly by HTTP handlers, route context, ingestion
continuation, or another service.

## 2. Repository-owned snapshot boundary

Both public aggregate read operations now enter the same boundary:

```text
GetLatestTrajectoryByICAO24
GetTrajectoryByID
↓
begin PostgreSQL transaction
isolation = REPEATABLE READ
access mode = READ ONLY
↓
read parent
read segments
read coverage gaps
↓
commit
```

The snapshot is owned by `TrajectoryRepository` whenever it was created from a
PostgreSQL pool. This makes consistency independent of the calling service and prevents
future composition roots from accidentally bypassing the protection.

## 3. Caller-owned transaction compatibility

`NewTrajectoryReadRepository` still supports binding the repository to an existing
`pgx` transaction. A transaction-bound repository has no pool ownership and therefore
reuses the caller snapshot without opening a nested transaction.

This preserves the Projection Intelligence transaction boundary and supports other
workflows that need to combine trajectory reads with additional evidence inside one
larger snapshot.

When `NewTrajectoryReadRepository` receives a pool, it records that pool as the snapshot
owner, so its public aggregate reads receive the same repository-owned protection as
`NewTrajectoryRepository`.

## 4. Failure behavior

The repository returns without a partial aggregate when:

```text
the transaction cannot start
the parent query fails
a child query fails
the caller context is cancelled
the read-only transaction cannot commit
```

An uncommitted transaction is rolled back through an independent bounded cleanup
context. Operation errors preserve their existing domain and repository semantics.

## 5. Concurrency evidence

The PostgreSQL integration test establishes a snapshot, reads one row, commits a second
row through another pooled connection, and reads again through the snapshot repository.
The second read still sees the original row count while a read after transaction commit
sees both rows.

This proves that the repository uses one repeatable-read snapshot rather than merely
issuing sequential queries on the same pool.

## 6. Regression protection

Permanent tests protect:

```text
both public aggregate reads enter withTrajectoryReadSnapshot
pool-backed reads use REPEATABLE READ
pool-backed reads use READ ONLY
parent and child queries receive the transaction-bound repository
transactions commit only after the aggregate is loaded
caller-owned transactions do not create nested transactions
pool-backed NewTrajectoryReadRepository retains snapshot ownership
concurrent commits do not change an active snapshot
```

The integration test is activated when `TEST_DATABASE_URL` is available and otherwise
skips without weakening the static architecture checks.

## 7. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/repository/postgres/trajectory_read_repository.go \
  internal/repository/postgres/trajectory_read_client.go \
  internal/repository/postgres/trajectory_read_snapshot.go \
  internal/repository/postgres/trajectory_read_snapshot_test.go \
  internal/repository/postgres/trajectory_read_snapshot_integration_test.go
go test -count=1 ./internal/repository/postgres
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
git diff --check
git status --short
```

## 8. Completion boundary

This increment closes production FlightTrajectory aggregate snapshot consistency.
It does not close:

```text
Ingestion Run transition integrity
trajectory relational constraints
shared migration filename parser
altitude precision policy
Traffic altitude-status semantics
timestamp and Unix-nanosecond consistency
large PostgreSQL repository decomposition
```

## 9. Finding history and root cause

### Finding

One logical FlightTrajectory aggregate was assembled by several independent pool reads
that did not share a PostgreSQL snapshot.

### Root cause

The repository modeled parent, segments, and coverage gaps as related queries but did not
model the aggregate read itself as one consistency boundary. Existing Projection code had
a caller-owned snapshot, which created a false sense that the lower-level repository was
safe for all callers.

### Failure scenario

```text
request reads trajectory parent at database state A
↓
another transaction commits new/replaced segment or coverage-gap data
↓
request reads children at database state B
↓
response combines parent from A with children from B
```

The result can be internally inconsistent even though every individual SQL query
succeeds.

### Impact

Mixed-snapshot aggregates can distort route context, historical analysis, ingestion
continuation, HTTP responses, and projection inputs. The defect is especially dangerous
because it produces plausible data rather than an obvious error.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1
analytical/data consistency** because inconsistent trajectory evidence can propagate into
higher-level analytics and predictions without failing loudly.

### Existing guarantees violated

```text
one logical aggregate is read from one coherent database state
repository callers do not need hidden transaction knowledge to obtain a valid aggregate
projection and non-projection paths observe compatible trajectory semantics
partial aggregate results are not returned after read failure
```

## 10. Considered and rejected alternatives

### Require every caller to open a transaction

Rejected because correctness would depend on composition-root discipline. New callers
could easily bypass the rule, which was the defect in the core repository boundary.

### Use `SERIALIZABLE` isolation

Rejected because the requirement is stable read visibility, not write-conflict
serialization. `REPEATABLE READ` supplies the needed snapshot with less coordination and
fewer avoidable retries.

### Read parent and children without a transaction but from one pooled connection

Rejected because connection affinity does not provide a stable snapshot under PostgreSQL
`READ COMMITTED`.

### Denormalize the full trajectory into one persisted row

Rejected because it would redesign storage and write paths to solve a read consistency
problem that PostgreSQL snapshots already solve directly.

### Chosen remediation

Make pool-backed public aggregate reads repository-owned `REPEATABLE READ`, `READ ONLY`
transactions while preserving caller-owned transaction reuse for larger workflows.

## 11. Why this solution and trade-offs

The repository is the narrowest layer that knows a FlightTrajectory requires several SQL
reads to form one domain aggregate. Putting the snapshot there prevents caller-specific
correctness drift.

Trade-offs:

```text
+ coherent parent/child aggregate reads
+ protection applies to current and future callers
+ existing caller-owned snapshots remain composable
- each pool-backed aggregate read opens a read-only transaction
- long-running aggregate reads can retain an older MVCC snapshot longer
- implementation must distinguish pool-owned from caller-owned transaction lifetime
```

The transaction overhead is accepted because correctness of analytical evidence is more
important than avoiding a bounded read-only transaction.

## 12. Adversarial review and remediation iterations

### Iteration 1 — projection-specific snapshot protection

Document 52 protected Projection Intelligence through a caller-owned repeatable-read
transaction.

### Iteration 2 — boundary challenge

Later review asked whether direct repository consumers received the same guarantee. They
did not; HTTP, route-context, ingestion-continuation, or future services could assemble a
mixed snapshot.

### Iteration 3 — repository-owned snapshot

Implementation commit `fcc601db509d8fb71d2f2db273548fec3832d3bd`
(`fix: enforce trajectory read snapshot consistency`) moved the default guarantee into
the production repository.

### Iteration 4 — nested-transaction challenge

The remediation preserved caller-owned transaction construction so Projection and other
larger workflows reuse an existing snapshot rather than accidentally nesting a second
transaction. Integration coverage then used a concurrent commit to prove snapshot
stability rather than relying only on source inspection.

## 13. Residual risks and limitations

The remediation guarantees snapshot consistency, not semantic correctness of the data
inside the snapshot. It does not prevent stale but coherent data, incorrect trajectory
construction, or manual database corruption.

Very long read transactions can retain PostgreSQL MVCC state longer; aggregate reads
should therefore remain bounded and should not perform unrelated remote work while the
snapshot is open.

## 14. Operational and deployment consequences

No schema migration is required. Database capacity planning should account for bounded
read-only transactions on aggregate trajectory reads. Monitoring should treat unusually
long transaction duration as an operational smell rather than weakening isolation.

## 15. Exact evidence

```text
implementation commit:
fcc601db509d8fb71d2f2db273548fec3832d3bd

regression coverage:
internal/repository/postgres/trajectory_read_snapshot_test.go
internal/repository/postgres/trajectory_read_snapshot_integration_test.go

protected implementation:
internal/repository/postgres/trajectory_read_repository.go
internal/repository/postgres/trajectory_read_snapshot.go
```

A historical pull-request number is not asserted because it is not preserved in the
currently searchable evidence. The implementation commit and concurrency integration test
are the canonical remediation evidence.

## 16. Final canonical status

```text
FINDING_GFA_DB_004_TRAJECTORY_READ_SNAPSHOT_CONSISTENCY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/61_STAGE_14_20_TRAJECTORY_READ_SNAPSHOT_CONSISTENCY.md
IMPLEMENTATION_COMMIT=fcc601db509d8fb71d2f2db273548fec3832d3bd
```
