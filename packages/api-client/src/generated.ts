/* eslint-disable */
// This file is generated from openapi/openapi.json. Do not edit manually.
// OpenAPI SHA-256: 3a4c8765e3f91af66c3d894e826615af1de08c0fd413a293ffdd45e45f1a9022

export type ActiveAircraftMetric = {
  readonly metric: "active_aircraft"
  readonly value: number
  readonly window_minutes: number
  readonly scope: MetricScope
  readonly observed_from: string
  readonly observed_to: string
  readonly calculated_at: string
  readonly confidence: MetricConfidence
  readonly sources: ReadonlyArray<MetricSource>
  readonly limitations: ReadonlyArray<string>
}

export type ActiveAircraftMetricResponse = {
  readonly success: true
  readonly data: ActiveAircraftMetric
}

export type AircraftListItem = {
  readonly icao24: string
  readonly registration: string
  readonly model: string
  readonly manufacturer: string
  readonly airline: string
  readonly country: string
}

export type AircraftListResponse = {
  readonly success: true
  readonly data: ReadonlyArray<AircraftListItem>
}

export type AircraftProfile = {
  readonly icao24: string
  readonly registration: string
  readonly model: string
  readonly manufacturer: string
  readonly aircraft_type: string
  readonly airline: string
  readonly country: string
}

export type AircraftResponse = {
  readonly success: true
  readonly data: AircraftProfile
}

export type AircraftRouteContext = {
  readonly icao24: string
  readonly trajectory_id: string
  readonly origin?: RouteContextAirportCandidate
  readonly destination?: RouteContextAirportCandidate
  readonly confidence: RouteContextConfidence
  readonly limitations: ReadonlyArray<RouteContextNotice>
  readonly generated_at: string
}

export type AircraftRouteContextResponse = {
  readonly success: true
  readonly data: AircraftRouteContext
}

export type AirportIntelligenceHistory = {
  readonly version: string
  readonly window: AirportIntelligenceWindow
  readonly icao_code: string
  readonly entries: ReadonlyArray<AirportStatistics>
  readonly limitations: ReadonlyArray<AirportIntelligenceLimitation>
  readonly generated_at: string
}

export type AirportIntelligenceHistoryResponse = {
  readonly success: true
  readonly data: AirportIntelligenceHistory
}

export type AirportIntelligenceLimitation = {
  readonly code: string
  readonly message: string
}

export type AirportIntelligenceOverview = {
  readonly version: string
  readonly window: AirportIntelligenceWindow
  readonly passport: AirportPassport
  readonly statistics: AirportStatistics
  readonly ranking: AirportRankingSummary
  readonly limitations: ReadonlyArray<AirportIntelligenceLimitation>
  readonly generated_at: string
}

export type AirportIntelligenceOverviewResponse = {
  readonly success: true
  readonly data: AirportIntelligenceOverview
}

export type AirportIntelligenceRanking = {
  readonly version: string
  readonly window: AirportIntelligenceWindow
  readonly weights: AirportRankingWeights
  readonly airports: ReadonlyArray<AirportRankedItem>
  readonly limitations: ReadonlyArray<AirportIntelligenceLimitation>
  readonly generated_at: string
}

export type AirportIntelligenceRankingResponse = {
  readonly success: true
  readonly data: AirportIntelligenceRanking
}

export type AirportIntelligenceTrends = {
  readonly version: string
  readonly window: AirportIntelligenceWindow
  readonly icao_code: string
  readonly compared_windows: number
  readonly window_duration_seconds: number
  readonly direction: string
  readonly baseline: AirportTrendPoint
  readonly current: AirportTrendPoint
  readonly peak: AirportTrendPoint
  readonly total_movements_change: number
  readonly movements_per_hour_change: number
  readonly movements_per_hour_change_percent: number
  readonly movements_per_hour_change_percent_known: boolean
  readonly active_routes_change: number
  readonly coverage_score_change: number
  readonly freshness_score_change: number
  readonly gap_count: number
  readonly gap_duration_seconds: number
  readonly observed_duration_seconds: number
  readonly continuity_score: number
  readonly limitations: ReadonlyArray<AirportIntelligenceLimitation>
  readonly generated_at: string
}

export type AirportIntelligenceTrendsResponse = {
  readonly success: true
  readonly data: AirportIntelligenceTrends
}

export type AirportIntelligenceWindow = {
  readonly start_time: string
  readonly end_time: string
  readonly as_of_time: string
  readonly completed_days: number
}

export type AirportListItem = {
  readonly icao_code: string
  readonly iata_code: string
  readonly name: string
  readonly city: string
  readonly country: string
  readonly latitude: number
  readonly longitude: number
}

export type AirportPassport = {
  readonly identity: AirportPassportIdentity
  readonly location: AirportPassportLocation
  readonly operations: AirportPassportOperations
  readonly data_quality: AirportPassportDataQuality
  readonly description: string
  readonly generated_at: string
}

export type AirportPassportDataQuality = {
  readonly freshness_score: number
  readonly coverage_score: number
  readonly observed_at: string
}

export type AirportPassportIdentity = {
  readonly icao_code: string
  readonly iata_code: string
  readonly name: string
}

export type AirportPassportLocation = {
  readonly city: string
  readonly country: string
  readonly latitude: number
  readonly longitude: number
  readonly elevation_m: number | null
  readonly elevation_status: string
  readonly timezone: string
}

export type AirportPassportOperations = {
  readonly arrivals: number
  readonly departures: number
  readonly activity: number
  readonly active_aircraft: number
}

export type AirportProfile = {
  readonly icao_code: string
  readonly iata_code: string
  readonly name: string
  readonly city: string
  readonly country: string
  readonly latitude: number
  readonly longitude: number
  readonly elevation_m: number | null
  readonly elevation_status: "observed" | "unknown" | "invalid"
  readonly timezone: string
  readonly description: string
}

export type AirportRankedItem = {
  readonly position: number
  readonly icao_code: string
  readonly iata_code: string
  readonly name: string
  readonly city: string
  readonly country: string
  readonly activity_score: number
  readonly data_confidence: number
  readonly movements_component: number
  readonly routes_component: number
  readonly observations_component: number
  readonly intensity_component: number
  readonly coverage_score: number
  readonly freshness_score: number
  readonly total_movements: number
  readonly active_routes: number
  readonly observed_samples: number
  readonly expected_samples: number
  readonly movements_per_hour: number
  readonly active_aircraft: number
}

export type AirportRankingSummary = {
  readonly position: number
  readonly total_airports: number
  readonly activity_score: number
  readonly data_confidence: number
  readonly movements_component: number
  readonly routes_component: number
  readonly observations_component: number
  readonly intensity_component: number
}

export type AirportRankingWeights = {
  readonly movements: number
  readonly routes: number
  readonly observations: number
  readonly intensity: number
  readonly coverage: number
  readonly freshness: number
}

export type AirportResponse = {
  readonly success: true
  readonly data: AirportProfile
}

export type AirportsResponse = {
  readonly success: true
  readonly data: ReadonlyArray<AirportListItem>
}

export type AirportStatistics = {
  readonly icao_code: string
  readonly window_start: string
  readonly window_end: string
  readonly arrivals: number
  readonly departures: number
  readonly total_movements: number
  readonly arrival_share: number
  readonly departure_share: number
  readonly movements_per_hour: number
  readonly active_aircraft: number
  readonly active_routes: number
  readonly observed_samples: number
  readonly expected_samples: number
  readonly coverage_score: number
  readonly freshness_score: number
  readonly latest_observation_at: string
  readonly generated_at: string
}

export type AirportTrendPoint = {
  readonly window_start: string
  readonly window_end: string
  readonly total_movements: number
  readonly movements_per_hour: number
  readonly active_routes: number
  readonly coverage_score: number
  readonly freshness_score: number
}

export type AirspaceConfidence = {
  readonly score: number
  readonly level: string
  readonly components: ReadonlyArray<AirspaceScoreComponent>
  readonly reasons: ReadonlyArray<AirspaceConfidenceReason>
}

export type AirspaceConfidenceReason = {
  readonly code: string
  readonly message: string
  readonly contribution: number
}

