# Document 29 — Reproducible Docker

Status: Implementation baseline v1.0
Project: Global Flight Analytics

---

## Purpose

This document defines the repeatable container build and local PostgreSQL runtime for the Go backend.

The container baseline is deliberately small:

```text
Go builder image
→ deterministic static Go binaries
→ scratch runtime image
→ non-root user
→ embedded healthcheck binary
→ database migrations copied into the image
```

The default Compose environment is local development infrastructure. It is not a production secret-management mechanism.

---

## Pinned Images

```text
Builder: golang:1.26.5-alpine3.24
PostgreSQL: postgres:16.14-alpine3.24
Runtime: scratch
```

The Go module remains the source of truth for dependencies through `go.mod` and `go.sum`.

---

## Backend Image Contents

The image contains these binaries:

```text
/app/server
/app/healthcheck
/app/migrate
/app/ingest
/app/reconcile
/app/import-airports
/app/verify-airports
```

Database migrations are available at:

```text
/app/migrations
```

The default command is:

```text
/app/server
```

---

## Security Baseline

The runtime image:

```text
uses scratch
runs as user 10001:10001
contains no package manager
contains no interactive shell
uses a read-only filesystem in Compose
drops Linux capabilities in Compose
enables no-new-privileges in Compose
exposes only the API port
```

The certificate authority bundle is copied from the builder image so HTTPS provider calls continue to work.

---

## Validate Configuration

Run from the repository root:

```bash
docker compose config
```

---

## Build the Backend Image

Run from the repository root:

```bash
docker build \
  --pull \
  --file apps/api/Dockerfile \
  --tag global-flight-analytics-api:local \
  .
```

Inspect the configured runtime user:

```bash
docker image inspect \
  global-flight-analytics-api:local \
  --format '{{.Config.User}}'
```

Expected result:

```text
10001:10001
```

---

## Start Local PostgreSQL, Migrations, and API

Run:

```bash
docker compose up \
  --build \
  --detach
```

Compose waits for PostgreSQL health, runs migrations to successful completion, and then starts the API.

Check service state:

```bash
docker compose ps
```

Check API health:

```bash
curl \
  --fail \
  --silent \
  --show-error \
  http://127.0.0.1:8080/api/v1/health
```

---

## Logs

```bash
docker compose logs \
  --follow \
  api
```

---

## Stop the Environment

Preserve PostgreSQL data:

```bash
docker compose down
```

Remove PostgreSQL data and recreate a clean database next time:

```bash
docker compose down \
  --volumes
```

---

## Port Overrides

Use another API host port:

```bash
API_HOST_PORT=18080 \
docker compose up \
  --build \
  --detach
```

Use another PostgreSQL host port:

```bash
POSTGRES_HOST_PORT=15432 \
docker compose up \
  --build \
  --detach
```

These values change host bindings only. Container-to-container ports remain unchanged.

---

## Verification Contract

Every backend Pull Request that changes the container scope must pass:

```text
Docker Compose configuration validation
backend image build
non-root user inspection
container health transition to healthy
HTTP health endpoint smoke test
```

The existing Go tests, static analysis, race checks, and PostgreSQL integration checks remain mandatory.
