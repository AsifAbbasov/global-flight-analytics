import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const document = readFileSync(
  new URL(
    '../../../docs/203_STAGE_18_REPLAY_EVIDENCE_QUALITY_POST_MERGE_CLOSURE.md',
    import.meta.url
  ),
  'utf8'
)

test('Stage 18 post-merge closure preserves exact merge and evidence semantics', () => {
  for (const marker of [
    'STAGE_18_REPLAY_EVIDENCE_QUALITY=CLOSED',
    'STAGE_18_MERGED_PR=159',
    'STAGE_18_APPROVED_HEAD=691d69b0241d1370ac901e1790d8a7736f8c33d8',
    'STAGE_18_MAIN_SHA=2e2a542e7c970441074ddc6c67aa7548ce9b91b3',
    'STAGE_18_POST_MERGE_CI=PASS',
    'STAGE_18_EVIDENCE_PROFILE=DESCRIPTIVE_ONLY',
    'STAGE_18_SYNTHETIC_QUALITY_SCORE=NONE',
    'STAGE_18_POSITION_INTERPOLATION=NONE',
    'STAGE_18_NEW_BACKEND_DATA=NONE',
    'STAGE_18_ADDITIONAL_COST=0_RUB',
    'STAGE_18_DOCUMENTATION=CLOSED',
  ]) {
    assert.match(document, new RegExp(marker))
  }
})

test('Stage 18 closure records final and post-merge validation matrices', () => {
  for (const evidence of [
    'Frontend CI #433',
    'run=34046106366',
    'Backend CI #771',
    'run=34046106253',
    'CodeQL #413',
    'run=34046106358',
    'API Load Baseline #301',
    'run=34046106343',
    'Playwright E2E #210',
    'run=34046106337',
    'Frontend CI #434',
    'run=34047359477',
    'Backend CI #772',
    'run=34047359503',
    'CodeQL #414',
    'run=34047359492',
    'Playwright E2E #211',
    'run=34047359497',
    'Vercel',
    'result=SUCCESS',
  ]) {
    assert.match(document, new RegExp(evidence.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }

  assert.match(document, /existing workflow\/event\/path policy does not require it/i)
  assert.match(document, /expected behavior, not a missing or failed check/i)
})

test('Stage 18 closure preserves the real CI rejection and remediation history', () => {
  assert.match(document, /Document 202 records the Stage 18 product need/i)
  assert.match(document, /appends the post-merge truth rather than rewriting Document 202/i)
  assert.match(document, /Frontend CI #428 \/ run `34045471852` failed/i)
  assert.match(document, /172 of 173 frontend tests passed/i)
  assert.match(document, /source-contract regex depended on one physical JSX line layout/i)
  assert.match(document, /not calibrated aviation\\s\+quality grades/)
  assert.match(document, /No additional product or architecture defect was discovered after merge/i)
})

test('Stage 18 closure keeps descriptive-only and zero-budget limitations explicit', () => {
  for (const limitation of [
    /describes only observations actually persisted/i,
    /FREE_V1 ingestion cadence can leave historical replay evidence sparse/i,
    /high observation density does not prove uniform temporal coverage/i,
    /P90 is descriptive and is not an SLO or aviation-quality threshold/i,
    /largest-gap share is not a calibrated severity score/i,
    /one persisted observation cannot establish an inter-sample distribution/i,
    /no intermediate position, route, flight phase, aircraft intent or pilot behavior is inferred/i,
    /not ATC-grade, navigation-grade or safety-critical evidence/i,
    /does not add denser commercial history or paid aviation data/i,
  ]) {
    assert.match(document, limitation)
  }

  assert.match(document, /New paid aviation provider = NO/)
  assert.match(document, /Additional cost\s+= 0 RUB/)
  assert.match(document, /must not be fabricated/i)
  assert.match(document, /only defensible implementation requires a paid source or paid infrastructure/i)
})
