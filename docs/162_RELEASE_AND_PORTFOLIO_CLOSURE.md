# Release and Portfolio Closure

<!-- RELEASE-TRUTH-DEPLOYMENT-REVISION-V1 -->

Status: Public production deployment verified; frontend visual redesign remains a separate product phase
Original source release SHA: `49e474e929dcca5b687464f0a47ce73fcd5a52a7`
Historically verified production application SHA (2026-08-02): `6bca02a8ed1487195b165ae9ced3ca687a373666`
Production migration evidence SHA: `31deab02507adc49bd296761d1551834e214b768`
Backend Continuous Integration: run `30715613342`, completed successfully
Frontend Continuous Integration: run `30715613361`, completed successfully
Original release-closure baseline: `03ac45dc2a515c77af8d992aa6489816f1cbe927`
Production verification date: 2026-08-02

## Purpose

This closure records the implemented system, exact source and Continuous Integration
evidence, and the verified public production deployment without hiding open-data,
free-tier, or visual-design limitations.

## Release state

```text
SOURCE_IMPLEMENTATION=CLOSED
BACKEND_CI=CLOSED
FRONTEND_CI=CLOSED
EXACT_COMMIT_CI_EVIDENCE=CLOSED
PRODUCTION_DATABASE_MIGRATION=CLOSED
PUBLIC_API_DEPLOYMENT=CLOSED
PUBLIC_NEXTJS_DEPLOYMENT=CLOSED
PRODUCTION_CORS=CLOSED
FULL_BROWSER_PRODUCTION_SMOKE=CLOSED
FRONTEND_VISUAL_REDESIGN=PLANNED_SEPARATE_PHASE
```

## Public production endpoints

- Frontend: `https://global-flight-analytics-web.vercel.app`
- API: `https://global-flight-analytics-api.onrender.com`
- Database: owner-controlled Neon PostgreSQL in the Frankfurt region

The public URLs contain no credentials. Database connection strings, mutation keys,
metrics keys, deployment tokens, and provider credentials remain exclusively in platform
secret stores.

## Exact evidence

The original portfolio source release is:

`49e474e929dcca5b687464f0a47ce73fcd5a52a7`

GitHub Actions evidence for that exact SHA:

| Workflow | Run identifier | Event | Conclusion |
| --- | ---: | --- | --- |
| Backend CI | `30715613342` | push | success |
| Frontend CI | `30715613361` | push | success |

Backend CI completed Backend Race Safety, Backend Quality, PostgreSQL 16 Integration and
Backend Container successfully. Frontend CI completed release contracts, dependency
security, production dependency audit, ESLint, TypeScript, eighty-two frontend tests and
the production build successfully.

Production migrations `019` through `029` were applied through the direct TLS Neon
connection workflow with evidence SHA
`31deab02507adc49bd296761d1551834e214b768`. The running API uses the pooled Neon
connection string.

During the production verification performed on 2026-08-02, Render and Vercel
served application revision `6bca02a8ed1487195b165ae9ced3ca687a373666`. The API version endpoint reported that
same revision during that revision-specific smoke test. The public aliases are mutable,
so this document does not assert that they continue to serve the historical SHA.

## Release truth and evidence freshness

Source revision, intended deployment revision, observed runtime revision, and current
repository `HEAD` are independent facts. Future deployment verification must obtain the
intended revision from Render deployment metadata, pass it explicitly as
`EXPECTED_API_REVISION`, and compare it with `/api/v1/version`. A local `git rev-parse HEAD`
must never be substituted automatically unless that exact commit is the deployment being
verified.

## Verified production smoke

The complete browser-to-API release smoke passed with the exact public frontend and API
origins:

```text
PRODUCTION_FRONTEND=PASS
PRODUCTION_API_HEALTH=PASS
PRODUCTION_API_READINESS=PASS
PRODUCTION_API_VERSION=PASS
PRODUCTION_CORS=PASS
PRODUCTION_RELEASE_SMOKE=PASS
```

The API-only lifecycle and revision smoke also passed before the full browser and CORS
verification.

## Independent release states

The release keeps separate evidence for:

1. source implementation;
2. exact-commit Continuous Integration;
3. production database migration;
4. public API deployment;
5. public Next.js deployment;
6. exact-origin CORS behavior;
7. browser-to-API production smoke.

All seven states are closed for the deployment recorded above. Future visual redesign
work does not invalidate this production evidence and must remain a separate product
increment with its own verification.

## Free-tier operational boundary

The Render API currently uses a free instance. It can spin down after inactivity and may
require a cold start before the first request succeeds. The production smoke scripts use
bounded retries for that operational boundary. This limitation affects initial latency,
not the recorded application revision, PostgreSQL readiness contract, or CORS policy.

## Evidence policy

Placeholders, guessed run identifiers, screenshots from another commit, unverified URLs,
and secret-bearing connection strings are prohibited release evidence. Public deployment
is recorded only because the exact URLs, API revision, readiness, frontend identity, and
CORS behavior were verified together.

## Remaining product phase

The current frontend is publicly deployed and technically integrated. A substantial visual
and interaction redesign remains planned as a separate frontend phase. This closure does
not claim that the present interface is the final design and does not modify frontend code.

## Scope boundary

This closure does not add authentication, billing, paid infrastructure, Kubernetes,
microservices, machine learning, satellite fusion, safety certification, proprietary
aviation feeds, or a custom domain.
