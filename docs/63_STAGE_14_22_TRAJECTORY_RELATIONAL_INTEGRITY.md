# Document 63 — Stage 14.22 Trajectory Relational Integrity

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: enforce complete PostgreSQL integrity for FlightTrajectory aggregates

## 1. Correctness problem

A FlightTrajectory is stored across three tables:

```text
flight_trajectories
trajectory_segments
coverage_gaps
```

Before this increment, child `trajectory_id` values could be null, segment sequence
numbers were not unique per trajectory, a coverage gap could reference a segment
owned by another trajectory, and parent counters were not required to match the
actual child rows.

Those gaps allowed a database state that could not represent one coherent domain
aggregate even when each individual row was syntactically valid.

## 2. Repository fail-fast boundary

`SaveTrajectory` now validates the aggregate before opening a PostgreSQL
transaction. It rejects:

```text
segment_count different from len(segments)
coverage_gap_count different from len(coverage_gaps)
point_count different from the sum of segment point_count values
point_count different from an available in-memory point collection
non-contiguous segment sequence numbers
negative stored counts
segment or coverage-gap ICAO24 different from the parent ICAO24
```

This is an early diagnostic boundary. PostgreSQL remains the authoritative
integrity boundary.

## 3. Child parent identity

Migration 018 makes `trajectory_id` mandatory for both child tables.

```text
trajectory_segments.trajectory_id → NOT NULL
coverage_gaps.trajectory_id → NOT NULL
```

Deleting a parent continues to use the existing cascade behavior. A child can no
longer exist as an unowned trajectory fragment.

## 4. Segment ordering

The database now requires:

```text
sequence_number > 0
UNIQUE (trajectory_id, sequence_number)
```

The deferred aggregate verifier additionally requires the final sequence to be
contiguous from `1` through the actual segment count.

## 5. Coverage-gap segment ownership

Each previous or next segment reference is protected by a composite foreign key:

```text
(trajectory_id, previous_segment_id)
    → trajectory_segments(trajectory_id, id)

(trajectory_id, next_segment_id)
    → trajectory_segments(trajectory_id, id)
```

A segment from another FlightTrajectory can no longer be attached to the gap even
when its standalone identifier exists.

## 6. Deferred aggregate verification

Constraint triggers run at transaction completion so the repository may continue
to insert the parent first and its children afterward.

At the final database state, PostgreSQL verifies:

```text
stored segment_count = actual segment rows
stored point_count = sum of segment point_count
stored coverage_gap_count = actual coverage-gap rows
segment sequence numbers are contiguous
all segment ICAO24 values match the parent
all coverage-gap ICAO24 values match the parent
```

The same protection applies to direct SQL and to future repository code.

## 7. Legacy-data policy

Migration 018 performs a complete preflight before adding constraints. It aborts
without silently rewriting data when it finds:

```text
unowned child rows
duplicate or non-positive sequence numbers
cross-trajectory gap references
stored-count mismatches
sequence gaps
child identity mismatches
```

Repairing such rows requires an explicit evidence-backed migration rather than an
automatic guess.

## 8. Regression protection

The permanent tests cover:

```text
canonical in-memory aggregate acceptance
stored-count mismatch rejection
sequence-gap rejection
child identity rejection
point-total rejection
validation before transaction creation
PostgreSQL NOT NULL enforcement
per-trajectory sequence uniqueness
same-trajectory gap references
stored counter and contiguous sequence verification
```

PostgreSQL integration tests run when `TEST_DATABASE_URL` is configured. They skip
without falsely reporting runtime database evidence when it is absent.

## 9. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/repository/postgres/trajectory_repository.go \
  internal/repository/postgres/trajectory_relational_integrity.go \
  internal/repository/postgres/trajectory_relational_integrity_test.go \
  internal/repository/postgres/trajectory_relational_integrity_integration_test.go
