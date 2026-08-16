# Document 187 — Realtime Flight Data Foundation

Status: IMPLEMENTED — scoped local verification complete; full release gate deferred by pre-existing frontend redesign work
Project: Global Flight Analytics

## Purpose

The original production traffic path was intentionally optimized for durable research analytics: scheduled provider ingestion persisted observations to PostgreSQL and the API read the newest displayable durable snapshot. That remains valuable for history, trajectory construction and analytics, but it is not an appropriate transport path for a smooth live-flight map.

This increment introduces a separate realtime foundation without replacing the mature durable analytics architecture.

## Architecture

```text
                    governed live providers
                            |
                            v
                      normalization
                            |
                            v
                    current-state store
                     /             \
                    /               \
                   v                 v
          live snapshot API      durable sampler
          GET /traffic/live          |
                   |                 v
                   v             PostgreSQL
                browser              |
                   |                 v
                   v             analytics/history
          client interpolation
```

The live path and durable path deliberately have different responsibilities. PostgreSQL remains the authority for persisted observation history, trajectories and analytical evidence. It is not used as the per-frame realtime transport.

## Current-state contract

The in-process bounded store keeps one latest normalized state per ICAO24 identifier and provides:

- deterministic newest-observation replacement and optional source priority at equal observation times;
- explicit observed-at and received-at timestamps;
- per-item source provenance;
- nullable telemetry so unavailable data is not confused with valid zero values;
- bounded capacity and deterministic oldest-state eviction;
- time-to-live stale-state eviction;
- deterministic snapshots with server time and monotonically increasing sequence number;
- optional simple geographic bounding-box filtering;
- explicit selected-aircraft inclusion even when the aircraft is outside the requested bounding box;
- bounded result limits and selected-aircraft limits.

The public `GET /api/v1/traffic/live` response is intentionally compact. Heavy aircraft profiles, trajectories, route intelligence, weather context and historical analytics remain separate APIs.

Client interpolation must be treated as display-only estimated motion between observations. Interpolated positions are never written back as provider observations and must never be presented as measured evidence.

## Zero-cost and provider-policy boundary

This architecture itself requires no paid infrastructure, Redis, Kafka, Kubernetes or new microservice topology.

Production source activation remains governed by provider terms and quota rather than by technical capability alone:

- Airplanes.live remains the currently configured production ingestion provider, but its public free API quota is not sufficient to honestly claim a continuous 5–10 second production polling cadence for this project. The existing durable production scheduler therefore remains the active path while the live foundation is built.
- OpenSky is fail-closed for `opensky` and `auto` provider modes unless `OPENSKY_OPERATIONAL_AGREEMENT_CONFIRMED=true` is explicitly configured after the required written operational agreement has actually been obtained.
- adsb.lol or any future community source may be added behind the existing provider adapter only after its operational-use expectations have been coordinated and accepted. This document does not activate such a provider.

Provider policy references used for this engineering decision:

- https://airplanes.live/api/
- https://opensky-network.org/about/terms-of-use
- https://api.adsb.lol/

## Durable sampling boundary

The existing Cloudflare and GitHub Actions production ingestion topology is retained as the durable sampling lane. This increment does not increase its provider request cadence and does not write browser animation frames to PostgreSQL.

A future approved zero-cost rapid collector can populate the same `Store` through `UpsertFlightStates` without changing the public live snapshot contract. That collector must run in a process that shares the current-state store or use another explicitly approved shared-state design; the current external scheduler cannot populate Render process memory directly.

## Analytical effect

The new architecture can improve freshness and, when a policy-compliant source supports denser sampling, provide richer raw evidence for trajectory reconstruction, flight-phase changes, route inference, traffic density, airport proximity and freshness/coverage analysis. It does not manufacture more accurate sensor measurements than the source actually supplies.

## Scope and status

```text
REALTIME_CURRENT_STATE_FOUNDATION=CLOSED
LIVE_SNAPSHOT_CONTRACT=CLOSED
ZERO_COST_LIVE_SOURCE_POLICY=CLOSED
DURABLE_SAMPLING_BOUNDARY=PRESERVED
RAPID_LIVE_PROVIDER_PRODUCTION_ACTIVATION=OPEN
FRONTEND_REALTIME_INTEGRATION=OPEN
```

The next product-facing step after exact source verification and a policy-compliant live-provider activation is to connect the frontend to this contract and implement GPU rendering plus client interpolation. The existing frontend is intentionally untouched by this increment.
