# Document 87 — Ingestion Durability, Replay and Partial Status Hardening

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics

## Purpose

This increment closes three related correctness findings from the original
Ingestion, Provider Adapters and Orchestration review:

```text
durable ingestion evidence before provider transport
replay-safe Flight State persistence
explicit partial completion after durable observation writes
```

It also repairs the migration catalog version collision introduced when the
OurAirports publication lifecycle was committed as a second migration 019.

## Durable pre-request ingestion run

`LoadAndProcessByPoint` now commits the `running` ingestion row before invoking
any provider method. Therefore a process crash, container termination or panic
inside provider transport leaves a durable row that startup stale-run recovery
can finalize.

The row initially records the provider chain source. After fallback selection,
the source is updated while the run is still `running` so the final evidence
records the provider that actually supplied the observations.

When orchestration proves that a local budget or polling denial did not execute
an external request, the provisional `running` row is deleted. This preserves
the previous rule that local denial must not become a false failed provider run.
Deletion is allowed only for a still-running row with no linked Flight States.

## Partial terminal status

The repository port and PostgreSQL repository now expose `MarkPartial`.

The ingestion service uses:

```text
StoredFlightStateCount > 0 and later processing error → partial
StoredFlightStateCount = 0 and processing error       → failed
```

This aligns the run status with the existing independent durability-unit
contract: source observations may commit before quality or trajectory
derivations fail and enter reconciliation.

## Replay-safe Flight State identity

Migration 023 installs the provider observation identity:

```text
(source_name, icao24, observed_at)
```

The migration fails closed when historical duplicates already exist rather than
silently deleting observations or cascading deletion into quality evidence.

The production insert uses:

```sql
ON CONFLICT (source_name, icao24, observed_at)
DO NOTHING
```

`FlightStateRepository.SaveFlightStatesCounted` returns the number of rows
actually inserted. The application service detects this optional capability and
propagates the real count into `ProcessAndStoreResult` and the ingestion run.
The legacy repository interface remains available for non-PostgreSQL tests and
adapters.

## Migration catalog repair

The canonical sequence is now:

```text
019_data_quality_parent_integrity.sql
020_stage14_correctness_hardening.sql
021_trajectory_query_profiles.sql
022_provider_publication_lifecycle.sql
023_ingestion_durability_replay_partial.sql
```

A permanent production-catalog regression test rejects duplicate versions,
sequence gaps and future accidental renaming of these ownership boundaries.

## Verification

The increment includes:

- ordering tests proving durable run creation precedes provider invocation;
- fail-fast tests proving provider transport does not start when run creation
  fails;
- provisional-run deletion tests for local access denial;
- partial-versus-failed terminal status tests;
- selected fallback source update coverage;
- counted replay persistence tests;
- PostgreSQL active-run lifecycle integration coverage;
- replay-safe SQL contract checks;
- migration catalog uniqueness and sequence checks;
- race detector coverage;
- full backend tests;
- `go vet`;
- code review policy gates;
- rollback-safe installation.

## Remaining review boundary

This document does not claim complete Ingestion review closure. Remaining
separate work includes exact in-memory deduplication semantics, Airplanes.live
nullable telemetry and conversion bounds, duration overflow protection,
malformed-record batch policy, provider constructor validation, and explicit
classification of multi-instance rate limiting and health-aware fallback.