export type AirspaceExplanation = {
  readonly code: string
  readonly message: string
}

export type AirspaceLimitation = {
  readonly code: string
  readonly message: string
  readonly scope: string
}

export type AirspaceOccupancy = {
  readonly bucket_duration_seconds: number
  readonly latitude_cell_degrees: number
  readonly longitude_cell_degrees: number
  readonly altitude_band_meters: number
  readonly buckets: ReadonlyArray<AirspaceOccupancyBucket>
  readonly metrics: AirspaceOccupancyMetrics
}

export type AirspaceOccupancyBucket = {
  readonly id: string
  readonly start_time: string
  readonly end_time: string
  readonly cells: ReadonlyArray<AirspaceOccupancyCell>
  readonly metrics: AirspaceOccupancyBucketMetrics
}

export type AirspaceOccupancyBucketMetrics = {
  readonly aircraft_count: number
  readonly occupied_cell_count: number
  readonly unknown_altitude_count: number
  readonly mean_quality_score: number
}

export type AirspaceOccupancyCell = {
  readonly id: string
  readonly bucket_id: string
  readonly bucket_start: string
  readonly bucket_end: string
  readonly latitude_index: number
  readonly longitude_index: number
  readonly altitude_band_index: number
  readonly altitude_known: boolean
  readonly aircraft_node_ids: ReadonlyArray<string>
  readonly aircraft_count: number
  readonly mean_quality_score: number
}

export type AirspaceOccupancyMetrics = {
  readonly bucket_count: number
  readonly expected_bucket_count: number
  readonly occupied_cell_count: number
  readonly aircraft_observation_count: number
  readonly unique_aircraft_count: number
  readonly unknown_altitude_count: number
  readonly peak_aircraft_per_bucket: number
  readonly peak_occupied_cells: number
  readonly mean_aircraft_per_bucket: number
  readonly temporal_coverage: number
}

export type AirspaceProvenance = {
  readonly input_fingerprint: string
  readonly scene_fingerprints: ReadonlyArray<string>
  readonly scan_fingerprints: ReadonlyArray<string>
  readonly risk_fingerprints: ReadonlyArray<string>
  readonly source_names: ReadonlyArray<string>
  readonly latest_observed_at: string
}

export type AirspaceRegionAnalytics = {
  readonly version: string
  readonly schema_version: string
  readonly status: string
  readonly region_code: string
  readonly window_start: string
  readonly window_end: string
  readonly occupancy: AirspaceOccupancy
  readonly sector_complexity: ReadonlyArray<AirspaceSectorComplexity>
  readonly metrics: AirspaceRegionMetrics
  readonly confidence: AirspaceConfidence
  readonly limitations: ReadonlyArray<AirspaceLimitation>
  readonly explanations: ReadonlyArray<AirspaceExplanation>
  readonly scope_guard: string
  readonly provenance: AirspaceProvenance
  readonly generated_at: string
}

export type AirspaceRegionAnalyticsResponse = {
  readonly success: true
  readonly data: AirspaceRegionAnalytics
}

export type AirspaceRegionMetrics = {
  readonly snapshot_count: number
  readonly bucket_count: number
  readonly unique_aircraft_count: number
  readonly aircraft_observation_count: number
  readonly occupied_cell_count: number
  readonly sector_report_count: number
  readonly current_aircraft_count: number
  readonly peak_aircraft_per_bucket: number
  readonly mean_aircraft_per_bucket: number
  readonly mean_complexity_score: number
  readonly peak_complexity_score: number
  readonly airspace_pressure_index: number
  readonly peak_airspace_pressure_index: number
  readonly moderate_sector_count: number
  readonly high_sector_count: number
  readonly severe_sector_count: number
  readonly contextual_risk_count: number
  readonly elevated_risk_count: number
  readonly high_risk_count: number
  readonly indeterminate_risk_count: number
  readonly unknown_altitude_count: number
  readonly temporal_coverage: number
  readonly occupancy_trend: string
  readonly highest_complexity_level: string
}

export type AirspaceScoreComponent = {
  readonly name: string
  readonly score: number
  readonly weight: number
}

export type AirspaceSectorComplexity = {
  readonly id: string
  readonly bucket_id: string
  readonly bucket_start: string
  readonly bucket_end: string
  readonly latitude_index: number
  readonly longitude_index: number
  readonly aircraft_node_ids: ReadonlyArray<string>
  readonly aircraft_count: number
  readonly altitude_band_count: number
  readonly unknown_altitude_count: number
  readonly candidate_pair_count: number
  readonly converging_pair_count: number
  readonly contextual_risk_count: number
  readonly elevated_risk_count: number
  readonly high_risk_count: number
  readonly indeterminate_risk_count: number
  readonly heading_dispersion: number
  readonly speed_variability: number
  readonly score: number
  readonly level: string
  readonly components: ReadonlyArray<AirspaceScoreComponent>
  readonly confidence: AirspaceConfidence
  readonly limitations: ReadonlyArray<AirspaceLimitation>
  readonly explanations: ReadonlyArray<AirspaceExplanation>
}

export type AnalyticalConfidence = {
  readonly level: string
  readonly score: number
  readonly reasons: ReadonlyArray<AnalyticalNotice>
}

export type AnalyticalConfidenceFactor = {
  readonly code: string
  readonly kind: string
  readonly weight: number
  readonly value: number
  readonly impact: number
  readonly message: string
}

export type AnalyticalConfidenceReport = {
  readonly base_score: number
  readonly penalty_score: number
  readonly score: number
  readonly level: string
  readonly factors: ReadonlyArray<AnalyticalConfidenceFactor>
  readonly reasons: ReadonlyArray<AnalyticalNotice>
  readonly warnings: ReadonlyArray<AnalyticalNotice>
  readonly limitations: ReadonlyArray<AnalyticalNotice>
  readonly evaluated_at: string
}

export type AnalyticalEligibility = {
  readonly capability: string
  readonly allowed: boolean
  readonly reasons: ReadonlyArray<string>
  readonly evaluated_at: string
}

export type AnalyticalFailure = {
  readonly code: string
  readonly message: string
  readonly retriable: boolean
}

export type AnalyticalMetric = {
  readonly metric: string
  readonly status: string
  readonly value?: number | null
  readonly has_value: boolean
  readonly confidence: AnalyticalConfidence
  readonly data_quality?: DataQualityReport
  readonly eligibility?: AnalyticalEligibility
  readonly scope: AnalyticalScope
  readonly sources: ReadonlyArray<AnalyticalSource>
  readonly warnings: ReadonlyArray<AnalyticalNotice>
  readonly limitations: ReadonlyArray<AnalyticalNotice>
  readonly calculated_at: string
  readonly failure?: AnalyticalFailure
  readonly confidence_report?: AnalyticalConfidenceReport
}

export type AnalyticalMetricResponse = {
  readonly success: true
  readonly data: AnalyticalMetric
}

export type AnalyticalNotice = {
  readonly code: string
  readonly message: string
}

export type AnalyticalScope = {
  readonly capability: string
  readonly input_count: number
  readonly allowed_count: number
  readonly denied_count: number
  readonly reasons: ReadonlyArray<AnalyticalScopeReason>
  readonly evaluated_at: string
}

export type AnalyticalScopeReason = {
  readonly reason: string
  readonly count: number
}

export type AnalyticalSource = {
  readonly name: string
  readonly role: string
  readonly observed_from?: string
  readonly observed_to?: string
  readonly retrieved_at?: string
  readonly limitations: ReadonlyArray<AnalyticalNotice>
}

export type CoverageGap = {
  readonly id: string
  readonly trajectory_id: string
  readonly previous_segment_id: string
  readonly next_segment_id: string
  readonly icao24: string
  readonly start_time: string
  readonly end_time: string
  readonly duration_seconds: number
  readonly distance_km: number
  readonly reason: string
  readonly filled_by: string
  readonly created_at: string
}

