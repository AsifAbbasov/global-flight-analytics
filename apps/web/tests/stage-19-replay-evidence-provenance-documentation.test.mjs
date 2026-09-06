import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const document = readFileSync(
  new URL('../../../docs/204_STAGE_19_REPLAY_EVIDENCE_PROVENANCE_PROFILE.md', import.meta.url),
  'utf8'
)

test('Stage 19 documentation records product purpose, evidence boundary and architecture', () => {
  for (const heading of [
    '## Product need',
    '## Source data and evidence boundary',
    '## Product behavior',
    '## Problem and root cause',
    '## Failure and adversarial scenarios',
    '## Considered solutions',
    '## Chosen architecture',
    '## Missing-data policy',
    '## Percentage semantics',
    '## Transition semantics',
    '## Regression protection',
    '## Expected result',
    '## Infrastructure and monetary impact',
    '## Residual limitations',
    '## Continuous Integration evidence',
    '## Future guard',
  ]) {
    assert.match(document, new RegExp(heading.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }
})

test('Stage 19 documentation preserves provenance-only semantics', () => {
  for (const marker of [
    'STAGE_19_EVIDENCE_CLASS=PERSISTED_SOURCE_LABELS_ONLY',
    'STAGE_19_PROVIDER_RANKING=NONE',
    'STAGE_19_PROVIDER_ACCURACY_SCORE=NONE',
    'STAGE_19_PROVIDER_SWITCH_TIME_INFERENCE=NONE',
    'STAGE_19_NEW_BACKEND_DATA=NONE',
    'STAGE_19_ADDITIONAL_COST=0_RUB',
  ]) {
    assert.match(document, new RegExp(marker))
  }

  assert.match(document, /source percentages are shares of persisted samples, not shares of elapsed time/i)
  assert.match(document, /cannot know when the upstream provider changed between them/i)
  assert.match(document, /Provider provenance must never be silently converted into provider quality/i)
})

test('Stage 19 documentation protects missing provenance and zero-budget behavior', () => {
  for (const limitation of [
    /unattributed evidence breaks transition continuity/i,
    /do not create a direct transition through unknown provenance/i,
    /provider-level accuracy measurement/i,
    /exact provider switch time between persisted observations/i,
    /continuous source ownership across an unobserved interval/i,
    /Additional cost\s+0 RUB/i,
  ]) {
    assert.match(document, limitation)
  }
})

test('Stage 19 documentation keeps future CI facts pending before validation exists', () => {
  for (const marker of [
    'STAGE_19_PRE_MERGE_CI=PENDING',
    'STAGE_19_POST_MERGE_CI=PENDING',
    'Frontend CI       PENDING',
    'Backend CI        PENDING',
    'CodeQL            PENDING',
    'API Load Baseline PENDING',
    'Playwright E2E    PENDING',
    'Vercel            PENDING',
  ]) {
    assert.match(document, new RegExp(marker))
  }

  assert.match(document, /must not be fabricated/i)
})
