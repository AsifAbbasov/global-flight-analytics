package weathertrust

import (
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
)

var (
	ErrInvalidPolicy = errors.New("weather trust policy is invalid")
	ErrResultInvalid = errors.New("weather trust result is invalid")
)

func Evaluate(input weathercontract.Result, policy Policy) (Result, error) {
	if err := policy.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrInvalidPolicy, err)
	}

	limitations := make([]Notice, 0)
	explanations := []Notice{
		{
			Code:    "weather_context_only",
			Message: "Weather trust controls contextual use and does not prove pilot intent, controller intent, rerouting reason, or maneuver cause.",
		},
	}

	contractReport := weathercontract.Validate(input)
	if contractReport.Status != weathercontract.ValidationStatusValid {
		for _, issue := range contractReport.Issues {
			limitations = append(limitations, Notice{
				Code:    "contract_" + issue.Code,
				Message: issue.Message,
			})
		}
		return finalize(
			input,
			policy,
			DecisionBlocked,
			false,
			0,
			0,
			0,
			0,
			nil,
			limitations,
			append(explanations, Notice{
				Code:    "weather_contract_invalid",
				Message: "Weather evidence was blocked because the canonical contract is invalid.",
			}),
		)
	}

	if input.Status == weathercontract.ResultStatusUnavailable {
		limitations = append(limitations, Notice{
			Code:    "weather_unavailable",
			Message: "The canonical weather result is unavailable.",
		})
		return finalize(
			input,
			policy,
			DecisionBlocked,
			false,
			input.Confidence.Score,
			0,
			0,
			0,
			nil,
			limitations,
			append(explanations, Notice{
				Code:    "weather_use_blocked",
				Message: "Unavailable weather evidence cannot be used by Weather Intelligence.",
			}),
		)
	}

	confidenceScore := clampUnit(input.Confidence.Score)
	freshnessScore, temporalBlocked, temporalNotices := evaluateFreshness(input, policy)
	limitations = append(limitations, temporalNotices...)
	completenessScore, featureBlocked, featureNotices := evaluateCompleteness(input, policy)
	limitations = append(limitations, featureNotices...)
	verticalScore, scopes, verticalLimited, verticalNotices := evaluateVertical(input)
	limitations = append(limitations, verticalNotices...)

	confidenceBlocked := confidenceScore < policy.MinimumUsableConfidence
	if confidenceBlocked {
		limitations = append(limitations, Notice{
			Code:    "weather_confidence_below_usable_minimum",
			Message: "Weather confidence is below the minimum required for analytical use.",
		})
	}

	components := policy.components(
		confidenceScore,
		freshnessScore,
		completenessScore,
		verticalScore,
	)
	score := weightedScore(components)
	scoreBlocked := score < policy.MinimumUsableScore
	if scoreBlocked {
		limitations = append(limitations, Notice{
			Code:    "weather_trust_score_below_usable_minimum",
			Message: "The combined weather trust score is below the minimum required for analytical use.",
		})
	}

	if temporalBlocked || featureBlocked || confidenceBlocked || scoreBlocked {
		return finalizeWithComponents(
			input,
			policy,
			DecisionBlocked,
			false,
			score,
			components,
			nil,
			limitations,
			append(explanations, Notice{
				Code:    "weather_use_blocked",
				Message: "Weather evidence failed one or more mandatory trust conditions.",
			}),
		)
	}

	limited := input.Status != weathercontract.ResultStatusComplete ||
		confidenceScore < policy.MinimumAllowedConfidence ||
		score < policy.MinimumAllowedScore ||
		verticalLimited ||
		len(input.Limitations) > 0

	if input.Status != weathercontract.ResultStatusComplete {
		limitations = append(limitations, Notice{
			Code:    "weather_contract_not_complete",
			Message: "The canonical weather result is limited and cannot receive a fully allowed trust decision.",
		})
	}
	if confidenceScore < policy.MinimumAllowedConfidence {
		limitations = append(limitations, Notice{
			Code:    "weather_confidence_below_allowed_minimum",
			Message: "Weather confidence supports limited use but not unrestricted contextual use.",
		})
	}
	if score < policy.MinimumAllowedScore {
		limitations = append(limitations, Notice{
			Code:    "weather_trust_score_below_allowed_minimum",
			Message: "The combined weather trust score supports limited use only.",
		})
	}
	for _, limitation := range input.Limitations {
		limitations = append(limitations, Notice{
			Code:    "contract_" + limitation.Code,
			Message: limitation.Message,
		})
	}

	if limited {
		return finalizeWithComponents(
			input,
			policy,
			DecisionLimited,
			true,
			score,
			components,
			scopes,
			limitations,
			append(explanations, Notice{
				Code:    "weather_use_limited",
				Message: "Weather evidence may be used only within the published scopes and limitations.",
			}),
		)
	}

	return finalizeWithComponents(
		input,
		policy,
		DecisionAllowed,
		true,
		score,
		components,
		scopes,
		limitations,
		append(explanations, Notice{
			Code:    "weather_use_allowed",
			Message: "Weather evidence satisfies the production trust policy for its published scopes.",
		}),
	)
}

