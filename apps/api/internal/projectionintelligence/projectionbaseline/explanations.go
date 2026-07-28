package projectionbaseline

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"

func baselineExplanations(onGround bool) []projectioncontract.Explanation {
	if onGround {
		return []projectioncontract.Explanation{
			{
				Code:    "stationary_on_ground_propagation",
				Message: "An allowed on-ground observation is projected as stationary; reported taxi speed, heading, and vertical rate are not propagated.",
			},
			{
				Code:    "explicit_uncertainty_growth",
				Message: "Horizontal and vertical uncertainty continue to grow across the forecast horizon.",
			},
		}
	}

	return []projectioncontract.Explanation{
		{
			Code:    "constant_ground_track_propagation",
			Message: "Each forecast point propagates the latest observed ground speed and heading over a spherical direct-geodesic step.",
		},
		{
			Code:    "linear_vertical_rate_propagation",
			Message: "When altitude is available, the selected geometric or barometric altitude reference is propagated using the latest observed vertical rate.",
		},
		{
			Code:    "explicit_uncertainty_growth",
			Message: "Horizontal and vertical uncertainty grow from caller-provided baseline values and rates.",
		},
	}
}
