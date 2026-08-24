# Document 68 — Stage 14.27 Flight Feature Timestamp Consistency

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: make exact Unix nanoseconds and PostgreSQL timestamp mirrors fail-closed and internally consistent

## 1. Correctness problem

`flight_feature_snapshots` stores each important instant twice:

```text
as_of_time timestamptz
as_of_time_unix_nano bigint
stored_at timestamptz
stored_at_unix_nano bigint
```

The PostgreSQL store wrote both representations from the same Go `time.Time`, but
all read queries selected only the Unix-nanosecond columns. A direct database
change, failed historical repair, or future writer could therefore make the
human-readable timestamp disagree with the exact timestamp without any runtime
error.

## 2. Canonical time policy

Unix nanoseconds remain the canonical exact representation for snapshot identity,
ordering, pagination, and returned Go values.

This choice is required because PostgreSQL `timestamptz` has microsecond precision,
while the feature contract and deterministic snapshot identity can contain
sub-microsecond values.

The `timestamptz` columns remain compatibility and operator-readable mirrors. They
are not independent sources of truth.

## 3. Read consistency boundary

Every insert return, direct read, latest read, and list read now selects both
representations.

The scanner:

```text
1. reconstructs the exact UTC instant from Unix nanoseconds;
2. normalizes the PostgreSQL mirror to UTC;
3. permits only the expected sub-microsecond precision difference;
4. rejects a difference of one microsecond or more as corrupt storage.
```

A mismatch returns `ErrCorruptSnapshot` with the responsible field:

```text
as_of_time
stored_at
```

Corrupt rows are never silently returned to Feature Pipeline, Route Intelligence,
or historical consumers.

## 4. Write consistency boundary

The insert path continues to derive both database values from one normalized UTC
instant. Snapshot uniqueness, exact lookup, latest ordering, and cursor filtering
continue to use `as_of_time_unix_nano`.

No independently supplied timestamp mirror is accepted by the application write
path.

## 5. Schema decision

No PostgreSQL migration is required.

The existing columns already support the contract, existing rows written by the
application are compatible, and removing the timestamp mirrors would be an
unnecessary destructive schema change. Runtime reads now make any future drift
visible and fail closed.

## 6. Precision examples

Accepted:

```text
exact Unix time:     2026-07-20T18:00:00.123456789Z
PostgreSQL mirror:   2026-07-20T18:00:00.123457Z
absolute difference: 211 nanoseconds
```

Rejected:

```text
exact Unix time:     2026-07-20T18:00:00.123456789Z
PostgreSQL mirror:   2026-07-20T18:00:00.123458Z
absolute difference: 1.211 microseconds
```

## 7. Regression protection

Tests protect:

```text
sub-microsecond PostgreSQL precision loss
one-microsecond mismatch rejection
as-of timestamp corruption reporting
stored timestamp corruption reporting
exact nanosecond reconstruction
all read-query timestamp mirror selection
Unix-nanosecond key and ordering ownership
PostgreSQL integration behavior when TEST_DATABASE_URL is available
```

The ownership test prevents a future query from returning to Unix-only scanning.

