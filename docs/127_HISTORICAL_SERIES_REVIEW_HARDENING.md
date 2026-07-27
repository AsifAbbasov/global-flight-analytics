# Historical Series Review Hardening

Status: closed

## Scope

This change hardens `apps/api/internal/historicalintelligence/historicalseries` as the generic time-series assembly boundary for Historical Intelligence.

## Accepted findings

- A single global coverage ratio was incorrectly copied into every bucket.
- Missing source and generation timestamps were replaced with synthetic window timestamps.
- Malformed limitations were silently deleted.
- Distinct window exclusions with the same reason could collapse into one limitation.
- Total sample accumulation was not protected from integer overflow.
- The main build function mixed validation, provenance, point construction, status derivation, confidence, and final contract validation.

## Corrected contracts

- Every `BucketValue` carries explicit `CoverageEvidence`.
- Dataset consumers bind complete or incomplete read evidence before series construction.
- Incomplete reads produce a conservative per-bucket lower bound and mark buckets without represented evidence unavailable.
- Series status is derived from bucket statuses rather than one floating-point state surrogate.
- Series confidence is the mean temporal coverage across planned buckets; sample count remains evidence volume rather than an artificial confidence multiplier.
- `LatestSourceUpdatedAt` and `GeneratedAt` are required and are never synthesized from the analytical window.
- Limitations fail closed when malformed or duplicated.
- Window exclusion limitations include unique codes and exact excluded intervals.
- Total sample count uses checked integer addition.
- `Build` delegates validation, point construction, status derivation, confidence, and final validation to focused helpers.
- Historical Traffic, Airport, and Route builders use the new coverage binding contract.
- A permanent strict audit is part of Backend Continuous Integration.

## Rejected or already-closed findings

- Zero coverage already produced an unavailable series before this increment.
- Mutable `historicalwindow.Plan` evidence was already canonicalized and validated by the Historical Window hardening.
- Generic `float64` value transport remains intentional: the central metric catalog rejects fractional count values and values beyond the exact integer boundary.
- A complete, fully observed empty bucket may legitimately have confidence `1` with sample count `0`; sample count is not source completeness.
- Nil-safe error formatting remains idiomatic Go and is not treated as a domain null state.
- Direct access to the canonical `historicalwindow.Plan` value object is an intentional adjacent-domain contract, not a Law of Demeter violation.

## Formal closure evidence

The Historical Series engineering remediation was completed by:

- Baseline commit: `0cc0dc7be96860120d69c8dd53158b36ac72d9c6`.
- Hardening commit: `02bee5fd59d13927d0ffb995844c83d07327a2f9` (`fix: harden historical series integrity`).
- Final engineering compatibility commit: `c863d03e5de711b78ab94027dbf951129665c110` (`fix: align historical contract audit with series hardening`).

GitHub Actions run `30305541816` for commit `c863d03e5de711b78ab94027dbf951129665c110` completed with all required jobs successful:

```text
Backend Quality=SUCCESS
Backend Race Safety=SUCCESS
PostgreSQL 16 Integration=SUCCESS
Backend Container=SUCCESS
```

Backend Quality executed the Historical Contract, Historical Window, Historical Read, and Historical Series strict audits successfully. The repository-level tests, Go vet analysis, race-safety checks, PostgreSQL integration checks, and backend container build were successful.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_SERIES_ENGINEERING_DEBT=CLOSED
HISTORICAL_SERIES_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
HISTORICAL_SERIES_REVIEW_STATUS=CLOSED
```