func evaluateFreshness(input weathercontract.Result, policy Policy) (float64, bool, []Notice) {
	if len(input.Samples) == 0 {
		return 0, true, []Notice{
			{
				Code:    "weather_samples_missing",
				Message: "Weather trust evaluation requires at least one sample.",
			},
		}
	}

	minimumScore := 1.0
	blocked := false
	notices := make([]Notice, 0)
	for _, sample := range input.Samples {
		score := 0.0
		sampleBlocked := false
		switch sample.Source.EvidenceKind {
		case weathercontract.EvidenceKindObservation:
			score, sampleBlocked = freshnessForPastEvidence(
				input.AsOfTime,
				sample.ValidAt,
				policy.MaximumObservationAge,
			)
		case weathercontract.EvidenceKindAnalysis:
			score, sampleBlocked = freshnessForPastEvidence(
				input.AsOfTime,
				sample.ValidAt,
				policy.MaximumAnalysisAge,
			)
		case weathercontract.EvidenceKindForecast:
			if sample.ValidAt.After(input.AsOfTime) {
				lead := sample.ValidAt.Sub(input.AsOfTime)
				if lead > policy.MaximumForecastLead {
					sampleBlocked = true
					score = 0
				} else {
					score = 1 - float64(lead)/float64(policy.MaximumForecastLead)
				}
			} else {
				score, sampleBlocked = freshnessForPastEvidence(
					input.AsOfTime,
					sample.ValidAt,
					policy.MaximumAnalysisAge,
				)
			}
		default:
			sampleBlocked = true
			score = 0
		}
		if sampleBlocked {
			blocked = true
		}
		if score < minimumScore {
			minimumScore = score
		}
	}

	if blocked {
		notices = append(notices, Notice{
			Code:    "weather_temporal_boundary_exceeded",
			Message: "At least one weather sample exceeds the production age or forecast-lead boundary.",
		})
	} else if minimumScore < 0.50 {
		notices = append(notices, Notice{
			Code:    "weather_evidence_aging",
			Message: "Weather evidence remains usable but is approaching its temporal boundary.",
		})
	}
	return clampUnit(minimumScore), blocked, notices
}

func freshnessForPastEvidence(
	asOfTime time.Time,
	validAt time.Time,
	maximumAge time.Duration,
) (float64, bool) {
	age := asOfTime.Sub(validAt)
	if age < 0 || age > maximumAge {
		return 0, true
	}
	return clampUnit(1 - float64(age)/float64(maximumAge)), false
}

