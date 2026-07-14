package routeresolver

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

var (
	ErrInvalidPartialConfidenceFactor = errors.New(
		"route resolver partial confidence factor must be finite and between zero and one",
	)
	ErrInvalidSameAirportConfidenceFactor = errors.New(
		"route resolver same-airport confidence factor must be finite and between zero and one",
	)
	ErrInvalidEndpointEvidence = errors.New(
		"route resolver endpoint evidence is structurally invalid",
	)
	ErrSourceNamesRequired = errors.New(
		"route resolver requires at least one provenance source name",
	)
	ErrGeneratedBeforeAsOfTime = errors.New(
		"route resolver generated time must not be before the analytical as-of time",
	)
	ErrContractValidation = errors.New(
		"route resolver produced an invalid route contract",
	)
)

type ContractValidationError struct {
	Report routecontract.ValidationReport
}

func (err *ContractValidationError) Error() string {
	if err == nil {
		return ErrContractValidation.Error()
	}

	return fmt.Sprintf(
		"%s: %d error(s), %d warning(s)",
		ErrContractValidation,
		err.Report.ErrorCount,
		err.Report.WarningCount,
	)
}

func (err *ContractValidationError) Unwrap() error {
	return ErrContractValidation
}
