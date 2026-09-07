import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'

const documentPath = '../../docs/207_STAGE_21_AIRPORT_CONGESTION_INTELLIGENCE.md'

function readDocument() {
  return fs.readFileSync(documentPath, 'utf8')
}

test('Stage 21 document preserves the research-only congestion boundary', () => {
  const document = readDocument()

  for (const marker of [
    'STAGE_21_ANALYTICAL_CLASS=RELATIVE_OBSERVED_ACTIVITY_PROXY',
    'STAGE_21_AIRPORT_CAPACITY_MODEL=NONE',
    'STAGE_21_RUNWAY_OCCUPANCY_MODEL=NONE',
    'STAGE_21_QUEUE_MODEL=NONE',
    'STAGE_21_SLOT_MODEL=NONE',
    'STAGE_21_DELAY_INFERENCE=NONE',
    'STAGE_21_OFFICIAL_CONGESTION_CLAIM=NONE',
    'STAGE_21_ADDITIONAL_COST=0_RUB',
  ]) {
    assert.match(document, new RegExp(marker.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }
})

test('Stage 21 document records requested-window and missing-latest evidence policy', () => {
  const document = readDocument()

  assert.match(document, /requested production window separately/i)
  assert.match(document, /trailing missing day/i)
  assert.match(document, /previous observed day is not silently substituted as current/i)
  assert.match(document, /Unknown is not converted to zero/i)
})

test('Stage 21 document retains real failed validation history', () => {
  const document = readDocument()

  assert.match(document, /Initial OpenAPI inventory failure/)
  assert.match(document, /Backend CI = FAILURE/)
  assert.match(document, /API Load Baseline = FAILURE/)
  assert.match(document, /Playwright E2E = FAILURE/)
  assert.match(document, /expected 38 embedded OpenAPI operations, got 39/)
})
