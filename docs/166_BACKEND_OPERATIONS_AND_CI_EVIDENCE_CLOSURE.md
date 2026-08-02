# Backend Operations and Continuous Integration Evidence Closure

<!-- RELEASE-TRUTH-DEPLOYMENT-REVISION-V1 -->

Status: Backend operations, production database migration, public API deployment, and production smoke verified
Original product release SHA: `49e474e929dcca5b687464f0a47ce73fcd5a52a7`
Historically verified production application SHA (2026-08-02): `6bca02a8ed1487195b165ae9ced3ca687a373666`
Production migration evidence SHA: `31deab02507adc49bd296761d1551834e214b768`
Backend Continuous Integration: run `30715613342`, completed successfully
Frontend Continuous Integration: run `30715613361`, completed successfully
Production verification date: 2026-08-02

## Purpose

This increment originally closed repository-side operational debt without changing the
Next.js product experience. The owner-controlled Neon, Render, and Vercel actions have now
also been completed and verified. This document records the resulting production evidence
without modifying frontend code or claiming that the current visual design is final.

## Closure state

```text
SOURCE_IMPLEMENTATION=CLOSED
EXACT_COMMIT_CI_EVIDENCE=CLOSED
BACKEND_DEPLOYMENT_SOURCE=CLOSED
PRODUCTION_DATABASE_MIGRATION_WORKFLOW=CLOSED
PRODUCTION_DATABASE_MIGRATION=CLOSED
PRODUCTION_API_SMOKE_CONTRACT=CLOSED
PUBLIC_API_DEPLOYMENT=CLOSED
PUBLIC_NEXTJS_DEPLOYMENT=CLOSED
PRODUCTION_CORS=CLOSED
FULL_BROWSER_PRODUCTION_SMOKE=CLOSED
FRONTEND_VISUAL_REDESIGN=PLANNED_SEPARATE_PHASE
```

## Production endpoints

- API: `https://global-flight-analytics-api.onrender.com`
- Frontend: `https://global-flight-analytics-web.vercel.app`

No credential-bearing value is recorded in this document. Neon URLs, mutation keys,
metrics keys, provider credentials, and platform tokens remain in owner-controlled secret
stores.

## Render free-plan boundary

`render.yaml` defines one free Docker web service in Frankfurt, deploys only after linked
checks pass, reads secrets from the Render Dashboard, and uses `/api/v1/ready` as the
health check. The Docker build keeps explicit `VCS_REF` precedence and falls back to
Render's `RENDER_GIT_COMMIT`, so `/api/v1/version` exposes the deployed revision.

A free Render web service does not support the paid pre-deploy command. Production
migrations therefore remain explicit and separate: the owner runs
`pnpm migrate:production-database` against the direct Neon URL before the first deploy and
before deploying commits that add migrations. The long-running API uses the pooled Neon
URL.

The current free instance can spin down after inactivity. Bounded smoke retries account for
cold starts; the limitation affects initial latency rather than the revision, readiness, or
CORS contracts.

## Production database evidence

Migrations `019` through `029` were applied successfully through the direct TLS Neon
connection workflow. Recorded evidence:

```text
PRODUCTION_DATABASE_MIGRATION_SHA=31deab02507adc49bd296761d1551834e214b768
PRODUCTION_DATABASE_MIGRATION=PASS
MIGRATION_COMMAND_EXIT=0
```

An operator backup created during the migration repair remains retained. This document does
not expose its contents or any database credential.

## Production API evidence

During the production verification performed on 2026-08-02, Render served
application revision:

`6bca02a8ed1487195b165ae9ced3ca687a373666`

The stable public API alias is mutable. This is evidence for the recorded verification
event and is not a perpetual assertion about the revision currently served by that alias.

The API established its PostgreSQL connection, started on the Render-assigned port, and
returned repeated `200` responses from `/api/v1/ready`.

The API-only production smoke verified liveness, PostgreSQL-backed readiness, version
fields, and exact build provenance:

```text
PRODUCTION_API_HEALTH=PASS
PRODUCTION_API_READINESS=PASS
PRODUCTION_API_VERSION=PASS
PRODUCTION_API_REVISION=PASS
PRODUCTION_API_SMOKE=PASS
```

## Browser and CORS evidence

Vercel deployed the Next.js application from `apps/web` with
`NEXT_PUBLIC_API_BASE_URL=https://global-flight-analytics-api.onrender.com`.

Render `API_ALLOWED_ORIGINS` was set to the exact stable Vercel production origin:

```text
https://global-flight-analytics-web.vercel.app
```

The full production smoke passed:

```text
PRODUCTION_FRONTEND=PASS
PRODUCTION_API_HEALTH=PASS
PRODUCTION_API_READINESS=PASS
PRODUCTION_API_VERSION=PASS
PRODUCTION_CORS=PASS
PRODUCTION_RELEASE_SMOKE=PASS
```

## Frontend scope boundary

The frontend is publicly deployed and technically integrated, but its visual and
interaction design still requires a separate redesign phase. No frontend source file is
changed by this evidence closure, and the redesign is not represented as unresolved
backend or infrastructure debt.

## Evidence-only attestation

Production claims in this document are limited to the exact public URLs, recorded
revisions, migration markers, lifecycle responses, frontend identity, and CORS smoke that
were observed on 2026-08-02. Future application or infrastructure changes require new
revision-specific evidence.
