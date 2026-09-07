# Stage 21 — Airport Congestion Intelligence Post-Merge Closure

Status: CLOSED

```text
STAGE_21_AIRPORT_CONGESTION_INTELLIGENCE=CLOSED
STAGE_21_MERGED_PR=164
STAGE_21_APPROVED_HEAD=f95c88aeb60b0b4e6c2e0cf7bff83697f2a9ed7f
STAGE_21_MAIN_SHA=63ff7c04cce26b73664dbecd307a50d9d6cdea04
STAGE_21_PREMERGE_EXACT_HEAD_CI=PASS
STAGE_21_PREMERGE_VERCEL=PASS
STAGE_21_POST_MERGE_CI=PASS
STAGE_21_POST_MERGE_VERCEL=PASS
STAGE_21_BACKEND_INITIAL_POSTMERGE_ATTEMPT=TIMEOUT_CANCELLED
STAGE_21_BACKEND_RETRY_ATTEMPT=PASS
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
```

## 1. Closure purpose

Document 207 is the immutable Stage 21 pre-merge engineering-history record. It intentionally preserves the analytical boundary, missing-data policy, real validation failures, remediation path and the fact that merge/post-merge evidence did not yet exist when that document was committed.

This document appends the actual exact-head merge and canonical-main evidence. Document 207 is not rewritten retroactively.

## 2. Product scope closed by Stage 21

Stage 21 adds Airport Congestion Intelligence as a bounded research-only relative observed-activity proxy over the existing Airport Intelligence completed-day history.

The production path is:

```text
existing persisted Airport Intelligence completed-day observations
        ↓
existing GetHistory production composition
        ↓
requested-window evidence accounting
        ↓
baseline median + prior observed peak comparison
        ↓
GET /api/v1/airports/{icao}/intelligence/congestion
        ↓
validated frontend parser + existing TanStack Query domain
        ↓
Unified Airport Analytics Workspace
        ↓
Activity pressure profile
```

No new provider, ingestion path, PostgreSQL table, migration, cache service, server or paid infrastructure was introduced.

## 3. Analytical and evidence contract

For a requested completed-day window, Stage 21 defines:

```text
current = latest expected completed UTC day, only when that day is observed
baseline = median movements/hour across prior observed completed-day windows
prior_peak = maximum movements/hour across prior observed completed-day windows
current_to_baseline_ratio = current / baseline, when baseline > 0
current_to_prior_peak_ratio = current / prior_peak, when prior_peak > 0
congestion_score = min(1, current_to_prior_peak_ratio)
```

The uncapped current/prior-peak ratio remains available separately so a new observed activity peak retains its magnitude even when the bounded display score reaches 1.

Missing evidence is not converted to zero. The requested production window is preserved independently of the compressed observed-history span so leading, internal and trailing gaps remain part of evidence coverage. When the latest expected completed day is absent, an older observed day is not silently substituted as current.

## 4. Interpretation boundary

The feature is not an operational airport-capacity or delay model.

```text
AIRPORT_CAPACITY_MODEL=NONE
RUNWAY_OCCUPANCY_MODEL=NONE
QUEUE_MODEL=NONE
SLOT_MODEL=NONE
SCHEDULE_ADHERENCE_MODEL=NONE
DELAY_INFERENCE=NONE
DELAY_CAUSE_INFERENCE=NONE
OFFICIAL_CONGESTION_CLAIM=NONE
CONTROLLER_WORKLOAD_CLAIM=NONE
LOW_MEDIUM_HIGH_CONGESTION_THRESHOLDS=NONE
```

The score describes relative observed movement intensity against the same airport's prior observed evidence. It does not establish declared capacity, runway utilization limits, queues, slot utilization, schedule adherence, delay attribution or official congestion status.

## 5. Public API and frontend closure

Stage 21 introduced:

```text
GET /api/v1/airports/{icao}/intelligence/congestion
```

The public API inventory therefore moved from 38 operations to 39 operations:

```text
39 operations = 38 GET + 1 POST
```

The canonical OpenAPI contract, embedded API document and generated TypeScript client were synchronized with the new operation. The frontend extends the existing Airport Intelligence workspace rather than creating a parallel product flow.

The Activity pressure profile preserves backend `*_known` semantics and exposes score, ratios, current/baseline/prior-peak movement intensity, new-peak evidence, observed/expected/gap/trailing-gap counts, evidence coverage/support and server-published scope guards without browser-side operational thresholds.

## 6. Preserved pre-merge validation history

Stage 21 does not rewrite its real development history into a fictitious first-pass success. Document 207 preserves the actual integration failures and remediations, including:

- the public OpenAPI operation inventory initially still expecting 38 operations after the 39th route was registered;
- the congestion DTO initially missing from the generic backend success-response union;
- the Playwright mock public path inventory initially missing the new route;
- touched Go files requiring `gofmt`;
- the embedded OpenAPI operation-count test still expecting 38 operations;
- a frontend source-contract assertion being more wording-specific than the actual research-only warning.

Those failures were remediated by synchronizing contracts, formatting, generated surfaces and semantic test expectations. No capacity model, delay inference, operational threshold or unsupported aviation claim was introduced to satisfy CI.

## 7. Final pre-merge exact-head evidence

The merge-authorized exact Stage 21 head was:

```text
f95c88aeb60b0b4e6c2e0cf7bff83697f2a9ed7f
```

All six GitHub validation workflows passed on that exact head:

```text
OpenAPI Contract #133 / run 34080720938 = SUCCESS
Frontend CI #498 / run 34080721020 = SUCCESS
Backend CI #836 / run 34080720950 = SUCCESS
API Load Baseline #360 / run 34080720971 = SUCCESS
CodeQL #478 / run 34080720936 = SUCCESS
Playwright E2E #269 / run 34080720937 = SUCCESS
```

