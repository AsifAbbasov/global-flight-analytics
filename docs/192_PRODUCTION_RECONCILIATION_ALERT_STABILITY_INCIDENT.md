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

---

## Canonical remediation record — GFA-OPS-454

### 1. Finding / symptom
Production could durably create `derived_reconciliation_tasks`, but no production scheduler or service actually consumed the repository-owned reconciliation worker. A real due backlog reached approximately 16.7 days.

### 2. Root cause
The reconciliation producer and worker implementation existed, but production composition operated only `/app/server`; no production workflow owned recurring execution of `cmd/reconcile`.

### 3. Failure scenario
Recoverable derived-data persistence failures create durable reconciliation tasks. The API continues serving while no production consumer claims those tasks. Pending work ages indefinitely and the oldest-due age grows even though the worker code itself is correct and tested.

### 4. Impact
Recoverable derived-data failures could remain unreconciled for days, leaving durable repair work unprocessed and undermining the intended eventual-repair guarantee.

### 5. Severity rationale
**P1 retrospective.** Production accumulated a real approximately 16.7-day reconciliation backlog because a durable repair queue had no operated consumer. The severity is reconstructed from observed production impact rather than claimed as an original historical label.

### 6. Existing guarantees violated
- every durable production repair queue requires an explicit production consumer owner;
- recoverable derived persistence failures must eventually receive bounded processing attempts;
- a compiled worker binary is not production reachability evidence;
- reconciliation work must be observable and recoverable without manual database mutation.

### 7. Considered solutions
- rely on operators to run `/app/reconcile` manually after alerts;
- run reconciliation continuously inside the web server;
- add a separate long-running service;
- add a bounded scheduled GitHub Actions consumer reusing the existing production database-secret boundary.

### 8. Chosen remediation
Add `Production Reconciliation` with an offset ten-minute schedule, manual `workflow_dispatch`, serialized execution, `cancel-in-progress=false`, and a maximum of 100 tasks per batch, executing the canonical `cmd/reconcile` path.

### 9. Why this solution was selected
It operates the existing tested reconciliation state machine without adding another paid always-on service, keeps database access in an existing GitHub Actions secret boundary, and bounds each recovery batch.

### 10. Rejected alternatives
Manual-only execution does not provide ownership; embedding reconciliation into the web server couples unrelated lifecycles; a new always-on service conflicts with the zero-cost deployment constraint and was unnecessary for bounded backlog processing.

### 11. Trade-offs
Scheduled GitHub execution is best-effort rather than real-time, and an independent reconciliation cron later proved too expensive for the free-tier wake budget. Document 194 therefore removes the independent cron while ingestion is offline and requires future reconciliation to share the ingestion wake window.

### 12. Regression tests / protection
Repository verification requires the production reconciliation workflow, canonical command, bounded batch size, serialized concurrency, production secret use and manual recovery path. Later free-tier policy explicitly owns cadence changes rather than deleting the consumer contract.

### 13. Adversarial review findings
The remediation does not mark rows complete manually, delete backlog records, re-enable traffic ingestion, or bypass the existing worker retry/lease semantics. Runtime closure required a real production worker run.

### 14. Remediation iterations
1. Alert investigation established that the backlog was real rather than notification noise.
2. Source review found a producer/consumer reachability gap.
3. PR #80 added the bounded production workflow.
4. PR #81 recorded production run `32372102564`, which completed successfully and observed no remaining due work.
5. Document 194 later changed scheduling for free-tier budget reasons without reopening the historical missing-consumer defect.

### 15. Residual risks and limitations
A bounded consumer does not guarantee exact execution cadence on GitHub-hosted schedules. Future FREE_V1 operation intentionally avoids an independent high-frequency reconciliation cron and must preserve consumer ownership inside the ingestion/recovery wake window.

### 16. Operational or deployment consequences
Production gained an explicit reconciliation execution path and manual recovery command. Under the current intentionally-offline ingestion state, reconciliation is manual-only to preserve free-tier scale-to-zero; future automatic execution should be coupled to successful ingestion.

### 17. Exact evidence
- PR #80 head `f1e79a0ae935b144870f46649926fc4066221c3e`, merge `1c98329c026e47377140f9f3eb5c2e438efd7a7b`;
- Backend CI `32364298598` SUCCESS;
- Frontend CI `32364298614` SUCCESS;
- CodeQL `32364298611` SUCCESS;
- API Load Baseline `32364298586` SUCCESS;
- incident backlog evidence around `1.445914e+06`–`1.448627e+06` seconds, approximately 16.7 days;
- production reconciliation run `32372102564` SUCCESS with `PRODUCTION_RECONCILIATION_BATCH=PASS`, `processed=0`, `failed=0`, `maximum_tasks=100`;
- PR #81 head `060347b8f7cf7c94008f0f618a5f8304902ce27e`, merge `40f6eeee1f3a40dc7d11409aa1c656f316004a45` recorded runtime recovery closure.

### 18. Final canonical status
**CLOSED.** The missing production-consumer defect is closed. Current cadence is governed separately by the FREE_V1 compute-budget policy.

### 19. Prevention / future guard
Any durable producer introduced to production must have an explicit operated consumer, reachability evidence, bounded execution policy and recovery path. Deployment-budget changes may alter cadence but must not silently remove consumer ownership.

---

