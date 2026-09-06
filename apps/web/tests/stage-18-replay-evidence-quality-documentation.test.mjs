import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const document = readFileSync(
  new URL('../../../docs/202_STAGE_18_REPLAY_EVIDENCE_QUALITY_PROFILE.md', import.meta.url),
  'utf8'
)

test('Stage 18 documentation records product purpose, architecture and expected result', () => {
  for (const heading of [
    '## Product need',
    '## Why this feature exists',
    '## Problem',
    '## Root cause',
    '## Failure scenario',
    '## Impact on the project',
    '## Considered solutions',
    '## Chosen architecture',
    '## Expected product result',
    '## Expected engineering result',
    '## Regression protection',
    '## Zero-budget impact',
    '## Residual limitations',
    '## Future guard',
  ]) {
    assert.match(document, new RegExp(heading.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }
})

test('Stage 18 documentation preserves descriptive-only evidence semantics', () => {
  for (const marker of [
    'STAGE_18_EVIDENCE_PROFILE=DESCRIPTIVE_ONLY',
    'STAGE_18_SYNTHETIC_QUALITY_SCORE=NONE',
    'STAGE_18_POSITION_INTERPOLATION=NONE',
    'STAGE_18_NEW_BACKEND_DATA=NONE',
    'STAGE_18_ADDITIONAL_COST=0_RUB',
  ]) {
    assert.match(document, new RegExp(marker))
  }

  assert.match(document, /synthetic 0–100 replay quality score/i)
  assert.match(document, /fixed labels such as GOOD \/ FAIR \/ POOR/i)
  assert.match(document, /descriptive evidence profile derived client-side/i)
  assert.match(document, /No single currently justified scalar/i)
})

test('Stage 18 documentation records adversarial and residual evidence limits', () => {
  for (const limitation of [
    /same sample count, very different coverage/i,
    /median hides one dominant gap/i,
    /attractive score hides calibration assumptions/i,
    /client feature silently adds a new data dependency/i,
    /FREE_V1 ingestion cadence can produce sparse historical replay data/i,
    /observation density does not mean observations are uniformly distributed/i,
    /no route is reconstructed inside gaps/i,
    /no ATC-, navigation- or safety-grade quality claim is made/i,
  ]) {
    assert.match(document, limitation)
  }
})

test('Stage 18 documentation keeps truthful CI state before validation', () => {
  assert.match(document, /STAGE_18_PRE_MERGE_CI=PENDING/)
  assert.match(document, /STAGE_18_POST_MERGE_CI=PENDING/)
  assert.match(document, /No CI\/review outcome is claimed yet/)
  assert.match(document, /fictional review history must not be invented/i)
})