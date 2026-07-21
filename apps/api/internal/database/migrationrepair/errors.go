package migrationrepair

import (
	"errors"
	"fmt"
)

var (
	ErrInspectorRequired = errors.New(
		"migration sequence repair inspector is required",
	)
	ErrPostgresPoolRequired = errors.New(
		"migration sequence repair postgres pool is required",
	)
	ErrContextRequired = errors.New(
		"migration sequence repair context is required",
	)
	ErrMigrationsDirectoryRequired = errors.New(
		"migration sequence repair migrations directory is required",
	)
	ErrRepairPlanInvalid = errors.New(
		"migration sequence repair plan is invalid",
	)
)

type InspectionError struct {
	Operation string
	Err       error
}

func (err *InspectionError) Error() string {
	if err == nil {
		return "inspect migration sequence repair state"
	}

	return fmt.Sprintf(
		"inspect migration sequence repair state during %s: %v",
		err.Operation,
		err.Err,
	)
}

func (err *InspectionError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}
