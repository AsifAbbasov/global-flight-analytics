# Stage 17 Observed Interval Comparison

Status: merge candidate pending final exact-head CI

```text
STAGE_17_OBSERVED_INTERVAL_COMPARISON=MERGE_CANDIDATE
STAGE_17_BASE_SHA=0fb818d23a60cd231f7ecbd290af2ccde03a488e
STAGE_17_PR=157
STAGE_17_INITIAL_HEAD=27676696fccddf364ca5f804bc1d8fc2b5532a80
STAGE_17_REMEDIATION_VALIDATION_SHA=21bd9474ac5bc1208251d078715122b2a29214ca
STAGE_17_EVIDENCE_CLASS=OBSERVED_ENDPOINTS
STAGE_17_POSITION_INTERPOLATION=NONE
STAGE_17_PATH_DISTANCE_CLAIM=NONE
STAGE_17_ADDITIONAL_COST=0_RUB
STAGE_17_INITIAL_PR_CI=FAILED_REMEDIATED
STAGE_17_REMEDIATION_CI=PASS
STAGE_17_DOCUMENTATION=PREMERGE_COMPLETE
STAGE_17_POST_MERGE_CI=PENDING
```

## Product need

Stage 15 made Historical Flight Replay trustworthy at the sample level. Stage 16 made navigation faithful to real elapsed observation time. The next user need is comparison: analysts need to select two persisted observations and understand how the recorded endpoints differ without manually reading two cards and subtracting values.

The comparison must remain useful even when the interval between selected observations is sparse. It therefore exposes both endpoint deltas and interval evidence quality.

## Why this feature exists

A replay user can already inspect one observation and the change from the immediately previous sample. That is insufficient for a wider observed interval such as the beginning and end of a climb, a long sparse segment, or two observations separated by several persisted points.

Stage 17 allows explicit two-endpoint comparison while preserving the central GFA guarantee: missing evidence is not silently converted into a continuous path.

## Expected product result

The user selects Observation A and Observation B from persisted replay states. The frontend normalizes the selection chronologically and shows only values derivable from those persisted endpoints and the persisted samples between them.

Expected visible outputs:

- earlier/later persisted timestamps;
- elapsed observed interval;
- endpoint great-circle displacement;
- observed sample count inside the selected interval;
- intermediate persisted sample count;
- largest persisted-sample gap inside the interval;
- altitude delta when both endpoints have supported altitude evidence;
- velocity delta;
- vertical-rate delta;
- shortest angular heading change;
- ground/airborne state transition.

## Problem

Without interval comparison, users either manually move between observations and remember values or overinterpret adjacent-sample change as if it described a wider time span.

Neither gives a concise, evidence-aware summary of an arbitrary pair of persisted observations.

## Root cause

The existing `buildFlightReplayObservedChange` intentionally compares only the immediately previous and current persisted samples. That narrow scope was correct for Stage 15 but does not represent an arbitrary user-selected interval.

The underlying data is already available in the browser; the missing capability is a comparison model and UI, not a new data source.

## Failure scenario

Assume two selected observations are ten minutes apart with only one intermediate persisted sample. A naive comparison could report straight-line endpoint distance as `distance travelled` or smooth the missing interval into a synthetic route. Both transform sparse evidence into a stronger claim than the data supports.

## Project impact

Without an explicit evidence-aware interval model:

- wider comparisons are cumbersome;
- endpoint displacement can be confused with route distance;
- sparse intervals can appear more complete than they are;
- future UI work may duplicate altitude or heading semantics inconsistently.

Severity: **P2 product semantics**. Stored evidence remains correct, but analytical UX lacks a safe arbitrary-interval comparison.

## Existing guarantees that must remain intact

```text
EVIDENCE_CLASS=OBSERVED
POSITION_INTERPOLATION=NONE
GAPS_REMAIN_UNKNOWN=YES
ENDPOINT_DISPLACEMENT_IS_NOT_PATH_DISTANCE=YES
UNSUPPORTED_ALTITUDE_DELTA=UNAVAILABLE
```

## Considered solutions

### Option A — calculate travelled distance by connecting all selected samples

