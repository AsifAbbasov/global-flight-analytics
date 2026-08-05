/* global process, console, URL, Buffer */
import http from 'node:http'
import { fileURLToPath } from 'node:url'

export const openAPIPaths = new Set([
  '/api/v1/health',
  '/api/v1/ready',
  '/api/v1/version',
  '/api/v1/regions',
  '/api/v1/regions/{code}',
  '/api/v1/airports',
  '/api/v1/airports/{icao}',
  '/api/v1/traffic/current',
  '/api/v1/metrics/active-aircraft',
  '/api/v1/aircraft/{icao24}/trajectory',
  '/api/v1/aircraft/{icao24}/route-context',
  '/api/v1/trajectories/{id}',
  '/api/v1/flights/{flightID}/states',
  '/api/v1/aircraft/{icao24}/latest-state',
  '/api/v1/flights',
  '/api/v1/flights/{id}',
  '/api/v1/aircraft',
  '/api/v1/aircraft/{icao24}',
])

export const supportedScenarios = new Set([
  'healthy',
  'traffic-error',
  'regions-error',
])

const regions = [
  {
    code: 'az',
    name: 'Azerbaijan',
    description: 'Azerbaijan regional observation view',
    bounds: {
      min_latitude: 38.3,
      max_latitude: 41.9,
      min_longitude: 44.7,
      max_longitude: 50.9,
    },
  },
  {
    code: 'tr',
    name: 'Turkey',
    description: 'Turkey regional observation view',
    bounds: {
      min_latitude: 35.8,
      max_latitude: 42.1,
      min_longitude: 25.7,
      max_longitude: 44.8,
    },
  },
]

const airports = [
  {
    icao_code: 'UBBB',
    iata_code: 'GYD',
    name: 'Heydar Aliyev International Airport',
    city: 'Baku',
    country: 'Azerbaijan',
    latitude: 40.4675,
    longitude: 50.0467,
    elevation_m: 3,
    elevation_status: 'observed',
    timezone: 'Asia/Baku',
    description: 'Deterministic Playwright fixture airport.',
  },
  {
    icao_code: 'LTFM',
    iata_code: 'IST',
    name: 'Istanbul Airport',
    city: 'Istanbul',
    country: 'Turkey',
    latitude: 41.2753,
    longitude: 28.7519,
    elevation_m: 99,
    elevation_status: 'observed',
    timezone: 'Europe/Istanbul',
    description: 'Deterministic Playwright fixture airport.',
  },
]

const traffic = [
  {
    region_code: 'az',
    icao24: '4b1801',
    callsign: 'AZAL101',
    latitude: 40.4093,
    longitude: 49.8671,
    altitude_m: 10_668,
    altitude_status: 'observed',
    altitude_source: 'barometric',
    velocity_mps: 230,
    heading_degrees: 285,
    on_ground: false,
    observed_at: '2026-08-04T18:00:00Z',
    aircraft_model: 'Airbus A320',
    airline: 'Azerbaijan Airlines',
    origin_country: 'Azerbaijan',
  },
  {
    region_code: 'tr',
    icao24: '4ba901',
    callsign: 'THY202',
    latitude: 41.0082,
    longitude: 28.9784,
    altitude_m: 9_753,
    altitude_status: 'observed',
    altitude_source: 'geometric',
    velocity_mps: 218,
    heading_degrees: 92,
    on_ground: false,
    observed_at: '2026-08-04T18:00:05Z',
    aircraft_model: 'Boeing 737-800',
    airline: 'Turkish Airlines',
    origin_country: 'Turkey',
  },
]

const aircraft = [
  {
    icao24: '4b1801',
    registration: '4K-AZ01',
    model: 'Airbus A320',
    manufacturer: 'Airbus',
    aircraft_type: 'narrow-body',
    airline: 'Azerbaijan Airlines',
    country: 'Azerbaijan',
  },
  {
    icao24: '4ba901',
    registration: 'TC-TK01',
    model: 'Boeing 737-800',
    manufacturer: 'Boeing',
    aircraft_type: 'narrow-body',
    airline: 'Turkish Airlines',
    country: 'Turkey',
  },
]

