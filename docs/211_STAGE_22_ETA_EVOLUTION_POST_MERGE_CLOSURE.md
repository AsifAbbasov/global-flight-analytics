# Stage 22 — Estimated Time of Arrival Evolution Post-Merge Closure

Status: CLOSED when this closure record is merged into canonical `main` and the resulting exact `main` SHA is independently verified.

```text
STAGE_22_ETA_EVOLUTION=CLOSURE_CANDIDATE
STAGE_22_MERGED_PR=165
STAGE_22_APPROVED_HEAD=eb8a2b2b7680b8e4590c520c16d60954a0cfd3a2
STAGE_22_FEATURE_MAIN_SHA=1020b06af9827057ac9296ef09914c232cd7f510
STAGE_22_PARENT_MAIN=e2189b8e38741abf1e4b87e1c937e6b623759f41
STAGE_22_PREMERGE_EXACT_HEAD_CI=PASS
STAGE_22_PREMERGE_VERCEL=PASS
STAGE_22_POST_MERGE_FEATURE_CI=PASS
STAGE_22_POST_MERGE_FEATURE_VERCEL=PASS
STAGE_22_POST_MERGE_API_LOAD=NOT_TRIGGERED_FRONTEND_ONLY_PATH_POLICY
STAGE_22_POST_MERGE_OPENAPI=NOT_TRIGGERED_FRONTEND_ONLY_PATH_POLICY
STAGE_22_EVIDENCE_CLASS=HISTORICALLY_RECOMPUTED_FROM_PERSISTED_OBSERVATIONS
STAGE_22_PERSISTED_FORECAST_HISTORY=NONE
STAGE_22_ETA_INTERPOLATION=NONE
STAGE_22_CAUSE_INFERENCE=NONE
STAGE_22_OPERATIONAL_GUIDANCE=NONE
STAGE_22_SAMPLE_CAP=6
STAGE_22_NEW_BACKEND_ENDPOINT=NONE
STAGE_22_NEW_DATABASE_TABLE=NONE
STAGE_22_NEW_MIGRATION=NONE
STAGE_22_NEW_PROVIDER=NONE
STAGE_22_ADDITIONAL_COST=0_RUB
```

## 1. Closure purpose

Document 208 is the immutable Stage 22 pre-merge engineering-history record. It intentionally preserves the original architecture/evidence contract, the first Frontend CI rejection, the Playwright foundation rejection, their remediations, and the fact that final merge/post-merge evidence did not yet exist when that document was authored.

This document appends the actual exact-head merge and canonical-main evidence. Document 208 is not rewritten retroactively.

Stage 22 is not formally closed merely because this file exists on a feature branch. Formal closure requires this record and its permanent regression test to be merged into canonical `main`, followed by exact-SHA validation of that resulting `main`.

## 2. Product scope closed by Stage 22

Stage 22 adds the Version 2 Estimated Time of Arrival Evolution Analyzer as a bounded, research-only historical recomputation over persisted replay observations and the existing Projection Intelligence path.

The production evidence flow is:

```text
persisted replay observations
        ↓
real persisted observed_at timestamps
        ↓
deterministic max-6 sampler
        ↓
existing Projection Intelligence with historical as_of_time
        ↓
current production projection applied to evidence available at that historical time
        ↓
ETA Evolution presentation
```

The feature does not claim that historical forecasts were persisted when they originally occurred.

## 3. Evidence and analytical contract

The canonical evidence class remains:

```text
HISTORICALLY_RECOMPUTED_FROM_PERSISTED_OBSERVATIONS
```

Each evolution point uses a real persisted replay observation timestamp as `as_of_time` and applies the current production Projection Intelligence implementation to evidence available at that historical time.

The feature therefore preserves these boundaries:

```text
PERSISTED_OBSERVATION_TIME=YES
PERSISTED_HISTORICAL_FORECAST_OUTPUT=NO
HISTORICAL_RECOMPUTATION=YES
CURRENT_MODEL_VERSION_APPLIED_TO_PAST_EVIDENCE=YES
MAX_PROJECTION_RECOMPUTATIONS=6
SYNTHETIC_TIME_SAMPLE=NONE
ETA_INTERPOLATION=NONE
CARRY_FORWARD=NONE
GAP_BRIDGING=NONE
```

Unavailable samples remain unavailable. An unavailable sampled point breaks adjacent ETA comparison. A failed transport request is not converted into an ETA value.

## 4. Interpretation boundary

ETA movement is descriptive evidence, not causal evidence.

```text
CAUSE_OF_ETA_CHANGE=UNKNOWN
DELAY_CAUSE_INFERENCE=NONE
ATC_CAUSE_INFERENCE=NONE
WEATHER_CAUSE_INFERENCE=NONE
AIRPORT_CONGESTION_CAUSE_INFERENCE=NONE
OPERATIONAL_GUIDANCE=NONE
```

