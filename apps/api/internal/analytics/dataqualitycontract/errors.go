package dataqualitycontract

import "errors"

var (
	ErrSourceNameRequired         = errors.New("source name is required")
	ErrSourceRecordTimeRequired   = errors.New("source record time is required")
	ErrReceivedAtRequired         = errors.New("received at is required")
	ErrReceivedBeforeSourceRecord = errors.New("received at must not be before source record time")
	ErrIngestionRunIDRequired     = errors.New("ingestion run id is required")
	ErrTransformationRequired     = errors.New("transformation is required")
	ErrAlgorithmVersionRequired   = errors.New("algorithm version is required")
	ErrInputFingerprintRequired   = errors.New("input fingerprint is required")

	ErrObservedAtRequired      = errors.New("observed at is required")
	ErrEvaluatedAtRequired     = errors.New("evaluated at is required")
	ErrObservationInFuture     = errors.New("observed at must not be after evaluated at")
	ErrExpectedIntervalInvalid = errors.New("expected interval must be positive")
	ErrStaleAfterInvalid       = errors.New("stale after must be greater than or equal to expected interval")
	ErrFreshnessScoreInvalid   = errors.New("freshness score must be finite and between zero and one")
	ErrFreshnessStatusInvalid  = errors.New("freshness status is invalid")

	ErrWindowStartRequired         = errors.New("window start is required")
	ErrWindowEndRequired           = errors.New("window end is required")
	ErrWindowRangeInvalid          = errors.New("window end must be after window start")
	ErrObservationOutsideWindow    = errors.New("observation time must be inside the requested window")
	ErrSamplingDensityScoreInvalid = errors.New("sampling density score must be finite and between zero and one")
	ErrSamplingCountsInvalid       = errors.New("sampling density counts are inconsistent")

	ErrPermissionReasonRequired = errors.New("denied permission requires at least one reason")
	ErrPermissionReasonInvalid  = errors.New("permission reason must not be blank")
	ErrEvaluatedAtMismatch      = errors.New("nested evaluation timestamps must match the report evaluated at time")
	ErrContractVersionInvalid   = errors.New("data quality contract version is invalid")
)