export type CurrentTrafficItem = {
  readonly icao24: string
  readonly callsign: string
  readonly latitude: number
  readonly longitude: number
  readonly altitude_m: number | null
  readonly altitude_status: "observed" | "ground" | "unknown" | "unavailable" | "invalid"
  readonly altitude_source: "geometric" | "barometric" | "ground" | "none"
  readonly velocity_mps: number
  readonly heading_degrees: number
  readonly on_ground: boolean
  readonly observed_at: string
  readonly aircraft_model: string
  readonly airline: string
  readonly origin_country: string
}

export type CurrentTrafficResponse = {
  readonly success: true
  readonly data: ReadonlyArray<CurrentTrafficItem>
}

export type CurrentWeather = {
  readonly snapshot_id: string
  readonly provider: string
  readonly latitude: number
  readonly longitude: number
  readonly observed_at: string
  readonly retrieved_at: string
  readonly stored_at: string
  readonly temperature_celsius: number | null
  readonly relative_humidity_percent: number | null
  readonly precipitation_mm: number | null
  readonly rain_mm: number | null
  readonly weather_code: number | null
  readonly cloud_cover_percent: number | null
  readonly surface_pressure_hpa: number | null
  readonly wind_speed_mps: number | null
  readonly wind_direction_degrees: number | null
  readonly wind_gusts_mps: number | null
}

export type CurrentWeatherResponse = {
  readonly success: true
  readonly data: CurrentWeather
}

export type DataQualityAnalyticsPermissions = {
  readonly route_inference: DataQualityPermission
  readonly phase_detection: DataQualityPermission
  readonly historical_analytics: DataQualityPermission
  readonly historical_similarity: DataQualityPermission
  readonly projection: DataQualityPermission
}

export type DataQualityFreshness = {
  readonly score: number
  readonly status: "fresh" | "aging" | "stale" | "unknown"
  readonly age_seconds: number
  readonly expected_interval_seconds: number
  readonly stale_after_seconds: number
  readonly observed_at: string
  readonly evaluated_at: string
  readonly explanation: string
}

export type DataQualityNotice = {
  readonly code: string
  readonly message: string
}

export type DataQualityPermission = {
  readonly allowed: boolean
  readonly reasons: ReadonlyArray<string>
}

export type DataQualityProvenance = {
  readonly source_name: string
  readonly source_record_time: string
  readonly received_at: string
  readonly ingestion_run_id: string
  readonly transformation: string
  readonly algorithm_version: string
  readonly input_fingerprint: string
}

export type DataQualityReport = {
  readonly contract_version: string
  readonly provenance: DataQualityProvenance
  readonly freshness: DataQualityFreshness
  readonly sampling_density: DataQualitySamplingDensity
  readonly analytics_permissions: DataQualityAnalyticsPermissions
  readonly missing_fields: ReadonlyArray<string>
  readonly warnings: ReadonlyArray<DataQualityNotice>
  readonly limitations: ReadonlyArray<DataQualityNotice>
  readonly evaluated_at: string
}

export type DataQualitySamplingDensity = {
  readonly score: number
  readonly observed_sample_count: number
  readonly expected_sample_count: number
  readonly covered_interval_count: number
  readonly total_interval_count: number
  readonly duplicate_sample_count: number
  readonly window_start: string
  readonly window_end: string
  readonly expected_interval: number
  readonly explanation: string
}

export type ErrorBody = {
  readonly code: string
  readonly message: string
}

export type ErrorResponse = {
  readonly success: false
  readonly error: ErrorBody
}

export type FlightListItem = {
  readonly id: string
  readonly aircraft_id: string
  readonly icao24: string
  readonly callsign: string
  readonly status: string
  readonly first_seen_at: string
  readonly last_seen_at: string
  readonly aircraft_model: string
  readonly airline: string
  readonly country: string
}

export type FlightListResponse = {
  readonly success: true
  readonly data: ReadonlyArray<FlightListItem>
}

export type FlightProfile = {
  readonly id: string
  readonly aircraft_id: string
  readonly icao24: string
  readonly callsign: string
  readonly status: string
  readonly first_seen_at: string
  readonly last_seen_at: string
  readonly aircraft_model: string
  readonly airline: string
  readonly country: string
}

export type FlightResponse = {
  readonly success: true
  readonly data: FlightProfile
}

export type FlightStateItem = {
  readonly id: string
  readonly flight_id: string
  readonly aircraft_id: string
  readonly icao24: string
  readonly callsign: string
  readonly latitude: number
  readonly longitude: number
  readonly barometric_altitude_m: number | null
  readonly barometric_altitude_status: "observed" | "ground" | "unknown" | "unavailable" | "invalid"
  readonly geometric_altitude_m: number | null
  readonly geometric_altitude_status: "observed" | "ground" | "unknown" | "unavailable" | "invalid"
  readonly velocity_mps: number
  readonly heading_degrees: number
  readonly vertical_rate_mps: number
  readonly on_ground: boolean
  readonly origin_country: string
  readonly observed_at: string
  readonly source_name: string
}

export type FlightStateListResponse = {
  readonly success: true
  readonly data: ReadonlyArray<FlightStateItem>
}

export type FlightStateResponse = {
  readonly success: true
  readonly data: FlightStateItem
}

export type HealthData = {
  readonly status: "ok"
}

export type HealthResponse = {
  readonly success: true
  readonly data: HealthData
}

export type HistoricalAggregateHistory = {
  readonly items: ReadonlyArray<HistoricalAggregateRecord>
  readonly has_more: boolean
  readonly next_cursor?: string
}

export type HistoricalAggregateHistoryResponse = {
  readonly success: true
  readonly data: HistoricalAggregateHistory
}

export type HistoricalAggregateRecord = {
  readonly id: string
  readonly input_fingerprint: string
  readonly stored_at: string
  readonly result: HistoricalResult
}

export type HistoricalAggregateRecordResponse = {
  readonly success: true
  readonly data: HistoricalAggregateRecord
}

export type HistoricalComparison = {
  readonly previous_window: HistoricalTimeWindow
  readonly current_value: number
  readonly previous_value: number
  readonly absolute_change: number
  readonly percentage_change?: number
  readonly direction: "unavailable" | "down" | "flat" | "up"
}

export type HistoricalConfidence = {
  readonly score: number
  readonly level: "none" | "low" | "medium" | "high"
  readonly sample_count: number
  readonly reasons: ReadonlyArray<IntelligenceConfidenceReason>
}

export type HistoricalMetric = {
  readonly name: "active_aircraft" | "flight_count" | "trajectory_count" | "observation_count" | "traffic_density" | "airport_departures" | "airport_arrivals" | "airport_operations" | "unique_aircraft" | "active_routes" | "route_observations" | "route_confidence" | "complete_route_ratio" | "partial_route_ratio" | "unavailable_route_ratio" | "great_circle_distance_km"
  readonly unit: string
  readonly aggregation: "count" | "sum" | "minimum" | "maximum" | "average" | "median" | "ratio"
}

export type HistoricalPoint = {
  readonly start_time: string
  readonly end_time: string
  readonly status: "unavailable" | "partial" | "complete"
  readonly value: number
  readonly sample_count: number
  readonly coverage_ratio: number
  readonly confidence: HistoricalConfidence
  readonly limitations: ReadonlyArray<IntelligenceLimitation>
}

export type HistoricalProvenance = {
  readonly builder_version: string
  readonly input_fingerprint: string
  readonly source_names: ReadonlyArray<string>
  readonly latest_source_updated_at: string
}

export type HistoricalResult = {
  readonly schema_version: string
  readonly status: "unavailable" | "partial" | "complete"
  readonly metric: HistoricalMetric
  readonly scope: HistoricalScope
  readonly window: HistoricalTimeWindow
  readonly granularity: "hour" | "day" | "week" | "custom"
  readonly points: ReadonlyArray<HistoricalPoint>
  readonly summary: HistoricalSummary
  readonly comparison?: HistoricalComparison
  readonly confidence: HistoricalConfidence
  readonly limitations: ReadonlyArray<IntelligenceLimitation>
  readonly provenance: HistoricalProvenance
  readonly generated_at: string
}

export type HistoricalScope = {
  readonly type: "global" | "region" | "airport" | "route"
  readonly region_code?: string
  readonly airport_icao_code?: string
  readonly origin_icao_code?: string
  readonly destination_icao_code?: string
}

