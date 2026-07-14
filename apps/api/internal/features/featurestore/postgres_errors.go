package featurestore

import (
	"errors"
	"fmt"
)

var (
	ErrPostgresPoolRequired = errors.New(
		"feature snapshot postgres pool is required",
	)
	ErrInvalidTrajectoryID = errors.New(
		"feature snapshot trajectory id must be a valid UUID",
	)
	ErrCorruptSnapshot = errors.New(
		"stored feature snapshot is internally inconsistent",
	)
)

type DatabaseError struct {
	Operation string
	Err       error
}

func (err *DatabaseError) Error() string {
	if err == nil {
		return "feature snapshot database operation failed"
	}

	return fmt.Sprintf(
		"feature snapshot database operation %s failed: %v",
		err.Operation,
		err.Err,
	)
}

func (err *DatabaseError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Err
}

type CorruptSnapshotError struct {
	Field string
}

func (err *CorruptSnapshotError) Error() string {
	if err == nil {
		return ErrCorruptSnapshot.Error()
	}

	return fmt.Sprintf(
		"%s: field=%s",
		ErrCorruptSnapshot,
		err.Field,
	)
}

func (err *CorruptSnapshotError) Unwrap() error {
	return ErrCorruptSnapshot
}
