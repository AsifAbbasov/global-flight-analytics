import {
  APIRequestError,
  requestAPIData,
  type APIRequestOptions,
} from '@/lib/api/client'
import type {
  AircraftRouteContext,
  RouteConfidenceLevel,
  RouteContextAirport,
  RouteContextAirportCandidate,
  RouteContextConfidence,
  RouteContextNotice,
} from '@/types/route-context'

const confidenceLevels = new Set<RouteConfidenceLevel>([
  'none',
  'low',
  'medium',
  'high',
])

export async function getAircraftRouteContext(
  icao24: string,
  options: APIRequestOptions = {}
): Promise<AircraftRouteContext> {
  const normalizedICAO24 = normalizeICAO24(icao24)
  const data = await requestAPIData<unknown>(
    `/api/v1/aircraft/${encodeURIComponent(
      normalizedICAO24
    )}/route-context`,
    options
  )

  return parseAircraftRouteContext(data)
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

function parseAircraftRouteContext(
  value: unknown
): AircraftRouteContext {
  const record = requireRecord(value, 'route context')

  return {
    icao24: requireString(record.icao24, 'icao24').toLowerCase(),
    trajectory_id: requireString(
      record.trajectory_id,
      'trajectory_id'
    ),
    origin: parseOptionalCandidate(record.origin, 'origin'),
    destination: parseOptionalCandidate(
      record.destination,
      'destination'
    ),
    confidence: parseConfidence(
      record.confidence,
      'confidence'
    ),
    limitations: requireArray(
      record.limitations,
      'limitations'
    ).map((item, index) =>
      parseNotice(item, `limitations[${index}]`)
    ),
    generated_at: requireTimestamp(
      record.generated_at,
      'generated_at'
    ),
  }
}

function parseOptionalCandidate(
  value: unknown,
  fieldName: string
): RouteContextAirportCandidate | undefined {
  if (value === undefined || value === null) {
    return undefined
  }

  const record = requireRecord(value, fieldName)

  return {
    airport: parseAirport(
      record.airport,
      `${fieldName}.airport`
    ),
    distance_km: requireNonNegativeNumber(
      record.distance_km,
      `${fieldName}.distance_km`
    ),
    confidence: parseConfidence(
      record.confidence,
      `${fieldName}.confidence`
    ),
  }
}

function parseAirport(
  value: unknown,
  fieldName: string
): RouteContextAirport {
  const record = requireRecord(value, fieldName)

  return {
    icao_code: requireString(
      record.icao_code,
      `${fieldName}.icao_code`
    ),
    iata_code: requireString(
      record.iata_code,
      `${fieldName}.iata_code`,
      true
    ),
    name: requireString(record.name, `${fieldName}.name`),
    city: requireString(
      record.city,
      `${fieldName}.city`,
      true
    ),
    country: requireString(
      record.country,
      `${fieldName}.country`,
      true
    ),
    latitude: requireLatitude(
      record.latitude,
      `${fieldName}.latitude`
    ),
    longitude: requireLongitude(
      record.longitude,
      `${fieldName}.longitude`
    ),
    elevation_m: requireFiniteNumber(
      record.elevation_m,
      `${fieldName}.elevation_m`
    ),
    timezone: requireString(
      record.timezone,
      `${fieldName}.timezone`,
      true
    ),
    description: requireString(
      record.description,
      `${fieldName}.description`,
      true
    ),
  }
}

function parseConfidence(
  value: unknown,
  fieldName: string
): RouteContextConfidence {
  const record = requireRecord(value, fieldName)
  const score = requireFiniteNumber(
    record.score,
    `${fieldName}.score`
  )
  if (score < 0 || score > 1) {
    throw invalidPayload(
      `${fieldName}.score must be between zero and one.`
    )
  }

  const level = requireString(
    record.level,
    `${fieldName}.level`
  )
  if (!confidenceLevels.has(level as RouteConfidenceLevel)) {
    throw invalidPayload(
      `${fieldName}.level contains an unsupported value.`
    )
  }

  return {
    score,
    level: level as RouteConfidenceLevel,
    reasons: requireArray(
      record.reasons,
      `${fieldName}.reasons`
    ).map((item, index) =>
      parseNotice(item, `${fieldName}.reasons[${index}]`)
    ),
  }
}

function parseNotice(
  value: unknown,
  fieldName: string
): RouteContextNotice {
  const record = requireRecord(value, fieldName)

  return {
    code: requireString(record.code, `${fieldName}.code`),
    message: requireString(
      record.message,
      `${fieldName}.message`
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

function invalidPayload(message: string): APIRequestError {
  return new APIRequestError(
    `The route context response is invalid: ${message}`
  )
}
