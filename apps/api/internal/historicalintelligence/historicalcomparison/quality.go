package historicalcomparison

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func assembleComparedResult(
	current historicalcontract.Result,
	previous historicalcontract.Result,
	comparison historicalcontract.PeriodComparison,
) historicalcontract.Result {
	result := current.Clone()
	result.Comparison = &comparison
	result.Limitations = comparisonLimitations(
		current,
		previous,
	)
	result.Provenance = comparisonProvenance(
		current,
		previous,
	)
	return result
}

func comparisonLimitations(
	current historicalcontract.Result,
	previous historicalcontract.Result,
) []historicalcontract.Limitation {
	result := append(
		[]historicalcontract.Limitation(nil),
		current.Limitations...,
	)
	result = appendUniqueLimitation(
		result,
		historicalcontract.Limitation{
			Code: "historical_comparison_period_quality",
			Message: fmt.Sprintf(
				"Comparison quality binds both periods: current_status=%s current_score=%.12g current_samples=%d previous_status=%s previous_score=%.12g previous_samples=%d.",
				current.Status,
				current.Confidence.Score,
				current.Confidence.SampleCount,
				previous.Status,
				previous.Confidence.Score,
				previous.Confidence.SampleCount,
			),
			Scope: "comparison",
		},
	)

	if len(previous.Limitations) > 0 {
		codes := make(
			[]string,
			0,
			len(previous.Limitations),
		)
		for _, limitation := range previous.Limitations {
			codes = append(codes, limitation.Code)
		}
		sort.Strings(codes)
		result = appendUniqueLimitation(
			result,
			historicalcontract.Limitation{
				Code: "historical_comparison_previous_period_limitations",
				Message: "Previous-period limitations affecting this comparison: " +
					strings.Join(codes, ", ") + ".",
				Scope: "comparison",
			},
		)
	}

	if current.Status ==
		historicalcontract.SeriesStatusPartial {
		result = appendUniqueLimitation(
			result,
			historicalcontract.Limitation{
				Code:    "historical_comparison_matched_partial_coverage",
				Message: "Both periods have the same partial per-bucket coverage profile; the comparison remains coverage-limited.",
				Scope:   "comparison",
			},
		)
	}

	sort.SliceStable(
		result,
		func(left int, right int) bool {
			if result[left].Scope != result[right].Scope {
				return result[left].Scope <
					result[right].Scope
			}
			return result[left].Code <
				result[right].Code
		},
	)
	return result
}

func appendUniqueLimitation(
	values []historicalcontract.Limitation,
	value historicalcontract.Limitation,
) []historicalcontract.Limitation {
	for _, existing := range values {
		if existing.Scope == value.Scope &&
			existing.Code == value.Code {
			return values
		}
	}
	return append(values, value)
}
