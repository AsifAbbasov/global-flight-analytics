package validator

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

const issueCodePrefix = "feature_validation."

var (
	icao24Pattern      = regexp.MustCompile(`^[A-F0-9]{6}$`)
	fingerprintPattern = regexp.MustCompile(
		`^sha256:[0-9a-f]{64}$`,
	)
)

func validateIdentity(
	collector *issueCollector,
	features flightfeatures.FlightFeatures,
) {
	if features.SchemaVersion != flightfeatures.SchemaVersionV1 {
		collector.error(
			"",
			"schema_version",
			issueCodePrefix+"unsupported_schema_version",
			fmt.Sprintf(
				"Schema version %q is unsupported; expected %q.",
				features.SchemaVersion,
				flightfeatures.SchemaVersionV1,
			),
		)
	}
	if strings.TrimSpace(features.TrajectoryID) == "" {
		collector.error(
			"",
			"trajectory_id",
			issueCodePrefix+"trajectory_id_required",
			"Trajectory ID is required.",
		)
	}
	if strings.TrimSpace(features.IdentityKey) == "" {
		collector.error(
			"",
			"identity_key",
			issueCodePrefix+"identity_key_required",
			"Stable trajectory identity key is required.",
		)
	}
	if !icao24Pattern.MatchString(features.ICAO24) {
		collector.error(
			"",
			"icao24",
			issueCodePrefix+"invalid_icao24",
			"ICAO24 must contain exactly six uppercase hexadecimal characters.",
		)
	}
	if features.Callsign != strings.TrimSpace(features.Callsign) {
		collector.warning(
			"",
			"callsign",
			issueCodePrefix+"callsign_not_normalized",
			"Callsign contains leading or trailing whitespace.",
		)
	}
}

func validateWindow(
	collector *issueCollector,
	features flightfeatures.FlightFeatures,
) {
	start := features.Window.StartTime
	end := features.Window.EndTime
	asOf := features.Window.AsOfTime
	extractedAt := features.ExtractedAt

	validateRequiredTimestamp(
		collector,
		"",
		"window.start_time",
		"window_start_time_required",
		start,
	)
	validateRequiredTimestamp(
		collector,
		"",
		"window.end_time",
		"window_end_time_required",
		end,
	)
	validateRequiredTimestamp(
		collector,
		"",
		"window.as_of_time",
		"as_of_time_required",
		asOf,
	)
	validateRequiredTimestamp(
		collector,
		"",
		"extracted_at",
		"extracted_at_required",
		extractedAt,
	)

	if !start.IsZero() && !end.IsZero() && end.Before(start) {
		collector.error(
			"",
			"window",
			issueCodePrefix+"invalid_feature_window",
			"Feature window end time is before its start time.",
		)
	}
	if !end.IsZero() && !asOf.IsZero() && asOf.Before(end) {
		collector.error(
			"",
			"window.as_of_time",
			issueCodePrefix+"as_of_before_window_end",
			"As-of time is before the feature window end and would permit future-data leakage.",
		)
	}
	if !asOf.IsZero() &&
		!extractedAt.IsZero() &&
		extractedAt.Before(asOf) {
		collector.error(
			"",
			"extracted_at",
			issueCodePrefix+"extracted_before_as_of",
			"Extraction time is before the declared as-of time.",
		)
	}
}

func validateProvenance(
	collector *issueCollector,
	features flightfeatures.FlightFeatures,
) {
	provenance := features.Provenance

	if strings.TrimSpace(provenance.ExtractorVersion) == "" {
		collector.error(
			"",
			"provenance.extractor_version",
			issueCodePrefix+"extractor_version_required",
			"Extractor version is required.",
		)
	}
	if !fingerprintPattern.MatchString(
		provenance.InputFingerprint,
	) {
		collector.error(
			"",
			"provenance.input_fingerprint",
			issueCodePrefix+"invalid_input_fingerprint",
			"Input fingerprint must use the sha256 prefix followed by 64 lowercase hexadecimal characters.",
		)
	}
	validateRequiredTimestamp(
		collector,
		"",
		"provenance.trajectory_updated_at",
		"trajectory_updated_at_required",
		provenance.TrajectoryUpdatedAt,
	)
	if !provenance.TrajectoryUpdatedAt.IsZero() &&
		!features.Window.AsOfTime.IsZero() &&
		provenance.TrajectoryUpdatedAt.After(
			features.Window.AsOfTime,
		) {
		collector.error(
			"",
			"provenance.trajectory_updated_at",
			issueCodePrefix+"trajectory_updated_after_as_of",
			"Trajectory update time is after the declared as-of time.",
		)
	}

	if len(provenance.SourceNames) == 0 {
		collector.warning(
			"",
			"provenance.source_names",
			issueCodePrefix+"source_names_unavailable",
			"No source provenance names are available.",
		)
		return
	}

	seen := make(map[string]struct{}, len(provenance.SourceNames))
	for index, sourceName := range provenance.SourceNames {
		path := fmt.Sprintf(
			"provenance.source_names[%d]",
			index,
		)
		if strings.TrimSpace(sourceName) == "" {
			collector.error(
				"",
				path,
				issueCodePrefix+"empty_source_name",
				"Source provenance name must not be empty.",
			)
			continue
		}
		if sourceName != strings.TrimSpace(sourceName) {
			collector.error(
				"",
				path,
				issueCodePrefix+"source_name_not_normalized",
				"Source provenance name contains leading or trailing whitespace.",
			)
		}
		if _, exists := seen[sourceName]; exists {
			collector.error(
				"",
				path,
				issueCodePrefix+"duplicate_source_name",
				fmt.Sprintf(
					"Source provenance name %q is duplicated.",
					sourceName,
				),
			)
		}
		seen[sourceName] = struct{}{}
	}

	if !sort.StringsAreSorted(provenance.SourceNames) {
		collector.error(
			"",
			"provenance.source_names",
			issueCodePrefix+"source_names_not_sorted",
			"Source provenance names must be sorted deterministically.",
		)
	}
}

