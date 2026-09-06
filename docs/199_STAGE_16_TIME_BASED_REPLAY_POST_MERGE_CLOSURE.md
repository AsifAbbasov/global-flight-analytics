# Stage 16 Time-Based Replay Post-Merge Closure

Status: closed

```text
STAGE_16_TIME_BASED_REPLAY_NAVIGATION=CLOSED
STAGE_16_MERGED_PR=155
STAGE_16_APPROVED_HEAD=ff5b3f0af6e146e4dd48502798a3a83c334f42be
STAGE_16_MAIN_SHA=b07e2a506455b695944c57e82327e7b06f0c0edc
STAGE_16_POST_MERGE_CI=PASS
STAGE_16_EVIDENCE_CLASS=OBSERVED
STAGE_16_TIME_CURSOR=REAL_ELAPSED_OBSERVATION_TIME
STAGE_16_POSITION_POLICY=LAST_PERSISTED_OBSERVATION
STAGE_16_POSITION_INTERPOLATION=NONE
STAGE_16_ADDITIONAL_COST=0_RUB
STAGE_16_DOCUMENTATION=CLOSED
```

## Purpose of this closure document

Document 198 records the Stage 16 design, implementation rationale, pre-merge evidence boundary and exact PR validation history as they existed before merge. It intentionally ended with post-merge verification pending because that evidence did not yet exist.

This document appends the post-merge truth instead of rewriting Document 198 after the fact. The engineering history therefore remains chronological and auditable:

```text
Document 198
pre-merge architecture + validation snapshot
        ↓
exact-head squash merge of PR #155
        ↓
post-merge validation on new main
        ↓
Document 199
closure evidence
```

This avoids a retrospective edit that would make a pre-merge document appear to have known future merge and CI results.

## What was delivered

Stage 16 changed Historical Flight Replay from sample-index playback semantics to real elapsed-time navigation using persisted `observed_at` timestamps.

The frontend can now move a time cursor through the actual observed time span while map position remains bound to the latest persisted observation at-or-before the cursor time.

The delivered product behavior includes:

- real elapsed-time replay cursor;
- explicit cursor timestamp;
- explicit `No observation at cursor` state inside evidence gaps;
- previous and next persisted observation jumps;
- start and end jumps;
- largest-gap inspection jump;
- elapsed-time playback at 1x / 5x / 10x / 30x;
- durable `replay_observation=<state.id>` identity preservation;
- existing discrete MapLibre observed-point evidence;
- no coordinate interpolation or synthesized historical positions.

## Why it was implemented

Stage 15 protected the primary truth boundary by advancing replay only across persisted observations. That was intentionally safe, but sample-index playback treated short and long historical intervals as equivalent UI steps.

Stage 16 separates two different concepts:

```text
TIME CURSOR
continuous elapsed historical time

POSITION EVIDENCE
latest persisted observation at-or-before that time
```

This makes temporal sparsity visible without weakening the observed-only evidence model.

## Expected result and actual result

Expected result:

- historical time spacing becomes visible and navigable;
- gaps remain explicit;
- aircraft position never becomes synthetic merely because the time cursor is continuous;
- no new paid data or infrastructure is required.

Actual result after merge matches that expectation. The merged implementation preserves observed-only position semantics, passes the full repository post-merge validation matrix, and adds no paid/runtime infrastructure dependency.

## Exact merge evidence

User-approved exact PR head:

```text
PR=155
APPROVED_HEAD=ff5b3f0af6e146e4dd48502798a3a83c334f42be
MERGE_METHOD=SQUASH
```

The merge operation used an exact-head guard. GitHub accepted the merge only while the PR head still matched the approved SHA.

Resulting main commit:

```text
MAIN_SHA=b07e2a506455b695944c57e82327e7b06f0c0edc
```

The resulting commit is the squash merge of PR #155 and is the authoritative Stage 16 merged implementation revision.

## Post-merge CI evidence

All expected push workflows on exact main SHA `b07e2a506455b695944c57e82327e7b06f0c0edc` completed successfully.

```text
Frontend CI #417
run=34037166018
result=SUCCESS

Backend CI #755
run=34037166006
result=SUCCESS

CodeQL #397
run=34037166092
result=SUCCESS

Playwright E2E #194
run=34037166020
result=SUCCESS

Vercel
result=SUCCESS
```

