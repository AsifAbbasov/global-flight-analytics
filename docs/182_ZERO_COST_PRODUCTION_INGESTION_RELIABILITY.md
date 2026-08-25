# Zero-Cost Production Ingestion Reliability Closure

Status: CANONICAL RECONCILIATION COMPLETE  
Incident-definition PR: #55, head `dc2f6f7042ea16d13cf25ce8700831df160b8158`, merge `22a220d6ec9776637fd15f01cac9556c6e96a51e`  
Reliability-foundation PR: #56, head `f847e54599efc5076495c758fac7436cc36bc312`, merge `463436371978d970b72d2ed3c2a6190ba9048b37`  
Cloudflare-cutover PR: #58, head `2ef88bf6524f95bfa00b7363fc40d4e54f5265d0`, merge `7dfc66685247a5a1aaea87b1391624d1014d7013`  
Final closure PR: #59, head `a983ebcb2920b7371fa8210f1cc1c8eb7cfa6cf6`, merge `3bcd07df3883827904cedfe5c71ca7ad58f6967c`

## 1. Purpose

This document records the closed zero-cost production traffic-ingestion
reliability architecture selected after the 2026-08-06 scheduled-execution gap.

The repository implementation lives at:

```text
infra/cloudflare/production-ingestion-reliability/
```

This reliability defect is distinct from `GFA-OPS-447`. `GFA-OPS-447` owned the earlier absence of any production ingestion execution path. `GFA-OPS-451` below owns the later failure of the already-existing GitHub-scheduled runtime to provide a sufficiently reliable production timing and recovery boundary.

## 2. Observed scheduled-execution gap

During exact-revision production validation on 2026-08-06, the public traffic snapshot was:

```text
age_seconds=18189
maximum_age_seconds=1800
```

The repository workflow declared a ten-minute GitHub Actions schedule, but the public snapshot was approximately five hours and three minutes old.

Investigation established that:

- `PRODUCTION_INGESTION_DATABASE_URL` was configured;
- the provider request path succeeded;
- bounded ingestion cycles could persist flight states and trajectories;
- the public API read from the same production database;
- application revision, deployment, health, readiness, version and CORS checks passed;
- no application-code, schema, provider-authentication or secret failure explained the missing scheduled executions.

Manual recovery run `31076668920` on revision `855f82bf97cf0db47d1a3918f75ea70f7f2b06fe` stored four trajectories and restored public freshness to five seconds. That closed the immediate stale-data incident but did not make GitHub-hosted cron a guaranteed ten-minute production scheduler.

## 3. Closed production architecture

```text
Cloudflare primary Cron: 3,13,23,33,43,53 * * * *
Cloudflare watchdog Cron: */5 * * * *
GitHub hourly fallback: 37 * * * *
Manual fallback: workflow_dispatch
Executor: Production Traffic Ingestion
Persistence: Neon PostgreSQL
Public verification: /api/v1/traffic/current
```

Cloudflare owns the primary ten-minute schedule. The Worker checks GitHub run
state before dispatch, suppresses queued or active duplicates, applies an
eight-minute recent-success deduplication window, and uses the watchdog to
recover stale or empty public traffic.

GitHub Actions remains the isolated bounded executor. It performs one ingestion
cycle, writes through the existing production database secret, and requires
public freshness before reporting success.

The foundation increment deliberately retained the existing GitHub `*/10` schedule until the Worker source, tests, authorization boundary, deployment and controlled recovery evidence were available. The later cutover then moved GitHub scheduling to the offset hourly fallback only after live Cloudflare evidence passed.

## 4. Security boundary

`GITHUB_ACTIONS_TOKEN` is the only Worker secret. It is restricted to
`AsifAbbasov/global-flight-analytics` with GitHub Actions read and write access.

Cloudflare does not receive:

- the Neon connection string;
- provider credentials;
- Render environment variables;
- mutation or metrics keys;
- Grafana credentials;
- repository-content write access.

The stable `/health` route exposes only non-secret configuration and a boolean
secret-configured marker. Preview URLs are disabled.

## 5. Repository and runtime contracts

