# Document 189 - Rapid Live Provider Activation and Server Wiring

Status: CODE-COMPLETE / PRODUCTION ACTIVATION BLOCKED BY EXTERNAL CONFIRMATION
Project: Global Flight Analytics
Date: 2026-08-18

## 1. Decision

The central live collector is now composable inside the existing Go API process. No second service, message broker, Redis instance, Kubernetes deployment or per-browser provider request path is introduced.

The selected rapid-live candidate is `adsb.lol`.

## 2. Current provider evidence

As of 2026-08-18 the official adsb.lol API documentation states:

- the API can be used for free;
- public API/data is licensed under ODbL;
- `/v2/point/{lat}/{lon}/{radius}` supports point queries up to 250 nautical miles;
- a future API key may be required and obtained through feeding;
- production users should contact the operator so API changes do not accidentally break their application.

Official references:

```text
https://api.adsb.lol/api/openapi.json
https://www.adsb.lol/docs/open-data/api/
https://github.com/adsblol/api
```

The provider repository describes rate limiting as dynamic. Therefore GFA does not invent a provider-owned quota.

## 3. GFA application safety cap

GFA applies its own conservative policy:

```text
6 requests / minute
360 requests / hour
8640 requests / day
collector interval >= 10 seconds
```

These are application-defined safety limits, not source-backed provider limits.

## 4. Fail-closed production activation

The live collector defaults to disabled.

Enabling requires:

```text
LIVE_TRAFFIC_COLLECTOR_ENABLED=true
LIVE_TRAFFIC_PROVIDER=adsb.lol
ADSB_LOL_PRODUCTION_CONTACT_CONFIRMED=true
```

The confirmation variable must only be set after the production-contact dependency has actually been resolved.

Until then, `LIVE_TRAFFIC_COLLECTOR_ENABLED=false` remains the production-safe configuration.

## 5. Runtime topology

```text
adsb.lol point endpoint
        |
        v
adsb.lol adapter
        |
        v
existing providerresponse
        |
        v
existing providerbudget (durable PostgreSQL state)
        |
        v
existing ingestionorchestrator
        |
        v
existing regionalprovider
        |
        v
Central Collector
        |
        v
existing live.Store
        |
        v
GET /api/v1/traffic/live
        |
        v
browser
```

Browsers never call adsb.lol directly.

## 6. Canonical data semantics

The adapter maps only provider-returned observations into canonical `FlightState`.

It preserves ICAO24, callsign, coordinates, altitude semantics, ground speed, track, vertical rate, on-ground semantics, observation time and source provenance. Missing position is rejected rather than converted into zero coordinates.

No intermediate animation positions are fabricated by the server.

## 7. Server lifecycle

When enabled, the collector is created once in `cmd/server`, receives the same root cancellation context as the API process, and writes to the exact `live.Store` passed into the HTTP server.

Provider outages do not crash the API. The collector retains its existing exponential backoff and stale-store eviction behavior. Configuration or composition failures remain startup failures.

## 8. Remaining external step

Before actual Render activation:

1. contact adsb.lol operator for production-use notice;
2. retain the response/evidence;
3. set `ADSB_LOL_PRODUCTION_CONTACT_CONFIRMED=true`;
4. configure Baku live target environment variables;
5. deploy the exact revision;
6. verify provider metrics and `/api/v1/traffic/live` freshness in production.

## 9. Status

```text
ADSB_LOL_ADAPTER=CLOSED
CENTRAL_COLLECTOR_SERVER_WIRING=CLOSED
PROVIDER_ORCHESTRATION_REUSED=YES
UPSTREAM_REQUESTS_PER_BROWSER=NO
ADSB_LOL_APPLICATION_CAP=6_PER_MINUTE
ADSB_LOL_MINIMUM_COLLECTOR_INTERVAL=10_SECONDS
ADSB_LOL_PRODUCTION_CONTACT_CONFIRMED=REQUIRED_FOR_ENABLE
RAPID_LIVE_PROVIDER_PRODUCTION_ACTIVATION=BLOCKED_EXTERNAL_CONFIRMATION
FRONTEND_REALTIME_INTEGRATION=OPEN
```
