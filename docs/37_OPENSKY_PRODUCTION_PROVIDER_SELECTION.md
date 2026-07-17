# Document 37 — OpenSky Production Provider Selection

Status: Implementation Foundation v1.0
Project: Global Flight Analytics
Scope: Controlled runtime selection between free regional traffic providers

## Purpose

This increment makes the bounded OpenSky client usable by the production ingestion command without making OpenSky mandatory and without claiming project-owned surveillance coverage.

## Runtime selection

The ingestion daemon accepts:

```text
TRAFFIC_PROVIDER=airplanes.live
TRAFFIC_PROVIDER=opensky
```

The default remains `airplanes.live`. OpenSky must be explicitly selected.

## OpenSky production path

```text
Regional point and radius in nautical miles
↓
Bounded OpenSky bounding box
↓
Provider budget and request coalescing
↓
OpenSky State Vector request
↓
Fifteen-second provider validity gate
↓
Canonical FlightState mapping
↓
Existing data quality and trajectory pipeline
↓
PostgreSQL
```

## Free-data controls

- A regional radius greater than 250 nautical miles is rejected.
- Requests that would exceed the configured regional credit boundary are rejected.
- Missing and stale OpenSky positions are not reconstructed as observed positions.
- Anonymous polling cannot be configured below ten seconds.
- Authenticated polling cannot be configured below five seconds.
- OAuth2 credentials are optional but must be configured as a complete pair.
- Provider response headers update the existing provider-reported budget controller.
- Transport and HTTP outcomes are reported to existing provider health collection.

## Environment variables

```text
TRAFFIC_PROVIDER=airplanes.live
OPENSKY_BASE_URL=https://opensky-network.org/api
OPENSKY_TOKEN_URL=https://auth.opensky-network.org/auth/realms/opensky-network/protocol/openid-connect/token
OPENSKY_CLIENT_ID=
OPENSKY_CLIENT_SECRET=
OPENSKY_TIMEOUT=15s
OPENSKY_POLLING_INTERVAL=10s
```

For authenticated OpenSky access, the default polling interval becomes five seconds.

## Explicit boundary

This increment provides deterministic provider selection. It does not yet implement automatic runtime fallback between `airplanes.live` and OpenSky. Automatic fallback requires ingestion-run provenance to record the provider that actually served each request. That is the next provider integration increment.
