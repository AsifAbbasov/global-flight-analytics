package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurepipeline"
)

const commandReportVersion = "flight-feature-materialization-report-v1"

var (
	errTrajectoryReaderRequired = errors.New(
		"flight feature trajectory reader is required",
	)
	errFeatureProcessorRequired = errors.New(
		"flight feature processor is required",
	)
	errAsOfBeforeTrajectoryEnd = errors.New(
		"flight feature as-of time precedes trajectory end",
	)
)

type trajectoryReader interface {
	GetTrajectoryByID(
		context.Context,
		string,
	) (trajectory.FlightTrajectory, error)
	GetLatestTrajectoryByICAO24(
		context.Context,
		string,
	) (trajectory.FlightTrajectory, error)
}

type featureProcessor interface {
	Process(
		context.Context,
		extractor.Request,
	) (featurepipeline.Result, error)
}

type materializationOperation struct {
	reader    trajectoryReader
	processor featureProcessor
}

type commandReport struct {
	Version              string    `json:"version"`
	Selector             string    `json:"selector"`
	SelectorValue        string    `json:"selector_value"`
	PipelineVersion      string    `json:"pipeline_version"`
	TrajectoryID         string    `json:"trajectory_id"`
	IdentityKey          string    `json:"identity_key"`
	ICAO24               string    `json:"icao24"`
	SchemaVersion        string    `json:"schema_version"`
	SnapshotID           string    `json:"snapshot_id"`
	AsOfTime             time.Time `json:"as_of_time"`
	ValidationStatus     string    `json:"validation_status"`
	CompletenessScore    float64   `json:"completeness_score"`
	InputQualityScore    float64   `json:"input_quality_score"`
	SupportingPointCount int       `json:"supporting_point_count"`
	LimitationCount      int       `json:"limitation_count"`
	InputFingerprint     string    `json:"input_fingerprint"`
	ExtractedAt          time.Time `json:"extracted_at"`
	StoredAt             time.Time `json:"stored_at"`
}

func newMaterializationOperation(
	reader trajectoryReader,
	processor featureProcessor,
) (*materializationOperation, error) {
	if reader == nil {
		return nil, errTrajectoryReaderRequired
	}
	if processor == nil {
		return nil, errFeatureProcessorRequired
	}
	return &materializationOperation{
		reader:    reader,
		processor: processor,
	}, nil
}

func (operation *materializationOperation) Execute(
	ctx context.Context,
	options commandOptions,
) (commandReport, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return commandReport{}, err
	}

	item, selector, selectorValue, err := operation.loadTrajectory(ctx, options)
	if err != nil {
		return commandReport{}, err
	}

	asOfTime := options.AsOfTime
	if asOfTime.IsZero() {
		asOfTime = item.EndTime.UTC()
	}
	if asOfTime.Before(item.EndTime.UTC()) {
		return commandReport{}, fmt.Errorf(
			"%w: as_of=%s trajectory_end=%s",
			errAsOfBeforeTrajectoryEnd,
			asOfTime.Format(time.RFC3339Nano),
			item.EndTime.UTC().Format(time.RFC3339Nano),
		)
	}

	result, err := operation.processor.Process(
		ctx,
		extractor.Request{
			Trajectory: item,
			AsOfTime:   asOfTime,
		},
	)
	if err != nil {
		return commandReport{}, fmt.Errorf(
			"materialize flight features: %w",
			err,
		)
	}

	features := result.Record.Features
	return commandReport{
		Version:              commandReportVersion,
		Selector:             selector,
		SelectorValue:        selectorValue,
		PipelineVersion:      result.PipelineVersion,
		TrajectoryID:         features.TrajectoryID,
		IdentityKey:          features.IdentityKey,
		ICAO24:               features.ICAO24,
		SchemaVersion:        string(features.SchemaVersion),
		SnapshotID:           result.Record.ID,
		AsOfTime:             features.Window.AsOfTime.UTC(),
		ValidationStatus:     string(features.Quality.Status),
		CompletenessScore:    features.Quality.CompletenessScore,
		InputQualityScore:    features.Quality.InputQualityScore,
		SupportingPointCount: features.Quality.SupportingPointCount,
		LimitationCount:      len(features.Quality.Limitations),
		InputFingerprint:     result.Record.InputFingerprint,
		ExtractedAt:          features.ExtractedAt.UTC(),
		StoredAt:             result.Record.StoredAt.UTC(),
	}, nil
}

func (operation *materializationOperation) loadTrajectory(
	ctx context.Context,
	options commandOptions,
) (trajectory.FlightTrajectory, string, string, error) {
	if strings.TrimSpace(options.TrajectoryID) != "" {
		item, err := operation.reader.GetTrajectoryByID(
			ctx,
			options.TrajectoryID,
		)
		return item, "trajectory_id", options.TrajectoryID, err
	}

	item, err := operation.reader.GetLatestTrajectoryByICAO24(
		ctx,
		options.ICAO24,
	)
	return item, "latest_icao24", options.ICAO24, err
}
