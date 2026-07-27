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
	ErrBucketValueInvalid = errors.New(
		"historical series bucket value must be finite and non-negative",
	)
	ErrBucketSampleCountInvalid = errors.New(
		"historical series bucket sample count must be non-negative",
	)
	ErrCoverageEvidenceInvalid = errors.New(
		"historical series bucket coverage evidence is inconsistent",
	)
	ErrDatasetCoverageInvalid = errors.New(
		"historical series dataset coverage evidence is invalid",
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
	ErrSourceNameInvalid = errors.New(
		"historical series provenance source names must be normalized and non-empty",
	)
	ErrSourceNameDuplicate = errors.New(
		"historical series provenance source names must be unique",
	)
	ErrLatestSourceTimeRequired = errors.New(
		"historical series latest source update time is required",
	)
	ErrLatestSourceTimeInvalid = errors.New(
		"historical series latest source update time must not exceed the analytical as-of time",
	)
	ErrGeneratedAtRequired = errors.New(
		"historical series generated time is required",
	)
	ErrGeneratedAtInvalid = errors.New(
		"historical series generated time must not precede the analytical as-of time",
	)
	ErrUnavailableBucketHasData = errors.New(
		"historical series unavailable bucket must not contain value or samples",
	)
	ErrLimitationInvalid = errors.New(
		"historical series limitations must have normalized non-empty code, message, and scope",
	)
	ErrLimitationDuplicate = errors.New(
		"historical series limitation scope and code combinations must be unique",
	)
	ErrSampleCountOverflow = errors.New(
		"historical series total sample count exceeds the platform integer range",
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
