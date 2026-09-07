import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

const closure = read('../../docs/211_STAGE_22_ETA_EVOLUTION_POST_MERGE_CLOSURE.md')
const premerge = read('../../docs/208_STAGE_22_ETA_EVOLUTION_ANALYZER.md')

test('Stage 22 closure protects exact feature merge identity', () => {
  assert.match(closure, /STAGE_22_ETA_EVOLUTION=CLOSURE_CANDIDATE/)
  assert.match(closure, /STAGE_22_MERGED_PR=165/)
  assert.match(
    closure,
    /STAGE_22_APPROVED_HEAD=eb8a2b2b7680b8e4590c520c16d60954a0cfd3a2/,
  )
  assert.match(
    closure,
    /STAGE_22_FEATURE_MAIN_SHA=1020b06af9827057ac9296ef09914c232cd7f510/,
  )
  assert.match(
    closure,
    /STAGE_22_PARENT_MAIN=e2189b8e38741abf1e4b87e1c937e6b623759f41/,
  )
})

test('Stage 22 closure protects final pre-merge exact-head evidence', () => {
  assert.match(closure, /Frontend CI #502 \/ run 34090595253 = SUCCESS/)
  assert.match(closure, /Backend CI #840 \/ run 34090595257 = SUCCESS/)
  assert.match(closure, /API Load Baseline #363 \/ run 34090595277 = SUCCESS/)
  assert.match(closure, /CodeQL #482 \/ run 34090595226 = SUCCESS/)
  assert.match(closure, /Playwright E2E #273 \/ run 34090595329 = SUCCESS/)
  assert.match(closure, /OpenAPI Contract = NOT_TRIGGERED_FRONTEND_ONLY_PATH_POLICY/)
  assert.match(closure, /STAGE_22_PREMERGE_VERCEL=PASS/)
  assert.match(closure, /VERCEL_PREMERGE_DEPLOYMENT=CiwGkDwTNBWRKDtmJgXXt9XcY3bN/)
})

test('Stage 22 closure protects canonical feature post-merge evidence', () => {
  assert.match(closure, /Frontend CI #503 \/ run 34092099738 = SUCCESS/)
  assert.match(closure, /Backend CI #841 \/ run 34092099782 = SUCCESS/)
  assert.match(closure, /CodeQL #483 \/ run 34092099737 = SUCCESS/)
  assert.match(closure, /Playwright E2E #274 \/ run 34092099739 = SUCCESS/)
  assert.match(
    closure,
    /STAGE_22_POST_MERGE_API_LOAD=NOT_TRIGGERED_FRONTEND_ONLY_PATH_POLICY/,
  )
  assert.match(
    closure,
    /STAGE_22_POST_MERGE_OPENAPI=NOT_TRIGGERED_FRONTEND_ONLY_PATH_POLICY/,
  )
  assert.match(closure, /STAGE_22_POST_MERGE_FEATURE_VERCEL=PASS/)
  assert.match(closure, /VERCEL_POSTMERGE_DEPLOYMENT=HXW8DXQXiGZqhR94bdwJd3GY2hYA/)
  assert.match(closure, /VERCEL_POSTMERGE_STATUS=Deployment has completed/)
})

test('Stage 22 closure preserves historically recomputed evidence semantics', () => {
  assert.match(
    closure,
    /STAGE_22_EVIDENCE_CLASS=HISTORICALLY_RECOMPUTED_FROM_PERSISTED_OBSERVATIONS/,
  )
  assert.match(closure, /STAGE_22_PERSISTED_FORECAST_HISTORY=NONE/)
  assert.match(closure, /STAGE_22_ETA_INTERPOLATION=NONE/)
  assert.match(closure, /STAGE_22_CAUSE_INFERENCE=NONE/)
  assert.match(closure, /STAGE_22_OPERATIONAL_GUIDANCE=NONE/)
  assert.match(closure, /STAGE_22_SAMPLE_CAP=6/)
  assert.match(closure, /MAX_PROJECTION_RECOMPUTATIONS=6/)
  assert.match(closure, /SYNTHETIC_TIME_SAMPLE=NONE/)
  assert.match(closure, /CARRY_FORWARD=NONE/)
  assert.match(closure, /GAP_BRIDGING=NONE/)
})

test('Stage 22 closure preserves the real validation rejection history', () => {
  assert.match(closure, /Frontend CI #487 \/ run `34063552010` failed/)
  assert.match(closure, /SOURCE_CONTRACT_FALSE_POSITIVE=YES/)
  assert.match(closure, /FALSE_MATCH=query\.refetch\(\)/)
  assert.match(closure, /Playwright E2E #263 \/ run `34063846532` failed/)
  assert.match(closure, /PLAYWRIGHT_FOUNDATION_POLICY_VIOLATION=YES/)
  assert.match(
    closure,
    /REMEDIATION=getByRole\('complementary', \{ name: 'Estimated Arrival Evolution' \}\)/,
  )
})

test('Stage 22 closure rejects unsupported operational and persistence claims', () => {
  assert.match(closure, /PERSISTED_HISTORICAL_FORECAST_OUTPUT=NO/)
  assert.match(closure, /CAUSE_OF_ETA_CHANGE=UNKNOWN/)
  assert.match(closure, /DELAY_CAUSE_INFERENCE=NONE/)
  assert.match(closure, /ATC_CAUSE_INFERENCE=NONE/)
  assert.match(closure, /WEATHER_CAUSE_INFERENCE=NONE/)
  assert.match(closure, /AIRPORT_CONGESTION_CAUSE_INFERENCE=NONE/)
  assert.match(closure, /OPERATIONAL_GUIDANCE=NONE/)
})

test('Stage 22 closure preserves zero-budget and no-new-backend-data boundary', () => {
  assert.match(closure, /STAGE_22_NEW_BACKEND_ENDPOINT=NONE/)
  assert.match(closure, /STAGE_22_NEW_DATABASE_TABLE=NONE/)
  assert.match(closure, /STAGE_22_NEW_MIGRATION=NONE/)
  assert.match(closure, /STAGE_22_NEW_PROVIDER=NONE/)
  assert.match(closure, /STAGE_22_ADDITIONAL_COST=0_RUB/)
  assert.match(closure, /NEW_PAID_SERVICE=NO/)
})

test('Document 208 remains immutable historical pre-merge evidence', () => {
  assert.match(premerge, /STAGE_22_ETA_EVOLUTION=IN_PROGRESS/)
  assert.match(premerge, /STAGE_22_FINAL_EXACT_HEAD_CI=PENDING/)
  assert.match(premerge, /STAGE_22_POST_MERGE_CI=PENDING/)
  assert.match(closure, /Document 208 is the immutable Stage 22 pre-merge engineering-history record/)
  assert.match(closure, /Document 208 is not rewritten retroactively/)
})

test('Stage 22 formal closure remains gated on closure PR merge and final main verification', () => {
  assert.match(
    closure,
    /STAGE_22_DOCUMENTATION_CLOSURE=AWAITING_CLOSURE_PR_MERGE/,
  )
  assert.match(
    closure,
    /Formal closure requires this record and its permanent regression test to be merged into canonical `main`/,
  )
  assert.match(closure, /STAGE_22_ETA_EVOLUTION=CLOSED/)
  assert.match(closure, /STAGE_22_CI_VERIFIED=YES/)
})
