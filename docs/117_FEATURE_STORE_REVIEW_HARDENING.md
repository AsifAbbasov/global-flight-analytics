# Document 117 — Feature Store Review Hardening

Status: IMPLEMENTED
Baseline commit: `92691d993d7340112399a40bd9ecbc62ddb240ad`

## Review classification

The reviewed report was produced on commit `a1689dc`. Several findings were
already closed before this increment:

- processing version participates in snapshot identity and PostgreSQL uniqueness;
- extraction fingerprints include processing identity and aircraft metadata;
- the pipeline depends on a narrow writer interface;
- timestamp mirrors are checked fail-closed on every PostgreSQL read;
- real PostgreSQL Feature Store tests and the production verifier run in Continuous Integration.

The remaining confirmed findings were output-identity conflicts, direct domain
model serialization, Memory/PostgreSQL validation divergence, incomplete write
validation proof, unbounded Memory Store growth and nil-context substitution.

## Output identity

Every stored record now carries an output fingerprint derived from a canonical,
versioned persistence payload. Idempotent replay requires both the input and output
fingerprints to match. A semantic output change under the same snapshot key and
input fingerprint is a conflict.

Operational timestamps are intentionally excluded from output identity:
`ExtractedAt`, validation time and aircraft metadata retrieval time record when a
run happened, not what analytical result it produced.

```text
FEATURE_STORE_OUTPUT_FINGERPRINT=ENFORCED
```

## Persistence contract

New rows use `flight-feature-snapshot-payload-v1`, an explicit persistence data
transfer object with stable JSON field names, strict unknown-field rejection and
an embedded output fingerprint. Legacy unversioned rows remain readable through a
separate compatibility decoder. No database migration is required because the
existing JSONB column and validation-report constraints remain compatible.

```text
FEATURE_STORE_PAYLOAD_VERSIONING=ENFORCED
```

## Store conformance

Memory and PostgreSQL stores now share:

- canonical Universally Unique Identifier trajectory validation;
- strict `sha256:` input fingerprint validation;
- non-finite numeric rejection;
- explicit nil-context rejection;
- complete validation audit proof for new writes.

The in-memory implementation is bounded by a configurable record capacity and
fails without evicting immutable historical evidence.

```text
FEATURE_STORE_IMPLEMENTATION_CONFORMANCE=ENFORCED
FEATURE_STORE_MEMORY_CAPACITY=BOUNDED
FEATURE_STORE_VALIDATION_PROOF=COMPLETE_ONLY
```

## Deliberately rejected mechanical findings

A separate sealed `ValidatedFlightFeatures` type was not introduced. Inside one
trusted internal binary, such a wrapper cannot prove that validation actually ran;
it only relocates construction. The store instead validates the complete durable
validator report, including audit state, validator version, status and validation
time.

The executor constructor name and cohesive transaction-sized methods are not
correctness defects. Renaming or splitting them solely to satisfy a mechanical
word or line-count rule would increase churn without strengthening invariants.

## Verification

The guarded installer runs targeted tests, permanent audits, race tests, the full
backend test suite, Go vet and static contracts in a detached worktree before
modifying the working tree, then repeats all gates in the working tree.

```text
FEATURE_STORE_REVIEW_STATUS=CLOSED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
DATABASE_MIGRATION=NOT_REQUIRED
```

## Canonical remediation history

Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, and Continuous Integration evidence. Severity labels are retrospective.

### GFA-DATA-167 — Idempotent Feature Store writes did not prove semantic output identity

1. **Finding / symptom:** the snapshot key and input fingerprint could match while the semantic output payload differed.
2. **Root cause:** persistence identity proved input/reprocessing identity but not the durable analytical result itself.
3. **Failure scenario:** two executions under the same snapshot key and input fingerprint produce different feature values; the store accepts the later value as an idempotent replay instead of a conflict.
4. **Impact:** silent mutation or replacement of analytical history under an identity that claims equivalence.
5. **Severity rationale:** **P1 retrospective** because immutable analytical snapshots could become semantically inconsistent without detection.
6. **Existing guarantees violated:** immutable snapshot identity and deterministic replay.
7. **Considered solutions:** compare selected fields, hash raw domain JSON, or define a canonical versioned output payload and fingerprint.
8. **Chosen remediation:** every new stored record carries an output fingerprint derived from the canonical versioned persistence payload; replay requires input and output fingerprints to match.
9. **Why this solution was selected:** it makes semantic equivalence explicit and versionable without coupling to incidental runtime timestamps.
10. **Rejected alternatives:** treating input identity alone as sufficient and hashing operational timestamps that intentionally vary between equivalent runs.
11. **Trade-offs:** persistence code owns an additional deterministic canonicalization contract.
12. **Regression tests / protection:** conflict/idempotency tests and `featurestorereviewaudit`.
13. **Adversarial review findings:** operational timestamps were deliberately excluded because they describe execution timing, not analytical output.
14. **Remediation iterations:** implemented in `624fcf44d3260bd35ac32f67c0730689713198c0`.
15. **Residual risks and limitations:** a future semantic field must be added intentionally to the canonical persistence payload/fingerprint contract.
16. **Operational or deployment consequences:** existing rows remain readable; no PostgreSQL migration required.
17. **Exact evidence:** commit `624fcf44d3260bd35ac32f67c0730689713198c0`, store conflict tests, permanent review audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** durable analytical stores must prove both input identity and semantic output identity before declaring a replay idempotent.