func validateGroupEvidence(
	collector *issueCollector,
	group flightfeatures.FeatureGroup,
	path string,
	evidence flightfeatures.GroupEvidence,
	expectedFieldCount int,
) {
	switch evidence.Status {
	case flightfeatures.AvailabilityStatusAvailable,
		flightfeatures.AvailabilityStatusPartial,
		flightfeatures.AvailabilityStatusUnavailable:
	default:
		collector.error(
			group,
			path+".status",
			issueCodePrefix+"unsupported_availability_status",
			fmt.Sprintf(
				"Availability status %q is unsupported.",
				evidence.Status,
			),
		)
	}

	if evidence.TotalFieldCount != expectedFieldCount {
		collector.error(
			group,
			path+".total_field_count",
			issueCodePrefix+"schema_field_count_mismatch",
			fmt.Sprintf(
				"Evidence total field count is %d; schema requires %d for group %q.",
				evidence.TotalFieldCount,
				expectedFieldCount,
				group,
			),
		)
	}
	if evidence.AvailableFieldCount < 0 {
		collector.error(
			group,
			path+".available_field_count",
			issueCodePrefix+"negative_available_field_count",
			"Available field count must not be negative.",
		)
	}
	if evidence.TotalFieldCount < 0 {
		collector.error(
			group,
			path+".total_field_count",
			issueCodePrefix+"negative_total_field_count",
			"Total field count must not be negative.",
		)
	}
	if evidence.AvailableFieldCount > evidence.TotalFieldCount {
		collector.error(
			group,
			path+".available_field_count",
			issueCodePrefix+"available_field_count_exceeds_total",
			"Available field count exceeds total field count.",
		)
	}
	if evidence.SupportingPointCount < 0 {
		collector.error(
			group,
			path+".supporting_point_count",
			issueCodePrefix+"negative_supporting_point_count",
			"Supporting point count must not be negative.",
		)
	}

	switch evidence.Status {
	case flightfeatures.AvailabilityStatusAvailable:
		if evidence.TotalFieldCount <= 0 ||
			evidence.AvailableFieldCount !=
				evidence.TotalFieldCount {
			collector.error(
				group,
				path,
				issueCodePrefix+"available_evidence_inconsistent",
				"Available evidence must expose every schema field in the group.",
			)
		}
	case flightfeatures.AvailabilityStatusPartial:
		if evidence.AvailableFieldCount <= 0 ||
			evidence.AvailableFieldCount >=
				evidence.TotalFieldCount {
			collector.error(
				group,
				path,
				issueCodePrefix+"partial_evidence_inconsistent",
				"Partial evidence must expose at least one but fewer than all schema fields.",
			)
		}
		collector.warning(
			group,
			path,
			issueCodePrefix+"feature_group_partial",
			fmt.Sprintf(
				"Feature group %q is only partially available.",
				group,
			),
		)
	case flightfeatures.AvailabilityStatusUnavailable:
		if evidence.AvailableFieldCount != 0 {
			collector.error(
				group,
				path,
				issueCodePrefix+"unavailable_evidence_inconsistent",
				"Unavailable evidence must expose zero available fields.",
			)
		}
		collector.warning(
			group,
			path,
			issueCodePrefix+"feature_group_unavailable",
			fmt.Sprintf(
				"Feature group %q is unavailable.",
				group,
			),
		)
	}

	validateLimitations(
		collector,
		group,
		path+".limitations",
		evidence.Limitations,
		true,
	)
}

