# Historical Replay Review Hardening

Status: closed

## Scope

This increment hardens
`apps/api/internal/historicalintelligence/historicalreplay` and the replay branch
of `apps/api/cmd/materialize-historical-intelligence`.

## Accepted findings

- Replay accepted a successful Materializer call without validating the returned
  Materialization version, plans, period summaries, historical results, record
  key, record identifier, persisted fingerprint, storage time, or comparison.
- Replay results did not independently preserve complete, partial, or failed
  execution state after separation from the returned Go error.
- The production command discarded a successfully persisted completed prefix
  whenever a later replay window failed.
- Separate per-window Materialization transactions could observe a shared period
  differently between adjacent calls even though the replay used one analytical
  as-of time.
- Global metric, scope, dataset-limit, generation-time, and replay-limit failures
  were not rejected before entering the window loop.
- Replay planning could use the larger Historical Window allocation limit and
  only afterwards apply the smaller replay-window limit.
- Nil context was replaced with a new background context, losing caller
  cancellation ownership.
- Replay had no canonical input fingerprint or public result validation contract.
- Regression coverage did not protect malformed Materialization outcomes,
  self-contained partial progress, continuity, early validation, bounded
  planning, clone isolation, result tampering, or production partial reporting.

## Corrected contracts

- Historical Replay advances to `historical-replay-v2`.
- Every successful Materializer call is validated before its record can enter the
  replay result. Validation covers the Materialization version, canonical
  one-bucket current and previous plans, read-summary windows, dataset limits,
  snapshot isolation label, all three Historical Results, required comparison,
  record identity, record payload, fingerprints, and storage causality.
- The replay `Result` records `complete`, `partial`, or `failed` status, planned
  and completed window counts, structured failure information, generated,
  started, and completed times, and a deterministic replay input fingerprint.
- `Result.Validate` reconstructs the replay fingerprint and verifies the plan,
  limits, status/count relationship, completed-prefix ordering, record contracts,
  failure location, and adjacent-period continuity.
- Global replay request validation occurs before Materialization. Metric and
  scope permission use the production `MetricSpecFor` catalog; dataset and both
  count limits are normalized once; generation time must be at or after the
  analytical as-of time and not after replay execution starts.
- Planning uses the lower of the normalized Historical Window bucket limit and
  replay window limit. A replay-window limit is reported as
  `WindowCountExceededError`; a lower bucket-allocation limit remains the
  Historical Window error. Every one-window Materialization request receives
  `MaximumBucketCount=1`.
- Adjacent Materializations must agree on the raw input fingerprint for their
  shared period. A mismatch stops the replay with an explicit partial result
  instead of silently persisting an internally inconsistent chain.
- The command operation now returns a non-empty partial report together with the
  replay error. The executable writes that JSON report to standard output and
  still returns a non-zero exit code and the failure on standard error.
- Replay and command operation reject nil context.

## Findings already resolved before this review

- Historical Materialization version two already removed the combined bounded
  read. Every Materialization call reads previous and current periods with
  independent limits inside one managed PostgreSQL repeatable-read transaction.
- Historical Contract version two already validates coverage, comparison-current
  value, confidence evidence, provenance, finite values, and metric-specific
  count, ratio, and continuous-number semantics.
- Historical Aggregate already enforces deterministic identity, canonical
  payload idempotency, full row-versus-JSON consistency, and storage-time
  causality.
- `WindowError.Unwrap` already returned its wrapped error. No correction was
  required.

## Qualified or rejected findings

- A replay-wide PostgreSQL transaction is deliberately not introduced. A long
  transaction across as many as ten thousand durable writes conflicts with the
  completed-prefix recovery model, increases database retention and lock costs,
  and would make partial persistence ambiguous. Shared-period fingerprint
  continuity detects the actual cross-call consistency risk without changing
  commit semantics.
- `MaximumBucketCount` and `MaximumWindowCount` remain separate because they are
  different operator controls: one bounds Historical Window planning and one
  bounds replay work. Their interaction is now deterministic and allocation is
  bounded by the lower value.
- A checkpoint or resume token is a future operational feature, not a missing
  integrity invariant. Aggregate writes are already idempotent, and a repeated
  replay remains safe. Resume semantics require a separate versioned API and are
  not mixed into this review correction.
- No arbitrary future-time tolerance is added. The causal rule is exact:
  `GeneratedAt` cannot precede `AsOfTime` and cannot exceed replay `StartedAt`.
- Returning `nil, error` from a Go constructor remains idiomatic and is not a
  domain null result.
- Function length is a review signal rather than an independent defect. Runner
  responsibilities are decomposed into request normalization, planning,
  Materialization outcome validation, fingerprinting, result validation, and
  execution control.
- Historical metric `float64` transport and comparison confidence semantics are
  retained closed Historical Contract and Historical Comparison decisions; they
  are not Replay defects.

```text
MATERIALIZATION_OUTCOME_VALIDATION=ENFORCED
REPLAY_RESULT_STATUS=EXPLICIT
PARTIAL_PROGRESS_REPORTING=PRESERVED
OVERLAPPING_PERIOD_CONTINUITY=VERIFIED
REPLAY_REQUEST_VALIDATION=EARLY
REPLAY_PLANNING_LIMITS=BOUNDED
NIL_CONTEXT_REJECTED=YES
REPLAY_INPUT_FINGERPRINT=BOUND
REPLAY_WIDE_TRANSACTION=REJECTED_BY_DESIGN
CHECKPOINT_RESUME_TOKEN=SEPARATE_PRODUCT_FEATURE
HISTORICAL_REPLAY_ENGINEERING_REMEDIATION=IMPLEMENTED
```

## Permanent verification

`apps/api/tools/historicalreplayreviewaudit` protects the replay version,
self-contained status and failure model, early request validation, bounded plan,
strict Materialization outcome validation, cross-call shared-period continuity,
canonical replay fingerprint, production completed-prefix reporting, nil-context
rejection, regression tests, and this review record in Backend Continuous
Integration.

## Formal closure evidence

The Historical Replay engineering remediation was committed and validated before
this administrative closure:

```text
ENGINEERING_BASELINE_COMMIT=d73c27b5e54108c7d2b9a009cb157496f7c67bde
ENGINEERING_REMEDIATION_COMMIT=38b14fbb8649a2e7e875cd4ae7ed73b6a954a068
ENGINEERING_GITHUB_ACTIONS_RUN=30390451707
Backend Quality=SUCCESS
Backend Quality Job=90380396908
PostgreSQL 16 Integration=SUCCESS
PostgreSQL 16 Integration Job=90380396909
Backend Race Safety=SUCCESS
Backend Race Safety Job=90380396961
Backend Container=SUCCESS
Backend Container Job=90380713650
```

All accepted findings are implemented. Findings already resolved before this
review and qualified or rejected findings retain their documented rationale.
No Historical Replay review item remains open, unclassified, or deferred.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_REPLAY_ENGINEERING_DEBT=CLOSED
HISTORICAL_REPLAY_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
HISTORICAL_REPLAY_REVIEW_STATUS=CLOSED
```
