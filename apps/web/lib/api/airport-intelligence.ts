// FRONTEND_UNIFIED_AIRPORT_ANALYTICS_WORKSPACE_V1

import {
  APIRequestError,
  requestAPIData,
} from '@/lib/api/client'
import type {
  AirportIntelligenceHistory,
  AirportIntelligenceLimitation,
  AirportIntelligenceOverview,
  AirportIntelligenceRanking,
  AirportIntelligenceRankingOptions,
  AirportIntelligenceTrends,
  AirportIntelligenceWindow,
  AirportIntelligenceWindowOptions,
  AirportPassport,
  AirportRankedItem,
  AirportRankingSummary,
  AirportRankingWeights,
  AirportStatistics,
  AirportTrendPoint,
} from '@/types/airport-intelligence'

const minimumDays = 1
const maximumDays = 365
const minimumLimit = 1
const maximumLimit = 200

export async function getAirportIntelligenceRanking(
  options: AirportIntelligenceRankingOptions
): Promise<AirportIntelligenceRanking> {
  const searchParams = buildSearchParams(options)
  searchParams.set('limit', String(normalizeInteger(options.limit, 'limit', minimumLimit, maximumLimit)))

  return parseRanking(
    await requestAPIData<unknown>('/api/v1/airports/intelligence/ranking', {
      signal: options.signal,
      searchParams,
    })
  )
}

export async function getAirportIntelligenceOverview(
  icaoCode: string,
  options: AirportIntelligenceWindowOptions
): Promise<AirportIntelligenceOverview> {
  return parseOverview(
    await requestAPIData<unknown>(
      `/api/v1/airports/${encodeURIComponent(normalizeICAOCode(icaoCode))}/intelligence/overview`,
      {
        signal: options.signal,
        searchParams: buildSearchParams(options),
      }
    )
  )
}

export async function getAirportIntelligenceHistory(
  icaoCode: string,
  options: AirportIntelligenceWindowOptions
): Promise<AirportIntelligenceHistory> {
  return parseHistory(
    await requestAPIData<unknown>(
      `/api/v1/airports/${encodeURIComponent(normalizeICAOCode(icaoCode))}/intelligence/history`,
      {
        signal: options.signal,
        searchParams: buildSearchParams(options),
      }
    )
  )
}

export async function getAirportIntelligenceTrends(
  icaoCode: string,
  options: AirportIntelligenceWindowOptions
): Promise<AirportIntelligenceTrends> {
  return parseTrends(
    await requestAPIData<unknown>(
      `/api/v1/airports/${encodeURIComponent(normalizeICAOCode(icaoCode))}/intelligence/trends`,
      {
        signal: options.signal,
        searchParams: buildSearchParams(options),
      }
    )
  )
}

function buildSearchParams(options: AirportIntelligenceWindowOptions): URLSearchParams {
  const searchParams = new URLSearchParams()
  searchParams.set('days', String(normalizeInteger(options.days, 'days', minimumDays, maximumDays)))
  return searchParams
}

function normalizeICAOCode(value: string): string {
  const normalized = value.trim().toUpperCase()
  if (!/^[A-Z0-9]{4}$/.test(normalized)) {
    throw new APIRequestError('Airport ICAO code must contain exactly four letters or digits.')
  }
  return normalized
}

function normalizeInteger(
  value: number,
  field: string,
  minimum: number,
  maximum: number
): number {
  if (!Number.isInteger(value) || value < minimum || value > maximum) {
    throw new APIRequestError(
      `Airport Intelligence ${field} must be between ${minimum} and ${maximum}.`
    )
  }
  return value
}

function parseRanking(value: unknown): AirportIntelligenceRanking {
  const item = objectValue(value, 'ranking')
  return {
    version: stringValue(item.version, 'ranking.version'),
    window: parseWindow(item.window, 'ranking.window'),
    weights: parseWeights(item.weights, 'ranking.weights'),
    airports: arrayValue(item.airports, 'ranking.airports').map((entry, index) =>
      parseRankedItem(entry, `ranking.airports[${index}]`)
    ),
    limitations: parseLimitations(item.limitations, 'ranking.limitations'),
    generated_at: timestampValue(item.generated_at, 'ranking.generated_at'),
  }
}

function parseOverview(value: unknown): AirportIntelligenceOverview {
  const item = objectValue(value, 'overview')
  return {
    version: stringValue(item.version, 'overview.version'),
    window: parseWindow(item.window, 'overview.window'),
    passport: parsePassport(item.passport, 'overview.passport'),
    statistics: parseStatistics(item.statistics, 'overview.statistics'),
    ranking: parseRankingSummary(item.ranking, 'overview.ranking'),
    limitations: parseLimitations(item.limitations, 'overview.limitations'),
    generated_at: timestampValue(item.generated_at, 'overview.generated_at'),
  }
}

