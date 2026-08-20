# Production Reconciliation and Alert Stability Incident

Status: Incident closed; production recovery verified
Date: 2026-08-20
Scope: reconciliation task consumption and Grafana reconciliation-backlog stability

---

## 1. Executive Summary

After the production-ingestion resilience incident was contained, Grafana repeatedly
sent alternating notifications for `Oldest pending reconciliation above 300 seconds`.
The notifications were initially easy to misread as notification noise because many of
the visible messages were `RESOLVED` with `A=0` and `B=0`.

Mailbox evidence later confirmed real `FIRING` notifications. Two verified examples
reported:

```text
2026-08-20T10:11:20Z  A=1.445914419564e+06  B=1
2026-08-20T10:56:20Z  A=1.448627344674e+06  B=1
```

The values correspond to an oldest due reconciliation task age of approximately
16.7 days. The backlog was therefore real. The repeated resolutions were a separate
observability-stability defect.

Two independent engineering gaps were identified:

1. the repository built a `reconcile` binary and implemented the reconciliation
   worker, but production had no scheduled consumer for `derived_reconciliation_tasks`;
2. the reconciliation alert used a 20-minute `last_over_time` lookback while the
   production metrics workflow writes samples on a 15-minute schedule. A delayed or
   missed scrape could therefore age the last sample out of the 20-minute window;
   the production renderer then converted the absent result to `vector(0)`, producing
   a false healthy value and a `RESOLVED` notification even though the database backlog
   remained present.

---

## 2. Reconciliation Producer/Consumer Gap

Recoverable derived persistence failures are intentionally written as durable
reconciliation tasks. The application calls `MarkPendingDerivation` for recoverable
quality-report and trajectory persistence failures.

The repository also contains a bounded worker command:

```text
/app/reconcile
```

The command claims due pending tasks, processes at most the configured maximum per
batch, retries recoverable failures with bounded backoff, and requeues abandoned
processing leases.

Before this remediation, the production Render service executed only:

```text
/app/server
```

and no GitHub Actions production workflow scheduled `/app/reconcile`. The system could
therefore produce reconciliation work without operating a production consumer.

The remediation adds `.github/workflows/production-reconciliation.yml` with:

```text
schedule=4,14,24,34,44,54 * * * *
manual workflow_dispatch=enabled
concurrency group=production-reconciliation
cancel-in-progress=false
maximum tasks per batch=100
```

The workflow reuses the existing production-ingestion database secret boundary and
runs only the repository-owned `cmd/reconcile` path. It does not contain provider
credentials and does not enable production traffic ingestion.

---

## 3. Grafana False-Resolution Mechanism

Production metrics are forwarded by GitHub Actions on this schedule:

```text
7,22,37,52 * * * *
```

That is a nominal 15-minute cadence, but GitHub-hosted scheduled execution is not a
hard real-time scheduler and can be delayed.

The reconciliation rule previously rendered approximately as:

```promql
max(last_over_time(
  global_flight_analytics_reconciliation_oldest_pending_age_seconds{deployment_revision=~".+"}[20m]
)) or vector(0)
```

The 20-minute lookback left only about five minutes of tolerance beyond the nominal
15-minute sample cadence. When the latest sample aged out, the left-hand expression
became empty and `or vector(0)` converted telemetry absence to a healthy backlog value.
A later metrics scrape restored the real, very large backlog value and the rule fired
again.

This produced the observed operational pattern:

```text
real backlog sample
  -> FIRING
  -> metrics sample ages out of 20m lookback
  -> vector(0)
  -> RESOLVED
  -> next metrics sample
  -> real backlog returns
  -> FIRING again
```

The dedicated `gfa-metrics-missing` rule already owns telemetry continuity and checks
for missing build metrics over 25 minutes. The reconciliation rule should therefore
remain conservative across ordinary scrape jitter rather than converting one delayed
scrape into a false database-health recovery.

The production renderer now enforces a 45-minute reconciliation lookback. This spans
three nominal metrics intervals, survives one missed/delayed scheduled scrape, and
allows the dedicated missing-metrics alert to become authoritative before the
reconciliation series can disappear from the backlog rule.

The threshold itself remains unchanged:

```text
oldest due reconciliation age > 300 seconds
for 15 minutes
```

This remediation does not weaken the reconciliation SLO.

---

## 4. Regression Protection