func validateTemporalFeatures(
	collector *issueCollector,
	features flightfeatures.FlightFeatures,
) {
	item := features.Temporal
	if item.Evidence.Status ==
		flightfeatures.AvailabilityStatusUnavailable {
		return
	}

	severity := relationshipSeverity(item.Evidence.Status)
	validateIntegerRange(
		collector,
		severity,
		flightfeatures.FeatureGroupTemporal,
		"temporal.start_hour_utc",
		item.StartHourUTC,
		0,
		23,
	)
	validateIntegerRange(
		collector,
		severity,
		flightfeatures.FeatureGroupTemporal,
		"temporal.end_hour_utc",
		item.EndHourUTC,
		0,
		23,
	)
	validateIntegerRange(
		collector,
		severity,
		flightfeatures.FeatureGroupTemporal,
		"temporal.start_weekday",
		int(item.StartWeekday),
		int(time.Sunday),
		int(time.Saturday),
	)
	validateIntegerRange(
		collector,
		severity,
		flightfeatures.FeatureGroupTemporal,
		"temporal.end_weekday",
		int(item.EndWeekday),
		int(time.Sunday),
		int(time.Saturday),
	)
	validateIntegerRange(
		collector,
		severity,
		flightfeatures.FeatureGroupTemporal,
		"temporal.start_minute_of_day_utc",
		item.StartMinuteOfDayUTC,
		0,
		1439,
	)
	validateIntegerRange(
		collector,
		severity,
		flightfeatures.FeatureGroupTemporal,
		"temporal.end_minute_of_day_utc",
		item.EndMinuteOfDayUTC,
		0,
		1439,
	)
	if item.DurationSeconds < 0 {
		addBySeverity(
			collector,
			severity,
			flightfeatures.FeatureGroupTemporal,
			"temporal.duration_seconds",
			issueCodePrefix+"negative_duration",
			"Temporal duration must not be negative.",
		)
	}

	start := features.Window.StartTime.UTC()
	end := features.Window.EndTime.UTC()
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return
	}

	expectedDuration := int64(end.Sub(start) / time.Second)
	if item.DurationSeconds != expectedDuration {
		addBySeverity(
			collector,
			severity,
			flightfeatures.FeatureGroupTemporal,
			"temporal.duration_seconds",
			issueCodePrefix+"duration_window_mismatch",
			fmt.Sprintf(
				"Temporal duration is %d seconds; feature window duration is %d seconds.",
				item.DurationSeconds,
				expectedDuration,
			),
		)
	}
	if item.StartHourUTC != start.Hour() {
		addBySeverity(
			collector,
			severity,
			flightfeatures.FeatureGroupTemporal,
			"temporal.start_hour_utc",
			issueCodePrefix+"start_hour_mismatch",
			"Start hour does not match the feature window start.",
		)
	}
	if item.EndHourUTC != end.Hour() {
		addBySeverity(
			collector,
			severity,
			flightfeatures.FeatureGroupTemporal,
			"temporal.end_hour_utc",
			issueCodePrefix+"end_hour_mismatch",
			"End hour does not match the feature window end.",
		)
	}
	if item.StartWeekday != start.Weekday() {
		addBySeverity(
			collector,
			severity,
			flightfeatures.FeatureGroupTemporal,
			"temporal.start_weekday",
			issueCodePrefix+"start_weekday_mismatch",
			"Start weekday does not match the feature window start.",
		)
	}
	if item.EndWeekday != end.Weekday() {
		addBySeverity(
			collector,
			severity,
			flightfeatures.FeatureGroupTemporal,
			"temporal.end_weekday",
			issueCodePrefix+"end_weekday_mismatch",
			"End weekday does not match the feature window end.",
		)
	}
	if item.StartMinuteOfDayUTC !=
		start.Hour()*60+start.Minute() {
		addBySeverity(
			collector,
			severity,
			flightfeatures.FeatureGroupTemporal,
			"temporal.start_minute_of_day_utc",
			issueCodePrefix+"start_minute_of_day_mismatch",
			"Start minute of day does not match the feature window start.",
		)
	}
	if item.EndMinuteOfDayUTC !=
		end.Hour()*60+end.Minute() {
		addBySeverity(
			collector,
			severity,
			flightfeatures.FeatureGroupTemporal,
			"temporal.end_minute_of_day_utc",
			issueCodePrefix+"end_minute_of_day_mismatch",
			"End minute of day does not match the feature window end.",
		)
	}

	expectedCrossesMidnight :=
		start.Year() != end.Year() ||
			start.YearDay() != end.YearDay()
	if item.CrossesUTCMidnight != expectedCrossesMidnight {
		addBySeverity(
			collector,
			severity,
			flightfeatures.FeatureGroupTemporal,
			"temporal.crosses_utc_midnight",
			issueCodePrefix+"utc_midnight_flag_mismatch",
			"UTC midnight crossing flag does not match the feature window.",
		)
	}
}

