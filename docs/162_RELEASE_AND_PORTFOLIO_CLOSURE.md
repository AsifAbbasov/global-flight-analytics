# Release and Portfolio Closure

<!-- RELEASE-TRUTH-DEPLOYMENT-REVISION-V1 -->

Status: V1.0.0 RELEASE PUBLISHED
Original source release SHA: `49e474e929dcca5b687464f0a47ce73fcd5a52a7`
Historically verified production application SHA (2026-08-02): `6bca02a8ed1487195b165ae9ced3ca687a373666`
Production migration evidence SHA: `31deab02507adc49bd296761d1551834e214b768`
Backend Continuous Integration: run `30715613342`, completed successfully
Frontend Continuous Integration: run `30715613361`, completed successfully
Original release-closure baseline: `03ac45dc2a515c77af8d992aa6489816f1cbe927`
Production verification date: 2026-08-02
Final exact-production validation date: 2026-09-06
Final validated production revision: `366a406bc33c7deb839d4c1a56feb82901402e75`
Render deployment: `dep-daegc2ss728c7380js90` — live
Production Smoke run: `34016471540`
Production Smoke job: `101441002735`
Published release tag: `v1.0.0`
Published release source SHA: `cc962c7c84b84d8e9b9b1306f65f054c6e0c4d70`
Published release URL: `https://github.com/AsifAbbasov/global-flight-analytics/releases/tag/v1.0.0`
Published release date: 2026-09-06

## Purpose

This closure records the implemented system, exact source and Continuous Integration
evidence, the verified public production deployment, and the published `v1.0.0` release
without hiding open-data, free-tier, or visual-design limitations.

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
FINAL_EXACT_PRODUCTION_VALIDATION=CLOSED
FINAL_RELEASE_DOCUMENTATION=CLOSED
V1_RELEASE_TAG=v1.0.0
V1_RELEASE_SHA=cc962c7c84b84d8e9b9b1306f65f054c6e0c4d70
V1_RELEASE=CLOSED
```

`FRONTEND_VISUAL_REDESIGN=PLANNED_SEPARATE_PHASE` is retained above as historical state from
the original August release closure. The later frontend redesign and Visual Polish V2 were
completed and closed in their own canonical evidence documents; this file does not rewrite
that historical release snapshot.

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

## Final exact-production validation — 2026-09-06

The final release candidate was validated from hosting metadata rather than inferred from a
local checkout. Render reported deployment `dep-daegc2ss728c7380js90` as `live` for exact
commit:

`366a406bc33c7deb839d4c1a56feb82901402e75`

GitHub Actions Production Smoke run `34016471540`, job `101441002735`, was then dispatched
explicitly with that same full SHA as `EXPECTED_API_REVISION`. The workflow checked out the
same `main` revision and produced:

```text
PRODUCTION_SMOKE_EXPECTED_REVISION=366a406bc33c7deb839d4c1a56feb82901402e75
PRODUCTION_SMOKE_REVISION_INPUT=PASS
PRODUCTION_FRONTEND=PASS
PRODUCTION_API_HEALTH=PASS
PRODUCTION_API_READINESS=PASS
PRODUCTION_API_VERSION=PASS
PRODUCTION_CORS=PASS
PRODUCTION_RELEASE_SMOKE=PASS
PRODUCTION_SMOKE_EVENT=workflow_dispatch
SCHEDULED_PRODUCTION_SMOKE=PASS
```

This closes the final exact-production validation boundary for the release candidate. It
proves the public frontend identity, API lifecycle, exact runtime revision and production
CORS contract together. It does not make the mutable public aliases immutable and does not
claim that production ingestion was permanently enabled; the FREE_V1 ingestion path remains
intentionally fail-closed after its separately recorded controlled validation.

The owner-side repository governance verifier was also executed against exact revision
`366a406bc33c7deb839d4c1a56feb82901402e75` and returned:

```text
DEPENDABOT_ALERTS_API=PASS
SECRET_SCANNING_ALERTS_API=PASS
CODEQL_ANALYSIS=PASS
MAIN_RULESET=PASS
ACTIONS_POLICY=PASS
REPOSITORY_GOVERNANCE_SETTINGS=PASS
```

Those markers are supporting release evidence for the controls the verifier actually checks.
They do not silently reclassify any separate canonical finding whose closure criteria extend
beyond that verifier.

## Published v1.0.0 release — 2026-09-06

After the final release-candidate documentation was merged, `main` resolved to exact revision
`cc962c7c84b84d8e9b9b1306f65f054c6e0c4d70`. Post-merge Backend CI #737, Frontend CI #396,
CodeQL #376 and Playwright E2E #173 all completed successfully on that revision, and the
Vercel deployment status was successful.

GitHub Release `v1.0.0` was then published from that exact post-merge revision as a full
release with `draft=false` and `prerelease=false`:

`https://github.com/AsifAbbasov/global-flight-analytics/releases/tag/v1.0.0`

The release tag is source evidence for the published portfolio revision. It does not imply
that mutable production aliases permanently serve the same source revision as the tag; the
runtime revision remains governed by the explicit deployment-revision validation policy.

## Release truth and evidence freshness

Source revision, intended deployment revision, observed runtime revision, current repository
`HEAD`, and a published release tag are independent facts. Future deployment verification
must obtain the intended revision from Render deployment metadata, pass it explicitly as
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
7. browser-to-API production smoke;
8. final exact-production revision validation;
9. final release documentation;
10. publication of the `v1.0.0` tag/release.

States 1–10 are closed. Publication is recorded only after GitHub exposed the real release
for tag `v1.0.0` targeting exact revision
`cc962c7c84b84d8e9b9b1306f65f054c6e0c4d70`.

## Free-tier operational boundary

The Render API currently uses a free instance. It can spin down after inactivity and may
require a cold start before the first request succeeds. The production smoke scripts use
bounded retries for that operational boundary. This limitation affects initial latency,
not the recorded application revision, PostgreSQL readiness contract, or CORS policy.

## Evidence policy

Placeholders, guessed run identifiers, screenshots from another commit, unverified URLs,
and secret-bearing connection strings are prohibited release evidence. Public deployment
is recorded only because the exact URLs, API revision, readiness, frontend identity, and
CORS behavior were verified together. Release publication is recorded only because GitHub
returned the real `v1.0.0` release with the expected exact target revision.

## Release publication closure

Final exact-production validation, final release documentation, post-merge CI, and
`v1.0.0` publication are closed. The current repository truth therefore records
`V1_RELEASE=CLOSED`. Further work belongs to post-v1 product development rather than the
v1 release-candidate closure sequence.

## Scope boundary

This closure does not add authentication, billing, paid infrastructure, Kubernetes,
microservices, machine learning, satellite fusion, safety certification, proprietary
aviation feeds, or a custom domain.