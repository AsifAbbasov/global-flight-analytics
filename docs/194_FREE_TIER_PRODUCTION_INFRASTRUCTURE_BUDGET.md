# 194 — Free-Tier Production Infrastructure Budget

Status: **HARDENING IN PROGRESS**

## 1. Incident signal

On 2026-08-21 Neon reported that the `global-flight-analytics` project had consumed 100% of its 100 CU-hour monthly compute allowance. During the same production degradation window, the GitHub `Production Metrics Scrape` workflow repeatedly received timeouts and HTTP 502/503 responses from the Render API metrics endpoint, and Grafana subsequently reported missing external metrics.

The GitHub evidence proves that the metrics-forwarding failures happened before Grafana remote write: the protected production metrics source itself was unavailable. The available evidence does **not** prove that Neon exhaustion was the only cause of every Render 5xx response, so the root cause is recorded as an infrastructure wake-pattern and capacity incident rather than a single-provider attribution.

## 2. Free-tier incompatibility in the previous schedule

The previous production schedules were:

- production reconciliation every 10 minutes;
- external production metrics scrape every 15 minutes;
- Cloudflare ingestion primary dispatch every 10 minutes;
- Cloudflare watchdog freshness checks every 5 minutes when dispatch was enabled;
- Render free web service idle spin-down after approximately 15 minutes;
- Neon free compute scale-to-zero after inactivity.

The independent reconciliation job connected directly to PostgreSQL every 10 minutes. That cadence was shorter than the observed database sleep cycle and therefore could keep a low-traffic database effectively active for most of the month.

The external metrics workflow also contacted the Render API every 15 minutes, which was near the Render free-service idle threshold and therefore acted as an accidental keep-alive.

The Cloudflare Worker remains safe while `DISPATCH_ENABLED=false` because it returns before any GitHub or Render network call.

This is incompatible with the project requirement that the v1 public portfolio deployment operate on free infrastructure without artificial keep-alive traffic.

## 3. Free-tier operating policy

### 3.1 No keep-alive traffic

Production monitoring, reconciliation, watchdog, or health checks MUST NOT exist only to keep Render or Neon awake.

Cold starts and database resumes are accepted platform behavior for the free v1 deployment.

Any scheduled component that touches the Render API must be treated as a potential database wake because the production API establishes its PostgreSQL pool during process startup when `DATABASE_URL` is configured.

### 3.2 Metrics cadence

`Production Metrics Scrape` runs every two hours:

```text
20 */2 * * *
```

The `:20` minute is intentionally staggered after the ingestion primary (`:17`) and watchdog (`:19`) windows. This keeps production verification inside one short infrastructure wake cluster rather than launching independent wake cycles.

Grafana's explicit missing-metrics alert uses a 180-minute lookback so the alert contract matches the sparse collection cadence.

### 3.3 Reconciliation cadence

While production traffic ingestion remains fail-closed, `Production Reconciliation` has no schedule and is manual-only.

When traffic ingestion is restored, reconciliation SHOULD execute in the same database wake window immediately after a successful ingestion batch instead of using an independent cron schedule.

That design turns:

```text
independent ingestion wake
+
independent reconciliation wake
```

into:

```text
one ingestion/reconciliation wake window
```

### 3.4 Cloudflare ingestion reliability cadence

The source-controlled FREE_V1 profile is now:

```text
primary dispatch:  17 */2 * * *
watchdog:          19 */2 * * *
metrics:           20 */2 * * *
```

The primary therefore requests at most one scheduled ingestion window every two hours. The watchdog runs two minutes later and metrics one minute after that, deliberately forming one staggered two-hour wake cluster.

The earlier recovery target was:

```text
17,47 * * * *
```

Controlled September 2026 runtime validation showed that this 30-minute profile was still too aggressive for the measured FREE_V1 sleep behavior. Neon operations repeatedly showed compute starts followed by suspension roughly twenty to twenty-four minutes later, with one longer active window when successive scheduled activity overlapped. A schedule that can re-wake compute every 30 minutes leaves too little guaranteed idle time and can recreate monthly allowance exhaustion even though scale-to-zero itself works.

While recovery remains open:

```text
DISPATCH_ENABLED=false
```

must remain fail-closed. With the kill switch active, both Cloudflare Cron Triggers perform zero GitHub and zero Render network calls.

The GitHub production ingestion workflow remains dispatch/manual-owned rather than becoming a second independent scheduler. Any future scheduler-frequency increase requires a fresh compute-budget calculation before activation.

### 3.5 Production smoke

The daily production release smoke remains scheduled once per day. It intentionally wakes the deployed API to verify health, readiness, version, CORS, and frontend availability. One bounded daily verification window is accepted because it provides meaningful release evidence and contributes only a small fraction of the monthly compute budget.

