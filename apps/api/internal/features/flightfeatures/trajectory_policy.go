package flightfeatures

// TrajectoryPointCountPolicy identifies how trajectory point count is selected
// when point records are materialized or only persisted metadata is available.
type TrajectoryPointCountPolicy string

// TrajectorySamplingPolicy identifies how observation instants contribute to
// sampling interval metrics.
type TrajectorySamplingPolicy string

// TrajectoryCoveragePolicy identifies the evidence and gap-union semantics used
// for temporal coverage.
type TrajectoryCoveragePolicy string

// TrajectoryPathPolicy identifies how continuous path parts are formed and how
// discontinuities contribute to path efficiency.
type TrajectoryPathPolicy string

const (
	CurrentTrajectoryPointCountPolicy TrajectoryPointCountPolicy = "canonical-unique-points-or-persisted-metadata-fallback-v1"
	CurrentTrajectorySamplingPolicy   TrajectorySamplingPolicy   = "unique-chronological-observation-instants-kahan-v1"
	CurrentTrajectoryCoveragePolicy   TrajectoryCoveragePolicy   = "non-invalid-segment-evidence-plus-clipped-gap-union-v1"
	CurrentTrajectoryPathPolicy       TrajectoryPathPolicy       = "continuous-parts-no-gap-or-segment-discontinuity-bridging-v1"
)

const (
	TrajectoryLimitationTemporalWindowUnavailable              = "trajectory_temporal_window_unavailable"
	TrajectoryLimitationPointRecordsUnmaterialized             = "trajectory_point_records_unmaterialized"
	TrajectoryLimitationPointTimestampMissing                  = "trajectory_point_timestamp_missing"
	TrajectoryLimitationPointOutsideWindow                     = "trajectory_point_outside_window"
	TrajectoryLimitationPointOrderCanonicalized                = "trajectory_point_order_canonicalized"
	TrajectoryLimitationDuplicateTimestampsCollapsed           = "trajectory_duplicate_timestamps_collapsed"
	TrajectoryLimitationInvalidPointCoordinates                = "trajectory_invalid_point_coordinates"
	TrajectoryLimitationSamplingEvidenceInsufficient           = "trajectory_sampling_evidence_insufficient"
	TrajectoryLimitationSamplingIntervalInvalid                = "trajectory_sampling_interval_invalid"
	TrajectoryLimitationSamplingAggregateNonFinite             = "trajectory_sampling_aggregate_non_finite"
	TrajectoryLimitationCoverageWindowUnavailable              = "trajectory_coverage_window_unavailable"
	TrajectoryLimitationCoverageObservationEvidenceUnavailable = "trajectory_coverage_observation_evidence_unavailable"
	TrajectoryLimitationCoverageGapWindowInvalid               = "trajectory_coverage_gap_window_invalid"
	TrajectoryLimitationCoverageGapOutsideWindow               = "trajectory_coverage_gap_outside_window"
	TrajectoryLimitationCoverageGapDurationMismatch            = "trajectory_coverage_gap_duration_metadata_mismatch"
	TrajectoryLimitationCoverageAggregateNonFinite             = "trajectory_coverage_aggregate_non_finite"
	TrajectoryLimitationCoverageRatioOutOfRange                = "trajectory_coverage_ratio_out_of_range"
	TrajectoryLimitationPathEvidenceInsufficient               = "trajectory_path_efficiency_evidence_insufficient"
	TrajectoryLimitationPathZeroDistance                       = "trajectory_path_efficiency_zero_path"
	TrajectoryLimitationPathAggregateNonFinite                 = "trajectory_path_aggregate_non_finite"
	TrajectoryLimitationPathRatioOutOfRange                    = "trajectory_path_ratio_out_of_range"
	TrajectoryLimitationPathSegmentFallback                    = "trajectory_path_segment_endpoint_fallback"
	TrajectoryLimitationPathPointEvidenceUnavailable           = "trajectory_path_point_evidence_unavailable"
	TrajectoryLimitationPathDiscontinuityExcluded              = "trajectory_path_discontinuity_excluded"
	TrajectoryLimitationPathInvalidSegmentStatus               = "trajectory_path_invalid_segment_status"
	TrajectoryLimitationPathInvalidSegmentCoordinates          = "trajectory_path_invalid_segment_coordinates"
	TrajectoryLimitationPathSegmentTimestampMissing            = "trajectory_path_segment_timestamp_missing"
	TrajectoryLimitationPathSegmentOutsideWindow               = "trajectory_path_segment_outside_window"
	TrajectoryLimitationQualityScoreInvalid                    = "trajectory_quality_score_invalid"
	TrajectoryLimitationSegmentSharesUndefined                 = "trajectory_segment_shares_undefined"
	TrajectoryLimitationSegmentStatusUnknown                   = "trajectory_segment_status_unknown"
)