## Canonical remediation record — GFA-OBS-455

### 1. Finding / symptom
The reconciliation backlog alert could emit `RESOLVED` with a healthy zero value while the database still contained a real, very old due backlog.

### 2. Root cause
The alert used a 20-minute `last_over_time` window against a nominal 15-minute GitHub metrics cadence. When a scrape was delayed or missed, the series could age out; `or vector(0)` converted telemetry absence into an apparent zero backlog.

### 3. Failure scenario
A real backlog produces a firing sample. No fresh metrics sample arrives inside the narrow extra tolerance. The series disappears from the 20-minute query window, the fallback emits zero and Grafana resolves the alert. The next metrics write restores the same large backlog and the alert fires again.

### 4. Impact
Monitoring could falsely report production recovery and create alternating `FIRING -> RESOLVED -> FIRING` notifications while the underlying reconciliation problem was unchanged, making real incident state easy to misinterpret.

### 5. Severity rationale
**P1 retrospective.** The defect produced false healthy production monitoring for a real approximately 16.7-day durable backlog. Because it directly undermined the trust boundary used to decide whether production repair work was healthy, it is classified P1 retrospectively.

### 6. Existing guarantees violated
- alert resolution must reflect domain recovery, not telemetry disappearance;
- missing telemetry and healthy zero backlog must remain distinguishable in ownership;
- reconciliation SLO threshold must not be weakened to hide sparse metrics;
- the dedicated missing-metrics rule must become authoritative before backlog evidence ages out.

### 7. Considered solutions
- remove `or vector(0)` entirely;
- shorten the metrics interval;
- weaken or raise the 300-second backlog threshold;
- increase the backlog lookback so it safely exceeds scrape jitter and the separate missing-metrics rule's ownership window.

### 8. Chosen remediation
Enforce a 45-minute reconciliation lookback, preserve the 300-second threshold and 15-minute `for` duration, and preserve the dedicated missing-metrics rule at 25 minutes.

### 9. Why this solution was selected
A 45-minute window spans three nominal metrics intervals and survives an ordinary delayed/missed scrape while allowing the separate 25-minute telemetry-continuity alert to fire before backlog evidence can disappear.

### 10. Rejected alternatives
Changing the backlog SLO would hide the domain problem; merely scraping more often worsens free-tier wake pressure; treating missing telemetry as zero without a wider evidence window preserves the false-recovery mechanism.

### 11. Trade-offs
The backlog rule can retain an older sample longer, so the separate missing-metrics alert must own telemetry freshness. This is deliberate separation of domain backlog state from observability continuity.

### 12. Regression tests / protection
Tests require exactly a 45-minute rendered reconciliation lookback, reject the old 20-minute form, preserve the 25-minute missing-metrics rule, preserve explicit pending-series fallback semantics and keep the 300-second backlog threshold unchanged.

### 13. Adversarial review findings
The remediation does not silence email delivery, delete the `RESOLVED` state, or weaken the backlog threshold. Runtime closure required a fresh metrics write and confirmation that the oscillating notification pattern did not recur.

### 14. Remediation iterations
1. Mailbox evidence separated real `FIRING` backlog alerts from apparently healthy `RESOLVED` notifications.
2. Query analysis identified the 20-minute lookback plus `vector(0)` mechanism.
3. PR #80 moved the lookback to 45 minutes with regression protection.
4. PR #81 recorded a fresh production metrics run and stable alert behavior after recovery.

### 15. Residual risks and limitations
Grafana still depends on scheduled external metrics delivery. Extended outages are handled by the separate missing-metrics alert rather than the backlog rule. Current FREE_V1 metrics cadence later moved to two hours and Document 194 correspondingly owns the 180-minute missing-metrics window.

### 16. Operational or deployment consequences
Grafana provisioning must deploy the widened reconciliation query. Operators should interpret the backlog rule together with the metrics-missing rule rather than treating a zero caused by missing telemetry as independent proof of database health.

### 17. Exact evidence
- PR #80 head `f1e79a0ae935b144870f46649926fc4066221c3e`, merge `1c98329c026e47377140f9f3eb5c2e438efd7a7b`;
- Backend CI `32364298598`, Frontend CI `32364298614`, CodeQL `32364298611`, API Load Baseline `32364298586` — all SUCCESS;
- real firing values `A=1.445914419564e+06` and `A=1.448627344674e+06` with `B=1`;
- production metrics run `32373146931` SUCCESS with `PRODUCTION_METRICS_SOURCE_PREFLIGHT=PASS`, `GRAFANA_ALLOY_CONFIG=PASS`, `GRAFANA_CLOUD_REMOTE_WRITE=PASS`, `GRAFANA_CLOUD_QUERY_EVIDENCE=PASS`;
- post-write verification observed no newer reconciliation `FIRING` and no metrics-missing notification during closure verification;
- PR #81 merge `40f6eeee1f3a40dc7d11409aa1c656f316004a45` records final incident closure evidence.

### 18. Final canonical status
**CLOSED.** The false-resolution mechanism at this historical cadence is closed; later FREE_V1 observability cadence changes are separately budget-governed.

### 19. Prevention / future guard
Whenever a Prometheus rule uses a missing-series fallback, verify that the evidence lookback cannot expire before the dedicated telemetry-continuity alert becomes authoritative. Cadence changes must update both windows together without weakening the domain SLO.
