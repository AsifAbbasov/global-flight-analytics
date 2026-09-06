# Stage 17 Observed Interval Comparison Post-Merge Closure

Status: closed

```text
STAGE_17_OBSERVED_INTERVAL_COMPARISON=CLOSED
STAGE_17_MERGED_PR=157
STAGE_17_APPROVED_HEAD=0ed82c4233ae53bd62e9e124add12e8e2b97f700
STAGE_17_MAIN_SHA=566f3e676eca9c0edde2c94d151fdaba2812b007
STAGE_17_POST_MERGE_CI=PASS
STAGE_17_EVIDENCE_CLASS=OBSERVED_ENDPOINTS
STAGE_17_POSITION_INTERPOLATION=NONE
STAGE_17_PATH_DISTANCE_CLAIM=NONE
STAGE_17_ADDITIONAL_COST=0_RUB
STAGE_17_DOCUMENTATION=CLOSED
```

## Purpose of this closure document

Document 200 records the Stage 17 product need, architecture, evidence boundaries, rejected alternatives, real CI rejection, remediation history, regression protection, pre-merge validation and residual limitations as they existed before merge. It intentionally carried `STAGE_17_POST_MERGE_CI=PENDING` because post-merge evidence did not yet exist.

This document appends the post-merge truth instead of rewriting Document 200 after the fact. The history therefore remains chronological and auditable:

```text
Document 200
product design + implementation + real CI remediation + pre-merge validation
        ↓
exact-head squash merge of PR #157
        ↓
post-merge validation on resulting main SHA
        ↓
Document 201
post-merge closure evidence
```

The pre-merge document is preserved as a historical snapshot rather than being made to appear as though it knew future merge and CI results.

## What was delivered

Stage 17 adds Observed Interval Comparison to Historical Flight Replay. A user can select two persisted observations and compare only values supported by those persisted endpoints and the persisted samples inside the selected interval.

The delivered product behavior includes:

- Observation A / Observation B selection over persisted replay states;
- chronological normalization when the user selects the endpoints in reverse order;
- elapsed observed interval;
- endpoint great-circle displacement;
- inclusive evidence sample count;
- intermediate persisted sample count;
- largest internal observation gap;
- altitude delta only when both endpoint altitude values have supported evidence;
- velocity delta;
- vertical-rate delta;
- shortest angular heading change;
- ground/airborne state transition;
- honest unavailable state when fewer than two persisted observations exist.

## Why it was implemented

Before Stage 17, replay could inspect a single sample and compare only the current observation with its immediately previous persisted observation. That was insufficient for an arbitrary wider interval.

The user needed a way to compare two real historical endpoints without manually reading and subtracting values, but the feature had to preserve the strongest existing replay guarantee: missing evidence must not become a reconstructed route, fabricated intermediate position or invented flight phase.

The browser already possessed the required persisted replay states, so the product need was a comparison model and UI rather than a new provider, backend endpoint or database structure.

## Expected result and actual result

Expected result:

- arbitrary persisted endpoints can be compared directly;
- sparse intervals expose their evidence quality instead of appearing continuous;
- endpoint displacement is not mislabeled as travelled path distance;
- unsupported altitude evidence remains unavailable;
- no intermediate coordinate, route, phase or intent is inferred;
- the feature requires no paid service or new runtime infrastructure.

Actual result after merge matches that expectation. The merged implementation reuses the existing observed-change semantics, exposes interval-local evidence quality, passes the post-merge validation matrix, deploys successfully to Vercel and adds no paid/runtime infrastructure dependency.

## Exact merge evidence

User-approved exact PR head:

```text
PR=157
APPROVED_HEAD=0ed82c4233ae53bd62e9e124add12e8e2b97f700
MERGE_METHOD=SQUASH
EXPECTED_HEAD_GUARD=ENFORCED
```

The merge operation used `expected_head_sha`, so GitHub would reject the merge if the PR head changed after approval.

