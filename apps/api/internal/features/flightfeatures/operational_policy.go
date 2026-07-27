package flightfeatures

// OperationalAggregationPolicy identifies how point observations contribute to
// operational aggregate metrics.
type OperationalAggregationPolicy string

// OperationalAltitudeSourcePolicy identifies how one altitude source is chosen
// for a trajectory-level altitude series.
type OperationalAltitudeSourcePolicy string

// OperationalHeadingPolicy identifies how chronological heading evidence is
// converted into cumulative heading change.
type OperationalHeadingPolicy string

const (
	CurrentOperationalAggregationPolicy    OperationalAggregationPolicy    = "observation-weighted-kahan-v1"
	CurrentOperationalAltitudeSourcePolicy OperationalAltitudeSourcePolicy = "single-source-prefer-barometric-v1"
	CurrentOperationalHeadingPolicy        OperationalHeadingPolicy        = "chronological-shortest-arc-contiguous-valid-runs-v1"
)

const (
	OperationalLimitationPointEvidenceUnavailable           = "operational_point_evidence_unavailable"
	OperationalLimitationPointEvidenceUnusable              = "operational_point_evidence_unusable"
	OperationalLimitationTemporalWindowUnavailable          = "operational_temporal_window_unavailable"
	OperationalLimitationPointTimestampMissing              = "operational_point_timestamp_missing"
	OperationalLimitationPointOutsideWindow                 = "operational_point_outside_window"
	OperationalLimitationPointOrderNonmonotonic             = "operational_point_order_nonmonotonic"
	OperationalLimitationDuplicatePointTimestamp            = "operational_duplicate_point_timestamp"
	OperationalLimitationAltitudeUnavailable                = "operational_altitude_unavailable"
	OperationalLimitationInvalidAltitudeObservations        = "operational_invalid_altitude_observations"
	OperationalLimitationGeometricAltitudeFallback          = "operational_geometric_altitude_fallback"
	OperationalLimitationMixedAltitudeSourceExcluded        = "operational_mixed_altitude_source_excluded"
	OperationalLimitationVelocityUnavailable                = "operational_velocity_unavailable"
	OperationalLimitationVelocityMeasurementUnavailable     = "operational_velocity_measurement_unavailable"
	OperationalLimitationInvalidVelocityObservations        = "operational_invalid_velocity_observations"
	OperationalLimitationVerticalRateUnavailable            = "operational_vertical_rate_unavailable"
	OperationalLimitationVerticalRateMeasurementUnavailable = "operational_vertical_rate_measurement_unavailable"
	OperationalLimitationInvalidVerticalRateObservations    = "operational_invalid_vertical_rate_observations"
	OperationalLimitationHeadingUnavailable                 = "operational_heading_unavailable"
	OperationalLimitationHeadingMeasurementUnavailable      = "operational_heading_measurement_unavailable"
	OperationalLimitationInvalidHeadingObservations         = "operational_invalid_heading_observations"
	OperationalLimitationHeadingSequenceGap                 = "operational_heading_sequence_gap"
	OperationalLimitationOnGroundUnavailable                = "operational_on_ground_unavailable"
	OperationalLimitationOnGroundMeasurementUnavailable     = "operational_on_ground_measurement_unavailable"
	OperationalLimitationAggregateNonFinite                 = "operational_aggregate_non_finite"
)
