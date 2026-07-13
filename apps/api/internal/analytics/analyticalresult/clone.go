package analyticalresult

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"

func (result Result[T]) Clone() Result[T] {
	clone := result
	clone.Confidence = cloneConfidence(result.Confidence)
	clone.Eligibility = cloneEligibility(result.Eligibility)
	clone.Sources = cloneSources(result.Sources)
	clone.Warnings = cloneNotices(result.Warnings)
	clone.Limitations = cloneNotices(result.Limitations)

	if result.Failure != nil {
		failure := *result.Failure
		clone.Failure = &failure
	}

	return clone
}

func cloneConfidence(
	confidence Confidence,
) Confidence {
	result := confidence
	result.Reasons = cloneNotices(
		confidence.Reasons,
	)
	return result
}

func cloneEligibility(
	eligibility *Eligibility,
) *Eligibility {
	if eligibility == nil {
		return nil
	}

	result := *eligibility
	result.Reasons = append(
		[]trajectoryeligibility.ReasonCode(nil),
		eligibility.Reasons...,
	)

	return &result
}

func cloneSources(
	sources []Source,
) []Source {
	if sources == nil {
		return nil
	}

	result := make(
		[]Source,
		len(sources),
	)

	for index, source := range sources {
		result[index] = source
		result[index].Limitations = cloneNotices(
			source.Limitations,
		)
	}

	return result
}

func cloneNotices(
	notices []Notice,
) []Notice {
	return append(
		[]Notice(nil),
		notices...,
	)
}