Stage 22 must not convert changing ETA values into claims about delay cause, weather impact, Air Traffic Control instruction, airport congestion cause, pilot intent, or operational flight guidance.

## 5. Bounded frontend architecture

Stage 22 remains a frontend-orchestrated increment that reuses existing production contracts:

```text
existing AircraftTrajectory
        ↓
existing useTrajectoryFlightReplay cache
        ↓
persisted FlightReplayPoint.observed_at
        ↓
deterministic max-6 sampler
        ↓
existing getProjectionIntelligence(...as_of_time...)
        ↓
TanStack useQueries historical recomputation
        ↓
pure ETA evolution model
        ↓
ETA Evolution panel inside Projection Intelligence
```

Historical recomputation queries preserve the bounded immutable-query behavior recorded by Document 208: `staleTime=Infinity`, `refetchInterval=false`, no refetch-on-window-focus, existing transport validation, and no component-level raw `fetch` path.

## 6. User-visible evidence guards

The frontend preserves machine-readable evidence metadata:

```text
data-eta-evolution-evidence=historically-recomputed-from-persisted-observations
data-eta-evolution-persisted-forecast-history=none
data-eta-evolution-interpolation=none
data-eta-evolution-cause-inference=none
```

The panel explicitly explains that historical points are recomputed now from persisted observations, are not immutable historical forecast records, contain no ETA interpolation between samples, and do not establish operational cause.

## 7. Preserved real validation history

Stage 22 does not rewrite its development history into a fictitious first-pass success.

### First rejection

Initial exact head:

```text
226526756f1a4b5e9350f77c048a647cfe82ca75
```

Frontend CI #487 / run `34063552010` failed because a regression-test regex intended to forbid raw browser `fetch()` also matched the `fetch(` substring inside legitimate `query.refetch()`.

```text
PRODUCT_TRANSPORT_DEFECT=NO
SOURCE_CONTRACT_FALSE_POSITIVE=YES
FALSE_MATCH=query.refetch()
REMEDIATION=/\bfetch\s*\(/
```

The remediation improved test precision without weakening the architecture boundary.

### Second rejection

Browser-complete exact head:

```text
00cec2e18c1430c9ca27d86b5776d508fed2f550
```

Playwright E2E #263 / run `34063846532` failed in repository foundation policy before Chromium because the new assertion used non-semantic `page.locator(...)`.

```text
PRODUCT_UI_DEFECT=NO
CHROMIUM_PRODUCT_VERDICT=NOT_REACHED
PLAYWRIGHT_FOUNDATION_POLICY_VIOLATION=YES
REMEDIATION=getByRole('complementary', { name: 'Estimated Arrival Evolution' })
```

The remediation changed browser-test selection only; product analytics and evidence semantics were not weakened.

## 8. Final pre-merge exact-head evidence

After Stage 22 was rebased onto the fully closed Stage 21 `main`, the merge-authorized exact head was:

```text
eb8a2b2b7680b8e4590c520c16d60954a0cfd3a2
```

The exact-head validation matrix was:

```text
Frontend CI #502 / run 34090595253 = SUCCESS
Backend CI #840 / run 34090595257 = SUCCESS
API Load Baseline #363 / run 34090595277 = SUCCESS
CodeQL #482 / run 34090595226 = SUCCESS
Playwright E2E #273 / run 34090595329 = SUCCESS
OpenAPI Contract = NOT_TRIGGERED_FRONTEND_ONLY_PATH_POLICY
```

Vercel also succeeded on the same exact head:

```text
STAGE_22_PREMERGE_VERCEL=PASS
VERCEL_PREMERGE_DEPLOYMENT=CiwGkDwTNBWRKDtmJgXXt9XcY3bN
```

## 9. Exact-head merge authorization and result

The user authorized a squash merge specifically for:

```text
APPROVED_PR=165
APPROVED_HEAD=eb8a2b2b7680b8e4590c520c16d60954a0cfd3a2
MERGE_METHOD=SQUASH
```

GitHub accepted the expected-head guarded merge and produced:

```text
MERGED_PR=165
FEATURE_MAIN_SHA=1020b06af9827057ac9296ef09914c232cd7f510
PARENT_MAIN=e2189b8e38741abf1e4b87e1c937e6b623759f41
```

The resulting canonical `main` commit is GitHub-verified and directly descends from the Stage 21-closed canonical base.

## 10. Canonical feature post-merge validation

The exact Stage 22 feature merge SHA is:

```text
1020b06af9827057ac9296ef09914c232cd7f510
```

The push-triggered post-merge workflow matrix for that exact SHA is:

