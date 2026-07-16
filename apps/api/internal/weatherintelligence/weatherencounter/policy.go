package weatherencounter

import "fmt"

const PolicyVersionV1 = "weather-encounter-profile-policy-v1"

type Policy struct {
	Version string

	MinimumCompleteProfileCoverage    float64
	MinimumCompleteCoreMetricCoverage float64
}

func DefaultPolicy() Policy {
	return Policy{
		Version: PolicyVersionV1,

		MinimumCompleteProfileCoverage:    0.95,
		MinimumCompleteCoreMetricCoverage: 0.75,
	}
}

func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 {
		return fmt.Errorf(
			"weather encounter policy version is invalid",
		)
	}
	if !unitInterval(
		policy.MinimumCompleteProfileCoverage,
	) ||
		policy.MinimumCompleteProfileCoverage <= 0 ||
		!unitInterval(
			policy.
				MinimumCompleteCoreMetricCoverage,
		) ||
		policy.
			MinimumCompleteCoreMetricCoverage <= 0 {
		return fmt.Errorf(
			"weather encounter policy thresholds are invalid",
		)
	}
	return nil
}

func (policy Policy) complete(
	alignmentStatus string,
	profileCoverage float64,
	temperatureCoverage float64,
	windSpeedCoverage float64,
	windDirectionCoverage float64,
) bool {
	return alignmentStatus == "complete" &&
		profileCoverage >=
			policy.MinimumCompleteProfileCoverage &&
		temperatureCoverage >=
			policy.
				MinimumCompleteCoreMetricCoverage &&
		windSpeedCoverage >=
			policy.
				MinimumCompleteCoreMetricCoverage &&
		windDirectionCoverage >=
			policy.
				MinimumCompleteCoreMetricCoverage
}
