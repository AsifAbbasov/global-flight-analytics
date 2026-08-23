# Document 197 — External Provider and Deployment Correspondence Evidence

Status: CURRENT EVIDENCE REGISTER
Date: 2026-08-23
Scope: external production-data-provider correspondence and selected deployment notifications relevant to release truth

## Purpose

This document records the external communications that materially changed Global Flight
Analytics production-provider or deployment decisions. It is an evidence register, not a
mail archive.

The repository stores only the minimum facts needed to explain engineering decisions:

- date;
- recipient/provider;
- subject;
- request or notification summary;
- response summary;
- resulting project decision.

Full private message bodies, Gmail identifiers, authentication data, private deployment
URLs and unrelated personal correspondence are intentionally not committed.

## Evidence boundary

The authoritative raw correspondence remains in the project owner's mail account. This
repository record is a faithful engineering summary of those messages.

A provider reply is never promoted into a stronger claim than the text supports. In
particular, general usage guidance is not rewritten as a guaranteed quota, SLA, dedicated
support commitment, API-stability promise or explicit approval of every requested polling
interval.

---

## 1. ADSB.lol — initial production-use inquiry

Date: 2026-08-18
Recipient: `info@adsb.lol`
Subject: `Production API usage notice — Global Flight Analytics`

Global Flight Analytics described itself as a non-commercial open-data aviation
analytics/research project and asked whether the public ADSB.lol API could be used as the
live aircraft source for production.

The initial proposal described:

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

The message also stated that the backend had conservative request-budget,
timeout/backoff and provider-cooldown controls and would not retry aggressively.

The project asked whether the pattern was acceptable and whether ADSB.lol preferred a
different cadence, specific User-Agent/contact identifier, API key, feeding requirement or
other production condition.

### Delivery follow-up

The same inquiry was forwarded/resubmitted to `info@adsb.lol` later on 2026-08-18. This
was a duplicate delivery of the same proposal, not a distinct provider-policy request.

---

## 2. ADSB.lol — first operator response

Date: 2026-08-18
Responder: Katia / ADSB.lol

The operator supplied general application guidance:

- visible ODbL attribution is required;
- identifiable User-Agent and contact information are required;
- HTTP errors and rate limits must be respected;
- a 4xx response should cause correction or slower requests rather than more aggressive
  retries;
- `Retry-After` must be respected when provided;
- caching and deduplication should be used where practical;
- geographic requests should be appropriately scoped;
- applications should not circumvent limits through multiple IP addresses;
- integrations should be designed for future API/authentication/limit changes;
- feeding the network is encouraged;
- guaranteed-limit API keys were described as future work, with GitHub Sponsors expected to
  be the first eligible group;
- ADSB.lol is volunteer-run and provides no guarantee of availability, API stability, data
  accuracy, completeness or freshness;
- bespoke infrastructure, SLAs and dedicated support were not available.

The reply did not explicitly approve the initially proposed one-request-per-10-seconds
cadence and did not provide a guaranteed quota.

---

## 3. ADSB.lol — project acknowledgement and compliance commitment

Date: 2026-08-18
Recipient: reply to Katia / ADSB.lol
Subject: `Re: Production API usage notice — Global Flight Analytics`

Global Flight Analytics acknowledged the guidance and committed to implement:

```text
VISIBLE_ODBL_ATTRIBUTION=REQUIRED
IDENTIFIABLE_USER_AGENT_WITH_CONTACT=REQUIRED
RATE_LIMIT_AND_RETRY_AFTER_HANDLING=REQUIRED
CACHE_AND_DEDUPLICATION=REQUIRED_WHERE_PRACTICAL
GEOGRAPHIC_SCOPE=BOUNDED
FUTURE_AUTHENTICATION_CHANGE=SUPPORTED_BY_DESIGN
```

This follow-up is an implementation commitment by the project, not additional provider
approval.

---

## 4. Airplanes.live — REST API access request

Date: 2026-08-20
Recipient: `contact@airplanes.live`
Subject: `REST API access request — Global Flight Analytics`

The project explained that it was a non-commercial aviation analytics/research project,
linked the repository, described the Azerbaijan/Caucasus/Turkey focus, and stated that it
did not resell Airplanes.live data or operate as a commercial flight-tracking service.

The request documented the production failure:

```text
provider=airplanes.live
endpoint=/v2/point/40.4093/49.8671/250
observed_status=HTTP 403
```

The project asked whether REST API access could be enabled and whether an API credential,
registration, IP allowlisting or other configuration was required.

---

## 5. Airplanes.live — provider response

Date: 2026-08-20
Responder: J / Airplanes.live

The reply explained that the general free API had been taken down because of bot/app
traffic and hosting-cost pressure.

The provider stated that:

- feeders can access the API from the same IP address as the feeder;
- users are encouraged to contribute a feeder;
- applications, businesses, websites and other higher-volume users should use sponsorship;
- sponsorship was required for the described non-feeder application path;
- the quoted plans at the time were USD 25/month and USD 50/month for developer
  sponsorship.

No free external API exception, credential, registration or project-specific allowlisting
was granted to Global Flight Analytics.

Engineering consequence:

```text
AIRPLANES_LIVE_FREE_EXTERNAL_API=UNAVAILABLE
AIRPLANES_LIVE_FEEDER_ACCESS=SAME_IP_ONLY
AIRPLANES_LIVE_PROJECT_EXCEPTION=NOT_GRANTED
AIRPLANES_LIVE_ZERO_BUDGET_PRODUCTION_PATH=INELIGIBLE
```

This correspondence converted the earlier `403` from an ambiguous provider failure into a
confirmed access-policy boundary.

---

## 6. ADSB.lol — revised bounded production/demo inquiry

