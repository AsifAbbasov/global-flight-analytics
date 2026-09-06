import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const qualityModel = readFileSync(
  new URL('../lib/replay/flight-replay-evidence-quality-model.ts', import.meta.url),
  'utf8'
)
const qualityComponent = readFileSync(
  new URL('../components/aircraft/flight-replay-evidence-quality.tsx', import.meta.url),
  'utf8'
)
const timeNavigation = readFileSync(
  new URL('../components/aircraft/flight-replay-time-navigation.tsx', import.meta.url),
  'utf8'
)
const testConfig = readFileSync(new URL('../tsconfig.test.json', import.meta.url), 'utf8')

test('Stage 18 reuses persisted replay evidence instead of creating a second data path', () => {
  assert.match(qualityModel, /buildFlightReplayAnalyticsSummary/)
  assert.match(qualityModel, /buildFlightReplayGaps/)
  assert.doesNotMatch(qualityModel, /fetch\(/)
  assert.doesNotMatch(qualityModel, /axios/)
  assert.doesNotMatch(qualityModel, /qualityScore|confidenceScore|goodThreshold|badThreshold/)
})

test('Stage 18 UI is descriptive only and exposes no synthetic quality grade', () => {
  for (const marker of [
    "aria-label='Replay evidence quality profile'",
    "data-flight-replay-evidence-quality='descriptive-only'",
    "data-flight-replay-evidence-score='none'",
    'Replay evidence quality',
    'No synthetic score',
    'Mean gap',
    'Median gap',
    'P90 gap',
    'Largest-gap share',
    'Observation density',
    'Altitude evidence',
    'Largest unobserved interval',
    'No persisted position exists between these endpoint observations',
    'not calibrated aviation quality grades',
  ]) {
    assert.match(qualityComponent, new RegExp(marker))
  }
})

test('Stage 18 keeps the single-sample state honest', () => {
  assert.match(
    qualityComponent,
    /Only one persisted observation is available, so sampling density and gap distribution\s+cannot be measured/
  )
  assert.match(qualityComponent, /Unavailable/)
})

test('Stage 18 integrates inside existing replay navigation and compiles the real model', () => {
  assert.match(timeNavigation, /FlightReplayEvidenceQuality/)
  assert.match(timeNavigation, /<FlightReplayEvidenceQuality replay=\{replay\} \/>/)
  assert.match(testConfig, /lib\/replay\/flight-replay-evidence-quality-model\.ts/)
  assert.doesNotMatch(timeNavigation, /fetch\(/)
  assert.doesNotMatch(timeNavigation, /axios/)
})