package routepipeline

import (
	"errors"
	"fmt"
)

var (
	ErrTrajectoryReaderRequired = errors.New(
		"route pipeline trajectory reader is required",
	)
	ErrAirportListerRequired = errors.New(
		"route pipeline airport lister is required",
	)
	ErrStoreRequired = errors.New(
		"route pipeline store is required",
	)
	ErrTrajectoryIDRequired = errors.New(
		"route pipeline trajectory id is required",
	)
	ErrTrajectoryIdentityMismatch = errors.New(
		"route pipeline loaded trajectory id does not match the request",
	)
	ErrInvalidAirportCatalogTTL = errors.New(
		"route pipeline airport catalog ttl must be greater than zero",
	)
	ErrAirportSourceNameRequired = errors.New(
		"route pipeline airport source name is required",
	)
	ErrNoAnalyticalAsOfTime = errors.New(
		"route pipeline cannot determine an analytical as-of time",
	)
	ErrPostgresPoolRequired = errors.New(
		"route pipeline postgres pool is required",
	)
)

type StageError struct {
	Stage Stage
	Err   error
}

func (err *StageError) Error() string {
	if err == nil {
		return "route pipeline stage failed"
	}

	return fmt.Sprintf(
		"route pipeline stage %s failed: %v",
		err.Stage,
		err.Err,
	)
}

func (err *StageError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Err
}

type ConstructionError struct {
	Component string
	Err       error
}

func (err *ConstructionError) Error() string {
	if err == nil {
		return "route pipeline construction failed"
	}

	return fmt.Sprintf(
		"route pipeline component %s construction failed: %v",
		err.Component,
		err.Err,
	)
}

func (err *ConstructionError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Err
}
