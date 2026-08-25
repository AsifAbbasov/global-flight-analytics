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

## Canonical remediation history

Historical reviewer identities and review-comment chronology are not reconstructed. The records below are reconstructed from repository source/tests, migration 029, remediation commit `18dde73b2d122d00476ea21accb256b33fc23527`, permanent strict audit, and exact Backend CI run `30374964285`. Severity labels are retrospective.

### GFA-DB-279 — Historical Aggregate region persistence required uppercase while the canonical contract required lowercase
1. **Finding / symptom:** migration 015 enforced uppercase region identifiers while Historical Contract v2 and aggregate key normalization used lowercase.
2. **Root cause:** persistence validation and domain canonicalization evolved under different region-normalization policies.
3. **Failure scenario:** a valid canonical lowercase regional result is rejected by PostgreSQL, or a legacy uppercase row has a deterministic identity incompatible with the corrected key.
4. **Impact:** regional Historical Aggregate persistence and replay identity can fail or diverge.
5. **Severity rationale:** **P1 retrospective** because the database could reject canonical domain output and legacy identities could not be safely rewritten in place.
6. **Existing guarantees violated:** domain and persistence canonical identity must agree exactly.
7. **Considered solutions:** switch domain back to uppercase, silently rewrite legacy rows, loosen the constraint, or migrate persistence to lowercase and fail on incompatible legacy identities.
8. **Chosen remediation:** migration 029 enforces lowercase grammar and explicitly fails when incompatible legacy regional rows require governed rematerialization.
9. **Why selected:** preserves corrected canonical identity without silently changing deterministic record IDs.
10. **Rejected alternatives:** permissive mixed-case persistence and silent legacy identity rewrite.
11. **Trade-offs:** deployments with legacy regional rows require explicit rematerialization.
12. **Regression tests / protection:** migration contract tests, production catalog regression, PostgreSQL verifier and strict audit.
13. **Adversarial review findings:** cross-module scope-normalizer consolidation remains separate and is not required for persistence correctness.
14. **Remediation iterations:** `18dde73b2d122d00476ea21accb256b33fc23527`; migration 029.
15. **Residual risks and limitations:** legacy rows created under incompatible identity policy cannot be reconstructed automatically without an explicit reset/rematerialization decision.
16. **Operational or deployment consequences:** migration may fail deliberately when unsafe legacy regional data exists.
17. **Exact evidence:** migration 029 checksum `4759cbb11050faa2a5b47be762f46be3a898e481e6e0fee9cfe7cdf0138c696d`; commit; run `30374964285` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** scope canonicalization changes require coordinated domain key, database constraint and deterministic-ID migration review.

### GFA-DATA-280 — Stored Historical JSON was not reconciled with denormalized row metadata
1. **Finding / symptom:** stored JSON could disagree with denormalized schema/metric/scope/window/status/confidence/fingerprint columns without read rejection.
2. **Root cause:** row scanning reconstructed the JSON result while trusting columns as independent query metadata.
3. **Failure scenario:** selection/filter columns identify one result while JSON payload describes another.
4. **Impact:** corrupted or drifted persistence can enter analytics under false identity/quality metadata.
5. **Severity rationale:** **P1 retrospective** because durable row contradiction could alter query selection and analytical trust.
6. **Existing guarantees violated:** denormalized columns are mirrors of one canonical stored result.
7. **Considered solutions:** trust JSON, trust columns, or reconcile all output-affecting mirrors.
8. **Chosen remediation:** every SELECT/RETURNING path reads complete metadata and row reconstruction validates it against canonical JSON.
9. **Why selected:** detects corruption and schema drift at the persistence boundary.
10. **Rejected alternatives:** precedence rules that silently choose one representation.
11. **Trade-offs:** inconsistent legacy rows fail closed.
12. **Regression tests / protection:** `TestScanRecordRejectsStoredMetadataMismatch`, migration JSON-metadata constraint and strict audit.
13. **Adversarial review findings:** domain/PostgreSQL contract files were already separated; no duplicate architecture finding created.
14. **Remediation iterations:** `18dde73b2d122d00476ea21accb256b33fc23527`.
15. **Residual risks and limitations:** fields intentionally not denormalized cannot be cross-checked at SQL-column level.
16. **Operational or deployment consequences:** corrupt rows are rejected instead of normalized silently.
17. **Exact evidence:** remediation commit, migration 029 `json_metadata_check`, run `30374964285`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every new denormalized field requires row↔JSON consistency enforcement and tests.

