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
- Neon free compute scale-to-zero after approximately 5 minutes of database inactivity.

The independent reconciliation job connected directly to PostgreSQL every 10 minutes. That cadence is shorter than two Neon autosuspend windows and therefore can keep a low-traffic database effectively active for most of the month.

The external metrics workflow also contacted the Render API every 15 minutes, which is approximately the Render free-service idle threshold and therefore acts as an accidental keep-alive.

The Cloudflare Worker is currently safe because `DISPATCH_ENABLED=false` returns before any GitHub or Render network call. However, its previous enabled profile would have checked the Render traffic endpoint every five minutes and could therefore recreate the same keep-awake pattern after provider recovery.

This is incompatible with the project requirement that the v1 public portfolio deployment operate on free infrastructure without artificial keep-alive traffic.

## 3. Free-tier operating policy

### 3.1 No keep-alive traffic

Production monitoring, reconciliation, watchdog, or health checks MUST NOT exist only to keep Render or Neon awake.

Cold starts and database resumes are accepted platform behavior for the free v1 deployment.

Any scheduled component that touches the Render API must be treated as a database wake because the production API establishes its PostgreSQL pool during process startup when `DATABASE_URL` is configured.

### 3.2 Metrics cadence

`Production Metrics Scrape` runs every two hours:

```text
20 */2 * * *
```

The `:20` minute is intentionally staggered after the future ingestion primary (`:17`) and watchdog (`:19`) windows. This avoids launching multiple cold-start requests simultaneously while still keeping all three operations inside one short infrastructure wake cluster.

Grafana's explicit missing-metrics alert is rendered with a 180-minute lookback so the alert contract matches the sparse collection cadence.

### 3.3 Reconciliation cadence

While production traffic ingestion is intentionally offline, `Production Reconciliation` has no schedule and is manual-only.

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

The future enabled free-tier profile is:

```text
primary dispatch:  17,47 * * * *
watchdog:          19 */2 * * *
metrics:           20 */2 * * *
```

The primary therefore requests at most two scheduled ingestion windows per hour. Every two hours the watchdog runs two minutes after the `:17` primary and metrics runs one minute after the watchdog. This deliberately forms one staggered wake cluster instead of three independent infrastructure wake events.

While recovery remains open:

```text
DISPATCH_ENABLED=false
```

must remain fail-closed. With the kill switch active, both Cloudflare Cron Triggers must perform zero GitHub and zero Render network calls.

The GitHub production ingestion workflow should remain dispatch/manual-owned rather than becoming a second independent high-frequency scheduler. A future scheduler-frequency increase requires a fresh compute-budget calculation before activation.

### 3.5 Production smoke

The daily production release smoke remains scheduled once per day. It intentionally wakes the deployed API to verify health, readiness, version, CORS, and frontend availability. One bounded daily verification window is accepted because it provides meaningful release evidence and contributes only a small fraction of the monthly compute budget.

## 4. Compute budget

Neon measures compute as:

```text
CU-hours = compute size in CU × active hours
```

At the minimum 0.25 CU, a database continuously active for 400 hours consumes 100 CU-hours.

For the free-tier observability cadence, a conservative wake-window estimate is:

```text
12 metrics wakes/day
× 5 minutes awake/wake
× 30 days
= 30 active hours/month

30 active hours
× 0.25 CU
= 7.5 CU-hours/month
```

For a future 30-minute ingestion cadence, a deliberately conservative upper-bound estimate that assumes every ingestion creates a separate five-minute minimum wake window is:

```text
48 ingestion wakes/day
× 5 minutes awake/wake
× 30 days
= 120 active hours/month

120 active hours
× 0.25 CU
= 30 CU-hours/month
```

Actual usage can be lower when metrics, watchdog, reconciliation, and user traffic reuse an already-awake window. It can also be higher because of autoscaling, retries, long-running queries, or user traffic.

The project therefore uses a safety target rather than trying to consume the full upstream allowance:

```text
TARGET_MONTHLY_NEON_COMPUTE <= 60 CU-hours
RESERVE_FOR_INTERACTIVE_AND_RECOVERY_WORK >= 40 CU-hours
```

## 5. Production activation budget

After traffic ingestion is restored, the target v1 cadence is intentionally low-frequency. A 30-minute ingestion cadence combined with reconciliation in the same wake window remains compatible with the analytical, non-real-time v1 product scope and leaves substantially more room than the previous independent 10-minute reconciliation loop.

