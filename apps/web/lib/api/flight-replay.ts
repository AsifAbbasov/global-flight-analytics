import {
  APIRequestError,
  requestAPIData,
  type APIRequestOptions,
} from '@/lib/api/client'
import type { AircraftTrajectory } from '@/types/trajectory'
import type {
  FlightReplay,
  FlightReplayAltitudeStatus,
  FlightReplayPoint,
} from '@/types/flight-replay'

const altitudeStatuses = new Set<FlightReplayAltitudeStatus>([
  'observed',
  'ground',
  'unknown',
  'unavailable',
  'invalid',
])

export async function getTrajectoryFlightReplay(
  trajectory: AircraftTrajectory,
  options: APIRequestOptions = {}
): Promise<FlightReplay> {
  const flightID = trajectory.flight_id.trim()
  if (flightID === '') {
    throw new APIRequestError(
      'Historical replay requires a durable flight identifier.'
    )
  }

  const data = await requestAPIData<unknown>(
    `/api/v1/flights/${encodeURIComponent(flightID)}/states`,
    options
  )
  const points = parseReplayPoints(data, trajectory)

  return {
    trajectory_id: trajectory.id,
    flight_id: flightID,
    icao24: trajectory.icao24.trim().toLowerCase(),
    callsign: trajectory.callsign.trim(),
    start_time: trajectory.start_time,
    end_time: trajectory.end_time,
    evidence_class: 'observed',
    interpolation_policy: 'none',
    points,
  }
}

function parseReplayPoints(
  value: unknown,
  trajectory: AircraftTrajectory
): FlightReplayPoint[] {
  if (!Array.isArray(value)) {
    throw invalidPayload('flight states must be an array.')
  }

  const startTime = Date.parse(trajectory.start_time)
  const endTime = Date.parse(trajectory.end_time)
  if (!Number.isFinite(startTime) || !Number.isFinite(endTime) || endTime < startTime) {
    throw invalidPayload('trajectory replay window is invalid.')
  }

  const expectedFlightID = trajectory.flight_id.trim()
  const expectedICAO24 = trajectory.icao24.trim().toLowerCase()
  const seen = new Set<string>()
  const points: FlightReplayPoint[] = []

  for (let index = 0; index < value.length; index++) {
    const point = parseReplayPoint(value[index], index)
    const observedAt = Date.parse(point.observed_at)

    if (point.flight_id !== expectedFlightID) continue
    if (point.icao24 !== expectedICAO24) continue
    if (observedAt < startTime || observedAt > endTime) continue
    if (seen.has(point.id)) continue

    seen.add(point.id)
    points.push(point)
  }

  return points.sort((left, right) => {
    const timeDifference = Date.parse(left.observed_at) - Date.parse(right.observed_at)
    return timeDifference === 0 ? left.id.localeCompare(right.id) : timeDifference
  })
}

function parseReplayPoint(value: unknown, index: number): FlightReplayPoint {
  const field = `flight_states[${index}]`
  const record = requireRecord(value, field)
  const barometricStatus = requireAltitudeStatus(
    record.barometric_altitude_status,
    `${field}.barometric_altitude_status`
  )
  const geometricStatus = requireAltitudeStatus(
    record.geometric_altitude_status,
    `${field}.geometric_altitude_status`
  )

  return {
    id: requireString(record.id, `${field}.id`),
    flight_id: requireString(record.flight_id, `${field}.flight_id`),
    icao24: requireICAO24(record.icao24, `${field}.icao24`),
    callsign: requireString(record.callsign, `${field}.callsign`, true),
    latitude: requireLatitude(record.latitude, `${field}.latitude`),
    longitude: requireLongitude(record.longitude, `${field}.longitude`),
    barometric_altitude_m: requireNullableAltitude(
      record.barometric_altitude_m,
      barometricStatus,
      `${field}.barometric_altitude_m`
    ),
    barometric_altitude_status: barometricStatus,
    geometric_altitude_m: requireNullableAltitude(
      record.geometric_altitude_m,
      geometricStatus,
      `${field}.geometric_altitude_m`
    ),
    geometric_altitude_status: geometricStatus,
    observed_at: requireTimestamp(record.observed_at, `${field}.observed_at`),
    source_name: requireString(record.source_name, `${field}.source_name`, true),
  }
}

function requireRecord(
  value: unknown,
  fieldName: string
): Record<string, unknown> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    throw invalidPayload(`${fieldName} must be an object.`)
  }
  return value as Record<string, unknown>
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
  return value.trim()
}

function requireICAO24(value: unknown, fieldName: string): string {
  const normalized = requireString(value, fieldName).toLowerCase()
  if (!/^[0-9a-f]{6}$/.test(normalized)) {
    throw invalidPayload(`${fieldName} must contain six hexadecimal characters.`)
  }
  return normalized
}

function requireTimestamp(value: unknown, fieldName: string): string {
  const timestamp = requireString(value, fieldName)
  if (Number.isNaN(Date.parse(timestamp))) {
    throw invalidPayload(`${fieldName} must be a valid timestamp.`)
  }
  return timestamp
}

function requireAltitudeStatus(
  value: unknown,
  fieldName: string
): FlightReplayAltitudeStatus {
  const status = requireString(value, fieldName) as FlightReplayAltitudeStatus
  if (!altitudeStatuses.has(status)) {
    throw invalidPayload(`${fieldName} contains an unsupported value.`)
  }
  return status
}

function requireNullableAltitude(
  value: unknown,
  status: FlightReplayAltitudeStatus,
  fieldName: string
): number | null {
  if (value === null) return null
  if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
    throw invalidPayload(`${fieldName} must be null or a non-negative number.`)
  }
  if (status !== 'observed' && status !== 'ground') {
    throw invalidPayload(`${fieldName} cannot carry a value for status ${status}.`)
  }
  return value
}

function requireLatitude(value: unknown, fieldName: string): number {
  if (typeof value !== 'number' || !Number.isFinite(value) || value < -90 || value > 90) {
    throw invalidPayload(`${fieldName} must be between -90 and 90.`)
  }
  return value
}

function requireLongitude(value: unknown, fieldName: string): number {
  if (typeof value !== 'number' || !Number.isFinite(value) || value < -180 || value > 180) {
    throw invalidPayload(`${fieldName} must be between -180 and 180.`)
  }
  return value
}

function invalidPayload(message: string): APIRequestError {
  return new APIRequestError(`The flight replay response is invalid: ${message}`)
}
