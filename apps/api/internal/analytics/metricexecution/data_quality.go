package metricexecution

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
)

func attachDataQuality[T any](
	result analyticalresult.Result[T],
	report *dataqualitycontract.Report,
) (analyticalresult.Result[T], error) {
	updated, err := result.WithDataQuality(report)
	if err != nil {
		return analyticalresult.Result[T]{},
			fmt.Errorf(
				"attach metric data quality: %w",
				err,
			)
	}

	return updated, nil
}
