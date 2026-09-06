# Stage 18 Replay Evidence Quality Profile Post-Merge Closure

Status: closed

```text
STAGE_18_REPLAY_EVIDENCE_QUALITY=CLOSED
STAGE_18_MERGED_PR=159
STAGE_18_APPROVED_HEAD=691d69b0241d1370ac901e1790d8a7736f8c33d8
STAGE_18_MAIN_SHA=2e2a542e7c970441074ddc6c67aa7548ce9b91b3
STAGE_18_POST_MERGE_CI=PASS
STAGE_18_EVIDENCE_PROFILE=DESCRIPTIVE_ONLY
STAGE_18_SYNTHETIC_QUALITY_SCORE=NONE
STAGE_18_POSITION_INTERPOLATION=NONE
STAGE_18_NEW_BACKEND_DATA=NONE
STAGE_18_ADDITIONAL_COST=0_RUB
STAGE_18_DOCUMENTATION=CLOSED
```

## Purpose of this closure document

Document 202 records the Stage 18 product need, architecture, evidence boundaries, rejected alternatives, real CI rejection, remediation history, regression protection and successful pre-merge remediation validation. It intentionally preserved future facts as pending while they did not yet exist.

This document appends the post-merge truth rather than rewriting Document 202 after the fact. The auditable sequence is therefore:

```text
Document 202
product design + implementation + real CI rejection + remediation + pre-merge evidence
        ↓
final exact-head validation of PR #159
        ↓
exact-head squash merge of PR #159
        ↓
post-merge validation on resulting main SHA
        ↓
Document 203
post-merge closure evidence
```

This preserves chronological evidence instead of making the pre-merge document appear to know future merge and CI outcomes.

## What was delivered

Stage 18 adds a visible Replay Evidence Quality Profile to Historical Flight Replay. The feature describes the persisted evidence already present in the browser rather than creating a synthetic quality verdict.

The delivered profile exposes:

- persisted observation count;
- measurable inter-observation interval count;
- arithmetic mean gap;
- median gap;
- nearest-rank P90 gap;
- largest exact observed gap;
- largest-gap share of the total observed span;
- observation density per hour when a positive observed span exists;
- altitude evidence coverage using the existing replay analytics semantics;
- exact persisted timestamps bounding the largest unobserved interval;
- an honest single-sample state when gap distribution and observation density cannot be measured.

The feature deliberately does not expose a 0-100 quality score, GOOD/FAIR/POOR grade, ATC/navigation/safety classification, synthetic intermediate position or reconstructed path.

## Why it was implemented

Before Stage 18, Historical Replay exposed persisted observations, individual gaps, replay analytics, time navigation and arbitrary observed interval comparison. Those capabilities were useful, but a user still had to mentally combine several fields to understand how sparse or concentrated the underlying historical evidence was.

A replay with ten observations across ten seconds and a replay with ten observations across two hours both have ten points but radically different evidence density. Likewise, a small median gap can coexist with one very large unobserved interval.

The product therefore needed a visible description of the sampling shape. The browser already had every required persisted observation, so a new provider, backend endpoint, database structure or paid service was unnecessary.

## Expected result and actual result

Expected result:

- sparse evidence remains visibly sparse;
- dense evidence is described by measurable timing properties rather than an opaque confidence score;
- large missing intervals remain explicit;
- altitude evidence remains independently qualified;
- the feature reuses existing replay semantics instead of implementing a second gap/coverage model;
- no new network path or backend contract is introduced;
- no paid infrastructure or aviation data is required.

Actual result after merge matches that expectation. The merged frontend computes the profile deterministically from already loaded persisted observations, preserves the existing observed-only boundary, passes the exact-head and post-merge validation matrices and deploys successfully to Vercel.

## Exact merge evidence

User-approved exact PR head:

```text
PR=159
APPROVED_HEAD=691d69b0241d1370ac901e1790d8a7736f8c33d8
MERGE_METHOD=SQUASH
EXPECTED_HEAD_GUARD=ENFORCED
```

The merge operation used `expected_head_sha`, so GitHub would reject the merge if the PR head moved after approval.

Resulting authoritative main revision:

```text
MAIN_SHA=2e2a542e7c970441074ddc6c67aa7548ce9b91b3
```

GitHub reports PR #159 as merged and `main` points to this exact GitHub-verified squash commit. Its parent is the previous authoritative main revision `cfd005f3e6259736f589c9b209820cc53c38c3cd`.

## Final pre-merge exact-head validation

The final merge-candidate head was validated independently after Document 202 recorded the successful remediation matrix:

