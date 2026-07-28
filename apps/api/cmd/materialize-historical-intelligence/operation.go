package main

import (
	"context"
	"errors"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalreplay"
)

var (
	errMaterializerRequired = errors.New(
		"Historical Intelligence materializer is required",
	)
	errReplayRunnerRequired = errors.New(
		"Historical Intelligence replay runner is required",
	)
	errClockRequired = errors.New(
		"Historical Intelligence command clock is required",
	)
	errOperationModeUnsupported = errors.New(
		"Historical Intelligence operation mode is unsupported",
	)
	errCommandContextRequired = errors.New(
		"Historical Intelligence command context is required",
	)
)

type commandMaterializer interface {
	Materialize(
		context.Context,
		historicalmaterialization.Request,
	) (historicalmaterialization.Outcome, error)
}

type commandReplayRunner interface {
	Run(
		context.Context,
		historicalreplay.Request,
	) (historicalreplay.Result, error)
}

type commandOperation struct {
	materializer commandMaterializer
	replayRunner commandReplayRunner
	now          func() time.Time
}

func newCommandOperation(
	materializer commandMaterializer,
	replayRunner commandReplayRunner,
	now func() time.Time,
) (*commandOperation, error) {
	if materializer == nil {
		return nil, errMaterializerRequired
	}
	if replayRunner == nil {
		return nil, errReplayRunnerRequired
	}
	if now == nil {
		return nil, errClockRequired
	}

	return &commandOperation{
		materializer: materializer,
		replayRunner: replayRunner,
		now:          now,
	}, nil
}

func (operation *commandOperation) Execute(
	ctx context.Context,
	options commandOptions,
) (commandReport, error) {
	if ctx == nil {
		return commandReport{},
			errCommandContextRequired
	}
	if err := ctx.Err(); err != nil {
		return commandReport{}, err
	}

	generatedAt := operation.now().UTC()
	if generatedAt.IsZero() {
		return commandReport{},
			errClockRequired
	}
	if generatedAt.Before(options.AsOfTime) {
		return commandReport{},
			errors.New(
				"command completion time must not be before as-of time",
			)
	}

	switch options.Mode {
	case operationModeMaterialize:
		return operation.executeMaterialize(
			ctx,
			options,
			generatedAt,
		)

	case operationModeReplay:
		return operation.executeReplay(
			ctx,
			options,
			generatedAt,
		)

	default:
		return commandReport{},
			errOperationModeUnsupported
	}
}

func (operation *commandOperation) executeMaterialize(
	ctx context.Context,
	options commandOptions,
	generatedAt time.Time,
) (commandReport, error) {
	outcome, err := operation.materializer.Materialize(
		ctx,
		historicalmaterialization.Request{
			StartTime: options.StartTime,
			EndTime:   options.EndTime,
			AsOfTime:  options.AsOfTime,

			Granularity: options.Granularity,
			MetricName:  options.MetricName,
			Scope:       options.Scope,

			DatasetLimit: options.DatasetLimit,
			MaximumBucketCount: options.
				MaximumBucketCount,
			GeneratedAt: generatedAt,
		},
	)
	if err != nil {
		return commandReport{}, err
	}

	return reportFromMaterialization(
		options,
		outcome,
		operation.now().UTC(),
	), nil
}

func (operation *commandOperation) executeReplay(
	ctx context.Context,
	options commandOptions,
	generatedAt time.Time,
) (commandReport, error) {
	replayed, err := operation.replayRunner.Run(
		ctx,
		historicalreplay.Request{
			StartTime: options.StartTime,
			EndTime:   options.EndTime,
			AsOfTime:  options.AsOfTime,

			Granularity: options.Granularity,
			MetricName:  options.MetricName,
			Scope:       options.Scope,

			DatasetLimit: options.DatasetLimit,
			MaximumBucketCount: options.
				MaximumBucketCount,
			MaximumWindowCount: options.
				MaximumWindowCount,
			GeneratedAt: generatedAt,
		},
	)
	if err != nil && replayed.Version == "" {
		return commandReport{}, err
	}

	report := reportFromReplay(
		options,
		replayed,
		operation.now().UTC(),
	)
	return report, err
}

func reportRecordFromAggregate(
	record historicalaggregate.Record,
) reportRecord {
	return reportRecord{
		ID:               record.ID,
		InputFingerprint: record.InputFingerprint,
		Window: reportWindowFromContract(
			record.Result.Window,
		),
		Status: string(
			record.Result.Status,
		),
		ConfidenceLevel: string(
			record.Result.Confidence.Level,
		),
		PointCount: record.Result.Summary.PointCount,
		Total:      record.Result.Summary.Total,
		StoredAt:   record.StoredAt.UTC(),
	}
}
