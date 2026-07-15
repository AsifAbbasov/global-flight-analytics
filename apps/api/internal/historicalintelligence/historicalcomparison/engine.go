package historicalcomparison

import (
	"reflect"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func Attach(
	current historicalcontract.Result,
	previous historicalcontract.Result,
) (historicalcontract.Result, error) {
	if err := validateSourceResult(
		current,
		ErrCurrentResultInvalid,
	); err != nil {
		return historicalcontract.Result{}, err
	}
	if err := validateSourceResult(
		previous,
		ErrPreviousResultInvalid,
	); err != nil {
		return historicalcontract.Result{}, err
	}

	if current.SchemaVersion !=
		previous.SchemaVersion {
		return historicalcontract.Result{},
			ErrSchemaMismatch
	}
	if current.Metric != previous.Metric {
		return historicalcontract.Result{},
			ErrMetricMismatch
	}
	if !reflect.DeepEqual(
		current.Scope,
		previous.Scope,
	) {
		return historicalcontract.Result{},
			ErrScopeMismatch
	}
	if current.Granularity !=
		previous.Granularity {
		return historicalcontract.Result{},
			ErrGranularityMismatch
	}
	if !current.Window.AsOfTime.Equal(
		previous.Window.AsOfTime,
	) {
		return historicalcontract.Result{},
			ErrAsOfTimeMismatch
	}
	if current.Window.Duration() !=
		previous.Window.Duration() {
		return historicalcontract.Result{},
			ErrWindowDurationMismatch
	}
	if !previous.Window.EndTime.Equal(
		current.Window.StartTime,
	) {
		return historicalcontract.Result{},
			ErrWindowNotAdjacent
	}
	if current.Status ==
		historicalcontract.SeriesStatusUnavailable ||
		previous.Status ==
			historicalcontract.SeriesStatusUnavailable ||
		current.Summary.PointCount == 0 ||
		previous.Summary.PointCount == 0 {
		return historicalcontract.Result{},
			ErrSeriesUnavailable
	}

	values, err := comparisonValues(
		current,
		previous,
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	absoluteChange := values.Current -
		values.Previous
	var percentageChange *float64
	if values.Previous != 0 {
		value := (absoluteChange /
			values.Previous) * 100
		percentageChange = &value
	}

	result := current.Clone()
	result.Comparison =
		&historicalcontract.PeriodComparison{
			PreviousWindow: historicalcontract.TimeWindow{
				StartTime: previous.Window.StartTime.UTC(),
				EndTime:   previous.Window.EndTime.UTC(),
				AsOfTime:  current.Window.AsOfTime.UTC(),
			},
			CurrentValue:     values.Current,
			PreviousValue:    values.Previous,
			AbsoluteChange:   absoluteChange,
			PercentageChange: percentageChange,
			Direction: historicalcontract.
				TrendDirectionForChange(
					absoluteChange,
				),
		}

	report := historicalcontract.Validate(result)
	if report.Status !=
		historicalcontract.ValidationStatusValid {
		return historicalcontract.Result{},
			&ResultValidationError{
				Kind:   ErrCurrentResultInvalid,
				Report: report.Clone(),
			}
	}

	return result.Clone(), nil
}

func comparisonValues(
	current historicalcontract.Result,
	previous historicalcontract.Result,
) (Values, error) {
	switch current.Metric.Aggregation {
	case historicalcontract.AggregationCount,
		historicalcontract.AggregationSum:
		return Values{
			Current:  current.Summary.Total,
			Previous: previous.Summary.Total,
		}, nil

	case historicalcontract.AggregationMinimum:
		return Values{
			Current:  current.Summary.Minimum,
			Previous: previous.Summary.Minimum,
		}, nil

	case historicalcontract.AggregationMaximum:
		return Values{
			Current:  current.Summary.Maximum,
			Previous: previous.Summary.Maximum,
		}, nil

	case historicalcontract.AggregationAverage,
		historicalcontract.AggregationRatio:
		return Values{
			Current:  current.Summary.Average,
			Previous: previous.Summary.Average,
		}, nil

	case historicalcontract.AggregationMedian:
		return Values{
			Current:  current.Summary.Median,
			Previous: previous.Summary.Median,
		}, nil

	default:
		return Values{},
			ErrAggregationUnsupported
	}
}

func validateSourceResult(
	result historicalcontract.Result,
	kind error,
) error {
	report := historicalcontract.Validate(result)
	if report.Status ==
		historicalcontract.ValidationStatusValid {
		return nil
	}

	return &ResultValidationError{
		Kind:   kind,
		Report: report.Clone(),
	}
}
