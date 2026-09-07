# Stage 21 — Airport Congestion Intelligence

Status: product implementation and frontend integration in progress; final exact-head validation pending

```text
STAGE_21_AIRPORT_CONGESTION_INTELLIGENCE=IN_PROGRESS
STAGE_21_BASE_MAIN=84989c72e4cc8fddb79e484f2c7de2e6e597ba0f
STAGE_21_ENDPOINT=GET_/api/v1/airports/{icao}/intelligence/congestion
STAGE_21_ANALYTICAL_CLASS=RELATIVE_OBSERVED_ACTIVITY_PROXY
STAGE_21_AIRPORT_CAPACITY_MODEL=NONE
STAGE_21_RUNWAY_OCCUPANCY_MODEL=NONE
STAGE_21_QUEUE_MODEL=NONE
STAGE_21_SLOT_MODEL=NONE
STAGE_21_DELAY_INFERENCE=NONE
STAGE_21_OFFICIAL_CONGESTION_CLAIM=NONE
STAGE_21_NEW_PROVIDER=NONE
STAGE_21_NEW_INGESTION_PATH=NONE
STAGE_21_NEW_DATABASE_DATA=NONE
STAGE_21_ADDITIONAL_COST=0_RUB
STAGE_21_FINAL_EXACT_HEAD_CI=PENDING
STAGE_21_POST_MERGE_CI=PENDING
```

## 1. Product need

Version 2 planning identified Airport Congestion Score as a remaining product gap after reconciliation against the repository. Existing Airport Intelligence already publishes completed-day airport statistics, history and trends, but it does not expose a dedicated view of how the latest observed completed-day movement intensity compares with that airport's own prior observed activity.

Stage 21 closes that gap without pretending that open traffic evidence represents airport capacity or official operational congestion.

## 2. Evidence boundary

The feature uses only existing persisted Airport Intelligence completed-day observations. The backend already derives arrivals, departures, movements per hour, active aircraft/routes, coverage and freshness from project data.

Stage 21 does not add:

- an airport-capacity data source;
- declared slot or schedule data;
- runway occupancy data;
- queue observations;
- delay observations;
- a new provider or ingestion path;
- a database table or migration;
- paid infrastructure.

Therefore the output is explicitly a **relative observed-activity proxy**, not an operational airport congestion measurement.

## 3. Selected analytical contract

For one requested completed-day window:

```text
current = latest expected completed UTC day, only when that day is observed
baseline = median movements/hour across prior observed completed-day windows
prior_peak = maximum movements/hour across prior observed completed-day windows
current_to_baseline_ratio = current / baseline, when baseline > 0
current_to_prior_peak_ratio = current / prior_peak, when prior_peak > 0
congestion_score = min(1, current_to_prior_peak_ratio)
```

The uncapped current/prior-peak ratio is retained separately so a new observed activity peak does not lose magnitude merely because the bounded display score reaches 1.

No LOW/MEDIUM/HIGH congestion thresholds are introduced. There is no calibration evidence for such labels.

## 4. Missing-data policy

A critical design issue was identified before final integration: the existing `history.History` value describes the span from the first observed entry to the last observed entry. If the requested window contains a leading or trailing missing day, deriving expected coverage only from that compressed history span can overstate evidence support.

Stage 21 therefore passes the requested production window separately into the congestion analyzer.

Consequences:

- expected window count comes from the requested completed-day interval;
- leading, internal and trailing missing days remain part of evidence coverage;
- `trailing_gap_window_count` is explicit;
- `current_window_is_latest_expected` is explicit;
- when the latest expected completed day is absent, the previous observed day is not silently substituted as current;
- score availability remains false when required current or denominator evidence is unavailable.

Unknown is not converted to zero.

## 5. Evidence support

The response preserves:

- observed window count;
- expected window count;
- gap window count;
- trailing gap window count;
- evidence coverage;
- conservative evidence support;
- baseline window count;
- baseline median movements/hour;
- prior observed peak movements/hour;
- explicit known flags for derived ratios and score.

Frontend presentation preserves the backend known/unknown state. A false `*_known` flag renders the value as unavailable rather than as zero.

## 6. Production composition

The endpoint is:

```text
GET /api/v1/airports/{icao}/intelligence/congestion
```

It uses the same `days` and optional `as_of_time` request semantics as existing Airport Intelligence reads.

Production composition reuses `GetHistory` and the existing Airport Intelligence observation reader. There is no second SQL path and no parallel historical store.

The public API operation inventory moves from:

```text
38 operations = 37 GET + 1 POST
```

to:

```text
39 operations = 38 GET + 1 POST
```

The canonical OpenAPI contract, embedded API documentation contract and generated TypeScript client are synchronized with that operation surface.

## 7. Frontend integration

Stage 21 extends the existing Unified Airport Analytics Workspace rather than creating a parallel product flow.

```text
existing airport-intelligence API transport
        ↓
validated congestion parser
        ↓
existing TanStack Query domain
        ↓
selected airport + completed-day window
        ↓
Activity pressure profile tab
        ↓
research-only evidence panel
```