export type HistoricalSummary = {
  readonly point_count: number
  readonly total: number
  readonly minimum: number
  readonly maximum: number
  readonly average: number
  readonly median: number
}

export type HistoricalTimeWindow = {
  readonly start_time: string
  readonly end_time: string
  readonly as_of_time: string
}

export type IntelligenceConfidence = {
  readonly score: number
  readonly level: "none" | "low" | "medium" | "high"
  readonly reasons: ReadonlyArray<IntelligenceConfidenceReason>
}

export type IntelligenceConfidenceReason = {
  readonly code: string
  readonly message: string
  readonly contribution: number
}

export type IntelligenceLimitation = {
  readonly code: string
  readonly message: string
  readonly scope: string
}

export type IntelligenceNotice = {
  readonly code: string
  readonly message: string
}

export type MetricConfidence = {
  readonly level: "high" | "medium" | "low" | "none"
  readonly score: number
  readonly reasons: ReadonlyArray<string>
}

export type MetricScope = {
  readonly type: "global" | "region"
  readonly code: string
}

export type MetricSource = {
  readonly name: string
  readonly role: string
}

export type OpenObject = {
  readonly [key: string]: unknown
}

export type ProjectionArrival = {
  readonly airport_icao_code: string
  readonly earliest_time: string
  readonly estimated_time: string
  readonly latest_time: string
  readonly confidence: IntelligenceConfidence
  readonly limitations: ReadonlyArray<IntelligenceLimitation>
}

export type ProjectionEvidence = {
  readonly neighbor_selection?: OpenObject
  readonly pattern_confidence?: OpenObject
  readonly freshness?: OpenObject
  readonly route_frequency?: OpenObject
}

export type ProjectionHorizon = {
  readonly as_of_time: string
  readonly end_time: string
  readonly step_seconds: number
  readonly duration_seconds: number
}

export type ProjectionInput = {
  readonly name: string
  readonly classification: string
  readonly source_name: string
  readonly observed_at: string
  readonly retrieved_at: string
  readonly limitation?: string
}

export type ProjectionIntelligence = {
  readonly version: string
  readonly strategy: string
  readonly fallback_reason: string
  readonly arrival_status: string
  readonly projection: ProjectionResult
  readonly evidence: ProjectionEvidence
  readonly notices: ReadonlyArray<IntelligenceNotice>
  readonly input_fingerprint: string
  readonly generated_at: string
}

export type ProjectionIntelligenceResponse = {
  readonly success: true
  readonly data: ProjectionIntelligence
}

export type ProjectionMethod = {
  readonly name: string
  readonly version: string
  readonly decision_class: string
}

export type ProjectionPoint = {
  readonly sequence: number
  readonly forecast_time: string
  readonly position: ProjectionPosition
  readonly uncertainty: ProjectionUncertainty
  readonly confidence: IntelligenceConfidence
}

export type ProjectionPosition = {
  readonly latitude: number
  readonly longitude: number
  readonly altitude_m?: number
}

export type ProjectionProvenance = {
  readonly input_fingerprint: string
  readonly inputs: ReadonlyArray<ProjectionInput>
  readonly latest_input_observed_at: string
}

export type ProjectionResult = {
  readonly schema_version: string
  readonly status: string
  readonly trajectory_id: string
  readonly flight_id: string
  readonly aircraft_id: string
  readonly icao24: string
  readonly callsign: string
  readonly method: ProjectionMethod
  readonly horizon: ProjectionHorizon
  readonly points: ReadonlyArray<ProjectionPoint>
  readonly arrival?: ProjectionArrival
  readonly confidence: IntelligenceConfidence
  readonly limitations: ReadonlyArray<IntelligenceLimitation>
  readonly explanations: ReadonlyArray<IntelligenceNotice>
  readonly scope_guard: string
  readonly provenance: ProjectionProvenance
  readonly generated_at: string
}

export type ProjectionUncertainty = {
  readonly horizontal_radius_m: number
  readonly vertical_radius_m?: number
}

export type ReadinessData = {
  readonly status: "ready"
}

export type ReadinessResponse = {
  readonly success: true
  readonly data: ReadinessData
}

export type RegionBounds = {
  readonly min_latitude: number
  readonly max_latitude: number
  readonly min_longitude: number
  readonly max_longitude: number
}

export type RegionItem = {
  readonly code: string
  readonly name: string
  readonly description: string
  readonly bounds: RegionBounds
}

export type RegionResponse = {
  readonly success: true
  readonly data: RegionItem
}

export type RegionsResponse = {
  readonly success: true
  readonly data: ReadonlyArray<RegionItem>
}

export type RouteContextAirport = {
  readonly icao_code: string
  readonly iata_code: string
  readonly name: string
  readonly city: string
  readonly country: string
  readonly latitude: number
  readonly longitude: number
  readonly elevation_m: number | null
  readonly elevation_status: string
  readonly timezone: string
  readonly description: string
}

export type RouteContextAirportCandidate = {
  readonly airport: RouteContextAirport
  readonly distance_km: number
  readonly confidence: RouteContextConfidence
}

export type RouteContextConfidence = {
  readonly score: number
  readonly level: string
  readonly reasons: ReadonlyArray<RouteContextNotice>
}

export type RouteContextNotice = {
  readonly code: string
  readonly message: string
}

export type RouteIntelligenceAirport = {
  readonly icao_code: string
  readonly iata_code: string
  readonly name: string
  readonly city: string
  readonly country: string
  readonly latitude: number
  readonly longitude: number
  readonly elevation_m: number | null
  readonly elevation_status: "observed" | "unknown" | "invalid"
  readonly timezone: string
}

export type RouteIntelligenceConfidence = {
  readonly score: number
  readonly level: "none" | "low" | "medium" | "high"
  readonly evidence_count: number
  readonly reasons: ReadonlyArray<RouteIntelligenceConfidenceReason>
}

export type RouteIntelligenceConfidenceReason = {
  readonly code: string
  readonly message: string
  readonly contribution: number
}

export type RouteIntelligenceEndpoint = {
  readonly role: "origin" | "destination"
  readonly airport: RouteIntelligenceAirport
  readonly distance_km: number
  readonly confidence: RouteIntelligenceConfidence
  readonly evidence: ReadonlyArray<RouteIntelligenceEvidence>
  readonly limitations: ReadonlyArray<RouteIntelligenceLimitation>
}

export type RouteIntelligenceEvidence = {
  readonly type: "trajectory_endpoint_proximity" | "ground_cycle" | "callsign_route_token" | "source_flight_identity" | "airport_activity" | "external_reference"
  readonly source_name: string
  readonly source_version: string
  readonly score: number
  readonly weight: number
  readonly observed_at: string
  readonly summary: string
  readonly attributes: ReadonlyArray<RouteIntelligenceEvidenceAttribute>
}

export type RouteIntelligenceEvidenceAttribute = {
  readonly key: string
  readonly value: string
}

export type RouteIntelligenceHistory = {
  readonly items: ReadonlyArray<RouteIntelligenceRecord>
  readonly has_more: boolean
  readonly next_before_as_of_time?: string
}

export type RouteIntelligenceHistoryResponse = {
  readonly success: true
  readonly data: RouteIntelligenceHistory
}

export type RouteIntelligenceLimitation = {
  readonly code: string
  readonly message: string
  readonly scope: string
}

export type RouteIntelligenceProvenance = {
  readonly resolver_version: string
  readonly input_fingerprint: string
  readonly trajectory_updated_at: string
  readonly source_names: ReadonlyArray<string>
}

export type RouteIntelligenceRecord = {
  readonly id: string
  readonly input_fingerprint: string
  readonly stored_at: string
  readonly result: RouteIntelligenceResult
}

export type RouteIntelligenceRecordResponse = {
  readonly success: true
  readonly data: RouteIntelligenceRecord
}

