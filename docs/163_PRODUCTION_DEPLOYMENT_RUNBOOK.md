# Production Deployment Runbook

## Purpose

This runbook deploys the existing architecture without introducing a second application
stack: Neon PostgreSQL, the existing Dockerized Go API on a Render-compatible web service,
and the existing Next.js application on Vercel.

Do not commit database passwords, mutation keys, provider credentials or deployment
tokens. Configure them in the platform secret stores.

## 1. Freeze the release commit

From a clean `main` branch:

```bash
git status -sb
git rev-parse HEAD
pnpm verify:release
```

Record the full SHA. Use the same value for the API build revision and the production
smoke command.

## 2. Create Neon PostgreSQL

Create one production project and database. Keep two connection strings:

- use a direct connection string for migrations and other session-dependent maintenance;
- use a pooled connection string for the running API.

Both strings must require TLS. Store them outside the repository.

Build the release image from the repository root and run migrations with the direct URL:

```bash
export RELEASE_SHA="$(git rev-parse HEAD)"
export RELEASE_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
export NEON_DIRECT_DATABASE_URL='paste the direct Neon URL in your terminal only'

docker build \
  --build-arg APP_VERSION="portfolio-${RELEASE_SHA}" \
  --build-arg VCS_REF="$RELEASE_SHA" \
  --build-arg BUILD_DATE="$RELEASE_DATE" \
  --file apps/api/Dockerfile \
  --tag global-flight-analytics-api:"$RELEASE_SHA" \
  .

docker run --rm \
  --env DATABASE_URL="$NEON_DIRECT_DATABASE_URL" \
  --env MIGRATIONS_DIR=/app/migrations \
  --env MIGRATION_TIMEOUT=2m \
  global-flight-analytics-api:"$RELEASE_SHA" \
  /app/migrate
```

Do not run schema migrations through the pooled URL.

## 3. Deploy the API

Create a Docker web service from the repository.

Use these settings:

| Setting | Value |
| --- | --- |
| Dockerfile | `apps/api/Dockerfile` |
| Docker build context | repository root |
| Health check path | `/api/v1/ready` |
| Runtime port | `API_PORT=10000` |
| Runtime database | pooled Neon connection string |
| Auto deploy | only from the intended release branch |

Required environment values:

```text
DATABASE_URL=<pooled Neon URL>
DATABASE_CONNECT_TIMEOUT=5s
API_PORT=10000
API_ALLOWED_ORIGINS=<final Vercel HTTPS origin>
API_BODY_LIMIT_BYTES=1048576
API_READ_TIMEOUT=10s
API_WRITE_TIMEOUT=15s
API_IDLE_TIMEOUT=60s
API_RATE_LIMIT_MAX=120
API_RATE_LIMIT_WINDOW=1m
API_MUTATION_KEY_SHA256=<owner-generated lowercase SHA-256 digest>
MIGRATIONS_DIR=/app/migrations
MIGRATION_TIMEOUT=2m
TRAFFIC_PROVIDER=airplanes.live
```

Add optional OpenSky credentials only when the provider is intentionally enabled. Keep
the raw mutation key outside the service; the backend stores only its digest.

The API must bind to `0.0.0.0` through the configured port. After deployment, verify:

```bash
curl --fail --silent --show-error https://your-api.example/api/v1/health
curl --fail --silent --show-error https://your-api.example/api/v1/ready
curl --fail --silent --show-error https://your-api.example/api/v1/version
```

## 4. Deploy the frontend

Import the same repository into Vercel and configure:

| Setting | Value |
| --- | --- |
| Framework | Next.js |
| Root Directory | `apps/web` |
| Production branch | the intended release branch |
| Environment | `NEXT_PUBLIC_API_BASE_URL=https://your-api.example` |

The API base URL is embedded in the browser build. Changing it requires a new frontend
deployment.

After Vercel assigns the final production origin, update `API_ALLOWED_ORIGINS` on the API
service to that exact HTTPS origin and redeploy the API. Do not use `*` for production
CORS.

## 5. Run the production smoke contract

```bash
FRONTEND_URL=https://your-frontend.example \
API_BASE_URL=https://your-api.example \
EXPECTED_API_REVISION="$(git rev-parse HEAD)" \
pnpm smoke:production
```

The smoke command verifies:

- frontend HTML and product identity;
- API liveness;
- PostgreSQL-backed readiness;
- version and build-provenance fields;
- optional exact API revision equality;
- the exact frontend CORS origin.

## 6. Rollback

A safe rollback uses a previously green commit and its matching database-compatible
image. Do not roll application code behind irreversible schema changes without reviewing
the migration history.

1. select the previous successful API deployment;
2. select the corresponding frontend deployment;
3. keep the database at the latest forward-compatible schema;
4. rerun `pnpm smoke:production` with the revision expected for the rolled-back API;
5. record the rollback reason and observed evidence.

## Platform references

- Vercel monorepo documentation: `https://vercel.com/docs/monorepos`
- Render Docker services: `https://render.com/docs/docker`
- Render web services and ports: `https://render.com/docs/web-services`
- Render health checks: `https://render.com/docs/health-checks`
- Neon connection pooling: `https://neon.com/docs/connect/connection-pooling`
