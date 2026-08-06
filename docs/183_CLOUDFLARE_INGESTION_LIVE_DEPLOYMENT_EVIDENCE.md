# Cloudflare Ingestion Live Deployment and Closure Evidence

Status: production ingestion reliability closed.

## 1. Deployment identity

```text
DEPLOYMENT_DATE=2026-08-06
CLOSURE_REPOSITORY_REVISION=7dfc66685247a5a1aaea87b1391624d1014d7013
WORKER_NAME=global-flight-analytics-production-ingestion-reliability
WORKER_VERSION_ID=b6e4d1eb-1c0e-4eb5-aad9-f902dfadc80a
WORKER_URL=https://global-flight-analytics-production-ingestion-reliability.aassifabbasov.workers.dev
PRIMARY_CRON=3,13,23,33,43,53 * * * *
WATCHDOG_CRON=*/5 * * * *
GITHUB_FALLBACK_CRON=37 * * * *
```

Preview URLs are disabled. The stable account-owned `workers.dev` route remains
enabled.

## 2. Deployment, health, and secret boundary

```text
CLOUDFLARE_WORKER_DEPLOYMENT=PASS
CLOUDFLARE_HEALTH_ENDPOINT=PASS
GITHUB_ACTIONS_SECRET=CONFIGURED
PRIMARY_CRON_TRIGGER=DEPLOYED
WATCHDOG_CRON_TRIGGER=DEPLOYED
CLOUDFLARE_GITHUB_AUTHORIZATION_BOUNDARY=PASS
LIVE_LOG_SECRET_EXPOSURE=NOT_OBSERVED
```

The Worker receives only the restricted GitHub Actions token. It does not
receive database, provider, Render, mutation, metrics, or Grafana credentials.

## 3. Watchdog stale recovery and fresh skip

The watchdog created GitHub Actions recovery run `31103550357` with exact
provenance:

```text
PRODUCTION_INGESTION_DISPATCH_SOURCE=cloudflare-watchdog
WATCHDOG_RECOVERY_RUN_ID=31103550357
GITHUB_RUN_ID=31103550357
GITHUB_RUN_CONCLUSION=success
LOADED=3
STORED=3
TRAJECTORIES=3
PRODUCTION_TRAFFIC_FRESHNESS=PASS
LATEST_AGE_SECONDS=6
STALE_TRAFFIC_RECOVERY_DISPATCH=PASS
POST_RECOVERY_PUBLIC_FRESHNESS=PASS
```

Subsequent watchdog executions observed fresh traffic and correctly skipped
recovery.

## 4. Primary dispatch and recent-success deduplication

A scheduled primary execution first proved the bounded recent-success decision:

```text
CLOUDFLARE_PRIMARY_SCHEDULE=PASS
RECENT_SUCCESS_DEDUPLICATION=PASS
PRIMARY_ACTION=skipped-recent-success
```

After the fallback cutover, Cloudflare created the real primary run:

```text
PRIMARY_DISPATCH_RUN_ID=31112274607
PRIMARY_DISPATCH_CREATED_AT=2026-08-06T14:43:43Z
PRIMARY_DISPATCH_HEAD_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013
PRIMARY_DISPATCH_CONCLUSION=success
PRODUCTION_INGESTION_DISPATCH_SOURCE=cloudflare-primary
CLOUDFLARE_PRIMARY_REAL_DISPATCH=PASS
PRIMARY_GITHUB_PROVENANCE=PASS
PRIMARY_POST_DISPATCH_FRESHNESS=PASS
```

## 5. Active-run duplicate suppression and GitHub fallback

The bounded fallback simulation created run `31113114700`. While that run was
`in_progress`, the 18:53 Asia/Baku Cloudflare primary execution observed it and
published:

```text
PRIMARY_ACTION=skipped-active-run
DECISION_MARKER=ACTIVE_RUN_DEDUPLICATION
ACTIVE_RUN_ID=31113114700
ACTIVE_RUN_DEDUPLICATION=PASS
```

The same run then completed successfully on the exact closure revision:

```text
ACTIVE_RUN_AND_FALLBACK_RUN_ID=31113114700
ACTIVE_RUN_AND_FALLBACK_HEAD_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013
FALLBACK_SIMULATION_RUN_ID=31113114700
FALLBACK_SIMULATION_HEAD_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013
FALLBACK_SIMULATION_CONCLUSION=success
PRODUCTION_INGESTION_DISPATCH_SOURCE=schedule
GITHUB_SCHEDULED_FALLBACK=PASS
FALLBACK_POST_INGESTION_FRESHNESS=PASS
```

This proves the cross-provider fallback can execute and the Cloudflare primary
does not create a duplicate while it is active.

## 6. Final exact-revision runtime validation

The complete read-only production runtime validator passed on the closure
revision:

```text
FINAL_RUNTIME_VALIDATION_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013
FINAL_RUNTIME_VALIDATION_COMPLETED_AT=2026-08-06T15:31:58Z
PRODUCTION_API_EXACT_REVISION=PASS
PRODUCTION_RELEASE_SMOKE=PASS
PRODUCTION_FRONTEND_API_CORS=PASS
PRODUCTION_TRAFFIC_FRESHNESS=PASS
PRODUCTION_TRAFFIC_DATA=PASS
PRODUCTION_API_DOCS_HTML=PASS
PRODUCTION_API_DOCS_SECURITY_HEADERS=PASS
PRODUCTION_API_DOCS_NO_CDN=PASS
PRODUCTION_OPENAPI_LIVE_PARITY=PASS
PRODUCTION_OPENAPI_ETAG=PASS
PRODUCTION_OPENAPI_CONDITIONAL_GET=PASS
PRODUCTION_MUTATION_AUTHENTICATION_BOUNDARY=PASS
VALIDATION_NON_MUTATING=PASS
LIVE_PRODUCTION_RUNTIME_VALIDATION=PASS
```

The command emits an owner-local validation log under `~/Downloads`.
The local report is supporting operational evidence and is not committed to the
repository. It contains no token values.

## 7. Evidence ownership and verification boundary

This document stores the exact closure metadata that can be reviewed in source:
run identities, exact revisions, the validation timestamp, and stable result
markers. The referenced GitHub Actions runs retain the immutable execution
history.

The owner-local runtime report supports the final validation event without
placing operational logs in source control. The repository verifier checks that
Documents 163, 182, and 183 remain internally consistent and contain no stale
closure `PENDING` markers. It does not replace live runtime revalidation after a
future deployment.

## 8. Final closure

```text
CLOUDFLARE_PRIMARY_SCHEDULE=PASS
CLOUDFLARE_WATCHDOG=PASS
CLOUDFLARE_GITHUB_AUTHORIZATION_BOUNDARY=PASS
ACTIVE_RUN_DEDUPLICATION=PASS
STALE_TRAFFIC_RECOVERY_DISPATCH=PASS
GITHUB_SCHEDULED_FALLBACK=PASS
MANUAL_RECOVERY=PASS
POST_RECOVERY_PUBLIC_FRESHNESS=PASS
LIVE_PRODUCTION_RUNTIME_VALIDATION=PASS
PRODUCTION_INGESTION_RELIABILITY=PASS
```

The production ingestion reliability implementation stage is closed. Token
rotation, alert response, and revision-specific revalidation remain normal
operations.
