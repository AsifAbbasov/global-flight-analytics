# Stage 23 — ETA Reliability Intelligence

Status: `IN_PROGRESS`
Project: Global Flight Analytics
Stage: 23
Feature PR: #171
Branch: `feat/stage-23-eta-reliability`
Base canonical main at stage start: `9e806f2f44c24d69f695ae37ab02a0f5a0df94f3`
Additional paid cost: `0 RUB`

<!-- STAGE-23-ETA-RELIABILITY:DOCUMENT -->

## 1. Product objective

Stage 23 adds one user-facing answer to the existing Aircraft Detail / Projection Intelligence workflow:

> How reliable have comparable ETA estimates been on persisted historical observations?

The stage is intentionally a vertical product slice rather than an analytics-only backend project. The required path is:

```text
persisted trajectory observations
        ↓
existing Projection Intelligence historical recomputation
        ↓
bounded ETA reliability aggregation
        ↓
read-only HTTP contract
        ↓
TypeScript client + TanStack Query
        ↓
Aircraft Detail / Estimated Arrival UI
```

A Stage 23 backend capability is not considered product-complete unless the same evidence is consumable and visible in the existing frontend ETA workflow.

## 2. User-facing surface

Stage 23 extends the existing `Projection and Estimated Arrival` panel with a `Historical ETA Reliability` section placed next to the current estimated arrival rather than creating a separate analytics dashboard.

The intended user-facing evidence includes:

- eligible historical sample count;
- median absolute ETA error;
- p80 absolute ETA error;
- share of eligible samples within five minutes;
- share of eligible samples within ten minutes;
- published ETA interval coverage against the historical endpoint proxy;
- explicit evidence status and limitations.

If eligible evidence is insufficient, the UI must show an unavailable or limited evidence state instead of manufacturing a reliability percentage.

## 3. Evidence classification

```text
STAGE_23_EVIDENCE_CLASS=HISTORICALLY_RECOMPUTED_FROM_PERSISTED_OBSERVATIONS
STAGE_23_OFFICIAL_ARRIVAL_TRUTH=NONE
STAGE_23_HISTORICAL_ARRIVAL_EVIDENCE=PERSISTED_ENDPOINT_PROXY
STAGE_23_TOUCHDOWN_TIME_CLAIM=NONE
STAGE_23_GATE_TIME_CLAIM=NONE
STAGE_23_SCHEDULE_ACCURACY_CLAIM=NONE
STAGE_23_OPERATIONAL_GUIDANCE=NONE
```

The project does not own an official airline, airport, gate, touchdown or certified arrival timestamp source for this feature.

Historical arrival evidence is therefore deliberately weaker and is represented as a persisted endpoint proxy: the last qualifying persisted trajectory observation close to the inferred destination airport under the bounded Stage 23 policy.

The endpoint proxy must never be relabeled as:

- actual touchdown time;
- actual gate-arrival time;
- official arrival time;
- schedule truth;
- ATC truth;
- operational flight-status truth.

## 4. Why the feature is useful

The current Projection Intelligence product can publish an ETA and a confidence score. Those values alone do not tell a user how similar estimates behaved against historical observed evidence.

Stage 23 converts existing historical projection/evaluation primitives into directly understandable evidence next to the ETA itself.

The product question changes from:

```text
What is the estimated arrival?
```

into:

```text
What is the estimated arrival, and how large were historical errors for comparable bounded estimates?
```

This is intended to improve interpretability, not to make a certified accuracy claim.

## 5. Zero-cost boundary

```text
STAGE_23_ADDITIONAL_COST=0_RUB
STAGE_23_NEW_PAID_PROVIDER=NONE
STAGE_23_NEW_EXTERNAL_API=NONE
STAGE_23_NEW_DATABASE_TABLE=NONE
STAGE_23_NEW_MIGRATION=NONE
STAGE_23_NEW_SERVER=NONE
STAGE_23_NEW_BACKGROUND_MATERIALIZER=NONE
STAGE_23_NEW_REDIS=NONE
STAGE_23_NEW_KAFKA=NONE
STAGE_23_NEW_GPU=NONE
```

Stage 23 reuses:

- the existing Go backend;
- the existing PostgreSQL / Neon data path;
- existing persisted trajectories;
- existing Projection Intelligence historical `as_of_time` support;
- the existing Projection Evaluation package;
- the existing Next.js frontend;
- existing TanStack Query infrastructure.

