import {
  APIRequestError,
  requestAPIData,
  type APIRequestOptions,
} from '@/lib/api/client'
import type { AircraftProfile } from '@/types/aircraft'

export async function getAircraftProfile(
  icao24: string,
  options: APIRequestOptions = {}
): Promise<AircraftProfile> {
  const normalizedICAO24 = normalizeICAO24(icao24)
  const data = await requestAPIData<unknown>(
    `/api/v1/aircraft/${encodeURIComponent(normalizedICAO24)}`,
    options
  )

  return parseAircraftProfile(data)
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

function parseAircraftProfile(value: unknown): AircraftProfile {
  const record = requireRecord(value, 'aircraft profile')

  return {
    icao24: requireString(record.icao24, 'icao24').toLowerCase(),
    registration: requireString(
      record.registration,
      'registration',
      true
    ),
    model: requireString(record.model, 'model', true),
    manufacturer: requireString(
      record.manufacturer,
      'manufacturer',
      true
    ),
    aircraft_type: requireString(
      record.aircraft_type,
      'aircraft_type',
      true
    ),
    airline: requireString(record.airline, 'airline', true),
    country: requireString(record.country, 'country', true),
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

function invalidPayload(message: string): APIRequestError {
  return new APIRequestError(
    `The aircraft profile response is invalid: ${message}`
  )
}
