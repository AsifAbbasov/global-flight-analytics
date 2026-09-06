import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const document = readFileSync(
  new URL(
    '../../../docs/200_STAGE_17_OBSERVED_INTERVAL_COMPARISON.md',
    import.meta.url
  ),
  'utf8'
)

test('Stage 17 documentation preserves purpose, evidence and zero-budget markers', () => {
  for (const marker of [
    'STAGE_17_OBSERVED_INTERVAL_COMPARISON=MERGE_CANDIDATE',
    'STAGE_17_BASE_SHA=0fb818d23a60cd231f7ecbd290af2ccde03a488e',
    'STAGE_17_PR=157',
    'STAGE_17_INITIAL_HEAD=27676696fccddf364ca5f804bc1d8fc2b5532a80',
    'STAGE_17_REMEDIATION_VALIDATION_SHA=21bd9474ac5bc1208251d078715122b2a29214ca',
    'STAGE_17_EVIDENCE_CLASS=OBSERVED_ENDPOINTS',
    'STAGE_17_POSITION_INTERPOLATION=NONE',
    'STAGE_17_PATH_DISTANCE_CLAIM=NONE',
    'STAGE_17_ADDITIONAL_COST=0_RUB',
    'STAGE_17_INITIAL_PR_CI=FAILED_REMEDIATED',
    'STAGE_17_REMEDIATION_CI=PASS',
    'STAGE_17_DOCUMENTATION=PREMERGE_COMPLETE',
    'STAGE_17_POST_MERGE_CI=PENDING',
  ]) {
    assert.match(document, new RegExp(marker))
  }
})

test('Stage 17 documentation contains the complete engineering decision narrative', () => {
  for (const heading of [
    '## Product need',
    '## Why this feature exists',
    '## Expected product result',
    '## Problem',
    '## Root cause',
    '## Failure scenario',
    '## Project impact',
    '## Existing guarantees that must remain intact',
    '## Considered solutions',
    '## Chosen architecture',
    '## Why this solution was selected',
    '## Implementation design',
    '## Adversarial scenario 1',
    '## Adversarial scenario 2',
    '## Adversarial scenario 3',
    '## Adversarial scenario 4',
    '## Initial implementation review outcome',
    '### Remediation record A',
    '### Remediation record B',
    '## Remediation validation',
    '## Trade-offs',
    '## Zero-budget impact',
    '## Regression protection',
    '## CI / review evidence',
    '## Residual limitations',
    '## Expected completion criteria',
    '## Future guard',
  ]) {
    assert.match(document, new RegExp(heading.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }
})

test('Stage 17 documentation preserves the real failed-then-remediated CI history', () => {
  for (const evidence of [
    'Frontend CI #420',
    '34038594838',
    '151 passing tests and 2 failures',
    'ERR_MODULE_NOT_FOUND',
    'tsconfig.test.json',
    'whitespace-tolerant',
    'Frontend CI #422',
    '34038754833',
    'Backend CI #760',
    '34038754813',
    'CodeQL #402',
    '34038754816',
    'API Load Baseline #292',
    '34038754841',
    'Playwright E2E #199',
    '34038754814',
    'Vercel preview',
  ]) {
    assert.match(document, new RegExp(evidence.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }

  assert.match(document, /remove the unit test or replace it only with source-text assertions/i)
  assert.match(document, /rewrite production JSX onto one line only to satisfy the test/i)
  assert.match(document, /second full exact-head CI cycle/i)
})

test('Stage 17 documentation explicitly rejects unsupported frontend claims', () => {
  assert.match(document, /not the actual travelled path/i)
  assert.match(document, /interpolat(e|ion)/i)
  assert.match(document, /new backend comparison endpoint/i)
  assert.match(document, /no intermediate coordinate, route, phase or aircraft intent is inferred/i)
  assert.match(document, /must not be added merely to make the UI look richer/i)
})
