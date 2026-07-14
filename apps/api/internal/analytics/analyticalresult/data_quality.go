package analyticalresult

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
)

func (result Result[T]) WithDataQuality(
	report *dataqualitycontract.Report,
) (Result[T], error) {
	updated := result.Clone()
	updated.DataQuality = cloneDataQuality(report)

	if err := updated.Validate(); err != nil {
		return Result[T]{},
			fmt.Errorf(
				"validate analytical result with data quality: %w",
				err,
			)
	}

	return updated, nil
}
