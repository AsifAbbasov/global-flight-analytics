import { APIRequestError, requestAPIData } from '@/lib/api/client'
import type {
  AirspaceConfidence,
  AirspaceNotice,
  AirspaceRegionAnalyticsRequest,
  AirspaceRegionAnalyticsResponse,
  AirspaceRegionMetrics,
} from '@/types/airspace-intelligence'

export async function getAirspaceRegionAnalytics(
  request: AirspaceRegionAnalyticsRequest
): Promise<AirspaceRegionAnalyticsResponse> {
  const regionCode = text(request.regionCode, 'regionCode').toLowerCase()
  const asOfTime = timestamp(request.asOfTime, 'asOfTime')
  if (!Number.isInteger(request.windowSeconds) || request.windowSeconds < 60) {
    throw new APIRequestError(
      'Airspace Intelligence window must be at least 60 whole seconds.'
    )
  }

  const searchParams = new URLSearchParams({
    as_of_time: asOfTime,
    window_seconds: String(request.windowSeconds),
  })
  const value = await requestAPIData<unknown>(
    `/api/v1/airspace/regions/${encodeURIComponent(regionCode)}/analytics`,
    {
      signal: request.signal,
      searchParams,
      timeoutMilliseconds: 20_000,
    }
  )
  return parseAirspaceRegionAnalyticsResponse(value)
}

export function parseAirspaceRegionAnalyticsResponse(
  value: unknown
): AirspaceRegionAnalyticsResponse {
  const root = record(value, 'response')
  const occupancy = record(root.occupancy, 'occupancy')
  const occupancyMetrics = record(occupancy.metrics, 'occupancy.metrics')
  const provenance = record(root.provenance, 'provenance')

  return {
    version: text(root.version, 'version'),
    schema_version: text(root.schema_version, 'schema_version'),
    status: text(root.status, 'status'),
    region_code: text(root.region_code, 'region_code').toLowerCase(),
    window_start: timestamp(root.window_start, 'window_start'),
    window_end: timestamp(root.window_end, 'window_end'),
    occupancy: {
      metrics: {
        bucket_count: whole(occupancyMetrics.bucket_count, 'occupancy.metrics.bucket_count'),
        expected_bucket_count: whole(
          occupancyMetrics.expected_bucket_count,
          'occupancy.metrics.expected_bucket_count'
        ),
        aircraft_observation_count: whole(
          occupancyMetrics.aircraft_observation_count,
          'occupancy.metrics.aircraft_observation_count'
        ),
        unique_aircraft_count: whole(
          occupancyMetrics.unique_aircraft_count,
          'occupancy.metrics.unique_aircraft_count'
        ),
        unknown_altitude_count: whole(
          occupancyMetrics.unknown_altitude_count,
          'occupancy.metrics.unknown_altitude_count'
        ),
        peak_aircraft_per_bucket: whole(
          occupancyMetrics.peak_aircraft_per_bucket,
          'occupancy.metrics.peak_aircraft_per_bucket'
        ),
        peak_occupied_cells: whole(
          occupancyMetrics.peak_occupied_cells,
          'occupancy.metrics.peak_occupied_cells'
        ),
        mean_aircraft_per_bucket: nonNegative(
          occupancyMetrics.mean_aircraft_per_bucket,
          'occupancy.metrics.mean_aircraft_per_bucket'
        ),
        temporal_coverage: ratio(
          occupancyMetrics.temporal_coverage,
          'occupancy.metrics.temporal_coverage'
        ),
      },
    },
    metrics: parseMetrics(root.metrics),
    confidence: parseConfidence(root.confidence),
    limitations: notices(root.limitations, 'limitations', true),
    explanations: notices(root.explanations, 'explanations', false),
    scope_guard: text(root.scope_guard, 'scope_guard'),
    provenance: {
      input_fingerprint: fingerprint(
        provenance.input_fingerprint,
        'provenance.input_fingerprint'
      ),
      source_names: list(provenance.source_names, 'provenance.source_names').map(
        (item, index) => text(item, `provenance.source_names[${index}]`)
      ),
      latest_observed_at: timestamp(
        provenance.latest_observed_at,
        'provenance.latest_observed_at'
      ),
    },
    generated_at: timestamp(root.generated_at, 'generated_at'),
  }
}

