package qualitypolicy

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"

const (
	StartingScore = 1.0
	MinimumScore  = 0.0
	MaximumScore  = 1.0

	MissingICAO24Penalty        = 0.50
	InvalidICAO24Penalty        = 0.50
	InvalidLatitudePenalty      = 0.35
	InvalidLongitudePenalty     = 0.35
	MissingObservedAtPenalty    = 0.40
	FutureObservedAtPenalty     = 0.40
	MissingCallsignPenalty      = 0.05
	MissingOriginCountryPenalty = 0.05
	MissingSourceNamePenalty    = 0.05
	InvalidAltitudePenalty      = 0.10
	InvalidVelocityPenalty      = 0.15
	InvalidVerticalRatePenalty  = 0.05
	InvalidHeadingPenalty       = 0.10

	HighConfidenceMinimumScore   = 0.85
	MediumConfidenceMinimumScore = 0.60
)

func ApplyPenalty(score float64, penalty float64) float64 {
	return score - penalty
}

func ClampScore(score float64) float64 {
	if score < MinimumScore {
		return MinimumScore
	}

	if score > MaximumScore {
		return MaximumScore
	}

	return score
}

func ConfidenceFromScore(score float64) dataquality.ConfidenceLevel {
	switch {
	case score >= HighConfidenceMinimumScore:
		return dataquality.ConfidenceLevelHigh
	case score >= MediumConfidenceMinimumScore:
		return dataquality.ConfidenceLevelMedium
	case score > MinimumScore:
		return dataquality.ConfidenceLevelLow
	default:
		return dataquality.ConfidenceLevelNone
	}
}
