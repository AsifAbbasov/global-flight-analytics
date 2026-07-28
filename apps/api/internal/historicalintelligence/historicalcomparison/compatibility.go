package historicalcomparison

import (
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func validateCompatibility(
	current historicalcontract.Result,
	previous historicalcontract.Result,
) error {
	switch {
	case current.SchemaVersion != previous.SchemaVersion:
		return ErrSchemaMismatch
	case current.Metric != previous.Metric:
		return ErrMetricMismatch
	case !current.Scope.Equal(previous.Scope):
		return ErrScopeMismatch
	case current.Granularity != previous.Granularity:
		return ErrGranularityMismatch
	case !current.Window.AsOfTime.Equal(
		previous.Window.AsOfTime,
	):
		return ErrAsOfTimeMismatch
	case current.Window.Duration() !=
		previous.Window.Duration():
		return ErrWindowDurationMismatch
	case !previous.Window.EndTime.Equal(
		current.Window.StartTime,
	):
		return ErrWindowNotAdjacent
	case current.Status ==
		historicalcontract.SeriesStatusUnavailable ||
		previous.Status ==
			historicalcontract.SeriesStatusUnavailable ||
		current.Summary.PointCount == 0 ||
		previous.Summary.PointCount == 0:
		return ErrSeriesUnavailable
	}

	return validateCoverageCompatibility(
		current,
		previous,
	)
}

func validateCoverageCompatibility(
	current historicalcontract.Result,
	previous historicalcontract.Result,
) error {
	if current.Status != previous.Status ||
		len(current.Points) != len(previous.Points) {
		return ErrCoverageMismatch
	}

	for index := range current.Points {
		currentPoint := current.Points[index]
		previousPoint := previous.Points[index]

		if currentPoint.Status != previousPoint.Status ||
			currentPoint.EndTime.Sub(
				currentPoint.StartTime,
			) != previousPoint.EndTime.Sub(
				previousPoint.StartTime,
			) ||
			math.Abs(
				currentPoint.CoverageRatio-
					previousPoint.CoverageRatio,
			) > coverageEqualityTolerance {
			return ErrCoverageMismatch
		}
	}

	return nil
}
