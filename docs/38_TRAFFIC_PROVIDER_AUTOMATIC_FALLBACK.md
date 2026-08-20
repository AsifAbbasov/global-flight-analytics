# Traffic Provider Automatic Fallback

Status: Production Provider Recovery Baseline v2.0
Project: Global Flight Analytics

## Purpose

This document defines the policy-aware automatic provider chain used by the
production traffic ingestion code after the August 2026 provider-access incident.

## Current default and ordered chain

The default provider is `adsb.lol`. Automatic mode always starts with ADSB.lol.

Airplanes.live is added only when `AIRPLANES_LIVE_ACCESS_APPROVED=true`.
OpenSky is added only when Airplanes.live is not approved and
`OPENSKY_OPERATIONAL_AGREEMENT_CONFIRMED=true`. When neither condition is
satisfied, `auto` is intentionally a single-provider ADSB.lol path.

## Supported provider modes

```text
TRAFFIC_PROVIDER=adsb.lol
TRAFFIC_PROVIDER=airplanes.live
TRAFFIC_PROVIDER=opensky
TRAFFIC_PROVIDER=auto
```

## Provider access boundaries

### ADSB.lol

The public API is documented as free and open, with public API/data licensed
under ODbL 1.0. Published rate limiting is dynamic rather than a fixed hard
quota. Global Flight Analytics therefore applies a project-owned conservative
budget of one request per minute with `PROJECT-CONSERVATIVE` provenance.

ADSB.lol asks production users to make contact so an application is not broken
accidentally by upstream changes. Production GitHub Actions therefore contains a
fail-closed `ADSBLOL_PRODUCTION_CONTACT_CONFIRMED` gate.

### Airplanes.live

General free external API access was confirmed unavailable to this project in
August 2026. Airplanes.live remains supported only as an explicit
feeder/sponsorship-compatible path and cannot be selected unless
`AIRPLANES_LIVE_ACCESS_APPROVED=true`.

### OpenSky

Operational REST use remains subject to the provider's written-agreement
boundary. Runtime refuses direct OpenSky selection and refuses to add OpenSky to
`auto` unless `OPENSKY_OPERATIONAL_AGREEMENT_CONFIRMED=true`.

## Provenance and fallback behavior

A successful ingestion run records the actual selected source. The provider
chain remains behind provider policy, durable provider budget, request
coalescing, health-aware selection, HTTP response observation and canonical
state mapping.

Production ingestion remains disabled/fail-closed until ADSB.lol production-use
contact is confirmed and a real production smoke, freshness verification and
subsequent scheduled-run verification all pass.