The Worker and repository contracts enforce:

- two exact Cron Triggers;
- fail-closed handling of unknown run states;
- exact dispatch provenance;
- exact GitHub `204` dispatch success;
- active-run suppression;
- recent-success deduplication;
- stale and empty-snapshot recovery;
- future-clock-skew rejection;
- hourly GitHub fallback;
- immutable Action pins;
- no committed Worker secret.

## 6. Exact source and CI chronology

### PR #55 — incident classification and reliability design

```text
HEAD=dc2f6f7042ea16d13cf25ce8700831df160b8158
MERGE=22a220d6ec9776637fd15f01cac9556c6e96a51e
Backend CI=31078316097 SUCCESS
Frontend CI=31078316212 SUCCESS
CodeQL=31078315797 SUCCESS
API Load Baseline=31078315805 SUCCESS
```

### PR #56 — zero-cost reliability foundation

```text
HEAD=f847e54599efc5076495c758fac7436cc36bc312
MERGE=463436371978d970b72d2ed3c2a6190ba9048b37
Backend CI=31080605862 SUCCESS
Frontend CI=31080605894 SUCCESS
CodeQL=31080605936 SUCCESS
API Load Baseline=31080605934 SUCCESS
OpenAPI Contract=31080606134 SUCCESS
Playwright E2E=31080605920 SUCCESS
```

### PR #58 — Cloudflare primary cutover

```text
HEAD=2ef88bf6524f95bfa00b7363fc40d4e54f5265d0
MERGE=7dfc66685247a5a1aaea87b1391624d1014d7013
Backend CI=31108437324 SUCCESS
Frontend CI=31108437341 SUCCESS
CodeQL=31108438528 SUCCESS
API Load Baseline=31108440177 SUCCESS
OpenAPI Contract=31108437331 SUCCESS
Playwright E2E=31108438516 SUCCESS
```

### PR #59 — final reliability evidence closure

```text
HEAD=a983ebcb2920b7371fa8210f1cc1c8eb7cfa6cf6
MERGE=3bcd07df3883827904cedfe5c71ca7ad58f6967c
Backend CI=31131144532 SUCCESS
Frontend CI=31131144610 SUCCESS
CodeQL=31131144666 SUCCESS
API Load Baseline=31131144637 SUCCESS
OpenAPI Contract=31131148232 SUCCESS
Playwright E2E=31131144949 SUCCESS
```

## 7. Live closure evidence

The following evidence passed on 2026-08-06:

```text
CLOUDFLARE_WORKER_DEPLOYMENT=PASS
CLOUDFLARE_HEALTH_ENDPOINT=PASS
CLOUDFLARE_PRIMARY_SCHEDULE=PASS
CLOUDFLARE_PRIMARY_REAL_DISPATCH=PASS
CLOUDFLARE_WATCHDOG=PASS
CLOUDFLARE_GITHUB_AUTHORIZATION_BOUNDARY=PASS
RECENT_SUCCESS_DEDUPLICATION=PASS
ACTIVE_RUN_DEDUPLICATION=PASS
STALE_TRAFFIC_RECOVERY_DISPATCH=PASS
GITHUB_SCHEDULED_FALLBACK=PASS
MANUAL_RECOVERY=PASS
POST_RECOVERY_PUBLIC_FRESHNESS=PASS
LIVE_PRODUCTION_RUNTIME_VALIDATION=PASS
PRODUCTION_INGESTION_RELIABILITY=PASS
```

Exact run identities:

```text
CLOSURE_REPOSITORY_REVISION=7dfc66685247a5a1aaea87b1391624d1014d7013
WATCHDOG_RECOVERY_RUN_ID=31103550357
PRIMARY_DISPATCH_RUN_ID=31112274607
PRIMARY_DISPATCH_HEAD_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013
ACTIVE_RUN_AND_FALLBACK_RUN_ID=31113114700
ACTIVE_RUN_AND_FALLBACK_HEAD_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013
FINAL_RUNTIME_VALIDATION_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013
FINAL_RUNTIME_VALIDATION_COMPLETED_AT=2026-08-06T15:31:58Z
```

