# Cloudflare Ingestion Live Deployment Evidence

Status: initial live deployment and recovery evidence verified; final reliability closure pending.

## 1. Deployment identity

```text
DEPLOYMENT_DATE=2026-08-06
REPOSITORY_REVISION=ba5f8331d39035c95e2c4c3caa4b2f9ae9db4e1e
WORKER_NAME=global-flight-analytics-production-ingestion-reliability
WORKER_VERSION_ID=b6e4d1eb-1c0e-4eb5-aad9-f902dfadc80a
WORKER_URL=https://global-flight-analytics-production-ingestion-reliability.aassifabbasov.workers.dev
PRIMARY_CRON=3,13,23,33,43,53 * * * *
WATCHDOG_CRON=*/5 * * * *
GITHUB_FALLBACK_CRON=37 * * * *
```

Preview URLs are intentionally disabled. The stable account-owned `workers.dev`
route remains enabled.

## 2. Deployment and health evidence

The corrected Wrangler deployment uploaded the Worker, preserved
`GITHUB_ACTIONS_TOKEN` as a hidden binding, and deployed both Cron Triggers.

The public `/health` route returned HTTP `200` with the exact expected
configuration and no credential material:

```text
CLOUDFLARE_WORKER_DEPLOYMENT=PASS
CLOUDFLARE_HEALTH_ENDPOINT=PASS
GITHUB_ACTIONS_SECRET=CONFIGURED
PRIMARY_CRON_TRIGGER=DEPLOYED
WATCHDOG_CRON_TRIGGER=DEPLOYED
```

## 3. Live scheduled-execution evidence

A primary execution at 2026-08-06 17:03 Asia/Baku completed successfully and
suppressed a duplicate because run `31103550357` had succeeded inside the
eight-minute deduplication window:

```text
CLOUDFLARE_PRIMARY_SCHEDULE=PASS
RECENT_SUCCESS_DEDUPLICATION=PASS
PRIMARY_ACTION=skipped-recent-success
```

Watchdog executions at 17:05 and 17:10 Asia/Baku observed three aircraft and
correctly skipped recovery while traffic remained below the 1,800-second
freshness limit:

```text
CLOUDFLARE_WATCHDOG=PASS
WATCHDOG_ACTION=skipped-fresh-traffic
AIRCRAFT_COUNT=3
OBSERVED_AGE_SECONDS=572,872
```

## 4. Stale-recovery dispatch evidence

Before the live tail session, the watchdog created GitHub Actions run
`31103550357` with exact provenance:

```text
PRODUCTION_INGESTION_DISPATCH_SOURCE=cloudflare-watchdog
GITHUB_RUN_ID=31103550357
GITHUB_RUN_STATUS=completed
GITHUB_RUN_CONCLUSION=success
```

The bounded ingestion cycle stored three usable states and three trajectories:

```text
LOADED=3
USABLE=3
STORED=3
TRAJECTORIES=3
```

The workflow then verified public freshness:

```text
PRODUCTION_TRAFFIC_FRESHNESS=PASS
LATEST_AGE_SECONDS=6
MAXIMUM_AGE_SECONDS=1800
STALE_TRAFFIC_RECOVERY_DISPATCH=PASS
POST_RECOVERY_PUBLIC_FRESHNESS=PASS
```

This proves the Worker GitHub token can read workflow state and create the
restricted `workflow_dispatch` request. Cloudflare never receives the production
database URL or provider credentials.

## 5. Secret boundary

The Worker health response exposes only a boolean token-configured marker.
Wrangler displayed the binding as `(hidden)`, and the observed live Worker logs
contained no Cloudflare or GitHub token values.

```text
CLOUDFLARE_GITHUB_AUTHORIZATION_BOUNDARY=PASS
LIVE_LOG_SECRET_EXPOSURE=NOT_OBSERVED
```

## 6. Fallback cutover

Cloudflare is now the primary ten-minute scheduler. GitHub Actions remains
configured as an offset hourly fallback:

```text
37 * * * *
```

The fallback cadence avoids the Cloudflare primary minutes and reduces
unnecessary duplicate scheduled runs. `workflow_dispatch` remains the final
manual recovery path.

## 7. Remaining closure evidence

The following evidence is still required:

```text
CLOUDFLARE_PRIMARY_REAL_DISPATCH=PENDING
ACTIVE_RUN_DEDUPLICATION=PENDING
GITHUB_SCHEDULED_FALLBACK=PENDING
FINAL_EXACT_REVISION_RUNTIME_VALIDATION=PENDING
PRODUCTION_INGESTION_RELIABILITY=PENDING
```

No final reliability `PASS` claim is permitted until all remaining markers are
verified.