Backend #755 additionally completed PostgreSQL 16 integration, backend quality, race safety, container build, non-root runtime verification, historical materializer verification, container health smoke, and the Backend CI Gate successfully.

CodeQL #397 completed both Go and JavaScript/TypeScript analyses plus the CodeQL Security Gate successfully.

Playwright #194 completed the Chromium end-to-end suite and evidence upload successfully.

API Load Baseline did not run on this push because the existing workflow/event/path policy does not trigger it for this post-merge push. This is expected behavior, not a missing or failed check. The PR itself had already passed API Load Baseline #288 / run `34036556415` on the final merge-candidate head.

## Review and remediation history

No architecture or regression failure was discovered during the final exact-head PR validation or post-merge validation.

Therefore there is no legitimate rejected-solution remediation story to invent for this closure.

The truthful history is:

1. Stage 16 selected continuous elapsed time + discrete observed position.
2. The first implementation validation cycle passed.
3. Pre-merge documentation was updated with that evidence.
4. The final exact-head PR cycle passed.
5. PR #155 was squash-merged using the explicitly approved exact head.
6. The post-merge exact-main validation cycle passed.
7. This document records that closure evidence.

## Evidence boundary after closure

The merged feature does not change the Stage 15/16 source-truth rules:

```text
EVIDENCE_CLASS=OBSERVED
POSITION_INTERPOLATION=NONE
GAPS_REMAIN_UNKNOWN=YES
POSITION_AT_TIME=LAST_PERSISTED_OBSERVATION_AT_OR_BEFORE_CURSOR
```

A time cursor inside a gap is navigation state, not aircraft-position evidence.

No coordinate between persisted observations is created, estimated, or represented as observed.

## Zero-budget and infrastructure impact

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

Stage 16 reuses existing persisted flight states, the existing flight-state endpoint, existing frontend computation, and existing MapLibre point layers.

## Regression protection

The merged repository contains permanent coverage for:

- elapsed-time cursor resolution;
- held previous persisted observation inside an evidence gap;
- exact observation detection;
- previous/next observation navigation;
- largest-gap midpoint navigation;
- elapsed-time playback completion;
- 1x/5x/10x/30x speed semantics;
- durable observation deep-link restoration;
- no coordinate interpolation;
- Point-based replay map evidence;
- browser visibility of time navigation;
- Stage 16 documentation/evidence semantics.

This closure adds a separate documentation contract so the exact merge revision, post-merge CI evidence, zero-budget boundary and residual limitations cannot silently disappear from future documentation changes.

## Residual limitations

Stage 16 is closed, but the following limits remain intentionally explicit:

- replay can only navigate observations that were actually persisted;
- the FREE_V1 ingestion cadence can leave historical replay sparse;
- missing observations cannot be reconstructed truthfully;
- the browser timer can be throttled in background tabs, so replay wall-clock pacing is not a scientific timing instrument;
- replay is not ATC-grade, navigation-grade or safety-critical evidence;
- elapsed-time navigation does not establish travelled path between persisted positions;
- replay does not establish flight phase, route, intent or pilot behavior from missing data;
- Stage 16 does not add commercial historical coverage or denser paid aviation data.

These are data/evidence limitations, not reasons to fabricate frontend values.

## Future guard

Any future replay feature must continue to declare before implementation:

1. source data;
2. evidence class: observed / derived / projected;
3. interpolation policy;
4. missing-data behavior;
5. user-visible uncertainty semantics;
6. infrastructure impact;
7. monetary cost;
8. regression protection.

If the desired frontend capability requires data that the current sources cannot truthfully supply, it must not be fabricated. If the only valid implementation requires paid data or infrastructure, that feature must stop at the zero-budget boundary until the user explicitly changes the budget constraint.

## Final status

```text
STAGE_16_IMPLEMENTATION=MERGED
STAGE_16_EXACT_HEAD_GUARD=PASS
STAGE_16_POST_MERGE_CI=PASS
STAGE_16_VERCEL=PASS
STAGE_16_POSITION_INTERPOLATION=NONE
STAGE_16_ADDITIONAL_COST=0_RUB
STAGE_16_DOCUMENTATION=CLOSED
STAGE_16_TIME_BASED_REPLAY_NAVIGATION=CLOSED
```
