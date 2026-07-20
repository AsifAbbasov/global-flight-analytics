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
  on_ground: boolean
  observed_at: string
  aircraft_model: string
  airline: string
  origin_country: string
}