export type RouteIntelligenceResult = {
  readonly schema_version: "route-intelligence-v1"
  readonly status: "unavailable" | "partial" | "complete"
  readonly trajectory_id: string
  readonly identity_key: string
  readonly flight_id: string
  readonly aircraft_id: string
  readonly icao24: string
  readonly callsign: string
  readonly window: RouteIntelligenceWindow
  readonly origin?: RouteIntelligenceEndpoint
  readonly destination?: RouteIntelligenceEndpoint
  readonly summary: RouteIntelligenceSummary
  readonly confidence: RouteIntelligenceConfidence
  readonly limitations: ReadonlyArray<RouteIntelligenceLimitation>
  readonly provenance: RouteIntelligenceProvenance
  readonly generated_at: string
}

export type RouteIntelligenceSummary = {
  readonly great_circle_distance_km: number
  readonly same_airport: boolean
}

export type RouteIntelligenceWindow = {
  readonly start_time: string
  readonly end_time: string
  readonly as_of_time: string
}

export type StabilityAnalysis = {
  readonly status: string
  readonly trend: string
  readonly health: string
  readonly metrics: StabilityAnalysisMetrics
  readonly confidence_score: number
  readonly confidence_level: string
  readonly input_fingerprint: string
}

export type StabilityAnalysisMetrics = {
  readonly version_count: number
  readonly transition_count: number
  readonly comparable_transition_count: number
  readonly stable_transition_share: number
  readonly comparable_transition_share: number
  readonly material_change_share: number
  readonly mean_stability_score: number
  readonly minimum_stability_score: number
  readonly score_standard_deviation: number
  readonly longest_stable_run: number
  readonly method_change_count: number
  readonly policy_change_count: number
  readonly implementation_change_count: number
  readonly input_change_count: number
  readonly output_change_count: number
  readonly mean_horizontal_shift_kilometers: number
  readonly maximum_horizontal_shift_kilometers: number
  readonly latest_level: string
}

export type StabilityConfidenceSummary = {
  readonly status: string
  readonly score: number
  readonly level: string
  readonly target_node_id: string
  readonly limiting_dependency_id?: string
  readonly input_fingerprint: string
}

export type StabilityFailure = {
  readonly rank: number
  readonly code: string
  readonly category: string
  readonly severity: string
  readonly classification: string
  readonly summary: string
  readonly detail: string
  readonly source: string
  readonly blocks_use: boolean
  readonly priority_score: number
  readonly evidence_fingerprints: ReadonlyArray<string>
}

export type StabilityFailureExplanation = {
  readonly status: string
  readonly primary_code: string
  readonly blocking_count: number
  readonly warning_count: number
  readonly unknown_cause_count: number
  readonly confidence_score: number
  readonly confidence_level: string
  readonly failures: ReadonlyArray<StabilityFailure>
  readonly input_fingerprint: string
}

export type StabilityIntelligence = {
  readonly version: string
  readonly trajectory_id: string
  readonly as_of_times: ReadonlyArray<string>
  readonly projections: ReadonlyArray<ProjectionIntelligence>
  readonly forecast_versions: ReadonlyArray<StabilityVersion>
  readonly transitions: ReadonlyArray<StabilityTransition>
  readonly forecast_analysis: StabilityAnalysis
  readonly propagated_confidence: StabilityConfidenceSummary
  readonly failure_explanation: StabilityFailureExplanation
  readonly unknown_intervention: StabilityInterventionGuard
  readonly scope_enforcement: StabilityScopeEnforcement
  readonly scope_guards: ReadonlyArray<string>
  readonly input_fingerprint: string
  readonly generated_at: string
}

export type StabilityIntelligenceResponse = {
  readonly success: true
  readonly data: StabilityIntelligence
}

export type StabilityInterventionGuard = {
  readonly status: string
  readonly claim_kind: string
  readonly decision: string
  readonly confidence_score: number
  readonly evidence_count: number
  readonly unknown_evidence_count: number
  readonly estimated_evidence_count: number
  readonly evidence_completeness: number
  readonly input_fingerprint: string
}

export type StabilityScopeEnforcement = {
  readonly status: string
  readonly decision: string
  readonly claim_count: number
  readonly allowed_count: number
  readonly limited_count: number
  readonly blocked_count: number
  readonly violations: ReadonlyArray<StabilityScopeViolation>
  readonly input_fingerprint: string
}

export type StabilityScopeViolation = {
  readonly code: string
  readonly claim_code: string
  readonly message: string
  readonly blocking: boolean
}

export type StabilityTransition = {
  readonly baseline_version_id: string
  readonly candidate_version_id: string
  readonly level: string
  readonly score: number
  readonly metrics: StabilityTransitionMetrics
  readonly input_fingerprint: string
  readonly evaluated_at: string
}

export type StabilityTransitionMetrics = {
  readonly aligned_point_count: number
  readonly aligned_point_share: number
  readonly mean_horizontal_shift_kilometers: number
  readonly maximum_horizontal_shift_kilometers: number
  readonly aggregate_confidence_delta: number
  readonly mean_relative_horizontal_uncertainty_change: number
  readonly arrival_comparable: boolean
  readonly arrival_shift_seconds: number
  readonly method_changed: boolean
  readonly policy_changed: boolean
  readonly implementation_changed: boolean
  readonly input_changed: boolean
  readonly output_changed: boolean
}

export type StabilityVersion = {
  readonly version_id: string
  readonly ordinal: number
  readonly parent_version_id?: string
  readonly method_name: string
  readonly method_version: string
  readonly policy_version: string
  readonly implementation_version: string
  readonly input_fingerprint: string
  readonly output_fingerprint: string
  readonly decision_fingerprint: string
  readonly created_at: string
}

export type Trajectory = {
  readonly id: string
  readonly identity_key: string
  readonly identity_basis: string
  readonly split_reason: string
  readonly flight_id: string
  readonly aircraft_id: string
  readonly icao24: string
  readonly callsign: string
  readonly start_time: string
  readonly end_time: string
  readonly duration_seconds: number
  readonly segment_count: number
  readonly point_count: number
  readonly coverage_gap_count: number
  readonly quality_score: number
  readonly source_name: string
  readonly segments: ReadonlyArray<TrajectorySegment>
  readonly coverage_gaps: ReadonlyArray<CoverageGap>
  readonly created_at: string
  readonly updated_at: string
}

export type TrajectoryResponse = {
  readonly success: true
  readonly data: Trajectory
}

export type TrajectorySegment = {
  readonly id: string
  readonly trajectory_id: string
  readonly flight_id: string
  readonly aircraft_id: string
  readonly icao24: string
  readonly callsign: string
  readonly sequence_number: number
  readonly status: string
  readonly quality_score: number
  readonly start_time: string
  readonly end_time: string
  readonly duration_seconds: number
  readonly start_latitude: number
  readonly start_longitude: number
  readonly end_latitude: number
  readonly end_longitude: number
  readonly point_count: number
  readonly source_name: string
  readonly created_at: string
}

export type TransponderEvidence = {
  readonly schema_version: string
  readonly evidence_only: boolean
  readonly confirmed_emergency: boolean
  readonly operational_alert: boolean
  readonly aircraft: TransponderEvidenceAircraft
  readonly observed_transponder_code: string
  readonly classification: TransponderEvidenceClassification
  readonly observation: TransponderEvidenceObservation
  readonly freshness: TransponderEvidenceFreshness
  readonly confidence: TransponderEvidenceConfidence
  readonly provenance: TransponderEvidenceProvenance
  readonly maximum_claim_strength: string
  readonly limitations: ReadonlyArray<string>
}

export type TransponderEvidenceAircraft = {
  readonly icao24: string
  readonly callsign?: string
}

export type TransponderEvidenceClassification = {
  readonly kind: string
  readonly label: string
}

export type TransponderEvidenceConfidence = {
  readonly level: string
  readonly reasons: ReadonlyArray<string>
}

export type TransponderEvidenceFreshness = {
  readonly status: string
  readonly first_observed_at: string
  readonly last_observed_at: string
  readonly as_of_time: string
  readonly age_seconds: number
  readonly maximum_fresh_age_seconds: number
}

export type TransponderEvidenceObservation = {
  readonly strength: string
  readonly observation_count: number
  readonly special_purpose_indicator_observed: boolean
}

export type TransponderEvidenceProvenance = {
  readonly fingerprint: string
  readonly source_names: ReadonlyArray<string>
}

