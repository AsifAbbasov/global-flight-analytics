export type FlightReplayAltitudeStatus =
  | 'observed'
  | 'ground'
  | 'unknown'
  | 'unavailable'
  | 'invalid'

export interface FlightReplayPoint {
  id: string
  flight_id: string
  icao24: string
  callsign: string
  latitude: number
  longitude: number
  barometric_altitude_m: number | null
  barometric_altitude_status: FlightReplayAltitudeStatus
  geometric_altitude_m: number | null
  geometric_altitude_status: FlightReplayAltitudeStatus
  observed_at: string
  source_name: string
}

export interface FlightReplay {
  trajectory_id: string
  flight_id: string
  icao24: string
  callsign: string
  start_time: string
  end_time: string
  evidence_class: 'observed'
  interpolation_policy: 'none'
  points: FlightReplayPoint[]
}
