# Document 72 — Stage 14.30 PostgreSQL Correctness Hardening

Status: Implemented current-scope baseline
Project: Global Flight Analytics
Scope: close Ingestion Run evidence invariants, Route and Historical timestamp mirror drift, and cancelled-context rollback risk

## 1. Correctness scope

This increment closes three independently confirmed PostgreSQL correctness gaps:

```text
Ingestion Run processed counts and error evidence were not tied to status
Route Results did not compare timestamptz mirrors with exact Unix nanoseconds
Historical Aggregate Results did not compare four timestamptz mirrors with exact Unix nanoseconds
Airport Import, Flight State and Trajectory writes rolled back with the caller context
```

It does not claim that the remaining Stage 14 maintainability and Clean Code register is complete.

## 2. Ingestion Run evidence contract

Repository finalization now validates before issuing SQL:

```text
all counters are non-negative
records_inserted + records_updated <= records_received
success has no error message
failed and partial require a non-empty normalized error message
only terminal statuses can enter the completion validator
```

Migration `020_stage14_correctness_hardening.sql` applies the same rules as PostgreSQL check constraints. Direct SQL and future repository implementations therefore cannot bypass the evidence contract.

## 3. Exact timestamp ownership

Unix nanoseconds remain the canonical exact representation. PostgreSQL `timestamptz` columns remain operator-readable mirrors with microsecond precision.

Every Route Result read now selects and validates:

```text
as_of_time ↔ as_of_time_unix_nano
stored_at ↔ stored_at_unix_nano
```

Every Historical Aggregate read now selects and validates:

```text
window_start ↔ window_start_unix_nano
window_end ↔ window_end_unix_nano
as_of_time ↔ as_of_time_unix_nano
stored_at ↔ stored_at_unix_nano
```

A difference below one microsecond is accepted as PostgreSQL precision loss. A difference of one microsecond or more fails closed as corrupt persisted evidence.

Migration 020 also installs database constraints for all six mirror pairs.

## 4. Independent rollback context

Repository transaction rollback now uses one shared helper with a fresh bounded context derived from `context.Background()`.

The helper owns rollback for:

```text
AirportRepository.UpsertImported
FlightStateRepository.SaveFlightStates
TrajectoryRepository.saveTrajectory
```

Caller cancellation continues to stop normal work and commit. It no longer prevents the deferred rollback attempt from reaching PostgreSQL.

## 5. Production catalog and integration evidence

The permanent Stage 14 PostgreSQL gate:

```text
applies the complete production migration catalog 001–020
runs the migrator a second time to prove no pending migrations remain
runs repository, Feature Store, Route Store and Historical Aggregate package tests
runs an isolated-schema integration test through the real migrator
proves PostgreSQL rejects invalid counters, invalid error semantics and timestamp drift
```

GitHub Actions now inspects the catalog before application and verifies the absence of pending migrations only after application. A fresh continuous-integration database is no longer incorrectly rejected merely because its migrations are initially pending.

## 6. Regression protection

Permanent tests protect:

```text
repository fail-fast Ingestion Run validation
PostgreSQL Ingestion Run check violations
Route mirror selection and scan validation
Historical mirror selection and scan validation
sub-microsecond precision tolerance
one-microsecond corruption rejection
independent bounded rollback context
production migration 020 ownership
continuous-integration execution of all affected packages
```

## 7. Completion boundary

This increment closes the three correctness groups named in this document.

Stage 14 remains reopened for the separate confirmed backlog, including large PostgreSQL method decomposition, query-contract cleanup, pagination, nil-context policy, nullable helper semantics, migration-repair generalization, repeated SQL and Scan contours, and evidence-backed trajectory index profiling.