const flights = [
  {
    id: 'flight-azal-101',
    aircraft_id: 'aircraft-4b1801',
    icao24: '4b1801',
    callsign: 'AZAL101',
    status: 'active',
    first_seen_at: '2026-08-04T17:45:00Z',
    last_seen_at: '2026-08-04T18:00:00Z',
    aircraft_model: 'Airbus A320',
    airline: 'Azerbaijan Airlines',
    country: 'Azerbaijan',
  },
  {
    id: 'flight-thy-202',
    aircraft_id: 'aircraft-4ba901',
    icao24: '4ba901',
    callsign: 'THY202',
    status: 'active',
    first_seen_at: '2026-08-04T17:40:00Z',
    last_seen_at: '2026-08-04T18:00:05Z',
    aircraft_model: 'Boeing 737-800',
    airline: 'Turkish Airlines',
    country: 'Turkey',
  },
]

const flightStates = [
  {
    id: 'state-azal-101-1',
    flight_id: 'flight-azal-101',
    aircraft_id: 'aircraft-4b1801',
    icao24: '4b1801',
    callsign: 'AZAL101',
    latitude: 40.4093,
    longitude: 49.8671,
    barometric_altitude_m: 10_668,
    barometric_altitude_status: 'observed',
    geometric_altitude_m: null,
    geometric_altitude_status: 'unavailable',
    velocity_mps: 230,
    heading_degrees: 285,
    vertical_rate_mps: 0,
    on_ground: false,
    origin_country: 'Azerbaijan',
    observed_at: '2026-08-04T18:00:00Z',
    source_name: 'playwright-fixture',
  },
  {
    id: 'state-thy-202-1',
    flight_id: 'flight-thy-202',
    aircraft_id: 'aircraft-4ba901',
    icao24: '4ba901',
    callsign: 'THY202',
    latitude: 41.0082,
    longitude: 28.9784,
    barometric_altitude_m: null,
    barometric_altitude_status: 'unavailable',
    geometric_altitude_m: 9_753,
    geometric_altitude_status: 'observed',
    velocity_mps: 218,
    heading_degrees: 92,
    vertical_rate_mps: 1.2,
    on_ground: false,
    origin_country: 'Turkey',
    observed_at: '2026-08-04T18:00:05Z',
    source_name: 'playwright-fixture',
  },
]

const trajectories = [
  {
    id: 'trajectory-azal-101',
    identity_key: 'icao24:4b1801',
    identity_basis: 'icao24',
    split_reason: 'none',
    flight_id: 'flight-azal-101',
    aircraft_id: 'aircraft-4b1801',
    icao24: '4b1801',
    callsign: 'AZAL101',
    start_time: '2026-08-04T17:45:00Z',
    end_time: '2026-08-04T18:00:00Z',
    duration_seconds: 900,
    segment_count: 1,
    point_count: 2,
    coverage_gap_count: 0,
    quality_score: 0.96,
    source_name: 'playwright-fixture',
    segments: [
      {
        id: 'segment-azal-101-1',
        trajectory_id: 'trajectory-azal-101',
        flight_id: 'flight-azal-101',
        aircraft_id: 'aircraft-4b1801',
        icao24: '4b1801',
        callsign: 'AZAL101',
        sequence_number: 1,
        status: 'complete',
        quality_score: 0.96,
        start_time: '2026-08-04T17:45:00Z',
        end_time: '2026-08-04T18:00:00Z',
        duration_seconds: 900,
        start_latitude: 40.4675,
        start_longitude: 50.0467,
        end_latitude: 40.4093,
        end_longitude: 49.8671,
        point_count: 2,
        source_name: 'playwright-fixture',
        created_at: '2026-08-04T18:00:10Z',
      },
    ],
    coverage_gaps: [],
    created_at: '2026-08-04T18:00:10Z',
    updated_at: '2026-08-04T18:00:10Z',
  },
]