func validateGeographicalFeatures(
	collector *issueCollector,
	features flightfeatures.FlightFeatures,
) {
	item := features.Geographical
	if item.Evidence.Status ==
		flightfeatures.AvailabilityStatusUnavailable {
		return
	}
	severity := relationshipSeverity(item.Evidence.Status)
	group := flightfeatures.FeatureGroupGeographical

	validateLatitude(
		collector,
		severity,
		group,
		"geographical.start_latitude",
		item.StartLatitude,
	)
	validateLongitude(
		collector,
		severity,
		group,
		"geographical.start_longitude",
		item.StartLongitude,
	)
	validateLatitude(
		collector,
		severity,
		group,
		"geographical.end_latitude",
		item.EndLatitude,
	)
	validateLongitude(
		collector,
		severity,
		group,
		"geographical.end_longitude",
		item.EndLongitude,
	)
	validateLatitude(
		collector,
		severity,
		group,
		"geographical.minimum_latitude",
		item.MinimumLatitude,
	)
	validateLatitude(
		collector,
		severity,
		group,
		"geographical.maximum_latitude",
		item.MaximumLatitude,
	)
	validateLongitude(
		collector,
		severity,
		group,
		"geographical.minimum_longitude",
		item.MinimumLongitude,
	)
	validateLongitude(
		collector,
		severity,
		group,
		"geographical.maximum_longitude",
		item.MaximumLongitude,
	)

	if finite(item.MinimumLatitude) &&
		finite(item.MaximumLatitude) &&
		item.MinimumLatitude > item.MaximumLatitude {
		addBySeverity(
			collector,
			severity,
			group,
			"geographical.latitude_bounds",
			issueCodePrefix+"latitude_bounds_reversed",
			"Minimum latitude exceeds maximum latitude.",
		)
	}
	if finite(item.MinimumLongitude) &&
		finite(item.MaximumLongitude) &&
		item.MinimumLongitude > item.MaximumLongitude &&
		!item.CrossesAntimeridian {
		addBySeverity(
			collector,
			severity,
			group,
			"geographical.longitude_bounds",
			issueCodePrefix+"longitude_bounds_reversed",
			"Minimum longitude exceeds maximum longitude without an antimeridian crossing.",
		)
	}

	validateNonNegativeFinite(
		collector,
		severity,
		group,
		"geographical.latitude_span_degrees",
		item.LatitudeSpanDegrees,
	)
	validateNonNegativeFinite(
		collector,
		severity,
		group,
		"geographical.longitude_span_degrees",
		item.LongitudeSpanDegrees,
	)
	if finite(item.LatitudeSpanDegrees) &&
		item.LatitudeSpanDegrees > 180 {
		addBySeverity(
			collector,
			severity,
			group,
			"geographical.latitude_span_degrees",
			issueCodePrefix+"latitude_span_out_of_range",
			"Latitude span must not exceed 180 degrees.",
		)
	}
	if finite(item.LongitudeSpanDegrees) &&
		item.LongitudeSpanDegrees > 360 {
		addBySeverity(
			collector,
			severity,
			group,
			"geographical.longitude_span_degrees",
			issueCodePrefix+"longitude_span_out_of_range",
			"Longitude span must not exceed 360 degrees.",
		)
	}
	if finite(item.MinimumLatitude) &&
		finite(item.MaximumLatitude) &&
		!approximatelyEqual(
			item.LatitudeSpanDegrees,
			item.MaximumLatitude-item.MinimumLatitude,
			collector.tolerance,
		) {
		addBySeverity(
			collector,
			severity,
			group,
			"geographical.latitude_span_degrees",
			issueCodePrefix+"latitude_span_mismatch",
			"Latitude span does not match the declared latitude bounds.",
		)
	}

	validateNonNegativeFinite(
		collector,
		severity,
		group,
		"geographical.great_circle_distance_km",
		item.GreatCircleDistanceKM,
	)
	validateNonNegativeFinite(
		collector,
		severity,
		group,
		"geographical.observed_path_distance_km",
		item.ObservedPathDistanceKM,
	)
	validateNonNegativeFinite(
		collector,
		severity,
		group,
		"geographical.maximum_displacement_km",
		item.MaximumDisplacementKM,
	)
	if finite(item.GreatCircleDistanceKM) &&
		finite(item.ObservedPathDistanceKM) &&
		item.GreatCircleDistanceKM >
			item.ObservedPathDistanceKM+collector.tolerance {
		addBySeverity(
			collector,
			severity,
			group,
			"geographical.observed_path_distance_km",
			issueCodePrefix+"path_shorter_than_great_circle",
			"Observed path distance is shorter than the endpoint great-circle distance.",
		)
	}
	if finite(item.MaximumDisplacementKM) &&
		finite(item.ObservedPathDistanceKM) &&
		item.MaximumDisplacementKM >
			item.ObservedPathDistanceKM+collector.tolerance {
		addBySeverity(
			collector,
			severity,
			group,
			"geographical.maximum_displacement_km",
			issueCodePrefix+"displacement_exceeds_path",
			"Maximum displacement exceeds observed path distance.",
		)
	}
	if item.UniqueGeographicCellCount <= 0 {
		addBySeverity(
			collector,
			severity,
			group,
			"geographical.unique_geographic_cell_count",
			issueCodePrefix+"geographic_cell_count_required",
			"At least one geographic cell is required when geographical features are available.",
		)
	}
	if item.GeographicCellPrecision <= 0 {
		addBySeverity(
			collector,
			severity,
			group,
			"geographical.geographic_cell_precision",
			issueCodePrefix+"geographic_cell_precision_required",
			"Geographic cell precision must be greater than zero when geographical features are available.",
		)
	}
}