### GFA-DATA-281 — Stored aggregate record identifiers were not recomputed from canonical result identity
1. **Finding / symptom:** reads accepted stored IDs without proving they matched the deterministic result key and input fingerprint.
2. **Root cause:** record ID was treated as trusted persistence identity rather than a derived invariant.
3. **Failure scenario:** a corrupted/misassociated ID survives scanning and can break idempotency, pagination or external references.
4. **Impact:** immutable record identity can disagree with its payload/key.
5. **Severity rationale:** **P1 retrospective** because deterministic identity is part of persistence correctness.
6. **Existing guarantees violated:** stored identifiers must be reproducible from canonical key plus fingerprint.
7. **Considered solutions:** trust database ID, add random IDs, or recompute and reject mismatch.
8. **Chosen remediation:** row reconstruction computes `expectedID := makeRecordID(...)` and rejects mismatch.
9. **Why selected:** proves identity instead of relying on storage history.
10. **Rejected alternatives:** accepting arbitrary stored identity and repairing it on read.
11. **Trade-offs:** malformed legacy rows become unreadable until repaired/rematerialized.
12. **Regression tests / protection:** `TestScanRecordRejectsStoredIdentifierMismatch`, strict audit and PostgreSQL verification.
13. **Adversarial review findings:** the existing four-field pagination cursor was already correct and is not duplicated as a finding.
14. **Remediation iterations:** `18dde73b2d122d00476ea21accb256b33fc23527`.
15. **Residual risks and limitations:** deterministic-ID correctness still depends on stable key/fingerprint policy versions.
16. **Operational or deployment consequences:** identity corruption fails closed.
17. **Exact evidence:** commit, stored-record tests, run `30374964285` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** derived persistent IDs must be verified on every reconstruction path.

### GFA-DATA-282 — Aggregate idempotency accepted same fingerprint with different canonical payload
1. **Finding / symptom:** an existing key/fingerprint pair was treated as idempotent even when the canonical JSON payload differed.
2. **Root cause:** fingerprint equality was treated as sufficient proof of complete semantic output equality.
3. **Failure scenario:** a bug or collision produces different result content under the same fingerprint and replay silently returns the existing record.
4. **Impact:** immutable result conflicts are hidden as successful idempotency.
5. **Severity rationale:** **P1 retrospective** because persistence could conceal contradictory analytical output.
6. **Existing guarantees violated:** idempotent replay requires same identity and same canonical payload.
7. **Considered solutions:** trust fingerprint, overwrite, version every replay, or compare canonical payload.
8. **Chosen remediation:** same key/fingerprint is idempotent only when canonical JSON matches; otherwise return `ErrResultPayloadConflict`.
9. **Why selected:** preserves immutability and turns contradictions into explicit evidence.
10. **Rejected alternatives:** overwrite and fingerprint-only replay acceptance.
11. **Trade-offs:** canonical serialization becomes part of the idempotency proof.
12. **Regression tests / protection:** `TestPostgresStoreRejectsSameFingerprintDifferentPayload`, runtime verifier conflict check and strict audit.
13. **Adversarial review findings:** no new hash algorithm is required; payload equality guards the semantic conflict.
14. **Remediation iterations:** `18dde73b2d122d00476ea21accb256b33fc23527`.
15. **Residual risks and limitations:** a deliberate schema-version change must define canonical payload compatibility explicitly.
16. **Operational or deployment consequences:** conflicting replay fails instead of silently succeeding.
17. **Exact evidence:** commit, `ErrResultPayloadConflict`, verifier PASS check, run `30374964285`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** idempotency at immutable stores must compare canonical output, not only an input fingerprint.

