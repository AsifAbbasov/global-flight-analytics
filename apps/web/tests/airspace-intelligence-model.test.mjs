import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL(
  '../.test-dist/lib/airspace/airspace-intelligence-model.js',
  import.meta.url
)
const importedModule = await import(moduleURL.href)
const model = importedModule.default ?? importedModule
const {
  resolveAirspaceAsOfTime,
  canRequestAirspaceRegionAnalytics,
} = model

const baseAircraft = {
  icao24: '4b1801',
  callsign: 'AZAL101',
  latitude: 40.45,
  longitude: 50,
  altitude_m: 3000,
  altitude_status: 'observed',
  altitude_source: 'geometric',
  velocity_mps: 200,
  heading_degrees: 90,
  on_ground: false,
  aircraft_model: 'A320',
  airline: 'AZAL',
  origin_country: 'Azerbaijan',
}

test('airspace request time uses the latest valid observed traffic timestamp', () => {
  const result = resolveAirspaceAsOfTime([
    {
      ...baseAircraft,
      observed_at: '2026-08-04T17:45:00Z',
    },
    {
      ...baseAircraft,
      icao24: '4b1802',
      observed_at: 'invalid',
    },
    {
      ...baseAircraft,
      icao24: '4b1803',
      observed_at: '2026-08-04T17:47:30+00:00',
    },
  ])

  assert.equal(result, '2026-08-04T17:47:30.000Z')
})

test('airspace request time remains unavailable without valid observations', () => {
  assert.equal(resolveAirspaceAsOfTime([]), null)
  assert.equal(
    resolveAirspaceAsOfTime([
      {
        ...baseAircraft,
        observed_at: 'invalid',
      },
    ]),
    null
  )
})

test('airspace analytics requires a bounded region and observed timestamp', () => {
  assert.equal(
    canRequestAirspaceRegionAnalytics('az', '2026-08-04T17:47:30Z'),
    true
  )
  assert.equal(
    canRequestAirspaceRegionAnalytics(' AZ ', '2026-08-04T17:47:30Z'),
    true
  )
  assert.equal(
    canRequestAirspaceRegionAnalytics('world', '2026-08-04T17:47:30Z'),
    false
  )
  assert.equal(
    canRequestAirspaceRegionAnalytics('az', null),
    false
  )
  assert.equal(canRequestAirspaceRegionAnalytics('   ', null), false)
})