## 8. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/features/featurestore
go test -count=1 ./internal/features/featurestore
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
git diff --check
git status --short
```

## 9. Completion boundary

This increment closes the known Flight Feature timestamp and Unix-nanosecond
consistency debt. The only known PostgreSQL hardening item remaining in Document
58 is responsibility-based decomposition of the large PostgreSQL repository
surface.

## 10. Route and Historical extension

Document 72 applies the same exact-Unix-nanosecond and PostgreSQL-mirror policy to
Flight Route Results and Historical Aggregate Results. All read queries now select both
representations, scanners fail closed on one-microsecond drift, and migration 020 adds
database checks for every mirror pair.

## 11. Finding history, root cause, and failure scenario

### Finding

Feature snapshots persisted the same instant in exact Unix nanoseconds and PostgreSQL
`timestamptz`, but reads trusted only the exact column and never verified the mirror.

### Root cause

The duplicate representation was treated as write-time convenience rather than a
consistency contract. Because PostgreSQL and Go have different timestamp precision,
future writers or repairs could make the two columns disagree while normal reads remained
successful.

### Failure scenario

```text
snapshot row is written or later modified
↓
as_of_time_unix_nano represents instant A
as_of_time mirror represents instant B beyond expected microsecond rounding
↓
read path selects only Unix nanoseconds
↓
application returns snapshot normally while operator-visible/database mirror tells a different story
```

### Impact

Timestamp drift can compromise deterministic snapshot identity, auditability, debugging,
feature ordering, historical comparison, and any tooling that relies on the human-readable
mirror. Silent disagreement is particularly dangerous because both values look valid in
isolation.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1
identity/data consistency** because timestamp identity is used for exact ordering and
reproducibility, and disagreement could remain undetected across analytical consumers.

### Existing guarantees violated

```text
exact Unix nanoseconds are the canonical instant
human-readable PostgreSQL timestamps are faithful mirrors within known precision loss
corrupt dual representations fail closed
all read paths verify both representations
ordering/pagination continue to use exact time
```

## 12. Considered and rejected alternatives

### Remove the PostgreSQL timestamp mirror columns

Rejected because they remain useful for operators and compatibility, existing data already
uses them, and destructive schema removal is unnecessary to enforce consistency.

### Make `timestamptz` canonical and drop nanosecond identity

Rejected because PostgreSQL microsecond precision cannot represent all exact snapshot
instants used by deterministic feature identity/order contracts.

### Continue reading only Unix nanoseconds and treat mirrors as best effort

Rejected because silent mirror corruption would remain invisible and operational evidence
could disagree indefinitely.

### Require exact equality between `timestamptz` and nanoseconds

Rejected because normal PostgreSQL precision rounding can differ by sub-microsecond
amounts without representing corruption.

### Chosen remediation

Keep Unix nanoseconds canonical, select both representations on every read, permit only the
known sub-microsecond precision difference, and fail closed on larger mismatch.

## 13. Why this solution and trade-offs

The solution preserves exact identity and operator-readable timestamps while making their
relationship executable rather than assumed.

Trade-offs:

```text
+ exact nanosecond identity preserved
+ silent mirror drift becomes detectable corruption
+ operator-readable timestamps retained
- read queries/scanners carry both timestamp forms
- every read performs a small consistency check
- future precision-policy changes require coordinated tests
```

The extra scan/check cost is negligible relative to the value of deterministic analytical
evidence.

## 14. Adversarial review and remediation iterations

### Iteration 1 — identify one-sided reads

Review found that writes derived both columns from one instant but reads selected only
Unix nanoseconds, leaving mirror drift unobserved.

### Iteration 2 — dual-read fail-closed validation

Implementation commit `d76c526601a35ad7964fd9e93513396b0b4e4d6b`
(`fix: enforce flight feature timestamp consistency`) made every read verify both forms.

### Iteration 3 — precision-boundary challenge

Tests distinguish expected sub-microsecond PostgreSQL rounding from corruption at one
microsecond or more. This prevents an over-strict equality check from rejecting valid rows.

### Iteration 4 — broader persistence challenge

Document 72 later applied the same exact/mirror policy to Route and Historical result
stores and added database checks, demonstrating that the finding represented a reusable
persistence invariant rather than a one-off Feature Store patch.

## 15. Residual risks and limitations

The remediation detects disagreement; it does not automatically repair corrupt historical
rows. Any repair must choose the canonical exact value based on evidence and should be
reviewed explicitly. Clock-source correctness before persistence is a separate concern.

## 16. Operational/deployment consequences

Rows whose mirror differs beyond the allowed precision now fail reads with explicit
corruption evidence. Operators should investigate/repair those rows rather than bypass the
check. No schema migration was required for the initial Feature Store remediation.

## 17. Exact evidence

```text
implementation commit:
d76c526601a35ad7964fd9e93513396b0b4e4d6b

regression coverage:
internal/features/featurestore timestamp consistency tests
read-query ownership tests
PostgreSQL integration behavior

later extension:
Document 72 — Route/Historical timestamp mirror hardening
```

## 18. Final canonical status

```text
FINDING_GFA_DB_011_FEATURE_TIMESTAMP_CONSISTENCY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/68_STAGE_14_27_FLIGHT_FEATURE_TIMESTAMP_CONSISTENCY.md
IMPLEMENTATION_COMMIT=d76c526601a35ad7964fd9e93513396b0b4e4d6b
```

Historical PR/reviewer identifiers are not invented when unavailable.
