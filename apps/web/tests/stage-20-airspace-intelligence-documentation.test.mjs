import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const document = readFileSync(
  new URL(
    '../../../docs/206_STAGE_20_AIRSPACE_INTELLIGENCE_FRONTEND_INTEGRATION.md',
    import.meta.url
  ),
  'utf8'
)

test('Stage 20 document protects product purpose and evidence boundaries', () => {
  for (const marker of [
    'STAGE_20_AIRSPACE_INTELLIGENCE_FRONTEND=IN_PROGRESS',
    'STAGE_20_BASE_MAIN=84989c72e4cc8fddb79e484f2c7de2e6e597ba0f',
    'STAGE_20_BACKEND_ANALYTICS_RECOMPUTATION=NONE',
    'STAGE_20_NEW_BACKEND_ENDPOINT=NONE',
    'STAGE_20_NEW_DATABASE_DATA=NONE',
    'STAGE_20_WORLD_ANALYTICS=DISABLED',
    'STAGE_20_AS_OF_SOURCE=LATEST_VALID_TRAFFIC_OBSERVED_AT',
    'STAGE_20_HEATMAP=DEFERRED_UNVERIFIED_GEOMETRY',
    'STAGE_20_OPERATIONAL_AIRSPACE_CLAIMS=NONE',
    'STAGE_20_ADDITIONAL_COST=0_RUB',
    'GET /api/v1/airspace/regions/{code}/analytics',
    'AIRSPACE_ANALYTICS=RESEARCH_ONLY',
    'OFFICIAL_SECTOR_CLAIM=NONE',
    'ATC_SUPPORT_CLAIM=NONE',
    'COLLISION_PREDICTION_CLAIM=NONE',
    'SYNTHETIC_CELL_GEOMETRY=PROHIBITED',
  ]) {
    assert.match(document, new RegExp(marker.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  }
})

test('Stage 20 document records validation and zero-budget scope honestly', () => {
  assert.match(document, /Continuous Integration and browser validation[\s\S]*remain pending/)
  assert.match(document, /final exact head/)
  assert.match(document, /New paid provider\s+= NO/)
  assert.match(document, /New PostgreSQL table\s+= NO/)
  assert.match(document, /New external aviation call\s+= NO/)
  assert.match(document, /Additional cost\s+= 0 RUB/)
  assert.match(document, /post-merge closure document must append the actual merge/)
})
