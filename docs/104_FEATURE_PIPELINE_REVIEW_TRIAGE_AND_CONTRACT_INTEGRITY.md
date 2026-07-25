# Document 104 — Feature Pipeline Review Triage and Contract Integrity

Status: IMPLEMENTED
Baseline commit: 48f274754fa0fbdbe4ed0a2b8f95985f38183629

## 1. Review baseline correction

The supplied review was performed against commit `bb9f351`.

The current repository no longer classifies
`internal/features/featurepipeline` as an unintegrated package. The production
command `cmd/materialize-flight-features` composes and executes the PostgreSQL
feature pipeline. Therefore the old BDUF/YAGNI finding is rejected as stale.

## 2. Findings fixed in this increment

```text
FP-03 Result no longer owns a second FlightFeatures value.
FP-04 Validation status copies are treated as an enforced invariant.
FP-05 Incomplete or contradictory validation reports are rejected, while identical codes remain valid on distinct issue paths.
FP-06 Pipeline depends on a narrow FeatureWriter interface.
FP-08 PostgreSQL Pool and Executor are mutually exclusive.
FP-10 Nil context is rejected instead of replaced.
FP-11 Typed-nil dependencies are rejected.
FP-12 PostgreSQL verifier is executed by Continuous Integration.
FP-13 PostgreSQL composition version is part of the version manifest.
FP-14 Materializer tests and the PostgreSQL verifier consume stored features through the new Result contract.
```

## 3. Deliberately rejected mechanical observations

The following are not correctness rules:

```text
a constructor returning nil together with an error;
the word With in a constructor name;
a fixed forty-line or fifty-line function threshold;
nil returned by Unwrap when the receiver itself is nil.
```

They can be discussed as style preferences, but they must not be recorded as
production blockers without a concrete failure mode.

## 4. Remaining schema-level blocker

Processing identity is still absent from the durable snapshot key.

This requires a dedicated migration because the PostgreSQL uniqueness
constraint, row scanning, record identity, memory-store key, compatibility
behavior for existing snapshots, and materializer replay semantics must change
atomically.

```text
FP-02_PROCESSING_IDENTITY_STATUS=OPEN
NEXT_INCREMENT=FEATURE_SNAPSHOT_PROCESSING_IDENTITY
```

The current increment intentionally does not hide this debt behind a source-only
fingerprint change.
