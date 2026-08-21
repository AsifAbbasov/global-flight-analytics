# 194 — Free-Tier Production Infrastructure Budget

Status: **HARDENING IN PROGRESS**

## 1. Incident signal

On 2026-08-21 Neon reported that the `global-flight-analytics` project had consumed 100% of its 100 CU-hour monthly compute allowance. During the same production degradation window, the GitHub `Production Metrics Scrape` workflow repeatedly received timeouts and HTTP 502/503 responses from the Render API metrics endpoint, and Grafana subsequently reported missing external metrics.

The GitHub evidence proves that the metrics-forwarding failures happened before Grafana remote write: the protected production metrics source itself was unavailable. The available evidence does **not** prove that Neon exhaustion was the only cause of every Render 5xx response, so the root cause is recorded as an infrastructure wake-pattern and capacity incident rather than a single-provider attribution.

## 2. Free-tier incompatibility in the previous schedule

The previous production schedules were:

- production reconciliation every 10 minutes;
- external production metrics scrape every 15 minutes;
- Render free web service idle spin-down after approximately 15 minutes;
- Neon free compute scale-to-zero after approximately 5 minutes of database inactivity.

The independent reconciliation job connected directly to PostgreSQL every 10 minutes. That cadence is shorter than two Neon autosuspend windows and therefore can keep a low-traffic database effectively active for most of the month.

The external metrics workflow also contacted the Render API every 15 minutes, which is approximately the Render free-service idle threshold and therefore acts as an accidental keep-alive.

This is incompatible with the project requirement that the v1 public portfolio deployment operate on free infrastructure without artificial keep-alive traffic.

## 3. Free-tier operating policy

### 3.1 No keep-alive traffic

Production monitoring, reconciliation, or health checks MUST NOT exist only to keep Render or Neon awake.

Cold starts and database resumes are accepted platform behavior for the free v1 deployment.

### 3.2 Metrics cadence

`Production Metrics Scrape` runs every two hours:

```text
17 */2 * * *
```

This allows both Render and Neon to return to zero/idle between monitoring windows.

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

This is an estimate, not an upstream quota guarantee. Real usage can be higher because of autoscaling, user traffic, ingestion, retries, and longer-running queries.

The project therefore uses a safety target rather than trying to consume the full upstream allowance:

```text
TARGET_MONTHLY_NEON_COMPUTE <= 60 CU-hours
RESERVE_FOR_INTERACTIVE_AND_RECOVERY_WORK >= 40 CU-hours
```

## 5. Production activation budget

After traffic ingestion is restored, the target v1 cadence is intentionally low-frequency. A 30-minute ingestion cadence combined with reconciliation in the same wake window remains compatible with the analytical, non-real-time v1 product scope and leaves substantially more room than the previous independent 10-minute reconciliation loop.

Any future increase in ingestion frequency requires a new CU-hour budget calculation before activation.

## 6. Required Neon settings

For the free v1 deployment:

- scale-to-zero/autosuspend must remain enabled;
- minimum compute should remain at the smallest available size (target 0.25 CU);
- maximum autoscaling should remain conservative unless measured production demand requires more;
- no external uptime service may ping the database or Render API merely to prevent sleep.

These console settings are operational prerequisites and are not encoded in this repository unless Neon infrastructure-as-code is introduced later.

## 7. Current recovery state

```text
NEON_MONTHLY_COMPUTE_ALLOWANCE=EXHAUSTED
PRODUCTION_RECONCILIATION_CRON=REMOVED
PRODUCTION_METRICS_SCRAPE_CADENCE=2_HOURS
GRAFANA_METRICS_MISSING_WINDOW=180_MINUTES
RENDER_KEEP_ALIVE_POLICY=PROHIBITED
PRODUCTION_INGESTION=INTENTIONALLY_OFFLINE
FREE_TIER_INFRASTRUCTURE_RECOVERY=IN_PROGRESS
```

The incident cannot be marked closed until the Neon allowance is available again (monthly reset or equivalent free recovery), production metrics are successfully collected at the new cadence, and observed compute usage demonstrates that the database is scaling to zero between production windows.
