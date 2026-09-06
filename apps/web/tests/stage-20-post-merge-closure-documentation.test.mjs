import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

const closure = read('../../docs/209_STAGE_20_AIRSPACE_INTELLIGENCE_POST_MERGE_CLOSURE.md')
const premerge = read('../../docs/206_STAGE_20_AIRSPACE_INTELLIGENCE_FRONTEND_INTEGRATION.md')

test('Stage 20 closure protects exact merge identity', () => {
  assert.match(closure, /STAGE_20_AIRSPACE_INTELLIGENCE_FRONTEND=CLOSED/)
  assert.match(closure, /STAGE_20_MERGED_PR=163/)
  assert.match(
    closure,
    /STAGE_20_APPROVED_HEAD=9556388d7cf9e9501f1d49fc6c39b263bfa6cffe/,
  )
  assert.match(
    closure,
    /STAGE_20_MAIN_SHA=d544b7cc2d5d551c262363102b2cddd9db4b2ec3/,
  )
})

test('Stage 20 closure protects final pre-merge workflow evidence', () => {
  assert.match(closure, /Frontend CI #448 \/ run 34056870283 = SUCCESS/)
  assert.match(closure, /Backend CI #786 \/ run 34056870279 = SUCCESS/)
  assert.match(closure, /CodeQL #428 \/ run 34056870268 = SUCCESS/)
  assert.match(closure, /API Load Baseline #312 \/ run 34056870276 = SUCCESS/)
  assert.match(closure, /Playwright E2E #225 \/ run 34056870310 = SUCCESS/)
})

test('Stage 20 closure protects canonical post-merge evidence', () => {
  assert.match(closure, /Frontend CI #495 \/ run 34064784852 = SUCCESS/)
  assert.match(closure, /Backend CI #833 \/ run 34064784891 = SUCCESS/)
  assert.match(closure, /CodeQL #475 \/ run 34064784848 = SUCCESS/)
  assert.match(closure, /Playwright E2E #266 \/ run 34064784850 = SUCCESS/)
  assert.match(closure, /STAGE_20_POST_MERGE_VERCEL=PASS/)
  assert.match(closure, /CHUchXoB7BkqVLUy8BF2HySunwLn/)
  assert.match(
    closure,
    /API Load Baseline does not run on this post-merge push under the existing workflow event\/path policy/,
  )
})

test('Stage 20 closure preserves the real first browser failure and remediation', () => {
  assert.match(closure, /Playwright E2E #221 \/ 34056126301 = FAILURE/)
  assert.match(closure, /RAW_64_HEX=ACCEPTED/)
  assert.match(closure, /SHA256_PREFIX_PLUS_64_HEX=ACCEPTED/)
  assert.match(closure, /ARBITRARY_FINGERPRINT_STRING=REJECTED/)
  assert.match(premerge, /Playwright E2E #221 \/ run 34056126301 = FAILURE/)
  assert.match(premerge, /HEAD=b803d6e6515877ed25d9fd5851bfbbc1f4f07129/)
})

test('Stage 20 closure preserves evidence-time and bounded-region semantics', () => {
  assert.match(closure, /STAGE_20_AS_OF_SOURCE=LATEST_VALID_TRAFFIC_OBSERVED_AT/)
  assert.match(closure, /BROWSER_NOW_AS_ANALYTICAL_EVIDENCE=NO/)
  assert.match(closure, /STAGE_20_WORLD_ANALYTICS_SUBSTITUTION=NONE/)
  assert.match(closure, /WORLD_AIRSPACE_ANALYTICS=NOT_CLAIMED/)
  assert.match(closure, /HIDDEN_REGION_SUBSTITUTION=NONE/)
})

test('Stage 20 closure rejects invented geometry and operational claims', () => {
  assert.match(closure, /STAGE_20_HEATMAP=DEFERRED_UNVERIFIED_GEOMETRY/)
  assert.match(closure, /SYNTHETIC_CELL_GEOMETRY=PROHIBITED/)
  assert.match(closure, /OFFICIAL_SECTOR_CLAIM=NONE/)
  assert.match(closure, /ATC_SUPPORT_CLAIM=NONE/)
  assert.match(closure, /CONTROLLER_WORKLOAD_CLAIM=NONE/)
  assert.match(closure, /REGULATORY_SEPARATION_CLAIM=NONE/)
  assert.match(closure, /COLLISION_PREDICTION_CLAIM=NONE/)
  assert.match(closure, /SAFETY_CRITICAL_GUIDANCE=NONE/)
})

test('Stage 20 closure preserves zero-budget and no-new-backend-data boundary', () => {
  assert.match(closure, /STAGE_20_NEW_BACKEND_DATA=NONE/)
  assert.match(closure, /STAGE_20_ADDITIONAL_COST=0_RUB/)
  assert.match(closure, /NEW_BACKEND_ENDPOINT=NO/)
  assert.match(closure, /NEW_POSTGRESQL_TABLE=NO/)
  assert.match(closure, /NEW_MIGRATION=NO/)
  assert.match(closure, /NEW_PAID_PROVIDER=NO/)
})

test('Document 206 remains explicitly historical rather than retroactively rewritten', () => {
  assert.match(premerge, /STAGE_20_AIRSPACE_INTELLIGENCE_FRONTEND=IN_PROGRESS/)
  assert.match(premerge, /A later post-merge closure document must append the actual merge and post-merge evidence without rewriting this pre-merge history/)
  assert.match(closure, /Document 206 remains the historical pre-merge engineering record/)
  assert.match(closure, /Document 209 is the canonical post-merge closure record/)
})