Rejected. A polyline through sparse observations is still not the actual travelled path. Summing segments would create a stronger route-distance claim than persisted evidence supports.

### Option B — interpolate missing coordinates before comparison

Rejected. This would fabricate positions and violate the observed-only replay contract.

### Option C — create a new backend comparison endpoint

Rejected for Stage 17. The browser already has every persisted state required. A new endpoint would add API surface, server load and maintenance without adding evidence.

### Option D — reuse existing endpoint-change semantics and add interval-quality metadata

Selected. The interval model delegates endpoint altitude/displacement/velocity/heading/state semantics to the already-tested `buildFlightReplayObservedChange`, then adds only interval-local sample counts and gap analysis.

## Chosen architecture

```text
existing persisted replay points
        ↓
user selects A and B
        ↓
chronological normalization
        ↓
endpoint subset [earlier, later]
        ↓
existing buildFlightReplayObservedChange semantics
        +
existing buildFlightReplayGaps over original interval
        ↓
Observed Interval Comparison UI
```

This avoids a second implementation of altitude evidence, great-circle displacement or heading-change semantics.

## Why this solution was selected

It maximizes product value while minimizing semantic drift and architecture cost:

- no provider change;
- no backend endpoint;
- no database migration;
- no new dependency;
- no duplicate evidence math;
- no interpolation;
- interval sparsity remains visible;
- additional monetary cost remains 0 RUB.

## Implementation design

### Interval identity

The user selects two persisted observation indexes. Selection order does not change the result: the model normalizes the pair chronologically.

### Endpoint evidence

The normalized start and end observations are passed through the existing two-point observed-change implementation. This preserves the same altitude-status, great-circle, velocity, heading and ground-state rules already used by Stage 15.

### Interval evidence quality

The model counts persisted samples inside the selected inclusive interval and computes the largest exact elapsed gap between adjacent persisted samples within that interval. A ten-minute comparison can therefore visibly disclose that only two or three actual observations support it.

### React lifecycle

Selection state lives in a keyed inner component. The key includes trajectory identity plus first/last persisted point identity and point count. When replay evidence changes materially, React remounts selector state instead of using effect-driven reset logic.

## Adversarial scenario 1 — endpoint displacement is presented as travelled path

Risk: a future label changes `Endpoint displacement` to `Distance travelled`.

Why this is wrong: two persisted coordinates do not establish the path between them.

Guard:

```text
STAGE_17_PATH_DISTANCE_CLAIM=NONE
```

The UI explicitly states that endpoint displacement is great-circle distance between persisted coordinates, not travelled path distance.

## Adversarial scenario 2 — missing altitude becomes zero

Risk: an unavailable endpoint altitude is coerced to zero and produces a false altitude delta.

Guard: Stage 17 reuses existing observed-change altitude semantics. If either endpoint lacks supported altitude evidence, `altitudeDeltaM` remains `null` and UI displays `Unavailable`.

## Adversarial scenario 3 — reverse selection changes semantic direction unpredictably

Risk: selecting B before A produces negative elapsed time or reverses state-transition meaning.

Guard: selected indexes are normalized chronologically before calculation; UI documents earlier → later direction.

## Adversarial scenario 4 — wide interval hides sparse evidence

Risk: users see large endpoint deltas but cannot tell whether dozens of observations or only two observations support the interval.

Guard: UI exposes inclusive evidence-sample count, intermediate sample count and largest internal gap.

## Initial implementation review outcome

The first full PR validation cycle on exact head `27676696fccddf364ca5f804bc1d8fc2b5532a80` produced a real Frontend CI rejection. This history is intentionally preserved instead of rewritten as if the first attempt had passed.

Frontend CI #420 / run `34038594838` reached ESLint and TypeScript successfully, then failed the frontend contract-test step with **151 passing tests and 2 failures**. Backend CI #758, CodeQL #400 and API Load Baseline #290 completed successfully on that initial head. Playwright #197 was cancelled after later remediation commits superseded the head.

### Remediation record A — interval model absent from test compilation output