func validateOperationalFeatures(
	collector *issueCollector,
	features flightfeatures.FlightFeatures,
) {
	item := features.Operational
	if item.Evidence.Status ==
		flightfeatures.AvailabilityStatusUnavailable {
		return
	}
	severity := relationshipSeverity(item.Evidence.Status)
	group := flightfeatures.FeatureGroupOperational

	values := []struct {
		path  string
		value float64
	}{
		{"operational.minimum_altitude_m", item.MinimumAltitudeM},
		{"operational.maximum_altitude_m", item.MaximumAltitudeM},
		{"operational.mean_altitude_m", item.MeanAltitudeM},
		{"operational.altitude_range_m", item.AltitudeRangeM},
		{"operational.mean_velocity_mps", item.MeanVelocityMPS},
		{"operational.maximum_velocity_mps", item.MaximumVelocityMPS},
		{"operational.mean_absolute_vertical_rate_mps", item.MeanAbsoluteVerticalRateMPS},
		{"operational.maximum_absolute_vertical_rate_mps", item.MaximumAbsoluteVerticalRateMPS},
		{"operational.heading_change_degrees", item.HeadingChangeDegrees},
		{"operational.ground_observation_share", item.GroundObservationShare},
		{"operational.airborne_observation_share", item.AirborneObservationShare},
	}
	for _, value := range values {
		validateFinite(
			collector,
			severity,
			group,
			value.path,
			value.value,
		)
	}

	if finite(item.MinimumAltitudeM) &&
		finite(item.MaximumAltitudeM) &&
		item.MinimumAltitudeM > item.MaximumAltitudeM {
		addBySeverity(
			collector,
			severity,
			group,
			"operational.altitude_bounds",
			issueCodePrefix+"altitude_bounds_reversed",
			"Minimum altitude exceeds maximum altitude.",
		)
	}
	if finite(item.MeanAltitudeM) &&
		finite(item.MinimumAltitudeM) &&
		finite(item.MaximumAltitudeM) &&
		(item.MeanAltitudeM <
			item.MinimumAltitudeM-collector.tolerance ||
			item.MeanAltitudeM >
				item.MaximumAltitudeM+collector.tolerance) {
		addBySeverity(
			collector,
			severity,
			group,
			"operational.mean_altitude_m",
			issueCodePrefix+"mean_altitude_outside_bounds",
			"Mean altitude is outside the declared altitude bounds.",
		)
	}
	if finite(item.AltitudeRangeM) &&
		item.AltitudeRangeM < 0 {
		addBySeverity(
			collector,
			severity,
			group,
			"operational.altitude_range_m",
			issueCodePrefix+"negative_altitude_range",
			"Altitude range must not be negative.",
		)
	}
	if finite(item.MinimumAltitudeM) &&
		finite(item.MaximumAltitudeM) &&
		finite(item.AltitudeRangeM) &&
		!approximatelyEqual(
			item.AltitudeRangeM,
			item.MaximumAltitudeM-item.MinimumAltitudeM,
			collector.tolerance,
		) {
		addBySeverity(
			collector,
			severity,
			group,
			"operational.altitude_range_m",
			issueCodePrefix+"altitude_range_mismatch",
			"Altitude range does not equal maximum altitude minus minimum altitude.",
		)
	}

	validateOrderedNonNegativePair(
		collector,
		severity,
		group,
		"operational.mean_velocity_mps",
		item.MeanVelocityMPS,
		"operational.maximum_velocity_mps",
		item.MaximumVelocityMPS,
		"velocity",
	)
	validateOrderedNonNegativePair(
		collector,
		severity,
		group,
		"operational.mean_absolute_vertical_rate_mps",
		item.MeanAbsoluteVerticalRateMPS,
		"operational.maximum_absolute_vertical_rate_mps",
		item.MaximumAbsoluteVerticalRateMPS,
		"absolute vertical rate",
	)
	validateNonNegativeFinite(
		collector,
		severity,
		group,
		"operational.heading_change_degrees",
		item.HeadingChangeDegrees,
	)
	validateRatio(
		collector,
		severity,
		group,
		"operational.ground_observation_share",
		item.GroundObservationShare,
	)
	validateRatio(
		collector,
		severity,
		group,
		"operational.airborne_observation_share",
		item.AirborneObservationShare,
	)
	if item.Evidence.SupportingPointCount > 0 &&
		finite(item.GroundObservationShare) &&
		finite(item.AirborneObservationShare) &&
		!approximatelyEqual(
			item.GroundObservationShare+
				item.AirborneObservationShare,
			1,
			collector.tolerance,
		) {
		addBySeverity(
			collector,
			severity,
			group,
			"operational.observation_shares",
			issueCodePrefix+"observation_shares_do_not_sum_to_one",
			"Ground and airborne observation shares must sum to one.",
		)
	}
}

