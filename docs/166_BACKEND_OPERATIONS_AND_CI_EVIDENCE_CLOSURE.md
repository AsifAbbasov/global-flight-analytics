# Backend Operations and Continuous Integration Evidence Closure

Status: Repository-side backend operations prepared; exact product release Continuous Integration closed
Product release SHA: `49e474e929dcca5b687464f0a47ce73fcd5a52a7`
Backend Continuous Integration: run `30715613342`, completed successfully
Frontend Continuous Integration: run `30715613361`, completed successfully
Date: 2026-08-01

## Purpose

This increment closes the remaining repository-side operational debt without changing the
Next.js product experience. It combines exact Continuous Integration evidence, a free-plan
Render Blueprint for the Dockerized Go API, a direct-database migration command, an
API-only production smoke contract, permanent tests and deployment documentation.

## Closure state

```text
SOURCE_IMPLEMENTATION=CLOSED
EXACT_COMMIT_CI_EVIDENCE=CLOSED
BACKEND_DEPLOYMENT_SOURCE=CLOSED
PRODUCTION_DATABASE_MIGRATION_WORKFLOW=CLOSED
PRODUCTION_API_SMOKE_CONTRACT=CLOSED
PUBLIC_API_DEPLOYMENT=PENDING_OWNER_CREDENTIALS
NEXTJS_CREATIVE_PHASE=DEFERRED_BY_OWNER
```

The pending cloud action is not an unidentified engineering debt. It requires the
owner's Neon and Render accounts, secret values and final public URL.

## Render free-plan boundary

`render.yaml` defines one free Docker web service in Frankfurt, deploys only after linked
checks pass, reads secrets from the Render Dashboard and uses `/api/v1/ready` as the
health check. The Docker build keeps explicit `VCS_REF` precedence and falls back to Render's `RENDER_GIT_COMMIT`, so `/api/v1/version` exposes the deployed revision.

A free Render web service does not support the paid pre-deploy command. The repository
therefore keeps production migrations explicit and separate: the owner runs
`pnpm migrate:production-database` against the direct Neon URL before the first deploy
and before deploying commits that add migrations. The long-running API uses the pooled
Neon URL.

## Production API evidence

After the service is public, the owner runs:

```bash
API_BASE_URL=https://your-api.example \
EXPECTED_API_REVISION="$(git rev-parse HEAD)" \
pnpm smoke:api-production
```

The command verifies API liveness, PostgreSQL-backed readiness, version fields and exact
build revision. CORS and browser integration remain part of the later Next.js creative
and public deployment phase.

## Deferred frontend phase

The Next.js visual and public deployment phase is deliberately deferred by the owner
because it requires a separate creative pass. No frontend code is changed by this
increment, and the deferred phase is not represented as an unresolved backend debt.

## Final evidence-only attestation

After this operational source increment is committed and its Backend and Frontend
workflows pass, one small documentation-only attestation may record that increment's
exact SHA and workflow run identifiers. That attestation does not reopen product scope.