func evaluateCompleteness(input weathercontract.Result, policy Policy) (float64, bool, []Notice) {
	if len(input.Samples) == 0 {
		return 0, true, nil
	}
	total := 0
	minimum := int(^uint(0) >> 1)
	for _, sample := range input.Samples {
		count := sample.Features.PresentCount()
		total += count
		if count < minimum {
			minimum = count
		}
	}
	average := float64(total) / float64(len(input.Samples))
	score := average / float64(policy.TargetFeatureCount)
	blocked := minimum < policy.MinimumFeatureCount
	notices := make([]Notice, 0)
	if blocked {
		notices = append(notices, Notice{
			Code:    "weather_features_insufficient",
			Message: "At least one weather sample contains fewer fields than the minimum trust policy permits.",
		})
	} else if average < float64(policy.TargetFeatureCount) {
		notices = append(notices, Notice{
			Code:    "weather_features_partial",
			Message: "Weather samples contain enough fields for limited use but do not reach the target feature count.",
		})
	}
	return clampUnit(score), blocked, notices
}

func evaluateVertical(input weathercontract.Result) (float64, []UsageScope, bool, []Notice) {
	if len(input.Samples) == 0 {
		return 0, nil, true, nil
	}
	totalScore := 0.0
	allFlightLevelApplicable := true
	hasSurface := false
	notices := make([]Notice, 0)

	for _, sample := range input.Samples {
		switch sample.Position.VerticalReference {
		case weathercontract.VerticalReferencePressureLevel:
			totalScore += 1
		case weathercontract.VerticalReferenceMeanSeaLevel:
			if sample.Position.AltitudeMeters != nil {
				totalScore += 1
			} else {
				totalScore += 0.50
				allFlightLevelApplicable = false
			}
		case weathercontract.VerticalReferenceSurface:
			totalScore += 0.35
			hasSurface = true
			allFlightLevelApplicable = false
		case weathercontract.VerticalReferenceUnknown:
			totalScore += 0.15
			allFlightLevelApplicable = false
		default:
			allFlightLevelApplicable = false
		}
	}

	scopes := make([]UsageScope, 0)
	if hasSurface {
		scopes = append(scopes, UsageScopeSurfaceContext)
	}
	if allFlightLevelApplicable {
		scopes = append(scopes, UsageScopeTrajectoryContext, UsageScopeProjectionUncertainty)
	}
	if len(scopes) == 0 {
		scopes = append(scopes, UsageScopeSurfaceContext)
	}

	limited := !allFlightLevelApplicable
	if hasSurface {
		notices = append(notices, Notice{
			Code:    "surface_weather_not_flight_level",
			Message: "Surface weather may be shown as context but must not be treated as weather at aircraft altitude.",
		})
	} else if limited {
		notices = append(notices, Notice{
			Code:    "vertical_weather_reference_incomplete",
			Message: "Weather vertical applicability is incomplete, so flight-level use is withheld.",
		})
	}

	return clampUnit(totalScore / float64(len(input.Samples))), normalizeScopes(scopes), limited, notices
}

func finalize(
	input weathercontract.Result,
	policy Policy,
	decision Decision,
	usable bool,
	confidenceScore float64,
	freshnessScore float64,
	completenessScore float64,
	verticalScore float64,
	scopes []UsageScope,
	limitations []Notice,
	explanations []Notice,
) (Result, error) {
	components := policy.components(confidenceScore, freshnessScore, completenessScore, verticalScore)
	return finalizeWithComponents(
		input,
		policy,
		decision,
		usable,
		weightedScore(components),
		components,
		scopes,
		limitations,
		explanations,
	)
}

func finalizeWithComponents(
	input weathercontract.Result,
	policy Policy,
	decision Decision,
	usable bool,
	score float64,
	components []Component,
	scopes []UsageScope,
	limitations []Notice,
	explanations []Notice,
) (Result, error) {
	result := Result{
		Version:          Version,
		Decision:         decision,
		Usable:           usable,
		AsOfTime:         input.AsOfTime.UTC(),
		Score:            clampUnit(score),
		Components:       append([]Component(nil), components...),
		AllowedScopes:    normalizeScopes(scopes),
		Limitations:      normalizeNotices(limitations),
		Explanations:     normalizeNotices(explanations),
		InputFingerprint: inputFingerprint(input, policy),
	}
	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrResultInvalid, err)
	}
	return result.Clone(), nil
}
