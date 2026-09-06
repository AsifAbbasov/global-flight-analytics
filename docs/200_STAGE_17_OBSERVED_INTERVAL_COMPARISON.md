# Stage 17 Observed Interval Comparison

Status: implemented, pending full PR validation

```text
STAGE_17_OBSERVED_INTERVAL_COMPARISON=IMPLEMENTED_PENDING_CI
STAGE_17_BASE_SHA=0fb818d23a60cd231f7ecbd290af2ccde03a488e
STAGE_17_EVIDENCE_CLASS=OBSERVED_ENDPOINTS
STAGE_17_POSITION_INTERPOLATION=NONE
STAGE_17_PATH_DISTANCE_CLAIM=NONE
STAGE_17_ADDITIONAL_COST=0_RUB
STAGE_17_INITIAL_PR_CI=PENDING
STAGE_17_POST_MERGE_CI=PENDING
```

## Product need

Stage 15 made Historical Flight Replay trustworthy at the sample level. Stage 16 made navigation faithful to real elapsed observation time. The next user need is comparison: analysts need to select two persisted observations and understand how the recorded endpoints differ without manually reading two cards and subtracting values.

The comparison must remain useful even when the interval between the selected observations is sparse. It therefore has to expose both endpoint deltas and interval evidence quality.

## Why this feature exists

A replay user can already inspect one observation and the change from the immediately previous sample. That is insufficient for a wider observed interval such as the beginning and end of a climb, a long sparse segment, or two observations separated by several persisted points.

Stage 17 allows explicit two-endpoint comparison while preserving the central GFA guarantee: missing evidence is not silently converted into a continuous path.

## Expected product result

The user selects Observation A and Observation B from persisted replay states. The frontend normalizes the selection chronologically and shows only values that can be derived from those persisted endpoints and the persisted samples between them.

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

Without an interval comparison, users have two weak choices:

1. manually move between observations and remember values;
2. overinterpret adjacent-sample change as if it described a wider time span.

Neither gives a concise, evidence-aware summary of an arbitrary pair of persisted observations.

## Root cause

The existing `buildFlightReplayObservedChange` intentionally compares only the immediately previous and current persisted samples. That narrow scope was correct for Stage 15 but does not represent an arbitrary user-selected interval.

The underlying data is already available in the browser; the missing capability is a comparison model and UI, not a new data source.

## Failure scenario

Assume two selected observations are ten minutes apart with only one intermediate persisted sample.

A naive comparison could report the straight-line distance between the endpoints as `distance travelled`, or could smooth the missing interval into a synthetic route. Both would transform sparse evidence into a stronger claim than the data supports.

## Project impact

Without an explicit evidence-aware interval model:

- wider comparisons are cumbersome;
- users may confuse endpoint displacement with route distance;
- sparse intervals can appear more complete than they are;
- future UI work may duplicate altitude or heading semantics inconsistently.

Severity: **P2 product semantics**. Stored evidence remains correct, but the analytical UX lacks a safe arbitrary-interval comparison.

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

Rejected.

A polyline through sparse observations is still not the actual travelled path. Summing line segments would create a stronger route-distance claim than the persisted observations justify.

### Option B — interpolate missing coordinates before comparison

Rejected.

This would fabricate positions and violate the observed-only replay contract.

### Option C — create a new backend comparison endpoint

Rejected for Stage 17.

The browser already has every persisted state needed for the comparison. A new endpoint would add API surface, server load and maintenance without adding evidence.

### Option D — reuse existing endpoint-change semantics and add interval-quality metadata

Selected.

The new interval model delegates endpoint altitude/displacement/velocity/heading/state semantics to the already-tested `buildFlightReplayObservedChange` model, then adds only interval-local sample counts and gap analysis.

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

This deliberately avoids a second independent implementation of altitude evidence, great-circle displacement or heading-change semantics.

## Why this solution was selected

It maximizes product value while minimizing semantic drift and architecture cost.

Benefits:

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

The model counts persisted samples inside the selected inclusive interval and computes the largest exact elapsed gap between adjacent persisted samples within that interval.

This lets the UI show that a ten-minute comparison may contain only two or three actual observations.

### React lifecycle

Selection state lives in a keyed inner component. The key includes trajectory identity plus first/last persisted point identity and point count. When replay evidence changes materially, React remounts the selector state instead of using effect-driven reset logic.

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

Guard: Stage 17 reuses the existing observed-change altitude semantics. If either endpoint lacks supported altitude evidence, `altitudeDeltaM` remains `null` and the UI displays `Unavailable`.

## Adversarial scenario 3 — reverse selection changes semantic direction unpredictably

Risk: selecting B before A could produce negative elapsed time or reverse state-transition meaning.

Guard: the model normalizes selected indexes chronologically before calculation and documents the earlier → later direction in the UI.

## Adversarial scenario 4 — wide interval hides sparse evidence

Risk: users see large endpoint deltas but cannot tell whether dozens of observations or only two observations support the interval.

Guard: the UI always exposes inclusive evidence-sample count, intermediate sample count and largest internal gap.

## Initial implementation review outcome

No historical review rejection is claimed at this point. The implementation has not yet completed its first full PR CI cycle.

If CI or review rejects an assumption, the exact failure, root cause, rejected remediation and final fix must be appended here rather than rewritten as if the first design had always been correct.

## Trade-offs

- endpoint displacement is deliberately not travelled path distance;
- heading change is the shortest angular endpoint difference, not accumulated turn;
- velocity and vertical-rate deltas compare endpoints only;
- sample count does not imply uniform temporal density;
- largest gap exposes sparsity but does not classify it with unsupported aviation severity thresholds;
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
- largest gap restricted to the selected interval;
- supported altitude delta;
- unavailable altitude remains unavailable;
- velocity and vertical-rate endpoint deltas;
- shortest angular heading change;
- evidence/path-distance UI markers;
- single-sample honest unavailable state;
- browser visibility in the existing aircraft intelligence journey;
- integration without a new fetch/API path.

## CI / review evidence

First PR validation cycle: **PENDING**.

The final merge-candidate document must record the exact PR number, implementation-validation SHA and workflow run IDs after the first complete validation cycle. A second exact-head cycle is required after that evidence is committed.

## Residual limitations

- comparison is limited to observations already persisted in the selected replay;
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
REGRESSION_TESTS=PASS
INITIAL_PR_CI=PASS
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