function parseHistory(value: unknown): AirportIntelligenceHistory {
  const item = objectValue(value, 'history')
  return {
    version: stringValue(item.version, 'history.version'),
    window: parseWindow(item.window, 'history.window'),
    icao_code: icaoValue(item.icao_code, 'history.icao_code'),
    entries: arrayValue(item.entries, 'history.entries').map((entry, index) =>
      parseStatistics(entry, `history.entries[${index}]`)
    ),
    limitations: parseLimitations(item.limitations, 'history.limitations'),
    generated_at: timestampValue(item.generated_at, 'history.generated_at'),
  }
}

function parseTrends(value: unknown): AirportIntelligenceTrends {
  const item = objectValue(value, 'trends')
  return {
    version: stringValue(item.version, 'trends.version'),
    window: parseWindow(item.window, 'trends.window'),
    icao_code: icaoValue(item.icao_code, 'trends.icao_code'),
    compared_windows: integerValue(item.compared_windows, 'trends.compared_windows'),
    window_duration_seconds: integerValue(
      item.window_duration_seconds,
      'trends.window_duration_seconds'
    ),
    direction: stringValue(item.direction, 'trends.direction'),
    baseline: parseTrendPoint(item.baseline, 'trends.baseline'),
    current: parseTrendPoint(item.current, 'trends.current'),
    peak: parseTrendPoint(item.peak, 'trends.peak'),
    total_movements_change: integerValue(
      item.total_movements_change,
      'trends.total_movements_change',
      true
    ),
    movements_per_hour_change: finiteNumber(
      item.movements_per_hour_change,
      'trends.movements_per_hour_change'
    ),
    movements_per_hour_change_percent: finiteNumber(
      item.movements_per_hour_change_percent,
      'trends.movements_per_hour_change_percent'
    ),
    movements_per_hour_change_percent_known: booleanValue(
      item.movements_per_hour_change_percent_known,
      'trends.movements_per_hour_change_percent_known'
    ),
    active_routes_change: integerValue(
      item.active_routes_change,
      'trends.active_routes_change',
      true
    ),
    coverage_score_change: finiteNumber(
      item.coverage_score_change,
      'trends.coverage_score_change'
    ),
    freshness_score_change: finiteNumber(
      item.freshness_score_change,
      'trends.freshness_score_change'
    ),
    gap_count: integerValue(item.gap_count, 'trends.gap_count'),
    gap_duration_seconds: integerValue(
      item.gap_duration_seconds,
      'trends.gap_duration_seconds'
    ),
    observed_duration_seconds: integerValue(
      item.observed_duration_seconds,
      'trends.observed_duration_seconds'
    ),
    continuity_score: ratioValue(item.continuity_score, 'trends.continuity_score'),
    limitations: parseLimitations(item.limitations, 'trends.limitations'),
    generated_at: timestampValue(item.generated_at, 'trends.generated_at'),
  }
}

function parseWindow(value: unknown, field: string): AirportIntelligenceWindow {
  const item = objectValue(value, field)
  return {
    start_time: timestampValue(item.start_time, `${field}.start_time`),
    end_time: timestampValue(item.end_time, `${field}.end_time`),
    as_of_time: timestampValue(item.as_of_time, `${field}.as_of_time`),
    completed_days: integerValue(item.completed_days, `${field}.completed_days`),
  }
}

function parseWeights(value: unknown, field: string): AirportRankingWeights {
  const item = objectValue(value, field)
  return {
    movements: ratioValue(item.movements, `${field}.movements`),
    routes: ratioValue(item.routes, `${field}.routes`),
    observations: ratioValue(item.observations, `${field}.observations`),
    intensity: ratioValue(item.intensity, `${field}.intensity`),
    coverage: ratioValue(item.coverage, `${field}.coverage`),
    freshness: ratioValue(item.freshness, `${field}.freshness`),
  }
}

