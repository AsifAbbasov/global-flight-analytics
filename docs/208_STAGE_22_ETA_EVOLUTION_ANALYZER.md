# Stage 22 — Estimated Time of Arrival Evolution Analyzer

Status: architecture and evidence contract locked; second validation rejection remediated; final exact-head validation pending

```text
STAGE_22_ETA_EVOLUTION=IN_PROGRESS
STAGE_22_BASE_MAIN=84989c72e4cc8fddb79e484f2c7de2e6e597ba0f
STAGE_22_ROADMAP_SOURCE=VERSION_2_ITEM_17
STAGE_22_EVIDENCE_CLASS=HISTORICALLY_RECOMPUTED_FROM_PERSISTED_OBSERVATIONS
STAGE_22_PERSISTED_FORECAST_HISTORY=NONE
STAGE_22_NEW_BACKEND_ENDPOINT=NONE
STAGE_22_NEW_DATABASE_TABLE=NONE
STAGE_22_NEW_MIGRATION=NONE
STAGE_22_NEW_PROVIDER=NONE
STAGE_22_ETA_INTERPOLATION=NONE
STAGE_22_CAUSE_INFERENCE=NONE
STAGE_22_OPERATIONAL_GUIDANCE=NONE
STAGE_22_SAMPLE_CAP=6
STAGE_22_ADDITIONAL_COST=0_RUB
STAGE_22_FIRST_VALIDATION_HEAD=226526756f1a4b5e9350f77c048a647cfe82ca75
STAGE_22_FIRST_FRONTEND_CI=FAIL
STAGE_22_FIRST_FAILURE_CLASS=SOURCE_CONTRACT_FALSE_POSITIVE
STAGE_22_SECOND_VALIDATION_HEAD=00cec2e18c1430c9ca27d86b5776d508fed2f550
STAGE_22_SECOND_PLAYWRIGHT_CI=FAIL_FOUNDATION
STAGE_22_SECOND_FAILURE_CLASS=NON_SEMANTIC_PLAYWRIGHT_LOCATOR
STAGE_22_SECOND_FAILURE_REMEDIATION=SEMANTIC_COMPLEMENTARY_ROLE_LOCATOR
STAGE_22_BROWSER_E2E=ADDED_REMEDIATION_PENDING_VALIDATION
STAGE_22_FINAL_EXACT_HEAD_CI=PENDING
STAGE_22_POST_MERGE_CI=PENDING
```

## 1. Product need

The Version 2 roadmap explicitly contains `Estimated Time of Arrival Evolution Analyzer`. The repository already has production Projection Intelligence and Estimated Arrival, but the normal product exposes only one projection result at one analytical `as_of_time`. A user cannot yet see how that estimate changed as additional persisted trajectory evidence arrived.

Stage 22 closes that product gap without inventing forecast history that the project did not persist.

## 2. Existing evidence that makes the increment possible

The existing Projection Intelligence HTTP contract already requires an explicit `as_of_time`:

```text
GET /api/v1/trajectories/{id}/projection-intelligence
  ?as_of_time=<RFC3339>
  &duration_seconds=<positive integer>
```

The backend Projection Intelligence read service loads a snapshot bounded to the requested trajectory and historical `as_of_time`, then composes the normal production projection from evidence available at that timestamp.

The existing Estimated Arrival contract can contain:

- airport ICAO code;
- earliest arrival time;
- estimated arrival time;
- latest arrival time;
- arrival confidence;
- limitations.

Historical replay already exposes persisted flight observations with exact `observed_at` timestamps and an explicit `interpolation_policy=none`.

Therefore Stage 22 can select a bounded set of persisted replay observation timestamps and request the existing production projection at those timestamps.

## 3. Critical terminology: recomputed history, not persisted forecast history

Stage 22 does **not** claim that prior projection outputs were stored at the time they originally occurred.

Each evolution point is a deterministic historical recomputation performed now from persisted evidence available at that historical `as_of_time` under the current production Projection Intelligence implementation and current policy/version.

The product must disclose this distinction.

```text
PERSISTED_OBSERVATION_TIME=YES
PERSISTED_HISTORICAL_FORECAST_OUTPUT=NO
HISTORICAL_RECOMPUTATION=YES
CURRENT_MODEL_VERSION_APPLIED_TO_PAST_EVIDENCE=YES
```

