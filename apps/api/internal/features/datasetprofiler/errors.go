package datasetprofiler

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

var ErrUnsupportedTargetSchema = errors.New(
	"dataset profiler target schema is unsupported",
)

type UnsupportedTargetSchemaError struct {
	SchemaVersion flightfeatures.SchemaVersion
}

func (err *UnsupportedTargetSchemaError) Error() string {
	if err == nil {
		return ErrUnsupportedTargetSchema.Error()
	}

	return fmt.Sprintf(
		"%s: %q",
		ErrUnsupportedTargetSchema,
		err.SchemaVersion,
	)
}

func (err *UnsupportedTargetSchemaError) Unwrap() error {
	return ErrUnsupportedTargetSchema
}
