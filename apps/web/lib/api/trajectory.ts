import {
  APIRequestError,
  requestAPIData,
  type APIRequestOptions,
} from '@/lib/api/client'
import type {
  AircraftTrajectory,
  CoverageGap,
  CoverageGapReason,
  FlightIdentityBasis,
  FlightSplitReason,
  TrajectorySegment,
  TrajectorySegmentStatus,
} from '@/types/trajectory'

const trajectorySegmentStatuses = new Set<TrajectorySegmentStatus>([
  'observed',
  'interpolated',
  'estimated',
  'invalid',
])

const coverageGapReasons = new Set<CoverageGapReason>([
  'time_gap',
  'movement_jump',
  'unknown',
])

const flightIdentityBases = new Set<FlightIdentityBasis>([
  'source_flight_id',
  'callsign_and_start_time',
  'aircraft_and_start_time',
])

const flightSplitReasons = new Set<FlightSplitReason>([
  'initial_observation',
  'source_flight_id_changed',
  'callsign_changed',
  'ground_cycle',
  'continued_from_previous_batch',
])

export async function getLatestAircraftTrajectory(
  icao24: string,
  options: APIRequestOptions = {}
): Promise<AircraftTrajectory> {
  const normalizedICAO24 = normalizeICAO24(icao24)
  const data = await requestAPIData<unknown>(
    `/api/v1/aircraft/${encodeURIComponent(
      normalizedICAO24
    )}/trajectory`,
    options
  )

  return parseAircraftTrajectory(data)
}

function normalizeICAO24(value: string): string {
  const normalized = value.trim().toLowerCase()

  if (!/^[0-9a-f]{6}$/.test(normalized)) {
    throw new APIRequestError(
      'ICAO24 must contain exactly six hexadecimal characters.'
    )
  }

  return normalized
}

function parseAircraftTrajectory(
  value: unknown
): AircraftTrajectory {
  const record = requireRecord(value, 'trajectory')
  const identityBasis = requireString(
    record.identity_basis,
    'identity_basis'
  )
  const splitReason = requireString(
    record.split_reason,
    'split_reason'
  )

  if (
    !flightIdentityBases.has(
      identityBasis as FlightIdentityBasis
    )
  ) {
    throw invalidPayload(
      'identity_basis contains an unsupported value.'
    )
  }

  if (
    !flightSplitReasons.has(
      splitReason as FlightSplitReason
    )
  ) {
    throw invalidPayload(
      'split_reason contains an unsupported value.'
    )
  }

  return {
    id: requireString(record.id, 'id'),
    identity_key: requireString(
      record.identity_key,
      'identity_key'
    ),
    identity_basis: identityBasis as FlightIdentityBasis,
    split_reason: splitReason as FlightSplitReason,
    flight_id: requireString(record.flight_id, 'flight_id', true),
    aircraft_id: requireString(
      record.aircraft_id,
      'aircraft_id',
      true
    ),
    icao24: requireString(record.icao24, 'icao24').toLowerCase(),
    callsign: requireString(record.callsign, 'callsign', true),
    start_time: requireTimestamp(record.start_time, 'start_time'),
    end_time: requireTimestamp(record.end_time, 'end_time'),
    duration_seconds: requireNonNegativeNumber(
      record.duration_seconds,
      'duration_seconds'
    ),
    segment_count: requireNonNegativeInteger(
      record.segment_count,
      'segment_count'
    ),
    point_count: requireNonNegativeInteger(
      record.point_count,
      'point_count'
    ),
    coverage_gap_count: requireNonNegativeInteger(
      record.coverage_gap_count,
      'coverage_gap_count'
    ),
    quality_score: requireFiniteNumber(
      record.quality_score,
      'quality_score'
    ),
    source_name: requireString(
      record.source_name,
      'source_name',
      true
    ),
    segments: requireArray(record.segments, 'segments').map(
      (item, index) => parseTrajectorySegment(item, index)
    ),
    coverage_gaps: requireArray(
      record.coverage_gaps,
      'coverage_gaps'
    ).map((item, index) => parseCoverageGap(item, index)),
    created_at: requireTimestamp(record.created_at, 'created_at'),
    updated_at: requireTimestamp(record.updated_at, 'updated_at'),
  }
}

function parseTrajectorySegment(
  value: unknown,
  index: number
): TrajectorySegment {
  const prefix = `segments[${index}]`
  const record = requireRecord(value, prefix)
  const status = requireString(record.status, `${prefix}.status`)

  if (
    !trajectorySegmentStatuses.has(
      status as TrajectorySegmentStatus
    )
  ) {
    throw invalidPayload(
      `${prefix}.status contains an unsupported value.`
    )
  }

  return {
    id: requireString(record.id, `${prefix}.id`),
    trajectory_id: requireString(
      record.trajectory_id,
      `${prefix}.trajectory_id`
    ),
    flight_id: requireString(
      record.flight_id,
      `${prefix}.flight_id`,
      true
    ),
    aircraft_id: requireString(
      record.aircraft_id,
      `${prefix}.aircraft_id`,
      true
    ),
    icao24: requireString(
      record.icao24,
      `${prefix}.icao24`
    ).toLowerCase(),
    callsign: requireString(
      record.callsign,
      `${prefix}.callsign`,
      true
    ),
    sequence_number: requireNonNegativeInteger(
      record.sequence_number,
      `${prefix}.sequence_number`
    ),
    status: status as TrajectorySegmentStatus,
    quality_score: requireFiniteNumber(
      record.quality_score,
      `${prefix}.quality_score`
    ),
    start_time: requireTimestamp(
      record.start_time,
      `${prefix}.start_time`
    ),
    end_time: requireTimestamp(
      record.end_time,
      `${prefix}.end_time`
    ),
    duration_seconds: requireNonNegativeNumber(
      record.duration_seconds,
      `${prefix}.duration_seconds`
    ),
    start_latitude: requireLatitude(
      record.start_latitude,
      `${prefix}.start_latitude`
    ),
    start_longitude: requireLongitude(
      record.start_longitude,
      `${prefix}.start_longitude`
    ),
    end_latitude: requireLatitude(
      record.end_latitude,
      `${prefix}.end_latitude`
    ),
    end_longitude: requireLongitude(
      record.end_longitude,
      `${prefix}.end_longitude`
    ),
    point_count: requireNonNegativeInteger(
      record.point_count,
      `${prefix}.point_count`
    ),
    source_name: requireString(
      record.source_name,
      `${prefix}.source_name`,
      true
    ),
    created_at: requireTimestamp(
      record.created_at,
      `${prefix}.created_at`
    ),
  }
}

