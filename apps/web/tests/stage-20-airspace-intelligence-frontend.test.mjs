import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

function read(relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8')
}

const regionalExperience = read('components/regional-traffic-experience.tsx')
const workspace = read('components/analytics/airspace-intelligence-workspace.tsx')
const panel = read('components/traffic/airspace-intelligence-panel.tsx')
const query = read('lib/queries/airspace-intelligence.ts')
const api = read('lib/api/airspace-intelligence.ts')
const model = read('lib/airspace/airspace-intelligence-model.ts')

test('Stage 20 exposes existing Airspace Intelligence through the regional product flow', () => {
  assert.match(regionalExperience, /FRONTEND_AIRSPACE_INTELLIGENCE_V1/)
  assert.match(regionalExperience, /AirspaceIntelligenceWorkspace/)
  assert.match(regionalExperience, /id='airspace-intelligence'/)
  assert.match(workspace, /useCurrentTraffic\(selectedRegion\.code\)/)
  assert.match(workspace, /resolveAirspaceAsOfTime/)
  assert.match(workspace, /useAirspaceRegionAnalytics/)
})

test('Stage 20 uses the existing read-only endpoint and no second raw network path', () => {
  assert.match(
    api,
    /\/api\/v1\/airspace\/regions\/\$\{encodeURIComponent\(regionCode\)\}\/analytics/
  )
  assert.match(api, /requestAPIData<unknown>/)
  assert.doesNotMatch(workspace, /\bfetch\s*\(/)
  assert.doesNotMatch(workspace, /axios/i)
  assert.doesNotMatch(panel, /\bfetch\s*\(/)
  assert.doesNotMatch(panel, /axios/i)
  assert.doesNotMatch(query, /\bfetch\s*\(/)
  assert.doesNotMatch(query, /axios/i)
})

test('Stage 20 preserves bounded-region and observed-time evidence semantics', () => {
  assert.match(model, /observed_at/)
  assert.match(model, /normalizedRegionCode !== 'world'/)
  assert.match(
    panel,
    /World option is a frontend-wide traffic view and is not treated as an[\s\S]*analytical airspace region/
  )
  assert.match(
    panel,
    /Waiting for an observed regional traffic timestamp before requesting[\s\S]*Airspace Intelligence/
  )
})

test('Stage 20 preserves research-only airspace scope guards', () => {
  assert.match(panel, /not air traffic control guidance or certified separation[\s\S]*monitoring/)
  assert.match(panel, /do not represent[\s\S]*official sectors/)
  assert.match(panel, /controller workload/)
  assert.match(panel, /regulatory separation minima/)
  assert.match(panel, /collision prediction/)
  assert.match(panel, /safety-critical guidance/)
  assert.doesNotMatch(panel, /safe to fly/i)
})

test('Stage 20 renders backend-owned metrics, confidence, limitations and provenance', () => {
  for (const token of [
    'current_aircraft_count',
    'unique_aircraft_count',
    'temporal_coverage',
    'mean_complexity_score',
    'peak_complexity_score',
    'airspace_pressure_index',
    'peak_airspace_pressure_index',
    'occupancy_trend',
    'highest_complexity_level',
    'confidence',
    'limitations',
    'explanations',
    'source_names',
    'latest_observed_at',
    'input_fingerprint',
  ]) {
    assert.match(panel + api, new RegExp(token))
  }
})

test('Stage 20 does not invent an occupancy heatmap without a proven geometry contract', () => {
  assert.doesNotMatch(regionalExperience + workspace + panel, /heatmap/i)
  assert.doesNotMatch(regionalExperience + workspace + panel, /latitude_index/)
  assert.doesNotMatch(regionalExperience + workspace + panel, /longitude_index/)
  assert.doesNotMatch(regionalExperience + workspace + panel, /altitude_band_index/)
})