```text
HEAD=691d69b0241d1370ac901e1790d8a7736f8c33d8

Frontend CI #433
run=34046106366
result=SUCCESS

Backend CI #771
run=34046106253
result=SUCCESS

CodeQL #413
run=34046106358
result=SUCCESS

API Load Baseline #301
run=34046106343
result=SUCCESS

Playwright E2E #210
run=34046106337
result=SUCCESS

Vercel preview
result=SUCCESS
```

This final cycle validated the exact code + regression contracts + completed pre-merge engineering history that was later authorized for squash merge.

## Post-merge CI evidence

All expected push workflows on exact main SHA `2e2a542e7c970441074ddc6c67aa7548ce9b91b3` completed successfully:

```text
Frontend CI #434
run=34047359477
result=SUCCESS

Backend CI #772
run=34047359503
result=SUCCESS

CodeQL #414
run=34047359492
result=SUCCESS

Playwright E2E #211
run=34047359497
result=SUCCESS

Vercel
result=SUCCESS
```

Frontend #434 completed ESLint, TypeScript validation, the full frontend contract suite, production build and Frontend CI gate successfully.

Backend #772 completed Backend Quality, PostgreSQL 16 Integration, Backend Race Safety, container build, non-root runtime verification, historical materializer verification, container health smoke and the Backend CI gate successfully.

CodeQL #414 completed both Go and JavaScript/TypeScript analysis and the security gate successfully.

Playwright #211 completed the Chromium end-to-end suite and evidence upload successfully.

API Load Baseline did not run on this post-merge push because the existing workflow/event/path policy does not require it for this push shape. This is expected behavior, not a missing or failed check. The final merge-candidate head had already passed API Load Baseline #301 / run `34046106343`.

## Real review and remediation history

Stage 18 contains a real pre-merge CI rejection and it remains preserved in Document 202 rather than being rewritten here.

The truthful sequence is:

1. Initial implementation was built from exact base main `cfd005f3e6259736f589c9b209820cc53c38c3cd`.
2. Initial PR validation head `cb93a00d46f874e0cd0cf48fe00f855948e65e7a` entered full CI.
3. Frontend CI #428 / run `34045471852` failed after ESLint and TypeScript passed; 172 of 173 frontend tests passed and one source-contract test failed.
4. The product UI already contained the required semantic disclosure that the descriptive metrics were not calibrated aviation quality grades.
5. Root cause: the source-contract regex depended on one physical JSX line layout and did not tolerate the newline between `aviation` and `quality`.
6. Bad fixes were rejected: rewriting production JSX purely for the regex, deleting the assertion, weakening it to a generic word, or bypassing the test.
7. The selected remediation made the semantic phrase whitespace-tolerant while preserving the full required wording: `/not calibrated aviation\s+quality grades/`.
8. Product metrics, UI semantics and evidence boundaries were unchanged.
9. Remediation validation head `9ca39c6f2b170ae260e5f67d21265c5fcf285c42` passed Frontend #431, Backend #769, CodeQL #411, API Load #299, Playwright #208 and Vercel.
10. Document 202 recorded this real failure → root cause → rejected fixes → remediation → successful validation history.
11. Final exact merge-candidate head `691d69b0241d1370ac901e1790d8a7736f8c33d8` passed the second complete validation cycle.
12. PR #159 was squash-merged only after explicit exact-head authorization.
13. Resulting main SHA `2e2a542e7c970441074ddc6c67aa7548ce9b91b3` passed all expected post-merge workflows and Vercel deployment.
14. This document records the final post-merge closure.

No additional product or architecture defect was discovered after merge, so no post-merge remediation story is invented.

## Evidence boundary after closure

The merged feature preserves these guarantees:

```text
EVIDENCE_PROFILE=DESCRIPTIVE_ONLY
SYNTHETIC_QUALITY_SCORE=NONE
GOOD_FAIR_POOR_THRESHOLDS=NONE
POSITION_INTERPOLATION=NONE
UNOBSERVED_INTERVALS_RECONSTRUCTED=NO
NEW_BACKEND_DATA=NONE
ATC_NAVIGATION_SAFETY_GRADE_CLAIM=NONE
```

Mean, median and P90 gaps describe only the actual adjacent persisted-observation time intervals.

Largest-gap share describes how dominant the largest measured missing interval is relative to the full observed span. It is not a severity classification.

Observation density describes persisted sample count per observed-span hour. It does not prove uniform polling or continuous aircraft coverage.

Altitude evidence coverage describes the proportion of persisted observations with supported altitude evidence. It does not upgrade unsupported altitude into known altitude.

