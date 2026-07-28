package historicalcomparison

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"

func validateSourceResults(
	current historicalcontract.Result,
	previous historicalcontract.Result,
) error {
	if err := validateSourceResult(
		current,
		ErrCurrentResultInvalid,
	); err != nil {
		return err
	}
	if err := validateSourceResult(
		previous,
		ErrPreviousResultInvalid,
	); err != nil {
		return err
	}
	if current.Comparison != nil ||
		previous.Comparison != nil {
		return ErrNestedComparisonUnsupported
	}
	return nil
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

func validateComparedResult(
	result historicalcontract.Result,
) (historicalcontract.Result, error) {
	report := historicalcontract.Validate(result)
	if report.Status !=
		historicalcontract.ValidationStatusValid {
		return historicalcontract.Result{},
			&ResultValidationError{
				Kind:   ErrComparisonResultInvalid,
				Report: report.Clone(),
			}
	}
	return result.Clone(), nil
}