## 4. Compute budget

Neon measures compute as:

```text
CU-hours = compute size in CU × active hours
```

At the minimum 0.25 CU, a database continuously active for 400 hours consumes 100 CU-hours.

### 4.1 Historical pre-reset model

The original recovery model assumed a five-minute database wake:

```text
48 ingestion wakes/day
× 5 minutes awake/wake
× 30 days
= 120 active hours/month

120 active hours
× 0.25 CU
= 30 CU-hours/month
```

That assumption was not retained after live post-reset observation because the real FREE_V1 active windows were materially longer.

### 4.2 Live-observed September wake behavior

During controlled scheduler validation on 2026-09-05, Neon operations recorded representative start/suspend pairs including:

```text
18:48:32Z start → 19:09:02Z suspend  ≈ 20.5 min
19:18:38Z start → 19:42:32Z suspend  ≈ 23.9 min
19:48:35Z start → 20:09:02Z suspend  ≈ 20.5 min
20:18Z start    → 21:09Z suspend     ≈ 50 min
21:18:26Z start → 21:38:47Z suspend  ≈ 20.4 min
```

The approximately fifty-minute window remained active because additional scheduled activity arrived before a clean sleep boundary. These observations prove that scale-to-zero works, but they also invalidate a five-minute wake assumption for capacity planning.

### 4.3 Why the 30-minute primary is rejected for FREE_V1

A deliberately conservative stress calculation using a twenty-minute wake for every 30-minute primary is:

```text
48 ingestion wakes/day
× 20 minutes awake/wake
× 30 days
= 480 active hours/month

480 active hours
× 0.25 CU
= 120 CU-hours/month
```

This is not a Neon forecast and does not claim every wake would be independent. It is a safety-bound calculation using measured wake duration. Because it exceeds the 100 CU-hour allowance before interactive usage, smoke, retries, or longer windows are added, the 30-minute profile is not acceptable for FREE_V1.

### 4.4 Two-hour primary budget

Using the same conservative twenty-minute wake assumption for one two-hour primary cluster:

```text
12 wake clusters/day
× 20 minutes awake/cluster
× 30 days
= 120 active hours/month

120 active hours
× 0.25 CU
= 30 CU-hours/month
```

The watchdog and metrics are intentionally placed at `:19` and `:20` so they can reuse the same wake cluster rather than create independent hourly activity. Real usage can be higher because of cold starts, user traffic, retries, longer queries, or autoscaling, so the project keeps an explicit safety reserve.

The project therefore retains:

```text
TARGET_MONTHLY_NEON_COMPUTE <= 60 CU-hours
RESERVE_FOR_INTERACTIVE_AND_RECOVERY_WORK >= 40 CU-hours
```

## 5. Production activation budget

After traffic ingestion is restored, FREE_V1 remains intentionally low-frequency and analytical rather than real-time.

The intended scheduler ownership is:

```text
Cloudflare primary scheduler (:17 every 2 hours)
        ↓
GitHub production ingestion workflow
        ↓
ingestion write batch
        ↓
reconciliation in the same wake window
        ↓
watchdog at :19 / metrics at :20 reuse the wake window
        ↓
idle → Render/Neon sleep
```

There must not be a second independent high-frequency ingestion scheduler once this profile is activated.

Any future increase in ingestion frequency requires a new CU-hour budget calculation and controlled runtime validation before activation.

## 6. Deployment profiles

The resource policy is a deployment concern, not a downgrade of domain architecture.

### FREE_V1

```text
cold starts accepted
2-hour ingestion target
2-hour metrics/watchdog
one daily release smoke
scale-to-zero enabled
no keep-alive traffic
```

### SCALE / PAID

```text
same API contracts
same domain model
same analytical modules
same provider abstractions
same database schema
same observability model
higher ingestion cadence / always-on capacity when justified
```

Moving from `FREE_V1` to a paid capacity profile must therefore be an operational configuration change, not a rewrite of the product architecture.

## 7. Required Neon settings

For the free v1 deployment:

- scale-to-zero/autosuspend must remain enabled;
- minimum compute should remain at the smallest available size (target 0.25 CU);
- maximum autoscaling should remain conservative unless measured production demand requires more;
- no external uptime service may ping the database or Render API merely to prevent sleep.

These console settings are operational prerequisites and are not encoded in this repository unless Neon infrastructure-as-code is introduced later.

## 8. Live Neon evidence

### 8.1 August incident evidence

A live Neon account inspection after the August incident recorded:

```text
project                         global-flight-analytics
primary branch                  production
PostgreSQL                      18
region                          aws-eu-central-1
compute range                   0.25–2 CU
active time                     ~373.6 hours
compute usage                   ~101.2 CU-hours
average effective compute       ~0.27 CU
quota reset                     2026-09-01 00:00 UTC
endpoint state after inactivity idle / suspended
```

