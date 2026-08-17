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
  '/api/v1/traffic/live',
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
  '/api/v1/aircraft/{icao24}/transponder-evidence/latest',
  '/api/v1/weather/current',
  '/api/v1/analytics/metrics/active-aircraft',
  '/api/v1/analytics/metrics/traffic-density',
  '/api/v1/analytics/metrics/airport-activity',
  '/api/v1/analytics/metrics/coverage-score',
  '/api/v1/analytics/metrics/data-freshness',
  '/api/v1/airports/intelligence/ranking',
  '/api/v1/airports/{icao}/intelligence/overview',
  '/api/v1/airports/{icao}/intelligence/history',
  '/api/v1/airports/{icao}/intelligence/trends',
  '/api/v1/historical-intelligence/aggregates/latest',
  '/api/v1/historical-intelligence/aggregates/history',
  '/api/v1/trajectories/{id}/projection-intelligence',
  '/api/v1/trajectories/{id}/stability-intelligence',
  '/api/v1/trajectories/{id}/weather-context',
  '/api/v1/airspace/regions/{code}/analytics',
  '/api/v1/trajectories/{id}/route-intelligence',
  '/api/v1/trajectories/{id}/route-intelligence/latest',
  '/api/v1/trajectories/{id}/route-intelligence/history',
])