### GFA-ARCH-283 — Historical Materialization depended on the full Aggregate Store although it only writes
1. **Finding / symptom:** the materializer required read/write Store behavior while its responsibility only needed persistence writes.
2. **Root cause:** dependency interface followed the concrete store surface instead of the consumer’s required behavior.
3. **Failure scenario:** materialization becomes coupled to unrelated read/pagination methods and test doubles must implement unnecessary operations.
4. **Impact:** higher coupling and broader change surface.
5. **Severity rationale:** **P3 retrospective** because this is maintainability/interface-segregation debt rather than incorrect output.
6. **Existing guarantees violated:** consumers should depend on the narrowest behavior contract they use.
7. **Considered solutions:** retain Store, define a local writer, or expose canonical aggregate `Writer` contract.
8. **Chosen remediation:** Aggregate exposes `Writer`; Historical Materialization depends only on it.
9. **Why selected:** keeps ownership in the aggregate domain and removes unnecessary dependency surface.
10. **Rejected alternatives:** duplicate materializer-specific write interface with the same semantics.
11. **Trade-offs:** one additional public domain interface to maintain.
12. **Regression tests / protection:** strict audit requires `type Writer interface` and forbids Materialization dependence on full Store.
13. **Adversarial review findings:** existing domain/PostgreSQL contract split was already present and not reopened.
14. **Remediation iterations:** `18dde73b2d122d00476ea21accb256b33fc23527`.
15. **Residual risks and limitations:** future materialization read requirements would need explicit new dependency review.
16. **Operational or deployment consequences:** none; composition/test simplification.
17. **Exact evidence:** remediation commit, audit, exact CI run.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** aggregate consumers must use `Writer`, `Reader`, or `Store` according to actual required capability.

### GFA-DATA-284 — Aggregate Put canonicalized caller input before validating the original domain value
1. **Finding / symptom:** storage normalization could repair/erase invalid caller-owned state before contract validation.
2. **Root cause:** canonicalization was executed as a prerequisite to validation instead of after validation.
3. **Failure scenario:** a malformed result becomes a valid-looking normalized value and is persisted rather than rejected.
4. **Impact:** invalid upstream evidence can be hidden by the persistence layer.
5. **Severity rationale:** **P1 retrospective** because the store could silently transform invalid analytical evidence into durable state.
6. **Existing guarantees violated:** persistence validates caller truth before deterministic storage normalization.
7. **Considered solutions:** validate normalized result only, validate both, or validate raw first then canonicalize.
8. **Chosen remediation:** `validateStorableResult` checks the unmodified caller value before canonical storage transformation.
9. **Why selected:** fail-closed evidence semantics preserve responsibility boundaries.
10. **Rejected alternatives:** persistence as an implicit repair layer.
11. **Trade-offs:** callers must explicitly fix invalid normalization rather than relying on Store cleanup.
12. **Regression tests / protection:** `TestPostgresStoreValidatesBeforeCanonicalization` and strict audit.
13. **Adversarial review findings:** normal deterministic canonicalization remains valid after raw contract success.
14. **Remediation iterations:** `18dde73b2d122d00476ea21accb256b33fc23527`.
15. **Residual risks and limitations:** validation covers the current Historical Contract; future fields need coordinated validation.
16. **Operational or deployment consequences:** invalid writes fail earlier.
17. **Exact evidence:** commit, regression test, run `30374964285` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** persistence normalization must never precede validation when normalization can erase a contract violation.

### GFA-OPS-285 — Historical Aggregate silently substituted nil caller contexts
1. **Finding / symptom:** Store methods replaced nil context with `context.Background()`.
2. **Root cause:** convenience fallback overrode caller lifecycle ownership.
3. **Failure scenario:** a caller bug loses deadline/cancellation semantics and database work continues outside the intended request lifetime.
4. **Impact:** unbounded or orphaned database work.
5. **Severity rationale:** **P2 retrospective** as an operational lifecycle correctness defect.
6. **Existing guarantees violated:** caller-owned context is mandatory at persistence boundaries.
7. **Considered solutions:** background fallback, panic, or typed error.
8. **Chosen remediation:** Store methods reject nil context with `ErrContextRequired` and preserve cancellation.
9. **Why selected:** explicit failure without process panic.
10. **Rejected alternatives:** implicit background lifetime.
11. **Trade-offs:** invalid callers must be corrected.
12. **Regression tests / protection:** `TestPostgresStoreRejectsNilContext`; strict audit forbids `nonNilContext` fallback.
13. **Adversarial review findings:** nil/error Go constructor behavior remains unrelated and intentionally retained.
14. **Remediation iterations:** `18dde73b2d122d00476ea21accb256b33fc23527`.
15. **Residual risks and limitations:** callers still choose appropriate deadlines.
16. **Operational or deployment consequences:** database work follows request cancellation consistently.
17. **Exact evidence:** commit, context tests, run `30374964285`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** new Store methods must call the same explicit context requirement before database work.

