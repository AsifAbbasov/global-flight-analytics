package historicalreplay

import (
	"context"
	"errors"
	"time"

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
		return Result{}, ErrContextRequired
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	startedAt := runner.now().UTC()
	if startedAt.IsZero() {
		return Result{}, ErrClockRequired
	}

	normalized, err := runner.normalizeRequest(
		request,
		startedAt,
	)
	if err != nil {
		return Result{}, err
	}

	plan, err := buildReplayPlan(
		ctx,
		normalized,
	)
	if err != nil {
		return Result{}, err
	}

	result := newReplayResult(
		normalized,
		plan,
		startedAt,
	)
	if !plan.HasBuckets() {
		return runner.failResult(
			result,
			historicalwindow.Bucket{},
			FailureCodeNoReplayWindow,
			ErrNoReplayWindow,
		)
	}

	previousCurrentFingerprint := ""
	for _, bucket := range plan.Buckets {
		if err := ctx.Err(); err != nil {
			return runner.failResult(
				result,
				bucket,
				FailureCodeContextCanceled,
				err,
			)
		}

		outcome, materializeErr :=
			runner.materializer.Materialize(
				ctx,
				normalized.materializationRequest(
					bucket,
				),
			)
		if materializeErr != nil {
			windowErr := &WindowError{
				Sequence:  bucket.Sequence,
				StartTime: bucket.StartTime,
				EndTime:   bucket.EndTime,
				Err:       materializeErr,
			}
			return runner.failResult(
				result,
				bucket,
				FailureCodeMaterialization,
				windowErr,
			)
		}

		if validationErr := validateOutcome(
			bucket,
			normalized,
			outcome,
		); validationErr != nil {
			windowErr := &WindowError{
				Sequence:  bucket.Sequence,
				StartTime: bucket.StartTime,
				EndTime:   bucket.EndTime,
				Err:       validationErr,
			}
			return runner.failResult(
				result,
				bucket,
				FailureCodeOutcomeContract,
				windowErr,
			)
		}

		previousFingerprint := outcome.
			PreviousResult.Provenance.
			InputFingerprint
		if previousCurrentFingerprint != "" &&
			previousFingerprint !=
				previousCurrentFingerprint {
			continuityErr :=
				&OutcomeContractError{
					Field: "previous_result.provenance.input_fingerprint",
					Err:   ErrReplayContinuityMismatch,
				}
			windowErr := &WindowError{
				Sequence:  bucket.Sequence,
				StartTime: bucket.StartTime,
				EndTime:   bucket.EndTime,
				Err:       continuityErr,
			}
			return runner.failResult(
				result,
				bucket,
				FailureCodeContinuityMismatch,
				windowErr,
			)
		}

		currentFingerprint := outcome.
			CurrentPeriodResult.Provenance.
			InputFingerprint
		result.Windows = append(
			result.Windows,
			WindowResult{
				Bucket:                         bucket,
				Record:                         outcome.Record.Clone(),
				PreviousPeriodInputFingerprint: previousFingerprint,
				CurrentPeriodInputFingerprint:  currentFingerprint,
			},
		)
		result.CompletedWindowCount =
			len(result.Windows)
		previousCurrentFingerprint =
			currentFingerprint
	}

	result.Status = StatusComplete
	result.CompletedAt = runner.completedAt(
		result.StartedAt,
	)
	if err := result.Validate(); err != nil {
		return result.Clone(),
			errors.Join(
				ErrResultInvalid,
				err,
			)
	}
	return result.Clone(), nil
}

func newReplayResult(
	request normalizedRequest,
	plan historicalwindow.Plan,
	startedAt time.Time,
) Result {
	result := Result{
		Version: Version,
		Status:  StatusFailed,

		Plan: plan.Clone(),

		MetricName:  request.MetricName,
		Scope:       request.Scope,
		Granularity: request.Granularity,

		DatasetLimit: request.DatasetLimit,
		MaximumBucketCount: request.
			MaximumBucketCount,
		MaximumWindowCount: request.
			MaximumWindowCount,

		PlannedWindowCount: len(plan.Buckets),
		Windows: make(
			[]WindowResult,
			0,
			len(plan.Buckets),
		),

		GeneratedAt: request.GeneratedAt,
		StartedAt:   startedAt,
	}
	result.InputFingerprint =
		replayInputFingerprint(
			result.Plan,
			result.MetricName,
			result.Scope,
			result.Granularity,
			result.DatasetLimit,
			result.MaximumBucketCount,
			result.MaximumWindowCount,
			result.GeneratedAt,
		)
	return result
}

func (runner *Runner) failResult(
	result Result,
	bucket historicalwindow.Bucket,
	code FailureCode,
	err error,
) (Result, error) {
	if len(result.Windows) > 0 {
		result.Status = StatusPartial
	} else {
		result.Status = StatusFailed
	}
	result.CompletedWindowCount =
		len(result.Windows)
	result.HasFailure = true
	result.Failure = Failure{
		Sequence:  bucket.Sequence,
		StartTime: bucket.StartTime.UTC(),
		EndTime:   bucket.EndTime.UTC(),
		Code:      code,
		Message:   err.Error(),
	}
	result.CompletedAt = runner.completedAt(
		result.StartedAt,
	)

	validationErr := result.Validate()
	if validationErr != nil {
		return result.Clone(),
			errors.Join(
				err,
				&ResultContractError{
					Field: "result",
					Err: errors.Join(
						ErrResultInvalid,
						validationErr,
					),
				},
			)
	}
	return result.Clone(), err
}

func (runner *Runner) completedAt(
	startedAt time.Time,
) time.Time {
	completedAt := runner.now().UTC()
	if completedAt.IsZero() ||
		completedAt.Before(startedAt) {
		return startedAt
	}
	return completedAt
}
