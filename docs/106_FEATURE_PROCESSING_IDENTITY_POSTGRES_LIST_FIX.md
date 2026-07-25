# Document 106 — Feature Processing Identity PostgreSQL List Fix

Status: IMPLEMENTED
Baseline commit: ab452c0cd039619e842c1991ec1bed10a42e5665

## Continuous Integration failure

Backend Continuous Integration run `30151121170` applied migration `026`
successfully. The PostgreSQL correctness step then failed before the feature
pipeline verifier.

The non-cursor PostgreSQL feature-snapshot list query accepted only three SQL
parameters:

```text
trajectory_id
schema_version
limit
```

The production caller already supplied four parameters:

```text
trajectory_id
schema_version
processing_version
limit
```

The SQL query therefore omitted the processing-version predicate and reused
placeholder `$3` for the limit.

## Correction

The non-cursor query now enforces:

```sql
AND processing_version = $3
LIMIT $4
```

A regression test owns the SQL placeholder layout and the production argument
order. The permanent feature-processing-identity audit also requires both
fragments.

## Closure condition

This correction is locally complete only after all tests, race tests, `go vet`
and architecture audits pass. Final FP-02 closure still requires a green
PostgreSQL 16 Integration job on the corrective commit.

```text
FEATURE_PROCESSING_IDENTITY_POSTGRES_LIST_FIX=ENFORCED
```

## Canonical UUID test contract

The PostgreSQL store normalizes trajectory UUID text before binding SQL
arguments. The regression test therefore compares the first argument with the
lowercase canonical form rather than the uppercase fixture spelling.
