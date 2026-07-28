package historicalcomparison

import (
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func buildPeriodComparison(
	values periodValues,
	previousWindow historicalcontract.TimeWindow,
) (historicalcontract.PeriodComparison, error) {
	absoluteChange := values.current - values.previous
	if !finite(absoluteChange) {
		return historicalcontract.PeriodComparison{},
			&ArithmeticError{
				Operation: "absolute_change",
			}
	}

	percentageChange, err := calculatePercentageChange(
		absoluteChange,
		values.previous,
	)
	if err != nil {
		return historicalcontract.PeriodComparison{}, err
	}

	return historicalcontract.PeriodComparison{
		PreviousWindow: historicalcontract.TimeWindow{
			StartTime: previousWindow.StartTime.UTC(),
			EndTime:   previousWindow.EndTime.UTC(),
			AsOfTime:  previousWindow.AsOfTime.UTC(),
		},
		CurrentValue:     values.current,
		PreviousValue:    values.previous,
		AbsoluteChange:   absoluteChange,
		PercentageChange: percentageChange,
		Direction: historicalcontract.
			TrendDirectionForChange(absoluteChange),
	}, nil
}

func calculatePercentageChange(
	absoluteChange float64,
	previousValue float64,
) (*float64, error) {
	if previousValue == 0 {
		return nil, nil
	}

	ratio := absoluteChange / previousValue
	if !finite(ratio) ||
		math.Abs(ratio) > math.MaxFloat64/100 {
		return nil, &ArithmeticError{
			Operation: "percentage_change",
		}
	}

	value := ratio * 100
	if !finite(value) {
		return nil, &ArithmeticError{
			Operation: "percentage_change",
		}
	}
	return &value, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}
