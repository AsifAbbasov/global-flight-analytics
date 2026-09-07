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

The service reuses production Projection Intelligence historical recomputation and computes only the bounded ETA absolute-error and interval-coverage definitions needed by the product. The production runtime does not import the offline-only Projection Evaluation package.

## 12. HTTP contract

The intended production read surface is:

```text
GET /api/v1/trajectories/{id}/eta-reliability
```

The endpoint is read-only.

The feature branch contains the source-backed 40-operation OpenAPI candidate, byte-identical root/embedded specifications and a regenerated TypeScript client exposing `getETAReliabilityByTrajectoryID`. Canonical `main` remains on the prior 39-operation contract until merge and independent post-merge verification. Stage 23 remains `IN_PROGRESS` until an exact-head authorized merge and post-merge closure are independently verified.

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

The UI deliberately requires a current Projection Intelligence arrival estimate before requesting historical ETA reliability. Historical reliability is contextual evidence beside an actual current ETA, not a standalone percentage. The deterministic Stage 23 browser scenario therefore supplies a bounded current ETA while production retains the honest `arrival === undefined` guard.

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

OpenAPI #143 remained a genuine release blocker until the canonical OpenAPI route/schema inventory and generated client were synchronized to the new read surface.

### 15.3 Head `5d96101274f863f80dc9d992b4a2558b49094016`

The OpenAPI synchronization had completed, but Playwright E2E #315 / run `34145986533` failed its foundation contract before Chromium execution. The only rejection was deterministic scenario ordering: the expected hard-coded scenario list placed `eta-reliability` after `intelligence-error`, while the validated sorted surface placed it before `healthy`.

This was test-contract drift, not a product transport defect. The expected sorted scenario list was corrected without weakening Playwright coverage.

### 15.4 Head `8849c188d7aa381f341b343afd1d27581f129a54`

Frontend CI #545 failed in the frontend contract-test step while ESLint and TypeScript validation had already succeeded. The new permanent Stage 23 test used repository-root paths even though `pnpm --filter web test` executes from `apps/web`, causing three `ENOENT` failures.

The permanent test was corrected to use workspace-relative paths. No production frontend or backend behavior changed.

### 15.5 Head `f2365626202428ccf10955286b5620e220f51ccc`

Backend CI #884 reached `go test ./...` after formatting succeeded. The Stage 23 ETA Reliability package itself passed, but `internal/http/apidocs` still asserted the historical embedded OpenAPI operation count of 39 and rejected the synchronized 40-operation contract.

The stale embedded API-doc test was updated from 39 to 40 operations. PostgreSQL integration remained successful; the failure was contract-test drift rather than an ETA reliability computation failure.

### 15.6 Head `15d040f251b84d3876c8f9cedc914d2fcc439169`

This head passed the major non-browser release gates:

```text
Frontend CI #547 = SUCCESS
Backend CI #885 = SUCCESS
OpenAPI Contract #176 = SUCCESS
API Load Baseline #404 = SUCCESS
CodeQL #528 = SUCCESS
Playwright E2E #318 = FAILURE
```

Playwright foundation passed and 20 of 21 Chromium product journeys passed. The only failing journey was the new Stage 23 ETA Reliability browser flow: `Historical ETA Reliability` never appeared because the deterministic `eta-reliability` scenario returned historical reliability evidence but its Projection Intelligence fixture still had `arrival_status: unavailable` and no current `projection.arrival`.

Production behavior was correct: `ETAReliabilityPanel` intentionally does not render or request historical reliability when the current projection has no ETA. The remediation therefore changed only the deterministic browser scenario to provide a bounded current ETA; the production `arrival === undefined` guard was preserved and permanently protected by the Stage 23 contract test.

### 15.7 Head `b2b73540ded8bd5ebb2bdd3012654b4d00aa1167`

A one-time self-cleaning workflow produced the corrected mock fixture and removed itself from the branch. Because this commit was authored by `github-actions[bot]`, GitHub classified the six recursively-created PR workflow runs as `action_required` rather than executing them as normal user-authored validation.

Those `action_required` runs are not pass evidence and are not treated as product failures. A subsequent normal user-authored permanent regression-test commit established a new exact-head validation boundary.

### 15.8 Validated implementation head `c55192a4077c541d9ac52b3d23f7eafc284a728b`

This user-authored head permanently protects both sides of the browser precondition: production requires a real current arrival estimate, while the deterministic Stage 23 browser scenario must supply one before historical reliability can be rendered.

