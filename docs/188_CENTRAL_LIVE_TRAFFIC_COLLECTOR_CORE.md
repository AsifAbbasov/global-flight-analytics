# Document 188 - Central Live Traffic Collector Core

Status: IMPLEMENTED CORE
Project: Global Flight Analytics
Date: 2026-08-18

## 1. Purpose

Realtime Flight Data Foundation introduced the hot in-memory current-state store and the public live snapshot API. This increment adds the provider-agnostic acquisition loop that will eventually populate that store from one centralized upstream source.

The collector is intentionally separated from provider construction. Existing provider adapters, regional-provider orchestration, request coalescing, budget enforcement, fallback decisions, response evidence and provider health remain the canonical upstream ownership boundaries.

## 2. Architecture

```text
existing regional provider
          |
          v
Central Live Collector
  - lifecycle cancellation
  - request timeout
  - target sequencing
  - retry-at evidence
  - exponential backoff
  - positive jitter
  - status/health snapshot
          |
          v
existing live.Store
          |
          v
GET /api/v1/traffic/live
```

The browser never calls the aviation provider.

## 3. Provider policy correction

Airplanes.live currently publishes:

```text
REST limit: 1 request / second
Free tier: 500 requests / day
point radius: up to 250 nautical miles
```

The canonical repository policy previously represented only the one-request-per-second limit. This increment records the 500-request daily free-tier limit as an additional source-backed fixed window.

For one target:

```text
24h / 500 = 172.8 seconds minimum average request spacing
```

Therefore Airplanes.live is not treated as a free rapid-live 5-10 second production source.

Official references:

```text
https://airplanes.live/api/
https://airplanes.live/api-guide/
```

## 4. adsb.lol activation boundary

The adsb.lol public API currently states that:

```text
the API can be used for free
public data/API license is ODbL
/v2/point/{lat}/{lon}/{radius} supports up to 250 nm
future API access may require an API key obtained by feeding
production users should contact the operator
```

Official references:

```text
https://api.adsb.lol/api/openapi.json
https://www.adsb.lol/docs/open-data/api/
```

This makes adsb.lol a strong zero-cost rapid-live candidate, but production activation remains open until the operator-contact dependency is resolved and encoded in repository policy.

## 5. OpenSky boundary

OpenSky remains behind the existing explicit operational-agreement gate. The collector core does not bypass that gate.

Official reference:

```text
https://opensky-network.org/about/terms-of-use
```

## 6. Collector semantics

The collector:

- accepts the existing provider LoadByPoint contract;
- supports one or more named geographic targets;
- executes targets sequentially;
- supports request spacing between targets;
- stops the current cycle after a provider failure rather than multiplying failed upstream requests;
- writes only provider-returned canonical FlightState observations into live.Store;
- preserves source and observation timestamps through the existing live-store adapter;
- uses exponential failure backoff capped by configuration;
- honors provider RetryAtTime evidence when it requires a longer delay;
- adds only non-negative jitter, so jitter never intentionally makes the configured cadence faster;
- exposes immutable status snapshots for future metrics and readiness work;
- exits cleanly when the server lifecycle context is cancelled.

## 7. Explicit non-goals

This increment does not:

- activate a new production provider;
- add WebSocket or Server-Sent Events;
- add Redis, Kafka, Kubernetes or a microservice;
- make an upstream request per browser;
- synthesize intermediate aircraft observations;
- persist animation frames;
- change PostgreSQL trajectory semantics;
- change the public live snapshot contract.

## 8. Status

```text
CENTRAL_LIVE_COLLECTOR_CORE=CLOSED
AIRPLANES_LIVE_FREE_DAILY_BUDGET=500
AIRPLANES_LIVE_RAPID_FREE_POLLING=NOT_SUPPORTED
ADSB_LOL_PRODUCTION_POLICY_REVIEW=OPEN
RAPID_LIVE_PROVIDER_PRODUCTION_ACTIVATION=OPEN
CENTRAL_COLLECTOR_SERVER_WIRING=OPEN
FRONTEND_REALTIME_INTEGRATION=OPEN
```

The next increment is provider activation and server composition: select a policy-compliant live provider, reuse the existing provider orchestration, construct the collector once inside the server process, bind its lifetime to the server context, and expose collector health through observability without changing the browser API.