const routeContexts = [
  {
    icao24: '4b1801',
    trajectory_id: 'trajectory-azal-101',
    origin: {
      airport: airports[0],
      distance_km: 3.4,
      confidence: {
        score: 0.91,
        level: 'high',
        reasons: [
          { code: 'NEAREST_START_AIRPORT', message: 'Nearest start airport.' },
        ],
      },
    },
    destination: undefined,
    confidence: {
      score: 0.64,
      level: 'medium',
      reasons: [
        { code: 'DESTINATION_UNAVAILABLE', message: 'Destination evidence is incomplete.' },
      ],
    },
    limitations: [
      { code: 'OPEN_DATA_ONLY', message: 'No filed flight plan is asserted.' },
    ],
    generated_at: '2026-08-04T18:00:15Z',
  },
]

const activeAircraftMetric = {
  metric: 'active_aircraft',
  value: 2,
  window_minutes: 15,
  scope: { type: 'global', code: '' },
  observed_from: '2026-08-04T17:45:05Z',
  observed_to: '2026-08-04T18:00:05Z',
  calculated_at: '2026-08-04T18:00:06Z',
  confidence: { level: 'high', score: 1, reasons: [] },
  sources: [{ name: 'playwright-fixture', role: 'observation' }],
  limitations: [],
}

function publicTrafficItem(item) {
  return {
    icao24: item.icao24,
    callsign: item.callsign,
    latitude: item.latitude,
    longitude: item.longitude,
    altitude_m: item.altitude_m,
    altitude_status: item.altitude_status,
    altitude_source: item.altitude_source,
    velocity_mps: item.velocity_mps,
    heading_degrees: item.heading_degrees,
    on_ground: item.on_ground,
    observed_at: item.observed_at,
    aircraft_model: item.aircraft_model,
    airline: item.airline,
    origin_country: item.origin_country,
  }
}

function success(data) {
  return {
    status: 200,
    body: {
      success: true,
      data,
    },
  }
}

function failure(status, code, message) {
  return {
    status,
    body: {
      success: false,
      error: {
        code,
        message,
      },
    },
  }
}

function normalizePath(pathname) {
  if (/^\/api\/v1\/regions\/[^/]+$/.test(pathname)) {
    return '/api/v1/regions/{code}'
  }
  if (/^\/api\/v1\/airports\/[^/]+$/.test(pathname)) {
    return '/api/v1/airports/{icao}'
  }
  if (/^\/api\/v1\/aircraft\/[^/]+\/trajectory$/.test(pathname)) {
    return '/api/v1/aircraft/{icao24}/trajectory'
  }
  if (/^\/api\/v1\/aircraft\/[^/]+\/route-context$/.test(pathname)) {
    return '/api/v1/aircraft/{icao24}/route-context'
  }
  if (/^\/api\/v1\/aircraft\/[^/]+\/latest-state$/.test(pathname)) {
    return '/api/v1/aircraft/{icao24}/latest-state'
  }
  if (/^\/api\/v1\/aircraft\/[^/]+$/.test(pathname)) {
    return '/api/v1/aircraft/{icao24}'
  }
  if (/^\/api\/v1\/trajectories\/[^/]+$/.test(pathname)) {
    return '/api/v1/trajectories/{id}'
  }
  if (/^\/api\/v1\/flights\/[^/]+\/states$/.test(pathname)) {
    return '/api/v1/flights/{flightID}/states'
  }
  if (/^\/api\/v1\/flights\/[^/]+$/.test(pathname)) {
    return '/api/v1/flights/{id}'
  }
  return pathname
}