### GFA-CONTRACT-168 — Feature Store serialized the Go domain model directly

1. **Finding / symptom:** persisted JSON shape was coupled directly to the in-process domain aggregate.
2. **Root cause:** no explicit versioned persistence data transfer object owned the storage contract.
3. **Failure scenario:** a benign Go refactor, field rename, or added exported field silently changes durable JSON compatibility or output fingerprints.
4. **Impact:** storage compatibility and replay behavior could drift with source-model changes.
5. **Severity rationale:** **P2 retrospective** because the defect threatens durable compatibility but does not by itself corrupt existing values.
6. **Existing guarantees violated:** stable persistence schema and explicit version evolution.
7. **Considered solutions:** freeze the domain struct forever, custom marshal methods on the domain type, or introduce a dedicated versioned persistence payload.
8. **Chosen remediation:** new rows use `flight-feature-snapshot-payload-v1` with stable field names, strict decoding, and a separate legacy decoder.
9. **Why this solution was selected:** decouples storage evolution from domain refactoring and makes compatibility policy explicit.
10. **Rejected alternatives:** direct `json.Marshal` of the domain aggregate and a breaking one-shot rewrite of legacy rows.
11. **Trade-offs:** mapping code must be maintained between domain and persistence representations.
12. **Regression tests / protection:** payload round-trip, unknown-field, legacy-read, and fingerprint tests.
13. **Adversarial review findings:** schema version one of the feature domain remains valid; persistence versioning is a separate storage concern.
14. **Remediation iterations:** `624fcf44d3260bd35ac32f67c0730689713198c0`.
15. **Residual risks and limitations:** future payload versions require explicit compatibility handling.
16. **Operational or deployment consequences:** no migration; legacy unversioned JSON remains readable.
17. **Exact evidence:** implementation commit `624fcf44d3260bd35ac32f67c0730689713198c0` and payload compatibility tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** persisted contracts must use explicit versioned payload owners instead of serializing evolving domain aggregates directly.

### GFA-DATA-169 — Memory and PostgreSQL stores enforced different write-validity contracts

1. **Finding / symptom:** in-memory and PostgreSQL implementations did not reject the same invalid snapshot inputs.
2. **Root cause:** validation logic was distributed across implementations instead of a shared conformance contract.
3. **Failure scenario:** tests pass against Memory Store while production PostgreSQL rejects or accepts materially different data.
4. **Impact:** environment-dependent correctness and misleading test confidence.
5. **Severity rationale:** **P1 retrospective** because divergent stores can publish inconsistent durable analytical state between test and production paths.
6. **Existing guarantees violated:** interchangeable Store semantics and test-production parity.
7. **Considered solutions:** test only PostgreSQL, duplicate validators carefully, or define shared validation helpers/conformance tests.
8. **Chosen remediation:** both stores enforce canonical trajectory UUID, strict `sha256:` fingerprints, non-finite rejection, nil-context rejection, and complete validation proof.
9. **Why this solution was selected:** preserves fast in-memory testing while requiring semantic equivalence with production storage.
10. **Rejected alternatives:** allowing Memory Store to remain a permissive test fake.
11. **Trade-offs:** Memory Store becomes stricter and slightly more complex.
12. **Regression tests / protection:** shared conformance tests, targeted Memory/PostgreSQL tests, and permanent audit.
13. **Adversarial review findings:** a sealed `ValidatedFlightFeatures` wrapper was rejected because type construction alone would not prove validation actually ran.
14. **Remediation iterations:** `624fcf44d3260bd35ac32f67c0730689713198c0`.
15. **Residual risks and limitations:** implementation-specific operational behavior may still differ where explicitly documented, but validity semantics must not.
16. **Operational or deployment consequences:** stricter test failures may surface previously hidden fixture defects.
17. **Exact evidence:** commit `624fcf44d3260bd35ac32f67c0730689713198c0`, Store conformance tests, PostgreSQL verifier.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every new Store implementation must satisfy the same semantic conformance suite before production use.

### GFA-DATA-170 — Feature Store writes did not require complete durable validation proof

