import type {
  TrafficAircraft,
  TrafficAltitudeSource,
  TrafficAltitudeStatus,
} from '@/types/traffic'

type TrafficAltitudeView = Pick<
  TrafficAircraft,
  'altitude_m' | 'altitude_status' | 'altitude_source'
>

export function formatTrafficAltitude(
  altitude: TrafficAltitudeView
): string {
  if (altitude.altitude_status === 'ground') {
    return 'Ground (0 m)'
  }

  if (
    altitude.altitude_status === 'observed' &&
    altitude.altitude_m !== null &&
    Number.isFinite(altitude.altitude_m)
  ) {
    const source = formatAltitudeSource(altitude.altitude_source)
    const value = new Intl.NumberFormat().format(
      Math.round(altitude.altitude_m)
    )

    return source ? `${value} m (${source})` : `${value} m`
  }

  return formatAltitudeStatus(altitude.altitude_status)
}

function formatAltitudeSource(
  source: TrafficAltitudeSource
): string {
  if (source === 'geometric') {
    return 'geometric'
  }

  if (source === 'barometric') {
    return 'barometric'
  }

  return ''
}

function formatAltitudeStatus(
  status: TrafficAltitudeStatus
): string {
  if (status === 'unknown') {
    return 'Unknown'
  }

  if (status === 'unavailable') {
    return 'Unavailable'
  }

  if (status === 'invalid') {
    return 'Invalid altitude evidence'
  }

  return 'Unavailable'
}
