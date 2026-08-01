# Production Deployment Runbook

## Purpose

This runbook deploys the existing PostgreSQL and Dockerized Go API without changing the
application architecture. The Next.js visual and public deployment phase is intentionally
deferred for a separate creative pass.

Do not commit database passwords, mutation keys, provider credentials or deployment
tokens. Configure them in platform secret stores.

## 1. Freeze the release candidate

```bash
git status -sb
git rev-parse HEAD
pnpm verify:release
pnpm verify:backend-operations-contract
```

Record the full SHA. Use the same SHA for production migrations and API smoke evidence.

## 2. Create Neon PostgreSQL

Create one project and keep two TLS connection strings. Use a direct connection string for migrations and maintenance, and use a pooled connection string for the running API.

- direct connection string for schema migrations and maintenance;
- pooled connection string for the long-running API.

Run migrations from a clean checkout with the direct URL:

```bash
PRODUCTION_DATABASE_MIGRATION_URL='paste the direct Neon URL in this terminal only' \
EXPECTED_RELEASE_SHA="$(git rev-parse HEAD)" \
pnpm migrate:production-database
```

The command rejects Neon hostnames containing `-pooler`, requires TLS and delegates to
the production `cmd/migrate` path. Do not store the URL in shell history when operating
on shared machines.

## 3. Create the Render Blueprint

The root `render.yaml` defines one free Render web service:

| Setting | Value |
| --- | --- |
| Runtime | Docker |
| Plan | Free portfolio service |
| Region | Frankfurt |
| Dockerfile | `apps/api/Dockerfile` |
| Build context | repository root |
| Health check | `/api/v1/ready` |
| Port | `API_PORT=10000` |
| Auto deploy | after linked checks pass |

During the initial Blueprint creation, supply:

```text
DATABASE_URL=<pooled Neon URL>
API_ALLOWED_ORIGINS=<temporary valid HTTPS origin until the frontend is deployed>
API_MUTATION_KEY_SHA256=<owner-generated lowercase SHA-256 digest>
```

The free Render web service does not support a pre-deploy command. Run
`pnpm migrate:production-database` before the initial deploy and before deploying any
commit that adds a migration.

Free Render services can spin down after an idle period. The API-only smoke command uses
bounded retries so the first request can wake the service before lifecycle validation.

A paid Render service may instead define a pre-deploy command that runs `/app/migrate`
with a direct database URL. Do not add that command to the free Blueprint because it
would document a capability unavailable on that plan.

## 4. Deploy and verify the API

After Blueprint creation, Render supplies `RENDER_GIT_COMMIT` during the Docker build.
The Dockerfile uses it only when an explicit `VCS_REF` build argument is absent, preserving
local and Continuous Integration build behavior. Wait for the CI-gated deploy and readiness
health check, then run:

```bash
API_BASE_URL=https://your-api.example \
EXPECTED_API_REVISION="$(git rev-parse HEAD)" \
pnpm smoke:api-production
```

Required markers:

```text
PRODUCTION_API_HEALTH=PASS
PRODUCTION_API_READINESS=PASS
PRODUCTION_API_VERSION=PASS
PRODUCTION_API_REVISION=PASS
PRODUCTION_API_SMOKE=PASS
```

## 5. Deferred Next.js phase

The later frontend deployment keeps the existing Vercel monorepo setting `Root Directory: apps/web`.
The current backend deployment may use a temporary valid HTTPS value for
`API_ALLOWED_ORIGINS`. When the Next.js creative phase is complete:

1. deploy the frontend;
2. replace `API_ALLOWED_ORIGINS` with the exact frontend HTTPS origin;
3. redeploy the API;
4. run `pnpm smoke:production` to verify frontend identity, API lifecycle, exact revision
   and CORS together.

## 6. Rollback

1. select the previous successful API deployment;
2. retain the latest forward-compatible database schema;
3. set `EXPECTED_API_REVISION` to the rollback revision;
4. rerun `pnpm smoke:api-production`;
5. record the reason and exact evidence.

## Platform references

- Render Blueprint specification: `https://render.com/docs/blueprint-spec`
- Render free-service limitations: `https://render.com/docs/free`
- Render health checks: `https://render.com/docs/health-checks`
- Neon connection pooling: `https://neon.com/docs/connect/connection-pooling`
