package featurepipeline

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

type Pipeline struct {
	extractor FeatureExtractor
	validator FeatureValidator
	writer    FeatureWriter
}

func New(config Config) (*Pipeline, error) {
	if dependencyMissing(config.Extractor) {
		return nil, ErrExtractorRequired
	}
	if dependencyMissing(config.Validator) {
		return nil, ErrValidatorRequired
	}
	if dependencyMissing(config.Writer) {
		return nil, ErrWriterRequired
	}

	return &Pipeline{
		extractor: config.Extractor,
		validator: config.Validator,
		writer:    config.Writer,
	}, nil
}

func (pipeline *Pipeline) Process(
	ctx context.Context,
	request extractor.Request,
) (Result, error) {
	if ctx == nil {
		return Result{}, ErrContextRequired
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	extracted, err := pipeline.extractor.Extract(
		ctx,
		cloneRequest(request),
	)
	if err != nil {
		return Result{}, newStageError(
			StageExtraction,
			err,
		)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	validated, report, err := pipeline.validator.Validate(
		ctx,
		extracted.Clone(),
	)
	if err != nil {
		return Result{}, newStageError(
			StageValidation,
			err,
		)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	result := Result{
		PipelineVersion:  Version,
		ValidationReport: report.Clone(),
	}

	if validated.Quality.Status != report.Status {
		return result.Clone(),
			&ValidationStatusMismatchError{
				FeatureStatus: validated.Quality.Status,
				ReportStatus:  report.Status,
			}
	}
	if err := validator.ValidateReport(
		report,
		validated.Quality.Status,
	); err != nil {
		return result.Clone(),
			newStageError(
				StageValidation,
				err,
			)
	}

	switch report.Status {
	case flightfeatures.ValidationStatusValid,
		flightfeatures.ValidationStatusLimited:
	default:
		return result.Clone(),
			&ValidationRejectedError{
				Status:   report.Status,
				Features: validated.Clone(),
				Report:   report.Clone(),
			}
	}

	record, err := pipeline.writer.Put(
		ctx,
		validated.Clone(),
	)
	if err != nil {
		return result.Clone(),
			newStageError(
				StageStorage,
				err,
			)
	}

	result.Record = record.Clone()

	return result.Clone(), nil
}

func newStageError(
	stage Stage,
	err error,
) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	return &StageError{
		Stage: stage,
		Err:   err,
	}
}

func cloneRequest(
	request extractor.Request,
) extractor.Request {
	cloned := request
	cloned.Trajectory = cloneTrajectory(
		request.Trajectory,
	)

	return cloned
}

func cloneTrajectory(
	item trajectory.FlightTrajectory,
) trajectory.FlightTrajectory {
	cloned := item
	cloned.Points = append(
		[]trajectory.TrackPoint4D(nil),
		item.Points...,
	)
	cloned.Segments = append(
		[]trajectory.TrajectorySegment(nil),
		item.Segments...,
	)
	cloned.CoverageGaps = append(
		[]trajectory.CoverageGap(nil),
		item.CoverageGaps...,
	)

	return cloned
}
