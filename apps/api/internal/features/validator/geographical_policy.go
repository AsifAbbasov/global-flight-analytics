package validator

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func validateGeographicalLongitudeEnvelope(
	collector *issueCollector,
	severity IssueSeverity,
	item flightfeatures.GeographicalFeatures,
) {
	if !finite(item.MinimumLongitude) ||
		!finite(item.MaximumLongitude) ||
		!finite(item.LongitudeSpanDegrees) {
		return
	}

	expectedSpan := flightfeatures.CircularLongitudeSpanDegrees(
		item.MinimumLongitude,
		item.MaximumLongitude,
	)
	if approximatelyEqual(
		item.LongitudeSpanDegrees,
		expectedSpan,
		collector.tolerance,
	) {
		return
	}

	addBySeverity(
		collector,
		severity,
		flightfeatures.FeatureGroupGeographical,
		"geographical.longitude_span_degrees",
		issueCodePrefix+"longitude_span_mismatch",
		fmt.Sprintf(
			"Longitude span %.6f degrees does not match the circular west/east envelope span %.6f degrees.",
			item.LongitudeSpanDegrees,
			expectedSpan,
		),
	)
}

func validateGeographicalDistanceRelationships(
	collector *issueCollector,
	severity IssueSeverity,
	item flightfeatures.GeographicalFeatures,
) {
	segmentFallback := flightfeatures.HasLimitationCode(
		item.Evidence.Limitations,
		flightfeatures.GeographicalLimitationSegmentEndpointFallback,
	)
	if segmentFallback {
		// Segment fallback deliberately excludes unobserved discontinuities from
		// path length. Endpoint distance or displacement may therefore exceed
		// the sum of movement observed inside individual segments.
		return
	}

	if finite(item.GreatCircleDistanceKM) &&
		finite(item.ObservedPathDistanceKM) &&
		item.GreatCircleDistanceKM >
			item.ObservedPathDistanceKM+collector.tolerance {
		addBySeverity(
			collector,
			severity,
			flightfeatures.FeatureGroupGeographical,
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
			flightfeatures.FeatureGroupGeographical,
			"geographical.maximum_displacement_km",
			issueCodePrefix+"displacement_exceeds_path",
			"Maximum displacement exceeds observed path distance.",
		)
	}
}
