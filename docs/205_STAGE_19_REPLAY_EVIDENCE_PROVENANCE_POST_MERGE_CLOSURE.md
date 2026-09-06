# Stage 19 Replay Evidence Provenance Profile Post-Merge Closure

Status: closed

```text
STAGE_19_REPLAY_EVIDENCE_PROVENANCE=CLOSED
STAGE_19_MERGED_PR=161
STAGE_19_APPROVED_HEAD=2405e76a5e8be9ab78d6165a7f68ce66eb5c4cba
STAGE_19_MAIN_SHA=64f6198bd96087d7bc479a78e1839db2f6914c68
STAGE_19_POST_MERGE_CI=PASS
STAGE_19_EVIDENCE_CLASS=PERSISTED_SOURCE_LABELS_ONLY
STAGE_19_PROVIDER_RANKING=NONE
STAGE_19_PROVIDER_ACCURACY_SCORE=NONE
STAGE_19_PROVIDER_SWITCH_TIME_INFERENCE=NONE
STAGE_19_NEW_BACKEND_DATA=NONE
STAGE_19_ADDITIONAL_COST=0_RUB
STAGE_19_DOCUMENTATION=CLOSED
```

## Purpose of this closure document

Document 204 records the Stage 19 product need, source-data boundary, chosen architecture, adversarial scenarios, rejected alternatives, the real first Continuous Integration rejection, remediation history, regression protection and successful pre-merge validation. It was written as the engineering-history document before the final merge outcome existed.

This document appends the post-merge truth instead of rewriting Document 204 after the fact. The auditable sequence is:

```text
Document 204
product design + implementation + real CI rejection + remediation + pre-merge evidence
        ↓
final exact-head validation of PR #161
        ↓
exact-head squash merge of PR #161
        ↓
post-merge validation on resulting main SHA
        ↓
Document 205
post-merge closure evidence
```

Document 204 therefore remains a historical pre-merge snapshot. Its earlier pending/in-progress statements must not be read as the current Stage 19 status after this closure document lands.

## What was delivered

Stage 19 adds a visible Replay Evidence Provenance Profile to Historical Flight Replay using source metadata already persisted on replay observations.

The delivered profile exposes:

- total persisted replay sample count;
- number of identified persisted source labels;
- number of unattributed persisted samples;
- persisted sample count per identified source;
- persisted-sample share per identified source;
- first and last persisted observation timestamps carrying each identified source label;
- adjacent persisted observation pairs whose identified source labels differ;
- exact persisted endpoint timestamps for each observed source-label transition;
- elapsed time between those persisted endpoints when timestamps are valid;
- jump-to-later-observation navigation using existing replay cursor semantics.

The feature deliberately does not expose provider ranking, provider accuracy score, a preferred-provider claim, elapsed-time ownership derived from sample share, exact provider switch-time inference inside gaps, or direct source transitions fabricated across unattributed observations.

## Why it was implemented

Stages 15 through 18 progressively made Historical Replay evidence-aware: observed-only playback, gap visibility, elapsed-time navigation, observed interval comparison and descriptive evidence-density metrics. After Stage 18, however, source provenance remained point-local: the user could see `source_name` for the currently selected observation but could not understand replay-wide source composition without manually stepping through the entire replay.

The existing replay payload already contains `FlightReplayPoint.source_name` and `FlightReplayPoint.observed_at` for every persisted point. Production ingestion already preserves the actual selected source on persisted observations. The missing capability was therefore presentation and deterministic aggregation over data already loaded in the browser, not a new data source or backend contract.

## Expected result and actual result

Expected result:

- single-source replay evidence is visibly distinguishable from multi-source replay evidence;
- unattributed persisted observations remain visible rather than silently disappearing;
- adjacent source-label changes are exposed only at their observed persisted endpoints;
- sample-share percentages remain sample shares rather than being misrepresented as elapsed-time coverage;
- unknown provenance breaks transition continuity instead of fabricating a direct provider-to-provider transition;
- provider provenance remains provenance, not a quality or accuracy grade;
- no second network path, new backend endpoint, database migration, provider dependency or paid infrastructure is introduced.

