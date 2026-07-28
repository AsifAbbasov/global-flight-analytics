# Historical Materialization Review Hardening

Status: implemented review remediation

## Scope

This increment hardens
`apps/api/internal/historicalintelligence/historicalmaterialization` and adds the
minimal two-period Historical Read boundary required by that orchestrator.

## Accepted findings

- A single bounded read over the combined previous and current windows allowed
  chronologically earlier rows to consume the complete dataset limit and create
  false current-period declines.
- Materialization did not verify that repository output matched the requested
  query or the current Historical Read version.
- One combined read summary could not identify which period was truncated or
  which period had weaker represented evidence.
- Materialization returned the newly built comparison even when the writer
  returned a canonical persisted result.
- The materialization-specific fingerprint repair excluded `GeneratedAt` and
  duplicated provenance ownership already assigned to Historical Comparison.
- Nil context was replaced with `context.Background`, discarding caller
  cancellation ownership.
- Lower-layer errors did not identify the failed orchestration stage.
- The exported `DatasetLimitOr` helper duplicated already-normalized request
  state.
- Regression coverage did not protect independent limits, two-period atomicity,
  snapshot metadata, persisted outcome identity, stage errors, generated-time
  identity, default and maximum limits, or clone isolation.

## Corrected contracts

- `historicalread.PeriodRepository` reads previous and current queries inside one
  managed repeatable-read transaction and applies an independent dataset and
  route-payload limit to each period.
- `Config.Repository` retains its original source-compatible type. Construction
  fails closed unless the concrete repository also supports
  `historicalread.PeriodRepository`; no duplicate dependency field is exposed.
- The previous and current queries must be adjacent and share one analytical
  as-of time.
- Materialization validates each snapshot version, exact normalized query, and
  shared supported isolation label before invoking a builder.
- Outcome exposes separate previous and current read summaries, including loaded
  rows, matched rows, route payload bytes, and every limit signal. The old
  aggregate `ReadSummary` remains as a deprecated source-compatibility view and
  must not be used for period-sensitive quality decisions.
- Builders receive only their own period snapshot. Their fingerprints no longer
  include unrelated rows from the adjacent period.
- Historical Comparison remains the sole owner of both-period comparison
  provenance and fingerprinting. Its version-two fingerprint already binds
  `GeneratedAt`.
- After persistence, `Record.Result` is the canonical current result returned in
  Outcome. Persistence identity and contract metadata are checked before return.
- Materialization rejects nil context and preserves caller cancellation.
- Typed stage errors identify request validation, planning, read, snapshot
  contract, builder, comparison, persistence, and persistence-contract failures.
- `DatasetLimitOr` is removed and normalized limits are assigned directly.
- Materialization version advances to `historical-materialization-v2`.

## Findings already resolved before this review

- A normal `historicalread.NewPostgres` repository already opens one read-only
  PostgreSQL repeatable-read transaction for every Snapshot. Passing a pool in
  production does not mean the four dataset queries execute outside a snapshot.
- Historical Aggregate already exposes a narrow `Writer`; Materialization no
  longer depends on the complete aggregate Store.
- Historical Comparison version two already requires matching per-bucket
  coverage profiles and records both periods' status, confidence score, sample
  count, and previous-period limitations. Current-series confidence remaining
  contract-consistent is an explicit closed design decision.

## Qualified or rejected findings

- Metric classification and scope permission already come from the single
  `MetricSpecFor` catalog. The remaining three-family builder switch is a finite
  domain dispatch, not three independent registries and not an Open/Closed
  Principle defect.
- Returning `nil, error` from a Go constructor is idiomatic and is not a domain
  null payload.
- Function and file length are review signals rather than independent defects.
  Materialize is decomposed here because period planning, read validation, and
  persistence validation are distinct responsibilities.
- Cross-module scope normalization consolidation is a separate contract
  migration and is not mixed into this period-read integrity correction.
- The combined latest source timestamp is defined as the newest evidence across
  both periods. Period-specific results and read summaries retain the evidence
  needed to inspect each side separately.
- Historical metrics are not monetary calculations; no decimal-library
  migration belongs in this module.

```text
INDEPENDENT_PERIOD_LIMITS=ENFORCED
ATOMIC_TWO_PERIOD_READ=ENFORCED
SNAPSHOT_QUERY_AND_VERSION=VERIFIED
PERIOD_READ_SUMMARIES=EXPLICIT
PERIOD_BUILDER_INPUTS=ISOLATED
HISTORICAL_COMPARISON_PROVENANCE_OWNER=PRESERVED
PERSISTED_RESULT_IS_CANONICAL=YES
GENERATED_AT_FINGERPRINT_IDENTITY=BOUND
NIL_CONTEXT_REJECTED=YES
STAGE_ERRORS=EXPLICIT
DATASET_LIMIT_HELPER=REMOVED
HISTORICAL_MATERIALIZATION_ENGINEERING_REMEDIATION=IMPLEMENTED
```

## Permanent verification

`apps/api/tools/historicalmaterializationreviewaudit` protects the independent
period limits, one-transaction two-period read, snapshot query and version
validation, explicit period summaries, persisted canonical outcome, comparison
provenance ownership, generated-time identity, nil-context rejection, stage
errors, regression tests, and this review record in Backend Continuous
Integration.
