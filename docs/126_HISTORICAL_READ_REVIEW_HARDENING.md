# Historical Read Review Hardening

Status: implemented review remediation

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