The average effective compute of approximately 0.27 CU was close to the 0.25 CU minimum, strongly supporting the conclusion that excessive active time and repeated wake windows — rather than sustained high autoscaling — dominated monthly compute exhaustion.

### 8.2 September post-reset runtime evidence

After the monthly reset, controlled production recovery established all of the following without claiming full-month safety:

- Render served the exact canonical repository revision `e7d4445778b75cfbfa7812857154640b48659c69`;
- Production Smoke run `33969466911` succeeded;
- controlled manual production ingestion run `33971235793` succeeded and stored ADSB.lol traffic;
- Cloudflare-origin primary dispatch run `33985176640` succeeded with `PRODUCTION_INGESTION_DISPATCH_SOURCE=cloudflare-primary`;
- subsequent automatic Cloudflare runs also succeeded while dispatch was temporarily enabled;
- Neon repeatedly returned to suspended state between windows, proving scale-to-zero works;
- the measured wake duration was materially longer than the original five-minute planning assumption;
- production dispatch was returned to `false` after the controlled validation window.

This evidence proves pipeline viability and scale-to-zero capability. It does **not** yet prove that the revised two-hour cadence stays within the monthly budget, so final recovery remains open pending controlled deployment and observation of the revised profile.

## 9. Current recovery state

```text
NEON_MONTHLY_COMPUTE_ALLOWANCE=AVAILABLE_AFTER_2026-09-01_RESET
NEON_SCALE_TO_ZERO=OBSERVED_WORKING
NEON_MINIMUM_COMPUTE=0.25_CU
PRODUCTION_RECONCILIATION_CRON=REMOVED
PRODUCTION_METRICS_SCRAPE_CADENCE=2_HOURS
PRODUCTION_METRICS_SCRAPE_MINUTE=20
CLOUDFLARE_PRIMARY_TARGET_CADENCE=2_HOURS
CLOUDFLARE_PRIMARY_MINUTE=17
CLOUDFLARE_WATCHDOG_TARGET_CADENCE=2_HOURS
CLOUDFLARE_WATCHDOG_MINUTE=19
CLOUDFLARE_DISPATCH_KILL_SWITCH=ACTIVE
PRODUCTION_WAKE_CLUSTER=STAGGERED
GRAFANA_METRICS_MISSING_WINDOW=180_MINUTES
RENDER_KEEP_ALIVE_POLICY=PROHIBITED
CONTROLLED_CLOUDFLARE_PRIMARY_RUNTIME=PASS
POST_RESET_SCALE_TO_ZERO=PASS
FREE_TIER_INFRASTRUCTURE_RECOVERY=IN_PROGRESS
```

The finding remains open until the revised two-hour source-controlled profile is merged, deployed from an exact merged revision, one controlled scheduled run succeeds under that profile, and Neon is again observed scaling to zero afterward.

---

## Canonical remediation record — GFA-OPS-456

### 1. Finding / symptom
The FREE_V1 production automation schedule was incompatible with the intended scale-to-zero cost model: independent production wakes consumed the monthly Neon compute allowance and undermined free-tier availability.

### 2. Root cause
Production cadence was originally designed per subsystem rather than from one shared infrastructure wake budget. Reconciliation, metrics, ingestion and watchdog activity could wake Render/Neon more frequently than the infrastructure could reliably return to sleep.

### 3. Failure scenario
Low user traffic would otherwise allow Render and Neon to sleep, but scheduled maintenance/monitoring work repeatedly wakes them. Neon remains active for a large fraction of the month at approximately minimum compute, eventually exhausting the 100 CU-hour allowance. The database/API then becomes unavailable or degraded before downstream observability can complete.

### 4. Impact
The zero-cost production environment exhausted its monthly database compute budget, production metrics source requests degraded, and runtime recovery became blocked until allowance reset/equivalent free recovery.

### 5. Severity rationale
**P1 retrospective.** This was a real production resource-exhaustion incident that consumed the full monthly free database compute allowance and materially degraded the production runtime. The severity is reconstructed from observed impact; the source incident did not record an original severity label.

### 6. Existing guarantees violated
- FREE_V1 must operate without artificial keep-alive traffic;
- scheduled components must share an explicit monthly compute/wake budget;
- monitoring and reconciliation must not prevent platform scale-to-zero by cadence alone;
- production recovery must reserve compute for interactive and incident work;
- cadence increases require resource-budget review before activation.

### 7. Considered solutions
- move to paid always-on infrastructure;
- keep existing cadences and accept quota exhaustion;
- disable all monitoring/reconciliation permanently;
- reduce and stagger cadences, remove independent reconciliation cron while ingestion is offline, share reconciliation with the ingestion wake window, and define an explicit CU-hour reserve.

