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
    'STAGE_17_OBSERVED_INTERVAL_COMPARISON=IMPLEMENTED_PENDING_CI',
    'STAGE_17_BASE_SHA=0fb818d23a60cd231f7ecbd290af2ccde03a488e',
    'STAGE_17_EVIDENCE_CLASS=OBSERVED_ENDPOINTS',
    'STAGE_17_POSITION_INTERPOLATION=NONE',
    'STAGE_17_PATH_DISTANCE_CLAIM=NONE',
    'STAGE_17_ADDITIONAL_COST=0_RUB',
    'STAGE_17_INITIAL_PR_CI=PENDING',
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

test('Stage 17 documentation explicitly rejects unsupported frontend claims', () => {
  assert.match(document, /not the actual travelled path/i)
  assert.match(document, /interpolat(e|ion)/i)
  assert.match(document, /new backend comparison endpoint/i)
  assert.match(document, /no intermediate coordinate, route, phase or aircraft intent is inferred/i)
  assert.match(document, /must not be added merely to make the UI look richer/i)
  assert.match(document, /No historical review rejection is claimed at this point/)
})