1. **Finding / symptom:** a snapshot could reach a Store without proving that the current validator completed a full audit.
2. **Root cause:** validation state was treated as metadata rather than a mandatory persistence precondition.
3. **Failure scenario:** code constructs apparently valid feature values or a partial validation report and writes them without a complete validator run.
4. **Impact:** unvalidated or partially validated analytical snapshots become durable evidence.
5. **Severity rationale:** **P1 retrospective** because Feature Store is the trust boundary for immutable analytical output.
6. **Existing guarantees violated:** persistence only after successful complete validation.
7. **Considered solutions:** trust pipeline call order, add a wrapper type, or validate durable audit fields at the Store boundary.
8. **Chosen remediation:** new writes require complete validation audit proof including validator version, status, audit completeness, and validation time.
9. **Why this solution was selected:** verifies real evidence instead of trusting construction path or a nominal type.
10. **Rejected alternatives:** sealed wrapper as proof of execution and pipeline-only enforcement.
11. **Trade-offs:** Store depends on the durable validation-report contract.
12. **Regression tests / protection:** incomplete-audit rejection tests and `featurestorereviewaudit`.
13. **Adversarial review findings:** mechanical constructor/method-size suggestions were rejected as unrelated to trust-boundary integrity.
14. **Remediation iterations:** `624fcf44d3260bd35ac32f67c0730689713198c0`.
15. **Residual risks and limitations:** validator correctness itself is governed separately by the Validator review chain.
16. **Operational or deployment consequences:** malformed legacy test fixtures may require explicit compatibility setup; new writes fail closed.
17. **Exact evidence:** implementation commit `624fcf44d3260bd35ac32f67c0730689713198c0`, validation-proof tests, Feature Store audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** persistence boundaries must validate concrete proof of required upstream trust gates rather than assume call ordering.

### GFA-PERF-171 — Memory Feature Store growth was unbounded

1. **Finding / symptom:** the in-memory Store could retain an unlimited number of immutable snapshots.
2. **Root cause:** test/development storage had no explicit capacity contract.
3. **Failure scenario:** a long-running process or broad test workload continually writes unique snapshots and grows memory without limit.
4. **Impact:** memory exhaustion and unpredictable local/runtime behavior.
5. **Severity rationale:** **P2 retrospective** because this is bounded-resource reliability, not analytical corruption.
6. **Existing guarantees violated:** bounded process-local memory use.
7. **Considered solutions:** eviction, external storage, unlimited test-only semantics, or explicit capacity with fail-closed writes.
8. **Chosen remediation:** configurable record capacity; the Store fails when full rather than evicting immutable historical evidence.
9. **Why this solution was selected:** avoids silently deleting historical snapshots while still bounding memory.
10. **Rejected alternatives:** arbitrary eviction of immutable evidence and introducing external cache/storage infrastructure for the in-memory implementation.
11. **Trade-offs:** writers must handle capacity errors.
12. **Regression tests / protection:** capacity-bound tests and Feature Store review audit.
13. **Adversarial review findings:** preserving immutable history was prioritized over a convenience eviction policy.
14. **Remediation iterations:** `624fcf44d3260bd35ac32f67c0730689713198c0`.
15. **Residual risks and limitations:** configured capacity must be appropriate for the intended local/test workload.
16. **Operational or deployment consequences:** deterministic memory ceiling; PostgreSQL production capacity is governed separately.
17. **Exact evidence:** commit `624fcf44d3260bd35ac32f67c0730689713198c0` and Memory Store capacity tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** process-local stores and caches require an explicit bounded-resource policy.

### GFA-OPS-172 — Feature Store accepted nil caller contexts

1. **Finding / symptom:** Store methods silently substituted background context for nil caller context.
2. **Root cause:** convenience fallback overrode lifecycle ownership.
3. **Failure scenario:** a caller bug starts database or in-memory work without the intended cancellation/deadline chain.
4. **Impact:** leaked work and inconsistent context semantics across Store implementations.
5. **Severity rationale:** **P2 retrospective** as a lifecycle correctness defect.
6. **Existing guarantees violated:** explicit caller context ownership at persistence boundaries.
7. **Considered solutions:** fallback, panic, or typed error.
8. **Chosen remediation:** Memory and PostgreSQL Stores reject nil contexts explicitly.
9. **Why this solution was selected:** fail-fast without panics and aligns Store semantics with the rest of hardened repository/provider code.
10. **Rejected alternatives:** implicit `context.Background()` substitution.
11. **Trade-offs:** invalid callers fail earlier.
12. **Regression tests / protection:** nil-context conformance tests and permanent Store audit.
13. **Adversarial review findings:** cleanup-specific independent contexts remain a separate, explicitly named exception where required.
14. **Remediation iterations:** `624fcf44d3260bd35ac32f67c0730689713198c0`.
15. **Residual risks and limitations:** upstream code must still choose meaningful deadlines.
16. **Operational or deployment consequences:** clearer cancellation propagation; no migration.
17. **Exact evidence:** implementation commit `624fcf44d3260bd35ac32f67c0730689713198c0` and Store context tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** Store interfaces must reject nil contexts unless a separately documented cleanup context owns the operation.