### 8. Chosen remediation
Move metrics to `20 */2 * * *`, make reconciliation manual-only while ingestion is offline, run the Cloudflare primary at `17 */2 * * *`, watchdog at `19 */2 * * *`, cluster wakes, prohibit keep-alive traffic, keep scale-to-zero enabled, and retain `TARGET_MONTHLY_NEON_COMPUTE <= 60 CU-hours` with at least 40 CU-hours reserved for interactive/recovery demand.

### 9. Why this solution was selected
It preserves the same product/domain architecture on free infrastructure, accepts cold starts as a deliberate FREE_V1 trade-off, and addresses the observed active-time mechanism instead of attributing every Render error to Neon without evidence.

### 10. Rejected alternatives
A paid tier violates the current zero-cost constraint; retaining former cadences repeats the wake pattern; disabling all production verification removes meaningful evidence; increasing timeouts does not reduce compute-active time.

### 11. Trade-offs
Traffic and monitoring evidence become less frequent and cold starts are expected. The configuration is appropriate for a non-real-time portfolio MVP, not a safety-critical or high-frequency flight-tracking SLA. Future higher cadence may require a paid capacity profile.

### 12. Regression tests / protection
Repository hardening protects the two-hour primary and watchdog cadence, removes the independent reconciliation cron while ingestion is offline, keeps the metrics cadence aligned to the same two-hour wake cluster, preserves the Cloudflare kill switch, rejects the superseded 30-minute primary in the source verifier, and documents a mandatory compute-budget review before any frequency increase.

### 13. Adversarial review findings
The incident evidence does not prove Neon exhaustion caused every Render 502/503/429. The September controlled window also does not prove the revised profile's full-month consumption. The finding is therefore scoped to the internal wake-pattern/budget incompatibility and remains open until the revised cadence receives runtime validation.

### 14. Remediation iterations
1. Neon reported 100% monthly compute allowance consumption.
2. Production schedule inventory identified independent high-frequency wake patterns.
3. FREE_V1 was redesigned around staggered wake windows and an explicit CU-hour budget.
4. PR #85 applied the first scheduling hardening and kept dispatch fail-closed.
5. The September quota reset allowed controlled manual and Cloudflare-origin runtime recovery.
6. Controlled automatic primary dispatch proved the full Cloudflare → GitHub → ADSB.lol → Neon → API path.
7. Neon operations showed real wake windows around twenty minutes, invalidating the five-minute planning assumption.
8. Production dispatch was returned to fail-closed and the FREE_V1 primary target was reduced from 30 minutes to two hours pending source-controlled merge/deploy validation.

### 15. Residual risks and limitations
Autoscaling, retries, user traffic, cold-start duration, metrics behavior and future provider cadence can raise compute above the model. Neon/Render console settings are operational prerequisites rather than repository-owned infrastructure-as-code. The 60-CU target is a project safety budget, not an upstream guarantee.

### 16. Operational or deployment consequences
Production ingestion remains fail-closed while this cadence correction is under review. Reconciliation is manual-only. Metrics remain every two hours. After merge, the exact merged Worker revision must be deployed and one bounded scheduled validation must demonstrate successful ingestion followed by Neon scale-to-zero before recovery can close.

### 17. Exact evidence
- PR #85 head `1a82b8eae63ff5e293830c630fed6a9102eb9480`, merge `4a95b7a0caae8e8581cf132945c2b1be3a7a3cca`;
- historical live Neon evidence: ~`373.6` active hours, ~`101.2` CU-hours, average effective compute ~`0.27` CU;
- monthly reset: `2026-09-01T00:00:00Z`;
- canonical production revision: `e7d4445778b75cfbfa7812857154640b48659c69`;
- Production Smoke run `33969466911` SUCCESS;
- manual production ingestion run `33971235793` SUCCESS;
- Cloudflare-origin production ingestion run `33985176640` SUCCESS with `cloudflare-primary` provenance;
- Neon post-reset scale-to-zero repeatedly observed through start/suspend operation pairs;
- repository correction target: primary `17 */2 * * *`, watchdog `19 */2 * * *`, metrics `20 */2 * * *`, dispatch fail-closed during review.

### 18. Final canonical status
**IN_PROGRESS.** Pipeline recovery and post-reset scale-to-zero are proven, but the newly measured wake duration requires a safer two-hour primary profile. Closure requires exact-merge deployment plus one controlled scheduled run and subsequent scale-to-zero evidence under that profile.

### 19. Prevention / future guard
Treat every scheduled production action that touches Render or Neon as a budgeted wake. Require a written monthly CU-hour calculation and wake-cluster review before increasing ingestion, watchdog, reconciliation, smoke or metrics frequency; keep FREE_V1 and SCALE/PAID as explicit deployment profiles rather than silently changing product architecture.
