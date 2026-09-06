# Stage 16 Time-Based Historical Replay Navigation

Status: merge candidate pending final exact-head CI

```text
STAGE_16_TIME_BASED_REPLAY_NAVIGATION=MERGE_CANDIDATE
STAGE_16_BASE_SHA=ceed3ecd31bd72d0de9491d23eda6bcb10cb2a98
STAGE_16_PR=155
STAGE_16_IMPLEMENTATION_VALIDATION_SHA=9601b4809e6e467e9820c6467b91086d59a0a355
STAGE_16_EVIDENCE_CLASS=OBSERVED
STAGE_16_TIME_CURSOR=REAL_ELAPSED_OBSERVATION_TIME
STAGE_16_POSITION_POLICY=LAST_PERSISTED_OBSERVATION
STAGE_16_POSITION_INTERPOLATION=NONE
STAGE_16_ADDITIONAL_COST=0_RUB
STAGE_16_INITIAL_PR_CI=PASS
STAGE_16_DOCUMENTATION=PREMERGE_COMPLETE
STAGE_16_POST_MERGE_CI=PENDING
```

## Product need

Stage 15 established trustworthy Historical Flight Replay from persisted Flight State observations. Its remaining navigation weakness was temporal: playback advanced by sample order, not by the actual elapsed time represented by `observed_at`.

That meant two adjacent samples separated by three seconds and two adjacent samples separated by five minutes consumed the same playback interval. The stored evidence was correct, but replay cadence did not communicate historical spacing faithfully.

Stage 16 makes elapsed observation time a first-class replay dimension without inventing aircraft positions.

## Why this feature exists

The product must answer two different questions simultaneously:

1. where was the aircraft actually observed?
2. where is the replay cursor in historical time?

When evidence is sparse, those are not the same thing.

A user must be able to move the replay cursor through an unobserved interval while the aircraft remains at the latest persisted observation until the timestamp of the next persisted observation is reached.

## Expected product result

Expected behavior:

```text
17:45:00 observed A                         17:52:30 observed B
        ●-------------------------------------------●
        ^                 ^                         ^
        |                 |                         |
   position A        time cursor in gap        position B
```

The cursor can move through the interval. The map does not create any coordinate between A and B.

User-visible capabilities:

- elapsed-time slider based on first/last persisted `observed_at`;
- cursor timestamp and elapsed-from-start display;
- exact-observation vs no-observation-at-cursor status;
- previous observation jump;
- next observation jump;
- start and end jumps;
- largest-gap midpoint jump for evidence inspection;
- 1x / 5x / 10x / 30x time-scale playback;
- existing `replay_observation=<state.id>` deep links preserved;
- existing sample selector retained as a direct jump to a persisted observation.

## Problem

Before Stage 16, replay state was primarily a `cursorIndex` and timer advancement meant one persisted sample per tick.

Example:

```text
sample A = 17:45:00
sample B = 17:45:03
sample C = 17:52:30
```

Sample stepping visually treats A→B and B→C as equal playback intervals even though the historical durations are radically different.

## Root cause

Stage 15 deliberately optimized first for evidence safety. A discrete sample cursor was the simplest way to guarantee that no intermediate position could be fabricated.

Once observed-only/no-interpolation semantics were protected by tests, the next safe evolution was to separate time navigation from position evidence.

## Failure scenario

If sample-index playback remained the only playback clock, a user could interpret evenly timed UI transitions as evidence of evenly spaced observations.

This would not fabricate coordinates, but it would compress evidence gaps and weaken temporal fidelity.

## Project impact

Primary impact:

- gap duration is harder to perceive during playback;
- replay speed means samples-per-second rather than historical time scale;
- long evidence gaps are visually compressed;
- replay is less useful for evidence-quality analysis.

Severity: **P2 product-semantics limitation**. Stored truth remains intact, but presentation underrepresents temporal sparsity.

## Existing guarantees that must remain intact

```text
EVIDENCE_CLASS=OBSERVED
POSITION_INTERPOLATION=NONE
GAPS_REMAIN_UNKNOWN=YES
SHARED_OBSERVATION_ID=DURABLE_STATE_ID
```

A continuous time cursor must never imply a continuous observed path.

## Considered solutions

### Option A — interpolate coordinates between observations

Rejected.

Smooth motion would create positions that were never persisted or observed.

### Option B — request denser historical data from a new provider

Rejected under current constraints.

It introduces an external dependency, can cross the zero-budget boundary, and still does not remove the need to represent missing evidence honestly.

### Option C — create a new backend replay endpoint that emits synthetic time frames

Rejected for Stage 16.

The existing flight-state endpoint already exposes every timestamp and persisted state required for truthful time navigation. A new backend contract would add complexity without adding evidence.

### Option D — continuous elapsed-time cursor + discrete observed position

Selected.

Time can move continuously; position resolves only to the latest persisted observation with `observed_at <= cursor_time`.

## Chosen architecture

```text
persisted flight_states
        ↓
existing GET /api/v1/flights/{flightID}/states
        ↓
sorted persisted replay points
        ↓
continuous elapsed-time cursor
        ↓
latest persisted observation <= cursor timestamp
        ↓
MapLibre Point evidence only
```

Two state concepts are intentionally separated:

```text
TIME CURSOR
continuous historical elapsed time

POSITION CURSOR
latest persisted observation at-or-before time cursor
```

This separation is the central Stage 16 architectural decision.

## Why this solution was selected

It improves product fidelity without weakening source truth.

