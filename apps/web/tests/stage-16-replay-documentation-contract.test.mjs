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
    'STAGE_16_TIME_BASED_REPLAY_NAVIGATION=PR_CANDIDATE',
    'STAGE_16_BASE_SHA=ceed3ecd31bd72d0de9491d23eda6bcb10cb2a98',
    'STAGE_16_EVIDENCE_CLASS=OBSERVED',
    'STAGE_16_TIME_CURSOR=REAL_ELAPSED_OBSERVATION_TIME',
    'STAGE_16_POSITION_POLICY=LAST_PERSISTED_OBSERVATION',
    'STAGE_16_POSITION_INTERPOLATION=NONE',
    'STAGE_16_ADDITIONAL_COST=0_RUB',
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
  assert.match(document, /time cursor only/)
  assert.match(document, /Position changes only when a persisted observation timestamp is crossed/)
  assert.match(document, /No review rejection will be invented/)
})

test('Stage 16 completion cannot be claimed before exact CI and documentation evidence exists', () => {
  assert.match(document, /Pending first exact-head PR validation/)
  assert.match(document, /PR number/)
  assert.match(document, /exact final head SHA/)
  assert.match(document, /POST_MERGE_CI=PASS/)
  assert.match(document, /DOCUMENTATION=COMPLETE/)
})
