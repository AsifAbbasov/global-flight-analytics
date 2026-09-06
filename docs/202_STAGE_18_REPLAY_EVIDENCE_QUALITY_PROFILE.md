# Stage 18 Replay Evidence Quality Profile

Status: pre-merge validation passed; final exact-head verification pending

```text
STAGE_18_REPLAY_EVIDENCE_QUALITY=IN_PROGRESS
STAGE_18_BASE_MAIN=cfd005f3e6259736f589c9b209820cc53c38c3cd
STAGE_18_IMPLEMENTATION_HEAD=8048e37747802ac65e3ceae6142331256d6b70aa
STAGE_18_FIRST_VALIDATION_HEAD=cb93a00d46f874e0cd0cf48fe00f855948e65e7a
STAGE_18_FIRST_FRONTEND_CI=FAIL
STAGE_18_REMEDIATION_HEAD=f814848601d5d78c736b5c7cd73e1e5735bf5dea
STAGE_18_REMEDIATION_VALIDATION_HEAD=9ca39c6f2b170ae260e5f67d21265c5fcf285c42
STAGE_18_REMEDIATION_VALIDATION=PASS
STAGE_18_EVIDENCE_PROFILE=DESCRIPTIVE_ONLY
STAGE_18_SYNTHETIC_QUALITY_SCORE=NONE
STAGE_18_POSITION_INTERPOLATION=NONE
STAGE_18_NEW_BACKEND_DATA=NONE
STAGE_18_ADDITIONAL_COST=0_RUB
STAGE_18_PRE_MERGE_CI=PASS
STAGE_18_FINAL_EXACT_HEAD_CI=PENDING
STAGE_18_POST_MERGE_CI=PENDING
```

## Product need

Historical Flight Replay already exposes persisted observations, gaps, replay analytics, time navigation and arbitrary observed interval comparison. The remaining trust problem is that a user can still see a replay with several analytical values without immediately understanding how sparse or concentrated the underlying sampling is.

Stage 18 adds a visible evidence-quality profile for the replay itself. The profile does not attempt to decide whether the data is universally "good" or "bad". It describes the persisted evidence so that the user can make that judgment in context.

## Why this feature exists

A replay with ten observations across ten seconds and a replay with ten observations across two hours both contain ten persisted samples, but they do not represent the same evidence density. Likewise, two datasets can have the same median gap while one contains a single very large unobserved interval.

The product therefore needs sampling-shape information that is visible next to replay navigation:

- how many observations exist;
- how many inter-observation intervals can actually be measured;
- mean gap;
- median gap;
- P90 gap;
- largest gap;
- how much of the observed span the largest gap occupies;
- observation density per hour;
- altitude evidence coverage;
- exact endpoints of the largest unobserved interval.

These values are all derivable from data that is already present in the browser.

## Problem

The replay UI previously exposed individual gap information and aggregate replay analytics, but the user had to mentally combine several fields to understand the sampling evidence profile.

There was also a product-risk temptation to collapse this into a single "quality score". A score would be visually convenient but would require calibration and thresholds that the current free-data foundation does not justify.

## Root cause

The current replay model intentionally preserves raw persisted observations and unobserved gaps. That architecture is correct, but it means evidence quality is multidimensional:

- sample count alone is insufficient;
- median gap alone can hide an outlier gap;
- largest gap alone does not show how dominant that gap is relative to the full span;
- altitude support is independent from temporal sampling density.

No single currently justified scalar represents all of those dimensions.

## Failure scenario

Without an explicit evidence profile, a user can see replay analytics and incorrectly assume that a dense-looking UI implies dense historical evidence.

Example:

```text
observation 1 ---- 60 s ---- observation 2 ---------------- 600 s ---------------- observation 3
```

A median-like summary alone can hide how much of the interval is dominated by one missing period. A synthetic quality score can make the situation worse by compressing distinct evidence limitations into an unexplained number.

## Impact on the project

If sparse evidence is visually overclaimed, Global Flight Analytics loses one of its strongest product properties: explicit separation between observed evidence and missing data.

The risk is not only cosmetic. A user may overinterpret replay-derived altitude, velocity or route context as though the aircraft had been continuously observed.

## Existing guarantee that must remain true

```text
OBSERVED_IS_OBSERVED
MISSING_IS_MISSING
POSITION_INTERPOLATION=NONE
UNOBSERVED_INTERVALS_ARE_NOT_RECONSTRUCTED
```

