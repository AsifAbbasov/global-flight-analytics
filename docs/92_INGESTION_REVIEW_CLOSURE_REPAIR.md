# Document 92 — Ingestion Review Closure Repair

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics
Baseline: `b7bf2b762290e55a45fa8d40641435248d1aeddf`

## 1. Scope

This increment closes the remaining verified findings from the Ingestion,
Provider Adapters and Orchestration review without changing the approved
modular-monolith architecture or adding runtime infrastructure.

The repaired boundaries are:

```text
Retry-After duration overflow protection
OpenSky OAuth expires_in duration overflow protection
Open-Meteo missing metric preservation
PostgreSQL NULL persistence for unavailable weather metrics
typed atomic fail-whole OurAirports publication parsing
PostgreSQL isolated fixture migration alignment
accurate review closure wording
```

## 2. Bounded external duration conversion

Both numeric `Retry-After` parsing paths reject values that cannot be represented
as `time.Duration` before multiplying by `time.Second`.

OpenSky OAuth token responses use the same fail-closed rule for `expires_in`.
A missing or non-positive lifetime still receives the existing 1,800-second
engineering default. A positive value above the representable duration limit is
classified as `ErrTokenResponseInvalid`.

## 3. Open-Meteo metric availability

Open-Meteo current-weather values decode through nullable pointers. The adapter
therefore distinguishes:

```text
JSON 0    → available metric with value zero
JSON null → unavailable metric
missing   → unavailable metric
```

`weather.CurrentSnapshot` carries explicit metric availability while preserving
legacy in-process callers: snapshots without explicit availability retain the
previous all-metrics-available interpretation.

The current weather Hypertext Transfer Protocol response uses JSON `null` for
unavailable metrics. It does not publish invented zero observations.

## 4. PostgreSQL weather persistence

Migration 025 formalizes nullable weather metric columns and records the meaning
of `NULL` through column comments.

`WeatherRepository` validates only available metrics and sends unavailable
metrics to PostgreSQL as `NULL`. Existing complete snapshots keep their previous
validation and persistence behavior.

Weather Context production selection already requires every weather metric to be
non-NULL. Incomplete snapshots therefore remain stored as truthful source
evidence but are not promoted into complete Weather Context analysis.

## 5. OurAirports atomic publication policy

OurAirports remains an atomic fail-whole publication import.

A malformed CSV row or a row that violates the airport parsing contract returns
the typed `AtomicPublicationError`, matches
`ErrAtomicPublicationRejected`, and returns no partial airport collection.

This is a deliberate publication policy, not an accidental parser side effect.

## 6. PostgreSQL fixture alignment

The Flight State altitude fixture now applies both migrations required by the
current repository insert contract:

```text
006_flight_state_altitude_semantics.sql
023_ingestion_durability_replay_partial.sql
```

The provider publication fixture now uses the canonical renamed migration:

```text
022_provider_publication_lifecycle.sql
```

These changes repair isolated integration schemas. They do not modify production
migration history.

## 7. Formal closure gates

The review may be marked closed only when a new commit containing this increment
passes all of the following:

```text
Go formatting
complete backend tests
Go vet
project architecture and contract audit
code review policy audit
Stage 14 final audit
critical race tests
PostgreSQL 16 migration apply and replay
PostgreSQL correctness integration tests
backend container verification
```

Local tests without `TEST_DATABASE_URL` do not independently prove the PostgreSQL
gate. A successful GitHub Backend Continuous Integration run on the new commit is
the final closure evidence.

## 8. Closure statement

When every gate in Section 7 passes on the same new commit:

```text
Open technical findings: 0
Unclassified findings: 0
Ingestion review: CLOSED
Release decision: ACCEPTABLE
```

Until then, the correct status remains:

```text
Ingestion review: CONDITIONALLY ACCEPTABLE
Formal closure: PENDING CONTINUOUS INTEGRATION
```
