# Production Traffic Provider Recovery

Status: Provider response received; compliance hardening in review; runtime activation still pending
Date: 2026-08-23
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

## 3. Operator response

The production-use inquiry sent on 2026-08-21 received an operator reply on the
same date. The reply did not promise a bespoke quota, SLA, dedicated support, API
stability, accuracy, completeness or freshness. It supplied general application
requirements applicable to the requested usage pattern:

- visible ODbL attribution in the application;
- an identifiable User-Agent with contact information;
- respect for HTTP errors and rate limits;
- respect for `Retry-After` when provided;
- caching and deduplication where practical;
- appropriately scoped geographic queries;
- no rate-limit circumvention through multiple IP addresses;
- design for future API/authentication/limit changes;
- no safety-critical dependency on the volunteer-run service.

The response did not object to the project description or bounded regional use,
but it is intentionally recorded as provider guidance rather than a guaranteed
quota or service commitment.

## 4. Implementation

The repository contains a dedicated `adsblol` adapter based on the same hardened
readsb-compatible HTTP, decoding, altitude, telemetry, malformed-batch and
provider-observation behavior already used by the Airplanes.live adapter.

Runtime provider modes are `adsb.lol`, `airplanes.live`, `opensky` and `auto`.
The default is `TRAFFIC_PROVIDER=adsb.lol`.

Automatic selection starts with ADSB.lol and only adds an explicitly approved
secondary provider.

Provider-response compliance hardening adds:

```text
IDENTIFIABLE_USER_AGENT=REQUIRED
PUBLIC_CONTACT_URL=https://github.com/AsifAbbasov/global-flight-analytics
VISIBLE_ADSBLOL_ODBL_ATTRIBUTION=REQUIRED
STANDARD_RETRY_AFTER_COOLDOWN=IMPLEMENTED
AGGRESSIVE_4XX_RETRY=PROHIBITED
```

The ADSB.lol HTTP client already forwards response headers to the provider
response controller. The controller parses standard `Retry-After` values and
applies provider cooldowns. The ADSB.lol client itself performs one HTTP request
per call and does not contain an aggressive retry loop.

## 5. Budget policy

ADSB.lol publishes dynamic load-based rate limiting rather than a stable hard
quota. Global Flight Analytics therefore applies:

```text
provider=adsb.lol
budget_mode=fixed-window
max_requests=1
window=minute
provenance=PROJECT-CONSERVATIVE
```

This is a project safety cap, not a claim about an upstream hard quota. The
planned production scheduler remains substantially slower at approximately one
bounded ingestion request every 30 minutes.

## 6. Fail-closed production activation

Production GitHub Actions refuses to execute unless:

```text
ADSBLOL_PRODUCTION_CONTACT_CONFIRMED=true
```

The operator response now provides the external evidence needed for that gate,
but the repository does not treat receipt of the email as permission to bypass
compliance hardening or runtime availability checks.

Airplanes.live additionally requires `AIRPLANES_LIVE_ACCESS_APPROVED=true`.
OpenSky additionally requires `OPENSKY_OPERATIONAL_AGREEMENT_CONFIRMED=true`.

## 7. Render boundary

The Render web service runs `/app/server`, not the ingestion command. Stale
`TRAFFIC_PROVIDER` and `AIRPLANES_LIVE_TIMEOUT` values are therefore removed from
`render.yaml`; provider configuration belongs to the ingestion runtime.

## 8. Production activation criteria

Provider recovery is not closed until:

```text
ADSBLOL_PRODUCTION_RESPONSE=RECEIVED
ADSBLOL_COMPLIANCE_HARDENING=MERGED
ADSBLOL_PRODUCTION_CONTACT_CONFIRMED=true
PRODUCTION_TRAFFIC_INGESTION_WORKFLOW=ENABLED_FOR_CONTROLLED_RUN
PRODUCTION_ADSBLOL_SMOKE=PASS
PRODUCTION_TRAFFIC_FRESHNESS=PASS
SUBSEQUENT_SCHEDULED_RUN=PASS
GRAFANA_FRESHNESS_RECOVERY=PASS
PRODUCTION_PROVIDER_RECOVERY=CLOSED
```

Until live runtime validation is possible:

```text
ADSBLOL_ADAPTER=IMPLEMENTED
ADSBLOL_PROVIDER_POLICY=IMPLEMENTED
ADSBLOL_PRODUCTION_RESPONSE=RECEIVED
PRODUCTION_WORKFLOW_SOURCE_READY=YES
PRODUCTION_WORKFLOW_CONTACT_GATE=ACTIVE
PRODUCTION_INGESTION=INTENTIONALLY_OFFLINE
PRODUCTION_PROVIDER_RECOVERY=OPEN_RUNTIME_VALIDATION
```

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
This remains deliberate even after the operator response: the reply explicitly
warns that API limits, authentication, endpoints and other details may change.
A cross-provider compatibility test can assert identical canonical semantics for
the currently overlapping v2 fields while preserving distinct source identity,
but the transport contracts should not be merged merely because their current
payload shapes overlap.
