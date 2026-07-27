# Historical Read Review Hardening

Status: closed

## Scope

This change hardens `apps/api/internal/historicalintelligence/historicalread` as the bounded PostgreSQL evidence boundary for Historical Intelligence.

## Corrected contracts

- All production datasets are read inside one read-only PostgreSQL `REPEATABLE READ` transaction.
- Flight and trajectory overlap predicates implement the half-open interval `[start, end)`.
- Mutable flight and trajectory rows are captured in append-only version tables and reconstructed at the analytical `AsOfTime`.
- Historical queries earlier than the version-history coverage boundary fail closed.
- Route membership is selected by trajectory event time rather than route calculation time.
- The latest admissible route version per trajectory is selected before the global row limit.
- Exact matched-row counts provide the coverage denominator; `limit + 1` remains only a bounded pagination sentinel.
- Route JSON has both a total byte budget and an explicit payload fingerprint.
- Route payload decoding is owned by the Historical Read boundary; downstream builders use decoded route results.
- Nullable identifiers retain explicit availability evidence instead of being erased by `COALESCE`.
- Numeric-to-float conversion uses explicit PostgreSQL rounding: twelve decimal places for quality scores and eight for coordinates.
- Alternative transaction executors are validated through the same record invariants as production reads.

## Compatibility decisions

`nil, error` constructor returns and nil-safe error unwrapping remain idiomatic Go and are not treated as defects. PostgreSQL-specific dependencies remain inside the PostgreSQL adapter. The old raw `RouteJSON` field is retained only for legacy test-fixture compatibility; production reads clear it after decoding, and downstream builders no longer parse persistence JSON.

## Database impact

Migration `028_harden_historical_read_snapshot.sql` creates temporal version history, capture triggers, the history coverage marker, and query-aligned indexes. Existing rows are backfilled as their currently known versions. Earlier overwritten states cannot be reconstructed retroactively, so an `AsOfTime` earlier than the migration coverage boundary is rejected.

## Formal closure evidence

The engineering remediation was completed by:

- Hardening commit: `b67546391984b4726e05d67a51471d401f921e29` (`fix: harden historical read integrity`).
- Final engineering commit: `98750a7eb5972cd770e6333f46cd0855eca8ad0e` (`test: align historical read fixtures with integrity`).

GitHub Actions run `30298888993` for commit `98750a7eb5972cd770e6333f46cd0855eca8ad0e` completed with all required jobs successful after the PostgreSQL infrastructure-only retry:

```text
Backend Quality=SUCCESS
Backend Race Safety=SUCCESS
PostgreSQL 16 Integration=SUCCESS
Backend Container=SUCCESS
```

The successful PostgreSQL job applied and verified the production migration catalog, ran the PostgreSQL correctness integration tests, and verified the PostgreSQL feature pipeline.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_READ_REVIEW_STATUS=CLOSED
```
