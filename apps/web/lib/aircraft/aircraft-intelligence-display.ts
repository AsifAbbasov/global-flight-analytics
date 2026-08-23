import type { AircraftProfile } from '../../types/aircraft'
import type { TrafficAircraft } from '../../types/traffic'

export type AircraftIntelligenceEvidence = 'observed' | 'profile'

export interface AircraftIntelligenceField {
  key: string
  label: string
  value: string
  evidence: AircraftIntelligenceEvidence
}

export interface AircraftIntelligenceDisplay {
  title: string
  subtitle: string
  observedFields: AircraftIntelligenceField[]
  profileFields: AircraftIntelligenceField[]
}

export function buildAircraftIntelligenceDisplay(
  selectedICAO24: string,
  aircraft: TrafficAircraft | undefined,
  profile: AircraftProfile | undefined
): AircraftIntelligenceDisplay {
  const normalizedICAO24 = selectedICAO24.trim().toUpperCase()
  const title =
    clean(aircraft?.callsign) || clean(profile?.registration) || normalizedICAO24

  return {
    title,
    subtitle: `ICAO24 ${normalizedICAO24}`,
    observedFields: aircraft ? buildObservedFields(aircraft) : [],
    profileFields: profile ? buildProfileFields(profile) : [],
  }
}

function buildObservedFields(aircraft: TrafficAircraft): AircraftIntelligenceField[] {
  const fields: Array<AircraftIntelligenceField | null> = [
    field('altitude', 'Altitude', formatAltitude(aircraft), 'observed'),
    field('speed', 'Speed', formatSpeed(aircraft.velocity_mps), 'observed'),
    field('heading', 'Heading', formatHeading(aircraft.heading_degrees), 'observed'),
    field('status', 'Status', aircraft.on_ground ? 'On ground' : 'Airborne', 'observed'),
    field('callsign', 'Callsign', clean(aircraft.callsign), 'observed'),
    field('country', 'Origin country', clean(aircraft.origin_country), 'observed'),
    field(
      'coordinates',
      'Position',
      formatPosition(aircraft.latitude, aircraft.longitude),
      'observed'
    ),
    field('observed-at', 'Observed', formatTimestamp(aircraft.observed_at), 'observed'),
  ]

  return fields.filter((value): value is AircraftIntelligenceField => value !== null)
}

function buildProfileFields(profile: AircraftProfile): AircraftIntelligenceField[] {
  const fields: Array<AircraftIntelligenceField | null> = [
    field('registration', 'Registration', clean(profile.registration), 'profile'),
    field('type', 'Type', clean(profile.aircraft_type), 'profile'),
    field('manufacturer', 'Manufacturer', clean(profile.manufacturer), 'profile'),
    field('model', 'Model', clean(profile.model), 'profile'),
    field('airline', 'Airline', clean(profile.airline), 'profile'),
    field('registration-country', 'Registration country', clean(profile.country), 'profile'),
  ]

  return fields.filter((value): value is AircraftIntelligenceField => value !== null)
}

function field(
  key: string,
  label: string,
  value: string,
  evidence: AircraftIntelligenceEvidence
): AircraftIntelligenceField | null {
  if (!value) return null
  return { key, label, value, evidence }
}

function clean(value: string | undefined): string {
  return value?.trim() ?? ''
}

function formatAltitude(aircraft: TrafficAircraft): string {
  if (aircraft.altitude_status === 'ground') return 'Ground (0 m)'

  if (
    aircraft.altitude_status !== 'observed' ||
    aircraft.altitude_m === null ||
    !Number.isFinite(aircraft.altitude_m)
  ) {
    return ''
  }

  const rounded = Math.round(aircraft.altitude_m).toLocaleString('en-US')
  if (aircraft.altitude_source === 'geometric') return `${rounded} m (geometric)`
  if (aircraft.altitude_source === 'barometric') return `${rounded} m (barometric)`
  return `${rounded} m`
}

function formatSpeed(value: number): string {
  if (!Number.isFinite(value) || value < 0) return ''
  return `${Math.round(value * 3.6).toLocaleString('en-US')} km/h`
}

function formatHeading(value: number): string {
  if (!Number.isFinite(value)) return ''
  const normalized = ((value % 360) + 360) % 360
  return `${Math.round(normalized)}°`
}

function formatPosition(latitude: number, longitude: number): string {
  if (
    !Number.isFinite(latitude) ||
    latitude < -90 ||
    latitude > 90 ||
    !Number.isFinite(longitude) ||
    longitude < -180 ||
    longitude > 180
  ) {
    return ''
  }

  return `${latitude.toFixed(4)}, ${longitude.toFixed(4)}`
}

function formatTimestamp(value: string): string {
  const timestamp = new Date(value)
  return Number.isNaN(timestamp.getTime()) ? '' : timestamp.toISOString()
}
