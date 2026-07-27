package flightfeatures

// GeographicalDistanceModel identifies the deterministic distance algorithm
// used by the geographical feature builder.
type GeographicalDistanceModel string

// GeographicCellPolicy identifies the deterministic decimal-degree bucket
// construction used for unique geographic-cell counts.
type GeographicCellPolicy string

const (
	CurrentGeographicalDistanceModel GeographicalDistanceModel = "mean-earth-sphere-haversine-v1"
	CurrentGeographicCellPolicy      GeographicCellPolicy      = "decimal-degree-round-half-away-from-zero-v1"
)

const (
	LimitationTrajectoryPointCountMetadataMismatch = "trajectory_point_count_metadata_mismatch"

	GeographicalLimitationTemporalWindowUnavailable              = "geographical_temporal_window_unavailable"
	GeographicalLimitationInvalidPointCoordinates                = "geographical_invalid_point_coordinates"
	GeographicalLimitationPointTimestampMissing                  = "geographical_point_timestamp_missing"
	GeographicalLimitationPointOutsideWindow                     = "geographical_point_outside_window"
	GeographicalLimitationPointEvidenceUnavailable               = "geographical_point_evidence_unavailable"
	GeographicalLimitationPointEvidenceUnusable                  = "geographical_point_evidence_unusable"
	GeographicalLimitationSinglePointSegmentFallback             = "geographical_single_point_segment_fallback"
	GeographicalLimitationSegmentEndpointFallback                = "geographical_segment_endpoint_fallback"
	GeographicalLimitationInvalidSegmentCoordinates              = "geographical_invalid_segment_coordinates"
	GeographicalLimitationInvalidSegmentStatus                   = "geographical_invalid_segment_status"
	GeographicalLimitationSegmentTimestampMissing                = "geographical_segment_timestamp_missing"
	GeographicalLimitationSegmentOutsideWindow                   = "geographical_segment_outside_window"
	GeographicalLimitationSegmentDiscontinuityExcluded           = "geographical_segment_discontinuity_excluded"
	GeographicalLimitationSegmentSupportingPointCountUnavailable = "geographical_segment_supporting_point_count_unavailable"
	GeographicalLimitationSegmentEvidenceUnusable                = "geographical_segment_evidence_unusable"
	GeographicalLimitationCoordinatesUnavailable                 = "geographical_coordinates_unavailable"
	GeographicalLimitationSingleCoordinate                       = "geographical_single_coordinate"
)

// CircularLongitudeSpanDegrees returns the eastward circular span from the
// western envelope bound to the eastern envelope bound. A western value larger
// than the eastern value represents a valid antimeridian-wrapping envelope.
func CircularLongitudeSpanDegrees(west float64, east float64) float64 {
	if west <= east {
		return east - west
	}
	return 360 - (west - east)
}

// HasLimitationCode reports whether evidence contains the requested code.
func HasLimitationCode(
	limitations []FeatureLimitation,
	code string,
) bool {
	for _, limitation := range limitations {
		if limitation.Code == code {
			return true
		}
	}
	return false
}