Stage 18 must make the evidence boundary easier to understand without weakening it.

## Considered solutions

### Option A — synthetic 0–100 replay quality score

Rejected.

A numeric score would require defensible weighting and calibration across gap size, sample density, altitude coverage and provider behavior. The project does not currently possess a validated calibration dataset for such a score.

### Option B — fixed labels such as GOOD / FAIR / POOR

Rejected.

Any fixed thresholds would be arbitrary unless calibrated against a documented downstream purpose. A sampling profile suitable for a visual demo may still be unsuitable for trajectory research, so a universal label would overstate what the evidence proves.

### Option C — request denser or commercial historical aviation data

Rejected under the current development constraint.

The feature does not require additional data to describe the evidence already stored. Buying denser history would violate the current `0 RUB` development budget and is unnecessary for this stage.

### Option D — new backend evidence-quality endpoint

Rejected.

The browser already has the exact replay states required for this calculation. Adding an endpoint would duplicate data movement and introduce another API contract without adding evidence.

### Option E — descriptive evidence profile derived client-side

Selected.

This option exposes factual dimensions without pretending that an uncalibrated score is objective.

## Chosen architecture

```text
existing persisted FlightReplay points
        ↓
existing buildFlightReplayGaps(...)
        ↓
existing buildFlightReplayAnalyticsSummary(...)
        ↓
flight-replay-evidence-quality-model.ts
        ↓
FlightReplayEvidenceQuality
        ↓
existing Historical replay time navigation
```

The new model deliberately reuses the existing replay gap and analytics calculations instead of introducing a second implementation of sample-gap or altitude-coverage semantics.

## Metrics and semantics

### Observations

Exact persisted sample count.

### Sampling intervals

Number of measurable intervals between adjacent persisted observations.

For `N` persisted chronological observations this is normally `N - 1`.

### Mean gap

Arithmetic mean of the actual adjacent persisted-observation time gaps.

No intermediate samples are synthesized.

### Median gap

Reuses the existing replay analytics median-gap semantics.

### P90 gap

Nearest-rank 90th percentile of the observed adjacent gap durations.

This is descriptive. It is not a service-level objective and not a quality threshold.

### Largest gap

Largest exact elapsed time between two adjacent persisted observations.

### Largest-gap share

```text
largest gap / total observed span
```

expressed as a rounded percentage.

This shows how dominant the worst missing interval is relative to the full replay span.

### Observation density

```text
persisted sample count / observed span in hours
```

This is reported only when a positive observed span exists.

It is a descriptive rate, not a statement about provider polling frequency and not a guarantee that observations are uniformly spaced.

### Altitude evidence

Reuses existing replay analytics altitude-coverage semantics. Unsupported altitude remains unsupported.

### Largest unobserved interval

The UI exposes the exact persisted timestamps that bound the largest measured gap and explicitly states that no persisted position exists between them.

## Why no score is produced

The absence of a score is an intentional product decision, not an unfinished calculation.

A single score would require answers to questions that are not currently justified by evidence:

- how many seconds of gap should be considered acceptable;
- whether altitude support should be weighted more or less than temporal density;
- how the score should differ by replay duration;
- whether the score represents visualization quality, analytical quality or aviation operational quality;
- how provider-specific sampling behavior should be normalized.

Until those questions can be calibrated against a documented purpose, the professional behavior is to expose the raw evidence profile.

## Single-sample behavior

One persisted observation cannot establish an inter-sample gap distribution or observation density over a positive span.

The UI therefore reports:

```text
sampling intervals = 0
mean gap = unavailable
median gap = unavailable
P90 gap = unavailable
largest gap = unavailable
largest-gap share = unavailable
observation density = unavailable
```

It also explains why these values cannot be measured.

This behavior is covered in the browser fixture instead of mutating the fixture to fabricate extra observations.

## Adversarial scenario 1 — same sample count, very different coverage

Two replays each contain 20 points. One spans two minutes; the other spans four hours.

A sample-count-only quality signal would make them appear equivalent.

Stage 18 exposes span-sensitive density and gap distribution instead.

## Adversarial scenario 2 — median hides one dominant gap

Most observations are ten seconds apart, but one gap lasts twenty minutes.

Median gap can remain small while the replay contains a major evidence hole.

Stage 18 exposes P90, largest gap, largest-gap share and the exact bounding timestamps.