func validateTrajectoryFeatures(
	collector *issueCollector,
	features flightfeatures.FlightFeatures,
) {
	item := features.Trajectory
	if item.Evidence.Status ==
		flightfeatures.AvailabilityStatusUnavailable {
		return
	}
	severity := relationshipSeverity(item.Evidence.Status)
	group := flightfeatures.FeatureGroupTrajectory

	counts := []struct {
		path  string
		value int
	}{
		{"trajectory.point_count", item.PointCount},
		{"trajectory.segment_count", item.SegmentCount},
		{"trajectory.coverage_gap_count", item.CoverageGapCount},
		{"trajectory.observed_segment_count", item.ObservedSegmentCount},
		{"trajectory.interpolated_segment_count", item.InterpolatedSegmentCount},
		{"trajectory.estimated_segment_count", item.EstimatedSegmentCount},
		{"trajectory.invalid_segment_count", item.InvalidSegmentCount},
	}
	for _, count := range counts {
		if count.value < 0 {
			addBySeverity(
				collector,
				severity,
				group,
				count.path,
				issueCodePrefix+"negative_count",
				fmt.Sprintf(
					"%s must not be negative.",
					count.path,
				),
			)
		}
	}

	validateRatio(
		collector,
		severity,
		group,
		"trajectory.quality_score",
		item.TrajectoryQualityScore,
	)
	shares := []struct {
		path  string
		value float64
		count int
	}{
		{"trajectory.observed_segment_share", item.ObservedSegmentShare, item.ObservedSegmentCount},
		{"trajectory.interpolated_segment_share", item.InterpolatedSegmentShare, item.InterpolatedSegmentCount},
		{"trajectory.estimated_segment_share", item.EstimatedSegmentShare, item.EstimatedSegmentCount},
		{"trajectory.invalid_segment_share", item.InvalidSegmentShare, item.InvalidSegmentCount},
	}
	for _, share := range shares {
		validateRatio(
			collector,
			severity,
			group,
			share.path,
			share.value,
		)
	}

	statusCountTotal :=
		item.ObservedSegmentCount +
			item.InterpolatedSegmentCount +
			item.EstimatedSegmentCount +
			item.InvalidSegmentCount
	if statusCountTotal != item.SegmentCount {
		addBySeverity(
			collector,
			severity,
			group,
			"trajectory.segment_counts",
			issueCodePrefix+"segment_status_count_mismatch",
			fmt.Sprintf(
				"Segment status counts total %d; segment count is %d.",
				statusCountTotal,
				item.SegmentCount,
			),
		)
	}

	if item.SegmentCount == 0 {
		for _, share := range shares {
			if finite(share.value) &&
				!approximatelyEqual(
					share.value,
					0,
					collector.tolerance,
				) {
				addBySeverity(
					collector,
					severity,
					group,
					share.path,
					issueCodePrefix+"nonzero_share_without_segments",
					"Segment share must be zero when segment count is zero.",
				)
			}
		}
		collector.warning(
			group,
			"trajectory.segment_count",
			issueCodePrefix+"no_trajectory_segments",
			"No trajectory segments support the extracted feature set.",
		)
	} else {
		shareTotal :=
			item.ObservedSegmentShare +
				item.InterpolatedSegmentShare +
				item.EstimatedSegmentShare +
				item.InvalidSegmentShare
		if finite(shareTotal) &&
			!approximatelyEqual(
				shareTotal,
				1,
				collector.tolerance,
			) {
			addBySeverity(
				collector,
				severity,
				group,
				"trajectory.segment_shares",
				issueCodePrefix+"segment_shares_do_not_sum_to_one",
				"Segment status shares must sum to one.",
			)
		}

		for _, share := range shares {
			expectedShare := float64(share.count) /
				float64(item.SegmentCount)
			if finite(share.value) &&
				!approximatelyEqual(
					share.value,
					expectedShare,
					collector.tolerance,
				) {
				addBySeverity(
					collector,
					severity,
					group,
					share.path,
					issueCodePrefix+"segment_share_count_mismatch",
					fmt.Sprintf(
						"Segment share %.6f does not match count-derived share %.6f.",
						share.value,
						expectedShare,
					),
				)
			}
		}
	}

	validateNonNegativeFinite(
		collector,
		severity,
		group,
		"trajectory.mean_sampling_interval_seconds",
		item.MeanSamplingIntervalSeconds,
	)
	validateNonNegativeFinite(
		collector,
		severity,
		group,
		"trajectory.maximum_sampling_gap_seconds",
		item.MaximumSamplingGapSeconds,
	)
	if finite(item.MeanSamplingIntervalSeconds) &&
		finite(item.MaximumSamplingGapSeconds) &&
		item.MaximumSamplingGapSeconds+
			collector.tolerance <
			item.MeanSamplingIntervalSeconds {
		addBySeverity(
			collector,
			severity,
			group,
			"trajectory.maximum_sampling_gap_seconds",
			issueCodePrefix+"maximum_gap_below_mean_interval",
			"Maximum sampling gap is below the mean sampling interval.",
		)
	}
	validateRatio(
		collector,
		severity,
		group,
		"trajectory.coverage_ratio",
		item.CoverageRatio,
	)
	validateRatio(
		collector,
		severity,
		group,
		"trajectory.path_efficiency_ratio",
		item.PathEfficiencyRatio,
	)

	if item.PointCount < 2 {
		collector.warning(
			group,
			"trajectory.point_count",
			issueCodePrefix+"insufficient_point_evidence",
			"Fewer than two trajectory points support the feature set.",
		)
	}
	if item.Evidence.SupportingPointCount != item.PointCount {
		addBySeverity(
			collector,
			severity,
			group,
			"trajectory.evidence.supporting_point_count",
			issueCodePrefix+"trajectory_supporting_point_count_mismatch",
			fmt.Sprintf(
				"Trajectory evidence reports %d supporting points; trajectory features report %d points.",
				item.Evidence.SupportingPointCount,
				item.PointCount,
			),
		)
	}

	if features.Geographical.Evidence.Status !=
		flightfeatures.AvailabilityStatusUnavailable &&
		features.Geographical.ObservedPathDistanceKM > 0 &&
		finite(item.PathEfficiencyRatio) {
		expectedRatio :=
			features.Geographical.GreatCircleDistanceKM /
				features.Geographical.ObservedPathDistanceKM
		if !approximatelyEqual(
			item.PathEfficiencyRatio,
			expectedRatio,
			collector.tolerance,
		) {
			addBySeverity(
				collector,
				severity,
				group,
				"trajectory.path_efficiency_ratio",
				issueCodePrefix+"path_efficiency_mismatch",
				fmt.Sprintf(
					"Path efficiency ratio %.6f does not match distance-derived ratio %.6f.",
					item.PathEfficiencyRatio,
					expectedRatio,
				),
			)
		}
	}
}

func validateAircraftFeatures(
	collector *issueCollector,
	features flightfeatures.FlightFeatures,
) {
	item := features.Aircraft
	values := []struct {
		path  string
		value string
	}{
		{"aircraft.registration", item.Registration},
		{"aircraft.manufacturer", item.Manufacturer},
		{"aircraft.model", item.Model},
		{"aircraft.aircraft_type", item.AircraftType},
		{"aircraft.airline", item.Airline},
		{"aircraft.country", item.Country},
	}

	availableCount := 0
	for _, value := range values {
		if value.value != strings.TrimSpace(value.value) {
			collector.warning(
				flightfeatures.FeatureGroupAircraft,
				value.path,
				issueCodePrefix+"aircraft_field_not_normalized",
				fmt.Sprintf(
					"%s contains leading or trailing whitespace.",
					value.path,
				),
			)
		}
		if strings.TrimSpace(value.value) != "" {
			availableCount++
		}
	}

	if availableCount != item.Evidence.AvailableFieldCount {
		severity := relationshipSeverity(item.Evidence.Status)
		if item.Evidence.Status ==
			flightfeatures.AvailabilityStatusUnavailable {
			severity = IssueSeverityError
		}
		addBySeverity(
			collector,
			severity,
			flightfeatures.FeatureGroupAircraft,
			"aircraft.evidence.available_field_count",
			issueCodePrefix+"aircraft_available_field_count_mismatch",
			fmt.Sprintf(
				"Aircraft evidence reports %d available fields; %d non-empty aircraft fields are present.",
				item.Evidence.AvailableFieldCount,
				availableCount,
			),
		)
	}
}