No paid aviation provider is introduced.

## 6. Bounded computation policy

The Stage 23 service is on-demand and bounded.

```text
STAGE_23_MAX_HISTORICAL_CANDIDATES=8
STAGE_23_FRONTEND_POLLING=NONE
STAGE_23_QUERY_STALE_TIME=5_MINUTES
STAGE_23_WINDOW_FOCUS_REFETCH=DISABLED
```

The feature must not create an unbounded replay loop over repository history for each frontend render.

Historical candidates are capped by policy. This protects the zero-budget Render/Neon profile and keeps the feature compatible with the project's free-tier operating model.

## 7. Historical recomputation semantics

For each eligible historical candidate, Stage 23 uses a real persisted historical analytical timestamp and invokes the existing production Projection Intelligence implementation at that historical `as_of_time`.

The feature does not invent synthetic observation timestamps and does not claim that historical forecast outputs were persisted when they were not.

```text
STAGE_23_PERSISTED_FORECAST_HISTORY=NONE
STAGE_23_SYNTHETIC_AS_OF_TIME=NONE
STAGE_23_HISTORICAL_RECOMPUTATION=CURRENT_PRODUCTION_IMPLEMENTATION
```

This means Stage 23 answers a bounded historical-recomputation question, not a stored-forecast-history question.

## 8. Method comparability

ETA error samples are only meaningful when the historical projection produced compatible arrival evidence under the current bounded evaluation policy.

The service records and exposes evidence status rather than silently mixing incompatible or unavailable historical evaluations.

A sample that cannot provide an eligible predicted arrival and endpoint proxy does not become a fabricated zero-error or carry-forward observation.

## 9. Metrics

The product metrics are descriptive historical aggregates.

### 9.1 Median absolute ETA error

Median of eligible absolute ETA errors in seconds.

### 9.2 P80 absolute ETA error

80th percentile of eligible absolute ETA errors in seconds.

### 9.3 Five-minute hit ratio

Fraction of eligible samples where absolute ETA error is less than or equal to five minutes.

### 9.4 Ten-minute hit ratio

Fraction of eligible samples where absolute ETA error is less than or equal to ten minutes.

### 9.5 ETA interval coverage

Fraction of eligible samples where the historical endpoint proxy falls inside the published ETA earliest/latest interval.

These metrics are not airport punctuality, airline on-time-performance or schedule adherence metrics.

## 10. Evidence status

The Stage 23 response distinguishes insufficient evidence from usable evidence.

Expected status vocabulary:

```text
unavailable
limited
complete
```

The exact response must always disclose sample count and limitations. A visually attractive percentage is not allowed to replace insufficient evidence.

## 11. Backend implementation

The Stage 23 backend package is:

```text
apps/api/internal/projectionintelligence/etareliability/
```

Current implementation files include:

```text
model.go
policy.go
service.go
service_test.go
```

The service composes existing Projection Intelligence and Projection Evaluation behavior instead of introducing a second forecasting engine.

## 12. HTTP contract

The intended production read surface is:

```text
GET /api/v1/trajectories/{id}/eta-reliability
```

The endpoint is read-only.

The public contract must be added to both canonical OpenAPI copies and the generated TypeScript client before Stage 23 can become review-ready.

At the current Stage 23 documentation state, that OpenAPI synchronization remains incomplete and therefore Stage 23 remains `IN_PROGRESS`.

## 13. Frontend integration

Current Stage 23 frontend files include:

```text
apps/web/types/eta-reliability.ts
apps/web/lib/api/eta-reliability.ts
apps/web/lib/queries/eta-reliability.ts
apps/web/components/aircraft/eta-reliability-panel.tsx
apps/web/components/aircraft/projection-intelligence-panel.tsx
```

The query policy deliberately avoids live polling:

```text
refetchInterval=false
refetchOnWindowFocus=false
staleTime=5 minutes
```

This is both a product and free-tier decision: historical reliability does not need minute-by-minute recomputation.

## 14. Frontend claim boundary

The frontend must communicate that:

- reliability is historical and evidence-bounded;
- sample size matters;
- endpoint evidence is a persisted observation proxy;
- no official arrival event is available;
- the result is research intelligence, not operational guidance.

