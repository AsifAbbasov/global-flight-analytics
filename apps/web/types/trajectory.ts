export type TrajectorySegmentStatus =
  | 'observed'
  | 'interpolated'
  | 'estimated'
  | 'invalid'

export type CoverageGapReason =
  | 'time_gap'
  | 'movement_jump'
  | 'unknown'

export interface TrajectorySegment {
  id: string
  trajectory_id: string
  flight_id: string
  aircraft_id: string
  icao24: string
  callsign: string
  sequence_number: number
  status: TrajectorySegmentStatus
  quality_score: number
  start_time: string
  end_time: string
  duration_seconds: number
  start_latitude: number
  start_longitude: number
  end_latitude: number
  end_longitude: number
  point_count: number
  source_name: string
  created_at: string
}

export interface CoverageGap {
  id: string
  trajectory_id: string
  previous_segment_id: string
  next_segment_id: string
  icao24: string
  start_time: string
  end_time: string
  duration_seconds: number
  distance_km: number
  reason: CoverageGapReason
  filled_by: string
  created_at: string
}

export interface AircraftTrajectory {
  id: string
  flight_id: string
  aircraft_id: string
  icao24: string
  callsign: string
  start_time: string
  end_time: string
  duration_seconds: number
  segment_count: number
  point_count: number
  coverage_gap_count: number
  quality_score: number
  source_name: string
  segments: TrajectorySegment[]
  coverage_gaps: CoverageGap[]
  created_at: string
  updated_at: string
}