func validateQuality(
	collector *issueCollector,
	features flightfeatures.FlightFeatures,
	policy Policy,
) {
	quality := features.Quality

	switch quality.Status {
	case flightfeatures.ValidationStatusUnvalidated,
		flightfeatures.ValidationStatusValid,
		flightfeatures.ValidationStatusLimited,
		flightfeatures.ValidationStatusInvalid:
	default:
		collector.error(
			"",
			"quality.status",
			issueCodePrefix+"unsupported_validation_status",
			fmt.Sprintf(
				"Validation status %q is unsupported.",
				quality.Status,
			),
		)
	}

	validateRatio(
		collector,
		IssueSeverityError,
		"",
		"quality.completeness_score",
		quality.CompletenessScore,
	)
	validateRatio(
		collector,
		IssueSeverityError,
		"",
		"quality.input_quality_score",
		quality.InputQualityScore,
	)
	if quality.SupportingPointCount < 0 {
		collector.error(
			"",
			"quality.supporting_point_count",
			issueCodePrefix+"negative_quality_supporting_point_count",
			"Quality supporting point count must not be negative.",
		)
	}

	evidenceGroups := []flightfeatures.GroupEvidence{
		features.Temporal.Evidence,
		features.Geographical.Evidence,
		features.Operational.Evidence,
		features.Trajectory.Evidence,
		features.Aircraft.Evidence,
	}
	availableFields := 0
	totalFields := 0
	expectedSupportingPoints := features.Trajectory.PointCount
	for _, evidence := range evidenceGroups {
		if evidence.AvailableFieldCount > 0 {
			availableFields += evidence.AvailableFieldCount
		}
		if evidence.TotalFieldCount > 0 {
			totalFields += evidence.TotalFieldCount
		}
		if evidence.SupportingPointCount >
			expectedSupportingPoints {
			expectedSupportingPoints =
				evidence.SupportingPointCount
		}
	}

	expectedCompleteness := 0.0
	if totalFields > 0 {
		expectedCompleteness =
			float64(availableFields) / float64(totalFields)
	}
	if finite(quality.CompletenessScore) &&
		!approximatelyEqual(
			quality.CompletenessScore,
			expectedCompleteness,
			collector.tolerance,
		) {
		collector.error(
			"",
			"quality.completeness_score",
			issueCodePrefix+"completeness_score_mismatch",
			fmt.Sprintf(
				"Completeness score %.6f does not match evidence-derived score %.6f.",
				quality.CompletenessScore,
				expectedCompleteness,
			),
		)
	}
	if quality.SupportingPointCount != expectedSupportingPoints {
		collector.error(
			"",
			"quality.supporting_point_count",
			issueCodePrefix+"quality_supporting_point_count_mismatch",
			fmt.Sprintf(
				"Quality reports %d supporting points; evidence requires %d.",
				quality.SupportingPointCount,
				expectedSupportingPoints,
			),
		)
	}
	if finite(quality.InputQualityScore) &&
		finite(features.Trajectory.TrajectoryQualityScore) &&
		!approximatelyEqual(
			quality.InputQualityScore,
			features.Trajectory.TrajectoryQualityScore,
			collector.tolerance,
		) {
		collector.error(
			"",
			"quality.input_quality_score",
			issueCodePrefix+"input_quality_score_mismatch",
			"Input quality score does not match the trajectory quality feature.",
		)
	}

	if ratioInRange(quality.CompletenessScore) &&
		quality.CompletenessScore+
			collector.tolerance <
			policy.MinimumValidCompletenessScore {
		collector.warning(
			"",
			"quality.completeness_score",
			issueCodePrefix+"completeness_below_valid_threshold",
			fmt.Sprintf(
				"Completeness score %.3f is below the valid threshold %.3f.",
				quality.CompletenessScore,
				policy.MinimumValidCompletenessScore,
			),
		)
	}
	if ratioInRange(quality.InputQualityScore) &&
		quality.InputQualityScore+
			collector.tolerance <
			policy.MinimumValidInputQualityScore {
		collector.warning(
			"",
			"quality.input_quality_score",
			issueCodePrefix+"input_quality_below_valid_threshold",
			fmt.Sprintf(
				"Input quality score %.3f is below the valid threshold %.3f.",
				quality.InputQualityScore,
				policy.MinimumValidInputQualityScore,
			),
		)
	}

	validateLimitations(
		collector,
		"",
		"quality.limitations",
		quality.Limitations,
		true,
	)
}

func validateLimitations(
	collector *issueCollector,
	group flightfeatures.FeatureGroup,
	path string,
	limitations []flightfeatures.FeatureLimitation,
	markValidAsWarning bool,
) {
	seen := make(map[string]struct{}, len(limitations))

	for index, limitation := range limitations {
		itemPath := fmt.Sprintf("%s[%d]", path, index)
		code := strings.TrimSpace(limitation.Code)
		message := strings.TrimSpace(limitation.Message)

		if code == "" {
			collector.error(
				group,
				itemPath+".code",
				issueCodePrefix+"limitation_code_required",
				"Feature limitation code is required.",
			)
		}
		if message == "" {
			collector.error(
				group,
				itemPath+".message",
				issueCodePrefix+"limitation_message_required",
				"Feature limitation message is required.",
			)
		}
		if limitation.Code != code ||
			limitation.Message != message {
			collector.error(
				group,
				itemPath,
				issueCodePrefix+"limitation_not_normalized",
				"Feature limitation code and message must not contain leading or trailing whitespace.",
			)
		}

		key := code + "\x00" + message
		if _, exists := seen[key]; exists {
			collector.warning(
				group,
				itemPath,
				issueCodePrefix+"duplicate_limitation",
				"Duplicate feature limitation was reported.",
			)
		}
		seen[key] = struct{}{}

		if markValidAsWarning &&
			code != "" &&
			message != "" &&
			!strings.HasPrefix(code, issueCodePrefix) {
			collector.warning(
				group,
				itemPath,
				code,
				message,
			)
		}
	}
}

