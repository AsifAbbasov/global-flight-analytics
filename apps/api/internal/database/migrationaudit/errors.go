package migrationaudit

import (
	"errors"
	"fmt"
)

var (
	ErrMigrationsDirRequired = errors.New(
		"migration audit directory is required",
	)
	ErrStateLoaderRequired = errors.New(
		"migration audit database state loader is required",
	)
	ErrPostgresPoolRequired = errors.New(
		"migration audit postgres pool is required",
	)
)

type LocalScanError struct {
	Path string
	Err  error
}

func (err *LocalScanError) Error() string {
	if err == nil {
		return "scan local migrations"
	}

	return fmt.Sprintf(
		"scan local migrations at %q: %v",
		err.Path,
		err.Err,
	)
}

func (err *LocalScanError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Err
}

type DatabaseInspectionError struct {
	Operation string
	Err       error
}

func (err *DatabaseInspectionError) Error() string {
	if err == nil {
		return "inspect migration database history"
	}

	return fmt.Sprintf(
		"inspect migration database history during %s: %v",
		err.Operation,
		err.Err,
	)
}

func (err *DatabaseInspectionError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Err
}
