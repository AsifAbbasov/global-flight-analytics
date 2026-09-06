import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const document = readFileSync(
  new URL(
    '../../../docs/199_STAGE_16_TIME_BASED_REPLAY_POST_MERGE_CLOSURE.md',
    import.meta.url
  ),
  'utf8'
)

test('Stage 16 post-merge closure preserves exact merge and evidence semantics', () => {
  for (const marker of [
    'STAGE_16_TIME_BASED_REPLAY_NAVIGATION=CLOSED',
    'STAGE_16_MERGED_PR=155',
    'STAGE_16_APPROVED_HEAD=ff5b3f0af6e146e4dd48502798a3a83c334f42be',
    'STAGE_16_MAIN_SHA=b07e2a506455b695944c57e82327e7b06f0c0edc',
    'STAGE_16_POST_MERGE_CI=PASS',
    'STAGE_16_EVIDENCE_CLASS=OBSERVED',
    'STAGE_16_POSITION_POLICY=LAST_PERSISTED_OBSERVATION',
    'STAGE_16_POSITION_INTERPOLATION=NONE',
    'STAGE_16_ADDITIONAL_COST=0_RUB',
    'STAGE_16_DOCUMENTATION=CLOSED',
  ]) {
    assert.match(document, new RegExp(marker))
  }
})

test('Stage 16 closure records the exact post-merge validation matrix', () => {
  for (const evidence of [
    'Frontend CI #417',
    'run=34037166018',
    'Backend CI #755',
    'run=34037166006',
    'CodeQL #397',
    'run=34037166092',
    'Playwright E2E #194',
    'run=34037166020',
    'Vercel',
    'result=SUCCESS',
    'API Load Baseline #288',
    'run `34036556415`',
  ]) {
    assert.match(document, new RegExp(evidence.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }

  assert.match(document, /existing workflow\/event\/path policy does not trigger it/i)
  assert.match(document, /expected behavior, not a missing or failed check/i)
})

test('Stage 16 closure preserves chronological documentation instead of rewriting history', () => {
  assert.match(document, /Document 198 records the Stage 16 design/i)
  assert.match(document, /appends the post-merge truth instead of rewriting Document 198/i)
  assert.match(document, /chronological and auditable/i)
  assert.match(document, /No architecture or regression failure was discovered/i)
  assert.match(document, /no legitimate rejected-solution remediation story to invent/i)
})

test('Stage 16 closure keeps residual evidence and zero-budget limits explicit', () => {
  for (const limitation of [
    /only navigate observations that were actually persisted/i,
    /FREE_V1 ingestion cadence can leave historical replay sparse/i,
    /missing observations cannot be reconstructed truthfully/i,
    /not ATC-grade, navigation-grade or safety-critical evidence/i,
    /does not establish travelled path/i,
    /does not add commercial historical coverage/i,
  ]) {
    assert.match(document, limitation)
  }

  assert.match(document, /New paid aviation provider = NO/)
  assert.match(document, /Additional cost\s+= 0 RUB/)
  assert.match(document, /must not be fabricated/i)
  assert.match(document, /only valid implementation requires paid data or infrastructure/i)
})
