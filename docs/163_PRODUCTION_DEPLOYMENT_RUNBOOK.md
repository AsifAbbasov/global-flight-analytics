# Production Deployment Runbook

<!-- RELEASE-TRUTH-DEPLOYMENT-REVISION-V1 -->

## Purpose

This runbook deploys the complete production path without changing the application
architecture:

- Neon PostgreSQL;
- the Dockerized Go API on Render;
- the Next.js frontend on Vercel;
- exact-origin Cross-Origin Resource Sharing between the frontend and API;
- production migration, lifecycle, revision, frontend, and browser integration smoke.

Do not commit database passwords, mutation keys, metrics keys, provider credentials,
deployment tokens, or private connection strings. Configure them in platform secret
stores.

## Verified production deployment

The deployment completed on 2026-08-02 with these public endpoints:

- Frontend: `https://global-flight-analytics-web.vercel.app`
- API: `https://global-flight-analytics-api.onrender.com`

Evidence revisions:

- historically verified production application SHA (2026-08-02): `6bca02a8ed1487195b165ae9ced3ca687a373666`;
- production migration evidence SHA: `31deab02507adc49bd296761d1551834e214b768`.

Verified markers:

```text
PRODUCTION_DATABASE_MIGRATION=PASS
PRODUCTION_API_SMOKE=PASS
PRODUCTION_FRONTEND=PASS
PRODUCTION_API_HEALTH=PASS
PRODUCTION_API_READINESS=PASS
PRODUCTION_API_VERSION=PASS
PRODUCTION_CORS=PASS
PRODUCTION_RELEASE_SMOKE=PASS
```

The current frontend is technically deployed and integrated. Visual redesign remains a
separate product phase and is not part of this deployment runbook.

## 1. Freeze the release candidate

```bash
git status -sb
git rev-parse HEAD
pnpm verify:release
pnpm verify:backend-operations-contract
```

Record the source candidate SHA separately from the deployment revision. After Render
selects or completes a deployment, copy the full SHA from Render deployment metadata into
`DEPLOYED_API_REVISION`. Use that explicit value for the API revision check and full
production smoke. Do not infer the deployed revision from the current local `HEAD`; the
repository may advance independently. A migration may be executed from an earlier compatible
SHA when later commits do not change the migration catalog; record every revision explicitly.

## 2. Create Neon PostgreSQL

Create one project and keep two TLS connection strings:

- a direct connection string for migrations and maintenance;
- a pooled connection string for the running API.

Run migrations from a clean checkout with the direct URL:

```bash
PRODUCTION_DATABASE_MIGRATION_URL='paste the direct Neon URL in this terminal only' \
EXPECTED_RELEASE_SHA="$(git rev-parse HEAD)" \
pnpm migrate:production-database
```

The command rejects Neon hostnames containing `-pooler`, requires TLS, and delegates to
the production `cmd/migrate` path. Do not place the direct URL in source control or shared
shell history.

Required successful markers include:

```text
PRODUCTION_DATABASE_MIGRATION_SHA=<full expected SHA>
PRODUCTION_DATABASE_MIGRATION=PASS
MIGRATION_COMMAND_EXIT=0
```

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

During initial Blueprint creation, supply:

```text
DATABASE_URL=<pooled Neon URL>
API_ALLOWED_ORIGINS=<temporary valid HTTPS origin until Vercel is deployed>
API_MUTATION_KEY_SHA256=<owner-generated lowercase SHA-256 digest>
METRICS_KEY_SHA256=<owner-generated lowercase SHA-256 digest>
```

The free Render web service does not support a pre-deploy command. Run
`pnpm migrate:production-database` before the initial deploy and before deploying any
commit that adds a migration.

A paid Render service may instead define a pre-deploy command that runs `/app/migrate`
with a direct database URL. Do not add that command to the free Blueprint because it
would document a capability unavailable on the current plan.

## 4. Deploy and verify the API

Render supplies `RENDER_GIT_COMMIT` during the Docker build. The Dockerfile uses it only
when an explicit `VCS_REF` build argument is absent, preserving local and Continuous
Integration behavior.

After the service is live and `/api/v1/ready` returns `200`, run:

```bash
DEPLOYED_API_REVISION='<full SHA from the intended Render deployment>'
API_BASE_URL=https://global-flight-analytics-api.onrender.com \
EXPECTED_API_REVISION="$DEPLOYED_API_REVISION" \
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

A `404` response for `/` is expected because the production lifecycle routes are under
`/api/v1/`.

## 5. Deploy Next.js on Vercel

Import the monorepo with these settings:

| Setting | Value |
| --- | --- |
| Framework | Next.js |
| Root Directory | `apps/web` |
| Production branch | `main` |
| Environment variable | `NEXT_PUBLIC_API_BASE_URL` |
| Environment value | `https://global-flight-analytics-api.onrender.com` |

Keep default Next.js build, install, and output settings unless the build log identifies a
specific monorepo problem.

After deployment, select the stable project production domain rather than a unique
commit-specific preview URL. The verified production origin is:

```text
https://global-flight-analytics-web.vercel.app
```

## 6. Close exact-origin CORS

In Render, replace the temporary `API_ALLOWED_ORIGINS` value with the exact Vercel
production origin:

```text
https://global-flight-analytics-web.vercel.app
```

The value must contain only the origin:

- no trailing slash;
- no path;
- no query parameters;
- no fragment;
- no `Value:` label;
- no quotation marks.

Save, rebuild, and deploy the Render service. Confirm PostgreSQL connection establishment,
API startup, and repeated `/api/v1/ready status=200` logs.

## 7. Run the full production smoke

```bash
DEPLOYED_API_REVISION='<full SHA from the intended Render deployment>'
FRONTEND_URL="https://global-flight-analytics-web.vercel.app" \
API_BASE_URL="https://global-flight-analytics-api.onrender.com" \
EXPECTED_API_REVISION="$DEPLOYED_API_REVISION" \
pnpm smoke:production
```

Required markers:

```text
PRODUCTION_FRONTEND=PASS
PRODUCTION_API_HEALTH=PASS
PRODUCTION_API_READINESS=PASS
PRODUCTION_API_VERSION=PASS
PRODUCTION_CORS=PASS
PRODUCTION_RELEASE_SMOKE=PASS
```

This command verifies frontend identity, API lifecycle, exact build provenance, and the
exact `Access-Control-Allow-Origin` response together.

## 8. Free-tier operational boundary

The Render free instance can spin down after inactivity and may delay the first request by
approximately fifty seconds or more. The smoke commands use bounded retries so the first
request can wake the service before lifecycle validation. This is a latency limitation,
not a deployment or database-integrity failure.

## 9. Rollback

1. select the previous successful API deployment;
2. retain the latest forward-compatible database schema;
3. set `EXPECTED_API_REVISION` to the rollback revision;
4. rerun `pnpm smoke:api-production`;
5. verify the frontend against the rollback API origin;
6. record the reason, exact revision, and smoke evidence.

## Platform references

- Render Blueprint specification: `https://render.com/docs/blueprint-spec`
- Render free-service limitations: `https://render.com/docs/free`
- Render health checks: `https://render.com/docs/health-checks`
- Neon connection pooling: `https://neon.com/docs/connect/connection-pooling`
- Vercel monorepos: `https://vercel.com/docs/monorepos`