1. **Finding / symptom:** `flight-replay-interval-model.test.mjs` failed with `ERR_MODULE_NOT_FOUND` for `.test-dist/lib/replay/flight-replay-interval-model.js`.
2. **Root cause:** `apps/web/tsconfig.test.json` has an explicit `include` list. The new pure model was added to production TypeScript but not to the separate test-compilation contract.
3. **Failure scenario:** a new model can typecheck in the app while its unit test can never load the compiled artifact.
4. **Impact:** intended interval math regression coverage is absent even though the test file exists.
5. **Severity rationale:** P2 test-contract defect; no production runtime defect was proven, but merge could not proceed because permanent model coverage was not executable.
6. **Guarantee affected:** new evidence math must be exercised by deterministic unit tests before merge.
7. **Initial solution:** add the pure model and a direct unit test importing its `.test-dist` artifact.
8. **Why it looked acceptable:** existing replay model tests use the same compile-then-import test architecture.
9. **CI/review objection:** Frontend CI #420 demonstrated the new model was outside the explicit test compilation scope.
10. **Rejected fix:** remove the unit test or replace it only with source-text assertions. That would make CI green by deleting behavioral coverage.
11. **Second attack scenario:** copying the interval math into the test would allow test and production implementations to diverge while both remain green.
12. **Chosen remediation:** add `lib/replay/flight-replay-interval-model.ts` to `tsconfig.test.json` so the real production model is compiled into `.test-dist` and executed by the unit test.
13. **Why selected:** it repairs the test architecture at its ownership boundary and preserves direct behavioral coverage.
14. **Regression test:** the same `flight-replay-interval-model.test.mjs` now imports and executes the compiled production model.
15. **CI evidence:** Frontend CI #422 / run `34038754833` passed ESLint, TypeScript, all frontend contract/unit tests and production build.
16. **Residual limitation:** future new pure test-target modules using this explicit compilation scheme must also be added to `tsconfig.test.json`.
17. **Operational consequence:** none; test-only compilation scope changed, runtime bundle and infrastructure are unchanged.
18. **Final status:** REMEDIATED.
19. **Future guard:** never satisfy a missing compiled test artifact by weakening or deleting the behavioral test.

### Remediation record B — source contract depended on JSX line formatting

1. **Finding / symptom:** Stage 17 source-contract test failed to match the sentence declaring that no intermediate coordinate, route, phase or intent is synthesized.
2. **Root cause:** the assertion used a literal-space regex while Prettier/JSX source split `no` and `intermediate` across a newline.
3. **Failure scenario:** harmless formatting can fail an evidence contract even though rendered user semantics remain unchanged.
4. **Impact:** creates noisy CI and pressure to modify production copy solely to satisfy source formatting.
5. **Severity rationale:** P3 test-harness fragility; product behavior and evidence wording were already correct.
6. **Guarantee affected:** source contracts should protect semantics, not incidental whitespace layout.
7. **Initial solution:** match the full semantic phrase as a literal string-derived regex.
8. **Why it looked acceptable:** other short source markers exist on one line and are stable.
9. **CI/review objection:** Frontend CI #420 proved this phrase crosses JSX whitespace boundaries.
10. **Rejected fix:** rewrite production JSX onto one line only to satisfy the test.
11. **Second attack scenario:** future formatter changes could reintroduce the same false failure.
12. **Chosen remediation:** make only the semantic boundary whitespace-tolerant with `no\s+intermediate...` while retaining the exact meaningful words.
13. **Why selected:** it preserves the user-visible evidence statement and removes formatting coupling.
14. **Regression test:** `stage-17-observed-interval-contract.test.mjs` continues to require all evidence/path markers and the full semantic phrase.
15. **CI evidence:** Frontend CI #422 / run `34038754833` passed the corrected contract.
16. **Residual limitation:** source-text contracts remain appropriate only for stable semantic markers; interactive behavior stays covered by Playwright.
17. **Operational consequence:** none; no runtime/UI wording changed.
18. **Final status:** REMEDIATED.
19. **Future guard:** source-contract regexes spanning formatted JSX must tolerate whitespace without weakening semantic tokens.

## Remediation validation

Exact remediation validation head:

```text
21bd9474ac5bc1208251d078715122b2a29214ca
```

Full validation matrix:

