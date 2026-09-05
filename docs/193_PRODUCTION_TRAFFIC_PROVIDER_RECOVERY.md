# Production Traffic Provider Recovery

Status: Controlled runtime path proven; final recovery reconciliation remains open while the revised FREE_V1 cadence is validated
Date: 2026-08-23
Last updated: 2026-09-06
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

## 3. Operator correspondence

### 3.1 Initial production-use inquiry — 2026-08-18

The project first contacted `info@adsb.lol` with the subject
`Production API usage notice — Global Flight Analytics`.

The message described Global Flight Analytics as a non-commercial open-data
aviation analytics/research project and proposed:

```text
endpoint=/v2/point/{lat}/{lon}/{radius}
initial_region=Baku / Azerbaijan
maximum_radius=250 NM
collector=one centralized backend collector
initial_proposed_cadence≈1 request / 10 seconds
safety_critical_use=NO
provider_attribution=PRESERVED
ODbL_terms=PRESERVED
```

The message also stated that the backend had its own request-budget,
timeout/backoff and provider-cooldown controls and would not retry aggressively.
It asked whether the usage was acceptable and whether ADSB.lol preferred a
different cadence, identifiable User-Agent/contact, API key, feeding requirement
or other production condition.

The same inquiry was forwarded/resubmitted later on 2026-08-18. That was a
duplicate delivery of the same request, not a separate production-policy proposal.

### 3.2 ADSB.lol response — 2026-08-18

Katia / ADSB.lol replied with general application requirements:

- visible ODbL attribution in the application;
- an identifiable User-Agent with contact information;
- respect for HTTP errors and rate limits;
- a 4xx response should cause correction or slower requests rather than more
  aggressive retries;
- respect for `Retry-After` when provided;
- caching and deduplication where practical;
- appropriately scoped geographic queries;
- no rate-limit circumvention through multiple IP addresses;
- design for future API/authentication/limit changes;
- feeding the network is encouraged;
- guaranteed-limit API keys were described as future work;
- no guarantee of service availability, API stability, data accuracy,
  completeness or freshness;
- no bespoke infrastructure, SLA or dedicated support.

The response did not explicitly approve the initially proposed one-request-per-
10-seconds cadence and did not provide a guaranteed quota.

### 3.3 Project acknowledgement — 2026-08-18

The project replied to Katia confirming that Global Flight Analytics would adopt
visible ODbL attribution, an identifiable User-Agent/contact, proper rate-limit
and `Retry-After` handling, caching/deduplication, geographically scoped requests
and an integration design able to accommodate future authentication/API-key
changes.

This acknowledgement is a project implementation commitment, not additional
provider approval.

### 3.4 Revised bounded production/demo inquiry — 2026-08-21

After the initial ADSB.lol guidance and the externally confirmed Airplanes.live
access-policy boundary, the project sent a stricter request to `info@adsb.lol`
with the subject
`Production API usage confirmation for open-source flight analytics project`.

The revised request described:

```text
endpoint=/v2/point/{lat}/{lon}/{radius}
maximum_radius=250 NM
normal_production_schedule≈1 request / 30 minutes
application_safety_cap<=1 request / minute
usage=research / analytics visualization
ODbL_attribution=PRESERVED
operational_ATC_use=NO
```

The project explicitly asked whether this revised pattern was acceptable and
whether additional identification, attribution, API-key or feeder setup was
required.

The 30-minute production cadence superseded the earlier 2026-08-18
one-request-per-10-seconds proposal at the time of that correspondence. It is
preserved here as historical inquiry evidence, not as the current FREE_V1
scheduler policy. Subsequent Neon runtime measurements caused the project to
reduce its own production cadence further to one request every two hours.

### 3.5 ADSB.lol response to revised inquiry — 2026-08-21

ADSB.lol replied the same day and repeated the same general application guidance:
visible ODbL attribution, identifiable User-Agent/contact, correct 4xx/rate-limit
handling, `Retry-After`, practical caching/deduplication, bounded geographic
queries, no multi-IP circumvention, design for future API/authentication changes,
and no availability/stability/accuracy/completeness/freshness guarantee or bespoke
SLA/support.

The response did not object to the project description or bounded regional-use
request, but it did not provide a guaranteed quota or a sentence promising that
the exact 30-minute pattern will always be accepted.

