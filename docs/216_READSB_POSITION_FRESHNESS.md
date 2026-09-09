# Document 216 — readsb-Inspired Position Freshness

Status: implementation candidate pending exact-head repository CI and stacked-base reconciliation
Date: 2026-09-09
Scope: `R1_POSITION_FRESHNESS`

## Purpose

This increment adopts the useful `seen` versus `seen_pos` distinction from readsb as an independently implemented GFA evidence contract. It does not embed readsb, copy readsb C code, require SDR hardware, or add any paid provider or infrastructure.

## Evidence semantics

GFA now separates two timestamps:

```text
position_observed_at  = time represented by latitude/longitude
message_observed_at   = optional latest provider message/contact time
```

`observed_at` remains a compatibility alias for the canonical position observation time. Message freshness must never make an older position appear fresh.

Provider mappings:

```text
readsb-compatible seen_pos -> position_observed_at
readsb-compatible seen     -> message_observed_at
OpenSky time_position      -> position_observed_at
OpenSky last_contact       -> message_observed_at
```

If a readsb-compatible response does not expose `seen_pos`, the existing `seen`-based position timestamp remains as a compatibility fallback. Missing message time remains null. Historical rows are not backfilled because no evidence exists from which to reconstruct a historical latest-message timestamp safely.

## Persistence

Migration `030_add_flight_state_message_observation_time.sql` adds nullable `flight_states.message_observed_at`. Existing `flight_states.observed_at` remains the position timestamp used by current traffic, trajectories, replay, projection inputs and position-based analytics.

## Public traffic contract

`GET /api/v1/traffic/current` retains `observed_at` and adds:

```text
position_observed_at: RFC3339 timestamp
message_observed_at:  RFC3339 timestamp or null
```

The OpenAPI contract documents that `message_observed_at` cannot substitute for position freshness.

## Frontend

The selected-aircraft intelligence surface shows position observation time independently from optional latest-message time. Freshness is evaluated relative to the successful traffic-query response timestamp using the same existing project policy as the Traffic Data Quality Lens:

```text
recent position window     = 5 minutes
accepted future clock skew = 1 minute
browser Date.now evidence  = not used
```

User-visible states are `Fresh position`, `Older position`, `Clock-skew warning`, `Invalid position time`, and `Freshness unavailable`.

## Zero-budget boundary

```text
New paid provider       = NO
New server              = NO
New cache               = NO
Redis / Valkey          = NO
New runtime service     = NO
SDR hardware required   = NO
readsb runtime required = NO
Additional cost         = 0 RUB
```

## Regression protection

Permanent tests cover readsb `seen_pos`/`seen` separation, compatibility fallback, missing-message null semantics, OpenSky `time_position`/`last_contact` separation, migration non-backfill policy, cross-provider canonical compatibility, frontend freshness classification and shared project freshness thresholds.

## Merge boundary

This branch is stacked on the dependency-security hotfix branch while PR #174 remains open. Validation on the stacked head does not authorize a later rebased head. After #174 enters canonical `main`, this branch must be reconciled to that exact main revision and the full required CI/Vercel matrix must pass again before merge authorization can be requested.
