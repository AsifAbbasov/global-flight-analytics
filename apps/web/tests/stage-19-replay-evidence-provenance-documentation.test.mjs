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
    '## Real first Continuous Integration rejection and remediation',
    '## Successful remediation validation',
    '## Why final exact-head validation is still pending',
    '## Expected result',
    '## Infrastructure and monetary impact',
    '## Residual limitations',
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

test('Stage 19 documentation preserves the real first CI rejection and remediation', () => {
  for (const evidence of [
    'STAGE_19_FIRST_VALIDATION_HEAD=88981e7036b3313d9aefad48983ac0bfd814902f',
    'STAGE_19_FIRST_FRONTEND_CI=FAIL',
    'STAGE_19_REMEDIATION_HEAD=282fa45ca6dba35cc7e1dc17bb811cafce917b7e',
    'Frontend CI #437',
    '34050031425',
    'Tests        190 total',
    'Pass         189',
    'Fail         1',
  ]) {
    assert.match(document, new RegExp(evidence))
  }

  assert.match(document, /formatting-sensitive source-contract defect/i)
  assert.match(document, /product component was not changed for the test/i)
  assert.match(document, /the\\s\+exact provider switch time between those observations is unknown/)
  assert.match(document, /Backend CI #775 \/ 34050031452\s+CANCELLED/)
  assert.match(document, /CodeQL #417 \/ 34050031494\s+CANCELLED/)
  assert.match(document, /API Load Baseline #303 \/ 34050031472 CANCELLED/)
  assert.match(document, /Playwright E2E #214 \/ 34050031497\s+CANCELLED/)
})

test('Stage 19 documentation records the successful remediation validation matrix', () => {
  for (const evidence of [
    'STAGE_19_REMEDIATION_VALIDATION=PASS',
    'Frontend CI #438',
    '34050082969',
    'Backend CI #776',
    '34050082973',
    'CodeQL #418',
    '34050082945',
    'API Load Baseline #304',
    '34050083049',
    'Playwright E2E #215',
    '34050083066',
    'Vercel',
    'SUCCESS',
  ]) {
    assert.match(document, new RegExp(evidence))
  }

  assert.match(document, /No second product, architecture or evidence defect was discovered/i)
})

test('Stage 19 documentation requires independent final exact-head verification', () => {
  assert.match(document, /STAGE_19_FINAL_EXACT_HEAD_CI=PENDING/)
  assert.match(document, /STAGE_19_POST_MERGE_CI=PENDING/)
  assert.match(document, /A commit cannot truthfully contain the future Continuous Integration run identifiers for itself/i)
  assert.match(document, /new independent exact-head Frontend, Backend, CodeQL, API Load, Playwright and Vercel cycle/i)
  assert.match(document, /Final run identifiers belong in pull-request metadata/i)
})