The repository therefore records the exchange as provider contact evidence and
usage guidance rather than a contractual service commitment.

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
ADSBLOL_COMPLIANCE_HARDENING=MERGED
```

The ADSB.lol HTTP client forwards response headers to the provider response
controller. The controller parses standard `Retry-After` values and applies
provider cooldowns. The ADSB.lol client itself performs one HTTP request per call
and does not contain an aggressive retry loop.

PR #100 carried the provider-guidance compliance increment. Its exact pull-request
head `489e398c1d371c7613848fa9b39798edeef9806d` completed Frontend CI, Backend CI,
CodeQL, API Load Baseline and Playwright E2E successfully before squash merge.
The resulting `main` commit is `47769076376d7e4596610059cdb2505c23598237`.
This evidence closes the repository-side compliance hardening only; it is not
runtime provider-recovery evidence.

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
current source-controlled FREE_V1 scheduler is substantially slower at one
bounded ingestion request every two hours. The reduction from the earlier
30-minute recovery target was driven by measured Neon wake duration and free-tier
compute safety, not by a new upstream ADSB.lol restriction.

## 6. Fail-closed production activation

Production GitHub Actions refuses to execute unless:

```text
ADSBLOL_PRODUCTION_CONTACT_CONFIRMED=true
```

The operator responses provide the external evidence needed to satisfy that gate,
but the repository does not treat receipt of email as permission to bypass
compliance hardening or runtime availability checks. The gate was explicitly
enabled for the controlled production runs and those runs succeeded.

Airplanes.live additionally requires `AIRPLANES_LIVE_ACCESS_APPROVED=true`.
OpenSky additionally requires `OPENSKY_OPERATIONAL_AGREEMENT_CONFIRMED=true`.

## 7. Render boundary

The Render web service runs `/app/server`, not the ingestion command. Stale
`TRAFFIC_PROVIDER` and `AIRPLANES_LIVE_TIMEOUT` values are therefore removed from
`render.yaml`; provider configuration belongs to the ingestion runtime.

## 8. Production activation criteria and post-reset evidence

The provider recovery criteria remain:

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

The September 2026 recovery window established the provider/runtime portion with
real production evidence:

```text
CANONICAL_PRODUCTION_REVISION=e7d4445778b75cfbfa7812857154640b48659c69
PRODUCTION_SMOKE_RUN_ID=33969466911
PRODUCTION_SMOKE=PASS
MANUAL_INGESTION_RUN_ID=33971235793
MANUAL_INGESTION=PASS
MANUAL_INGESTION_PROVIDER=adsb.lol
MANUAL_INGESTION_STORED=5
MANUAL_TRAFFIC_FRESHNESS=PASS
CLOUDFLARE_PRIMARY_RUN_ID=33985176640
CLOUDFLARE_PRIMARY_DISPATCH_SOURCE=cloudflare-primary
CLOUDFLARE_PRIMARY=PASS
CLOUDFLARE_PRIMARY_PROVIDER=adsb.lol
CLOUDFLARE_PRIMARY_STORED=5
CLOUDFLARE_PRIMARY_TRAFFIC_FRESHNESS=PASS
SUBSEQUENT_AUTOMATIC_PRIMARY_RUNS=PASS
NEON_SCALE_TO_ZERO=OBSERVED_WORKING
```

Run `33985176640` executed on the exact canonical revision and recorded
`PRODUCTION_INGESTION_DISPATCH_SOURCE=cloudflare-primary`. ADSB.lol returned five
usable observations, all five were stored, five trajectories were produced, and
the production freshness verifier passed with data within the configured
30-minute freshness boundary.

The controlled window also proved that the provider path was not the remaining
FREE_V1 blocker. Neon wake-duration measurements showed the then-configured
30-minute primary was too aggressive for the 100 CU-hour monthly allowance. The
Cloudflare kill switch was therefore returned to `false`, and the primary target
was reduced to `17 */2 * * *` before permanent activation.

Current status is therefore:

```text
ADSBLOL_ADAPTER=IMPLEMENTED
ADSBLOL_PROVIDER_POLICY=IMPLEMENTED
ADSBLOL_PRODUCTION_RESPONSE=RECEIVED
ADSBLOL_COMPLIANCE_HARDENING=MERGED
PRODUCTION_WORKFLOW_SOURCE_READY=YES
PRODUCTION_WORKFLOW_CONTACT_GATE=PASS
CONTROLLED_PRODUCTION_ADSBLOL_SMOKE=PASS
CONTROLLED_PRODUCTION_TRAFFIC_FRESHNESS=PASS
CONTROLLED_CLOUDFLARE_PRIMARY=PASS
SUBSEQUENT_SCHEDULED_RUN=PASS
PRODUCTION_INGESTION=INTENTIONALLY_FAIL_CLOSED
PRODUCTION_PROVIDER_RECOVERY=OPEN_FINAL_RECONCILIATION
```

Provider recovery is not marked `CLOSED` in this document yet because the final
FREE_V1 production state must use the revised two-hour scheduler, be deployed from
an exact merged revision, and complete the remaining observability/final-recovery
reconciliation without weakening the free-tier budget boundary.

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
This remains deliberate even after the operator responses: the replies explicitly
warn that API limits, authentication, endpoints and other details may change.
A cross-provider compatibility test can assert identical canonical semantics for
the currently overlapping v2 fields while preserving distinct source identity,
but the transport contracts should not be merged merely because their current
payload shapes overlap.