export type TransponderEvidenceResponse = {
  readonly success: true
  readonly data: TransponderEvidence
}

export type VersionData = {
  readonly version: string
  readonly revision: string
  readonly built_at: string
}

export type VersionResponse = {
  readonly success: true
  readonly data: VersionData
}

export type WeatherAlignmentResult = {
  readonly version: string
  readonly status: string
  readonly trajectory_id: string
  readonly as_of_time: string
  readonly trust_decision: string
  readonly trust_score: number
  readonly point_count: number
  readonly aligned_count: number
  readonly unmatched_count: number
  readonly coverage_ratio: number
  readonly matches: ReadonlyArray<OpenObject>
  readonly limitations: ReadonlyArray<IntelligenceNotice>
  readonly explanations: ReadonlyArray<IntelligenceNotice>
  readonly input_fingerprint: string
  readonly generated_at: string
}

export type WeatherContext = {
  readonly version: string
  readonly trajectory_id: string
  readonly as_of_time: string
  readonly weather: WeatherFeatureResult
  readonly trust: WeatherTrustResult
  readonly alignment: WeatherAlignmentResult
  readonly encounter: WeatherEncounterResult
  readonly uncertainty: WeatherUncertaintyResult
  readonly input_fingerprint: string
  readonly generated_at: string
}

export type WeatherContextResponse = {
  readonly success: true
  readonly data: WeatherContext
}

export type WeatherDirectionSummary = {
  readonly present_count: number
  readonly coverage_ratio: number
  readonly mean_direction_degrees?: number
  readonly concentration?: number
}

export type WeatherEncounterResult = {
  readonly version: string
  readonly status: string
  readonly trajectory_id: string
  readonly as_of_time: string
  readonly alignment_status: string
  readonly alignment_coverage_ratio: number
  readonly point_count: number
  readonly encounter_point_count: number
  readonly unprofiled_point_count: number
  readonly profile_coverage_ratio: number
  readonly temperature_celsius: WeatherMetricSummary
  readonly precipitation_millimeters: WeatherMetricSummary
  readonly cloud_cover_percent: WeatherMetricSummary
  readonly wind_speed_meters_per_second: WeatherMetricSummary
  readonly wind_gusts_meters_per_second: WeatherMetricSummary
  readonly wind_direction_degrees: WeatherDirectionSummary
  readonly limitations: ReadonlyArray<IntelligenceNotice>
  readonly explanations: ReadonlyArray<IntelligenceNotice>
  readonly input_fingerprint: string
  readonly generated_at: string
}

export type WeatherFeatureResult = {
  readonly schema_version: string
  readonly status: string
  readonly trajectory_id: string
  readonly as_of_time: string
  readonly samples: ReadonlyArray<WeatherSample>
  readonly confidence: IntelligenceConfidence
  readonly limitations: ReadonlyArray<IntelligenceLimitation>
  readonly explanations: ReadonlyArray<IntelligenceNotice>
  readonly scope_guard: string
  readonly input_fingerprint: string
  readonly source_names: ReadonlyArray<string>
  readonly latest_available_at: string
  readonly generated_at: string
}

export type WeatherMetricSummary = {
  readonly present_count: number
  readonly coverage_ratio: number
  readonly minimum?: number
  readonly maximum?: number
  readonly mean?: number
}

export type WeatherSample = {
  readonly sequence: number
  readonly position: WeatherSamplePosition
  readonly source: WeatherSampleSource
  readonly features: OpenObject
  readonly valid_at: string
  readonly available_at: string
  readonly retrieved_at: string
}

export type WeatherSamplePosition = {
  readonly latitude: number
  readonly longitude: number
  readonly altitude_meters?: number
  readonly vertical_reference: string
}

export type WeatherSampleSource = {
  readonly provider: string
  readonly dataset: string
  readonly evidence_kind: string
  readonly horizontal_resolution_kilometers?: number
  readonly temporal_resolution_seconds: number
}

export type WeatherTrustComponent = {
  readonly name: string
  readonly score: number
  readonly weight: number
}

export type WeatherTrustResult = {
  readonly version: string
  readonly decision: string
  readonly usable: boolean
  readonly as_of_time: string
  readonly score: number
  readonly components: ReadonlyArray<WeatherTrustComponent>
  readonly allowed_scopes: ReadonlyArray<string>
  readonly limitations: ReadonlyArray<IntelligenceNotice>
  readonly explanations: ReadonlyArray<IntelligenceNotice>
  readonly input_fingerprint: string
}

export type WeatherUncertaintyResult = {
  readonly version: string
  readonly status: string
  readonly trajectory_id: string
  readonly as_of_time: string
  readonly severity_score: number
  readonly weather_multiplier: number
  readonly point_adjustments: ReadonlyArray<OpenObject>
  readonly limitations: ReadonlyArray<IntelligenceNotice>
  readonly explanations: ReadonlyArray<IntelligenceNotice>
  readonly input_fingerprint: string
  readonly generated_at: string
}

export interface OperationDefinition {
  readonly method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  readonly path: string
  readonly protected: boolean
  readonly hasBody: boolean
  readonly parameters: ReadonlyArray<{
    readonly name: string
    readonly in: string
    readonly required: boolean
  }>
}