Exact-head validation completed successfully:

```text
Frontend CI #550 / run 34149930933 = SUCCESS
Backend CI #888 / run 34149930944 = SUCCESS
OpenAPI Contract #179 / run 34149930961 = SUCCESS
API Load Baseline #407 / run 34149930956 = SUCCESS
CodeQL #531 / run 34149930946 = SUCCESS
Playwright E2E #321 / run 34149930942 = SUCCESS
Vercel deployment 5Nafn9zhNJJNkV9fw7WXDGm57QBs = SUCCESS
```

The same PR state check reported `mergeable=true`, with zero review threads and zero submitted reviews. PR #171 remained a draft and no merge authorization had been granted.

Earlier Vercel preview attempts in the remediation window were rejected by the free-tier build-rate limit. No paid upgrade was purchased. The validated implementation head later deployed successfully on the existing zero-cost Vercel path, so the temporary external rate-limit blocker did not become a product or cost-scope change.

## 16. Current validation status

The feature implementation has a complete successful exact-head evidence matrix on `c55192a4077c541d9ac52b3d23f7eafc284a728b`. This document update intentionally creates a later documentation head, so validation is not transferred automatically: the documentation head must run its own exact-head CI before PR #171 can become review-ready.

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
STAGE_23_IMPLEMENTATION_SEQUENCE=ALIGNED_V1_8
STAGE_23_OPENAPI_CANDIDATE_OPERATIONS=40
STAGE_23_OPENAPI_CANDIDATE_GET_OPERATIONS=39
STAGE_23_GENERATED_CLIENT=ALIGNED
STAGE_23_DEDICATED_PLAYWRIGHT_JOURNEY=PASS_ON_VALIDATED_IMPLEMENTATION_HEAD
STAGE_23_DOCUMENTATION_REGRESSION_TEST=INSTALLED
STAGE_23_VALIDATED_IMPLEMENTATION_HEAD=c55192a4077c541d9ac52b3d23f7eafc284a728b
STAGE_23_VALIDATED_IMPLEMENTATION_GITHUB_CI=6_OF_6_PASS
STAGE_23_VALIDATED_IMPLEMENTATION_VERCEL=PASS
STAGE_23_EXACT_HEAD_FINAL_CI=NOT_YET_AVAILABLE
STAGE_23_POST_MERGE_CI=NOT_APPLICABLE
```

`STAGE_23_EXACT_HEAD_FINAL_CI=NOT_YET_AVAILABLE` is an authoring-time governance marker for this pre-merge evidence update. It must not be rewritten recursively merely to insert its own future commit SHA; the final review-ready exact head and its validation matrix are recorded in PR #171 after this documentation head is independently verified.

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

Document 25 is aligned to Implementation Baseline v1.8: Version 2 reconciliation is closed, the historical statement that reconciliation itself did not create Stage 23 is preserved, and a later append-only Stage 23 section records the explicit zero-cost frontend product decision.

No synthetic finding ID is created merely because a product feature exists. `GFA-SEC-445` remains the independent existing repository-security finding and is not modified by Stage 23.

## 18. Remaining engineering work before review-ready

Product implementation, formatting, source/OpenAPI/generated-client synchronization, permanent regression tests, dedicated browser coverage and the zero-cost Vercel deployment have all been validated on the implementation head recorded above.

The remaining pre-review work is governance-only:

1. independently validate the documentation head created by this evidence update;
2. re-check exact-head mergeability, review threads and reviews;
3. record that final exact-head matrix in PR #171 without creating another recursive evidence commit;
4. mark the PR ready for review only if every Review-ready gate below is satisfied.

No additional product code, provider, infrastructure, migration or paid service is required by the currently known Stage 23 scope.

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
STAGE_23_VALIDATED_IMPLEMENTATION_HEAD=c55192a4077c541d9ac52b3d23f7eafc284a728b
STAGE_23_VALIDATED_IMPLEMENTATION_GITHUB_CI=6_OF_6_PASS
STAGE_23_VALIDATED_IMPLEMENTATION_VERCEL=PASS
STAGE_23_FINAL_EXACT_HEAD=UNKNOWN
STAGE_23_CLOSED=NO
```

This document is the canonical pre-merge engineering record for Stage 23. A future post-merge closure document, if the feature is successfully merged and independently verified, must preserve this failure and remediation history rather than replacing it.