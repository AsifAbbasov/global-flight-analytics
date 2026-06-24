export interface TrafficAircraft {
  icao24: string
  callsign: string
  latitude: number
  longitude: number
  altitude_m: number
  velocity_mps: number
  heading_degrees: number
  on_ground: boolean
  observed_at: string
  aircraft_model: string
  airline: string
  origin_country: string
}
