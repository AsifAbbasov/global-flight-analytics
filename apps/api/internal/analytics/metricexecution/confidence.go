package metricexecution

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/confidencereport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const (
	factorCodeEligibleContributorCoverage = "eligible_contributor_coverage"
)

func buildTrajectoryConfidence(
	allowed []trajectory.FlightTrajectory,
	inputCount int,
	evaluatedAt time.Time,
	capability trajectoryeligibility.Capability,
) (
	[]confidencereport.Factor,
	[]analyticalresult.Notice,
) {
	if inputCount == 0 {
		return []confidencereport.Factor{
				confidencereport.Evidence(
					confidencereport.
						FactorCodeMethodStability,
					0.25,
					1,
					"The metric formula is deterministic.",
				),
				confidencereport.Evidence(
					factorCodeEligibleContributorCoverage,
					0.75,
					0,
					"No trajectory observations were available.",
				),
			},
			[]analyticalresult.Notice{
				{
					Code:    NoticeCodeNoTrajectoryObservations,
					Message: "No trajectory observations were available; the published value is based on an empty observation set.",
				},
			}
	}

	qualityTotal := 0.0
	identityTotal := 0.0
	freshnessTotal := 0.0
	coverageTotal := 0.0
	gapPenaltyTotal := 0.0
	futureObservationCount := 0

	maximumAge := 5 * time.Minute
	pointTarget := 5.0

	if capability ==
		trajectoryeligibility.
			CapabilityAirportActivity {
		maximumAge = 15 * time.Minute
		pointTarget = 3
	}

	for _, item := range allowed {
		qualityTotal += clampUnit(
			item.QualityScore,
		)
		identityTotal +=
			identityReliability(
				item,
			)

		freshness, future :=
			trajectoryFreshness(
				item,
				evaluatedAt,
				maximumAge,
			)
		freshnessTotal += freshness
		if future {
			futureObservationCount++
		}

		pointCount := item.PointCount
		if len(item.Points) > pointCount {
			pointCount = len(item.Points)
		}

		coverageTotal += clampUnit(
			float64(pointCount) /
				pointTarget,
		)

		gapCount := item.CoverageGapCount
		if len(item.CoverageGaps) >
			gapCount {
			gapCount =
				len(item.CoverageGaps)
		}

		gapPenaltyTotal += clampUnit(
			float64(gapCount) / 5,
		)
	}

	allowedCount := len(allowed)
	divisor := float64(allowedCount)
	if allowedCount == 0 {
		divisor = 1
	}

	factors := []confidencereport.Factor{
		confidencereport.Evidence(
			confidencereport.
				FactorCodeTrajectoryQuality,
			0.30,
			qualityTotal/divisor,
			"Trajectory quality supports the metric result.",
		),
		confidencereport.Evidence(
			confidencereport.
				FactorCodeIdentityReliability,
			0.20,
			identityTotal/divisor,
			"Flight identity reliability supports contributor attribution.",
		),
		confidencereport.Evidence(
			confidencereport.
				FactorCodeDataFreshness,
			0.20,
			freshnessTotal/divisor,
			"Observation freshness supports the metric result.",
		),
		confidencereport.Evidence(
			confidencereport.
				FactorCodeObservationCoverage,
			0.15,
			coverageTotal/divisor,
			"Trajectory point coverage supports the metric result.",
		),
		confidencereport.Evidence(
			factorCodeEligibleContributorCoverage,
			0.15,
			float64(allowedCount)/
				float64(inputCount),
			"Eligible contributor coverage supports the metric result.",
		),
	}

	if gapPenaltyTotal > 0 {
		factors = append(
			factors,
			confidencereport.Penalty(
				confidencereport.
					FactorCodeCoverageGapPenalty,
				0.10,
				gapPenaltyTotal/divisor,
				"Coverage gaps reduce confidence in the aggregate metric.",
			),
		)
	}

	limitations := make(
		[]analyticalresult.Notice,
		0,
		1,
	)

	if futureObservationCount > 0 {
		limitations = append(
			limitations,
			analyticalresult.Notice{
				Code:    NoticeCodeFutureObservationTime,
				Message: "One or more trajectory end times are later than the evaluation time and do not contribute freshness confidence.",
			},
		)
	}

	return factors, limitations
}

func identityReliability(
	item trajectory.FlightTrajectory,
) float64 {
	if strings.TrimSpace(
		item.IdentityKey,
	) == "" {
		return 0
	}

	switch item.IdentityBasis {
	case trajectory.FlightIdentityBasisSourceFlightID,
		trajectory.FlightIdentityBasisCallsignAndStartTime:
		return 1

	case trajectory.FlightIdentityBasisAircraftAndStartTime:
		return 0.50

	default:
		return 0.25
	}
}

func trajectoryFreshness(
	item trajectory.FlightTrajectory,
	evaluatedAt time.Time,
	maximumAge time.Duration,
) (float64, bool) {
	if item.EndTime.IsZero() ||
		maximumAge <= 0 {
		return 0, false
	}

	endTime := item.EndTime.UTC()
	referenceTime := evaluatedAt.UTC()

	if endTime.After(referenceTime) {
		return 0, true
	}

	age := referenceTime.Sub(endTime)
	if age >= maximumAge {
		return 0, false
	}

	return 1 -
			float64(age)/
				float64(maximumAge),
		false
}

func methodConfidenceFactors(
	message string,
) []confidencereport.Factor {
	return []confidencereport.Factor{
		confidencereport.Evidence(
			confidencereport.
				FactorCodeMethodStability,
			0.70,
			1,
			message,
		),
		confidencereport.Evidence(
			confidencereport.
				FactorCodeSourceCoverage,
			0.30,
			1,
			"Required metric inputs are present and valid.",
		),
	}
}

func mergeNotices(
	collections ...[]analyticalresult.Notice,
) []analyticalresult.Notice {
	byCode := make(
		map[string]analyticalresult.Notice,
	)

	for _, collection := range collections {
		for _, notice := range collection {
			if _, exists := byCode[notice.Code]; exists {
				continue
			}

			byCode[notice.Code] = notice
		}
	}

	result := make(
		[]analyticalresult.Notice,
		0,
		len(byCode),
	)

	for _, notice := range byCode {
		result = append(
			result,
			notice,
		)
	}

	sort.SliceStable(
		result,
		func(
			left int,
			right int,
		) bool {
			return result[left].Code <
				result[right].Code
		},
	)

	return result
}

func clampUnit(
	value float64,
) float64 {
	if math.IsNaN(value) ||
		math.IsInf(value, 0) {
		return 0
	}

	return math.Max(
		0,
		math.Min(
			1,
			value,
		),
	)
}
