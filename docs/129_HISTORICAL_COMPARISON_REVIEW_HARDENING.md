# Historical Comparison Review Hardening

Status: implemented review remediation

## Scope

This increment hardens `apps/api/internal/historicalintelligence/historicalcomparison` as the atomic domain boundary for adjacent-period Historical Intelligence comparisons.

## Accepted findings

- The previous period could be partial while the current period was complete, producing a mathematically valid but domain-misleading percentage from incomparable coverage.
- Direct `Attach` returned a derived result whose builder version, fingerprint, source union, and latest source time described only the current period until the materializer performed a later optional repair.
- `Attach` mixed source validation, compatibility checks, value selection, arithmetic, result construction, provenance, and final validation.
- Scope identity depended on `reflect.DeepEqual`, so future non-semantic fields could silently alter comparison behavior.
- The package exported the generic `Values` helper without an external use case.
- Percentage overflow was detected only by final contract validation and was misclassified as an invalid current source result.
- Regression coverage did not prove mismatch classification, coverage comparability, all aggregation selectors, downward and flat direction, direct provenance, arithmetic overflow, nested comparison rejection, or previous-period limitations.

## Corrected contracts

- Current and previous series must have identical top-level status, point count, bucket duration, bucket status, and per-bucket coverage ratio within the Historical Contract ratio tolerance.
- Complete-versus-partial and differently partial series fail closed with `ErrCoverageMismatch`; a coverage difference cannot masquerade as metric growth.
- Matched partial periods remain comparable. The current-series confidence stays contract-consistent, while a comparison-scoped quality limitation records both statuses, confidence scores, sample counts, previous-period limitations, and the matched-partial constraint.
- `Attach` now constructs a standalone provenance atomically. Builder identity includes Historical Comparison plus both source builders; semantic fingerprinting binds both complete source results; sources are merged and sorted; latest source time is the later period evidence time.
- Nested comparison attachment is rejected so a derived result cannot be treated as a raw source period.
- Scope comparison uses the explicit domain method `Scope.Equal`.
- Value selection uses a focused aggregation selector registry and the helper type is package-private.
- Percentage calculation preserves full finite `float64` domain precision, treats a zero previous value as explicitly undefined, and rejects non-finite arithmetic with `ErrComparisonArithmeticInvalid`.
- Final derived-result validation uses `ErrComparisonResultInvalid`, keeping source validity and comparison construction failures distinct.

## Deliberately retained contracts

- The module does not process money. Historical metrics are exact counts within the IEEE-754 exact-integer boundary, bounded ratios, or continuous aviation measurements. Replacing the project-wide `float64` transport with a decimal library would be a breaking cross-contract migration without a domain benefit here.
- Percentage values are not rounded in the domain. The finite full-precision value is stored; presentation layers may apply display rounding without changing analytical identity.
- `PercentageChange == nil` remains the explicit contract representation for an undefined percentage when the previous value is zero. Replacing it with a new object would break persistence and API schemas without improving the mathematical state.
- `Summary.Average` is the unweighted temporal mean of represented bucket values. `Summary.Median` is the temporal median of represented bucket values. Neither field claims to be an observation-weighted mean or a raw-observation median.
- The nil receiver behavior of `ResultValidationError.Unwrap` remains idiomatic: a nil error wrapper has no underlying error.
- A finite enumeration switch is not inherently an Open/Closed Principle violation. The comparison module nevertheless now uses a selector registry to keep the orchestration path focused.

```text
COMPARISON_COVERAGE_PROFILE_MATCH=ENFORCED
PREVIOUS_PERIOD_QUALITY=BOUND
COMPARISON_PROVENANCE=ATOMIC
COMPARISON_FINGERPRINT_BOTH_PERIODS=BOUND
COMPARISON_ARITHMETIC_FINITE=ENFORCED
SCOPE_EQUALITY=EXPLICIT
PERCENTAGE_ZERO_BASE=UNDEFINED_OPTIONAL
TEMPORAL_SUMMARY_SEMANTICS=DOCUMENTED
HISTORICAL_COMPARISON_ENGINEERING_REMEDIATION=IMPLEMENTED
```

## Permanent verification

`apps/api/tools/historicalcomparisonreviewaudit` protects the version-two comparison boundary, explicit scope equality, coverage-profile compatibility, explicit two-period quality evidence, atomic provenance, both-period fingerprint identity, finite percentage arithmetic, package-private helper surface, regression tests, and this review record in Backend Continuous Integration.
