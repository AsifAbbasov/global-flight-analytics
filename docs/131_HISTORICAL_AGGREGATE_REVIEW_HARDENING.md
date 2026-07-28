# Historical Aggregate Review Hardening

Status: closed

## Scope

This increment hardens
`apps/api/internal/historicalintelligence/historicalaggregate` at the immutable
PostgreSQL aggregate boundary.

## Accepted findings

- Migration 015 required uppercase region identifiers while the version-two
  Historical Contract and aggregate key canonicalization require lowercase.
- Stored JSON was not checked against all denormalized identity and quality
  columns.
- Stored record identifiers were not recomputed from the result key and input
  fingerprint during reads.
- An existing record with the same key and fingerprint was treated as an
  idempotent replay even when its canonical payload differed.
- Historical Materialization depended on the full Store although it only writes.
- Put canonicalized the result before validating the caller-owned domain value.
- Nil contexts were replaced with `context.Background`, discarding cancellation
  ownership.
- The injected clock could produce a zero `StoredAt` or a storage time before
  `GeneratedAt`.
- Migration 015 kept timestamp and nanosecond mirrors without a database
  consistency constraint.

## Corrected contracts

- Migration 029 replaces the region constraint with the canonical lowercase
  grammar and adds timestamp-mirror and JSON-metadata consistency constraints.
  It fails explicitly when legacy regional rows exist because their
  deterministic identifiers were created under an incompatible uppercase key
  and must be rematerialized rather than silently rewritten.
- Every SELECT and INSERT RETURNING path reads the complete denormalized row
  identity.
- Row reconstruction validates schema version, metric, scope, granularity,
  window, status, confidence level, fingerprint, scope key, and deterministic
  record identifier against the JSON result.
- Replay idempotency now requires the same key, fingerprint, and canonical JSON
  payload. A same-fingerprint payload difference returns
  `ErrResultPayloadConflict`.
- `Writer` is a dedicated aggregate contract and Historical Materialization
  depends only on it.
- The unmodified caller result is validated before deterministic storage
  canonicalization.
- Store methods reject nil context and preserve caller cancellation.
- `StoredAt` must be non-zero and not precede `GeneratedAt`.
- The complete four-field keyset cursor and lexicographic SQL remain protected.

## Findings already resolved before this review

- Full tuple pagination already uses window end, window start, as-of time, and
  record identifier. The attached review described the original
  `BeforeWindowEnd` contract, not current main.
- Domain contracts and PostgreSQL contracts are already split between
  `historicalaggregatecontract`, aggregate aliases, and
  `postgres_contracts.go`.
- Historical Contract generation two owns a single `MetricSpecFor` catalog that
  binds each metric name to its unit, aggregation, value kind, builder family,
  and scopes. `ResultKey` therefore does not need to duplicate unit and
  aggregation.

## Qualified or rejected findings

- Function length is a review signal, not independent evidence of a defect.
  This increment decomposes stored-row integrity because it has a coherent
  responsibility, not to satisfy a line-count quota.
- Words such as `With` are not globally forbidden by
  `docs/82_CODE_REVIEW_STANDARD.md`; `NewPostgresWithExecutor` is not changed
  without a demonstrated semantic ambiguity.
- Returning `nil, error` from Go constructors and returning nil from a nil
  receiver's `Unwrap` are idiomatic Go contracts, not domain null payloads.
- The status/confidence index is retained. Absence from the current Store query
  set does not prove that operational and diagnostic SQL consumers do not use
  it.
- The migration metric allow-list is a versioned persistence contract. Adding a
  new production metric deliberately requires both catalog and schema review;
  this is not an Open/Closed Principle defect.
- `float64` remains appropriate for non-monetary historical analytics. This
  store introduces no additional rounding policy.
- Cross-module scope-normalizer consolidation is a separate migration and is
  not mixed into this persistence integrity correction.

```text
REGION_SCOPE_DATABASE_CONTRACT=LOWERCASE
FULL_TUPLE_CURSOR=PREEXISTING_AND_VERIFIED
STORED_METADATA_JSON_CONSISTENCY=ENFORCED
STORED_RECORD_IDENTITY=ENFORCED
IDEMPOTENCY_CANONICAL_PAYLOAD=ENFORCED
MATERIALIZER_WRITER_INTERFACE=ENFORCED
RAW_DOMAIN_VALIDATION_BEFORE_CANONICALIZATION=ENFORCED
NIL_CONTEXT_REJECTED=YES
STORED_AT_CAUSALITY=ENFORCED
TIMESTAMP_MIRROR_DATABASE_CONSISTENCY=ENFORCED
HISTORICAL_AGGREGATE_ENGINEERING_REMEDIATION=IMPLEMENTED
```

## Migration evidence

```text
MIGRATION_VERSION=029
MIGRATION_NAME=harden_historical_aggregate_integrity
MIGRATION_SHA256=4759cbb11050faa2a5b47be762f46be3a898e481e6e0fee9cfe7cdf0138c696d
```

## Permanent verification

`apps/api/tools/historicalaggregatereviewaudit` protects the aggregate writer
boundary, full tuple cursor, canonical lowercase region contract, complete
stored-row consistency, deterministic identifier reconstruction, canonical
payload idempotency, raw validation order, nil-context rejection, StoredAt
causality, migration 029, regression tests, and this review record in Backend
Continuous Integration.

## Formal closure evidence

The Historical Aggregate engineering remediation was committed and validated
before this administrative closure:

```text
ENGINEERING_BASELINE_COMMIT=82ebd68d0372c885d724308d2291c61dab2de378
ENGINEERING_REMEDIATION_COMMIT=18dde73b2d122d00476ea21accb256b33fc23527
ENGINEERING_GITHUB_ACTIONS_RUN=30374964285
Backend Quality=SUCCESS
Backend Quality Job=90328145394
Backend Race Safety=SUCCESS
Backend Race Safety Job=90328145512
PostgreSQL 16 Integration=SUCCESS
PostgreSQL 16 Integration Job=90328145602
Backend Container=SUCCESS
Backend Container Job=90328492165
```

All accepted findings are implemented. Qualified, rejected, or pre-resolved
findings retain their documented rationale, and no Historical Aggregate review
item remains open, unclassified, or deferred.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_AGGREGATE_ENGINEERING_DEBT=CLOSED
HISTORICAL_AGGREGATE_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
HISTORICAL_AGGREGATE_REVIEW_STATUS=CLOSED
```
