# Production ingestion reliability Worker

Status: production ingestion reliability closed and live-verified.

This Worker implements the zero-cost reliability design for production traffic
ingestion:

- the primary Cron Trigger requests a GitHub workflow dispatch every ten minutes;
- the watchdog checks public traffic freshness every five minutes;
- queued or active workflow runs suppress duplicate dispatches;
- a recent successful run inside the deduplication window suppresses an
  unnecessary dispatch;
- stale or empty traffic requests one recovery dispatch;
- the Worker never receives the Neon database URL or provider credentials.

## Security boundary

`GITHUB_ACTIONS_TOKEN` is the only Worker secret. Create a fine-grained token
restricted to `AsifAbbasov/global-flight-analytics` with GitHub Actions read and
write access. Do not place the token in this directory, Wrangler configuration,
GitHub variables, logs, documentation, or shell history.

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

Cloudflare recommends testing scheduled handlers through Wrangler's scheduled
test route. Start local development:

```bash
npx --yes wrangler@4.94.0 dev --test-scheduled --config wrangler.jsonc
```

Then invoke either configured expression:

```bash
curl 'http://127.0.0.1:8787/__scheduled?cron=3%2C13%2C23%2C33%2C43%2C53+*+*+*+*'
curl 'http://127.0.0.1:8787/__scheduled?cron=*%2F5+*+*+*+*'
```

The Worker is deployed at the stable account-owned `workers.dev` route. Repository
installation must never configure secrets or perform production deployment. Live primary
dispatch, watchdog recovery, active-run suppression, bounded GitHub fallback, public
freshness, and exact-revision runtime validation are closed. Secret rotation and incident
response remain separate owner-controlled operations.