Detailed repository-recorded closure evidence is documented in
`docs/183_CLOUDFLARE_INGESTION_LIVE_DEPLOYMENT_EVIDENCE.md`.

## 8. Evidence ownership and verification boundary

The repository records exact run identities, the closure revision, the
runtime-validation timestamp, and stable closure markers. GitHub Actions retains
the immutable execution history. The final validator log remains owner-local,
non-secret supporting evidence and is not committed to the repository.

The permanent verifier protects cross-document consistency and rejects stale
`PENDING` claims. It does not independently prove that a historical production
event occurred; that event evidence remains owned by the referenced GitHub runs
and the owner-controlled runtime report.

## 9. Operational ownership

The architecture is historically closed, but operations remain owner-controlled:

- rotate the Cloudflare API token and GitHub fine-grained token before expiry;
- investigate a failed watchdog recovery or stale public snapshot;
- update the expected deployed revision after a new API deployment;
- preserve manual dispatch as the final recovery path;
- rerun exact-revision production validation after production application changes.

These are ongoing operational duties, not a reopening of the historical scheduler-reliability implementation finding.

Later provider withdrawal, provider authorization, Neon quota exhaustion, or an intentionally disabled production ingestion switch are separate operational incidents owned by later documents. They do not rewrite this historical closure unless the same scheduler-reliability defect recurs.

---

## Canonical remediation record — GFA-OPS-451

### 1. Finding / symptom

Production ingestion depended on GitHub-hosted scheduled execution as its sole regular timing authority, yet a live validation observed a traffic snapshot `18,189` seconds old despite a declared ten-minute cron cadence and an otherwise healthy ingestion/data path.

### 2. Root cause

The architecture treated a best-effort GitHub scheduled workflow as the primary production scheduler without an independent freshness watchdog, cross-provider timing authority, or automatic stale-data recovery path. The executor itself was healthy; the reliability gap was in schedule ownership and detection/recovery redundancy.

### 3. Failure scenario

GitHub scheduled events fail to execute near the intended cadence for several hours. No application or provider error fires because no ingestion run starts. The API stays healthy and continues serving the last persisted observations, which become materially stale until an owner notices and manually dispatches the workflow.

### 4. Impact

The public current-traffic product surface could remain apparently available while its core observations were hours stale. That is a production correctness/availability failure for a feature explicitly presented as current traffic, even though no stored data was corrupted.

### 5. Severity rationale

**P1 retrospective.** The observed incident left production current-traffic data approximately five hours stale while API/application health remained green, creating a silent material product-correctness gap. This severity is reconstructed from the failure mode; no original severity label is claimed.

### 6. Existing guarantees violated

- current-traffic production must have an explicit reliable timing owner;
- public API health must not be treated as proof that observation time is advancing;
- freshness degradation must be detected independently of scheduler execution;
- one scheduler/provider failure must have a bounded recovery path under the zero-cost architecture;
- duplicate recovery dispatches must remain controlled.

### 7. Considered solutions

- keep GitHub `*/10` as the only primary scheduler and rely on manual recovery;
- add a paid always-on worker or paid cron service;
- move ingestion directly into Cloudflare;
- use Cloudflare Cron as an independent primary timing authority/watchdog while retaining GitHub Actions as the isolated bounded executor, offset fallback, and manual final fallback.

### 8. Chosen remediation

Deploy a repository-owned Cloudflare Worker with a ten-minute primary Cron Trigger and five-minute freshness watchdog; inspect GitHub workflow run state before dispatch; suppress active/recent duplicate runs; dispatch stale/empty recovery only when needed; retain an offset hourly GitHub schedule and manual `workflow_dispatch`; continue using the existing bounded ingestion workflow as the only database-writing executor.

### 9. Why this solution was selected

It adds an independent zero-cost scheduling/detection plane without duplicating ingestion logic or moving database/provider credentials into Cloudflare. The architecture separates timing, freshness detection, bounded execution, persistence, and final public verification.

### 10. Rejected alternatives

