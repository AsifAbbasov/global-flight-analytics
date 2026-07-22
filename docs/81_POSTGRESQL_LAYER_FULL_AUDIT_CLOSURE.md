# PostgreSQL Layer Full Audit Closure

## Status

The PostgreSQL Layer audit originally performed at commit `bb9f351` is closed against the current codebase.

Every finding is classified as one of the following:

- **fixed** through production code and executable tests;
- **not applicable** because the original implementation no longer exists;
- **deliberately rejected** as a mechanical rule that does not improve correctness.

No finding is left unclassified.

## Critical and integrity findings

| Original finding | Final status | Closure evidence |
| --- | --- | --- |
| Migration and history write were not atomic | fixed | one transaction, transaction-scoped advisory lock, rollback tests, concurrent runner test |
| Unsafe migration baseline | fixed | baseline command and runner behavior removed |
| Data Quality report could exist without Flight State | fixed | parent integrity migration, persisted-state query, integration tests |
| Trajectory aggregate read lacked one snapshot | fixed | repeatable-read snapshot client and aggregate read tests |
| Ingestion Run terminal transitions were uncontrolled | fixed | guarded transition predicates, database constraints and tests |
| Trajectory ownership constraints were incomplete | fixed | relational-integrity migration and repository contracts |
| Migration filename parsers disagreed | fixed | canonical migrationfile package used by migrator, audit and repair |
| Altitude precision policy was implicit | fixed | explicit domain and PostgreSQL integer conversion policy |
| NULL telemetry became semantic zero | fixed | availability and status semantics are preserved through domain, repository and HTTP layers |
| Traffic ignored altitude status | fixed | traffic altitude selection uses status semantics |
| Timestamp and Unix nanoseconds could diverge | fixed | consistency constraints and feature-store contract tests |

## Remaining structural findings closed by this increment

### Metrics query mode

The boolean `UseBounds` mode has been removed. `ActiveAircraftQuery` now owns an explicit `ActiveAircraftQueryScope`, and PostgreSQL uses separate global and bounded statements. This prevents a boolean flag from changing SQL behavior invisibly.

### Data Quality persistence mode

The empty reconciliation task identifier is no longer a sentinel. Live and reconciled writes use a typed `dataQualityWriteRequest` mode. A nil context is rejected instead of being replaced with `context.Background()`.

### Migration audit context

`migrationaudit.Auditor.Audit` rejects a nil context with `ErrContextRequired`. Cancellation semantics remain owned by the caller.

### Repository context and rollback coverage

Flight State reconciliation, Ingestion Run, reconciliation queue and Source HTTP validator repositories now reject a nil context through the shared repository context contract. No repository silently creates a background request. Trajectory snapshot rollback uses an explicitly named, bounded rollback context that is independent from request cancellation.

### Trajectory identifiers and naming

`ListTrajectoriesByEndTimeAndBounds` is renamed to `ListTrajectoriesWithinBounds`. Identifier queries pass native UUID values through ordered `unnest($1::uuid[])`; the UUID primary key is no longer converted to text and caller order remains preserved.

## Performance findings

Airport catalogue access is paginated through `ListPage` and a bounded maximum page size. The compatibility `List` method iterates those bounded pages.

Trajectory plans are verified by executable PostgreSQL tests using `EXPLAIN (ANALYZE, BUFFERS)`. Index assertions are retained for latest-by-aircraft, end-time ordering, segment ownership and coverage-gap ordering. Indexes are added only with plan evidence.

## Explicit architectural decisions

### Long methods

The proposal to split every function solely because it exceeds fifty lines is **deliberately rejected**. Methods are decomposed when they mix transactions, persistence ownership, validation or mapping responsibilities. A line-count threshold by itself is not a correctness contract.

### Words such as `And`

A repository-wide prohibition on the word `And` is **deliberately rejected**. The one ambiguous public trajectory method from the audit was renamed because its contract became clearer, not because a word is globally forbidden.

### Nullable adapter values

The former raw pointer helpers are no longer applicable. PostgreSQL adapters use explicit driver values and pgx nullable types. Database nullability remains inside the adapter boundary and is translated into domain availability semantics.

### Historical migration repair

`migrationrepair` is retained as a verification-only historical recovery tool. It is **not applicable** to production runtime composition and is prohibited from server, ingestion, reconciliation and service dependency trees.

## Permanent verification

The `postgreslayeraudit` tool and Stage 14 verifier enforce:

1. no boolean `UseBounds` SQL mode;
2. no Data Quality empty-task sentinel;
3. no nil-context fallback in PostgreSQL repositories or migration audit, including reconciliation, ingestion and Source HTTP validator paths;
4. native UUID trajectory queries;
5. mandatory provenance instead of fabricated `unknown` sources;
6. bounded airport pagination;
7. executable query-plan evidence;
8. migration repair isolation from production runtime;
9. a complete closure classification document.

The closure is evidence-based. It does not claim that the project can never acquire new technical debt; it closes the findings of the referenced PostgreSQL Layer audit.