Actual result after merge matches that expectation. The merged frontend computes the provenance profile deterministically from already loaded replay points, preserves the existing observed-only replay boundary, passes the final exact-head and post-merge validation matrices and deploys successfully to Vercel.

## Exact merge evidence

User-approved exact PR head:

```text
PR=161
APPROVED_HEAD=2405e76a5e8be9ab78d6165a7f68ce66eb5c4cba
MERGE_METHOD=SQUASH
EXPECTED_HEAD_GUARD=ENFORCED
```

The merge operation used `expected_head_sha`, so GitHub would have rejected the merge if PR #161 had moved from the approved head.

Resulting authoritative main revision:

```text
MAIN_SHA=64f6198bd96087d7bc479a78e1839db2f6914c68
```

GitHub reports PR #161 as merged and `main` points to this exact GitHub-verified squash commit. Its direct parent is the previous authoritative main revision `cbe6905a96f0656d868e328ba7c6ab38e756cfd8`.

## Final pre-merge exact-head validation

The final merge-candidate head was independently validated after Document 204 had already recorded the real failure/remediation sequence:

```text
HEAD=2405e76a5e8be9ab78d6165a7f68ce66eb5c4cba

Frontend CI #440
run=34050341592
result=SUCCESS

Backend CI #778
run=34050341638
result=SUCCESS

CodeQL #420
run=34050341680
result=SUCCESS

API Load Baseline #306
run=34050341750
result=SUCCESS

Playwright E2E #217
run=34050341642
result=SUCCESS

Vercel preview
result=SUCCESS
```

This final cycle validated the exact code, regression contracts and completed pre-merge engineering history that was later authorized for squash merge.

## Post-merge CI evidence

All expected push workflows on exact main SHA `64f6198bd96087d7bc479a78e1839db2f6914c68` completed successfully:

```text
Frontend CI #441
run=34054495884
result=SUCCESS

Backend CI #779
run=34054496033
result=SUCCESS

CodeQL #421
run=34054495862
result=SUCCESS

Playwright E2E #218
run=34054495873
result=SUCCESS

Vercel
result=SUCCESS
```

Frontend #441 validates the merged frontend quality gates on the resulting main revision.

Backend #779 validates the existing backend quality, PostgreSQL 16 integration, race-safety and container/runtime verification gates on the same resulting main revision.

CodeQL #421 validates the security-analysis workflow on the same resulting main revision.

Playwright #218 validates the Chromium end-to-end browser journey, including the Stage 19 provenance profile integration, on the resulting main revision.

Vercel reports `Deployment has completed` with `state=success` for the same main SHA.

API Load Baseline did not run on this post-merge push because the existing workflow/event/path policy does not require it for this push shape. This is expected behavior, not a missing or failed check. The final merge-candidate head had already passed API Load Baseline #306 / run `34050341750`.

## Real review and remediation history

Stage 19 contains a real pre-merge CI rejection and it remains preserved rather than rewritten away.

The truthful sequence is:

1. Stage 19 implementation was based on exact Stage 18-closed main `cbe6905a96f0656d868e328ba7c6ab38e756cfd8`.
2. Initial feature head `88981e7036b3313d9aefad48983ac0bfd814902f` entered Pull Request CI.
3. Frontend CI #437 / run `34050031425` passed ESLint and TypeScript, then failed the frontend contract suite with 190 total tests, 189 passing and one failing.
4. The failing assertion required the literal source sequence `the exact provider switch time between those observations is unknown`.
5. The production UI already contained the full semantic disclosure, but harmless JSX formatting inserted whitespace/newline between `the` and `exact`.
6. Root cause: the source-contract regular expression was formatting-sensitive rather than semantic.
7. Bad fixes were rejected: rewriting product copy merely to satisfy the regex, deleting or weakening the assertion, or bypassing the contract.
8. The selected remediation preserved the full required semantic phrase while tolerating source formatting: `/the\s+exact provider switch time between those observations is unknown/`.
9. Product behavior, provenance semantics, evidence boundaries and architecture were unchanged.
10. Remediation head `282fa45ca6dba35cc7e1dc17bb811cafce917b7e` passed Frontend CI #438, Backend CI #776, CodeQL #418, API Load Baseline #304, Playwright E2E #215 and Vercel.
11. Document 204 recorded the real failure → root cause → rejected fixes → remediation → successful validation history.
12. Final exact merge-candidate head `2405e76a5e8be9ab78d6165a7f68ce66eb5c4cba` then passed a second complete validation cycle.
13. PR #161 was squash-merged only after explicit exact-head authorization.
14. Resulting main SHA `64f6198bd96087d7bc479a78e1839db2f6914c68` passed all expected post-merge push workflows and Vercel deployment.
15. This document records the final post-merge closure.

The other workflows from the first failed head were cancelled after the remediation push and are not misclassified as failures or successes. No additional product or architecture defect was discovered after merge, so no fictional post-merge remediation story is created.

## Evidence boundary after closure

The merged feature preserves these guarantees:

```text
EVIDENCE_CLASS=PERSISTED_SOURCE_LABELS_ONLY
SOURCE_LABEL_INPUT=FlightReplayPoint.source_name
TIMESTAMP_INPUT=FlightReplayPoint.observed_at
SAMPLE_SHARE_IS_ELAPSED_TIME_SHARE=NO
PROVIDER_RANKING=NONE
PROVIDER_ACCURACY_SCORE=NONE
PREFERRED_PROVIDER_CLAIM=NONE
PROVIDER_SWITCH_TIME_INFERENCE=NONE
UNKNOWN_SOURCE_BRIDGING=NONE
POSITION_INTERPOLATION=NONE
NEW_BACKEND_DATA=NONE
```

A source label identifies the persisted attribution attached to a replay observation. It does not certify the provider as more accurate, more complete or more reliable than another provider.

A source sample share is calculated over persisted replay observations. It does not prove the same share of elapsed replay time, request traffic, provider uptime or real-world aviation coverage.

A source transition exists only when two adjacent persisted observations both carry identified source labels and those labels differ. The feature knows the two observed endpoint timestamps; it does not know the exact upstream provider switch instant inside the interval.

An unattributed observation breaks source-transition continuity. For example, `source A → unavailable → source B` is not reported as a direct `source A → source B` transition.

No position, route, phase, intent, provider quality or provider ownership is reconstructed inside an unobserved interval.

## Why the selected architecture remains correct

The final architecture remains:

```text
existing Historical Replay HTTP response
        ↓
FlightReplayPoint[]
        ↓
source_name + observed_at
        ↓
flight-replay-evidence-provenance-model.ts
        ↓
FlightReplayEvidenceProvenance
        ↓
existing FlightReplayTimeNavigation
```

This remains preferable because it:

- consumes provenance already transported by the existing replay contract;
- avoids a duplicate backend aggregate endpoint;
- avoids a second frontend network path;
- keeps provenance logic deterministic and bounded by the loaded replay points;
- preserves missing attribution rather than inventing provider identity;
- avoids unsupported provider ranking and quality claims;
- reuses existing replay cursor semantics for navigation;
- requires no provider change, ingestion change, database migration or new runtime service;
- remains straightforward to unit-test and browser-test;
- stays inside the zero-budget boundary.

## Zero-budget and infrastructure impact

```text
New paid aviation provider = NO
New backend endpoint        = NO
New PostgreSQL table        = NO
New database migration      = NO
New ingestion path          = NO
New server                  = NO
New cache                   = NO
New persistent storage      = NO
New runtime dependency      = NO
Additional external calls   = NO
Additional cost             = 0 RUB
```

Stage 19 reuses existing persisted replay states, existing frontend replay data and existing application infrastructure. It adds no production request volume to external aviation providers.