go test -count=1 ./internal/repository/postgres
go test -count=1 ./internal/database/migrationaudit ./internal/database/migrator
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
git diff --check
git status --short
```

## 10. Completion statement

This increment closes the known FlightTrajectory relational-integrity debt for
child ownership, segment ordering, same-trajectory gap references, child identity,
and parent stored counters.

## 11. Finding history, root cause, and failure scenario

### Finding

The relational schema allowed trajectory fragments that were individually valid rows but
could not form one valid FlightTrajectory aggregate.

### Root cause

Domain-level aggregate invariants were partially represented as repository conventions and
stored counters, but not fully encoded as relational constraints. Child ownership,
sequence uniqueness, cross-reference ownership, and aggregate-count consistency were
therefore separable from the parent identity.

### Failure scenario

```text
trajectory T1 exists
trajectory T2 exists
↓
a coverage gap for T1 references a segment owned by T2
and/or
segment sequence numbers for T1 are duplicated or non-contiguous
and/or
stored segment_count / point_count / gap_count no longer matches child rows
↓
queries successfully read rows that cannot describe one coherent trajectory
```

### Impact

Malformed aggregates could corrupt trajectory reconstruction, route context, historical
analytics, feature extraction, reconciliation, and projection inputs while remaining
syntactically valid in PostgreSQL.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1 data
integrity** because the database could durably represent impossible trajectory aggregates
that downstream analytics would treat as canonical evidence.

### Existing guarantees violated

```text
every child belongs to exactly one parent trajectory
segment order is positive, unique, and contiguous within a trajectory
gap references cannot cross trajectory ownership
stored aggregate counters match durable child rows
child ICAO24 identity matches parent identity
```

## 12. Considered and rejected alternatives

### Keep validation only in `SaveTrajectory`

Rejected because direct SQL, migration code, future repositories, or bugs in alternate
writers could bypass application validation.

### Remove stored counters instead of validating them

Rejected because counters are part of existing read/performance contracts and can remain
useful if PostgreSQL verifies that they agree with the aggregate.

### Normalize coverage-gap references into globally unique segment IDs only

Rejected because global identifier existence does not prove same-trajectory ownership.
Composite foreign keys express the actual domain relationship.

### Rewrite legacy invalid rows automatically during migration

Rejected because sequence, identity, or cross-trajectory repairs would require guessing
historical intent. Migration 018 fails closed for explicit evidence-backed repair.

### Chosen remediation

Use fail-fast aggregate validation in the repository plus authoritative PostgreSQL
constraints and deferred aggregate verification at transaction completion.

## 13. Why this solution and trade-offs

This duplicates selected checks intentionally at two boundaries: application validation
improves diagnostics before work begins; PostgreSQL guarantees durable correctness for all
writers.

Trade-offs:

```text
+ invalid aggregate states become unrepresentable after transaction commit
+ direct SQL/future writers receive the same protection
+ early repository errors remain understandable
- migration 018 and deferred constraint triggers increase schema complexity
- writes perform additional integrity work before commit
- legacy inconsistent data blocks deployment until explicitly repaired
```

The extra validation cost is accepted because trajectory writes are correctness-sensitive
and impossible aggregates are more expensive to diagnose later.

## 14. Adversarial review and remediation iterations

### Iteration 1 — repository aggregate validation

Repository validation made known counter, ordering, and identity mismatches fail before
transaction creation.

### Iteration 2 — bypass challenge

Review considered direct SQL and future writer paths. Application validation alone was
not sufficient, so the durable constraints were added in migration 018.

### Iteration 3 — ownership challenge

A standalone segment foreign key was not enough for coverage gaps because it could point
to a real segment belonging to another trajectory. Composite `(trajectory_id, segment_id)`
keys were used instead.

### Iteration 4 — legacy repair challenge

Migration preflight deliberately aborts instead of guessing how to repair conflicting
historical trajectory evidence.

Implementation commit:
`5bb4a6aab7b16bc13e8477ca31f11eaa27e808ff`
(`fix: enforce trajectory relational integrity`).

## 15. Residual risks and limitations

Relational integrity does not prove that observed trajectory points are physically or
semantically correct; it proves that persisted aggregate structure is internally
consistent. Administrator-level constraint disabling can bypass the contract.

Large aggregate validation also remains bounded by normal transaction/write performance
considerations; profiling should guide future optimization rather than weakening the
constraints.

## 16. Operational/deployment consequences

Migration 018 performs fail-closed preflight. Any legacy violation requires explicit
operator repair before deployment proceeds. Constraint-trigger failures should be treated
as integrity findings, not retried blindly as transient database errors.

## 17. Exact evidence

```text
implementation commit:
5bb4a6aab7b16bc13e8477ca31f11eaa27e808ff

migration:
018_trajectory_relational_integrity.sql

regression coverage:
internal/repository/postgres/trajectory_relational_integrity_test.go
internal/repository/postgres/trajectory_relational_integrity_integration_test.go
```

## 18. Final canonical status

```text
FINDING_GFA_DB_006_TRAJECTORY_RELATIONAL_INTEGRITY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/63_STAGE_14_22_TRAJECTORY_RELATIONAL_INTEGRITY.md
IMPLEMENTATION_COMMIT=5bb4a6aab7b16bc13e8477ca31f11eaa27e808ff
```

Historical PR/reviewer identifiers are not invented where the searchable repository does
not preserve them.
