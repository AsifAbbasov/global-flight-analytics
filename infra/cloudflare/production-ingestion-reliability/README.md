# Production ingestion reliability Worker

Status: production ingestion reliability implementation is hardened; production dispatch remains intentionally disabled while provider and free-tier recovery are open.

This Worker implements the zero-cost reliability design for production traffic ingestion:

- the primary Cron Trigger requests a GitHub workflow dispatch every 30 minutes;
- the watchdog checks public traffic freshness every two hours;
- the primary and watchdog schedules are deliberately sparse so they do not act as an artificial Render/Neon keep-alive;
- queued or active workflow runs suppress duplicate dispatches;
- a recent successful run inside the deduplication window suppresses an unnecessary dispatch;
- a recent failed run opens a bounded six-hour circuit breaker instead of creating an endless failure/dispatch loop;
- `DISPATCH_ENABLED=false` is a fail-closed production kill switch and is intentionally active while the upstream provider and free-tier infrastructure incidents are unresolved;
- stale or empty traffic requests one recovery dispatch;
- the Worker never receives the Neon database URL or provider credentials.

## Free-tier scheduling boundary

The deployed schedule is:

```text
primary:  17,47 * * * *
watchdog: 19 */2 * * *
```

When dispatch is disabled, both Cron Triggers return before any GitHub or Render network request.

When production traffic ingestion is restored, the primary cadence provides at most two scheduled ingestion wake windows per hour. The watchdog runs every two hours and is intentionally placed two minutes after the `:17` primary window so a healthy system can reuse the same Render/Neon wake period instead of creating a separate keep-alive cycle.

The GitHub production ingestion workflow should remain dispatch/manual-owned rather than becoming a second independent high-frequency scheduler. Reconciliation should execute in the same database wake window after ingestion instead of using its own cron.

Any increase in primary or watchdog frequency requires a fresh free-tier compute-budget review.

## Security boundary

`GITHUB_ACTIONS_TOKEN` is the only Worker secret. Create a fine-grained token restricted to `AsifAbbasov/global-flight-analytics` with GitHub Actions read and write access. Do not place the token in this directory, Wrangler configuration, GitHub variables, logs, documentation, or shell history.

Configure the secret interactively:

```bash
cd infra/cloudflare/production-ingestion-reliability
npx --yes wrangler@4.94.0 secret put GITHUB_ACTIONS_TOKEN --config wrangler.jsonc
```

## Local verification

```bash
node --test test/*.test.mjs
npx --yes wrangler@4.94.0 deploy --dry-run --config wrangler.jsonc
```

Cloudflare recommends testing scheduled handlers through Wrangler's scheduled test route. Start local development:

```bash
npx --yes wrangler@4.94.0 dev --test-scheduled --config wrangler.jsonc
```

Then invoke either configured expression:

```bash
curl 'http://127.0.0.1:8787/__scheduled?cron=17%2C47+*+*+*+*'
curl 'http://127.0.0.1:8787/__scheduled?cron=19+*%2F2+*+*+*'
```

The Worker is deployed at the stable account-owned `workers.dev` route. Repository installation must never configure secrets or perform production deployment. Live primary dispatch, watchdog recovery, active-run suppression, bounded GitHub fallback, public freshness, and exact-revision runtime validation remain established capabilities. Secret rotation, provider recovery, free-tier budget verification, and incident response remain separate owner-controlled operations.