export function resolveMockResponse({
  method,
  requestURL,
  scenario = 'healthy',
}) {
  const url = new URL(requestURL, 'http://127.0.0.1')
  const pathname = url.pathname

  if (method === 'GET' && pathname === '/api/v1/health') {
    return success({ status: 'ok' })
  }
  if (method === 'GET' && pathname === '/api/v1/ready') {
    return success({ status: 'ready' })
  }
  if (method === 'GET' && pathname === '/api/v1/version') {
    return success({
      version: 'e2e',
      revision: 'playwright-fixture',
      built_at: '2026-08-04T18:00:00Z',
    })
  }

  if (
    scenario === 'regions-error' &&
    method === 'GET' &&
    (pathname === '/api/v1/regions' ||
      normalizePath(pathname) === '/api/v1/regions/{code}')
  ) {
    return failure(
      503,
      'REGION_FIXTURE_UNAVAILABLE',
      'The deterministic region fixture is unavailable.',
    )
  }

  if (method === 'GET' && pathname === '/api/v1/regions') {
    return success(regions)
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) === '/api/v1/regions/{code}'
  ) {
    const code = decodeURIComponent(pathname.split('/').at(-1) ?? '').toLowerCase()
    const region = regions.find(item => item.code === code)
    return region
      ? success(region)
      : failure(404, 'REGION_NOT_FOUND', 'Region not found')
  }

  if (method === 'GET' && pathname === '/api/v1/airports') {
    return success(
      airports.map(item => ({
        icao_code: item.icao_code,
        iata_code: item.iata_code,
        name: item.name,
        city: item.city,
        country: item.country,
        latitude: item.latitude,
        longitude: item.longitude,
      })),
    )
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) === '/api/v1/airports/{icao}'
  ) {
    const icao = decodeURIComponent(pathname.split('/').at(-1) ?? '').toUpperCase()
    const airport = airports.find(item => item.icao_code === icao)
    return airport
      ? success(airport)
      : failure(404, 'AIRPORT_NOT_FOUND', 'Airport not found')
  }

  if (method === 'GET' && pathname === '/api/v1/traffic/current') {
    if (scenario === 'traffic-error') {
      return failure(
        503,
        'TRAFFIC_FIXTURE_UNAVAILABLE',
        'The deterministic traffic fixture is unavailable.',
      )
    }
    const requestedRegion = url.searchParams.get('region')?.trim().toLowerCase()
    const selectedTraffic =
      !requestedRegion || requestedRegion === 'world'
        ? traffic
        : traffic.filter(item => item.region_code === requestedRegion)
    return success(selectedTraffic.map(publicTrafficItem))
  }

  if (method === 'GET' && pathname === '/api/v1/metrics/active-aircraft') {
    const region = url.searchParams.get('region')?.trim().toLowerCase()
    return success({
      ...activeAircraftMetric,
      scope:
        region && region !== 'world'
          ? { type: 'region', code: region }
          : activeAircraftMetric.scope,
    })
  }

  if (method === 'GET' && pathname === '/api/v1/aircraft') {
    return success(
      aircraft.map(item => ({
        icao24: item.icao24,
        registration: item.registration,
        model: item.model,
        manufacturer: item.manufacturer,
        airline: item.airline,
        country: item.country,
      })),
    )
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) === '/api/v1/aircraft/{icao24}'
  ) {
    const icao24 = decodeURIComponent(pathname.split('/').at(-1) ?? '').toLowerCase()
    const item = aircraft.find(candidate => candidate.icao24 === icao24)
    return item
      ? success(item)
      : failure(404, 'AIRCRAFT_NOT_FOUND', 'Aircraft not found')
  }

  if (method === 'GET' && pathname === '/api/v1/flights') {
    return success(flights)
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) === '/api/v1/flights/{id}'
  ) {
    const id = decodeURIComponent(pathname.split('/').at(-1) ?? '')
    const item = flights.find(candidate => candidate.id === id)
    return item ? success(item) : failure(404, 'FLIGHT_NOT_FOUND', 'Flight not found')
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) === '/api/v1/flights/{flightID}/states'
  ) {
    const flightID = decodeURIComponent(pathname.split('/').at(-2) ?? '')
    return success(flightStates.filter(item => item.flight_id === flightID))
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) === '/api/v1/aircraft/{icao24}/latest-state'
  ) {
    const icao24 = decodeURIComponent(pathname.split('/').at(-2) ?? '').toLowerCase()
    const item = flightStates.find(candidate => candidate.icao24 === icao24)
    return item
      ? success(item)
      : failure(404, 'FLIGHT_STATE_NOT_FOUND', 'Flight state not found')
  }

  if (
    method === 'GET' &&
    normalizePath(pathname) === '/api/v1/aircraft/{icao24}/trajectory'
  ) {
    const icao24 = decodeURIComponent(pathname.split('/').at(-2) ?? '').toLowerCase()
    const item = trajectories.find(candidate => candidate.icao24 === icao24)
    return item
      ? success(item)
      : failure(404, 'TRAJECTORY_NOT_FOUND', 'Trajectory not found')
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) === '/api/v1/trajectories/{id}'
  ) {
    const id = decodeURIComponent(pathname.split('/').at(-1) ?? '')
    const item = trajectories.find(candidate => candidate.id === id)
    return item
      ? success(item)
      : failure(404, 'TRAJECTORY_NOT_FOUND', 'Trajectory not found')
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) === '/api/v1/aircraft/{icao24}/route-context'
  ) {
    const icao24 = decodeURIComponent(pathname.split('/').at(-2) ?? '').toLowerCase()
    const item = routeContexts.find(candidate => candidate.icao24 === icao24)
    return item
      ? success(item)
      : failure(404, 'ROUTE_CONTEXT_NOT_FOUND', 'Route context not found')
  }

  return failure(404, 'E2E_ROUTE_NOT_FOUND', 'Mock API route not found')
}

