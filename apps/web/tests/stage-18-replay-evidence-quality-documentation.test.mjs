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

test('Stage 18 documentation preserves the real first CI rejection and remediation', () => {
  for (const evidence of [
    'STAGE_18_FIRST_VALIDATION_HEAD=cb93a00d46f874e0cd0cf48fe00f855948e65e7a',
    'STAGE_18_FIRST_FRONTEND_CI=FAIL',
    'STAGE_18_REMEDIATION_HEAD=f814848601d5d78c736b5c7cd73e1e5735bf5dea',
    'Frontend CI #428',
    'run=34045471852',
    'Tests=173',
    'Pass=172',
    'Fail=1',
  ]) {
    assert.match(document, new RegExp(evidence))
  }

  assert.match(document, /formatting-sensitive literal sequence/i)
  assert.match(document, /test-contract defect, not an evidence-model or product-UI defect/i)
  assert.match(document, /rewriting the production UI sentence onto one source line/i)
  assert.match(document, /not calibrated aviation\\s\+quality grades/)
  assert.match(document, /No fictional product-review rejection is recorded/i)
})

test('Stage 18 documentation records the successful remediation validation matrix', () => {
  for (const evidence of [
    'STAGE_18_REMEDIATION_VALIDATION_HEAD=9ca39c6f2b170ae260e5f67d21265c5fcf285c42',
    'STAGE_18_REMEDIATION_VALIDATION=PASS',
    'STAGE_18_PRE_MERGE_CI=PASS',
    'Frontend CI #431',
    'run=34045730187',
    'Backend CI #769',
    'run=34045730140',
    'CodeQL #411',
    'run=34045730178',
    'API Load Baseline #299',
    'run=34045729973',
    'Playwright E2E #208',
    'run=34045730064',
    'Vercel preview',
    'result=SUCCESS',
  ]) {
    assert.match(document, new RegExp(evidence))
  }

  assert.match(document, /No second product or architecture defect was discovered/i)
})

test('Stage 18 documentation requires independent final exact-head verification', () => {
  assert.match(document, /STAGE_18_FINAL_EXACT_HEAD_CI=PENDING/)
  assert.match(document, /STAGE_18_POST_MERGE_CI=PENDING/)
  assert.match(document, /A commit cannot contain a truthful assertion of its own future CI run IDs/i)
  assert.match(document, /final complete exact-head CI\/Vercel cycle/i)
  assert.match(document, /PR metadata, which does not change the commit SHA/i)
})