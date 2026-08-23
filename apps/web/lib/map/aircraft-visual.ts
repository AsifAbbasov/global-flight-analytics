import type { TrafficAircraft } from '../../types/traffic'

export type AircraftMotionState = 'airborne' | 'ground'

export interface AircraftVisualState {
  key: string
  label: string
  headingDegrees: number
  motionState: AircraftMotionState
  isSelected: boolean
  latitude: number
  longitude: number
}

export function buildAircraftVisualState(
  aircraft: TrafficAircraft,
  selectedAircraftICAO24: string | null
): AircraftVisualState | null {
  if (!hasValidCoordinates(aircraft)) {
    return null
  }

  const key = aircraft.icao24.trim().toLowerCase()
  if (!key) {
    return null
  }

  return {
    key,
    label: aircraft.callsign.trim() || aircraft.icao24,
    headingDegrees: normalizeAircraftHeading(aircraft.heading_degrees),
    motionState: aircraft.on_ground ? 'ground' : 'airborne',
    isSelected:
      key === (selectedAircraftICAO24?.trim().toLowerCase() ?? null),
    latitude: aircraft.latitude,
    longitude: aircraft.longitude,
  }
}

export function normalizeAircraftHeading(headingDegrees: number): number {
  if (!Number.isFinite(headingDegrees)) {
    return 0
  }

  return ((headingDegrees % 360) + 360) % 360
}

function hasValidCoordinates(aircraft: TrafficAircraft): boolean {
  return (
    Number.isFinite(aircraft.latitude) &&
    aircraft.latitude >= -90 &&
    aircraft.latitude <= 90 &&
    Number.isFinite(aircraft.longitude) &&
    aircraft.longitude >= -180 &&
    aircraft.longitude <= 180
  )
}