function writeJSON(response, resolved) {
  const body = JSON.stringify(resolved.body)
  response.writeHead(resolved.status, {
    'Access-Control-Allow-Headers': 'Accept,Content-Type',
    'Access-Control-Allow-Methods': 'GET,POST,OPTIONS',
    'Access-Control-Allow-Origin': 'http://127.0.0.1:3000',
    'Cache-Control': 'no-store',
    'Content-Length': Buffer.byteLength(body),
    'Content-Type': 'application/json; charset=utf-8',
    'X-Request-ID': 'playwright-e2e-request',
  })
  response.end(body)
}

async function readJSON(request) {
  const chunks = []
  let size = 0
  for await (const chunk of request) {
    size += chunk.length
    if (size > 16_384) {
      throw new Error('request body is too large')
    }
    chunks.push(chunk)
  }
  if (chunks.length === 0) return {}
  return JSON.parse(Buffer.concat(chunks).toString('utf8'))
}

export function createMockAPIServer({
  host = '127.0.0.1',
  port = 8091,
} = {}) {
  let scenario = 'healthy'

  const server = http.createServer(async (request, response) => {
    try {
      const requestURL = new URL(
        request.url ?? '/',
        `http://${host}:${port}`,
      )

      if (request.method === 'OPTIONS') {
        response.writeHead(204, {
          'Access-Control-Allow-Headers': 'Accept,Content-Type',
          'Access-Control-Allow-Methods': 'GET,POST,OPTIONS',
          'Access-Control-Allow-Origin': 'http://127.0.0.1:3000',
          'Cache-Control': 'no-store',
        })
        response.end()
        return
      }

      if (
        request.method === 'POST' &&
        requestURL.pathname === '/__e2e/scenario'
      ) {
        const payload = await readJSON(request)
        if (!supportedScenarios.has(payload.scenario)) {
          writeJSON(
            response,
            failure(
              400,
              'INVALID_E2E_SCENARIO',
              'The requested Playwright scenario is invalid.',
            ),
          )
          return
        }
        scenario = payload.scenario
        writeJSON(response, success({ scenario }))
        return
      }

      writeJSON(
        response,
        resolveMockResponse({
          method: request.method ?? 'GET',
          requestURL: requestURL.toString(),
          scenario,
        }),
      )
    } catch (error) {
      writeJSON(
        response,
        failure(
          500,
          'E2E_MOCK_FAILURE',
          error instanceof Error ? error.message : 'Mock API failure',
        ),
      )
    }
  })

  return {
    server,
    getScenario: () => scenario,
    listen: () =>
      new Promise((resolve, reject) => {
        server.once('error', reject)
        server.listen(port, host, () => {
          server.off('error', reject)
          resolve(server.address())
        })
      }),
    close: () =>
      new Promise((resolve, reject) => {
        server.close(error => {
          if (error) reject(error)
          else resolve()
        })
      }),
  }
}

const currentFile = fileURLToPath(import.meta.url)
if (process.argv[1] && currentFile === process.argv[1]) {
  const host = process.env.PLAYWRIGHT_API_HOST ?? '127.0.0.1'
  const port = Number(process.env.PLAYWRIGHT_API_PORT ?? '8091')
  const mock = createMockAPIServer({ host, port })
  await mock.listen()
  console.log(`PLAYWRIGHT_MOCK_API_READY=http://${host}:${port}`)
}