```text
Frontend CI #503 / run 34092099738 = SUCCESS
Backend CI #841 / run 34092099782 = SUCCESS
CodeQL #483 / run 34092099737 = SUCCESS
Playwright E2E #274 / run 34092099739 = SUCCESS
API Load Baseline = NOT_TRIGGERED_FRONTEND_ONLY_PATH_POLICY
OpenAPI Contract = NOT_TRIGGERED_FRONTEND_ONLY_PATH_POLICY
```

Only four GitHub Actions workflows were triggered on the canonical post-merge push for this frontend-only increment. The pre-merge API Load success is not transferred to the merge SHA and is not represented as post-merge evidence.

Vercel succeeded independently on the exact canonical feature merge SHA:

```text
STAGE_22_POST_MERGE_FEATURE_VERCEL=PASS
VERCEL_POSTMERGE_DEPLOYMENT=HXW8DXQXiGZqhR94bdwJd3GY2hYA
VERCEL_POSTMERGE_STATUS=Deployment has completed
```

## 11. Stage 20/21 overlap preservation

The final Stage 22 branch was rebuilt on the Stage 21-closed canonical base. Two real overlaps with Stage 20 were reconciled rather than overwritten:

```text
apps/web/e2e/tests/advanced-intelligence.spec.mjs
apps/web/tsconfig.test.json
```

The advanced-intelligence Chromium journey preserves Airspace Intelligence assertions alongside ETA Evolution assertions. The test TypeScript configuration preserves the Stage 20 airspace model and includes the Stage 22 ETA evolution model/projection types.

The successful post-merge Frontend and Playwright gates on the canonical merge SHA provide repository-real regression evidence for that reconciliation.

## 12. Regression protection

The feature's permanent model, source-contract and Chromium tests continue to protect:

- maximum six historical Projection recomputations;
- real persisted timestamps only;
- first/last observation retention when sampling;
- transparent adjacent/net ETA arithmetic;
- unavailable sample gap behavior;
- no carry-forward or interpolation;
- existing Projection Intelligence transport reuse;
- historical query refetch restrictions;
- machine-readable evidence guards;
- no fake persisted-forecast claim;
- no cause inference;
- integration inside the existing Projection Intelligence product flow.

This closure increment adds a permanent documentation contract protecting:

- merged PR identity;
- merge-authorized exact head;
- canonical feature merge SHA and parent;
- exact pre-merge workflow identifiers;
- exact post-merge workflow identifiers;
- explicit post-merge `NOT_TRIGGERED` status for API Load and OpenAPI rather than invented PASS evidence;
- Vercel exact-SHA success;
- preservation of both real validation rejections;
- Document 208 as historical pre-merge evidence;
- recomputed-history semantics and six-sample bound;
- no persisted forecast history, interpolation, cause inference or operational guidance;
- zero-budget/no-new-backend-data boundary.

## 13. Infrastructure and cost closure

```text
NEW_BACKEND_ENDPOINT=NO
NEW_PROVIDER=NO
NEW_INGESTION_PATH=NO
NEW_DATABASE_TABLE=NO
NEW_MIGRATION=NO
NEW_CACHE_SERVICE=NO
NEW_SERVER=NO
NEW_PAID_SERVICE=NO
ADDITIONAL_COST=0_RUB
```

Stage 22 therefore closes the product increment without adding infrastructure or cost.

## 14. Residual limitations

Stage 22 remains bounded by the evidence described in Document 208:

- historical points are recomputations under the current implementation, not immutable forecasts generated by older software versions;
- the six-sample cap means the series is a bounded sample, not every forecastable instant;
- unavailable samples create real gaps and are not bridged;
- ETA movement cannot identify delay, weather, Air Traffic Control, airport congestion, pilot intent or another operational cause;
- the analyzer is research-only and not operational flight guidance.

A future immutable forecast ledger would be a different evidence class and would require its own persistence, versioning, retention and validation contract.

## 15. Formal closure boundary

This document is a closure candidate until it is merged into canonical `main` together with its permanent regression test.

After that merge, the resulting exact canonical `main` SHA must independently pass the repository-applicable final validation gates. Only then may the project state be reported as:

```text
STAGE_22_ETA_EVOLUTION=CLOSED
STAGE_22_CI_VERIFIED=YES
```

Until then the truthful state is:

```text
STAGE_22_FEATURE_IMPLEMENTATION=MERGED
STAGE_22_FEATURE_POST_MERGE_CI=PASS
STAGE_22_FEATURE_POST_MERGE_VERCEL=PASS
STAGE_22_DOCUMENTATION_CLOSURE=AWAITING_CLOSURE_PR_MERGE
STAGE_22_ETA_EVOLUTION=CLOSURE_CANDIDATE
```

Once formal Stage 22 closure is complete, the remaining Version 2 work is the final Version 2 reconciliation audit rather than another automatically implied feature stage.
