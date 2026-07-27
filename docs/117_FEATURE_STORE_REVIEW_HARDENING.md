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