## Adversarial scenario 3 — attractive score hides calibration assumptions

A future developer introduces `qualityScore=84` based on undocumented thresholds.

The number appears authoritative even though no validation dataset defines what 84 means.

Permanent source-contract tests explicitly forbid synthetic `qualityScore` / `confidenceScore` logic in the Stage 18 model and require the UI marker:

```text
data-flight-replay-evidence-score=none
```

## Adversarial scenario 4 — client feature silently adds a new data dependency

A developer attempts to call another API to improve the evidence profile.

The current feature does not need this. Source contracts reject `fetch` / `axios` in the evidence-quality model and time-navigation integration.

## Expected product result

The user should be able to look at Historical Replay and understand not only what observations exist but also how the evidence is distributed in time.

The UI should make sparse data look sparse and dense data look dense without pretending that the project has an aviation-grade calibration model.

## Expected engineering result

- no backend endpoint;
- no PostgreSQL migration;
- no provider change;
- no ingestion change;
- no new infrastructure;
- no new paid service;
- no position interpolation;
- no synthetic replay-quality score;
- pure deterministic frontend model;
- real production model included in the frontend unit-test compilation scope;
- source contracts preserve evidence semantics;
- browser test preserves the honest single-sample state.

## Regression protection

Stage 18 adds:

- `apps/web/tests/flight-replay-evidence-quality-model.test.mjs`;
- `apps/web/tests/stage-18-replay-evidence-quality-contract.test.mjs`;
- browser assertions inside the existing aircraft intelligence Playwright journey;
- explicit inclusion of the real model in `apps/web/tsconfig.test.json`.

The unit tests verify multi-sample calculations and single-sample uncertainty.

The source contract verifies:

- reuse of existing replay analytics/gap semantics;
- no secondary network path;
- no synthetic quality/confidence score;
- no undocumented good/bad thresholds;
- descriptive-only UI markers;
- integration inside existing time navigation;
- production-model compilation in `.test-dist`.

## Initial implementation history

The initial Stage 18 implementation was created from exact base main:

```text
cfd005f3e6259736f589c9b209820cc53c38c3cd
```

Implementation head before documentation:

```text
8048e37747802ac65e3ceae6142331256d6b70aa
```

Initial PR validation head:

```text
cb93a00d46f874e0cd0cf48fe00f855948e65e7a
```

## First CI rejection and real remediation history

The first full validation cycle produced a real Frontend CI failure. This is preserved instead of being rewritten as though the first implementation had passed.

```text
Frontend CI #428
run=34045471852
head=cb93a00d46f874e0cd0cf48fe00f855948e65e7a
ESLint=PASS
TypeScript=PASS
Frontend contract tests=FAIL
Production build=SKIPPED
Tests=173
Pass=172
Fail=1
```

The failing test was:

```text
Stage 18 UI is descriptive only and exposes no synthetic quality grade
```

### Root cause of the rejection

The product UI already contained the required semantic disclosure:

```text
These values describe the persisted sample set only. They are not calibrated aviation
quality grades and must not be interpreted as ATC-, navigation- or safety-grade coverage.
```

The source-contract test incorrectly searched for the formatting-sensitive literal sequence:

```text
not calibrated aviation quality grades
```

Because JSX wrapped the sentence across a newline between `aviation` and `quality`, the contract failed even though the rendered semantic wording was present. The failure was therefore a test-contract defect, not an evidence-model or product-UI defect.

### Failure scenario

A source contract tied to one physical line layout can fail after harmless formatting, Prettier output or JSX wrapping even when the required user-visible evidence disclosure is unchanged.

This creates false-negative CI and encourages developers to change production markup merely to satisfy test formatting.

### Bad fixes rejected

The following responses were rejected:

1. rewriting the production UI sentence onto one source line only to satisfy the regex;
2. deleting the wording assertion;
3. weakening the contract to check only for a generic word such as `quality`;
4. disabling or bypassing the failed test.

All four would either couple production formatting to test mechanics or weaken the evidence guarantee.

### Chosen remediation

The source contract was changed to preserve the semantic phrase while allowing ordinary source whitespace:

```text
/not calibrated aviation\s+quality grades/
```

The production model, UI wording, metrics and evidence semantics were not changed.

Remediation code head:

```text
f814848601d5d78c736b5c7cd73e1e5735bf5dea
```

### Why this remediation was selected