It also preserves the project constraints:

- existing API reused;
- no database migration;
- no provider change;
- no server-side replay engine;
- no new persistent state;
- durable observation links preserved;
- missing evidence remains visible;
- additional monetary cost remains 0 RUB.

## Implementation design

### Replay time model

Deterministic helpers now cover:

- observation timestamp → elapsed cursor seconds;
- elapsed cursor → held persisted observation;
- exact persisted observation detection;
- seconds since last observation;
- previous/next observation cursor positions;
- largest-gap midpoint navigation;
- elapsed-time cursor advancement;
- deep-link observation ID → elapsed cursor restoration.

### Playback clock

The browser clock uses a 250 ms UI tick.

At time scale `S`:

```text
cursor advance per tick = 0.25 seconds * S
```

Supported scales:

```text
1x
5x
10x
30x
```

Only time advances continuously. Position changes only when another persisted observation timestamp is crossed.

### Gap behavior

Example:

```text
cursor                 = 18:03:00
previous observation   = 18:01:00
next observation       = 18:06:00
```

Resolved state:

```text
map position           = persisted position at 18:01:00
seconds since observed = 120
exact observation      = false
```

No coordinate is computed for 18:03:00.

## Adversarial scenario 1 — continuous slider is mistaken for continuous evidence

Risk: a future implementation could move the aircraft marker smoothly with the time slider.

Guard:

```text
STAGE_16_POSITION_POLICY=LAST_PERSISTED_OBSERVATION
STAGE_16_POSITION_INTERPOLATION=NONE
```

The UI explicitly reports `No observation at cursor` while the cursor is inside a gap.

## Adversarial scenario 2 — timer silently regresses to sample-per-tick behavior

Risk: reusing the old discrete `advanceFlightReplayCursor` would restore Stage 15 sample-clock semantics.

Guard: source contracts require `advanceFlightReplayTimeCursor`, a fixed elapsed tick and multiplication by replay speed.

## Adversarial scenario 3 — deep-link identity is replaced by time-only identity

Risk: introducing a new time query parameter would weaken durable observation identity and create two competing share contracts.

Chosen behavior: retain `replay_observation=<state.id>` and resolve that persisted ID to its elapsed cursor offset during initialization.

## Initial implementation review outcome

The first Stage 16 architecture reached full PR CI without an architectural or test rejection.

No historical rejection is invented for this stage.

This means there is no legitimate `initial solution rejected → remediation` story to record for the first validation cycle. The relevant engineering evidence is that the initial continuous-time/discrete-position separation passed lint, type checking, contracts, production build, CodeQL, API load baseline, backend repository guards and Chromium E2E unchanged.

## Trade-offs

- sparse evidence intentionally produces long periods where the aircraft marker does not move;
- browser timers can be throttled in background tabs;
- cursor time is UI navigation state, not a scientific timing instrument;
- largest-gap navigation points to the interval midpoint for inspection, not to an aircraft position;
- Stage 16 does not improve source sampling density.

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

- elapsed-time cursor resolution;
- held previous observation inside a gap;
- exact persisted observation detection;
- previous/next observation jumps;
- largest-gap midpoint calculation;
- exact observed-span playback completion;
- 1x/5x/10x/30x speed contract;
- deep-link observation restoration into elapsed time;
- no coordinate interpolation contract;
- Point-based replay map representation;
- browser visibility of the time-navigation workspace;
- documentation completeness and zero-budget markers.

## CI / review evidence

Pull request:

```text
PR=155
IMPLEMENTATION_VALIDATION_HEAD=9601b4809e6e467e9820c6467b91086d59a0a355
BASE=ceed3ecd31bd72d0de9491d23eda6bcb10cb2a98
```

First complete validation cycle on that implementation head:

```text
Frontend CI #414
run=34036243832
result=SUCCESS

Backend CI #752
run=34036243790
result=SUCCESS

CodeQL #394
run=34036243838
result=SUCCESS

API Load Baseline #286
run=34036243839
result=SUCCESS

Playwright E2E #191
run=34036243846
result=SUCCESS

Vercel preview
result=SUCCESS
```

The documentation update that records this evidence necessarily creates a newer PR head. Therefore merge readiness requires a second full exact-head CI cycle after this documentation commit. The final exact head is authoritative in PR metadata and final merge authorization, not self-embedded into the commit that would need to contain its own SHA.

## Residual limitations

- replay can only navigate evidence that was actually persisted;
- FREE_V1 ingestion cadence can leave historical replay sparse;
- missing observations cannot be reconstructed truthfully;
- background-tab timer throttling can reduce wall-clock playback precision;
- replay is not ATC-grade or navigation-grade;
- time navigation does not establish travelled path between samples;
- time navigation does not establish flight phase, route or intent;
- no commercial historical coverage is added.

## Expected completion criteria

Pre-merge requirements:

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
INITIAL_PR_CI=PASS
DOCUMENTATION=PREMERGE_COMPLETE
FINAL_EXACT_HEAD_CI=PASS
```

Post-merge closure requirement:

```text
POST_MERGE_CI=PASS
```

## Future guard

Any later replay feature proposing smoother historical movement must declare before implementation:

1. source data;
2. observed / derived / projected evidence class;
3. interpolation policy;
4. missing-data behavior;
5. user-visible uncertainty semantics;
6. infrastructure impact;
7. monetary cost;
8. regression protection.

A continuous path or smooth marker movement cannot be described as observed unless the intermediate positions actually exist as persisted observations.
