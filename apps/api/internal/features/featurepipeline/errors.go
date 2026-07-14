package featurepipeline

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

var (
	ErrExtractorRequired = errors.New(
		"feature processing pipeline extractor is required",
	)
	ErrValidatorRequired = errors.New(
		"feature processing pipeline validator is required",
	)
	ErrStoreRequired = errors.New(
		"feature processing pipeline store is required",
	)
	ErrValidationRejected = errors.New(
		"feature processing pipeline rejected validation result",
	)
	ErrValidationStatusMismatch = errors.New(
		"validated features status does not match validation report status",
	)
)

type StageError struct {
	Stage Stage
	Err   error
}

func (err *StageError) Error() string {
	if err == nil {
		return "feature processing pipeline stage failed"
	}

	return fmt.Sprintf(
		"feature processing pipeline %s stage failed: %v",
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

type ValidationRejectedError struct {
	Status flightfeatures.ValidationStatus
	Report validator.Report
}

func (err *ValidationRejectedError) Error() string {
	if err == nil {
		return ErrValidationRejected.Error()
	}

	return fmt.Sprintf(
		"%s: status=%q errors=%d warnings=%d",
		ErrValidationRejected,
		err.Status,
		err.Report.ErrorCount,
		err.Report.WarningCount,
	)
}

func (err *ValidationRejectedError) Unwrap() error {
	return ErrValidationRejected
}

type ValidationStatusMismatchError struct {
	FeatureStatus flightfeatures.ValidationStatus
	ReportStatus  flightfeatures.ValidationStatus
}

func (err *ValidationStatusMismatchError) Error() string {
	if err == nil {
		return ErrValidationStatusMismatch.Error()
	}

	return fmt.Sprintf(
		"%s: features=%q report=%q",
		ErrValidationStatusMismatch,
		err.FeatureStatus,
		err.ReportStatus,
	)
}

func (err *ValidationStatusMismatchError) Unwrap() error {
	return ErrValidationStatusMismatch
}

type ConstructionError struct {
	Component string
	Err       error
}

func (err *ConstructionError) Error() string {
	if err == nil {
		return "construct feature processing pipeline component"
	}

	return fmt.Sprintf(
		"construct feature processing pipeline component %q: %v",
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

const (
	ComponentExtractorComposition = "extractor_composition"
	ComponentValidator            = "validator"
	ComponentPipeline             = "pipeline"
)
