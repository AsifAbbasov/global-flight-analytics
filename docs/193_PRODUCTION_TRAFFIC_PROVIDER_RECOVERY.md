# Production Traffic Provider Recovery

Status: Implementation foundation complete; production activation pending external confirmation and runtime evidence
Date: 2026-08-21
Scope: policy-aware recovery of production traffic ingestion after the Airplanes.live access incident

## 1. Recovery objective

Restore a valid open-data production traffic provider path without bypassing
provider terms, hiding source failures, or weakening existing ingestion,
provenance, budget, health, quality, trajectory and monitoring controls.

## 2. Candidate provider

ADSB.lol is the recovery candidate because its current public documentation
states that the public API is free, public API/data are ODbL 1.0, the v2 point
endpoint supports bounded regional requests, rate limits are dynamic, future API
keys may be tied to feeding, and production users should contact the operator.

Canonical references:

```text
https://api.adsb.lol/
https://www.adsb.lol/docs/open-data/api/
https://www.adsb.lol/privacy-license/
```

## 3. Implementation

The repository contains a dedicated `adsblol` adapter based on the same hardened
readsb-compatible HTTP, decoding, altitude, telemetry, malformed-batch and
provider-observation behavior already used by the Airplanes.live adapter.

Runtime provider modes are `adsb.lol`, `airplanes.live`, `opensky` and `auto`.
The default is `TRAFFIC_PROVIDER=adsb.lol`.

Automatic selection starts with ADSB.lol and only adds an explicitly approved
secondary provider.

## 4. Budget policy

ADSB.lol publishes dynamic load-based rate limiting rather than a stable hard
quota. Global Flight Analytics therefore applies:

```text
provider=adsb.lol
budget_mode=fixed-window
max_requests=1
window=minute
provenance=PROJECT-CONSERVATIVE
```

This is a project safety cap, not a claim about an upstream hard quota.

## 5. Fail-closed production activation

Production GitHub Actions refuses to execute unless:

```text
ADSBLOL_PRODUCTION_CONTACT_CONFIRMED=true
```

Airplanes.live additionally requires `AIRPLANES_LIVE_ACCESS_APPROVED=true`.
OpenSky additionally requires `OPENSKY_OPERATIONAL_AGREEMENT_CONFIRMED=true`.

## 6. Render boundary

The Render web service runs `/app/server`, not the ingestion command. Stale
`TRAFFIC_PROVIDER` and `AIRPLANES_LIVE_TIMEOUT` values are therefore removed from
`render.yaml`; provider configuration belongs to the ingestion runtime.

## 7. Production activation criteria

Provider recovery is not closed until:

```text
ADSBLOL_PRODUCTION_CONTACT=CONFIRMED
PRODUCTION_TRAFFIC_INGESTION_WORKFLOW=ENABLED
PRODUCTION_ADSBLOL_SMOKE=PASS
PRODUCTION_TRAFFIC_FRESHNESS=PASS
SUBSEQUENT_SCHEDULED_RUN=PASS
GRAFANA_FRESHNESS_RECOVERY=PASS
PRODUCTION_PROVIDER_RECOVERY=CLOSED
```

Until then:

```text
ADSBLOL_ADAPTER=IMPLEMENTED
ADSBLOL_PROVIDER_POLICY=IMPLEMENTED
PRODUCTION_WORKFLOW_SOURCE_READY=YES
PRODUCTION_WORKFLOW_CONTACT_GATE=ACTIVE
PRODUCTION_INGESTION=INTENTIONALLY_OFFLINE
PRODUCTION_PROVIDER_RECOVERY=OPEN
```


<!-- PROVIDER-CONFIG-AND-COMPATIBILITY-HARDENING-V1 -->
## 9. Provider configuration and compatibility hardening

Provider configuration is intentionally split into nested provider-owned
configuration blocks:

```text
TrafficProviderConfig
├── ADSBLOL
├── AirplanesLive
└── OpenSky
```

Runtime access eligibility is centralized on `TrafficProviderConfig` through
`RequireEligible` and `AutomaticCandidates`. The provider factory consumes that
policy instead of owning provider-specific agreement/access rules.

The ADSB.lol and Airplanes.live public v2 adapters currently remain separate.
This is deliberate until a real ADSB.lol production response provides evidence
that the overlapping contracts are stable enough for a shared protocol
abstraction. A cross-provider compatibility test asserts identical canonical
semantics for the overlapping v2 fields while preserving distinct source
identity.

This avoids both failure modes:

- duplicated provider-selection/access policy;
- premature protocol abstraction based only on endpoint similarity.

A shared `readsbv2` transport/model layer may be extracted after live response
evidence confirms the contract boundary.