This is materially different from a future persisted `forecast_versions` ledger. If a later stage stores immutable forecast outputs at generation time, those records must be presented as a different evidence class.

## 4. Sampling policy

Stage 22 must not call Projection Intelligence for every replay observation. That could multiply backend work for long trajectories without adding proportional product value.

The first increment uses at most six persisted replay observation timestamps.

Selection rules:

1. input observations are sorted by `observed_at`;
2. duplicate/invalid timestamps are rejected from sampling;
3. when the replay contains at most six valid observations, every valid timestamp is selected;
4. otherwise the first and last valid observations are always selected;
5. the remaining points are deterministic approximately-even positions across the observed index span;
6. only real persisted observation timestamps are selected;
7. no synthetic timestamp is inserted between observations.

```text
MAX_PROJECTION_RECOMPUTATIONS=6
FIRST_OBSERVATION_INCLUDED=YES
LAST_OBSERVATION_INCLUDED=YES
SYNTHETIC_TIME_SAMPLE=NONE
```

## 5. Evolution point contract

For each sampled persisted observation timestamp, Stage 22 records a presentation point containing only information returned by the existing Projection Intelligence result:

- `as_of_time`;
- projection availability/status;
- production strategy/method;
- arrival status;
- airport ICAO code when an arrival estimate is attached;
- earliest / estimated / latest arrival timestamps when attached;
- arrival confidence score/level;
- production input fingerprint;
- projection scope guard.

A request that returns an evidence-bounded unavailable result remains an unavailable evolution point. A transport/service error is not converted into an ETA value.

## 6. Derived comparison semantics

Frontend comparison may derive only transparent arithmetic between returned ETA timestamps.

For adjacent sampled points only, when **both** points expose an available estimated arrival:

```text
eta_change_seconds = current estimated arrival - immediately previous sampled estimated arrival
```

For each available ETA point:

```text
arrival_window_seconds = latest arrival - earliest arrival
```

An unavailable sampled point intentionally breaks adjacent ETA comparison. Stage 22 does not skip over that gap and compare the next available point with an older available point as though the missing sample did not exist.

The analyzer may summarize:

- first available ETA;
- latest available ETA;
- net ETA change across those two available endpoints;
- largest absolute **observed adjacent sampled** ETA change;
- count of sampled observations;
- count of available ETA points;
- count of unavailable ETA points.

`net ETA change` is an endpoint summary and does not imply uninterrupted ETA evidence between the endpoints.

It must not infer why an ETA changed.

```text
CAUSE_OF_ETA_CHANGE=UNKNOWN
DELAY_CAUSE_INFERENCE=NONE
ATC_CAUSE_INFERENCE=NONE
WEATHER_CAUSE_INFERENCE=NONE
AIRPORT_CONGESTION_CAUSE_INFERENCE=NONE
```

Those causes require independent evidence and cannot be inferred merely from changing forecast values.

## 7. Missing-data policy

Unknown remains unknown.

Stage 22 must not:

- interpolate an ETA between sampled observations;
- carry a previous ETA forward as if it were a new estimate;
- bridge an unavailable sampled point when calculating adjacent ETA change;
- convert an unavailable arrival estimate to zero duration;
- manufacture an airport code when the backend withholds arrival;
- treat a failed request as a stable forecast.

The UI must keep gaps visible.

## 8. Frontend architecture

