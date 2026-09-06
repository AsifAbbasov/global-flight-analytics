import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const document = readFileSync(
  new URL(
    '../../../docs/205_STAGE_19_REPLAY_EVIDENCE_PROVENANCE_POST_MERGE_CLOSURE.md',
    import.meta.url
  ),
  'utf8'
)

test('Stage 19 post-merge closure preserves exact merge and provenance semantics', () => {
  for (const marker of [
    'STAGE_19_REPLAY_EVIDENCE_PROVENANCE=CLOSED',
    'STAGE_19_MERGED_PR=161',
    'STAGE_19_APPROVED_HEAD=2405e76a5e8be9ab78d6165a7f68ce66eb5c4cba',
    'STAGE_19_MAIN_SHA=64f6198bd96087d7bc479a78e1839db2f6914c68',
    'STAGE_19_POST_MERGE_CI=PASS',
    'STAGE_19_EVIDENCE_CLASS=PERSISTED_SOURCE_LABELS_ONLY',
    'STAGE_19_PROVIDER_RANKING=NONE',
    'STAGE_19_PROVIDER_ACCURACY_SCORE=NONE',
    'STAGE_19_PROVIDER_SWITCH_TIME_INFERENCE=NONE',
    'STAGE_19_NEW_BACKEND_DATA=NONE',
    'STAGE_19_ADDITIONAL_COST=0_RUB',
    'STAGE_19_DOCUMENTATION=CLOSED',
  ]) {
    assert.match(document, new RegExp(marker))
  }
})

test('Stage 19 closure records final and post-merge validation matrices', () => {
  for (const evidence of [
    'Frontend CI #440',
    'run=34050341592',
    'Backend CI #778',
    'run=34050341638',
    'CodeQL #420',
    'run=34050341680',
    'API Load Baseline #306',
    'run=34050341750',
    'Playwright E2E #217',
    'run=34050341642',
    'Frontend CI #441',
    'run=34054495884',
    'Backend CI #779',
    'run=34054496033',
    'CodeQL #421',
    'run=34054495862',
    'Playwright E2E #218',
    'run=34054495873',
    'Vercel',
    'result=SUCCESS',
  ]) {
    assert.match(document, new RegExp(evidence.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }

  assert.match(document, /existing workflow\/event\/path policy does not require it/i)
  assert.match(document, /expected behavior, not a missing or failed check/i)
})

test('Stage 19 closure preserves historical pre-merge truth and real remediation history', () => {
  assert.match(document, /Document 204 records the Stage 19 product need/i)
  assert.match(document, /appends the post-merge truth instead of rewriting Document 204/i)
  assert.match(document, /Document 204 therefore remains a historical pre-merge snapshot/i)
  assert.match(document, /Frontend CI #437 \/ run `34050031425`/)
  assert.match(document, /190 total tests, 189 passing and one failing/i)
  assert.match(document, /source-contract regular expression was formatting-sensitive/i)
  assert.match(document, /the\\s\+exact provider switch time between those observations is unknown/)
  assert.match(document, /No additional product or architecture defect was discovered after merge/i)
})

test('Stage 19 closure keeps provenance-only and zero-budget limitations explicit', () => {
  for (const limitation of [
    /provenance describes only source labels actually persisted/i,
    /source label does not prove provider accuracy, completeness, reliability or superiority/i,
    /persisted sample share does not prove elapsed-time coverage or provider uptime/i,
    /exact provider switch instant inside that interval remains unknown/i,
    /unknown provenance prevents a fabricated direct transition/i,
    /request-level fallback reasons are not reconstructed/i,
    /no continuous provider ownership is projected across unobserved intervals/i,
    /not ATC-, navigation- or safety-grade provenance certification/i,
  ]) {
    assert.match(document, limitation)
  }

  assert.match(document, /New paid aviation provider = NO/)
  assert.match(document, /Additional cost\s+= 0 RUB/)
  assert.match(document, /must never be silently converted into provider quality/i)
  assert.match(document, /must not be fabricated/i)
  assert.match(document, /only defensible implementation requires a paid source or paid infrastructure/i)
})
