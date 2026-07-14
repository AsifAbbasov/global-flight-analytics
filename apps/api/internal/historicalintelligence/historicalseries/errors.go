package historicalseries

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

var (
	ErrPlanVersionInvalid = errors.New(
		"historical series requires the current historical window plan version",
	)
	ErrPlanWindowInvalid = errors.New(
		"historical series plan does not expose a positive usable window",
	)
	ErrBucketValueCountInvalid = errors.New(
		"historical series bucket value count must match planned buckets",
	)
	ErrBucketValueOrderInvalid = errors.New(
		"historical series bucket values must match planned bucket order and boundaries",
	)
	ErrCoverageRatioInvalid = errors.New(
		"historical series coverage ratio must be between zero and one",
	)
	ErrBuilderVersionRequired = errors.New(
		"historical series builder version is required",
	)
	ErrFingerprintInvalid = errors.New(
		"historical series input fingerprint must use sha256:<64 lowercase hexadecimal characters>",
	)
	ErrSourceNamesRequired = errors.New(
		"historical series requires at least one provenance source name",
	)
	ErrLatestSourceTimeInvalid = errors.New(
		"historical series latest source update time must not exceed the analytical as-of time",
	)
	ErrGeneratedAtInvalid = errors.New(
		"historical series generated time must not precede the analytical as-of time",
	)
	ErrUnavailableBucketHasData = errors.New(
		"historical series unavailable bucket must not contain value or samples",
	)
)

type ContractValidationError struct {
	Report historicalcontract.ValidationReport
}

func (err *ContractValidationError) Error() string {
	if err == nil {
		return "historical series contract validation failed"
	}

	return fmt.Sprintf(
		"historical series contract validation failed: errors=%d warnings=%d",
		err.Report.ErrorCount,
		err.Report.WarningCount,
	)
}
