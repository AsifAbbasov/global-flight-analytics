# Zero-Cost Production Ingestion Reliability Foundation

Status: Cloudflare Worker deployed and initial live evidence verified; GitHub hourly fallback configured; final production closure pending.

## 1. Purpose

This document records the executable foundation for the zero-cost production
traffic reliability design selected after the 2026-08-06 stale-data incident.

The repository implementation lives at:

```text
infra/cloudflare/production-ingestion-reliability/
```

The Cloudflare Worker, both Cron Triggers, encrypted GitHub credential, public
health route, fresh-data watchdog skip, and stale-data recovery dispatch are now
verified live. Final production reliability closure remains intentionally open until
the remaining primary-dispatch, active-run, fallback, and exact-revision evidence pass.

## 2. Implemented repository surface

```text
infra/cloudflare/production-ingestion-reliability/
├── README.md
├── src/
│   └── index.mjs
├── test/
│   └── index.test.mjs
└── wrangler.jsonc
```

The Worker implements:

- a primary Cron Trigger at minutes `3,13,23,33,43,53`;
- a watchdog Cron Trigger every five minutes;
- exact GitHub workflow-run lookup for `main`;
- fail-closed active-run suppression;
- an eight-minute recent-success deduplication window;
- exact `workflow_dispatch` requests with dispatch provenance;
- public traffic freshness parsing aligned with the production verifier;
- stale and empty-snapshot recovery dispatch;
- future-clock-skew rejection;
- a no-store `/health` endpoint that never exposes the token.

## 3. Current production safety boundary

Cloudflare now owns the primary ten-minute schedule through:

```text
3,13,23,33,43,53 * * * *
```

GitHub Actions remains available as an offset hourly fallback:

```yaml
schedule:
  - cron: '37 * * * *'
```

The fallback cutover was permitted only after the Worker deployment, encrypted
secret, `/health` endpoint, scheduled execution, fresh watchdog skip, stale
recovery dispatch, public freshness recovery, and secret-safe live logs were
verified. Final closure remains blocked until a primary Cloudflare dispatch,
active-run suppression, one fallback execution or bounded simulation, and the
complete exact-revision production validator pass.

## 4. GitHub authorization

Use a fine-grained token restricted to:

```text
Repository: AsifAbbasov/global-flight-analytics
Repository permission: Actions — Read and write
```

The Worker uses GitHub Actions read access to inspect recent workflow runs and
write access to create a workflow dispatch. The token does not need the Neon
connection string, provider credentials, Render environment variables, API
mutation key, Grafana credentials, or repository-content write access.

## 5. Dispatch provenance

The production ingestion workflow now accepts the optional input:

```text
dispatch_source
```

Accepted values are:

```text
manual
schedule
cloudflare-primary
cloudflare-watchdog
```

Each run prints:

```text
PRODUCTION_INGESTION_DISPATCH_SOURCE=<value>
```

The ingestion command, database secret, bounded cycle, concurrency group, and
public freshness verification remain unchanged.

## 6. Free-tier capacity

The configured Worker uses two Cron Triggers. Normal operation generates:

```text
primary invocations per day=144
watchdog invocations per day=288
total scheduled invocations per day=432
```

This is intentionally far below the documented Workers Free request and Cron
Trigger limits. The Worker performs at most three outbound requests during a
stale recovery decision:

1. public traffic freshness request;
2. GitHub workflow-run lookup;
3. GitHub workflow dispatch.

## 7. Deployment procedure

Pin Wrangler during validation and deployment:

```bash
cd infra/cloudflare/production-ingestion-reliability
npx --yes wrangler@4.94.0 login
npx --yes wrangler@4.94.0 secret put GITHUB_ACTIONS_TOKEN --config wrangler.jsonc
npx --yes wrangler@4.94.0 deploy --dry-run --config wrangler.jsonc
npx --yes wrangler@4.94.0 deploy --config wrangler.jsonc
```

Record the returned Worker URL and verify:

```bash
curl --fail --silent --show-error '<worker-url>/health'
```

Deployment must be followed by the controlled live proof described in
`docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md`.

## 8. Evidence markers

Repository verification must publish:

```text
CLOUDFLARE_RELIABILITY_WORKER_TESTS=PASS
CLOUDFLARE_RELIABILITY_SOURCE_CONTRACT=PASS
CLOUDFLARE_RELIABILITY_WRANGLER_DRY_RUN=PASS
ZERO_COST_INGESTION_RELIABILITY_FOUNDATION=PASS
```

Production closure remains blocked until the live evidence markers documented
in the production deployment runbook are recorded.

## 9. Live deployment and remaining closure work

The Worker is deployed at:

```text
https://global-flight-analytics-production-ingestion-reliability.aassifabbasov.workers.dev
```

Initial live evidence is recorded in
`docs/183_CLOUDFLARE_INGESTION_LIVE_DEPLOYMENT_EVIDENCE.md`.

The following work remains:

- observe one exact primary Cloudflare dispatch after the hourly fallback cutover;
- capture one live active-run deduplication decision;
- verify one hourly GitHub fallback execution or bounded fallback simulation;
- rerun the complete exact-revision production validator;
- update final documentation markers only after every required proof passes.

Until those conditions pass, `PRODUCTION_INGESTION_RELIABILITY=PASS` must not be
claimed.