Repository tests now require all of the following:

- rendered reconciliation lookback is exactly 45 minutes;
- the old 20-minute reconciliation lookback is not retained;
- the separate missing-metrics rule remains at 25 minutes;
- missing pending-series fallback remains explicit;
- a bounded production reconciliation workflow exists;
- the workflow has a ten-minute offset schedule and manual recovery path;
- concurrent reconciliation runs cannot cancel each other;
- the production database credential remains a GitHub Actions secret;
- the worker batch is bounded to 100 tasks;
- the workflow executes the canonical `cmd/reconcile` command.

---

## 5. Recovery Criteria

This incident is not operationally closed merely because repository tests pass.
Production closure requires:

1. merge the remediation to `main`;
2. require all pull-request checks to pass;
3. verify Grafana observability provisioning succeeds from the merged revision;
4. manually run `Production Reconciliation` against `main`;
5. inspect the worker summary for processed/completed/retry/failed counts;
6. repeat bounded recovery batches if due backlog remains;
7. run/observe a fresh production metrics scrape;
8. verify the reconciliation backlog alert reports a genuine healthy value rather
   than a value caused by telemetry absence;
9. verify the repeated `FIRING -> RESOLVED -> FIRING` notification pattern stops.

---

## 6. Safety Boundaries

This remediation does not:

- re-enable `Production Traffic Ingestion`;
- change the Airplanes.live provider configuration;
- bypass the OpenSky operational-agreement gate;
- mark an unapproved provider as production-ready;
- delete reconciliation rows manually;
- rewrite reconciliation status directly in PostgreSQL;
- suppress Grafana email delivery merely to hide the symptom;
- change the 300-second backlog threshold.

Recovery uses the same repository-owned reconciliation state machine that production
code and integration tests already exercise.

---

## 7. Production Recovery Verification

The repository remediation was merged to `main` at revision:

```text
1c98329c026e47377140f9f3eb5c2e438efd7a7b
fix: run production reconciliation and stabilize backlog alert (#80)
```

Grafana observability provisioning for the merged rules completed successfully.

The owner then manually dispatched `Production Reconciliation` against `main`:

```text
run=32372102564
conclusion=success
```

The production worker connected through the configured production database secret and
reported:

```text
requeued_stale=0
processed=0
completed=0
retries=0
failed=0
requeued_by_new_signal=0
maximum_tasks=100
PRODUCTION_RECONCILIATION_BATCH=PASS
```

Because `processed=0` and `failed=0`, no due reconciliation work remained at the time of
the production verification batch. No manual PostgreSQL status rewrite or task deletion
was used.

A fresh production metrics collection was then manually dispatched:

```text
run=32373146931
conclusion=success
PRODUCTION_METRICS_SOURCE_PREFLIGHT=PASS
GRAFANA_ALLOY_CONFIG=PASS
GRAFANA_CLOUD_REMOTE_WRITE=PASS
GRAFANA_CLOUD_QUERY_EVIDENCE=PASS
```

Mailbox verification after the fresh metrics write found no newer reconciliation
`FIRING` notification and no production-metrics-missing notification. The prior real
backlog alert had already resolved, and the repeated `FIRING -> RESOLVED -> FIRING`
pattern did not recur during closure verification.

The recovery criteria in Section 5 are therefore satisfied.

---

## 8. Current State

```text
REAL_RECONCILIATION_BACKLOG=CONFIRMED
BACKLOG_AGE_AT_INCIDENT≈16.7_DAYS
PRODUCTION_RECONCILIATION_CONSUMER=PASS
PRODUCTION_RECONCILIATION_BACKLOG=EMPTY
PRODUCTION_RECONCILIATION_BATCH=PASS
RECONCILIATION_ALERT_LOOKBACK=45_MINUTES
RECONCILIATION_ALERT_THRESHOLD=300_SECONDS_UNCHANGED
METRICS_MISSING_ALERT=25_MINUTES_PRESERVED
GRAFANA_CLOUD_REMOTE_WRITE=PASS
GRAFANA_CLOUD_QUERY_EVIDENCE=PASS
RECONCILIATION_ALERT_STABILITY=PASS
PRODUCTION_TRAFFIC_INGESTION=INTENTIONALLY_OFFLINE
PRODUCTION_RECOVERY_VERIFICATION=PASS
RECONCILIATION_INCIDENT=CLOSED
```