```text
Frontend CI #422
run=34038754833
result=SUCCESS

Backend CI #760
run=34038754813
result=SUCCESS

CodeQL #402
run=34038754816
result=SUCCESS

API Load Baseline #292
run=34038754841
result=SUCCESS

Playwright E2E #199
run=34038754814
result=SUCCESS

Vercel preview
result=SUCCESS
```

No product architecture change was required by the CI rejection. Both remediations repaired test reachability/robustness while preserving the original Stage 17 evidence model.

## Trade-offs

- endpoint displacement is deliberately not travelled path distance;
- heading change is shortest angular endpoint difference, not accumulated turn;
- velocity and vertical-rate deltas compare endpoints only;
- sample count does not imply uniform temporal density;
- largest gap exposes sparsity but does not invent aviation severity thresholds;
- the feature does not reconstruct route, phase or intent.

## Zero-budget impact

```text
New paid aviation provider = NO
New backend endpoint        = NO
New database                = NO
New server                  = NO
New cache                   = NO
New persistent storage      = NO
New paid map service        = NO
New runtime dependency      = NO
Additional cost             = 0 RUB
```

## Regression protection

Permanent tests cover:

- chronological normalization of reversed A/B selection;
- endpoint comparison across non-adjacent persisted observations;
- elapsed interval duration;
- inclusive/intermediate sample counts;
- largest gap restricted to selected interval;
- supported altitude delta;
- unavailable altitude remains unavailable;
- velocity and vertical-rate endpoint deltas;
- shortest angular heading change;
- evidence/path-distance UI markers;
- single-sample honest unavailable state;
- browser visibility in existing aircraft intelligence journey;
- integration without a new fetch/API path;
- executable test compilation of the new interval model;
- whitespace-tolerant semantic source contract.

## CI / review evidence

Pull request:

```text
PR=157
BASE=0fb818d23a60cd231f7ecbd290af2ccde03a488e
INITIAL_HEAD=27676696fccddf364ca5f804bc1d8fc2b5532a80
INITIAL_FRONTEND_CI=FAILED
REMEDIATION_VALIDATION_HEAD=21bd9474ac5bc1208251d078715122b2a29214ca
REMEDIATION_CI=PASS
```

The documentation commit that records this remediation necessarily creates a newer PR head. Merge readiness therefore requires a second full exact-head CI cycle after this documentation update. The final exact head belongs in PR metadata and exact-head merge authorization rather than self-referencing its own commit SHA.

## Residual limitations

- comparison is limited to observations actually persisted in the selected replay;
- FREE_V1 ingestion cadence can leave intervals sparse;
- endpoint displacement does not establish travelled distance;
- endpoint heading difference does not establish cumulative turns;
- endpoint velocity delta does not describe acceleration between observations;
- no intermediate coordinate, route, phase or aircraft intent is inferred;
- comparison is analytical UI evidence, not ATC- or navigation-grade data.

## Expected completion criteria

```text
TWO_PERSISTED_OBSERVATIONS_REQUIRED=YES
REVERSED_SELECTION_NORMALIZED=YES
ENDPOINT_CHANGE_REUSES_EXISTING_MODEL=YES
INTERVAL_SAMPLE_COUNT_VISIBLE=YES
INTERVAL_LARGEST_GAP_VISIBLE=YES
POSITION_INTERPOLATION=NONE
PATH_DISTANCE_CLAIM=NONE
UNSUPPORTED_ALTITUDE_DELTA=UNAVAILABLE
ADDITIONAL_COST=0_RUB
INITIAL_PR_CI=FAILED_REMEDIATED
REMEDIATION_CI=PASS
REGRESSION_TESTS=PASS
DOCUMENTATION=PREMERGE_COMPLETE
FINAL_EXACT_HEAD_CI=PASS
POST_MERGE_CI=PASS
```

## Future guard

Any future interval metric must declare before implementation:

1. exact source fields;
2. whether it describes endpoints or the full interval;
3. observed / derived / projected evidence class;
4. missing-data behavior;
5. whether temporal gaps affect interpretation;
6. whether it assumes an unobserved path;
7. infrastructure and monetary cost;
8. regression protection.

A metric that requires unavailable frontend data or an unsupported reconstruction must not be added merely to make the UI look richer.