### GFA-DATA-286 — Aggregate storage time could be zero or precede result generation
1. **Finding / symptom:** injected clock output was not required to be non-zero or causally after/equal to `GeneratedAt`.
2. **Root cause:** storage timestamp was treated as operational metadata rather than analytical provenance with causal constraints.
3. **Failure scenario:** persisted result claims it was stored before it existed, or has an unknown zero storage time.
4. **Impact:** invalid provenance, ordering and freshness evidence.
5. **Severity rationale:** **P1 retrospective** because durable temporal provenance could become impossible.
6. **Existing guarantees violated:** storage time must be concrete and causally follow generation.
7. **Considered solutions:** accept clock output, substitute current time, clamp to `GeneratedAt`, or reject invalid time.
8. **Chosen remediation:** `StoredAt` must be non-zero and not precede `GeneratedAt`; invalid clocks return `ErrStoredAtInvalid`.
9. **Why selected:** fail closed without inventing or mutating provenance.
10. **Rejected alternatives:** timestamp clamping and fallback clocks.
11. **Trade-offs:** misconfigured/injected clocks fail persistence explicitly.
12. **Regression tests / protection:** `TestPostgresStoreRejectsInvalidStoredAt`, migration causality constraint and audit.
13. **Adversarial review findings:** no arbitrary time tolerance is needed for this causal relationship.
14. **Remediation iterations:** `18dde73b2d122d00476ea21accb256b33fc23527`.
15. **Residual risks and limitations:** wall-clock correctness remains an infrastructure concern; only causal ordering is enforced here.
16. **Operational or deployment consequences:** invalid clock behavior blocks writes.
17. **Exact evidence:** commit, `stored_at_causality_check`, tests, run `30374964285`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every new persisted processing timestamp must define and enforce its causal relation to domain timestamps.

### GFA-DB-287 — Historical Aggregate timestamp mirrors lacked database-level consistency constraints
1. **Finding / symptom:** timestamp columns and nanosecond mirror columns could disagree while the row remained database-valid.
2. **Root cause:** migration 015 stored redundant temporal representations without a CHECK invariant.
3. **Failure scenario:** SQL filtering/order uses one timestamp representation while reconstructed identity/JSON uses another.
4. **Impact:** contradictory temporal identity and query behavior.
5. **Severity rationale:** **P1 retrospective** because durable database state could encode two different times for the same field.
6. **Existing guarantees violated:** redundant timestamp mirrors must be atomically consistent.
7. **Considered solutions:** remove mirrors, trust application writes, repair on read, or add database CHECK constraints.
8. **Chosen remediation:** migration 029 adds timestamp-mirror consistency constraints and schema verification.
9. **Why selected:** prevents contradiction at the lowest durable boundary.
10. **Rejected alternatives:** application-only validation and read-time repair.
11. **Trade-offs:** migration rejects inconsistent existing data and adds constraint maintenance.
12. **Regression tests / protection:** migration contract test requires `timestamp_mirror_check`; PostgreSQL verifier pins migration 029.
13. **Adversarial review findings:** full tuple cursor was already correct and remains protected separately.
14. **Remediation iterations:** `18dde73b2d122d00476ea21accb256b33fc23527`; migration 029.
15. **Residual risks and limitations:** consistency proves representations match, not that the source timestamp itself is externally accurate.
16. **Operational or deployment consequences:** inconsistent writes are rejected by PostgreSQL.
17. **Exact evidence:** migration 029, remediation commit, run `30374964285` SUCCESS, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** redundant durable temporal representations require database-level equality constraints and migration regression tests.
