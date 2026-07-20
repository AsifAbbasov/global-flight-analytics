export type RouteConfidenceLevel =
  | 'none'
  | 'low'
  | 'medium'
  | 'high'

export interface RouteContextNotice {
  code: string
  message: string
}

export interface RouteContextConfidence {
  score: number
  level: RouteConfidenceLevel
  reasons: RouteContextNotice[]
}

export type AirportElevationStatus = 'observed' | 'unknown' | 'invalid'

export interface RouteContextAirport {
  icao_code: string
  iata_code: string
  name: string
  city: string
  country: string
  latitude: number
  longitude: number
  elevation_m: number | null
  elevation_status: AirportElevationStatus
  timezone: string
  description: string
}

export interface RouteContextAirportCandidate {
  airport: RouteContextAirport
  distance_km: number
  confidence: RouteContextConfidence
}

export interface AircraftRouteContext {
  icao24: string
  trajectory_id: string
  origin?: RouteContextAirportCandidate
  destination?: RouteContextAirportCandidate
  confidence: RouteContextConfidence
  limitations: RouteContextNotice[]
  generated_at: string
}
