# Zero-Cost Production Ingestion Reliability Closure

Status: production ingestion reliability closed.

## 1. Purpose

This document records the closed zero-cost production traffic-ingestion
reliability architecture selected after the 2026-08-06 scheduled-execution gap.

The repository implementation lives at:

```text
infra/cloudflare/production-ingestion-reliability/
```

## 2. Closed production architecture

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

## 3. Security boundary

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

## 4. Repository and runtime contracts

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

## 5. Live closure evidence

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
WATCHDOG_RECOVERY_RUN_ID=31103550357
PRIMARY_DISPATCH_RUN_ID=31112274607
PRIMARY_DISPATCH_HEAD_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013
ACTIVE_RUN_AND_FALLBACK_RUN_ID=31113114700
ACTIVE_RUN_AND_FALLBACK_HEAD_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013
FINAL_RUNTIME_VALIDATION_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013
FINAL_RUNTIME_VALIDATION_COMPLETED_AT=2026-08-06T15:31:58Z
```

Detailed immutable runtime evidence is recorded in
`docs/183_CLOUDFLARE_INGESTION_LIVE_DEPLOYMENT_EVIDENCE.md`.

## 6. Operational ownership

The architecture is closed, but operations remain owner-controlled:

- rotate the Cloudflare API token and GitHub fine-grained token before expiry;
- investigate a failed watchdog recovery or stale public snapshot;
- update the expected deployed revision after a new API deployment;
- preserve manual dispatch as the final recovery path;
- rerun exact-revision production validation after production application
  changes.

These are ongoing operational duties, not open implementation defects.
