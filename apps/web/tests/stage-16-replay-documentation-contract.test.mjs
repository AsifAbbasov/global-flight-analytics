import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const document = readFileSync(
  new URL(
    '../../../docs/198_STAGE_16_TIME_BASED_HISTORICAL_REPLAY_NAVIGATION.md',
    import.meta.url
  ),
  'utf8'
)

test('Stage 16 documentation preserves product purpose, evidence boundary and zero-budget contract', () => {
  for (const marker of [
    'STAGE_16_TIME_BASED_REPLAY_NAVIGATION=MERGE_CANDIDATE',
    'STAGE_16_BASE_SHA=ceed3ecd31bd72d0de9491d23eda6bcb10cb2a98',
    'STAGE_16_PR=155',
    'STAGE_16_IMPLEMENTATION_VALIDATION_SHA=9601b4809e6e467e9820c6467b91086d59a0a355',
    'STAGE_16_EVIDENCE_CLASS=OBSERVED',
    'STAGE_16_TIME_CURSOR=REAL_ELAPSED_OBSERVATION_TIME',
    'STAGE_16_POSITION_POLICY=LAST_PERSISTED_OBSERVATION',
    'STAGE_16_POSITION_INTERPOLATION=NONE',
    'STAGE_16_ADDITIONAL_COST=0_RUB',
    'STAGE_16_INITIAL_PR_CI=PASS',
    'STAGE_16_DOCUMENTATION=PREMERGE_COMPLETE',
    'STAGE_16_POST_MERGE_CI=PENDING',
  ]) {
    assert.match(document, new RegExp(marker))
  }

  for (const heading of [
    '## Product need',
    '## Why this feature exists',
    '## Expected product result',
    '## Problem',
    '## Root cause',
    '## Failure scenario',
    '## Project impact',
    '## Considered solutions',
    '## Chosen architecture',
    '## Why this solution was selected',
    '## Implementation design',
    '## Adversarial scenario 1',
    '## Adversarial scenario 2',
    '## Adversarial scenario 3',
    '## Initial implementation review outcome',
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

test('Stage 16 documentation explicitly rejects unsupported frontend data and synthetic movement', () => {
  assert.match(document, /request denser historical data from a new provider/i)
  assert.match(document, /Rejected under current constraints/)
  assert.match(document, /No coordinate is computed for 18:03:00/)
  assert.match(document, /Only time advances continuously/)
  assert.match(document, /Position changes only when another persisted observation timestamp is crossed/)
  assert.match(document, /No historical rejection is invented/)
})

test('Stage 16 documentation records exact first-cycle CI evidence and requires final exact-head verification', () => {
  for (const evidence of [
    'Frontend CI #414',
    'run=34036243832',
    'Backend CI #752',
    'run=34036243790',
    'CodeQL #394',
    'run=34036243838',
    'API Load Baseline #286',
    'run=34036243839',
    'Playwright E2E #191',
    'run=34036243846',
    'Vercel preview',
    'result=SUCCESS',
  ]) {
    assert.match(document, new RegExp(evidence.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }

  assert.match(document, /merge readiness requires a second full exact-head CI cycle/)
  assert.match(document, /FINAL_EXACT_HEAD_CI=PASS/)
  assert.match(document, /POST_MERGE_CI=PASS/)
})