- GitHub-only scheduling was rejected as the observed single reliability boundary;
- paid workers were rejected by the project's zero-cost constraint;
- direct Cloudflare ingestion was rejected because it would duplicate or relocate the mature Go ingestion/data-secret boundary;
- manual-only recovery was rejected because it does not detect or repair stale data automatically.

### 11. Trade-offs

The design introduces another platform and a restricted GitHub Actions token, plus cross-provider coordination logic. Cloudflare Cron and GitHub fallback are still external best-effort services; the architecture reduces single-scheduler dependence rather than creating a hard real-time SLA.

### 12. Regression tests / protection

Reliability tests cover exact cron expressions, run-state lookup, fail-closed unknown statuses, recent-success and active-run suppression, stale/empty recovery dispatch, future clock skew, provenance, security boundaries, and fallback cadence. Source verification and Wrangler dry-run are required by Backend/release CI. Document 183 preserves exact live runtime evidence.

### 13. Adversarial review findings

The remediation deliberately kept the operating GitHub ten-minute schedule primary while the Cloudflare foundation was only source code. Cutover occurred only after deployment, secret, health, cron, watchdog recovery, freshness, and secret-safe-log evidence passed, preventing the reliability fix from creating a new scheduling gap.

### 14. Remediation iterations

1. PR #55 classified the observed five-hour scheduled-execution gap and selected the zero-cost architecture while preserving manual recovery.
2. PR #56 added the tested Cloudflare reliability foundation but explicitly did not deploy or cut over production.
3. PR #58 deployed/cut over Cloudflare primary and moved GitHub to the hourly fallback after initial live evidence.
4. PR #59 recorded the real primary dispatch, active-run suppression, fallback proof and final exact-revision runtime validation, then closed the reliability stage.

### 15. Residual risks and limitations

Cloudflare, GitHub Actions, the provider and database remain external dependencies. The architecture cannot guarantee exact ten-minute execution or safety-critical freshness. Token rotation and exact-revision revalidation remain owner duties. Later provider/free-tier incidents are separate failure modes and may intentionally keep ingestion offline.

### 16. Operational or deployment consequences

Cloudflare becomes the primary scheduler and freshness watchdog, GitHub's scheduled event becomes an offset hourly fallback, and manual dispatch remains the final recovery path. Cloudflare receives only the restricted GitHub Actions token; database/provider secrets remain with the existing executor/runtime boundaries.

### 17. Exact evidence

- incident PR #55 head `dc2f6f7042ea16d13cf25ce8700831df160b8158`, merge `22a220d6ec9776637fd15f01cac9556c6e96a51e`;
- observed snapshot age `18189` seconds versus `1800` maximum;
- manual recovery run `31076668920` on `855f82bf97cf0db47d1a3918f75ea70f7f2b06fe`, restoring freshness to five seconds;
- PR #56 head `f847e54599efc5076495c758fac7436cc36bc312`, merge `463436371978d970b72d2ed3c2a6190ba9048b37`, all six exact-head workflows SUCCESS;
- PR #58 head `2ef88bf6524f95bfa00b7363fc40d4e54f5265d0`, merge/closure revision `7dfc66685247a5a1aaea87b1391624d1014d7013`, all six exact-head workflows SUCCESS;
- watchdog recovery run `31103550357`;
- primary dispatch run `31112274607` on closure revision;
- active-run/fallback run `31113114700` on closure revision;
- final runtime validation `7dfc66685247a5a1aaea87b1391624d1014d7013` at `2026-08-06T15:31:58Z`;
- PR #59 head `a983ebcb2920b7371fa8210f1cc1c8eb7cfa6cf6`, merge `3bcd07df3883827904cedfe5c71ca7ad58f6967c`, all six exact-head workflows SUCCESS;
- Document 183 records `PRODUCTION_INGESTION_RELIABILITY=PASS`.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

Keep the cross-provider scheduler/watchdog/fallback contracts, duplicate-dispatch guards, public freshness verification, exact-revision runtime validation, and stale-closure-document verifier. Do not infer current production freshness from historical closure after a later provider, quota, deployment, or kill-switch incident.
