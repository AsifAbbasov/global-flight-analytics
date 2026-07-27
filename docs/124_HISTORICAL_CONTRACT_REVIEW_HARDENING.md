# Historical Contract Review Hardening

## Baseline

`39549504bbeff1a6c272153bf3dcde469b766202`

## Corrected integrity boundaries

- One production metric catalog owns metric name, unit, aggregation, value kind, builder family, and allowed scopes.
- Only the sixteen metrics materialized by Traffic, Airport, and Route builders are advertised as supported.
- Four reserved names remain source-compatible constants but are not accepted as materializable metrics.
- Count values remain in the heterogeneous float64 transport field but must be exact non-negative safe integers.
- Ratio values use an explicit absolute tolerance; continuous values use a dimensionless relative tolerance; count comparisons are exact.
- Unavailable bucket confidence is canonical zero evidence.
- Partial series require at least one represented partial or complete bucket.
- A zero-coverage result with no represented bucket is canonical unavailable and does not expose unavailable points as analytical observations.
- Partial and unavailable evidence require explicit limitations.
- Comparison current values are bound to the aggregation-selected current summary.
- Confidence reason contributions must reconcile to the declared score.
- The schema registry now describes every semantic field already present in the result model.
- Aggregate region normalization is lowercase and matches the contract.
- Zero-event complete buckets remain valid because source coverage and event count are independent concepts.

## Deliberately retained contracts

- `Point.Value` remains `float64` because one versioned series model carries count, ratio, rate, and distance metrics.
- Optional comparison fields remain Go pointers; nil is the idiomatic representation of absent optional comparison evidence.
- Custom granularity remains supported because the production window planner and materializer already use it.
- Region scope remains a structural type but no current production metric catalog entry allows it.

## Versions

```text
Historical Contract implementation: historical-intelligence-contract-v2
Historical Contract validation: historical-intelligence-contract-validation-v2
Schema: historical-intelligence-v1
PostgreSQL migration: not required
```

## Closure

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_CONTRACT_REVIEW_STATUS=CLOSED_PENDING_EXACT_COMMIT_CI
```