The frontend must not say or imply:

```text
ETA accuracy guaranteed
actual landing time
official arrival accuracy
ATC-grade reliability
safe operational ETA
```

## 15. Validation history — rejected intermediate heads

Validation is SHA-specific. Earlier failures remain historical evidence and must not be relabeled as passing after later remediation.

### 15.1 Head `0d1f49abe255459a45dac67bbdd7e3d0fd117106`

PR #171 was opened as a draft on this head to use CI as an early compile/contract gate while Stage 23 was still explicitly incomplete.

Observed workflow results included:

```text
Frontend CI #513 / run 34119546928 = SUCCESS
OpenAPI Contract #142 / run 34119546967 = FAILURE
API Load Baseline #370 / run 34119546975 = FAILURE
```

The OpenAPI failure was expected contract drift because the new source route had not yet been added to the canonical OpenAPI surface.

API Load Baseline did not fail a performance threshold. Its backend Docker build exposed a real Go compile contract defect:

```text
dto.ETAReliabilityResponse did not satisfy response.SuccessPayload
```

That defect was remediated by explicitly adding the new response DTO to the backend success-payload union.

### 15.2 Head `a9710c428e8988f4099b51b55702fe56cc203444`

After the success-payload remediation, the next observed workflow matrix was:

```text
Frontend CI #514 / run 34119801776 = SUCCESS
API Load Baseline #371 / run 34119801694 = SUCCESS
Playwright E2E #285 / run 34119801740 = SUCCESS
CodeQL #495 / run 34119801725 = SUCCESS
Backend CI #852 / run 34119801824 = FAILURE
OpenAPI Contract #143 / run 34119801804 = FAILURE
```

Backend #852 failed at the `Verify Go formatting` gate before Go tests/vet/audits could run. The unformatted files were reported as:

```text
internal/http/dto/eta_reliability.go
internal/projectionintelligence/etareliability/model.go
internal/projectionintelligence/etareliability/policy.go
internal/projectionintelligence/etareliability/service.go
internal/projectionintelligence/etareliability/service_test.go
internal/server/eta_reliability_runtime.go
internal/server/projection_database_composition.go
```

PostgreSQL 16 Integration and Backend Race Safety on the same workflow completed successfully, but the aggregate Backend CI result remains `FAILURE` and must not be reported as pass.

OpenAPI #143 remains a genuine release blocker until the canonical OpenAPI route/schema inventory and generated client are synchronized to the new read surface.

## 16. Current validation status

The canonical documentation surfaces are now aligned to the in-progress Stage 23 state, but implementation validation is not complete:

```text
STAGE_23_STATUS=IN_PROGRESS
STAGE_23_PR=171
STAGE_23_PR_DRAFT=YES
STAGE_23_REVIEW_READY=NO
STAGE_23_MERGE_READY=NO
STAGE_23_MERGE_AUTHORIZATION=NOT_GRANTED
STAGE_23_DOCUMENT_214=ALIGNED_IN_PROGRESS
STAGE_23_DOCUMENT_INDEX=ALIGNED_V2_4
STAGE_23_README=ALIGNED_IN_PROGRESS
STAGE_23_ROADMAP=ALIGNED_V1_3
STAGE_23_DOCUMENTATION_REGRESSION_TEST=INSTALLED
STAGE_23_EXACT_HEAD_FINAL_CI=NOT_YET_AVAILABLE
STAGE_23_POST_MERGE_CI=NOT_APPLICABLE
```

No earlier successful workflow may be transferred to a later Stage 23 head.

## 17. Documentation and governance alignment

Stage 23 is registered in all high-level documentation surfaces needed for the in-progress feature:

```text
docs/214_STAGE_23_ETA_RELIABILITY_INTELLIGENCE.md
docs/DOCUMENT_INDEX.md — Documentation Index v2.4
README.md — Stage 23 in-progress product summary
docs/24_MVP_VERSION_ROADMAP.md — Architecture Baseline v1.3 / explicit Stage 23 product increment
apps/web/tests/stage-23-eta-reliability-documentation.test.mjs
```

Document 24 also repairs the stale Version 2 candidate language and records the already verified Version 2 closure from Document 213.

