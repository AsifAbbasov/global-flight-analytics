package historicalcomparison

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"

func Attach(
	current historicalcontract.Result,
	previous historicalcontract.Result,
) (historicalcontract.Result, error) {
	if err := validateSourceResults(current, previous); err != nil {
		return historicalcontract.Result{}, err
	}
	if err := validateCompatibility(current, previous); err != nil {
		return historicalcontract.Result{}, err
	}

	values, err := selectPeriodValues(current, previous)
	if err != nil {
		return historicalcontract.Result{}, err
	}
	comparison, err := buildPeriodComparison(
		values,
		previous.Window,
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	result := assembleComparedResult(
		current,
		previous,
		comparison,
	)
	return validateComparedResult(result)
}