Resulting authoritative main revision:

```text
MAIN_SHA=566f3e676eca9c0edde2c94d151fdaba2812b007
```

GitHub reports PR #157 as merged and `main` points to this exact verified squash commit.

## Post-merge CI evidence

All expected push workflows on exact main SHA `566f3e676eca9c0edde2c94d151fdaba2812b007` completed successfully.

```text
Frontend CI #425
run=34043307643
result=SUCCESS

Backend CI #763
run=34043307600
result=SUCCESS

CodeQL #405
run=34043307583
result=SUCCESS

Playwright E2E #202
run=34043307626
result=SUCCESS

Vercel
result=SUCCESS
```

Backend CI #763 completed Backend Quality, PostgreSQL 16 Integration, Backend Race Safety, container build, non-root runtime verification, historical materializer verification, container health smoke and the final Backend CI gate successfully.

CodeQL #405 completed the repository security analysis successfully on the resulting main revision.

Playwright #202 completed the Chromium end-to-end suite and evidence upload successfully.

API Load Baseline did not run on this post-merge push because the existing workflow/event/path policy does not require it for this push shape. This is expected behavior, not a missing or failed check. The final merge-candidate head had already passed API Load Baseline #294 / run `34039054637`.

## Review and remediation history

Stage 17 has a real pre-merge CI remediation history and it remains preserved in Document 200 rather than being rewritten here.

The truthful sequence is:

1. Initial Stage 17 implementation head `27676696fccddf364ca5f804bc1d8fc2b5532a80` entered full validation.
2. Frontend CI #420 / run `34038594838` failed after ESLint and TypeScript passed, with 151 tests passing and 2 failing.
3. Root cause A: the new pure interval model was absent from the explicit `tsconfig.test.json` compilation scope.
4. Root cause B: one source-contract regex depended on one-line JSX whitespace.
5. Bad fixes were rejected: deleting the behavioral unit test, duplicating production math in tests, or rewriting production JSX solely to satisfy formatting-coupled regex behavior.
6. The test compilation scope was corrected to compile the real production model.
7. The semantic source-contract became whitespace-tolerant without weakening its required evidence wording.
8. Remediation validation head `21bd9474ac5bc1208251d078715122b2a29214ca` passed the full validation matrix.
9. Document 200 recorded the real failure → root cause → rejected fixes → remediation → residual limitations history.
10. Final merge-candidate head `0ed82c4233ae53bd62e9e124add12e8e2b97f700` passed the second complete validation cycle.
11. PR #157 was squash-merged only after explicit exact-head authorization.
12. The resulting `main` revision passed all expected post-merge workflows and Vercel deployment.
13. This document records the final post-merge closure.

No additional architecture defect was discovered after merge, so no post-merge remediation story is invented.

## Evidence boundary after closure

The merged feature preserves the following guarantees:

```text
EVIDENCE_CLASS=OBSERVED_ENDPOINTS
POSITION_INTERPOLATION=NONE
PATH_DISTANCE_CLAIM=NONE
GAPS_REMAIN_UNKNOWN=YES
UNSUPPORTED_ALTITUDE_DELTA=UNAVAILABLE
INTERVAL_DIRECTION=EARLIER_TO_LATER
```

Endpoint displacement remains great-circle distance between two persisted coordinates. It is not travelled path distance.

Heading change remains the shortest angular difference between endpoint headings. It is not cumulative turning through the interval.

Velocity and vertical-rate values are endpoint deltas. They do not establish acceleration or continuous vertical behavior between observations.

Sample count and largest gap describe the persisted evidence density inside the interval; they do not establish complete coverage or aviation-grade severity.

## Why the selected architecture remains correct

The final architecture reuses the existing `buildFlightReplayObservedChange` semantics for endpoint deltas and adds only interval-local sample-count and gap analysis.

This remains preferable because it:

- avoids a second altitude/displacement/heading implementation;
- avoids semantic drift between adjacent-sample and arbitrary-interval comparison;
- requires no new API contract;
- requires no provider change;
- requires no database migration;
- keeps sparse evidence visible;
- prevents route/path reconstruction from being smuggled into a comparison feature;
- remains testable as deterministic frontend logic;
- remains inside the 0 RUB development budget.

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

Stage 17 reuses existing persisted flight states, the existing frontend replay data path, existing observed-change semantics and existing MapLibre/Next.js infrastructure.

## Regression protection

The merged repository contains permanent coverage for:

- chronological normalization of reversed endpoint selection;
- comparison across non-adjacent persisted observations;
- elapsed interval duration;
- inclusive and intermediate sample counts;
- largest gap restricted to the selected interval;
- supported altitude delta;
- unavailable altitude remaining unavailable;
- velocity and vertical-rate endpoint deltas;
- shortest angular heading change;
- evidence and path-distance wording;
- honest single-sample unavailable state;
- browser visibility inside the existing aircraft-intelligence journey;
- integration without a new fetch/API path;
- real production model compilation into the test output;
- whitespace-tolerant semantic source-contract behavior;
- Document 200 pre-merge evidence semantics.

This closure adds a separate documentation contract so the exact merge revision, post-merge CI evidence, zero-budget boundary and residual limitations cannot silently disappear from future documentation changes.

## Residual limitations

Stage 17 is closed, but the following limits remain intentionally explicit:

- comparison is limited to observations actually persisted in the selected replay;
- FREE_V1 ingestion cadence can leave intervals sparse;
- endpoint displacement does not establish travelled path distance;
- endpoint heading difference does not establish cumulative turns;
- endpoint velocity delta does not establish acceleration between observations;
- vertical-rate delta does not reconstruct continuous climb or descent behavior;
- no intermediate coordinate, route, flight phase, aircraft intent or pilot behavior is inferred;
- largest gap exposes sparsity but does not create an aviation severity classification;
- sample count does not imply uniform temporal density;
- the feature is analytical replay evidence, not ATC-grade, navigation-grade or safety-critical data;
- Stage 17 does not add commercial historical coverage or denser paid aviation data.

These are evidence limitations, not reasons to fabricate richer frontend values.

## Operational consequences

There is no new production service, database migration, provider request volume, cache, persistent storage or runtime dependency to operate.

The main operational consequence is positive: the browser performs deterministic comparison over data it already loaded, while CI permanently guards the evidence semantics and test ownership boundaries that failed during initial validation.

## Future guard

Any future interval/replay metric must declare before implementation:

1. exact source fields available to the frontend;
2. observed / derived / projected evidence class;
3. whether the metric describes endpoints or the full interval;
4. missing-data behavior;
5. whether gaps change interpretation;
6. whether it assumes any unobserved path or position;
7. whether the UI wording could overstate evidence;
8. regression protection;
9. infrastructure impact;
10. monetary cost.

If a desired frontend capability requires data that the current provider/backend does not truthfully expose, it must not be fabricated. If the only defensible implementation requires a paid source or paid infrastructure, implementation stops at the zero-budget boundary until the user explicitly changes that constraint.

## Final status

```text
STAGE_17_IMPLEMENTATION=MERGED
STAGE_17_EXACT_HEAD_GUARD=PASS
STAGE_17_POST_MERGE_CI=PASS
STAGE_17_VERCEL=PASS
STAGE_17_EVIDENCE_CLASS=OBSERVED_ENDPOINTS
STAGE_17_POSITION_INTERPOLATION=NONE
STAGE_17_PATH_DISTANCE_CLAIM=NONE
STAGE_17_ADDITIONAL_COST=0_RUB
STAGE_17_DOCUMENTATION=CLOSED
STAGE_17_OBSERVED_INTERVAL_COMPARISON=CLOSED
```
