package weatheruncertainty

import "fmt"

const PolicyVersionV1 = "weather-adjusted-uncertainty-policy-v1"

type Weights struct {
	WindSpeed       float64
	WindGust        float64
	Precipitation   float64
	CloudCover      float64
	EvidenceQuality float64
}

type Policy struct {
	Version string

	MaximumUncertaintyMultiplier float64
	MaximumConfidenceReduction   float64
	NearTermEffectFraction       float64

	WindSpeedReferenceMetersPerSecond float64
	WindSpeedHighMetersPerSecond      float64
	WindGustReferenceMetersPerSecond  float64
	WindGustHighMetersPerSecond       float64

	PrecipitationReferenceMillimeters float64
	PrecipitationHighMillimeters      float64

	CloudCoverReferencePercent float64
	CloudCoverHighPercent      float64

	Weights Weights
}

func DefaultPolicy() Policy {
	return Policy{
		Version:                      PolicyVersionV1,
		MaximumUncertaintyMultiplier: 2.50,
		MaximumConfidenceReduction:   0.30,
		NearTermEffectFraction:       0.50,

		WindSpeedReferenceMetersPerSecond: 12,
		WindSpeedHighMetersPerSecond:      35,
		WindGustReferenceMetersPerSecond:  18,
		WindGustHighMetersPerSecond:       50,

		PrecipitationReferenceMillimeters: 0.50,
		PrecipitationHighMillimeters:      5,

		CloudCoverReferencePercent: 40,
		CloudCoverHighPercent:      100,

		Weights: Weights{
			WindSpeed:       0.30,
			WindGust:        0.20,
			Precipitation:   0.15,
			CloudCover:      0.10,
			EvidenceQuality: 0.25,
		},
	}
}

func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 {
		return fmt.Errorf("weather uncertainty policy version is invalid")
	}
	if !finite(policy.MaximumUncertaintyMultiplier) ||
		policy.MaximumUncertaintyMultiplier < 1 ||
		!unitInterval(policy.MaximumConfidenceReduction) ||
		!unitInterval(policy.NearTermEffectFraction) {
		return fmt.Errorf("weather uncertainty policy output limits are invalid")
	}

	thresholds := []struct {
		name      string
		reference float64
		high      float64
	}{
		{
			name:      "wind speed",
			reference: policy.WindSpeedReferenceMetersPerSecond,
			high:      policy.WindSpeedHighMetersPerSecond,
		},
		{
			name:      "wind gust",
			reference: policy.WindGustReferenceMetersPerSecond,
			high:      policy.WindGustHighMetersPerSecond,
		},
		{
			name:      "precipitation",
			reference: policy.PrecipitationReferenceMillimeters,
			high:      policy.PrecipitationHighMillimeters,
		},
		{
			name:      "cloud cover",
			reference: policy.CloudCoverReferencePercent,
			high:      policy.CloudCoverHighPercent,
		},
	}

	for _, threshold := range thresholds {
		if !finite(threshold.reference) ||
			!finite(threshold.high) ||
			threshold.reference < 0 ||
			threshold.high <= threshold.reference {
			return fmt.Errorf("weather uncertainty %s thresholds are invalid", threshold.name)
		}
	}

	weights := []float64{
		policy.Weights.WindSpeed,
		policy.Weights.WindGust,
		policy.Weights.Precipitation,
		policy.Weights.CloudCover,
		policy.Weights.EvidenceQuality,
	}
	weightTotal := 0.0
	for _, weight := range weights {
		if !finite(weight) || weight < 0 {
			return fmt.Errorf("weather uncertainty policy weight is invalid")
		}
		weightTotal += weight
	}
	if absolute(weightTotal-1) > 1e-9 {
		return fmt.Errorf("weather uncertainty policy weights must sum to one")
	}
	return nil
}

func (policy Policy) components(
	windSpeedScore float64,
	windGustScore float64,
	precipitationScore float64,
	cloudCoverScore float64,
	evidenceQualityScore float64,
) []Component {
	return []Component{
		{
			Name:   ComponentWindSpeed,
			Score:  clampUnit(windSpeedScore),
			Weight: policy.Weights.WindSpeed,
		},
		{
			Name:   ComponentWindGust,
			Score:  clampUnit(windGustScore),
			Weight: policy.Weights.WindGust,
		},
		{
			Name:   ComponentPrecipitation,
			Score:  clampUnit(precipitationScore),
			Weight: policy.Weights.Precipitation,
		},
		{
			Name:   ComponentCloudCover,
			Score:  clampUnit(cloudCoverScore),
			Weight: policy.Weights.CloudCover,
		},
		{
			Name:   ComponentEvidenceQuality,
			Score:  clampUnit(evidenceQualityScore),
			Weight: policy.Weights.EvidenceQuality,
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
