import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL(
  '../.test-dist/lib/replay/flight-replay-evidence-provenance-model.js',
  import.meta.url
)
const importedModule = await import(moduleURL.href)
const provenanceModel = importedModule.default ?? importedModule
const { buildFlightReplayEvidenceProvenanceProfile } = provenanceModel

const basePoint = {
  flight_id: 'flight-1',
  icao24: '4b1801',
  callsign: 'AZAL101',
  latitude: 40.45,
  longitude: 50,
  barometric_altitude_m: 3000,
  barometric_altitude_status: 'observed',
  geometric_altitude_m: null,
  geometric_altitude_status: 'unavailable',
  velocity_mps: 200,
  heading_degrees: 90,
  vertical_rate_mps: 0,
  on_ground: false,
  origin_country: 'Azerbaijan',
}

const replay = {
  trajectory_id: 'trajectory-1',
  flight_id: 'flight-1',
  icao24: '4b1801',
  callsign: 'AZAL101',
  start_time: '2026-08-04T17:45:00Z',
  end_time: '2026-08-04T17:49:00Z',
  evidence_class: 'observed',
  interpolation_policy: 'none',
  points: [
    {
      ...basePoint,
      id: 'state-1',
      source_name: 'adsb.lol',
      observed_at: '2026-08-04T17:45:00Z',
    },
    {
      ...basePoint,
      id: 'state-2',
      source_name: 'adsb.lol',
      observed_at: '2026-08-04T17:46:00Z',
    },
    {
      ...basePoint,
      id: 'state-3',
      source_name: 'opensky',
      observed_at: '2026-08-04T17:47:00Z',
    },
    {
      ...basePoint,
      id: 'state-4',
      source_name: 'opensky',
      observed_at: '2026-08-04T17:48:00Z',
    },
    {
      ...basePoint,
      id: 'state-5',
      source_name: 'adsb.lol',
      observed_at: '2026-08-04T17:49:00Z',
    },
  ],
}

test('provenance profile describes persisted source composition and adjacent transitions', () => {
  const profile = buildFlightReplayEvidenceProvenanceProfile(replay)

  assert.equal(profile.sampleCount, 5)
  assert.equal(profile.identifiedSourceCount, 2)
  assert.equal(profile.unattributedSampleCount, 0)
  assert.deepEqual(profile.sources, [
    {
      sourceName: 'adsb.lol',
      sampleCount: 3,
      sampleSharePercent: 60,
      firstObservedAt: '2026-08-04T17:45:00Z',
      lastObservedAt: '2026-08-04T17:49:00Z',
    },
    {
      sourceName: 'opensky',
      sampleCount: 2,
      sampleSharePercent: 40,
      firstObservedAt: '2026-08-04T17:47:00Z',
      lastObservedAt: '2026-08-04T17:48:00Z',
    },
  ])
  assert.deepEqual(profile.transitions, [
    {
      fromIndex: 1,
      toIndex: 2,
      fromSourceName: 'adsb.lol',
      toSourceName: 'opensky',
      startObservedAt: '2026-08-04T17:46:00Z',
      endObservedAt: '2026-08-04T17:47:00Z',
      elapsedSeconds: 60,
    },
    {
      fromIndex: 3,
      toIndex: 4,
      fromSourceName: 'opensky',
      toSourceName: 'adsb.lol',
      startObservedAt: '2026-08-04T17:48:00Z',
      endObservedAt: '2026-08-04T17:49:00Z',
      elapsedSeconds: 60,
    },
  ])
  assert.equal(Object.hasOwn(profile, 'score'), false)
  assert.equal(Object.hasOwn(profile, 'providerRank'), false)
  assert.equal(Object.hasOwn(profile, 'accuracy'), false)
})

test('provenance profile does not invent a transition across unattributed evidence', () => {
  const profile = buildFlightReplayEvidenceProvenanceProfile({
    ...replay,
    points: [
      replay.points[0],
      {
        ...replay.points[1],
        source_name: '   ',
      },
      replay.points[2],
    ],
  })

  assert.equal(profile.sampleCount, 3)
  assert.equal(profile.identifiedSourceCount, 2)
  assert.equal(profile.unattributedSampleCount, 1)
  assert.equal(profile.transitions.length, 0)
  assert.equal(profile.sources[0]?.sampleSharePercent, 33)
  assert.equal(profile.sources[1]?.sampleSharePercent, 33)
})

test('provenance profile keeps invalid transition duration unavailable', () => {
  const profile = buildFlightReplayEvidenceProvenanceProfile({
    ...replay,
    points: [
      replay.points[0],
      {
        ...replay.points[2],
        observed_at: 'invalid-timestamp',
      },
    ],
  })

  assert.equal(profile.transitions.length, 1)
  assert.equal(profile.transitions[0]?.elapsedSeconds, null)
})

test('provenance profile handles missing replay without synthetic defaults', () => {
  assert.deepEqual(buildFlightReplayEvidenceProvenanceProfile(undefined), {
    sampleCount: 0,
    identifiedSourceCount: 0,
    unattributedSampleCount: 0,
    sources: [],
    transitions: [],
  })
})