Vercel also reported `Deployment has completed` for the same exact head.

```text
STAGE_21_GITHUB_EXACT_HEAD_VALIDATION=6_OF_6_PASS
STAGE_21_VERCEL_EXACT_HEAD_VALIDATION=PASS
```

## 8. Exact-head merge authorization and result

The user authorized a squash merge tied specifically to:

```text
APPROVED_HEAD=f95c88aeb60b0b4e6c2e0cf7bff83697f2a9ed7f
MERGE_METHOD=SQUASH
```

GitHub accepted the merge with the expected-head guard and produced:

```text
MERGED_PR=164
MAIN_SHA=63ff7c04cce26b73664dbecd307a50d9d6cdea04
PARENT_MAIN=bdf4810e558bafe1b027a42947c0bf937f2348d3
```

The resulting canonical main commit is GitHub-verified.

## 9. Canonical post-merge validation

The Stage 21 feature merge SHA is:

```text
63ff7c04cce26b73664dbecd307a50d9d6cdea04
```

The canonical post-merge workflow matrix is:

```text
OpenAPI Contract #134 / run 34083139299 = SUCCESS
Frontend CI #499 / run 34083139329 = SUCCESS
Backend CI #837 / run 34083139296 / attempt 2 = SUCCESS
API Load Baseline #361 / run 34083139369 = SUCCESS
CodeQL #479 / run 34083139308 = SUCCESS
Playwright E2E #270 / run 34083139400 = SUCCESS
```

Vercel succeeded on the same canonical merge SHA:

```text
STAGE_21_POST_MERGE_VERCEL=PASS
VERCEL_DEPLOYMENT=AagXsgugSre4HQp2RY4hwyqJ6srG
```

## 10. Backend post-merge retry history

Backend CI #837 preserves an important infrastructure retry rather than being represented as first-pass success.

Attempt 1 reached successful Backend Quality, Backend Race Safety and PostgreSQL 16 Integration. Its Backend Container job then exceeded the configured 20-minute job budget while executing `Build backend container image`. GitHub terminated that job as `cancelled`, and the aggregate Backend CI Gate consequently reported failure.

```text
BACKEND_RUN=34083139296
BACKEND_ATTEMPT_1_CONTAINER=TIMEOUT_CANCELLED
BACKEND_ATTEMPT_1_GATE=FAILURE
SOURCE_SHA_CHANGED_FOR_RETRY=NO
```

Only the failed/cancelled backend workflow job was retried; no source-code commit was introduced to manufacture a new exact head. GitHub recorded run attempt 2 on the same canonical SHA.

Attempt 2 passed the container build and all subsequent runtime checks:

```text
BACKEND_ATTEMPT_2_BUILD_CONTAINER=SUCCESS
BACKEND_ATTEMPT_2_NON_ROOT_RUNTIME=SUCCESS
BACKEND_ATTEMPT_2_MATERIALIZER=SUCCESS
BACKEND_ATTEMPT_2_CONTAINER_HEALTH_SMOKE=SUCCESS
BACKEND_ATTEMPT_2_WORKFLOW=SUCCESS
```

The successful same-SHA retry supports an infrastructure-timeout interpretation for attempt 1; it does not erase the cancelled first attempt.

## 11. Regression protection

Permanent Stage 21 backend/frontend/contract/browser tests introduced with the feature continue to protect the analytical and presentation behavior.

This post-merge closure adds a separate permanent documentation contract protecting:

- merged PR number;
- merge-authorized exact head;
- resulting canonical main SHA;
- exact pre-merge workflow identifiers;
- exact post-merge workflow identifiers;
- Vercel post-merge success;
- the cancelled first Backend #837 attempt and successful same-SHA retry;
- preservation of Document 207 as historical pre-merge evidence;
- relative-observed-activity semantics;
- no airport-capacity, runway-occupancy, queue, slot or delay model;
- no official congestion claim or invented LOW/MEDIUM/HIGH thresholds;
- zero-budget/no-new-data boundary.

## 12. Infrastructure and cost closure

```text
NEW_PROVIDER=NO
NEW_INGESTION_PATH=NO
NEW_DATABASE_TABLE=NO
NEW_MIGRATION=NO
NEW_CACHE_SERVICE=NO
NEW_SERVER=NO
NEW_PAID_SERVICE=NO
ADDITIONAL_COST=0_RUB
```

Stage 21 therefore closes its product increment without changing the project's zero-budget infrastructure boundary.

## 13. Residual limitations

Stage 21 cannot answer questions requiring evidence the project does not possess, including:

- declared airport capacity;
- runway occupancy or runway throughput limits;
- terminal, taxiway or departure queues;
- slot utilization;
- schedule adherence;
- delay magnitude or delay attribution;
- controller workload;
- official operational congestion status.

Future work may add richer operational context only when an appropriate source and explicit evidence contract exist. Stage 21's relative observed-activity proxy must not be reinterpreted as those missing concepts.

## 14. Final closure state

```text
STAGE_21_PRODUCT_IMPLEMENTATION=MERGED
STAGE_21_PREMERGE_EXACT_HEAD_CI=PASS
STAGE_21_POST_MERGE_FEATURE_SHA_CI=PASS
STAGE_21_POST_MERGE_FEATURE_SHA_VERCEL=PASS
STAGE_21_DOCUMENTATION=CLOSED
STAGE_21_AIRPORT_CONGESTION_INTELLIGENCE=CLOSED
```

Stage 21 product evidence is closed by this record. Document 207 remains the historical pre-merge engineering record; Document 210 is the canonical post-merge closure record once this documentation change is merged and its resulting canonical main SHA is independently verified under the repository's normal closure governance.
