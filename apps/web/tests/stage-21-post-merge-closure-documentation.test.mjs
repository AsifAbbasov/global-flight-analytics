import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

const closure = read('../../docs/210_STAGE_21_AIRPORT_CONGESTION_INTELLIGENCE_POST_MERGE_CLOSURE.md')
const premerge = read('../../docs/207_STAGE_21_AIRPORT_CONGESTION_INTELLIGENCE.md')

test('Stage 21 closure protects exact merge identity', () => {
  assert.match(closure, /STAGE_21_AIRPORT_CONGESTION_INTELLIGENCE=CLOSED/)
  assert.match(closure, /STAGE_21_MERGED_PR=164/)
  assert.match(
    closure,
    /STAGE_21_APPROVED_HEAD=f95c88aeb60b0b4e6c2e0cf7bff83697f2a9ed7f/,
  )
  assert.match(
    closure,
    /STAGE_21_MAIN_SHA=63ff7c04cce26b73664dbecd307a50d9d6cdea04/,
  )
})

test('Stage 21 closure protects final pre-merge workflow evidence', () => {
  assert.match(closure, /OpenAPI Contract #133 \/ run 34080720938 = SUCCESS/)
  assert.match(closure, /Frontend CI #498 \/ run 34080721020 = SUCCESS/)
  assert.match(closure, /Backend CI #836 \/ run 34080720950 = SUCCESS/)
  assert.match(closure, /API Load Baseline #360 \/ run 34080720971 = SUCCESS/)
  assert.match(closure, /CodeQL #478 \/ run 34080720936 = SUCCESS/)
  assert.match(closure, /Playwright E2E #269 \/ run 34080720937 = SUCCESS/)
  assert.match(closure, /STAGE_21_PREMERGE_VERCEL=PASS/)
})

test('Stage 21 closure protects canonical post-merge evidence', () => {
  assert.match(closure, /OpenAPI Contract #134 \/ run 34083139299 = SUCCESS/)
  assert.match(closure, /Frontend CI #499 \/ run 34083139329 = SUCCESS/)
  assert.match(
    closure,
    /Backend CI #837 \/ run 34083139296 \/ attempt 2 = SUCCESS/,
  )
  assert.match(closure, /API Load Baseline #361 \/ run 34083139369 = SUCCESS/)
  assert.match(closure, /CodeQL #479 \/ run 34083139308 = SUCCESS/)
  assert.match(closure, /Playwright E2E #270 \/ run 34083139400 = SUCCESS/)
  assert.match(closure, /STAGE_21_POST_MERGE_VERCEL=PASS/)
  assert.match(closure, /VERCEL_DEPLOYMENT=AagXsgugSre4HQp2RY4hwyqJ6srG/)
})

test('Stage 21 closure preserves the real backend timeout and same-SHA retry', () => {
  assert.match(closure, /STAGE_21_BACKEND_INITIAL_POSTMERGE_ATTEMPT=TIMEOUT_CANCELLED/)
  assert.match(closure, /STAGE_21_BACKEND_RETRY_ATTEMPT=PASS/)
  assert.match(closure, /BACKEND_ATTEMPT_1_CONTAINER=TIMEOUT_CANCELLED/)
  assert.match(closure, /BACKEND_ATTEMPT_1_GATE=FAILURE/)
  assert.match(closure, /SOURCE_SHA_CHANGED_FOR_RETRY=NO/)
  assert.match(closure, /BACKEND_ATTEMPT_2_BUILD_CONTAINER=SUCCESS/)
  assert.match(closure, /BACKEND_ATTEMPT_2_CONTAINER_HEALTH_SMOKE=SUCCESS/)
  assert.match(closure, /BACKEND_ATTEMPT_2_WORKFLOW=SUCCESS/)
})

test('Stage 21 closure preserves relative observed-activity semantics', () => {
  assert.match(closure, /STAGE_21_ANALYTICAL_CLASS=RELATIVE_OBSERVED_ACTIVITY_PROXY/)
  assert.match(closure, /current = latest expected completed UTC day, only when that day is observed/)
  assert.match(closure, /congestion_score = min\(1, current_to_prior_peak_ratio\)/)
  assert.match(closure, /Missing evidence is not converted to zero/)
  assert.match(closure, /an older observed day is not silently substituted as current/)
})

test('Stage 21 closure rejects unsupported operational claims', () => {
  assert.match(closure, /STAGE_21_AIRPORT_CAPACITY_MODEL=NONE/)
  assert.match(closure, /STAGE_21_RUNWAY_OCCUPANCY_MODEL=NONE/)
  assert.match(closure, /STAGE_21_QUEUE_MODEL=NONE/)
  assert.match(closure, /STAGE_21_SLOT_MODEL=NONE/)
  assert.match(closure, /STAGE_21_DELAY_INFERENCE=NONE/)
  assert.match(closure, /STAGE_21_OFFICIAL_CONGESTION_CLAIM=NONE/)
  assert.match(closure, /LOW_MEDIUM_HIGH_CONGESTION_THRESHOLDS=NONE/)
})

test('Stage 21 closure preserves zero-budget and no-new-data boundary', () => {
  assert.match(closure, /STAGE_21_NEW_PROVIDER=NONE/)
  assert.match(closure, /STAGE_21_NEW_INGESTION_PATH=NONE/)
  assert.match(closure, /STAGE_21_NEW_DATABASE_DATA=NONE/)
  assert.match(closure, /STAGE_21_ADDITIONAL_COST=0_RUB/)
  assert.match(closure, /NEW_DATABASE_TABLE=NO/)
  assert.match(closure, /NEW_MIGRATION=NO/)
  assert.match(closure, /NEW_PAID_SERVICE=NO/)
})

test('Document 207 remains historical rather than retroactively rewritten', () => {
  assert.match(premerge, /STAGE_21_AIRPORT_CONGESTION_INTELLIGENCE=IN_PROGRESS/)
  assert.match(premerge, /STAGE_21_POST_MERGE_CI=PENDING/)
  assert.match(closure, /Document 207 remains the historical pre-merge engineering record/)
  assert.match(closure, /Document 210 is the canonical post-merge closure record/)
})