export const supportedScenarios = new Set([
  'healthy',
  'traffic-error',
  'regions-error',
  'aircraft-error',
  'airport-error',
  'historical-error',
  'intelligence-error',
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
    id: '11111111-1111-4111-8111-111111111111',
    identity_key: 'icao24:4b1801',
    identity_basis: 'aircraft_and_start_time',
    split_reason: 'initial_observation',
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
        trajectory_id: '11111111-1111-4111-8111-111111111111',
        flight_id: 'flight-azal-101',
        aircraft_id: 'aircraft-4b1801',
        icao24: '4b1801',
        callsign: 'AZAL101',
        sequence_number: 1,
        status: 'observed',
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
    trajectory_id: '11111111-1111-4111-8111-111111111111',
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

const advancedTrajectoryID = '11111111-1111-4111-8111-111111111111'
const advancedFingerprint = `sha256:${'a'.repeat(64)}`
const advancedOutputFingerprint = `sha256:${'b'.repeat(64)}`
const advancedDecisionFingerprint = `sha256:${'c'.repeat(64)}`

const transponderEvidence = {
  schema_version: 'transponder-evidence-v1',
  evidence_only: true,
  confirmed_emergency: false,
  operational_alert: false,
  aircraft: { icao24: '4b1801', callsign: 'AZAL101' },
  observed_transponder_code: '7700',
  classification: { kind: 'special-purpose', label: 'Observed special transponder code' },
  observation: {
    strength: 'single-observation',
    observation_count: 1,
    special_purpose_indicator_observed: true,
  },
  freshness: {
    status: 'fresh',
    first_observed_at: '2026-08-04T18:00:00Z',
    last_observed_at: '2026-08-04T18:00:00Z',
    as_of_time: '2026-08-04T18:00:05Z',
    age_seconds: 5,
    maximum_fresh_age_seconds: 300,
  },
  confidence: {
    level: 'low',
    reasons: ['One persisted observation is evidence, not confirmation.'],
  },
  provenance: {
    fingerprint: advancedFingerprint,
    source_names: ['playwright-fixture'],
  },
  maximum_claim_strength: 'observed_evidence_only',
  limitations: ['No confirmed emergency or operational alert is asserted.'],
}

const currentWeather = {
  snapshot_id: 'weather-fixture-1',
  provider: 'open-meteo-fixture',
  latitude: 40.4093,
  longitude: 49.8671,
  observed_at: '2026-08-04T18:00:00Z',
  retrieved_at: '2026-08-04T18:00:02Z',
  stored_at: '2026-08-04T18:00:03Z',
  temperature_celsius: 29.4,
  relative_humidity_percent: 54,
  precipitation_mm: 0,
  rain_mm: null,
  weather_code: 1,
  cloud_cover_percent: 18,
  surface_pressure_hpa: 1008.3,
  wind_speed_mps: 5.8,
  wind_direction_degrees: 175,
  wind_gusts_mps: null,
}

function analyticalMetricFixture(metric, value) {
  return {
    metric,
    status: 'available',
    value,
    has_value: true,
    confidence: {
      level: 'high',
      score: 0.91,
      reasons: [{ code: 'fixture_evidence', message: 'Deterministic fixture evidence.' }],
    },
    scope: {
      capability: metric,
      input_count: 2,
      allowed_count: 2,
      denied_count: 0,
      reasons: [],
      evaluated_at: '2026-08-04T18:00:05Z',
    },
    sources: [
      {
        name: 'playwright-fixture',
        role: 'observation',
        observed_from: '2026-08-04T17:45:00Z',
        observed_to: '2026-08-04T18:00:00Z',
        retrieved_at: '2026-08-04T18:00:05Z',
        limitations: [],
      },
    ],
    warnings: [],
    limitations: [
      {
        code: 'not_operational_air_traffic_control',
        message: 'This fixture is not operational aviation guidance.',
      },
    ],
    calculated_at: '2026-08-04T18:00:06Z',
  }
}

const airportIntelligenceWindow = {
  start_time: '2026-07-06T00:00:00Z',
  end_time: '2026-08-05T00:00:00Z',
  as_of_time: '2026-08-05T12:00:00Z',
  completed_days: 30,
}

const airportStatistics = {
  icao_code: 'UBBB',
  window_start: airportIntelligenceWindow.start_time,
  window_end: airportIntelligenceWindow.end_time,
  arrivals: 21,
  departures: 19,
  total_movements: 40,
  arrival_share: 0.525,
  departure_share: 0.475,
  movements_per_hour: 0.0556,
  active_aircraft: 16,
  active_routes: 9,
  observed_samples: 120,
  expected_samples: 144,
  coverage_score: 0.8333,
  freshness_score: 0.97,
  latest_observation_at: '2026-08-04T23:58:00Z',
  generated_at: '2026-08-05T12:00:01Z',
}

const airportRanking = {
  version: 'airport-intelligence-production-v1',
  window: airportIntelligenceWindow,
  weights: {
    movements: 0.25,
    routes: 0.2,
    observations: 0.15,
    intensity: 0.15,
    coverage: 0.15,
    freshness: 0.1,
  },
  airports: [
    {
      position: 1,
      icao_code: 'UBBB',
      iata_code: 'GYD',
      name: 'Heydar Aliyev International Airport',
      city: 'Baku',
      country: 'Azerbaijan',
      activity_score: 0.88,
      data_confidence: 0.9,
      movements_component: 0.9,
      routes_component: 0.82,
      observations_component: 0.87,
      intensity_component: 0.79,
      coverage_score: 0.8333,
      freshness_score: 0.97,
      total_movements: 40,
      active_routes: 9,
      observed_samples: 120,
      expected_samples: 144,
      movements_per_hour: 0.0556,
      active_aircraft: 16,
    },
  ],
  limitations: [
    {
      code: 'completed_days_only',
      message: 'Only completed UTC days are included.',
    },
  ],
  generated_at: '2026-08-05T12:00:01Z',
}

const airportOverview = {
  version: 'airport-intelligence-production-v1',
  window: airportIntelligenceWindow,
  passport: {
    identity: {
      icao_code: 'UBBB',
      iata_code: 'GYD',
      name: 'Heydar Aliyev International Airport',
    },
    location: {
      city: 'Baku',
      country: 'Azerbaijan',
      latitude: 40.4675,
      longitude: 50.0467,
      elevation_m: 3,
      elevation_status: 'observed',
      timezone: 'Asia/Baku',
    },
    operations: {
      arrivals: 21,
      departures: 19,
      activity: 40,
      active_aircraft: 16,
    },
    data_quality: {
      freshness_score: 0.97,
      coverage_score: 0.8333,
      observed_at: '2026-08-04T23:58:00Z',
    },
    description: 'Deterministic airport intelligence fixture.',
    generated_at: '2026-08-05T12:00:01Z',
  },
  statistics: airportStatistics,
  ranking: {
    position: 1,
    total_airports: 2,
    activity_score: 0.88,
    data_confidence: 0.9,
    movements_component: 0.9,
    routes_component: 0.82,
    observations_component: 0.87,
    intensity_component: 0.79,
  },
  limitations: airportRanking.limitations,
  generated_at: '2026-08-05T12:00:01Z',
}

const airportHistory = {
  version: 'airport-intelligence-production-v1',
  window: airportIntelligenceWindow,
  icao_code: 'UBBB',
  entries: [airportStatistics],
  limitations: airportRanking.limitations,
  generated_at: '2026-08-05T12:00:01Z',
}

const airportTrendPoint = {
  window_start: '2026-08-04T00:00:00Z',
  window_end: '2026-08-05T00:00:00Z',
  total_movements: 40,
  movements_per_hour: 1.67,
  active_routes: 9,
  coverage_score: 0.8333,
  freshness_score: 0.97,
}

const airportTrends = {
  version: 'airport-intelligence-production-v1',
  window: airportIntelligenceWindow,
  icao_code: 'UBBB',
  compared_windows: 2,
  window_duration_seconds: 86_400,
  direction: 'up',
  baseline: { ...airportTrendPoint, total_movements: 36 },
  current: airportTrendPoint,
  peak: airportTrendPoint,
  total_movements_change: 4,
  movements_per_hour_change: 0.17,
  movements_per_hour_change_percent: 11.11,
  movements_per_hour_change_percent_known: true,
  active_routes_change: 1,
  coverage_score_change: 0.03,
  freshness_score_change: 0.01,
  gap_count: 0,
  gap_duration_seconds: 0,
  observed_duration_seconds: 172_800,
  continuity_score: 1,
  limitations: airportRanking.limitations,
  generated_at: '2026-08-05T12:00:01Z',
}

const historicalResult = {
  schema_version: 'historical-intelligence-v1',
  status: 'complete',
  metric: {
    name: 'active_aircraft',
    unit: 'aircraft',
    aggregation: 'count',
  },
  scope: { type: 'global' },
  window: {
    start_time: '2026-08-04T00:00:00Z',
    end_time: '2026-08-05T00:00:00Z',
    as_of_time: '2026-08-05T12:00:00Z',
  },
  granularity: 'hour',
  points: [
    {
      start_time: '2026-08-04T17:00:00Z',
      end_time: '2026-08-04T18:00:00Z',
      status: 'complete',
      value: 2,
      sample_count: 2,
      coverage_ratio: 1,
      confidence: {
        score: 0.95,
        level: 'high',
        sample_count: 2,
        reasons: [
          {
            code: 'fixture_complete',
            message: 'Complete deterministic bucket.',
            contribution: 0.95,
          },
        ],
      },
      limitations: [],
    },
  ],
  summary: {
    point_count: 1,
    total: 2,
    minimum: 2,
    maximum: 2,
    average: 2,
    median: 2,
  },
  confidence: {
    score: 0.95,
    level: 'high',
    sample_count: 2,
    reasons: [],
  },
  limitations: [],
  provenance: {
    builder_version: 'fixture-v1',
    input_fingerprint: advancedFingerprint,
    source_names: ['playwright-fixture'],
    latest_source_updated_at: '2026-08-04T18:00:00Z',
  },
  generated_at: '2026-08-05T12:00:01Z',
}

const historicalRecord = {
  id: 'historical-fixture-1',
  input_fingerprint: advancedFingerprint,
  stored_at: '2026-08-05T12:00:02Z',
  result: historicalResult,
}

function projectionFixture(asOfTime = '2026-08-04T18:00:00Z') {
  return {
    version: 'projection-read-v1',
    strategy: 'continuation',
    fallback_reason: '',
    arrival_status: 'unavailable',
    projection: {
      schema_version: 'projection-v1',
      status: 'available',
      trajectory_id: advancedTrajectoryID,
      flight_id: 'flight-azal-101',
      aircraft_id: 'aircraft-4b1801',
      icao24: '4b1801',
      callsign: 'AZAL101',
      method: {
        name: 'continuation',
        version: 'v1',
        decision_class: 'bounded_research_projection',
      },
      horizon: {
        as_of_time: asOfTime,
        end_time: '2026-08-04T18:10:00Z',
        step_seconds: 60,
        duration_seconds: 600,
      },
      points: [
        {
          sequence: 1,
          forecast_time: '2026-08-04T18:01:00Z',
          position: {
            latitude: 40.42,
            longitude: 49.91,
            altitude_m: 10_600,
          },
          uncertainty: {
            horizontal_radius_m: 850,
            vertical_radius_m: 120,
          },
          confidence: {
            score: 0.72,
            level: 'medium',
            reasons: [
              {
                code: 'bounded_horizon',
                message: 'Short deterministic horizon.',
                contribution: 0.72,
              },
            ],
          },
        },
      ],
      confidence: {
        score: 0.72,
        level: 'medium',
        reasons: [],
      },
      limitations: [
        {
          code: 'not_navigation_guidance',
          message: 'Projection is research-only.',
          scope: 'projection',
        },
      ],
      explanations: [
        {
          code: 'continuation_selected',
          message: 'Continuation strategy selected.',
        },
      ],
      scope_guard: 'research_only',
      provenance: {
        input_fingerprint: advancedFingerprint,
        inputs: [
          {
            name: 'trajectory',
            classification: 'observation',
            source_name: 'playwright-fixture',
            observed_at: '2026-08-04T18:00:00Z',
            retrieved_at: '2026-08-04T18:00:01Z',
          },
        ],
        latest_input_observed_at: '2026-08-04T18:00:00Z',
      },
      generated_at: '2026-08-04T18:00:02Z',
    },
    evidence: {
      freshness: { status: 'fresh' },
    },
    notices: [],
    input_fingerprint: advancedFingerprint,
    generated_at: '2026-08-04T18:00:02Z',
  }
}

const stabilityIntelligence = {
  version: 'stability-intelligence-v1',
  trajectory_id: advancedTrajectoryID,
  as_of_times: ['2026-08-04T17:55:00Z', '2026-08-04T18:00:00Z'],
  projections: [
    projectionFixture('2026-08-04T17:55:00Z'),
    projectionFixture('2026-08-04T18:00:00Z'),
  ],
  forecast_versions: [
    {
      version_id: 'forecast-v1',
      ordinal: 1,
      method_name: 'continuation',
      method_version: 'v1',
      policy_version: 'policy-v1',
      implementation_version: 'implementation-v1',
      input_fingerprint: advancedFingerprint,
      output_fingerprint: advancedOutputFingerprint,
      decision_fingerprint: advancedDecisionFingerprint,
      created_at: '2026-08-04T17:55:02Z',
    },
    {
      version_id: 'forecast-v2',
      ordinal: 2,
      parent_version_id: 'forecast-v1',
      method_name: 'continuation',
      method_version: 'v1',
      policy_version: 'policy-v1',
      implementation_version: 'implementation-v1',
      input_fingerprint: advancedFingerprint,
      output_fingerprint: advancedOutputFingerprint,
      decision_fingerprint: advancedDecisionFingerprint,
      created_at: '2026-08-04T18:00:02Z',
    },
  ],
  transitions: [
    {
      baseline_version_id: 'forecast-v1',
      candidate_version_id: 'forecast-v2',
      level: 'stable',
      score: 0.91,
      metrics: {
        aligned_point_count: 1,
        aligned_point_share: 1,
        mean_horizontal_shift_kilometers: 0.2,
        maximum_horizontal_shift_kilometers: 0.2,
        aggregate_confidence_delta: 0,
        mean_relative_horizontal_uncertainty_change: 0,
        arrival_comparable: false,
        arrival_shift_seconds: 0,
        method_changed: false,
        policy_changed: false,
        implementation_changed: false,
        input_changed: false,
        output_changed: false,
      },
      input_fingerprint: advancedFingerprint,
      evaluated_at: '2026-08-04T18:00:03Z',
    },
  ],
  forecast_analysis: {
    status: 'available',
    trend: 'stable',
    health: 'healthy',
    metrics: {
      version_count: 2,
      transition_count: 1,
      comparable_transition_count: 1,
      stable_transition_share: 1,
      comparable_transition_share: 1,
      material_change_share: 0,
      mean_stability_score: 0.91,
      minimum_stability_score: 0.91,
      score_standard_deviation: 0,
      longest_stable_run: 1,
      method_change_count: 0,
      policy_change_count: 0,
      implementation_change_count: 0,
      input_change_count: 0,
      output_change_count: 0,
      mean_horizontal_shift_kilometers: 0.2,
      maximum_horizontal_shift_kilometers: 0.2,
      latest_level: 'stable',
    },
    confidence_score: 0.9,
    confidence_level: 'high',
    input_fingerprint: advancedFingerprint,
  },
  propagated_confidence: {
    status: 'available',
    score: 0.72,
    level: 'medium',
    target_node_id: 'forecast-v2',
    input_fingerprint: advancedFingerprint,
  },
  failure_explanation: {
    status: 'none',
    primary_code: 'none',
    blocking_count: 0,
    warning_count: 0,
    unknown_cause_count: 0,
    confidence_score: 1,
    confidence_level: 'high',
    failures: [],
    input_fingerprint: advancedFingerprint,
  },
  unknown_intervention: {
    status: 'blocked',
    claim_kind: 'unknown_intervention',
    decision: 'insufficient_evidence',
    confidence_score: 0,
    evidence_count: 0,
    unknown_evidence_count: 0,
    estimated_evidence_count: 0,
    evidence_completeness: 0,
    input_fingerprint: advancedFingerprint,
  },
  scope_enforcement: {
    status: 'allowed',
    decision: 'research_only',
    claim_count: 1,
    allowed_count: 1,
    limited_count: 0,
    blocked_count: 0,
    violations: [],
    input_fingerprint: advancedFingerprint,
  },
  scope_guards: ['not_navigation_guidance'],
  input_fingerprint: advancedFingerprint,
  generated_at: '2026-08-04T18:00:04Z',
}

function stabilityIntelligenceFixture(asOfTimes) {
  const normalizedAsOfTimes = asOfTimes
    .map(value => String(value).trim())
    .filter(Boolean)

  const effectiveAsOfTimes =
    normalizedAsOfTimes.length >= 2
      ? normalizedAsOfTimes
      : stabilityIntelligence.as_of_times

  const projections = effectiveAsOfTimes.map(asOfTime =>
    projectionFixture(asOfTime)
  )
  const forecastVersions = effectiveAsOfTimes.map((asOfTime, index) => ({
    version_id: `forecast-v${index + 1}`,
    ordinal: index + 1,
    ...(index > 0 ? { parent_version_id: `forecast-v${index}` } : {}),
    method_name: 'continuation',
    method_version: 'v1',
    policy_version: 'policy-v1',
    implementation_version: 'implementation-v1',
    input_fingerprint: advancedFingerprint,
    output_fingerprint: advancedOutputFingerprint,
    decision_fingerprint: advancedDecisionFingerprint,
    created_at: new Date(Date.parse(asOfTime) + 2_000).toISOString(),
  }))
  const transitions = forecastVersions.slice(1).map((candidate, index) => ({
    baseline_version_id: forecastVersions[index].version_id,
    candidate_version_id: candidate.version_id,
    level: 'stable',
    score: 0.91,
    metrics: {
      aligned_point_count: 1,
      aligned_point_share: 1,
      mean_horizontal_shift_kilometers: 0.2,
      maximum_horizontal_shift_kilometers: 0.2,
      aggregate_confidence_delta: 0,
      mean_relative_horizontal_uncertainty_change: 0,
      arrival_comparable: false,
      arrival_shift_seconds: 0,
      method_changed: false,
      policy_changed: false,
      implementation_changed: false,
      input_changed: false,
      output_changed: false,
    },
    input_fingerprint: advancedFingerprint,
    evaluated_at: new Date(
      Date.parse(effectiveAsOfTimes[index + 1]) + 3_000,
    ).toISOString(),
  }))

  return {
    ...stabilityIntelligence,
    as_of_times: effectiveAsOfTimes,
    projections,
    forecast_versions: forecastVersions,
    transitions,
    forecast_analysis: {
      ...stabilityIntelligence.forecast_analysis,
      metrics: {
        ...stabilityIntelligence.forecast_analysis.metrics,
        version_count: forecastVersions.length,
        transition_count: transitions.length,
        comparable_transition_count: transitions.length,
        longest_stable_run: transitions.length,
        latest_level: transitions.at(-1)?.level ?? 'stable',
      },
    },
    propagated_confidence: {
      ...stabilityIntelligence.propagated_confidence,
      target_node_id: forecastVersions.at(-1)?.version_id ?? 'forecast-v1',
    },
    generated_at: new Date(
      Date.parse(effectiveAsOfTimes.at(-1)) + 4_000,
    ).toISOString(),
  }
}

const emptyWeatherMetric = {
  present_count: 0,
  coverage_ratio: 0,
}

const weatherContext = {
  version: 'weather-context-read-result-v1',
  trajectory_id: advancedTrajectoryID,
  as_of_time: '2026-08-04T18:00:00Z',
  weather: {
    schema_version: 'weather-feature-v1',
    status: 'available',
    trajectory_id: advancedTrajectoryID,
    as_of_time: '2026-08-04T18:00:00Z',
    samples: [
      {
        sequence: 1,
        position: {
          latitude: 40.4093,
          longitude: 49.8671,
          altitude_meters: 10_668,
          vertical_reference: 'barometric',
        },
        source: {
          provider: 'open-meteo-fixture',
          dataset: 'forecast',
          evidence_kind: 'estimated_weather',
          horizontal_resolution_kilometers: 11,
          temporal_resolution_seconds: 3600,
        },
        features: { temperature_celsius: 29.4 },
        valid_at: '2026-08-04T18:00:00Z',
        available_at: '2026-08-04T17:55:00Z',
        retrieved_at: '2026-08-04T18:00:01Z',
      },
    ],
    confidence: {
      score: 0.7,
      level: 'medium',
      reasons: [],
    },
    limitations: [],
    explanations: [],
    scope_guard: 'research_only',
    input_fingerprint: advancedFingerprint,
    source_names: ['open-meteo-fixture'],
    latest_available_at: '2026-08-04T17:55:00Z',
    generated_at: '2026-08-04T18:00:02Z',
  },
  trust: {
    version: 'weather-trust-v1',
    decision: 'limited',
    usable: true,
    as_of_time: '2026-08-04T18:00:00Z',
    score: 0.7,
    components: [{ name: 'freshness', score: 0.9, weight: 0.5 }],
    allowed_scopes: ['research_projection'],
    limitations: [],
    explanations: [],
    input_fingerprint: advancedFingerprint,
  },
  alignment: {
    version: 'weather-alignment-v1',
    status: 'available',
    trajectory_id: advancedTrajectoryID,
    as_of_time: '2026-08-04T18:00:00Z',
    trust_decision: 'limited',
    trust_score: 0.7,
    point_count: 1,
    aligned_count: 1,
    unmatched_count: 0,
    coverage_ratio: 1,
    matches: [{ trajectory_sequence: 1, weather_sequence: 1 }],
    limitations: [],
    explanations: [],
    input_fingerprint: advancedFingerprint,
    generated_at: '2026-08-04T18:00:03Z',
  },
  encounter: {
    version: 'weather-encounter-v1',
    status: 'available',
    trajectory_id: advancedTrajectoryID,
    as_of_time: '2026-08-04T18:00:00Z',
    alignment_status: 'available',
    alignment_coverage_ratio: 1,
    point_count: 1,
    encounter_point_count: 1,
    unprofiled_point_count: 0,
    profile_coverage_ratio: 1,
    temperature_celsius: {
      present_count: 1,
      coverage_ratio: 1,
      minimum: 29.4,
      maximum: 29.4,
      mean: 29.4,
    },
    precipitation_millimeters: emptyWeatherMetric,
    cloud_cover_percent: emptyWeatherMetric,
    wind_speed_meters_per_second: emptyWeatherMetric,
    wind_gusts_meters_per_second: emptyWeatherMetric,
    wind_direction_degrees: {
      present_count: 0,
      coverage_ratio: 0,
    },
    limitations: [],
    explanations: [],
    input_fingerprint: advancedFingerprint,
    generated_at: '2026-08-04T18:00:04Z',
  },
  uncertainty: {
    version: 'weather-uncertainty-v1',
    status: 'available',
    trajectory_id: advancedTrajectoryID,
    as_of_time: '2026-08-04T18:00:00Z',
    severity_score: 0.2,
    weather_multiplier: 1.1,
    point_adjustments: [{ sequence: 1, multiplier: 1.1 }],
    limitations: [],
    explanations: [],
    input_fingerprint: advancedFingerprint,
    generated_at: '2026-08-04T18:00:05Z',
  },
  input_fingerprint: advancedFingerprint,
  generated_at: '2026-08-04T18:00:06Z',
}

const airspaceAnalytics = {
  version: 'airspace-production-v1',
  schema_version: 'airspace-region-analytics-v1',
  status: 'available',
  region_code: 'az',
  window_start: '2026-08-04T17:55:00Z',
  window_end: '2026-08-04T18:00:00Z',
  occupancy: {
    bucket_duration_seconds: 60,
    latitude_cell_degrees: 0.25,
    longitude_cell_degrees: 0.25,
    altitude_band_meters: 1_000,
    buckets: [],
    metrics: {
      bucket_count: 5,
      expected_bucket_count: 5,
      occupied_cell_count: 2,
      aircraft_observation_count: 2,
      unique_aircraft_count: 2,
      unknown_altitude_count: 0,
      peak_aircraft_per_bucket: 2,
      peak_occupied_cells: 2,
      mean_aircraft_per_bucket: 0.4,
      temporal_coverage: 1,
    },
  },
  sector_complexity: [],
  metrics: {
    snapshot_count: 2,
    bucket_count: 5,
    unique_aircraft_count: 2,
    aircraft_observation_count: 2,
    occupied_cell_count: 2,
    sector_report_count: 0,
    current_aircraft_count: 2,
    peak_aircraft_per_bucket: 2,
    mean_aircraft_per_bucket: 0.4,
    mean_complexity_score: 0,
    peak_complexity_score: 0,
    airspace_pressure_index: 0.1,
    peak_airspace_pressure_index: 0.1,
    moderate_sector_count: 0,
    high_sector_count: 0,
    severe_sector_count: 0,
    contextual_risk_count: 0,
    elevated_risk_count: 0,
    high_risk_count: 0,
    indeterminate_risk_count: 0,
    unknown_altitude_count: 0,
    temporal_coverage: 1,
    occupancy_trend: 'stable',
    highest_complexity_level: 'low',
  },
  confidence: {
    score: 0.82,
    level: 'high',
    components: [{ name: 'coverage', score: 1, weight: 0.5 }],
    reasons: [],
  },
  limitations: [
    {
      code: 'research_only',
      message: 'Not suitable for separation or air traffic control.',
      scope: 'airspace',
    },
  ],
  explanations: [],
  scope_guard: 'research_only',
  provenance: {
    input_fingerprint: advancedFingerprint,
    scene_fingerprints: [advancedFingerprint],
    scan_fingerprints: [],
    risk_fingerprints: [],
    source_names: ['playwright-fixture'],
    latest_observed_at: '2026-08-04T18:00:00Z',
  },
  generated_at: '2026-08-04T18:00:06Z',
}

const routeIntelligenceTrajectoryID =
  '11111111-1111-4111-8111-111111111111'
const routeIntelligenceRecordID =
  '22222222-2222-4222-8222-222222222222'
const routeIntelligenceMutationKey =
  'playwright-route-intelligence-key-v1'

const routeIntelligenceRecord = {
  id: routeIntelligenceRecordID,
  input_fingerprint: `sha256:${'d'.repeat(64)}`,
  stored_at: '2026-08-05T18:00:02Z',
  result: {
    schema_version: 'route-intelligence-v1',
    status: 'complete',
    trajectory_id: routeIntelligenceTrajectoryID,
    identity_key: `flight-identity-${'e'.repeat(64)}`,
    flight_id: '33333333-3333-4333-8333-333333333333',
    aircraft_id: '44444444-4444-4444-8444-444444444444',
    icao24: '4B1801',
    callsign: 'AZAL101',
    window: {
      start_time: '2026-08-05T16:00:00Z',
      end_time: '2026-08-05T18:00:00Z',
      as_of_time: '2026-08-05T18:00:00Z',
    },
    origin: {
      role: 'origin',
      airport: {
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
      },
      distance_km: 7.4,
      confidence: {
        score: 0.92,
        level: 'high',
        evidence_count: 1,
        reasons: [
          {
            code: 'trajectory_endpoint_proximity',
            message: 'The first trajectory segment is close to UBBB.',
            contribution: 0.62,
          },
        ],
      },
      evidence: [
        {
          type: 'trajectory_endpoint_proximity',
          source_name: 'playwright-fixture',
          source_version: 'route-intelligence-v1',
          score: 0.92,
          weight: 0.8,
          observed_at: '2026-08-05T16:00:00Z',
          summary: 'Observed origin endpoint proximity.',
          attributes: [
            { key: 'airport_icao', value: 'UBBB' },
            { key: 'distance_km', value: '7.4' },
          ],
        },
      ],
      limitations: [],
    },
    destination: {
      role: 'destination',
      airport: {
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
      },
      distance_km: 11.2,
      confidence: {
        score: 0.88,
        level: 'high',
        evidence_count: 1,
        reasons: [
          {
            code: 'trajectory_endpoint_proximity',
            message: 'The final trajectory segment is close to LTFM.',
            contribution: 0.58,
          },
        ],
      },
      evidence: [
        {
          type: 'trajectory_endpoint_proximity',
          source_name: 'playwright-fixture',
          source_version: 'route-intelligence-v1',
          score: 0.88,
          weight: 0.8,
          observed_at: '2026-08-05T18:00:00Z',
          summary: 'Observed destination endpoint proximity.',
          attributes: [
            { key: 'airport_icao', value: 'LTFM' },
            { key: 'distance_km', value: '11.2' },
          ],
        },
      ],
      limitations: [],
    },
    summary: {
      great_circle_distance_km: 1789.4,
      same_airport: false,
    },
    confidence: {
      score: 0.9,
      level: 'high',
      evidence_count: 2,
      reasons: [
        {
          code: 'endpoint_pair_supported',
          message: 'Both inferred endpoints have independent evidence.',
          contribution: 0.9,
        },
      ],
    },
    limitations: [
      {
        code: 'inferred_not_filed',
        message: 'This route is inferred from observations and is not a filed flight plan.',
        scope: 'route',
      },
    ],
    provenance: {
      resolver_version: 'route-resolver-v1',
      input_fingerprint: `sha256:${'d'.repeat(64)}`,
      trajectory_updated_at: '2026-08-05T18:00:00Z',
      source_names: ['playwright-fixture'],
    },
    generated_at: '2026-08-05T18:00:01Z',
  },
}

const routeIntelligenceHistory = {
  items: [routeIntelligenceRecord],
  has_more: true,
  next_before_as_of_time: '2026-08-05T18:00:00Z',
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
  if (/^\/api\/v1\/aircraft\/[^/]+\/transponder-evidence\/latest$/.test(pathname)) {
    return '/api/v1/aircraft/{icao24}/transponder-evidence/latest'
  }
  if (/^\/api\/v1\/airports\/[^/]+\/intelligence\/overview$/.test(pathname)) {
    return '/api/v1/airports/{icao}/intelligence/overview'
  }
  if (/^\/api\/v1\/airports\/[^/]+\/intelligence\/history$/.test(pathname)) {
    return '/api/v1/airports/{icao}/intelligence/history'
  }
  if (/^\/api\/v1\/airports\/[^/]+\/intelligence\/trends$/.test(pathname)) {
    return '/api/v1/airports/{icao}/intelligence/trends'
  }
  if (/^\/api\/v1\/trajectories\/[^/]+\/projection-intelligence$/.test(pathname)) {
    return '/api/v1/trajectories/{id}/projection-intelligence'
  }
  if (/^\/api\/v1\/trajectories\/[^/]+\/stability-intelligence$/.test(pathname)) {
    return '/api/v1/trajectories/{id}/stability-intelligence'
  }
  if (/^\/api\/v1\/trajectories\/[^/]+\/weather-context$/.test(pathname)) {
    return '/api/v1/trajectories/{id}/weather-context'
  }
  if (/^\/api\/v1\/airspace\/regions\/[^/]+\/analytics$/.test(pathname)) {
    return '/api/v1/airspace/regions/{code}/analytics'
  }
  if (/^\/api\/v1\/trajectories\/[^/]+\/route-intelligence$/.test(pathname)) {
    return '/api/v1/trajectories/{id}/route-intelligence'
  }
  if (/^\/api\/v1\/trajectories\/[^/]+\/route-intelligence\/latest$/.test(pathname)) {
    return '/api/v1/trajectories/{id}/route-intelligence/latest'
  }
  if (/^\/api\/v1\/trajectories\/[^/]+\/route-intelligence\/history$/.test(pathname)) {
    return '/api/v1/trajectories/{id}/route-intelligence/history'
  }
  return pathname
}

export function resolveMockResponse({
  method,
  requestURL,
  scenario = 'healthy',
  headers = {},
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

  if (method === 'GET' && pathname === '/api/v1/traffic/live') {
    if (scenario === 'traffic-error') {
      return failure(
        503,
        'TRAFFIC_FIXTURE_UNAVAILABLE',
        'The deterministic traffic fixture is unavailable.',
      )
    }

    const serverTime = '2026-08-04T18:00:10Z'
    const limitCandidate = Number.parseInt(
      url.searchParams.get('limit') ?? '1500',
      10,
    )
    const limit =
      Number.isInteger(limitCandidate) &&
      limitCandidate >= 1 &&
      limitCandidate <= 5000
        ? limitCandidate
        : 1500

    const liveAircraft = traffic.map((item, index) => ({
      icao24: item.icao24,
      callsign: item.callsign,
      latitude: item.latitude,
      longitude: item.longitude,
      altitude_m: item.altitude_m,
      velocity_mps: item.velocity_mps,
      heading_degrees: item.heading_degrees,
      vertical_rate_mps: index === 0 ? 0 : 1.2,
      on_ground: item.on_ground,
      observed_at: item.observed_at,
      received_at: '2026-08-04T18:00:06Z',
      source: 'playwright-fixture',
      freshness_ms:
        Date.parse(serverTime) - Date.parse(item.observed_at),
    }))

    return success({
      server_time: serverTime,
      sequence: 42,
      aircraft: liveAircraft.slice(0, limit),
      total_active: liveAircraft.length,
      matching: liveAircraft.length,
      truncated: liveAircraft.length > limit,
    })
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

  if (
    scenario === 'aircraft-error' &&
    method === 'GET' &&
    [
      '/api/v1/aircraft/{icao24}',
      '/api/v1/aircraft/{icao24}/trajectory',
      '/api/v1/aircraft/{icao24}/route-context',
    ].includes(normalizePath(pathname))
  ) {
    return failure(
      503,
      'AIRCRAFT_FIXTURE_UNAVAILABLE',
      'The deterministic aircraft intelligence fixture is unavailable.',
    )
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

  if (
    method === 'GET' &&
    normalizePath(pathname) ===
      '/api/v1/aircraft/{icao24}/transponder-evidence/latest'
  ) {
    return success(transponderEvidence)
  }

  if (method === 'GET' && pathname === '/api/v1/weather/current') {
    return success(currentWeather)
  }

  const analyticalMetrics = new Map([
    ['/api/v1/analytics/metrics/active-aircraft', ['active_aircraft', 2]],
    ['/api/v1/analytics/metrics/traffic-density', ['traffic_density', 0.0042]],
    ['/api/v1/analytics/metrics/airport-activity', ['airport_activity', 40]],
    ['/api/v1/analytics/metrics/coverage-score', ['coverage_score', 0.8333]],
    ['/api/v1/analytics/metrics/data-freshness', ['data_freshness', 0.97]],
  ])
  if (method === 'GET' && analyticalMetrics.has(pathname)) {
    const [metric, value] = analyticalMetrics.get(pathname)
    return success(analyticalMetricFixture(metric, value))
  }

  if (
    scenario === 'airport-error' &&
    method === 'GET' &&
    (
      pathname === '/api/v1/airports/intelligence/ranking' ||
      [
        '/api/v1/airports/{icao}/intelligence/overview',
        '/api/v1/airports/{icao}/intelligence/history',
        '/api/v1/airports/{icao}/intelligence/trends',
      ].includes(normalizePath(pathname))
    )
  ) {
    return failure(
      503,
      'AIRPORT_INTELLIGENCE_FIXTURE_UNAVAILABLE',
      'The deterministic Airport Intelligence fixture is unavailable.',
    )
  }

  if (
    method === 'GET' &&
    pathname === '/api/v1/airports/intelligence/ranking'
  ) {
    return success(airportRanking)
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) ===
      '/api/v1/airports/{icao}/intelligence/overview'
  ) {
    return success(airportOverview)
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) ===
      '/api/v1/airports/{icao}/intelligence/history'
  ) {
    return success(airportHistory)
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) ===
      '/api/v1/airports/{icao}/intelligence/trends'
  ) {
    return success(airportTrends)
  }

  if (
    scenario === 'historical-error' &&
    method === 'GET' &&
    (
      pathname === '/api/v1/historical-intelligence/aggregates/latest' ||
      pathname === '/api/v1/historical-intelligence/aggregates/history'
    )
  ) {
    return failure(
      503,
      'HISTORICAL_INTELLIGENCE_FIXTURE_UNAVAILABLE',
      'The deterministic Historical Intelligence fixture is unavailable.',
    )
  }

  if (
    method === 'GET' &&
    pathname === '/api/v1/historical-intelligence/aggregates/latest'
  ) {
    return success(historicalRecord)
  }
  if (
    method === 'GET' &&
    pathname === '/api/v1/historical-intelligence/aggregates/history'
  ) {
    return success({
      items: [historicalRecord],
      has_more: true,
      next_cursor: 'fixture-cursor-v1',
    })
  }

  if (
    scenario === 'intelligence-error' &&
    method === 'GET' &&
    [
      '/api/v1/trajectories/{id}/projection-intelligence',
      '/api/v1/trajectories/{id}/stability-intelligence',
      '/api/v1/trajectories/{id}/weather-context',
    ].includes(normalizePath(pathname))
  ) {
    return failure(
      503,
      'ADVANCED_INTELLIGENCE_FIXTURE_UNAVAILABLE',
      'The deterministic advanced intelligence fixture is unavailable.',
    )
  }

  if (
    method === 'GET' &&
    normalizePath(pathname) ===
      '/api/v1/trajectories/{id}/projection-intelligence'
  ) {
    return success(projectionFixture())
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) ===
      '/api/v1/trajectories/{id}/stability-intelligence'
  ) {
    const requestedAsOfTimes = (
      url.searchParams.get('as_of_times') ?? ''
    ).split(',')
    return success(stabilityIntelligenceFixture(requestedAsOfTimes))
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) === '/api/v1/trajectories/{id}/weather-context'
  ) {
    return success(weatherContext)
  }
  if (
    method === 'GET' &&
    normalizePath(pathname) === '/api/v1/airspace/regions/{code}/analytics'
  ) {
    return success(airspaceAnalytics)
  }

  if (
    method === 'POST' &&
    normalizePath(pathname) ===
      '/api/v1/trajectories/{id}/route-intelligence'
  ) {
    const candidate = String(
      headers['x-internal-api-key'] ??
        headers['X-Internal-API-Key'] ??
        '',
    )
    if (candidate !== routeIntelligenceMutationKey) {
      return failure(
        401,
        'MUTATION_AUTHENTICATION_REQUIRED',
        'Valid internal mutation credentials are required',
      )
    }
    return success(routeIntelligenceRecord)
  }

  if (
    method === 'GET' &&
    normalizePath(pathname) ===
      '/api/v1/trajectories/{id}/route-intelligence/latest'
  ) {
    return success(routeIntelligenceRecord)
  }

  if (
    method === 'GET' &&
    normalizePath(pathname) ===
      '/api/v1/trajectories/{id}/route-intelligence/history'
  ) {
    return success(routeIntelligenceHistory)
  }

  return failure(404, 'E2E_ROUTE_NOT_FOUND', 'Mock API route not found')
}

function writeJSON(response, resolved) {
  const body = JSON.stringify(resolved.body)
  response.writeHead(resolved.status, {
    'Access-Control-Allow-Headers': 'Accept,Content-Type,X-Internal-API-Key',
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
          'Access-Control-Allow-Headers': 'Accept,Content-Type,X-Internal-API-Key',
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
          headers: request.headers,
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