export interface OperationParameters {
  readonly getActiveAircraftMetric: {
    readonly query?: {
      readonly region?: string
      readonly window_minutes?: number
    }
  }
  readonly getAircraftByICAO24: {
    readonly path: {
      readonly icao24: string
    }
  }
  readonly getAircraftRouteContextByICAO24: {
    readonly path: {
      readonly icao24: string
    }
  }
  readonly getAirportByICAO: {
    readonly path: {
      readonly icao: string
    }
  }
  readonly getAirportIntelligenceHistory: {
    readonly path: {
      readonly icao: string
    }
    readonly query?: {
      readonly days?: number
      readonly as_of_time?: string
    }
  }
  readonly getAirportIntelligenceOverview: {
    readonly path: {
      readonly icao: string
    }
    readonly query?: {
      readonly days?: number
      readonly as_of_time?: string
    }
  }
  readonly getAirportIntelligenceRanking: {
    readonly query?: {
      readonly days?: number
      readonly as_of_time?: string
      readonly limit?: number
    }
  }
  readonly getAirportIntelligenceTrends: {
    readonly path: {
      readonly icao: string
    }
    readonly query?: {
      readonly days?: number
      readonly as_of_time?: string
    }
  }
  readonly getAirspaceRegionAnalytics: {
    readonly path: {
      readonly code: string
    }
    readonly query: {
      readonly as_of_time: string
      readonly window_seconds?: number
    }
  }
  readonly getAnalyticalActiveAircraft: {
    readonly query?: {
      readonly window_minutes?: number
      readonly limit?: number
      readonly region?: string
    }
  }
  readonly getAnalyticalAirportActivity: {
    readonly query: {
      readonly window_minutes?: number
      readonly limit?: number
      readonly airport_icao: string
      readonly radius_kilometers?: number
    }
  }
  readonly getAnalyticalCoverageScore: {
    readonly query?: {
      readonly window_minutes?: number
      readonly region?: string
    }
  }
  readonly getAnalyticalDataFreshness: {
    readonly query?: {
      readonly window_minutes?: number
      readonly region?: string
    }
  }
  readonly getAnalyticalTrafficDensity: {
    readonly query: {
      readonly window_minutes?: number
      readonly limit?: number
      readonly region: string
    }
  }
  readonly getCurrentTraffic: {
    readonly query?: {
      readonly region?: string
    }
  }
  readonly getCurrentWeather: {
    readonly query: {
      readonly lat: number
      readonly lon: number
    }
  }
  readonly getFlightByID: {
    readonly path: {
      readonly id: string
    }
  }
  readonly getHealth: {}
  readonly getLatestFlightStateByICAO24: {
    readonly path: {
      readonly icao24: string
    }
  }
  readonly getLatestHistoricalIntelligenceAggregate: {
    readonly query: {
      readonly metric: "active_aircraft" | "flight_count" | "trajectory_count" | "observation_count" | "traffic_density" | "airport_departures" | "airport_arrivals" | "airport_operations" | "unique_aircraft" | "active_routes" | "route_observations" | "route_confidence" | "complete_route_ratio" | "partial_route_ratio" | "unavailable_route_ratio" | "great_circle_distance_km"
      readonly scope: "global" | "region" | "airport" | "route"
      readonly granularity: "hour" | "day" | "week" | "custom"
      readonly region_code?: string
      readonly airport_icao?: string
      readonly origin_icao?: string
      readonly destination_icao?: string
    }
  }
  readonly getLatestRouteIntelligenceByTrajectoryID: {
    readonly path: {
      readonly id: string
    }
  }
  readonly getLatestTrajectoryByICAO24: {
    readonly path: {
      readonly icao24: string
    }
  }
  readonly getLatestTransponderEvidence: {
    readonly path: {
      readonly icao24: string
    }
  }
  readonly getProjectionIntelligenceByTrajectoryID: {
    readonly path: {
      readonly id: string
    }
    readonly query: {
      readonly as_of_time: string
      readonly duration_seconds?: number
    }
  }
  readonly getReadiness: {}
  readonly getRegionByCode: {
    readonly path: {
      readonly code: string
    }
  }
  readonly getStabilityIntelligenceByTrajectoryID: {
    readonly path: {
      readonly id: string
    }
    readonly query: {
      readonly as_of_times: string
      readonly duration_seconds: number
    }
  }
  readonly getTrajectoryByID: {
    readonly path: {
      readonly id: string
    }
  }
  readonly getVersion: {}
  readonly getWeatherContextByTrajectoryID: {
    readonly path: {
      readonly id: string
    }
    readonly query: {
      readonly as_of_time: string
      readonly duration_seconds: number
    }
  }
  readonly listAircraft: {}
  readonly listAirports: {}
  readonly listFlights: {}
  readonly listFlightStatesByFlightID: {
    readonly path: {
      readonly flightID: string
    }
  }
  readonly listHistoricalIntelligenceAggregateHistory: {
    readonly query: {
      readonly metric: "active_aircraft" | "flight_count" | "trajectory_count" | "observation_count" | "traffic_density" | "airport_departures" | "airport_arrivals" | "airport_operations" | "unique_aircraft" | "active_routes" | "route_observations" | "route_confidence" | "complete_route_ratio" | "partial_route_ratio" | "unavailable_route_ratio" | "great_circle_distance_km"
      readonly scope: "global" | "region" | "airport" | "route"
      readonly granularity: "hour" | "day" | "week" | "custom"
      readonly region_code?: string
      readonly airport_icao?: string
      readonly origin_icao?: string
      readonly destination_icao?: string
      readonly limit?: number
      readonly cursor?: string
    }
  }
  readonly listRegions: {}
  readonly listRouteIntelligenceHistoryByTrajectoryID: {
    readonly path: {
      readonly id: string
    }
    readonly query?: {
      readonly limit?: number
      readonly before_as_of_time?: string
    }
  }
  readonly processRouteIntelligenceByTrajectoryID: {
    readonly path: {
      readonly id: string
    }
  }
}

export interface OperationResponses {
  readonly getActiveAircraftMetric: ActiveAircraftMetricResponse
  readonly getAircraftByICAO24: AircraftResponse
  readonly getAircraftRouteContextByICAO24: AircraftRouteContextResponse
  readonly getAirportByICAO: AirportResponse
  readonly getAirportIntelligenceHistory: AirportIntelligenceHistoryResponse
  readonly getAirportIntelligenceOverview: AirportIntelligenceOverviewResponse
  readonly getAirportIntelligenceRanking: AirportIntelligenceRankingResponse
  readonly getAirportIntelligenceTrends: AirportIntelligenceTrendsResponse
  readonly getAirspaceRegionAnalytics: AirspaceRegionAnalyticsResponse
  readonly getAnalyticalActiveAircraft: AnalyticalMetricResponse
  readonly getAnalyticalAirportActivity: AnalyticalMetricResponse
  readonly getAnalyticalCoverageScore: AnalyticalMetricResponse
  readonly getAnalyticalDataFreshness: AnalyticalMetricResponse
  readonly getAnalyticalTrafficDensity: AnalyticalMetricResponse
  readonly getCurrentTraffic: CurrentTrafficResponse
  readonly getCurrentWeather: CurrentWeatherResponse
  readonly getFlightByID: FlightResponse
  readonly getHealth: HealthResponse
  readonly getLatestFlightStateByICAO24: FlightStateResponse
  readonly getLatestHistoricalIntelligenceAggregate: HistoricalAggregateRecordResponse
  readonly getLatestRouteIntelligenceByTrajectoryID: RouteIntelligenceRecordResponse
  readonly getLatestTrajectoryByICAO24: TrajectoryResponse
  readonly getLatestTransponderEvidence: TransponderEvidenceResponse
  readonly getProjectionIntelligenceByTrajectoryID: ProjectionIntelligenceResponse
  readonly getReadiness: ReadinessResponse
  readonly getRegionByCode: RegionResponse
  readonly getStabilityIntelligenceByTrajectoryID: StabilityIntelligenceResponse
  readonly getTrajectoryByID: TrajectoryResponse
  readonly getVersion: VersionResponse
  readonly getWeatherContextByTrajectoryID: WeatherContextResponse
  readonly listAircraft: AircraftListResponse
  readonly listAirports: AirportsResponse
  readonly listFlights: FlightListResponse
  readonly listFlightStatesByFlightID: FlightStateListResponse
  readonly listHistoricalIntelligenceAggregateHistory: HistoricalAggregateHistoryResponse
  readonly listRegions: RegionsResponse
  readonly listRouteIntelligenceHistoryByTrajectoryID: RouteIntelligenceHistoryResponse
  readonly processRouteIntelligenceByTrajectoryID: RouteIntelligenceRecordResponse
}

export type OperationId = keyof OperationParameters

