# Document 107 — Feature Processing Identity Test Fixture Alignment

Status: IMPLEMENTED
Baseline commit: f18d43689d53301db862bc10c0445c90dc6f277d

## Continuous Integration evidence

Backend Continuous Integration run `30157732731` applied the complete production
migration catalog successfully. The PostgreSQL correctness step then failed in:

```text
TestPostgresStorePreservesExactNanosecondsAndRejectsMirrorDrift
```

The exact PostgreSQL error was:

```text
column "processing_version" of relation "flight_feature_snapshots" does not exist
SQLSTATE 42703
```

## Root cause

The timestamp-consistency integration test creates an isolated PostgreSQL schema
and manually creates `flight_feature_snapshots`. Its fixture still represented
the pre-processing-identity schema:

- no `processing_version` column;
- uniqueness on trajectory, schema and as-of time only.

The production migration was correct, but it could not affect the isolated
schema created later by the test.

## Correction

The fixture DDL is now a named source contract and includes:

```sql
processing_version text NOT NULL
```

Its uniqueness contract now includes:

```text
trajectory_id
schema_version
processing_version
as_of_time_unix_nano
```

A unit test owns these fragments, and the permanent feature-processing-identity
audit verifies that the PostgreSQL integration fixture remains processing-aware.

## Closure condition

Final FP-02 closure still requires a green Backend Continuous Integration run on
the corrective commit, including the PostgreSQL feature-pipeline verifier.

```text
FEATURE_PROCESSING_IDENTITY_TEST_FIXTURE_ALIGNMENT=ENFORCED
```