The requirement is semantic continuity of the disclosure, not a particular JSX line break. A whitespace-tolerant expression keeps the meaningful phrase mandatory while making the contract robust to source formatting.

### Second attack scenario

A future formatter may replace one newline with several spaces or vice versa. The new contract continues to require the full ordered phrase and therefore survives formatting changes without allowing the disclosure to disappear.

### Regression protection after remediation

The Stage 18 source-contract test still requires all evidence markers, metric labels and the full `not calibrated aviation ... quality grades` phrase. Only whitespace representation became flexible.

No fictional product-review rejection is recorded. The only rejection is the real Frontend CI #428 source-contract failure above.

## Successful remediation validation

The complete repository validation matrix was rerun after the remediation and documentation-contract update on exact head:

```text
9ca39c6f2b170ae260e5f67d21265c5fcf285c42
```

All required PR workflows completed successfully:

```text
Frontend CI #431
run=34045730187
result=SUCCESS

Backend CI #769
run=34045730140
result=SUCCESS

CodeQL #411
run=34045730178
result=SUCCESS

API Load Baseline #299
run=34045729973
result=SUCCESS

Playwright E2E #208
run=34045730064
result=SUCCESS

Vercel preview
result=SUCCESS
```

Frontend #431 proves that ESLint, TypeScript, the full frontend contract suite and production build all pass after the whitespace-tolerant remediation.

Backend #769 proves that repository-wide backend quality, PostgreSQL integration, race safety, container build, non-root runtime verification, historical materializer verification and container health smoke remain green even though Stage 18 is frontend-only.

CodeQL #411 completed successfully for Go and JavaScript/TypeScript. API Load #299 preserved the existing performance baseline. Playwright #208 completed the Chromium end-to-end suite and evidence upload successfully, including the new honest single-sample replay-evidence profile assertions. Vercel preview for the same exact head completed successfully.

No second product or architecture defect was discovered in this cycle, so no additional remediation story is invented.

## Why another exact-head cycle is still required

Recording the successful remediation matrix changes the documentation commit and therefore changes the PR head. A commit cannot contain a truthful assertion of its own future CI run IDs without creating a self-referential commit chain.

The repository therefore uses this two-layer rule:

1. Document 202 records the completed pre-merge validation and real remediation history on `9ca39c6f...`.
2. The resulting documentation head must pass one final complete exact-head CI/Vercel cycle without further code changes.
3. That final exact merge-candidate head and its independent validation matrix are recorded in PR metadata, which does not change the commit SHA.

This preserves both auditable documentation and an exact-head merge guard without an infinite sequence of evidence-only commits.

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

## Residual limitations

Stage 18 does not improve the underlying historical data density. It only describes persisted evidence more clearly.

Residual limitations remain:

- FREE_V1 ingestion cadence can produce sparse historical replay data;
- observation density does not mean observations are uniformly distributed;
- P90 and largest-gap statistics can be unstable for very small sample sets;
- the profile is limited to observations actually persisted by GFA;
- no route is reconstructed inside gaps;
- no aircraft intent or flight phase is inferred;
- no ATC-, navigation- or safety-grade quality claim is made;
- no commercial historical coverage is added;
- altitude coverage describes supported altitude evidence only and does not validate altitude accuracy.

## Future guard

Any future change that wants to convert this descriptive profile into a score or grade must first provide:

1. a clearly defined downstream purpose;
2. documented metric weighting;
3. justified thresholds;
4. a calibration/validation dataset;
5. regression tests for score stability and boundary behavior;
6. explicit documentation of known false-positive and false-negative cases.

Without that evidence, the profile must remain descriptive-only.

If the only defensible implementation of a future evidence-quality feature requires paid data or paid infrastructure, development of that feature must stop and the budget boundary must be reported before implementation.

## Final pre-merge closure requirement

The completed remediation validation satisfies the pre-merge quality requirements. The evidence-recording commit must now pass a final no-change exact-head cycle:

```text
ESLINT=PASS
TYPESCRIPT=PASS
FRONTEND_CONTRACTS=PASS
PRODUCTION_BUILD=PASS
PLAYWRIGHT=PASS
CODEQL=PASS
BACKEND_GUARDS=PASS
API_LOAD_BASELINE=PASS
VERCEL=PASS
DOCUMENTATION=COMPLETE
FINAL_EXACT_HEAD_CI=PASS
```

Until that final exact-head cycle completes, PR #159 is not merge-ready.