The largest unobserved interval is bounded only by the two persisted timestamps on either side of the gap. No position, route, phase or behavior inside that gap is inferred.

## Why the selected architecture remains correct

The final architecture reuses `buildFlightReplayGaps(...)` and `buildFlightReplayAnalyticsSummary(...)` and adds a small deterministic frontend evidence-profile model.

This remains preferable because it:

- avoids duplicating gap semantics;
- avoids duplicating altitude-coverage semantics;
- requires no second network path;
- requires no backend endpoint;
- requires no provider change;
- requires no database migration;
- keeps missing observations visible as missing;
- avoids an uncalibrated scalar quality score;
- remains deterministic and straightforward to unit-test;
- stays inside the 0 RUB development budget.

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

Stage 18 reuses existing persisted replay states, existing frontend replay data, existing Next.js/React execution and existing analytics/gap helpers. It adds no production request volume to an external aviation source.

## Regression protection

The merged repository permanently protects Stage 18 through:

- unit tests executing the real production evidence-quality model from `.test-dist`;
- explicit compilation of that production model in `apps/web/tsconfig.test.json`;
- source contracts requiring reuse of the existing replay gap/analytics semantics;
- source contracts forbidding a second `fetch`/`axios` data path in the Stage 18 model and integration;
- source contracts forbidding synthetic `qualityScore` / `confidenceScore` logic;
- source contracts preserving descriptive-only UI markers;
- whitespace-tolerant semantic protection for the calibration disclosure;
- browser coverage of the honest single-sample state inside the existing aircraft-intelligence journey;
- Document 202 documentation-contract coverage of product purpose, real remediation history and zero-budget/evidence boundaries.

This closure adds a separate documentation contract so the exact merge revision, final and post-merge CI evidence, zero-budget boundary and residual limitations cannot silently disappear from future documentation changes.

## Residual limitations

Stage 18 is closed, but the following limits remain intentional and explicit:

- the profile describes only observations actually persisted in the selected replay;
- FREE_V1 ingestion cadence can leave historical replay evidence sparse;
- a high observation density does not prove uniform temporal coverage;
- mean and median gaps can still hide local irregularities, which is why largest gap, largest-gap share and P90 are exposed separately;
- P90 is descriptive and is not an SLO or aviation-quality threshold;
- largest-gap share is not a calibrated severity score;
- altitude coverage does not say anything about the accuracy of each supported altitude beyond its existing evidence semantics;
- one persisted observation cannot establish an inter-sample distribution or positive-span density;
- no intermediate position, route, flight phase, aircraft intent or pilot behavior is inferred across gaps;
- the feature is not ATC-grade, navigation-grade or safety-critical evidence;
- Stage 18 does not add denser commercial history or paid aviation data.

These are evidence limitations, not reasons to fabricate richer frontend values.

## Operational consequences

There is no new service, migration, provider dependency, cache, storage layer or paid runtime to operate.

The browser performs deterministic descriptive calculations over replay data it already loaded. The operational burden is therefore limited to the existing frontend bundle/tests and the permanent CI contracts that protect the evidence semantics.

## Future guard

Any future replay evidence-quality feature must declare before implementation:

1. exact source fields already available to the frontend;
2. observed / derived / projected evidence class;
3. whether any threshold or score is empirically calibrated;
4. what downstream purpose that calibration represents;
5. missing-data behavior;
6. how gaps affect interpretation;
7. whether any unobserved position or path would be inferred;
8. regression protection;
9. infrastructure impact;
10. monetary cost.

A synthetic score or qualitative grade must not be introduced merely because it is visually convenient. It requires a documented purpose and defensible calibration evidence.

If a desired frontend capability requires data the current provider/backend does not truthfully expose, it must not be fabricated. If the only defensible implementation requires a paid source or paid infrastructure, implementation stops at the zero-budget boundary until the user explicitly changes that constraint.

## Final status

```text
STAGE_18_IMPLEMENTATION=MERGED
STAGE_18_EXACT_HEAD_GUARD=PASS
STAGE_18_FINAL_PRE_MERGE_CI=PASS
STAGE_18_POST_MERGE_CI=PASS
STAGE_18_VERCEL=PASS
STAGE_18_EVIDENCE_PROFILE=DESCRIPTIVE_ONLY
STAGE_18_SYNTHETIC_QUALITY_SCORE=NONE
STAGE_18_POSITION_INTERPOLATION=NONE
STAGE_18_NEW_BACKEND_DATA=NONE
STAGE_18_ADDITIONAL_COST=0_RUB
STAGE_18_DOCUMENTATION=CLOSED
STAGE_18_REPLAY_EVIDENCE_QUALITY=CLOSED
```