The first increment is frontend-orchestrated and reuses existing query/data boundaries:

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
ETA Evolution panel inside Projection Intelligence product flow
```

No raw component-level `fetch` or axios path is allowed.

The Projection Intelligence panel observes the existing `useLatestAircraftTrajectory` TanStack query key rather than creating a new trajectory transport path. Historical replay likewise reuses the existing `useTrajectoryFlightReplay` cache contract.

## 9. Query behavior

Historical recomputation queries are immutable for the selected trajectory/update version and historical `as_of_time` during a browser session. They therefore must not use the one-minute live refetch behavior of the current projection query.

The evolution query:

- uses existing `getProjectionIntelligence` transport validation;
- uses the standard projection duration;
- remains disabled without a trajectory/replay sample;
- uses `staleTime=Infinity` for the historical query key;
- uses `refetchInterval=false`;
- disables refetch-on-window-focus;
- preserves bounded retry policy for transient server failures;
- permits an explicit user retry without fabricating a replacement point.

## 10. UI evidence boundary

The panel prominently states:

- historical points are recomputed now from persisted observations;
- they are not immutable forecasts stored at those historical moments;
- no movement or ETA is interpolated between samples;
- change direction is descriptive and does not establish delay or operational cause.

The UI exposes machine-readable guards:

```text
data-eta-evolution-evidence=historically-recomputed-from-persisted-observations
data-eta-evolution-persisted-forecast-history=none
data-eta-evolution-interpolation=none
data-eta-evolution-cause-inference=none
```

## 11. Regression protection

The real production ETA evolution model is compiled through `apps/web/tsconfig.test.json` and executed from `.test-dist`.

Permanent model tests protect:

- the six-sample maximum;
- selection from real persisted timestamps only;
- first/last observation retention when sampling;
- transparent adjacent and net ETA arithmetic;
- arrival-window width;
- unavailable-sample gap behavior;
- no carry-forward after unavailable evidence;
- transport errors remaining unavailable rather than becoming synthetic ETA values.

Permanent source/documentation contracts protect:

- reuse of `getProjectionIntelligence` rather than a new raw network path;
- no historical background refetch interval;
- machine-readable recomputation/interpolation/cause guards;
- no fake persisted-forecast claim;
- integration inside the existing Projection Intelligence product flow;
- zero-budget and no-new-backend-data boundaries;
- the distinction between current-model historical recomputation and an immutable forecast ledger.

The existing Chromium `advanced-intelligence.spec.mjs` journey is extended rather than creating a detached browser scenario. It requires the visible `Estimated Arrival Evolution` panel, the historically-recomputed evidence class, `persisted-forecast-history=none`, `interpolation=none`, `cause-inference=none`, and visible copy explaining both non-persisted forecast history and no ETA interpolation. The permanent source contract also requires those Chromium assertions to remain present.

The browser assertion uses the semantic `complementary` landmark named by `Estimated Arrival Evolution`. Stage 22 does not use a CSS `page.locator(...)` escape hatch for this panel.

## 12. Rejected alternatives

### New `forecast_versions` table now

Rejected for the first increment. It would only start collecting forecasts after deployment and would not solve historical evolution for trajectories already persisted. It also adds a migration, write path, retention policy and operational lifecycle before the value of the product view is proven.

### Recompute every replay observation

Rejected as unbounded repeated projection work.

### Generate evenly spaced wall-clock timestamps

Rejected because those timestamps would not themselves be persisted observation evidence.

### Bridge missing samples for a smoother ETA line

Rejected because an unavailable sampled projection is real missing analytical evidence. A smooth line across that gap would visually imply continuity the project did not observe.

### Infer delay / weather / congestion causes

Rejected because ETA movement alone cannot establish causality.

## 13. Cost and infrastructure

```text
NEW_PROVIDER=NO
NEW_DATABASE_TABLE=NO
NEW_MIGRATION=NO
NEW_CACHE_SERVICE=NO
NEW_SERVER=NO
NEW_PAID_SERVICE=NO
ADDITIONAL_COST=0_RUB
```

The increment reuses already persisted flight states, trajectory identity, the existing Projection Intelligence read path and existing frontend infrastructure.

## 14. Real first validation rejection and remediation

The first Stage 22 PR validation is preserved rather than rewritten away.

Initial exact feature head:

```text
HEAD=226526756f1a4b5e9350f77c048a647cfe82ca75
```

Frontend CI #487 / run `34063552010` reached the Stage 22 implementation with:

```text
ESLINT=PASS
TYPESCRIPT=PASS
ETA_EVOLUTION_MODEL_TESTS=PASS
FRONTEND_CONTRACT_TESTS=204_PASS_1_FAIL
```

The only failing test was:

```text
Stage 22 reuses persisted replay timestamps and existing Projection Intelligence transport
```

The source contract attempted to forbid raw browser fetches with:

```text
/fetch\s*\(/
```

That pattern also matched the `fetch(` substring inside the legitimate TanStack method call `query.refetch()`. The production query did not contain a raw `fetch()` transport path; it called the existing `getProjectionIntelligence` API wrapper as intended.

Root cause:

```text
PRODUCT_TRANSPORT_DEFECT=NO
SOURCE_CONTRACT_FALSE_POSITIVE=YES
FALSE_MATCH=query.refetch()
```

Rejected remediation:

- removing the raw-fetch guard;
- changing production query behavior merely to satisfy the regex;
- deleting explicit retry support.

Selected remediation:

```text
RAW_FETCH_GUARD=/\bfetch\s*\(/
```

The word boundary rejects actual `fetch(` calls but does not match the `fetch` substring inside `refetch(`. The same semantic guard is applied to the Projection Intelligence integration source check.

This remediation changes regression-test precision only. It does not loosen the architectural ban on a second raw network path.

## 15. Real second validation rejection and remediation

After adding explicit browser coverage, the browser/documentation-complete exact head was:

```text
HEAD=00cec2e18c1430c9ca27d86b5776d508fed2f550
```

Playwright E2E #263 / run `34063846532` failed during the repository Playwright foundation contract **before Chromium was started**.

Foundation result:

```text
PLAYWRIGHT_FOUNDATION_TESTS=15
PLAYWRIGHT_FOUNDATION_PASS=13
PLAYWRIGHT_FOUNDATION_FAIL=2
CHROMIUM_SUITE=SKIPPED
```

The two failures reported:

```text
Playwright tests must prefer semantic locators
browser product coverage uses semantic locators and no public deployment
```

The new Stage 22 browser assertion located the ETA Evolution panel using:

```text
page.locator("[data-eta-evolution-evidence='historically-recomputed-from-persisted-observations']")
```

The repository has an explicit Playwright quality contract rejecting `page.locator(...)` in product E2E tests in favor of semantic role/text/label locators. The data attribute itself remains useful as evidence metadata, but it must not be the primary browser locator.

Root cause:

```text
PRODUCT_UI_DEFECT=NO
CHROMIUM_PRODUCT_VERDICT=NOT_REACHED
PLAYWRIGHT_FOUNDATION_POLICY_VIOLATION=YES
NON_SEMANTIC_LOCATOR=page.locator
```

Rejected remediation:

- weakening or deleting the repository semantic-locator verifier;
- excluding the Stage 22 spec from the policy;
- keeping a CSS selector merely because the element exposes a data attribute.

Selected remediation:

```text
page.getByRole('complementary', { name: 'Estimated Arrival Evolution' })
```

The production `<aside aria-labelledby='eta-evolution-title'>` already exposes a named complementary landmark, so the browser test can locate the product surface semantically. The test still verifies all four `data-eta-evolution-*` evidence attributes after locating the panel by role.

This remediation changes the Playwright assertion only. It does not change product analytics, evidence semantics, API usage, sampling, or UI claims.

## 16. Residual limitations

Historically recomputed forecast points use the current projection implementation and current policy against historical evidence. They are therefore suitable for analyzing how the current system would have evolved as evidence accumulated, but they are not an audit ledger of what an older software version actually predicted at that time.

The sample cap intentionally means the displayed series is not every forecastable instant. It is a bounded evidence sample over persisted observations.

The current frontend orchestration can trigger at most six Projection Intelligence recomputations when a selected trajectory/replay changes. This is intentionally bounded, but it is still more backend work than the single current projection request and should remain protected against unbounded sample growth.

## 17. Pre-merge status

Final workflow IDs belong in PR metadata after the documentation-complete head exists; mutating this document merely to record its own future exact SHA would recursively invalidate validation.

```text
STAGE_22_ARCHITECTURE=LOCKED
STAGE_22_PRODUCT_IMPLEMENTATION=IMPLEMENTED_PREMERGE
STAGE_22_DOCUMENTATION=FIRST_AND_SECOND_VALIDATION_REJECTIONS_RECORDED
STAGE_22_FIRST_VALIDATION_HEAD=226526756f1a4b5e9350f77c048a647cfe82ca75
STAGE_22_FIRST_FRONTEND_CI=FAIL
STAGE_22_FIRST_FAILURE_REMEDIATION=SEMANTIC_RAW_FETCH_GUARD
STAGE_22_SECOND_VALIDATION_HEAD=00cec2e18c1430c9ca27d86b5776d508fed2f550
STAGE_22_SECOND_PLAYWRIGHT_CI=FAIL_FOUNDATION
STAGE_22_SECOND_FAILURE_REMEDIATION=SEMANTIC_COMPLEMENTARY_ROLE_LOCATOR
STAGE_22_BROWSER_E2E=REMEDIATED_PENDING_VALIDATION
STAGE_22_FINAL_EXACT_HEAD_CI=PENDING
STAGE_22_POST_MERGE_CI=PENDING
STAGE_22_ETA_EVOLUTION=IN_PROGRESS
```