function parseCoverageGap(
  value: unknown,
  index: number
): CoverageGap {
  const prefix = `coverage_gaps[${index}]`
  const record = requireRecord(value, prefix)
  const reason = requireString(record.reason, `${prefix}.reason`)

  if (!coverageGapReasons.has(reason as CoverageGapReason)) {
    throw invalidPayload(
      `${prefix}.reason contains an unsupported value.`
    )
  }

  return {
    id: requireString(record.id, `${prefix}.id`),
    trajectory_id: requireString(
      record.trajectory_id,
      `${prefix}.trajectory_id`
    ),
    previous_segment_id: requireString(
      record.previous_segment_id,
      `${prefix}.previous_segment_id`,
      true
    ),
    next_segment_id: requireString(
      record.next_segment_id,
      `${prefix}.next_segment_id`,
      true
    ),
    icao24: requireString(
      record.icao24,
      `${prefix}.icao24`
    ).toLowerCase(),
    start_time: requireTimestamp(
      record.start_time,
      `${prefix}.start_time`
    ),
    end_time: requireTimestamp(
      record.end_time,
      `${prefix}.end_time`
    ),
    duration_seconds: requireNonNegativeNumber(
      record.duration_seconds,
      `${prefix}.duration_seconds`
    ),
    distance_km: requireNonNegativeNumber(
      record.distance_km,
      `${prefix}.distance_km`
    ),
    reason: reason as CoverageGapReason,
    filled_by: requireString(
      record.filled_by,
      `${prefix}.filled_by`,
      true
    ),
    created_at: requireTimestamp(
      record.created_at,
      `${prefix}.created_at`
    ),
  }
}

function requireRecord(
  value: unknown,
  fieldName: string
): Record<string, unknown> {
  if (
    typeof value !== 'object' ||
    value === null ||
    Array.isArray(value)
  ) {
    throw invalidPayload(`${fieldName} must be an object.`)
  }

  return value as Record<string, unknown>
}

function requireArray(
  value: unknown,
  fieldName: string
): unknown[] {
  if (!Array.isArray(value)) {
    throw invalidPayload(`${fieldName} must be an array.`)
  }

  return value
}

function requireString(
  value: unknown,
  fieldName: string,
  allowEmpty = false
): string {
  if (typeof value !== 'string') {
    throw invalidPayload(`${fieldName} must be a string.`)
  }

  if (!allowEmpty && value.trim() === '') {
    throw invalidPayload(`${fieldName} must not be empty.`)
  }

  return value
}

function requireTimestamp(
  value: unknown,
  fieldName: string
): string {
  const timestamp = requireString(value, fieldName)

  if (Number.isNaN(Date.parse(timestamp))) {
    throw invalidPayload(
      `${fieldName} must be a valid timestamp.`
    )
  }

  return timestamp
}

function requireFiniteNumber(
  value: unknown,
  fieldName: string
): number {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    throw invalidPayload(
      `${fieldName} must be a finite number.`
    )
  }

  return value
}

function requireNonNegativeNumber(
  value: unknown,
  fieldName: string
): number {
  const number = requireFiniteNumber(value, fieldName)

  if (number < 0) {
    throw invalidPayload(
      `${fieldName} must not be negative.`
    )
  }

  return number
}

function requireNonNegativeInteger(
  value: unknown,
  fieldName: string
): number {
  const number = requireNonNegativeNumber(value, fieldName)

  if (!Number.isInteger(number)) {
    throw invalidPayload(
      `${fieldName} must be an integer.`
    )
  }

  return number
}

function requireLatitude(
  value: unknown,
  fieldName: string
): number {
  const latitude = requireFiniteNumber(value, fieldName)

  if (latitude < -90 || latitude > 90) {
    throw invalidPayload(
      `${fieldName} must be between -90 and 90.`
    )
  }

  return latitude
}

function requireLongitude(
  value: unknown,
  fieldName: string
): number {
  const longitude = requireFiniteNumber(value, fieldName)

  if (longitude < -180 || longitude > 180) {
    throw invalidPayload(
      `${fieldName} must be between -180 and 180.`
    )
  }

  return longitude
}

function invalidPayload(message: string): APIRequestError {
  return new APIRequestError(
    `The trajectory response is invalid: ${message}`
  )
}

// STAGE-14-1-TRAJECTORY-RUNTIME-PARSER-FIX
