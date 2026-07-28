package historicalroute

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
)

func routeLimitations(
	metricName historicalcontract.MetricName,
	snapshot historicalread.Snapshot,
) []historicalcontract.Limitation {
	limitations := []historicalcontract.Limitation{
		{
			Code:    "probable_route_intelligence_only",
			Message: "Historical route metrics are derived from probable Route Intelligence results rather than filed flight-plan data.",
			Scope:   "series",
		},
	}

	switch metricName {
	case historicalcontract.MetricNameActiveRoutes:
		limitations = append(limitations, historicalcontract.Limitation{
			Code:    "active_routes_are_unique_directional_pairs",
			Message: "Active routes count unique complete origin-destination airport pairs per bucket; sample count records the validated route results supporting that distinct count.",
			Scope:   "metric",
		})
	case historicalcontract.MetricNameRouteConfidence:
		limitations = append(limitations, historicalcontract.Limitation{
			Code:    "route_confidence_is_unweighted_result_mean",
			Message: "Route confidence is the compensated arithmetic mean of validated result-level confidence scores, including unavailable results with their contract-defined zero score; evidence quality is already represented inside each result score.",
			Scope:   "metric",
		})
	case historicalcontract.MetricNameCompleteRouteRatio,
		historicalcontract.MetricNamePartialRouteRatio,
		historicalcontract.MetricNameUnavailableRouteRatio:
		limitations = append(limitations, historicalcontract.Limitation{
			Code:    "route_status_ratio_uses_validated_result_denominator",
			Message: "Route status ratios use all validated latest route results in each global bucket as the denominator and are not defined for a complete airport-pair scope.",
			Scope:   "metric",
		})
	case historicalcontract.MetricNameGreatCircleDistanceKM:
		limitations = append(limitations, historicalcontract.Limitation{
			Code:    "route_distance_recomputed_from_coordinates",
			Message: "Great-circle distance is recomputed from validated endpoint coordinates for complete routes instead of trusting the persisted summary value.",
			Scope:   "metric",
		})
	}

	if snapshot.RouteLimitReached {
		limitations = append(limitations, historicalcontract.Limitation{
			Code:    "historical_route_dataset_limit_reached",
			Message: "The bounded historical route read reached its dataset limit; global represented coverage uses the exact matched-route denominator.",
			Scope:   "series",
		})
	}
	if snapshot.RouteByteLimitReached {
		limitations = append(limitations, historicalcontract.Limitation{
			Code:    "historical_route_payload_byte_limit_reached",
			Message: "The bounded historical route read reached its payload byte budget; global unrepresented routes remain explicitly incomplete.",
			Scope:   "series",
		})
	}
	return limitations
}
