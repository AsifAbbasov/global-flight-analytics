package projectionbaseline

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"

type HorizontalFallbackPolicy string

const (
	HorizontalFallbackAllowLimited HorizontalFallbackPolicy = "allow_limited"
	HorizontalFallbackReject       HorizontalFallbackPolicy = "reject"
)

func (policy HorizontalFallbackPolicy) IsKnown() bool {
	switch policy {
	case HorizontalFallbackAllowLimited,
		HorizontalFallbackReject:
		return true
	default:
		return false
	}
}

func (config Config) effectiveHorizontalFallbackPolicy() HorizontalFallbackPolicy {
	if config.HorizontalFallbackPolicy == "" {
		return HorizontalFallbackAllowLimited
	}
	return config.HorizontalFallbackPolicy
}

func (policy HorizontalFallbackPolicy) allows(
	decision trajectoryeligibility.Decision,
) bool {
	return policy == HorizontalFallbackAllowLimited &&
		!decision.Allowed &&
		len(decision.Reasons) == 1 &&
		decision.Reasons[0] == trajectoryeligibility.ReasonMissingAltitude
}