## Regression protection

The merged repository permanently protects Stage 19 through:

- unit tests executing the real compiled production provenance model from `.test-dist`;
- explicit compilation of the provenance model in `apps/web/tsconfig.test.json`;
- model tests covering multi-source composition, source transitions, unattributed evidence and invalid transition timestamps;
- source contracts requiring existing `source_name` and `observed_at` evidence;
- source contracts forbidding a second `fetch`/`axios` data path;
- source contracts forbidding provider ranking, accuracy-score and preferred-provider fields;
- whitespace-tolerant semantic protection for the exact switch-time limitation disclosure;
- browser coverage of the provenance profile inside the existing aircraft-intelligence Playwright journey;
- Document 204 documentation-contract coverage of purpose, evidence boundaries, real remediation history and zero-budget scope.

This closure adds a separate permanent documentation contract so the exact approved head, resulting main SHA, final/post-merge validation evidence, provenance-only semantics, zero-budget boundary and residual limitations cannot silently disappear from future documentation changes.

## Residual limitations

Stage 19 is closed, but these limits remain intentional and explicit:

- provenance describes only source labels actually persisted on the selected replay observations;
- absent or blank `source_name` remains unattributed;
- a source label does not prove provider accuracy, completeness, reliability or superiority;
- persisted sample share does not prove elapsed-time coverage or provider uptime;
- adjacent different source labels only bound a change between two persisted endpoints;
- the exact provider switch instant inside that interval remains unknown;
- unknown provenance prevents a fabricated direct transition through the unknown point;
- request-level fallback reasons are not reconstructed from replay points;
- provider selection decision internals are not inferred from source labels alone;
- no continuous provider ownership is projected across unobserved intervals;
- no denser commercial history, satellite coverage or paid provider evidence is added;
- the feature is descriptive research tooling, not ATC-, navigation- or safety-grade provenance certification.

These are evidence limitations, not reasons to fabricate richer frontend claims.

## Operational consequences

There is no new service, migration, provider dependency, cache, storage layer or paid runtime to operate.

The browser performs deterministic descriptive calculations over replay data it already loaded. Operational burden remains limited to the existing frontend bundle/tests and permanent Continuous Integration contracts.

## Future guard

Any future replay provenance feature must declare before implementation:

1. exact persisted provenance fields it consumes;
2. whether the output describes samples, requests, ingestion runs or elapsed time;
3. missing-provenance behavior;
4. whether a source transition is directly observed or only bounded by adjacent observations;
5. whether any provider ranking, quality or accuracy claim is calibrated;
6. what evidence supports that calibration;
7. whether a new backend/provider data path is actually necessary;
8. regression protection;
9. infrastructure impact;
10. monetary cost.

Provider provenance must never be silently converted into provider quality, accuracy, completeness or preference.

If a desired capability requires evidence the existing provider/backend does not truthfully expose, it must not be fabricated. If the only defensible implementation requires a paid source or paid infrastructure, implementation stops at the zero-budget boundary until that constraint is explicitly changed.

## Final status

```text
STAGE_19_IMPLEMENTATION=MERGED
STAGE_19_EXACT_HEAD_GUARD=PASS
STAGE_19_FINAL_PRE_MERGE_CI=PASS
STAGE_19_POST_MERGE_CI=PASS
STAGE_19_VERCEL=PASS
STAGE_19_EVIDENCE_CLASS=PERSISTED_SOURCE_LABELS_ONLY
STAGE_19_PROVIDER_RANKING=NONE
STAGE_19_PROVIDER_ACCURACY_SCORE=NONE
STAGE_19_PROVIDER_SWITCH_TIME_INFERENCE=NONE
STAGE_19_NEW_BACKEND_DATA=NONE
STAGE_19_ADDITIONAL_COST=0_RUB
STAGE_19_DOCUMENTATION=CLOSED
STAGE_19_REPLAY_EVIDENCE_PROVENANCE=CLOSED
```
