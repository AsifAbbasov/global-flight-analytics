package featurepipeline

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type Pipeline struct {
	extractor FeatureExtractor
	validator FeatureValidator
	store     featurestore.Store
}

func New(config Config) (*Pipeline, error) {
	if config.Extractor == nil {
		return nil, ErrExtractorRequired
	}
	if config.Validator == nil {
		return nil, ErrValidatorRequired
	}
	if config.Store == nil {
		return nil, ErrStoreRequired
	}

	return &Pipeline{
		extractor: config.Extractor,
		validator: config.Validator,
		store:     config.Store,
	}, nil
}

func (pipeline *Pipeline) Process(
	ctx context.Context,
	request extractor.Request,
) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
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
		Features:         validated.Clone(),
		ValidationReport: report.Clone(),
	}

	if validated.Quality.Status != report.Status {
		return result.Clone(),
			&ValidationStatusMismatchError{
				FeatureStatus: validated.Quality.Status,
				ReportStatus:  report.Status,
			}
	}

	switch report.Status {
	case flightfeatures.ValidationStatusValid,
		flightfeatures.ValidationStatusLimited:
	default:
		return result.Clone(),
			&ValidationRejectedError{
				Status: report.Status,
				Report: report.Clone(),
			}
	}

	record, err := pipeline.store.Put(
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
