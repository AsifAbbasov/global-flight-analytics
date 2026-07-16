package weathertrust

import (
	"fmt"
	"time"
)

const PolicyVersionV1 = "weather-trust-policy-v1"

type Weights struct {
	ContractConfidence    float64
	TemporalFreshness     float64
	FeatureCompleteness   float64
	VerticalApplicability float64
}

type Policy struct {
	Version string

	MaximumObservationAge time.Duration
	MaximumAnalysisAge    time.Duration
	MaximumForecastLead   time.Duration

	MinimumFeatureCount int
	TargetFeatureCount  int

	MinimumUsableConfidence  float64
	MinimumAllowedConfidence float64
	MinimumUsableScore       float64
	MinimumAllowedScore      float64

	Weights Weights
}

func DefaultPolicy() Policy {
	return Policy{
		Version: PolicyVersionV1,

		MaximumObservationAge: 45 * time.Minute,
		MaximumAnalysisAge:    2 * time.Hour,
		MaximumForecastLead:   6 * time.Hour,

		MinimumFeatureCount: 3,
		TargetFeatureCount:  8,

		MinimumUsableConfidence:  0.35,
		MinimumAllowedConfidence: 0.70,
		MinimumUsableScore:       0.40,
		MinimumAllowedScore:      0.75,

		Weights: Weights{
			ContractConfidence:    0.35,
			TemporalFreshness:     0.30,
			FeatureCompleteness:   0.20,
			VerticalApplicability: 0.15,
		},
	}
}

func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 {
		return fmt.Errorf("weather trust policy version is invalid")
	}
	if policy.MaximumObservationAge <= 0 ||
		policy.MaximumAnalysisAge <= 0 ||
		policy.MaximumForecastLead <= 0 {
		return fmt.Errorf("weather trust policy time limits must be greater than zero")
	}
	if policy.MinimumFeatureCount <= 0 ||
		policy.TargetFeatureCount < policy.MinimumFeatureCount {
		return fmt.Errorf("weather trust feature thresholds are invalid")
	}
	for name, value := range map[string]float64{
		"minimum usable confidence":  policy.MinimumUsableConfidence,
		"minimum allowed confidence": policy.MinimumAllowedConfidence,
		"minimum usable score":       policy.MinimumUsableScore,
		"minimum allowed score":      policy.MinimumAllowedScore,
	} {
		if !unitInterval(value) {
			return fmt.Errorf("weather trust %s is invalid", name)
		}
	}
	if policy.MinimumAllowedConfidence < policy.MinimumUsableConfidence ||
		policy.MinimumAllowedScore < policy.MinimumUsableScore {
		return fmt.Errorf("weather trust allowed thresholds must not be below usable thresholds")
	}

	weightTotal := policy.Weights.ContractConfidence +
		policy.Weights.TemporalFreshness +
		policy.Weights.FeatureCompleteness +
		policy.Weights.VerticalApplicability
	if !finite(policy.Weights.ContractConfidence) ||
		!finite(policy.Weights.TemporalFreshness) ||
		!finite(policy.Weights.FeatureCompleteness) ||
		!finite(policy.Weights.VerticalApplicability) ||
		policy.Weights.ContractConfidence < 0 ||
		policy.Weights.TemporalFreshness < 0 ||
		policy.Weights.FeatureCompleteness < 0 ||
		policy.Weights.VerticalApplicability < 0 ||
		absolute(weightTotal-1) > 1e-9 {
		return fmt.Errorf("weather trust policy weights are invalid")
	}
	return nil
}

func (policy Policy) components(
	confidenceScore float64,
	freshnessScore float64,
	completenessScore float64,
	verticalScore float64,
) []Component {
	return []Component{
		{
			Name:   ComponentContractConfidence,
			Score:  clampUnit(confidenceScore),
			Weight: policy.Weights.ContractConfidence,
		},
		{
			Name:   ComponentTemporalFreshness,
			Score:  clampUnit(freshnessScore),
			Weight: policy.Weights.TemporalFreshness,
		},
		{
			Name:   ComponentFeatureCompleteness,
			Score:  clampUnit(completenessScore),
			Weight: policy.Weights.FeatureCompleteness,
		},
		{
			Name:   ComponentVerticalApplicability,
			Score:  clampUnit(verticalScore),
			Weight: policy.Weights.VerticalApplicability,
		},
	}
}

func weightedScore(components []Component) float64 {
	total := 0.0
	for _, component := range components {
		total += component.Score * component.Weight
	}
	return clampUnit(total)
}