function parseRankedItem(value: unknown, field: string): AirportRankedItem {
  const item = objectValue(value, field)
  return {
    position: integerValue(item.position, `${field}.position`),
    icao_code: icaoValue(item.icao_code, `${field}.icao_code`),
    iata_code: stringValue(item.iata_code, `${field}.iata_code`, true),
    name: stringValue(item.name, `${field}.name`),
    city: stringValue(item.city, `${field}.city`, true),
    country: stringValue(item.country, `${field}.country`, true),
    activity_score: ratioValue(item.activity_score, `${field}.activity_score`),
    data_confidence: ratioValue(item.data_confidence, `${field}.data_confidence`),
    movements_component: ratioValue(item.movements_component, `${field}.movements_component`),
    routes_component: ratioValue(item.routes_component, `${field}.routes_component`),
    observations_component: ratioValue(
      item.observations_component,
      `${field}.observations_component`
    ),
    intensity_component: ratioValue(item.intensity_component, `${field}.intensity_component`),
    coverage_score: ratioValue(item.coverage_score, `${field}.coverage_score`),
    freshness_score: ratioValue(item.freshness_score, `${field}.freshness_score`),
    total_movements: integerValue(item.total_movements, `${field}.total_movements`),
    active_routes: integerValue(item.active_routes, `${field}.active_routes`),
    observed_samples: integerValue(item.observed_samples, `${field}.observed_samples`),
    expected_samples: integerValue(item.expected_samples, `${field}.expected_samples`),
    movements_per_hour: nonNegativeNumber(
      item.movements_per_hour,
      `${field}.movements_per_hour`
    ),
    active_aircraft: integerValue(item.active_aircraft, `${field}.active_aircraft`),
  }
}

function parsePassport(value: unknown, field: string): AirportPassport {
  const item = objectValue(value, field)
  const identity = objectValue(item.identity, `${field}.identity`)
  const location = objectValue(item.location, `${field}.location`)
  const operations = objectValue(item.operations, `${field}.operations`)
  const quality = objectValue(item.data_quality, `${field}.data_quality`)
  return {
    identity: {
      icao_code: icaoValue(identity.icao_code, `${field}.identity.icao_code`),
      iata_code: stringValue(identity.iata_code, `${field}.identity.iata_code`, true),
      name: stringValue(identity.name, `${field}.identity.name`),
    },
    location: {
      city: stringValue(location.city, `${field}.location.city`, true),
      country: stringValue(location.country, `${field}.location.country`, true),
      latitude: boundedNumber(location.latitude, `${field}.location.latitude`, -90, 90),
      longitude: boundedNumber(location.longitude, `${field}.location.longitude`, -180, 180),
      elevation_m: nullableFiniteNumber(location.elevation_m, `${field}.location.elevation_m`),
      elevation_status: stringValue(
        location.elevation_status,
        `${field}.location.elevation_status`
      ),
      timezone: stringValue(location.timezone, `${field}.location.timezone`, true),
    },
    operations: {
      arrivals: integerValue(operations.arrivals, `${field}.operations.arrivals`),
      departures: integerValue(operations.departures, `${field}.operations.departures`),
      activity: integerValue(operations.activity, `${field}.operations.activity`),
      active_aircraft: integerValue(
        operations.active_aircraft,
        `${field}.operations.active_aircraft`
      ),
    },
    data_quality: {
      freshness_score: ratioValue(
        quality.freshness_score,
        `${field}.data_quality.freshness_score`
      ),
      coverage_score: ratioValue(
        quality.coverage_score,
        `${field}.data_quality.coverage_score`
      ),
      observed_at: timestampValue(
        quality.observed_at,
        `${field}.data_quality.observed_at`
      ),
    },
    description: stringValue(item.description, `${field}.description`, true),
    generated_at: timestampValue(item.generated_at, `${field}.generated_at`),
  }
}

function parseStatistics(value: unknown, field: string): AirportStatistics {
  const item = objectValue(value, field)
  return {
    icao_code: icaoValue(item.icao_code, `${field}.icao_code`),
    window_start: timestampValue(item.window_start, `${field}.window_start`),
    window_end: timestampValue(item.window_end, `${field}.window_end`),
    arrivals: integerValue(item.arrivals, `${field}.arrivals`),
    departures: integerValue(item.departures, `${field}.departures`),
    total_movements: integerValue(item.total_movements, `${field}.total_movements`),
    arrival_share: ratioValue(item.arrival_share, `${field}.arrival_share`),
    departure_share: ratioValue(item.departure_share, `${field}.departure_share`),
    movements_per_hour: nonNegativeNumber(
      item.movements_per_hour,
      `${field}.movements_per_hour`
    ),
    active_aircraft: integerValue(item.active_aircraft, `${field}.active_aircraft`),
    active_routes: integerValue(item.active_routes, `${field}.active_routes`),
    observed_samples: integerValue(item.observed_samples, `${field}.observed_samples`),
    expected_samples: integerValue(item.expected_samples, `${field}.expected_samples`),
    coverage_score: ratioValue(item.coverage_score, `${field}.coverage_score`),
    freshness_score: ratioValue(item.freshness_score, `${field}.freshness_score`),
    latest_observation_at: timestampValue(
      item.latest_observation_at,
      `${field}.latest_observation_at`
    ),
    generated_at: timestampValue(item.generated_at, `${field}.generated_at`),
  }
}