Date: 2026-08-21
Recipient: `info@adsb.lol`
Subject: `Production API usage confirmation for open-source flight analytics project`

After the first ADSB.lol guidance and the Airplanes.live access-policy response, Global
Flight Analytics sent a stricter production/demo proposal:

```text
endpoint=/v2/point/{lat}/{lon}/{radius}
maximum_radius=250 NM
normal_production_schedule≈1 request / 30 minutes
application_safety_cap<=1 request / minute
usage=research / analytics visualization
ODbL_attribution=PRESERVED
operational_ATC_use=NO
```

The project explicitly asked whether this revised pattern was acceptable and whether
additional identification, attribution, API-key or feeder setup was desired.

The revised cadence supersedes the earlier 2026-08-18 one-request-per-10-seconds proposal
for the current zero-cost production plan.

---

## 7. ADSB.lol — response to revised inquiry

Date: 2026-08-21
Responder: Katia / ADSB.lol

ADSB.lol repeated the same general application guidance supplied on 2026-08-18:

- visible ODbL attribution;
- identifiable User-Agent/contact;
- correct 4xx/rate-limit handling;
- respect for `Retry-After`;
- caching/deduplication;
- geographically scoped queries;
- no multi-IP rate-limit circumvention;
- design for future API/authentication/limit changes;
- no service-availability, API-stability, accuracy, completeness or freshness guarantee;
- no bespoke SLA or dedicated support.

The response did not object to the project description or bounded regional-use request,
but it did **not** contain an explicit guaranteed quota or a sentence promising that the
exact 30-minute pattern will always be accepted.

The repository therefore records this as external contact evidence and provider guidance,
not as a contractual SLA or guaranteed production entitlement.

Current project policy derived from the correspondence:

```text
provider=adsb.lol
budget_mode=fixed-window
max_requests=1
window=minute
budget_provenance=PROJECT-CONSERVATIVE
planned_scheduler_cadence≈1 request / 30 minutes
```

---

## 8. Automated Vercel preview failure notification — not outbound correspondence

Date: 2026-08-23
Sender: Vercel automated notifications
Subject: `Failed preview deployment on team 'asifabbasov's projects'`
Environment: Preview
Branch: `feat/frontend-aviation-basemap-map-chrome-20260823`
Commit: `a2c941a0ea9cd7421254857cf8f490b7eca4bd19`

This was an automated platform notification; Global Flight Analytics did **not** send a
support request to Vercel in this thread.

The notification reported:

```text
VERCEL_PREVIEW_BUILD=FAILED
COMMAND=pnpm run build
EXIT_CODE=1
NEXT_COMPILE=SUCCESS
FAILURE_STAGE=TypeScript stage reached
```

The supplied email/log excerpt does not include the final TypeScript diagnostic, so this
document does not invent one.

This notification describes a stale feature-branch preview and is not current production
truth. Later verified frontend closure work superseded that branch, and GitHub recorded
Vercel `success` for the subsequently merged frontend revisions, including current release
governance work.

```text
HISTORICAL_STALE_PREVIEW_FAILURE=PRESERVED
CURRENT_MAIN_FAILURE_INFERRED_FROM_OLD_PREVIEW=PROHIBITED
```

---

## 9. Communications not present

A Gmail review of the August 2026 Global Flight Analytics correspondence found no outbound
OpenSky email and no outbound Vercel support email.

OpenSky therefore remains governed by its documented provider/access requirements and the
repository's explicit runtime agreement gate; no project-specific email approval is claimed.

Automated GitHub Actions, Dependabot, Vercel bot and Grafana alert emails are operational
notifications and are not counted as provider correspondence unless they materially support
a documented incident or deployment-evidence statement.

---

## 10. Canonical documentation mapping

The evidence in this register is interpreted by the following domain documents:

- `191_PRODUCTION_INGESTION_RESILIENCE_INCIDENT_CLOSURE.md` — Airplanes.live `403`, external
  access-policy confirmation and fail-closed incident containment;
- `193_PRODUCTION_TRAFFIC_PROVIDER_RECOVERY.md` — ADSB.lol correspondence, compliance
  hardening and controlled-activation criteria;
- `38_TRAFFIC_PROVIDER_AUTOMATIC_FALLBACK.md` — provider eligibility gates and zero-budget
  fallback policy;
- `169_RELEASE_TRUTH_AND_DEPLOYMENT_REVISION_CLOSURE.md` — historical deployment evidence
  versus mutable/current deployment truth;
- `196_FRONTEND_VISUAL_POLISH_V2_CLOSURE.md` — final frontend exact-head evidence and later
  Vercel success, which must not be contradicted by a stale preview failure.

## Current correspondence-derived state

```text
AIRPLANES_LIVE_PROVIDER_RESPONSE=RECEIVED
AIRPLANES_LIVE_FREE_EXTERNAL_API=UNAVAILABLE
AIRPLANES_LIVE_PROJECT_EXCEPTION=NOT_GRANTED
ADSBLOL_INITIAL_CONTACT=COMPLETE
ADSBLOL_REVISED_PRODUCTION_CONTACT=COMPLETE
ADSBLOL_PROVIDER_GUIDANCE=RECEIVED
ADSBLOL_EXTERNAL_CONTACT_EVIDENCE=SATISFIED
ADSBLOL_GUARANTEED_QUOTA=NOT_PROVIDED
ADSBLOL_SLA=NOT_PROVIDED
ADSBLOL_PRODUCTION_CONTACT_CONFIRMED_RUNTIME_GATE=STILL_EXPLICITLY_CONTROLLED
HISTORICAL_VERCEL_PREVIEW_FAILURE=PRESERVED_AS_NONCURRENT_EVIDENCE
PRODUCTION_PROVIDER_RECOVERY=OPEN_RUNTIME_VALIDATION
```
