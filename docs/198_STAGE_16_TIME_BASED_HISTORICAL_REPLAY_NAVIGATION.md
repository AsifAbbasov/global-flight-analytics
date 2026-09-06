# Stage 16 Time-Based Historical Replay Navigation

Status: PR candidate

```text
STAGE_16_TIME_BASED_REPLAY_NAVIGATION=PR_CANDIDATE
STAGE_16_BASE_SHA=ceed3ecd31bd72d0de9491d23eda6bcb10cb2a98
STAGE_16_EVIDENCE_CLASS=OBSERVED
STAGE_16_TIME_CURSOR=REAL_ELAPSED_OBSERVATION_TIME
STAGE_16_POSITION_POLICY=LAST_PERSISTED_OBSERVATION
STAGE_16_POSITION_INTERPOLATION=NONE
STAGE_16_ADDITIONAL_COST=0_RUB
```

## Product need

Stage 15 established a trustworthy Historical Flight Replay built from persisted Flight State observations. The remaining navigation weakness is that the primary playback mechanic advances by array/sample order rather than by the actual elapsed time represented by `observed_at` timestamps.

A user can therefore see three samples but cannot directly understand whether those samples are separated by seconds or by several minutes without reading the evidence-gap panel. This weakens the analytical value of replay even though the underlying timestamps already exist.

Stage 16 makes elapsed time a first-class replay navigation dimension without creating any new aircraft positions.

## Why this feature exists

The product needs to answer two different questions at the same time:

1. **where was the aircraft actually observed?**
2. **where is the replay cursor in real historical time?**

Those are not the same thing when observations are sparse.

The user must be able to move the historical cursor through a five-minute evidence gap while the map honestly keeps the aircraft at the last known persisted observation until another persisted observation timestamp is reached.

## Expected product result

The expected user-visible behavior is:

```text
17:45:00 observed A
        ↓
        ●──────────────────────────────────────────────●
        ↑                                              ↑
 cursor can move through elapsed time          17:52:30 observed B
 aircraft remains at A until B is reached
```

The interface adds:

- a real elapsed-time slider from the first to the last persisted `observed_at`;
- explicit cursor timestamp and elapsed-from-start display;
- exact-observation vs no-observation-at-cursor status;
- previous observation jump;
- next observation jump;
- start and end jumps;
- largest-gap midpoint jump for evidence inspection;
- time-scaled playback at 1x, 5x, 10x and 30x;
- preservation of the existing durable `replay_observation=<state.id>` deep-link behavior.

The existing sample selector remains useful as a direct jump to a persisted observation, but it no longer defines playback time semantics.

## Problem

The Stage 15 workspace stores replay progress as a `cursorIndex` and advances it once per timer tick.

That means:

```text
sample 1 -> sample 2 -> sample 3
```

is treated as evenly spaced playback even if real timestamps are:

```text
17:45:00 -> 17:45:03 -> 17:52:30
```

The visual playback cadence therefore does not represent the historical elapsed-time cadence.

## Root cause

The original Stage 15 goal was to establish the evidence boundary first: observed samples only, no interpolation. A discrete sample index was the safest initial cursor because it could not synthesize intermediate positions.

Once that guarantee was established and protected by tests, the remaining limitation became explicit: safe sample stepping is not equivalent to true time navigation.

## Failure scenario

Assume persisted observations:

```text
A = 18:01:00
B = 18:01:10
C = 18:06:10
```

With sample-index playback, A→B and B→C consume the same UI interval even though the second historical interval is thirty times larger.

A user can incorrectly interpret the playback cadence as evidence that observations were evenly distributed.

## Project impact

The issue does not corrupt stored data, but it weakens analytical fidelity:

- historical elapsed time is visually compressed into sample count;
- large gaps are less intuitive to inspect;
- playback speed means “samples per second” rather than historical time scale;
- the replay is less useful for evidence-quality analysis.

Severity classification: **P2 product-semantics limitation**. It does not create false positions because Stage 15 forbids interpolation, but it can misrepresent temporal spacing.

## Existing guarantees that must remain intact

Stage 16 must preserve all Stage 15 evidence guarantees:

```text
EVIDENCE_CLASS=OBSERVED
POSITION_INTERPOLATION=NONE
GAPS_REMAIN_UNKNOWN=YES
SHARED_OBSERVATION_ID=DURABLE_STATE_ID
```

In particular, a continuous time cursor must never be converted into a continuous aircraft position.

## Considered solutions

### Option A — interpolate aircraft coordinates between observations

Rejected.

This would make time playback visually smooth but would create aircraft positions that were never persisted or observed.

### Option B — request denser historical data from a new provider

Rejected under current constraints.

It introduces a new external dependency and can cross the zero-budget boundary. It also does not solve the fundamental need to represent missing evidence honestly.

### Option C — add a new backend replay endpoint that emits synthetic time frames

Rejected for this stage.

The existing `GET /api/v1/flights/{flightID}/states` already contains every datum required for truthful time navigation. A new backend API would add complexity without new evidence.

### Option D — keep position discrete, make only the cursor continuous in elapsed time

Selected.

The cursor can move through real historical seconds while the map resolves the latest persisted observation whose timestamp is less than or equal to the cursor timestamp.

## Chosen architecture

```text
persisted flight_states
        ↓
existing flight states endpoint
        ↓
sorted persisted replay points
        ↓
continuous elapsed-time cursor
        ↓
latest observed sample <= cursor time
        ↓
MapLibre observed point only
```

The state model is split into two concepts:

```text
TIME CURSOR
continuous elapsed historical time

POSITION CURSOR
latest persisted observation at-or-before time cursor
```

This separation is the core Stage 16 architectural decision.

## Why this solution was selected

It provides real historical time semantics while preserving source truth.

It also:

- reuses the existing API and persisted data;
- requires no database migration;
- requires no provider change;
- requires no server-side replay engine;
- preserves durable observation deep links;
- makes gaps visible instead of hiding them;
- keeps the additional monetary cost at 0 RUB.

## Implementation design

### Replay model

The replay model adds deterministic helpers for:

- observation timestamp → elapsed cursor seconds;
- elapsed cursor → current held observation;
- exact observation detection;
- seconds since last persisted observation;
- previous/next observation jumps;
- largest-gap midpoint navigation;
- elapsed-time playback advancement;
- deep-link observation → elapsed cursor restoration.

### Playback clock

Playback uses a 250 ms UI tick.

At speed `S`, each tick advances:

```text
0.25 seconds * S
```

Supported time scales:

```text
1x
5x
10x
30x
```

The timer changes the time cursor only. Position changes only when a persisted observation timestamp is crossed.

### Map behavior inside a gap

If the cursor is between two observations:

```text
cursor = 18:03:00
previous observation = 18:01:00
next observation = 18:06:00
```

then:

```text
current map position = position observed at 18:01:00
seconds since observed = 120
exact observation at cursor = false
```

No coordinate is computed for 18:03:00.

## Adversarial scenario 1 — continuous slider implies continuous evidence

A reviewer or future developer could assume that a continuous slider justifies continuous map movement.

Guard:

```text
STAGE_16_POSITION_POLICY=LAST_PERSISTED_OBSERVATION
STAGE_16_POSITION_INTERPOLATION=NONE
```

The UI explicitly says “No observation at cursor” while inside a gap.

## Adversarial scenario 2 — time playback accidentally becomes sample playback again

If the timer calls the old `advanceFlightReplayCursor`, Stage 16 silently regresses to sample-per-tick semantics.

Guard: source contracts require `advanceFlightReplayTimeCursor`, a fixed elapsed tick and multiplication by replay speed.

## Adversarial scenario 3 — deep links lose observation identity

A time-only query parameter would make old shared observation links less durable and would create a second identity mechanism.

Chosen behavior: retain `replay_observation=<state.id>`. The observation ID resolves to its elapsed-time position when replay initializes.

## Trade-offs

- sparse data still produces long periods where the map marker does not move;
- that behavior is intentionally truthful rather than visually smooth;
- the browser timer is not a scientific timing instrument and can be throttled in background tabs;
- replay cursor time is UI state, not a server-side historical clock;
- largest-gap navigation jumps to the midpoint of the longest observed interval for inspection, not to a claimed aircraft position.

## Zero-budget impact

```text
New paid aviation provider = NO
New backend endpoint        = NO
New database                = NO
New server                  = NO
New cache                   = NO
New persistent storage      = NO
New map provider            = NO
New runtime dependency      = NO
Additional cost             = 0 RUB
```

## Regression protection

Permanent tests cover:

- real elapsed-time cursor resolution;
- position held at the previous persisted observation inside a gap;
- exact persisted observation detection;
- previous/next observation navigation;
- largest-gap midpoint calculation;
- elapsed-time playback completion at the exact observed-span end;
- 1x/5x/10x/30x time-scale contract;
- deep-link observation restoration into time cursor;
- source-level prohibition against coordinate interpolation;
- MapLibre replay remaining Point-based rather than a synthetic replay line.

## CI / review evidence

Pending first exact-head PR validation.

This section must be updated before Stage 16 is considered merge-ready with:

- PR number;
- exact final head SHA;
- Frontend CI run;
- Backend CI run;
- CodeQL run;
- API Load Baseline run when triggered by PR policy;
- Playwright E2E run;
- Vercel preview status;
- any real remediation history discovered by CI/review.

No review rejection will be invented if the first architecture passes unchanged.

## Residual limitations

- replay can only navigate evidence that is actually persisted;
- FREE_V1 ingestion cadence can leave replay sparse;
- missing historical observations cannot be reconstructed truthfully;
- browser background throttling can make wall-clock playback less precise;
- the feature is not ATC-grade or navigation-grade;
- time navigation does not establish travelled path between observations;
- time navigation does not establish flight phase, intent or route;
- commercial historical coverage is not added by this stage.

## Expected completion criteria

Stage 16 can be marked CLOSED only when all of the following are true:

```text
TIME_CURSOR_USES_OBSERVED_AT=YES
PLAYBACK_USES_ELAPSED_TIME=YES
POSITION_INTERPOLATION=NONE
GAP_STATE_VISIBLE=YES
PREVIOUS_NEXT_OBSERVATION_JUMPS=YES
LARGEST_GAP_NAVIGATION=YES
DEEP_LINK_IDENTITY_PRESERVED=YES
ADDITIONAL_COST=0_RUB
REGRESSION_TESTS=PASS
PR_CI=PASS
POST_MERGE_CI=PASS
DOCUMENTATION=COMPLETE
```

## Future guard

Any later feature that proposes smoother historical movement must first declare:

1. evidence source;
2. whether positions are observed, derived or projected;
3. interpolation policy;
4. missing-data behavior;
5. user-visible uncertainty semantics;
6. infrastructure impact;
7. monetary cost.

A continuous historical line or marker motion cannot be introduced under the label “observed” unless the intermediate positions actually exist as persisted observations.