export const operationDefinitions = {
  getActiveAircraftMetric: {
      method: "GET",
      path: "/api/v1/metrics/active-aircraft",
      protected: false,
      hasBody: false,
      parameters: [{"name":"region","in":"query","required":false},{"name":"window_minutes","in":"query","required":false}],
    },
  getAircraftByICAO24: {
      method: "GET",
      path: "/api/v1/aircraft/{icao24}",
      protected: false,
      hasBody: false,
      parameters: [{"name":"icao24","in":"path","required":true}],
    },
  getAircraftRouteContextByICAO24: {
      method: "GET",
      path: "/api/v1/aircraft/{icao24}/route-context",
      protected: false,
      hasBody: false,
      parameters: [{"name":"icao24","in":"path","required":true}],
    },
  getAirportByICAO: {
      method: "GET",
      path: "/api/v1/airports/{icao}",
      protected: false,
      hasBody: false,
      parameters: [{"name":"icao","in":"path","required":true}],
    },
  getAirportIntelligenceHistory: {
      method: "GET",
      path: "/api/v1/airports/{icao}/intelligence/history",
      protected: false,
      hasBody: false,
      parameters: [{"name":"icao","in":"path","required":true},{"name":"days","in":"query","required":false},{"name":"as_of_time","in":"query","required":false}],
    },
  getAirportIntelligenceOverview: {
      method: "GET",
      path: "/api/v1/airports/{icao}/intelligence/overview",
      protected: false,
      hasBody: false,
      parameters: [{"name":"icao","in":"path","required":true},{"name":"days","in":"query","required":false},{"name":"as_of_time","in":"query","required":false}],
    },
  getAirportIntelligenceRanking: {
      method: "GET",
      path: "/api/v1/airports/intelligence/ranking",
      protected: false,
      hasBody: false,
      parameters: [{"name":"days","in":"query","required":false},{"name":"as_of_time","in":"query","required":false},{"name":"limit","in":"query","required":false}],
    },
  getAirportIntelligenceTrends: {
      method: "GET",
      path: "/api/v1/airports/{icao}/intelligence/trends",
      protected: false,
      hasBody: false,
      parameters: [{"name":"icao","in":"path","required":true},{"name":"days","in":"query","required":false},{"name":"as_of_time","in":"query","required":false}],
    },
  getAirspaceRegionAnalytics: {
      method: "GET",
      path: "/api/v1/airspace/regions/{code}/analytics",
      protected: false,
      hasBody: false,
      parameters: [{"name":"code","in":"path","required":true},{"name":"as_of_time","in":"query","required":true},{"name":"window_seconds","in":"query","required":false}],
    },
  getAnalyticalActiveAircraft: {
      method: "GET",
      path: "/api/v1/analytics/metrics/active-aircraft",
      protected: false,
      hasBody: false,
      parameters: [{"name":"window_minutes","in":"query","required":false},{"name":"limit","in":"query","required":false},{"name":"region","in":"query","required":false}],
    },
  getAnalyticalAirportActivity: {
      method: "GET",
      path: "/api/v1/analytics/metrics/airport-activity",
      protected: false,
      hasBody: false,
      parameters: [{"name":"window_minutes","in":"query","required":false},{"name":"limit","in":"query","required":false},{"name":"airport_icao","in":"query","required":true},{"name":"radius_kilometers","in":"query","required":false}],
    },
  getAnalyticalCoverageScore: {
      method: "GET",
      path: "/api/v1/analytics/metrics/coverage-score",
      protected: false,
      hasBody: false,
      parameters: [{"name":"window_minutes","in":"query","required":false},{"name":"region","in":"query","required":false}],
    },
  getAnalyticalDataFreshness: {
      method: "GET",
      path: "/api/v1/analytics/metrics/data-freshness",
      protected: false,
      hasBody: false,
      parameters: [{"name":"window_minutes","in":"query","required":false},{"name":"region","in":"query","required":false}],
    },
  getAnalyticalTrafficDensity: {
      method: "GET",
      path: "/api/v1/analytics/metrics/traffic-density",
      protected: false,
      hasBody: false,
      parameters: [{"name":"window_minutes","in":"query","required":false},{"name":"limit","in":"query","required":false},{"name":"region","in":"query","required":true}],
    },
  getCurrentTraffic: {
      method: "GET",
      path: "/api/v1/traffic/current",
      protected: false,
      hasBody: false,
      parameters: [{"name":"region","in":"query","required":false}],
    },
  getCurrentWeather: {
      method: "GET",
      path: "/api/v1/weather/current",
      protected: false,
      hasBody: false,
      parameters: [{"name":"lat","in":"query","required":true},{"name":"lon","in":"query","required":true}],
    },
  getFlightByID: {
      method: "GET",
      path: "/api/v1/flights/{id}",
      protected: false,
      hasBody: false,
      parameters: [{"name":"id","in":"path","required":true}],
    },
  getHealth: {
      method: "GET",
      path: "/api/v1/health",
      protected: false,
      hasBody: false,
      parameters: [],
    },
  getLatestFlightStateByICAO24: {
      method: "GET",
      path: "/api/v1/aircraft/{icao24}/latest-state",
      protected: false,
      hasBody: false,
      parameters: [{"name":"icao24","in":"path","required":true}],
    },
  getLatestHistoricalIntelligenceAggregate: {
      method: "GET",
      path: "/api/v1/historical-intelligence/aggregates/latest",
      protected: false,
      hasBody: false,
      parameters: [{"name":"metric","in":"query","required":true},{"name":"scope","in":"query","required":true},{"name":"granularity","in":"query","required":true},{"name":"region_code","in":"query","required":false},{"name":"airport_icao","in":"query","required":false},{"name":"origin_icao","in":"query","required":false},{"name":"destination_icao","in":"query","required":false}],
    },
  getLatestRouteIntelligenceByTrajectoryID: {
      method: "GET",
      path: "/api/v1/trajectories/{id}/route-intelligence/latest",
      protected: false,
      hasBody: false,
      parameters: [{"name":"id","in":"path","required":true}],
    },
  getLatestTrajectoryByICAO24: {
      method: "GET",
      path: "/api/v1/aircraft/{icao24}/trajectory",
      protected: false,
      hasBody: false,
      parameters: [{"name":"icao24","in":"path","required":true}],
    },
  getLatestTransponderEvidence: {
      method: "GET",
      path: "/api/v1/aircraft/{icao24}/transponder-evidence/latest",
      protected: false,
      hasBody: false,
      parameters: [{"name":"icao24","in":"path","required":true}],
    },
  getProjectionIntelligenceByTrajectoryID: {
      method: "GET",
      path: "/api/v1/trajectories/{id}/projection-intelligence",
      protected: false,
      hasBody: false,
      parameters: [{"name":"id","in":"path","required":true},{"name":"as_of_time","in":"query","required":true},{"name":"duration_seconds","in":"query","required":false}],
    },
  getReadiness: {
      method: "GET",
      path: "/api/v1/ready",
      protected: false,
      hasBody: false,
      parameters: [],
    },
  getRegionByCode: {
      method: "GET",
      path: "/api/v1/regions/{code}",
      protected: false,
      hasBody: false,
      parameters: [{"name":"code","in":"path","required":true}],
    },
  getStabilityIntelligenceByTrajectoryID: {
      method: "GET",
      path: "/api/v1/trajectories/{id}/stability-intelligence",
      protected: false,
      hasBody: false,
      parameters: [{"name":"id","in":"path","required":true},{"name":"as_of_times","in":"query","required":true},{"name":"duration_seconds","in":"query","required":true}],
    },
  getTrajectoryByID: {
      method: "GET",
      path: "/api/v1/trajectories/{id}",
      protected: false,
      hasBody: false,
      parameters: [{"name":"id","in":"path","required":true}],
    },
  getVersion: {
      method: "GET",
      path: "/api/v1/version",
      protected: false,
      hasBody: false,
      parameters: [],
    },
  getWeatherContextByTrajectoryID: {
      method: "GET",
      path: "/api/v1/trajectories/{id}/weather-context",
      protected: false,
      hasBody: false,
      parameters: [{"name":"id","in":"path","required":true},{"name":"as_of_time","in":"query","required":true},{"name":"duration_seconds","in":"query","required":true}],
    },
  listAircraft: {
      method: "GET",
      path: "/api/v1/aircraft",
      protected: false,
      hasBody: false,
      parameters: [],
    },
  listAirports: {
      method: "GET",
      path: "/api/v1/airports",
      protected: false,
      hasBody: false,
      parameters: [],
    },
  listFlights: {
      method: "GET",
      path: "/api/v1/flights",
      protected: false,
      hasBody: false,
      parameters: [],
    },
  listFlightStatesByFlightID: {
      method: "GET",
      path: "/api/v1/flights/{flightID}/states",
      protected: false,
      hasBody: false,
      parameters: [{"name":"flightID","in":"path","required":true}],
    },
  listHistoricalIntelligenceAggregateHistory: {
      method: "GET",
      path: "/api/v1/historical-intelligence/aggregates/history",
      protected: false,
      hasBody: false,
      parameters: [{"name":"metric","in":"query","required":true},{"name":"scope","in":"query","required":true},{"name":"granularity","in":"query","required":true},{"name":"region_code","in":"query","required":false},{"name":"airport_icao","in":"query","required":false},{"name":"origin_icao","in":"query","required":false},{"name":"destination_icao","in":"query","required":false},{"name":"limit","in":"query","required":false},{"name":"cursor","in":"query","required":false}],
    },
  listRegions: {
      method: "GET",
      path: "/api/v1/regions",
      protected: false,
      hasBody: false,
      parameters: [],
    },
  listRouteIntelligenceHistoryByTrajectoryID: {
      method: "GET",
      path: "/api/v1/trajectories/{id}/route-intelligence/history",
      protected: false,
      hasBody: false,
      parameters: [{"name":"id","in":"path","required":true},{"name":"limit","in":"query","required":false},{"name":"before_as_of_time","in":"query","required":false}],
    },
  processRouteIntelligenceByTrajectoryID: {
      method: "POST",
      path: "/api/v1/trajectories/{id}/route-intelligence",
      protected: true,
      hasBody: false,
      parameters: [{"name":"id","in":"path","required":true}],
    }
} as const satisfies Record<OperationId, OperationDefinition>