The Version 2 reconciliation subsection of Document 25 was written before Stage 23 existed and states that the reconciliation itself did not create Stage 23. That sentence remains historically true about the reconciliation operation. The later explicit product decision that authorizes Stage 23 is owned by Document 24 Section 21 and this Document 214; it must not be misread as a retroactive claim that Version 2 reconciliation created Stage 23.

No synthetic finding ID is created merely because a product feature exists. `GFA-SEC-445` remains the independent existing repository-security finding and is not modified by Stage 23.

## 18. Remaining engineering work before review-ready

The stage remains blocked on all of the following:

1. run `gofmt`-equivalent formatting over every changed Go file and revalidate;
2. synchronize canonical OpenAPI public route/schema contract;
3. keep root and embedded OpenAPI byte-identical;
4. regenerate the canonical TypeScript API client from OpenAPI;
5. update route inventory/count assertions from the previous surface to the new source-backed surface;
6. add dedicated production E2E/mock assertions for the ETA Reliability user path rather than relying only on unrelated Playwright success;
7. complete a full exact-head GitHub CI matrix and Vercel check;
8. verify review threads/reviews and mergeability on the exact final head.

The Stage 23 documentation/claim-boundary regression test, README alignment, Document Index registration and roadmap registration are already present and are no longer listed as unfinished engineering work.

## 19. Review-ready gate

Stage 23 may become review-ready only when all of the following are true on one exact PR head:

```text
Backend CI=SUCCESS
Frontend CI=SUCCESS
OpenAPI Contract=SUCCESS
API Load Baseline=SUCCESS
CodeQL=SUCCESS
Playwright E2E=SUCCESS
Vercel=SUCCESS
Review threads=0 unresolved
OpenAPI/source inventory aligned
Stage 23 permanent regression tests=PASS
Document 214 current-state evidence aligned
```

## 20. Merge gate

Review-ready is not merge authorization.

Merge requires a separate explicit exact-head authorization after the final review-ready SHA is known.

If the head moves, prior authorization and prior exact-head validation are not transferred.

## 21. Closure gate

Stage 23 must not be declared `CLOSED` merely because PR #171 merges.

Formal closure requires:

1. exact-head authorized merge;
2. resulting canonical `main` SHA identified;
3. independent post-merge workflows evaluated on that canonical SHA;
4. Vercel status evaluated on that canonical SHA where applicable;
5. non-triggered workflows reported as `NOT_TRIGGERED`, never as pass;
6. post-merge closure evidence recorded separately without rewriting this pre-merge history.

## 22. Scaling path

Stage 23 is intentionally reusable.

If the trajectory-level product proves useful, the same bounded evidence semantics can later support:

- airport-level ETA reliability summaries;
- route-level ETA reliability summaries;
- reliability comparison by forecast horizon;
- method comparison where evidence is sufficient;
- ETA Evolution interpretation against historical reliability.

Those extensions are not part of Stage 23 and require separate product decisions. Stage 23 must not pre-build them speculatively.

## 23. Non-goals

Stage 23 does not implement:

```text
flight schedule punctuality
commercial flight-status replacement
airport operations truth
runway touchdown detection
gate arrival detection
ATC surveillance
safety-critical guidance
paid aviation feeds
full weather-grid modeling
Fréchet trajectory matching
persistent trajectory spatial index
new ML model training
```

## 24. Current disposition

```text
STAGE_23_ETA_RELIABILITY=IN_PROGRESS
STAGE_23_FRONTEND_CONSUMER=YES
STAGE_23_ZERO_COST=YES
STAGE_23_DATA_FOR_DATA_SAKE=NO
STAGE_23_ARCHITECTURE_FOR_ARCHITECTURE_SAKE=NO
STAGE_23_OFFICIAL_ARRIVAL_TRUTH=NONE
STAGE_23_ENDPOINT_PROXY_DISCLOSURE=REQUIRED
STAGE_23_SAMPLE_SIZE_DISCLOSURE=REQUIRED
STAGE_23_DOCUMENTATION_ALIGNMENT=COMPLETE_FOR_IN_PROGRESS_STATE
STAGE_23_FINAL_EXACT_HEAD=UNKNOWN
STAGE_23_CLOSED=NO
```

This document is the canonical pre-merge engineering record for Stage 23. A future post-merge closure document, if the feature is successfully merged and independently verified, must preserve this failure and remediation history rather than replacing it.
