package dataqualityintegration

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
)

const (
	PermissionReasonHistoricalSimilarityRequiresTwoTrajectories = "historical_similarity_requires_two_trajectories"
	PermissionReasonHistoricalSimilarityNoComparablePair        = "historical_similarity_no_comparable_pair"

	LimitationCodeHistoricalSimilarityUnavailable = "historical_similarity_unavailable"
	LimitationCodeHistoricalSimilarityPartial     = "historical_similarity_partial_evidence"

	DefaultMaximumSimilarityPairChecks = 1_000
)

func evaluateHistoricalSimilarity(
	items []trajectory.FlightTrajectory,
) (
	dataqualitycontract.Permission,
	[]dataqualitycontract.Notice,
	error,
) {
	if len(items) < 2 {
		permission, err := dataqualitycontract.DeniedPermission(
			PermissionReasonHistoricalSimilarityRequiresTwoTrajectories,
		)
		if err != nil {
			return dataqualitycontract.Permission{},
				nil,
				err
		}

		return permission,
			[]dataqualitycontract.Notice{
				{
					Code:    LimitationCodeHistoricalSimilarityUnavailable,
					Message: "Historical trajectory similarity requires at least two retained trajectories.",
				},
			},
			nil
	}

	engine := historicalsimilarity.NewDefault()
	checkedPairCount := 0
	rejectedPairCount := 0
	truncated := false

	for left := 0; left < len(items); left++ {
		for right := left + 1; right < len(items); right++ {
			if checkedPairCount >=
				DefaultMaximumSimilarityPairChecks {
				truncated = true
				break
			}
			checkedPairCount++

			_, err := engine.Compare(
				items[left],
				items[right],
			)
			if err == nil {
				limitations := make(
					[]dataqualitycontract.Notice,
					0,
					1,
				)
				if rejectedPairCount > 0 ||
					truncated {
					limitations = append(
						limitations,
						dataqualitycontract.Notice{
							Code: LimitationCodeHistoricalSimilarityPartial,
							Message: fmt.Sprintf(
								"Historical similarity found a comparable trajectory pair after checking %d pairs; %d earlier pairs were not comparable and pair evaluation truncation=%t.",
								checkedPairCount,
								rejectedPairCount,
								truncated,
							),
						},
					)
				}

				return dataqualitycontract.
						AllowedPermission(),
					limitations,
					nil
			}
			rejectedPairCount++
		}
		if truncated {
			break
		}
	}

	permission, err := dataqualitycontract.DeniedPermission(
		PermissionReasonHistoricalSimilarityNoComparablePair,
	)
	if err != nil {
		return dataqualitycontract.Permission{},
			nil,
			err
	}

	return permission,
		[]dataqualitycontract.Notice{
			{
				Code: LimitationCodeHistoricalSimilarityUnavailable,
				Message: fmt.Sprintf(
					"Historical similarity checked %d trajectory pairs but found no pair with enough usable chronological coordinates; pair evaluation truncation=%t.",
					checkedPairCount,
					truncated,
				),
			},
		},
		nil
}
