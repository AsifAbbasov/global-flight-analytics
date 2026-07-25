# Document 105 — Feature Snapshot Processing Identity

Status: IMPLEMENTED
Baseline commit: 312afe2b9ddcc05da0c2068e50c05e0741a7a1c1

## Problem

The previous durable snapshot identity used only trajectory, feature schema and
as-of time. Different processing contracts could collide in one snapshot slot.

## Contract

Every snapshot now carries processing version in provenance, the in-memory key,
record identifier, PostgreSQL row, uniqueness boundary and read filters.

Existing rows are backfilled as:

```text
flight-feature-processing-legacy-v1
```

New production materialization uses:

```text
flight-feature-processing-pipeline-v1
```

Blank versions at old internal call sites normalize to the current version for
source compatibility. Explicit legacy and future versions remain readable.

```text
FP-02_PROCESSING_IDENTITY_STATUS=CLOSED
FEATURE_SNAPSHOT_PROCESSING_IDENTITY=ENFORCED
```

## Legacy compatibility

The migration updates both the relational processing-version column and the
embedded JSON provenance. Existing identifiers are accepted through the
documented legacy identifier algorithm; new identifiers include processing
version.

## Continuous verification

The PostgreSQL verifier stores two records with the same trajectory, schema,
as-of time and input fingerprint but distinct processing versions.

## Migration catalog ownership

Migration `026_flight_feature_processing_identity.sql` is registered in the
canonical production migration catalog. The catalog regression test requires
exactly twenty-six contiguous migrations and verifies the canonical filename.

## Semantic contract audit

The feature-pipeline contract audit parses the Go syntax tree. It validates the
types of `Config.Writer` and `Config.ProcessingVersion` without depending on
gofmt whitespace alignment.
