import type { TrafficAircraft } from '../../types/traffic'

export type AircraftExplorerSort =
  | 'recent'
  | 'callsign'
  | 'altitude-descending'
  | 'speed-descending'

export type AircraftExplorerMotionFilter = 'all' | 'airborne' | 'ground'

export interface AircraftExplorerOptions {
  query?: string
  sort?: AircraftExplorerSort
  motion?: AircraftExplorerMotionFilter
  requireAltitudeEvidence?: boolean
  limit?: number
}

export interface AircraftExplorerModel {
  items: TrafficAircraft[]
  totalCount: number
  matchedCount: number
  displayedCount: number
  airborneCount: number
  groundCount: number
  unknownAltitudeCount: number
}

const defaultLimit = 100

export function buildAircraftExplorerModel(
  aircraft: TrafficAircraft[],
  options: AircraftExplorerOptions = {}
): AircraftExplorerModel {
  const query = normalizeSearchValue(options.query ?? '')
  const sort = options.sort ?? 'recent'
  const motion = options.motion ?? 'all'
  const requireAltitudeEvidence = options.requireAltitudeEvidence ?? false
  const limit = normalizeLimit(options.limit)

  const matched = aircraft.filter(
    item =>
      matchesQuery(item, query) &&
      matchesMotion(item, motion) &&
      matchesAltitudeEvidence(item, requireAltitudeEvidence)
  )
  const sorted = [...matched].sort((left, right) =>
    compareAircraft(left, right, sort)
  )

  return {
    items: sorted.slice(0, limit),
    totalCount: aircraft.length,
    matchedCount: sorted.length,
    displayedCount: Math.min(sorted.length, limit),
    airborneCount: sorted.filter(item => !item.on_ground).length,
    groundCount: sorted.filter(item => item.on_ground).length,
    unknownAltitudeCount: sorted.filter(item => item.altitude_m === null).length,
  }
}

function normalizeLimit(value: number | undefined): number {
  if (value === undefined) return defaultLimit
  if (!Number.isSafeInteger(value) || value <= 0) return defaultLimit
  return value
}

function normalizeSearchValue(value: string): string {
  return value.trim().toLocaleLowerCase('en-US')
}

function matchesQuery(item: TrafficAircraft, query: string): boolean {
  if (!query) return true

  return [
    item.icao24,
    item.callsign,
    item.aircraft_model,
    item.airline,
    item.origin_country,
  ].some(value => normalizeSearchValue(value).includes(query))
}

function matchesMotion(
  item: TrafficAircraft,
  motion: AircraftExplorerMotionFilter
): boolean {
  if (motion === 'airborne') return !item.on_ground
  if (motion === 'ground') return item.on_ground
  return true
}

function matchesAltitudeEvidence(
  item: TrafficAircraft,
  requireAltitudeEvidence: boolean
): boolean {
  if (!requireAltitudeEvidence) return true
  if (item.on_ground) return item.altitude_status === 'ground'

  return (
    item.altitude_status === 'observed' &&
    item.altitude_m !== null &&
    Number.isFinite(item.altitude_m)
  )
}

function compareAircraft(
  left: TrafficAircraft,
  right: TrafficAircraft,
  sort: AircraftExplorerSort
): number {
  let result = 0

  switch (sort) {
    case 'callsign':
      result = compareText(displayIdentifier(left), displayIdentifier(right))
      break
    case 'altitude-descending':
      result = compareNullableNumberDescending(left.altitude_m, right.altitude_m)
      break
    case 'speed-descending':
      result = compareNumberDescending(left.velocity_mps, right.velocity_mps)
      break
    case 'recent':
      result = compareTimestampDescending(left.observed_at, right.observed_at)
      break
  }

  if (result !== 0) return result
  return compareText(left.icao24, right.icao24)
}

function displayIdentifier(item: TrafficAircraft): string {
  return item.callsign.trim() || item.icao24
}

function compareText(left: string, right: string): number {
  return left.localeCompare(right, 'en', {
    numeric: true,
    sensitivity: 'base',
  })
}

function compareNumberDescending(left: number, right: number): number {
  return right - left
}

function compareNullableNumberDescending(
  left: number | null,
  right: number | null
): number {
  if (left === null && right === null) return 0
  if (left === null) return 1
  if (right === null) return -1
  return compareNumberDescending(left, right)
}

function compareTimestampDescending(left: string, right: string): number {
  const leftTimestamp = parseTimestamp(left)
  const rightTimestamp = parseTimestamp(right)

  if (leftTimestamp === null && rightTimestamp === null) return 0
  if (leftTimestamp === null) return 1
  if (rightTimestamp === null) return -1
  return rightTimestamp - leftTimestamp
}

function parseTimestamp(value: string): number | null {
  const timestamp = Date.parse(value)
  return Number.isFinite(timestamp) ? timestamp : null
}
