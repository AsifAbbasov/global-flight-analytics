// FRONTEND_RESEARCH_SNAPSHOT_EXPORT_V1
import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL(
  '../.test-dist/lib/traffic/traffic-snapshot-export.js',
  import.meta.url
)
const importedModule = await import(moduleURL.href)
const exportModule = importedModule.default ?? importedModule
const { buildTrafficSnapshotExport } = exportModule

function aircraft(overrides = {}) {
  return {
    icao24: 'abc123',
    callsign: 'AZAL101',
    latitude: 40.4,
    longitude: 49.9,
    altitude_m: 10000,
    altitude_status: 'observed',
    altitude_source: 'geometric',
    velocity_mps: 220,
    heading_degrees: 90,
    on_ground: false,
    observed_at: '2026-08-01T12:00:00Z',
    aircraft_model: 'Airbus A320',
    airline: 'Azerbaijan Airlines',
    origin_country: 'Azerbaijan',
    ...overrides,
  }
}

const context = {
  regionCode: 'AZ',
  regionName: 'Azerbaijan',
  snapshotUpdatedAt: Date.parse('2026-08-01T12:34:56.789Z'),
  generatedAt: '2026-08-01T12:35:00.000Z',
  selectedAircraftICAO24: 'ABC123',
}

test('CSV uses a fixed schema and deterministic ICAO24 ordering', () => {
  const result = buildTrafficSnapshotExport(
    [aircraft({ icao24: 'def456' }), aircraft({ icao24: 'ABC123' })],
    context
  )
  const lines = result.csv.content.trimEnd().split('\n')

  assert.equal(lines.length, 3)
  assert.match(lines[0], /^schema_version,region_code,region_name,/)
  const header = lines[0].split(',')
  const icao24Index = header.indexOf('icao24')
  assert.equal(lines[1].split(',')[icao24Index], 'abc123')
  assert.equal(lines[2].split(',')[icao24Index], 'def456')
})

test('CSV escapes commas, quotes and line breaks without losing fields', () => {
  const result = buildTrafficSnapshotExport(
    [
      aircraft({
        callsign: 'A,"B"',
        airline: 'Line one\nLine two',
      }),
    ],
    context
  )

  assert.match(result.csv.content, /"A,""B"""/)
  assert.match(result.csv.content, /"Line one\nLine two"/)
})

test('null and non-finite optional numbers serialize safely', () => {
  const result = buildTrafficSnapshotExport(
    [aircraft({ altitude_m: null, velocity_mps: Number.NaN })],
    context
  )
  const line = result.csv.content.trimEnd().split('\n')[1]

  assert.ok(line.includes(',observed,geometric,,90,false,'))
  const geoJSON = JSON.parse(result.geoJSON.content)
  assert.equal(geoJSON.features[0].properties.altitude_m, null)
  assert.equal(geoJSON.features[0].properties.velocity_mps, null)
})

test('GeoJSON uses longitude-latitude order and snapshot metadata', () => {
  const result = buildTrafficSnapshotExport([aircraft()], context)
  const geoJSON = JSON.parse(result.geoJSON.content)

  assert.deepEqual(geoJSON.features[0].geometry.coordinates, [49.9, 40.4])
  assert.equal(geoJSON.metadata.region_code, 'AZ')
  assert.equal(geoJSON.metadata.selected_aircraft_icao24, 'abc123')
  assert.equal(geoJSON.metadata.aircraft_count, 1)
  assert.match(geoJSON.metadata.evidence_boundary, /Current API snapshot/)
})

test('GeoJSON excludes invalid coordinates and reports exclusions', () => {
  const result = buildTrafficSnapshotExport(
    [
      aircraft({ icao24: 'valid1' }),
      aircraft({ icao24: 'badlat', latitude: 91 }),
      aircraft({ icao24: 'badlon', longitude: -181 }),
    ],
    context
  )
  const geoJSON = JSON.parse(result.geoJSON.content)

  assert.equal(result.totalAircraftCount, 3)
  assert.equal(result.geoJSONFeatureCount, 1)
  assert.equal(result.excludedInvalidCoordinateCount, 2)
  assert.equal(geoJSON.metadata.excluded_invalid_coordinate_count, 2)
})

test('filenames are deterministic and use the snapshot timestamp', () => {
  const result = buildTrafficSnapshotExport([], context)

  assert.equal(
    result.csv.filename,
    'global-flight-analytics-az-20260801T123456Z.csv'
  )
  assert.equal(
    result.geoJSON.filename,
    'global-flight-analytics-az-20260801T123456Z.geojson'
  )
})

test('invalid generated timestamps are rejected instead of guessed', () => {
  assert.throws(
    () =>
      buildTrafficSnapshotExport([], {
        ...context,
        generatedAt: 'not-a-date',
      }),
    /generatedAt/
  )
})
