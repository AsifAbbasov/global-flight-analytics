package historicalcomparison

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"

type summaryValueSelector func(
	historicalcontract.Summary,
) float64

var comparisonValueSelectors = map[historicalcontract.Aggregation]summaryValueSelector{
	historicalcontract.AggregationCount: func(
		summary historicalcontract.Summary,
	) float64 {
		return summary.Total
	},
	historicalcontract.AggregationSum: func(
		summary historicalcontract.Summary,
	) float64 {
		return summary.Total
	},
	historicalcontract.AggregationMinimum: func(
		summary historicalcontract.Summary,
	) float64 {
		return summary.Minimum
	},
	historicalcontract.AggregationMaximum: func(
		summary historicalcontract.Summary,
	) float64 {
		return summary.Maximum
	},
	historicalcontract.AggregationAverage: func(
		summary historicalcontract.Summary,
	) float64 {
		return summary.Average
	},
	historicalcontract.AggregationMedian: func(
		summary historicalcontract.Summary,
	) float64 {
		return summary.Median
	},
	historicalcontract.AggregationRatio: func(
		summary historicalcontract.Summary,
	) float64 {
		return summary.Average
	},
}

func selectPeriodValues(
	current historicalcontract.Result,
	previous historicalcontract.Result,
) (periodValues, error) {
	selector, exists := comparisonValueSelectors[current.Metric.Aggregation]
	if !exists {
		return periodValues{},
			ErrAggregationUnsupported
	}

	return periodValues{
		current:  selector(current.Summary),
		previous: selector(previous.Summary),
	}, nil
}
