import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

test('Stage 22 reuses persisted replay timestamps and existing Projection Intelligence transport', () => {
  const query = read('lib/queries/eta-evolution.ts')
  const model = read('lib/projection/eta-evolution-model.ts')

  assert.match(query, /getProjectionIntelligence/)
  assert.match(query, /selectETAEvolutionAsOfTimes/)
  assert.match(query, /staleTime:\s*Number\.POSITIVE_INFINITY/)
  assert.match(query, /refetchInterval:\s*false/)
  assert.doesNotMatch(query, /\bfetch\s*\(/)
  assert.doesNotMatch(query, /axios/i)
  assert.match(model, /etaEvolutionMaximumSamples\s*=\s*6/)
  assert.match(model, /point\.observed_at/)
})

test('Stage 22 UI discloses recomputation and rejects fake persisted history or cause inference', () => {
  const component = read('components/aircraft/eta-evolution-panel.tsx')

  assert.match(
    component,
    /data-eta-evolution-evidence='historically-recomputed-from-persisted-observations'/
  )
  assert.match(component, /data-eta-evolution-persisted-forecast-history='none'/)
  assert.match(component, /data-eta-evolution-interpolation='none'/)
  assert.match(component, /data-eta-evolution-cause-inference='none'/)
  assert.match(component, /not immutable forecast outputs stored at those past/)
  assert.match(component, /No ETA is interpolated between samples/)
  assert.match(component, /does not establish/)
  assert.doesNotMatch(component, /confirmed delay cause|ATC caused|weather caused|congestion caused/i)
})

test('Stage 22 remains integrated with the existing Projection Intelligence product flow', () => {
  const projectionPanel = read('components/aircraft/projection-intelligence-panel.tsx')

  assert.match(projectionPanel, /ETAEvolutionPanel/)
  assert.match(projectionPanel, /useLatestAircraftTrajectory/)
  assert.doesNotMatch(projectionPanel, /\bfetch\s*\(/)
  assert.doesNotMatch(projectionPanel, /axios/i)
})

test('Stage 22 browser journey protects recomputation evidence guards', () => {
  const browser = read('e2e/tests/advanced-intelligence.spec.mjs')

  assert.match(browser, /Estimated Arrival Evolution/)
  assert.match(browser, /historically-recomputed-from-persisted-observations/)
  assert.match(browser, /data-eta-evolution-persisted-forecast-history/)
  assert.match(browser, /data-eta-evolution-interpolation/)
  assert.match(browser, /data-eta-evolution-cause-inference/)
  assert.match(browser, /No ETA is interpolated between samples/)
})

test('Stage 22 documentation preserves recomputed-history and zero-budget boundaries', () => {
  const document = read('../../docs/208_STAGE_22_ETA_EVOLUTION_ANALYZER.md')

  assert.match(document, /STAGE_22_EVIDENCE_CLASS=HISTORICALLY_RECOMPUTED_FROM_PERSISTED_OBSERVATIONS/)
  assert.match(document, /STAGE_22_PERSISTED_FORECAST_HISTORY=NONE/)
  assert.match(document, /STAGE_22_ETA_INTERPOLATION=NONE/)
  assert.match(document, /STAGE_22_CAUSE_INFERENCE=NONE/)
  assert.match(document, /STAGE_22_SAMPLE_CAP=6/)
  assert.match(document, /STAGE_22_ADDITIONAL_COST=0_RUB/)
  assert.match(document, /current projection implementation and current policy against historical evidence/i)
  assert.match(document, /Chromium/i)
})
