import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const provenanceModel = readFileSync(
  new URL('../lib/replay/flight-replay-evidence-provenance-model.ts', import.meta.url),
  'utf8'
)
const provenanceComponent = readFileSync(
  new URL('../components/aircraft/flight-replay-evidence-provenance.tsx', import.meta.url),
  'utf8'
)
const timeNavigation = readFileSync(
  new URL('../components/aircraft/flight-replay-time-navigation.tsx', import.meta.url),
  'utf8'
)
const testConfig = readFileSync(new URL('../tsconfig.test.json', import.meta.url), 'utf8')

test('Stage 19 uses persisted replay source labels without a second data path', () => {
  assert.match(provenanceModel, /point\.source_name/)
  assert.match(provenanceModel, /point\.observed_at/)
  assert.doesNotMatch(provenanceModel, /fetch\(/)
  assert.doesNotMatch(provenanceModel, /axios/)
  assert.doesNotMatch(
    provenanceModel,
    /providerRank|accuracyScore|qualityScore|confidenceScore|preferredProvider/
  )
})

test('Stage 19 UI preserves descriptive provenance semantics', () => {
  for (const marker of [
    "aria-label='Replay evidence provenance profile'",
    "data-flight-replay-provenance='persisted-source-labels-only'",
    "data-flight-replay-provider-ranking='none'",
    "data-flight-replay-provider-accuracy-claim='none'",
    'Replay evidence provenance',
    'Observed provenance',
    'Source composition',
    'Adjacent source-label transitions',
    'Percentages are shares of persisted samples, not shares of elapsed time',
  ]) {
    assert.match(provenanceComponent, new RegExp(marker))
  }

  assert.match(
    provenanceComponent,
    /does not\s+rank provider accuracy, infer a better source, or reconstruct the exact switch instant/
  )
  assert.match(
    provenanceComponent,
    /the\s+exact provider switch time between those observations is unknown/
  )
})

test('Stage 19 integrates with existing replay navigation and compiles the real model', () => {
  assert.match(timeNavigation, /FlightReplayEvidenceProvenance/)
  assert.match(timeNavigation, /flightReplayObservationCursorSeconds/)
  assert.match(timeNavigation, /<FlightReplayEvidenceProvenance/)
  assert.match(
    testConfig,
    /lib\/replay\/flight-replay-evidence-provenance-model\.ts/
  )
  assert.doesNotMatch(timeNavigation, /fetch\(/)
  assert.doesNotMatch(timeNavigation, /axios/)
})