func validateRequiredTimestamp(
	collector *issueCollector,
	group flightfeatures.FeatureGroup,
	path string,
	codeSuffix string,
	value time.Time,
) {
	if value.IsZero() {
		collector.error(
			group,
			path,
			issueCodePrefix+codeSuffix,
			fmt.Sprintf("%s is required.", path),
		)
		return
	}

	_, offset := value.Zone()
	if offset != 0 {
		collector.error(
			group,
			path,
			issueCodePrefix+"timestamp_not_utc",
			fmt.Sprintf("%s must be normalized to UTC.", path),
		)
	}
}

func relationshipSeverity(
	status flightfeatures.AvailabilityStatus,
) IssueSeverity {
	if status == flightfeatures.AvailabilityStatusPartial {
		return IssueSeverityWarning
	}

	return IssueSeverityError
}

func addBySeverity(
	collector *issueCollector,
	severity IssueSeverity,
	group flightfeatures.FeatureGroup,
	path string,
	code string,
	message string,
) {
	if severity == IssueSeverityWarning {
		collector.warning(group, path, code, message)
		return
	}

	collector.error(group, path, code, message)
}

func validateIntegerRange(
	collector *issueCollector,
	severity IssueSeverity,
	group flightfeatures.FeatureGroup,
	path string,
	value int,
	minimum int,
	maximum int,
) {
	if value < minimum || value > maximum {
		addBySeverity(
			collector,
			severity,
			group,
			path,
			issueCodePrefix+"integer_out_of_range",
			fmt.Sprintf(
				"%s must be between %d and %d.",
				path,
				minimum,
				maximum,
			),
		)
	}
}

func validateLatitude(
	collector *issueCollector,
	severity IssueSeverity,
	group flightfeatures.FeatureGroup,
	path string,
	value float64,
) {
	if !finite(value) || value < -90 || value > 90 {
		addBySeverity(
			collector,
			severity,
			group,
			path,
			issueCodePrefix+"latitude_out_of_range",
			fmt.Sprintf(
				"%s must be a finite latitude between -90 and 90.",
				path,
			),
		)
	}
}

func validateLongitude(
	collector *issueCollector,
	severity IssueSeverity,
	group flightfeatures.FeatureGroup,
	path string,
	value float64,
) {
	if !finite(value) || value < -180 || value > 180 {
		addBySeverity(
			collector,
			severity,
			group,
			path,
			issueCodePrefix+"longitude_out_of_range",
			fmt.Sprintf(
				"%s must be a finite longitude between -180 and 180.",
				path,
			),
		)
	}
}

func validateFinite(
	collector *issueCollector,
	severity IssueSeverity,
	group flightfeatures.FeatureGroup,
	path string,
	value float64,
) {
	if !finite(value) {
		addBySeverity(
			collector,
			severity,
			group,
			path,
			issueCodePrefix+"non_finite_value",
			fmt.Sprintf("%s must be finite.", path),
		)
	}
}

func validateNonNegativeFinite(
	collector *issueCollector,
	severity IssueSeverity,
	group flightfeatures.FeatureGroup,
	path string,
	value float64,
) {
	if !finite(value) || value < 0 {
		addBySeverity(
			collector,
			severity,
			group,
			path,
			issueCodePrefix+"negative_or_non_finite_value",
			fmt.Sprintf(
				"%s must be finite and non-negative.",
				path,
			),
		)
	}
}

func validateRatio(
	collector *issueCollector,
	severity IssueSeverity,
	group flightfeatures.FeatureGroup,
	path string,
	value float64,
) {
	if !ratioInRange(value) {
		addBySeverity(
			collector,
			severity,
			group,
			path,
			issueCodePrefix+"ratio_out_of_range",
			fmt.Sprintf(
				"%s must be a finite ratio between zero and one.",
				path,
			),
		)
	}
}

func validateOrderedNonNegativePair(
	collector *issueCollector,
	severity IssueSeverity,
	group flightfeatures.FeatureGroup,
	meanPath string,
	meanValue float64,
	maximumPath string,
	maximumValue float64,
	label string,
) {
	validateNonNegativeFinite(
		collector,
		severity,
		group,
		meanPath,
		meanValue,
	)
	validateNonNegativeFinite(
		collector,
		severity,
		group,
		maximumPath,
		maximumValue,
	)

	if finite(meanValue) &&
		finite(maximumValue) &&
		meanValue > maximumValue+collector.tolerance {
		addBySeverity(
			collector,
			severity,
			group,
			maximumPath,
			issueCodePrefix+"maximum_below_mean",
			fmt.Sprintf(
				"Maximum %s is below mean %s.",
				label,
				label,
			),
		)
	}
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func approximatelyEqual(
	left float64,
	right float64,
	tolerance float64,
) bool {
	if !finite(left) || !finite(right) {
		return false
	}

	difference := math.Abs(left - right)
	scale := math.Max(
		1,
		math.Max(math.Abs(left), math.Abs(right)),
	)

	return difference <= tolerance*scale
}