The UI exposes:

- bounded observed-activity score when known;
- current/baseline ratio when known;
- current/prior-peak ratio when known;
- current movements/hour;
- baseline and prior-peak values;
- new-peak evidence;
- observed/expected/gap/trailing-gap counts;
- evidence coverage/support;
- whether the current observation is the latest expected completed day;
- server-published score semantics, explanation and scope guard.

The UI includes machine-readable guards:

```text
data-airport-congestion-scope=relative-observed-activity-only
data-airport-capacity-model=none
data-airport-delay-inference=none
```

## 8. Regression protection

Backend unit tests protect:

- relative baseline and prior-peak arithmetic;
- bounded score with uncapped magnitude retained separately;
- zero-denominator unavailability;
- insufficient history;
- invalid evidence scores;
- requested-window gap accounting;
- missing latest expected day behavior.

Frontend tests protect:

- unknown score/ratios remain null;
- evidence ratios remain bounded;
- production endpoint and TanStack Query integration;
- research-only UI data attributes and wording;
- absence of LOW/MEDIUM/HIGH congestion labels;
- deterministic mock API support.

The Airport Intelligence Playwright journey is extended through the `Activity pressure` tab and verifies the evidence scope guard in Chromium.

## 9. Real validation failures and remediation history

Stage 21 preserves the actual validation history rather than rewriting it into a fictitious first-pass success.

### 9.1 Initial OpenAPI inventory failure

The new production route was registered before the canonical public-operation inventory and OpenAPI document were synchronized. OpenAPI Contract correctly rejected:

```text
unclassified source operation: GET /api/v1/airports/{icao}/intelligence/congestion
public operation count: found 39, expected 38
OpenAPI missing operation count: 1
```

The remediation updated the canonical 39-operation contract, embedded copy and generated TypeScript client.

### 9.2 Backend response and browser mock contract failures

On exact head:

```text
bf2f0a22c54b9965803688061caedf9521ab8353
```

GitHub validation showed:

```text
OpenAPI Contract = SUCCESS
Frontend CI = SUCCESS
CodeQL = SUCCESS
Backend CI = FAILURE
API Load Baseline = FAILURE
Playwright E2E = FAILURE
```

Root causes were contract integration omissions rather than analytical failures:

- new Go files required `gofmt`;
- `dto.AirportCongestionIntelligenceResponse` was missing from the generic success-response union;
- the Playwright mock public path inventory did not yet include the new operation.

API Load failed while compiling the same backend response contract; it did not establish a load-performance regression.

The remediation formatted the Go files, added the response type to the success envelope and synchronized the mock public path surface.

### 9.3 Embedded OpenAPI operation-count failure

After the above remediation, the congestion analyzer, response envelope, handlers, server, race tests and PostgreSQL integration passed. Backend Quality then exposed one remaining historical test assertion:

```text
expected 38 embedded OpenAPI operations, got 39
```

The assertion was updated to the new canonical 39-operation surface.

## 10. Temporary generation/remediation tooling

Large generated OpenAPI and mock/API workspace files were synchronized using temporary feature-branch-only GitHub runner scripts because this execution environment did not provide a reliable direct local GitHub checkout for deterministic repository generators.

Those temporary scripts and temporary `contents: write` workflow permissions were removed after their generated/product changes were persisted. The final Stage 21 diff must not contain self-writing CI or Stage 21 generation/remediation scripts.

## 11. Infrastructure and cost

```text
NEW_PROVIDER=NO
NEW_DATABASE_TABLE=NO
NEW_MIGRATION=NO
NEW_CACHE=NO
NEW_SERVER=NO
NEW_PAID_SERVICE=NO
ADDITIONAL_COST=0_RUB
```

The additional HTTP read is composed over existing Airport Intelligence data and product infrastructure.

## 12. Residual limitations

The feature cannot answer questions that require evidence the project does not possess, including:

- airport declared capacity;
- runway occupancy or runway throughput limits;
- slot utilization;
- terminal or taxiway queues;
- schedule adherence;
- delay attribution;
- controller workload;
- official congestion status.

A later version may add richer operational context only when an appropriate source and evidence contract exist. It must not reinterpret the Stage 21 relative activity proxy as those missing concepts.

## 13. Pre-merge status

At this document commit, product and browser integration are still being validated. Final exact-head workflow identifiers belong in PR metadata after the documentation-complete head exists; mutating this document merely to record its own future SHA would recursively invalidate exact-head validation.

```text
STAGE_21_PRODUCT_IMPLEMENTATION=IMPLEMENTED_PREMERGE
STAGE_21_FRONTEND_INTEGRATION=IMPLEMENTED_PREMERGE
STAGE_21_OPENAPI_SURFACE=39_PUBLIC_OPERATIONS
STAGE_21_RESEARCH_ONLY_BOUNDARY=ENFORCED
STAGE_21_FINAL_EXACT_HEAD_CI=PENDING
STAGE_21_POST_MERGE_CI=PENDING
STAGE_21_AIRPORT_CONGESTION_INTELLIGENCE=IN_PROGRESS
```
