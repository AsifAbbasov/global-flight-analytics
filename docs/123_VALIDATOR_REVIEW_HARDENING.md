# Validator Review Hardening

## Scope

This increment hardens `apps/api/internal/features/validator` as the production trust gate for Flight Feature snapshots at baseline `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2`.

## Confirmed findings

The prior severity policy downgraded relationship and numerical integrity failures in partial groups to warnings. Since the Feature Pipeline persists both valid and limited snapshots, non-finite values and contradictory relationships could reach the Feature Store. Quality limitations were also merged with earlier validator output, allowing stale validation findings to survive a corrected revalidation.

## Rejected or stale findings

The current materialization path no longer rejects ordinary trajectories merely because `TrajectoryUpdatedAt` is later than `AsOfTime`; that validator rule is absent. The geographical schema already contains fifteen authoritative analytical fields, while `GeographicCellPrecision` is an intentional processing-configuration mirror outside schema counts. Longitude span reconciliation is already implemented through the circular longitude envelope policy. Returning a nil pointer with a non-nil constructor error remains idiomatic Go and is not a domain-integrity violation.

## Implemented policy

- Mathematical integrity is always error-severity, regardless of `available` or `partial` status.
- Evidence incompleteness remains warning-severity only when it is normalized and explained by a domain limitation.
- Partial and unavailable evidence without an explanation is invalid.
- Unavailable required groups must expose canonical zero-value payloads so stale or non-finite values cannot be hidden.
- Available geographical, operational, and trajectory groups require observation support.
- Operational ground and airborne shares are reconciled only when on-ground fields are claimed available by the typed limitation contract.
- Quality limitations are rebuilt from current group evidence on every validation pass before current validator issues are merged.
- Validator-owned stale limitations and stale group-derived aggregate limitations are not retained after source evidence is corrected.
- Numeric tolerance is dimensionless and relative; it is never added directly to values expressed in degrees, kilometres, metres, seconds, or ratios.
- Nil validation contexts are rejected.

## Version boundary

```text
Validator: flight-feature-validator-v5 -> flight-feature-validator-v6
Processing Pipeline: flight-feature-processing-pipeline-v11 -> flight-feature-processing-pipeline-v12
Schema: flight-features-v1 unchanged
PostgreSQL migration: not required
```

## Permanent evidence

The increment adds regression tests for partial mathematical corruption, stale limitation removal, unavailable residual payloads, zero-support available operational evidence, missing partial explanations, longitude envelope mismatch, and nil contexts. `tools/validatorreviewaudit` is part of Backend Quality in Continuous Integration.

## Review classification

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
VALIDATOR_REVIEW_STATUS=PENDING_EXACT_COMMIT_CI
```
