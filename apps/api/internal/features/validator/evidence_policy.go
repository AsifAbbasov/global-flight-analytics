package validator

import (
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func collectCurrentGroupLimitations(
	features flightfeatures.FlightFeatures,
) []flightfeatures.FeatureLimitation {
	groups := []flightfeatures.GroupEvidence{
		features.Temporal.Evidence,
		features.Geographical.Evidence,
		features.Operational.Evidence,
		features.Trajectory.Evidence,
		features.Aircraft.Evidence,
	}
	result := make([]flightfeatures.FeatureLimitation, 0)
	seen := make(map[string]struct{})
	for _, evidence := range groups {
		for _, limitation := range stripValidatorLimitations(evidence.Limitations) {
			key := limitation.Code + "\x00" + limitation.Message
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, limitation)
		}
	}
	return result
}

func hasNonValidatorLimitation(
	limitations []flightfeatures.FeatureLimitation,
) bool {
	for _, limitation := range limitations {
		code := strings.TrimSpace(limitation.Code)
		message := strings.TrimSpace(limitation.Message)
		if code != "" && message != "" && !strings.HasPrefix(code, issueCodePrefix) {
			return true
		}
	}
	return false
}

func emitLimitationsAsWarnings(
	collector *issueCollector,
	group flightfeatures.FeatureGroup,
	path string,
	limitations []flightfeatures.FeatureLimitation,
) {
	for index, limitation := range limitations {
		code := strings.TrimSpace(limitation.Code)
		message := strings.TrimSpace(limitation.Message)
		if code == "" || message == "" || strings.HasPrefix(code, issueCodePrefix) {
			continue
		}
		collector.warning(
			group,
			path+"["+decimalIndex(index)+"]",
			code,
			message,
		)
	}
}

func decimalIndex(value int) string {
	if value == 0 {
		return "0"
	}
	digits := [20]byte{}
	position := len(digits)
	for value > 0 {
		position--
		digits[position] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[position:])
}
