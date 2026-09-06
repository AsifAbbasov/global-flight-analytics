import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const intervalModel = readFileSync(
  new URL('../lib/replay/flight-replay-interval-model.ts', import.meta.url),
  'utf8'
)
const intervalComponent = readFileSync(
  new URL('../components/aircraft/flight-replay-interval-comparison.tsx', import.meta.url),
  'utf8'
)
const timeNavigation = readFileSync(
  new URL('../components/aircraft/flight-replay-time-navigation.tsx', import.meta.url),
  'utf8'
)

test('Stage 17 compares persisted observations without adding a second evidence math implementation', () => {
  assert.match(intervalModel, /buildFlightReplayObservedChange/)
  assert.match(intervalModel, /buildFlightReplayGaps/)
  assert.match(intervalModel, /points: \[start, end\]/)
  assert.doesNotMatch(intervalModel, /interpolat(e|ion)/i)
  assert.doesNotMatch(intervalModel, /travelledPath|traveledPath|routeDistance/)
})

test('Stage 17 UI exposes interval quality and explicitly refuses path inference', () => {
  for (const marker of [
    "aria-label='Observed interval comparison'",
    "data-flight-replay-interval-evidence='observed-endpoints-only'",
    "data-flight-replay-interval-interpolation='none'",
    "data-flight-replay-path-distance='not-claimed'",
    'Evidence samples',
    'Largest internal gap',
    'Endpoint displacement',
    'not travelled path distance',
    'no intermediate coordinate, route, phase or intent is synthesized',
    'At least two persisted observations are required for interval comparison',
  ]) {
    assert.match(intervalComponent, new RegExp(marker))
  }
})

test('Stage 17 integration stays inside existing replay time navigation', () => {
  assert.match(timeNavigation, /FlightReplayIntervalComparison/)
  assert.match(timeNavigation, /<FlightReplayIntervalComparison replay=\{replay\} \/>/)
  assert.doesNotMatch(timeNavigation, /fetch\(/)
  assert.doesNotMatch(timeNavigation, /axios/)
})
