# Historical Series Review Hardening

Status: implemented review remediation

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
