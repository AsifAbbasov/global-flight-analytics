import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

const document = read('../../docs/214_STAGE_23_ETA_RELIABILITY_INTELLIGENCE.md')
const index = read('../../docs/DOCUMENT_INDEX.md')
const roadmap = read('../../docs/24_MVP_VERSION_ROADMAP.md')
const readme = read('../../README.md')

test('Stage 23 documentation protects product-first vertical slice', () => {
  assert.match(document, /STAGE_23_ETA_RELIABILITY=IN_PROGRESS/)
  assert.match(document, /STAGE_23_FRONTEND_CONSUMER=YES/)
  assert.match(document, /STAGE_23_DATA_FOR_DATA_SAKE=NO/)
  assert.match(document, /STAGE_23_ARCHITECTURE_FOR_ARCHITECTURE_SAKE=NO/)
  assert.match(document, /Historical ETA Reliability/)
  assert.match(document, /Aircraft Detail \/ Projection Intelligence/)
})

test('Stage 23 documentation protects honest historical evidence semantics', () => {
  assert.match(
    document,
    /STAGE_23_EVIDENCE_CLASS=HISTORICALLY_RECOMPUTED_FROM_PERSISTED_OBSERVATIONS/,
  )
  assert.match(document, /STAGE_23_OFFICIAL_ARRIVAL_TRUTH=NONE/)
  assert.match(document, /STAGE_23_HISTORICAL_ARRIVAL_EVIDENCE=PERSISTED_ENDPOINT_PROXY/)
  assert.match(document, /STAGE_23_TOUCHDOWN_TIME_CLAIM=NONE/)
  assert.match(document, /STAGE_23_GATE_TIME_CLAIM=NONE/)
  assert.match(document, /STAGE_23_OPERATIONAL_GUIDANCE=NONE/)
  assert.match(document, /STAGE_23_PERSISTED_FORECAST_HISTORY=NONE/)
  assert.match(document, /STAGE_23_SYNTHETIC_AS_OF_TIME=NONE/)
})

test('Stage 23 documentation protects bounded zero-cost computation', () => {
  assert.match(document, /STAGE_23_ADDITIONAL_COST=0_RUB/)
  assert.match(document, /STAGE_23_NEW_PAID_PROVIDER=NONE/)
  assert.match(document, /STAGE_23_NEW_EXTERNAL_API=NONE/)
  assert.match(document, /STAGE_23_NEW_DATABASE_TABLE=NONE/)
  assert.match(document, /STAGE_23_NEW_MIGRATION=NONE/)
  assert.match(document, /STAGE_23_NEW_SERVER=NONE/)
  assert.match(document, /STAGE_23_NEW_BACKGROUND_MATERIALIZER=NONE/)
  assert.match(document, /STAGE_23_MAX_HISTORICAL_CANDIDATES=8/)
  assert.match(document, /STAGE_23_FRONTEND_POLLING=NONE/)
})

test('Stage 23 documentation preserves sample-size and proxy disclosure', () => {
  assert.match(document, /STAGE_23_ENDPOINT_PROXY_DISCLOSURE=REQUIRED/)
  assert.match(document, /STAGE_23_SAMPLE_SIZE_DISCLOSURE=REQUIRED/)
  assert.match(document, /eligible historical sample count/)
  assert.match(document, /unavailable or limited evidence state/)
  assert.match(document, /The frontend must not say or imply:/)
  assert.match(document, /ETA accuracy guaranteed/)
  assert.match(document, /actual landing time/)
  assert.match(document, /official arrival accuracy/)
})

test('Stage 23 documentation preserves rejected validation history', () => {
  assert.match(document, /0d1f49abe255459a45dac67bbdd7e3d0fd117106/)
  assert.match(document, /Frontend CI #513 \/ run 34119546928 = SUCCESS/)
  assert.match(document, /OpenAPI Contract #142 \/ run 34119546967 = FAILURE/)
  assert.match(document, /API Load Baseline #370 \/ run 34119546975 = FAILURE/)
  assert.match(document, /dto\.ETAReliabilityResponse did not satisfy response\.SuccessPayload/)

  assert.match(document, /a9710c428e8988f4099b51b55702fe56cc203444/)
  assert.match(document, /Frontend CI #514 \/ run 34119801776 = SUCCESS/)
  assert.match(document, /API Load Baseline #371 \/ run 34119801694 = SUCCESS/)
  assert.match(document, /Playwright E2E #285 \/ run 34119801740 = SUCCESS/)
  assert.match(document, /CodeQL #495 \/ run 34119801725 = SUCCESS/)
  assert.match(document, /Backend CI #852 \/ run 34119801824 = FAILURE/)
  assert.match(document, /OpenAPI Contract #143 \/ run 34119801804 = FAILURE/)
  assert.match(document, /Verify Go formatting/)
})

test('Stage 23 documentation refuses premature merge or closure claims', () => {
  assert.match(document, /STAGE_23_PR_DRAFT=YES/)
  assert.match(document, /STAGE_23_REVIEW_READY=NO/)
  assert.match(document, /STAGE_23_MERGE_READY=NO/)
  assert.match(document, /STAGE_23_MERGE_AUTHORIZATION=NOT_GRANTED/)
  assert.match(document, /STAGE_23_EXACT_HEAD_FINAL_CI=NOT_YET_AVAILABLE/)
  assert.match(document, /STAGE_23_CLOSED=NO/)
})

test('Stage 23 is registered in canonical documentation surfaces', () => {
  assert.match(index, /Documentation Index v2\.4/)
  assert.match(index, /## Document 214 — Stage 23 ETA Reliability Intelligence/)
  assert.match(index, /214_STAGE_23_ETA_RELIABILITY_INTELLIGENCE\.md/)

  assert.match(roadmap, /Architecture Baseline v1\.3/)
  assert.match(roadmap, /## 21\. Stage 23 Product Increment — ETA Reliability Intelligence/)
  assert.match(roadmap, /VERSION_2_RELEASE_CLOSURE=CLOSED/)
  assert.match(roadmap, /STAGE_23_ADDITIONAL_COST=0_RUB/)
  assert.match(roadmap, /FRONTEND_CONSUMER=REQUIRED/)
  assert.match(roadmap, /No future Stage 24 is implied by this roadmap amendment/)

  assert.match(readme, /<!-- STAGE-23-ETA-RELIABILITY:README -->/)
  assert.match(readme, /## Stage 23 — ETA Reliability Intelligence/)
  assert.match(readme, /STAGE_23_ETA_RELIABILITY=IN_PROGRESS/)
  assert.match(readme, /STAGE_23_ADDITIONAL_COST=0_RUB/)
  assert.match(readme, /PERSISTED_ENDPOINT_PROXY/)
  assert.match(readme, /OFFICIAL_ARRIVAL_TRUTH=NONE/)
  assert.match(readme, /VERSION_2_RECONCILIATION=CLOSED/)
})

test('Stage 23 canonical document records documentation alignment without synthetic finding creation', () => {
  assert.match(document, /STAGE_23_DOCUMENTATION_ALIGNMENT=COMPLETE_FOR_IN_PROGRESS_STATE/)
  assert.match(document, /STAGE_23_DOCUMENT_214=ALIGNED_IN_PROGRESS/)
  assert.match(document, /STAGE_23_DOCUMENT_INDEX=ALIGNED_V2_4/)
  assert.match(document, /STAGE_23_README=ALIGNED_IN_PROGRESS/)
  assert.match(document, /STAGE_23_ROADMAP=ALIGNED_V1_3/)
  assert.match(document, /No synthetic finding ID is created merely because a product feature exists/)
})