function parseRankingSummary(value: unknown, field: string): AirportRankingSummary {
  const item = objectValue(value, field)
  return {
    position: integerValue(item.position, `${field}.position`),
    total_airports: integerValue(item.total_airports, `${field}.total_airports`),
    activity_score: ratioValue(item.activity_score, `${field}.activity_score`),
    data_confidence: ratioValue(item.data_confidence, `${field}.data_confidence`),
    movements_component: ratioValue(item.movements_component, `${field}.movements_component`),
    routes_component: ratioValue(item.routes_component, `${field}.routes_component`),
    observations_component: ratioValue(
      item.observations_component,
      `${field}.observations_component`
    ),
    intensity_component: ratioValue(item.intensity_component, `${field}.intensity_component`),
  }
}

function parseTrendPoint(value: unknown, field: string): AirportTrendPoint {
  const item = objectValue(value, field)
  return {
    window_start: timestampValue(item.window_start, `${field}.window_start`),
    window_end: timestampValue(item.window_end, `${field}.window_end`),
    total_movements: integerValue(item.total_movements, `${field}.total_movements`),
    movements_per_hour: nonNegativeNumber(
      item.movements_per_hour,
      `${field}.movements_per_hour`
    ),
    active_routes: integerValue(item.active_routes, `${field}.active_routes`),
    coverage_score: ratioValue(item.coverage_score, `${field}.coverage_score`),
    freshness_score: ratioValue(item.freshness_score, `${field}.freshness_score`),
  }
}

function parseLimitations(value: unknown, field: string): AirportIntelligenceLimitation[] {
  return arrayValue(value, field).map((entry, index) => {
    const item = objectValue(entry, `${field}[${index}]`)
    return {
      code: stringValue(item.code, `${field}[${index}].code`),
      message: stringValue(item.message, `${field}[${index}].message`),
    }
  })
}

function objectValue(value: unknown, field: string): Record<string, unknown> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    invalid(`${field} must be an object.`)
  }
  return value as Record<string, unknown>
}

function arrayValue(value: unknown, field: string): unknown[] {
  if (!Array.isArray(value)) invalid(`${field} must be an array.`)
  return value
}

function stringValue(value: unknown, field: string, allowEmpty = false): string {
  if (typeof value !== 'string') invalid(`${field} must be a string.`)
  if (!allowEmpty && value.trim() === '') invalid(`${field} must not be empty.`)
  return value
}

function icaoValue(value: unknown, field: string): string {
  const normalized = stringValue(value, field).trim().toUpperCase()
  if (!/^[A-Z0-9]{4}$/.test(normalized)) invalid(`${field} must be a four-character ICAO code.`)
  return normalized
}

function finiteNumber(value: unknown, field: string): number {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    invalid(`${field} must be a finite number.`)
  }
  return value
}

function nullableFiniteNumber(value: unknown, field: string): number | null {
  return value === null ? null : finiteNumber(value, field)
}

function nonNegativeNumber(value: unknown, field: string): number {
  const parsed = finiteNumber(value, field)
  if (parsed < 0) invalid(`${field} must not be negative.`)
  return parsed
}

function boundedNumber(value: unknown, field: string, minimum: number, maximum: number): number {
  const parsed = finiteNumber(value, field)
  if (parsed < minimum || parsed > maximum) invalid(`${field} is outside its allowed range.`)
  return parsed
}

function integerValue(value: unknown, field: string, signed = false): number {
  const parsed = finiteNumber(value, field)
  if (!Number.isInteger(parsed) || (!signed && parsed < 0)) {
    invalid(`${field} must be ${signed ? 'an integer' : 'a non-negative integer'}.`)
  }
  return parsed
}

function ratioValue(value: unknown, field: string): number {
  const parsed = finiteNumber(value, field)
  if (parsed < 0 || parsed > 1) invalid(`${field} must be between zero and one.`)
  return parsed
}

function booleanValue(value: unknown, field: string): boolean {
  if (typeof value !== 'boolean') invalid(`${field} must be a boolean.`)
  return value
}

function timestampValue(value: unknown, field: string): string {
  const parsed = stringValue(value, field)
  if (Number.isNaN(Date.parse(parsed))) invalid(`${field} must be a valid timestamp.`)
  return parsed
}

function invalid(message: string): never {
  throw new APIRequestError(`The Airport Intelligence response is invalid: ${message}`)
}
