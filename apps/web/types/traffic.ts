export type TrafficAltitudeStatus =
  | 'observed'
  | 'ground'
  | 'unknown'
  | 'unavailable'
  | 'invalid'

export type TrafficAltitudeSource =
  | 'geometric'
  | 'barometric'
  | 'ground'
  | 'none'

export type TrafficPositionSource =
  | 'adsb'
  | 'adsr'
  | 'tisb'
  | 'adsc'
  | 'asterix'
  | 'mlat'
  | 'flarm'

export interface TrafficAircraft {
  icao24: string
  callsign: string
  latitude: number
  longitude: number
  altitude_m: number | null
  altitude_status: TrafficAltitudeStatus
  altitude_source: TrafficAltitudeSource
  velocity_mps: number
  heading_degrees: number
  vertical_rate_mps: number | null
  on_ground: boolean
  observed_at: string
  position_observed_at: string
  message_observed_at: string | null
  position_source?: TrafficPositionSource | null
  source_name?: string
  aircraft_model: string
  airline: string
  origin_country: string
}