function parseMetrics(value: unknown): AirspaceRegionMetrics {
  const metrics = record(value, 'metrics')
  return {
    snapshot_count: whole(metrics.snapshot_count, 'metrics.snapshot_count'),
    bucket_count: whole(metrics.bucket_count, 'metrics.bucket_count'),
    unique_aircraft_count: whole(
      metrics.unique_aircraft_count,
      'metrics.unique_aircraft_count'
    ),
    aircraft_observation_count: whole(
      metrics.aircraft_observation_count,
      'metrics.aircraft_observation_count'
    ),
    occupied_cell_count: whole(
      metrics.occupied_cell_count,
      'metrics.occupied_cell_count'
    ),
    sector_report_count: whole(
      metrics.sector_report_count,
      'metrics.sector_report_count'
    ),
    current_aircraft_count: whole(
      metrics.current_aircraft_count,
      'metrics.current_aircraft_count'
    ),
    peak_aircraft_per_bucket: whole(
      metrics.peak_aircraft_per_bucket,
      'metrics.peak_aircraft_per_bucket'
    ),
    mean_aircraft_per_bucket: nonNegative(
      metrics.mean_aircraft_per_bucket,
      'metrics.mean_aircraft_per_bucket'
    ),
    mean_complexity_score: ratio(
      metrics.mean_complexity_score,
      'metrics.mean_complexity_score'
    ),
    peak_complexity_score: ratio(
      metrics.peak_complexity_score,
      'metrics.peak_complexity_score'
    ),
    airspace_pressure_index: ratio(
      metrics.airspace_pressure_index,
      'metrics.airspace_pressure_index'
    ),
    peak_airspace_pressure_index: ratio(
      metrics.peak_airspace_pressure_index,
      'metrics.peak_airspace_pressure_index'
    ),
    moderate_sector_count: whole(
      metrics.moderate_sector_count,
      'metrics.moderate_sector_count'
    ),
    high_sector_count: whole(metrics.high_sector_count, 'metrics.high_sector_count'),
    severe_sector_count: whole(
      metrics.severe_sector_count,
      'metrics.severe_sector_count'
    ),
    contextual_risk_count: whole(
      metrics.contextual_risk_count,
      'metrics.contextual_risk_count'
    ),
    elevated_risk_count: whole(
      metrics.elevated_risk_count,
      'metrics.elevated_risk_count'
    ),
    high_risk_count: whole(metrics.high_risk_count, 'metrics.high_risk_count'),
    indeterminate_risk_count: whole(
      metrics.indeterminate_risk_count,
      'metrics.indeterminate_risk_count'
    ),
    unknown_altitude_count: whole(
      metrics.unknown_altitude_count,
      'metrics.unknown_altitude_count'
    ),
    temporal_coverage: ratio(metrics.temporal_coverage, 'metrics.temporal_coverage'),
    occupancy_trend: text(metrics.occupancy_trend, 'metrics.occupancy_trend'),
    highest_complexity_level: text(
      metrics.highest_complexity_level,
      'metrics.highest_complexity_level'
    ),
  }
}

function parseConfidence(value: unknown): AirspaceConfidence {
  const confidence = record(value, 'confidence')
  return {
    score: ratio(confidence.score, 'confidence.score'),
    level: text(confidence.level, 'confidence.level'),
    reasons: list(confidence.reasons, 'confidence.reasons').map((item, index) => {
      const reason = record(item, `confidence.reasons[${index}]`)
      return {
        code: text(reason.code, `confidence.reasons[${index}].code`),
        message: text(reason.message, `confidence.reasons[${index}].message`),
        contribution: finite(
          reason.contribution,
          `confidence.reasons[${index}].contribution`
        ),
      }
    }),
  }
}

function notices(
  value: unknown,
  field: string,
  includeScope: boolean
): AirspaceNotice[] {
  return list(value, field).map((item, index) => {
    const entry = record(item, `${field}[${index}]`)
    return {
      code: text(entry.code, `${field}[${index}].code`),
      message: text(entry.message, `${field}[${index}].message`),
      ...(includeScope
        ? { scope: text(entry.scope, `${field}[${index}].scope`) }
        : {}),
    }
  })
}

function record(value: unknown, field: string): Record<string, unknown> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    invalid(`${field} must be an object.`)
  }
  return value as Record<string, unknown>
}

function list(value: unknown, field: string): unknown[] {
  if (!Array.isArray(value)) invalid(`${field} must be an array.`)
  return value
}

function text(value: unknown, field: string): string {
  if (typeof value !== 'string' || value.trim() === '') {
    invalid(`${field} must be a non-empty string.`)
  }
  return value.trim()
}

function finite(value: unknown, field: string): number {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    invalid(`${field} must be finite.`)
  }
  return value
}

function nonNegative(value: unknown, field: string): number {
  const result = finite(value, field)
  if (result < 0) invalid(`${field} must not be negative.`)
  return result
}

function whole(value: unknown, field: string): number {
  const result = nonNegative(value, field)
  if (!Number.isInteger(result)) invalid(`${field} must be a whole number.`)
  return result
}

function ratio(value: unknown, field: string): number {
  const result = finite(value, field)
  if (result < 0 || result > 1) invalid(`${field} must be between zero and one.`)
  return result
}

function timestamp(value: unknown, field: string): string {
  const result = text(value, field)
  if (Number.isNaN(Date.parse(result))) invalid(`${field} must be a valid timestamp.`)
  return new Date(result).toISOString()
}

function fingerprint(value: unknown, field: string): string {
  const result = text(value, field).toLowerCase()
  if (!/^[0-9a-f]{64}$/.test(result)) {
    invalid(`${field} must be a 64-character hexadecimal fingerprint.`)
  }
  return result
}

function invalid(message: string): never {
  throw new APIRequestError(`The Airspace Intelligence response is invalid: ${message}`)
}
