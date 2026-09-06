import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

function source(relativePath) {
  return readFileSync(new URL(`../../../${relativePath}`, import.meta.url), 'utf8')
}

const document = source(
  'docs/197_STAGE_15_HISTORICAL_FLIGHT_REPLAY_PRODUCT_AND_EVIDENCE_HARDENING.md'
)

test('Stage 15 replay closure preserves canonical evidence and zero-budget markers', () => {
  assert.match(document, /Status: closed/)
  assert.match(document, /STAGE_15_HISTORICAL_FLIGHT_REPLAY=CLOSED/)
  assert.match(
    document,
    /STAGE_15_IMPLEMENTATION_SHA=6ed51f563818a619bb2346e6c4ad50d4256d60ec/
  )
  assert.match(document, /HISTORICAL_FLIGHT_REPLAY_EVIDENCE=OBSERVED_ONLY/)
  assert.match(document, /HISTORICAL_FLIGHT_REPLAY_INTERPOLATION=NONE/)
  assert.match(document, /HISTORICAL_FLIGHT_REPLAY_COST_BOUNDARY=0_RUB/)
  assert.match(document, /STAGE_15_DOCUMENTATION=CLOSED/)
})

test('Stage 15 replay closure records every delivered product increment and exact final head', () => {
  const expected = [
    ['#149', '9483d05dbb1828f1efd19f7d6e782c8b16f92939'],
    ['#150', 'f3db7c5045915b13c17797969d15896f1d206e74'],
    ['#151', '9c1496445d124137eccbc68b36ecc3ac4318f8c7'],
    ['#152', '3f7d06edfb7ff125b007fb6c822e6c9e87ea4dff'],
    ['#153', 'cc783b28410b029309eed5afc819608ab36441ee'],
  ]

  for (const [pr, sha] of expected) {
    assert.ok(document.includes(pr), `missing ${pr}`)
    assert.ok(document.includes(sha), `missing ${sha}`)
  }
})

test('Stage 15 replay closure permanently documents the evidence honesty boundary', () => {
  assert.match(document, /no intermediate position is inferred/i)
  assert.match(document, /endpoint displacement is not travelled path distance/i)
  assert.match(document, /EVIDENCE_CLASS=OBSERVED/)
  assert.match(document, /INTERPOLATION_POLICY=NONE/)
  assert.match(document, /gaps are evidence that the intermediate position is unknown/i)
})

test('Stage 15 replay closure preserves real remediation history instead of fabricated review history', () => {
  assert.match(document, /useEffect/)
  assert.match(document, /Frontend CI\/ESLint rejected/i)
  assert.match(document, /keyed component/i)
  assert.match(document, /Trail OFF must still allow replay data to exist/i)

  assert.match(document, /Playwright #180 failure/i)
  assert.match(document, /getByText\('0\.0 m\/s', \{ exact: true \}\)/)
  assert.match(document, /Playwright #181 run `34024855653` completed successfully/i)

  assert.match(document, /no historical CI\/review rejection occurred/i)
  assert.match(document, /no historical review rejection occurred/i)
})

test('Stage 15 replay closure protects adversarial semantics for analytics, links and movement', () => {
  assert.match(document, /flight-phase and predictive labels are deferred/i)
  assert.match(document, /replay_observation=<state\.id>/)
  assert.match(document, /adding an earlier observation must not cause an existing shared link/i)
  assert.match(document, /350° → 10°.*20°/)
  assert.match(document, /two identical endpoints can still hide movement between observations/i)
})

test('Stage 15 replay closure records exact post-merge evidence and residual limitations', () => {
  assert.match(document, /Frontend CI #411\s+run 34029950587\s+SUCCESS/)
  assert.match(document, /CodeQL #391\s+run 34029950613\s+SUCCESS/)
  assert.match(document, /Playwright E2E #188 run 34029950618\s+SUCCESS/)
  assert.match(document, /only persisted observations can be replayed/i)
  assert.match(document, /browser clipboard behavior remains subject to browser permission/i)
  assert.match(document, /future free-tier\/provider\/retention limits can reduce data availability/i)
})

test('Stage 15 future guard requires evidence, interpolation, provenance and cost declarations', () => {
  assert.match(document, /source data used/)
  assert.match(document, /evidence class/)
  assert.match(document, /interpolation policy/)
  assert.match(document, /identity\/provenance key/)
  assert.match(document, /incremental monetary cost/)
  assert.match(document, /A feature that cannot answer these questions is not ready to merge/)
})
