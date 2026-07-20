# Document 68 — Stage 14.27 Flight Feature Timestamp Consistency

Status: Implemented
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