The intended scheduler ownership is:

```text
Cloudflare primary scheduler (:17 / :47)
        ↓
GitHub production ingestion workflow
        ↓
ingestion write batch
        ↓
reconciliation in the same wake window
        ↓
watchdog / metrics reuse the existing wake window where applicable
        ↓
idle → Render/Neon sleep
```

There must not be a second independent high-frequency ingestion scheduler once this profile is activated.

Any future increase in ingestion frequency requires a new CU-hour budget calculation before activation.

## 6. Deployment profiles

The resource policy is a deployment concern, not a downgrade of domain architecture.

### FREE_V1

```text
cold starts accepted
30-minute ingestion target
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

## 8. Live Neon evidence — 2026-08-21

A live Neon account inspection after the incident recorded:

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

The endpoint was observed suspended roughly six minutes after its last recorded activity, confirming that scale-to-zero is functioning in practice. The average effective compute of approximately 0.27 CU is very close to the 0.25 CU minimum, which strongly supports the conclusion that excessive active time and repeated wake windows — rather than sustained high autoscaling — dominated the monthly compute exhaustion.

This evidence does not prove that every Render 502/503/429 response was caused by Neon. It narrows the database-side incident mechanism to active-time budget exhaustion and validates the free-tier scheduling changes as the primary corrective action.

## 9. Current recovery state

```text
NEON_MONTHLY_COMPUTE_ALLOWANCE=EXHAUSTED
NEON_SCALE_TO_ZERO=OBSERVED_WORKING
NEON_AVERAGE_EFFECTIVE_COMPUTE≈0.27_CU
NEON_QUOTA_RESET=2026-09-01T00:00:00Z
PRODUCTION_RECONCILIATION_CRON=REMOVED
PRODUCTION_METRICS_SCRAPE_CADENCE=2_HOURS
PRODUCTION_METRICS_SCRAPE_MINUTE=20
CLOUDFLARE_PRIMARY_TARGET_CADENCE=30_MINUTES
CLOUDFLARE_WATCHDOG_TARGET_CADENCE=2_HOURS
CLOUDFLARE_DISPATCH_KILL_SWITCH=ACTIVE
PRODUCTION_WAKE_CLUSTER=STAGGERED
GRAFANA_METRICS_MISSING_WINDOW=180_MINUTES
RENDER_KEEP_ALIVE_POLICY=PROHIBITED
PRODUCTION_INGESTION=INTENTIONALLY_OFFLINE
FREE_TIER_INFRASTRUCTURE_RECOVERY=IN_PROGRESS
```

The incident cannot be marked closed until the Neon allowance is available again (monthly reset or equivalent free recovery), production metrics are successfully collected at the new cadence, and observed compute usage demonstrates that the database is scaling to zero between production windows.

---

## Canonical remediation record — GFA-OPS-456

### 1. Finding / symptom
The FREE_V1 production automation schedule was incompatible with the intended scale-to-zero cost model: independent reconciliation and metrics jobs repeatedly woke Neon/Render often enough to consume the monthly Neon compute allowance and undermine free-tier availability.

### 2. Root cause
Production cadence was designed per subsystem rather than from one shared infrastructure wake budget. Reconciliation connected directly to PostgreSQL every ten minutes, metrics touched Render about every fifteen minutes, and the enabled watchdog profile could touch Render every five minutes. Those independent schedules were shorter than or near the platforms' autosuspend/idle windows.

### 3. Failure scenario
Low user traffic would otherwise allow Render and Neon to sleep, but independent scheduled maintenance/monitoring work repeatedly wakes them. Neon remains active for a large fraction of the month at approximately minimum compute, eventually exhausting the 100 CU-hour allowance. The database/API then becomes unavailable or degraded before Grafana remote write is reached.

### 4. Impact
The zero-cost production environment exhausted its monthly database compute budget, production metrics source requests degraded with timeouts/5xx responses, and runtime recovery became blocked until allowance reset/equivalent free recovery. The issue also threatened to recur after provider recovery if high-frequency watchdog scheduling returned.

### 5. Severity rationale
**P1 retrospective.** This was a real production resource-exhaustion incident that consumed the full monthly free database compute allowance and materially degraded the production runtime. The severity is reconstructed from the observed impact; the source incident did not record an original severity label.

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
- reduce and stagger cadences, remove independent reconciliation cron while ingestion is offline, share future reconciliation with the ingestion wake window, and define an explicit CU-hour reserve.

### 8. Chosen remediation
Move metrics to `20 */2 * * *`, make reconciliation manual-only while ingestion is offline, target future Cloudflare primary at `:17/:47`, watchdog at `19 */2`, cluster wakes, prohibit keep-alive traffic, keep scale-to-zero enabled, and define `TARGET_MONTHLY_NEON_COMPUTE <= 60 CU-hours` with at least 40 CU-hours reserved for interactive/recovery demand.

### 9. Why this solution was selected
It preserves the same product/domain architecture on free infrastructure, accepts cold starts as a deliberate FREE_V1 trade-off, and addresses the observed active-time mechanism instead of attributing every Render error to Neon without evidence.

### 10. Rejected alternatives
A paid tier violates the current zero-cost constraint; retaining the former cadences repeats the wake pattern; disabling all production verification removes meaningful evidence; increasing timeouts does not reduce compute-active time.

### 11. Trade-offs
Metrics and watchdog evidence become less frequent and cold starts are expected. The configuration is appropriate for a non-real-time portfolio MVP, not a safety-critical or high-frequency flight-tracking SLA. Future higher cadence may require a paid capacity profile.

### 12. Regression tests / protection
Repository hardening changes the scheduled workflow contracts, removes the independent reconciliation cron while ingestion is offline, updates the Grafana missing-metrics window to match sparse collection, preserves the Cloudflare kill switch and documents a mandatory compute-budget review before any frequency increase.

### 13. Adversarial review findings
The incident evidence does not prove Neon exhaustion caused every Render 502/503/429, so the canonical finding is scoped to the internal wake-pattern/budget incompatibility rather than claiming a single-cause explanation for all upstream failures. Scale-to-zero was observed functioning after inactivity, strengthening the active-time diagnosis.

### 14. Remediation iterations
1. Neon reported 100% monthly compute allowance consumption.
2. Production schedule inventory identified independent 10-minute/15-minute wake patterns and the potential 5-minute watchdog pattern.
3. The FREE_V1 policy was redesigned around staggered wake windows and an explicit CU-hour budget.
4. PR #85 applied the hardening on current `main` with branch-protection-compatible exact history.
5. Final recovery remains intentionally open pending allowance availability and post-reset runtime evidence.

### 15. Residual risks and limitations
Autoscaling, retries, user traffic, cold-start duration and future provider cadence can raise compute above the model. Neon/Render console settings are operational prerequisites rather than repository-owned infrastructure-as-code. The 60-CU target is a project safety budget, not an upstream guarantee.

### 16. Operational or deployment consequences
Production ingestion remains intentionally offline and Cloudflare dispatch disabled. Reconciliation is manual-only. Metrics run every two hours. After recovery, ingestion/reconciliation/watchdog/metrics should reuse clustered wake windows and operators must observe real post-reset compute behavior before closing the finding.

### 17. Exact evidence
- PR #85 head `1a82b8eae63ff5e293830c630fed6a9102eb9480`, merge `4a95b7a0caae8e8581cf132945c2b1be3a7a3cca`;
- Backend CI `32478940782` SUCCESS;
- Frontend CI `32478940687` SUCCESS;
- CodeQL `32478940737` SUCCESS;
- API Load Baseline `32478940692` SUCCESS;
- live Neon evidence: ~`373.6` active hours, ~`101.2` CU-hours, average effective compute ~`0.27` CU, endpoint observed suspended roughly six minutes after last activity;
- monthly reset recorded as `2026-09-01T00:00:00Z`;
- repository state: reconciliation cron removed, metrics cadence two hours, dispatch kill switch active, `FREE_TIER_INFRASTRUCTURE_RECOVERY=IN_PROGRESS`.

### 18. Final canonical status
**IN_PROGRESS.** Repository-side scheduling/budget hardening is merged, but production recovery cannot be closed until the monthly allowance is available again, the sparse production metrics path succeeds, and observed runtime evidence shows scale-to-zero between production windows.

### 19. Prevention / future guard
Treat every scheduled production action that touches Render or Neon as a budgeted wake. Require a written monthly CU-hour calculation and wake-cluster review before increasing ingestion, watchdog, reconciliation, smoke or metrics frequency; keep FREE_V1 and SCALE/PAID as explicit deployment profiles rather than silently changing product architecture.
