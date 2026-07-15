package historicalreplay

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

type Runner struct {
	materializer Materializer
	now          func() time.Time
}

func New(
	config Config,
) (*Runner, error) {
	if config.Materializer == nil {
		return nil, ErrMaterializerRequired
	}
	if config.Now == nil {
		config.Now = time.Now
	}

	return &Runner{
		materializer: config.Materializer,
		now:          config.Now,
	}, nil
}

func (runner *Runner) Run(
	ctx context.Context,
	request Request,
) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	maximumWindowCount :=
		request.MaximumWindowCount
	if maximumWindowCount == 0 {
		maximumWindowCount =
			DefaultMaximumWindowCount
	}
	if maximumWindowCount < 1 ||
		maximumWindowCount >
			MaximumWindowCount {
		return Result{},
			ErrMaximumWindowCountInvalid
	}

	generatedAt := request.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = runner.now()
	}
	generatedAt = generatedAt.UTC()

	plan, err := historicalwindow.Build(
		ctx,
		historicalwindow.Request{
			StartTime: request.StartTime,
			EndTime:   request.EndTime,
			AsOfTime:  request.AsOfTime,
			Granularity: request.
				Granularity,
			MaximumBucketCount: request.
				MaximumBucketCount,
		},
	)
	if err != nil {
		return Result{}, err
	}
	if !plan.HasBuckets() {
		return Result{
				Version: Version,
				Plan:    plan.Clone(),
				Windows: []WindowResult{},
			},
			ErrNoReplayWindow
	}
	if len(plan.Buckets) >
		maximumWindowCount {
		return Result{
				Version: Version,
				Plan:    plan.Clone(),
				Windows: []WindowResult{},
			},
			&WindowCountExceededError{
				Count:   len(plan.Buckets),
				Maximum: maximumWindowCount,
			}
	}

	result := Result{
		Version: Version,
		Plan:    plan.Clone(),
		Windows: make(
			[]WindowResult,
			0,
			len(plan.Buckets),
		),
	}

	for _, bucket := range plan.Buckets {
		if err := ctx.Err(); err != nil {
			return result.Clone(), err
		}

		outcome, materializeErr :=
			runner.materializer.Materialize(
				ctx,
				historicalmaterialization.Request{
					StartTime: bucket.StartTime,
					EndTime:   bucket.EndTime,
					AsOfTime:  request.AsOfTime,

					Granularity: request.
						Granularity,
					MetricName: request.
						MetricName,
					Scope: request.Scope,

					DatasetLimit: request.
						DatasetLimit,
					MaximumBucketCount: request.
						MaximumBucketCount,
					GeneratedAt: generatedAt,
				},
			)
		if materializeErr != nil {
			return result.Clone(),
				&WindowError{
					Sequence:  bucket.Sequence,
					StartTime: bucket.StartTime,
					EndTime:   bucket.EndTime,
					Err:       materializeErr,
				}
		}

		result.Windows = append(
			result.Windows,
			WindowResult{
				Bucket: bucket,
				Record: outcome.Record.Clone(),
			},
		)
	}

	return result.Clone(), nil
}
