import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const document = readFileSync(
  new URL(
    '../../../docs/201_STAGE_17_OBSERVED_INTERVAL_COMPARISON_POST_MERGE_CLOSURE.md',
    import.meta.url
  ),
  'utf8'
)

test('Stage 17 post-merge closure preserves exact merge and evidence semantics', () => {
  for (const marker of [
    'STAGE_17_OBSERVED_INTERVAL_COMPARISON=CLOSED',
    'STAGE_17_MERGED_PR=157',
    'STAGE_17_APPROVED_HEAD=0ed82c4233ae53bd62e9e124add12e8e2b97f700',
    'STAGE_17_MAIN_SHA=566f3e676eca9c0edde2c94d151fdaba2812b007',
    'STAGE_17_POST_MERGE_CI=PASS',
    'STAGE_17_EVIDENCE_CLASS=OBSERVED_ENDPOINTS',
    'STAGE_17_POSITION_INTERPOLATION=NONE',
    'STAGE_17_PATH_DISTANCE_CLAIM=NONE',
    'STAGE_17_ADDITIONAL_COST=0_RUB',
    'STAGE_17_DOCUMENTATION=CLOSED',
  ]) {
    assert.match(document, new RegExp(marker))
  }
})

test('Stage 17 closure records the exact post-merge validation matrix', () => {
  for (const evidence of [
    'Frontend CI #425',
    'run=34043307643',
    'Backend CI #763',
    'run=34043307600',
    'CodeQL #405',
    'run=34043307583',
    'Playwright E2E #202',
    'run=34043307626',
    'Vercel',
    'result=SUCCESS',
    'API Load Baseline #294',
    'run `34039054637`',
  ]) {
    assert.match(document, new RegExp(evidence.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }

  assert.match(document, /existing workflow\/event\/path policy does not require it/i)
  assert.match(document, /expected behavior, not a missing or failed check/i)
})

test('Stage 17 closure preserves real remediation history instead of rewriting it', () => {
  assert.match(document, /Document 200 records the Stage 17 product need/i)
  assert.match(document, /appends the post-merge truth instead of rewriting Document 200/i)
  assert.match(document, /Frontend CI #420 \/ run `34038594838` failed/i)
  assert.match(document, /151 tests passing and 2 failing/i)
  assert.match(document, /new pure interval model was absent from the explicit `tsconfig\.test\.json` compilation scope/i)
  assert.match(document, /source-contract regex depended on one-line JSX whitespace/i)
  assert.match(document, /No additional architecture defect was discovered after merge/i)
})

test('Stage 17 closure keeps evidence and zero-budget limitations explicit', () => {
  for (const limitation of [
    /limited to observations actually persisted/i,
    /FREE_V1 ingestion cadence can leave intervals sparse/i,
    /endpoint displacement does not establish travelled path distance/i,
    /endpoint heading difference does not establish cumulative turns/i,
    /endpoint velocity delta does not establish acceleration/i,
    /no intermediate coordinate, route, flight phase, aircraft intent or pilot behavior is inferred/i,
    /not ATC-grade, navigation-grade or safety-critical data/i,
    /does not add commercial historical coverage or denser paid aviation data/i,
  ]) {
    assert.match(document, limitation)
  }

  assert.match(document, /New paid aviation provider = NO/)
  assert.match(document, /Additional cost\s+= 0 RUB/)
  assert.match(document, /must not be fabricated/i)
  assert.match(document, /only defensible implementation requires a paid source or paid infrastructure/i)
})
