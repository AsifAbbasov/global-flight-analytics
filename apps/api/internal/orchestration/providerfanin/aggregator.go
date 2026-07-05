package providerfanin

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type BatchStatus string

const (
	BatchStatusEmpty BatchStatus = "empty"

	BatchStatusSucceeded BatchStatus = "succeeded"

	BatchStatusPartial BatchStatus = "partial"

	BatchStatusFailed BatchStatus = "failed"
)

type Success struct {
	TaskID     string
	Provider   providerpolicy.Provider
	RequestKey string
	Value      any
	Shared     bool
}

type Failure struct {
	TaskID     string
	Provider   providerpolicy.Provider
	RequestKey string
	Err        error
}

type Envelope struct {
	Status BatchStatus

	TotalCount   int
	SuccessCount int
	FailureCount int

	Successes []Success
	Failures  []Failure
}

func Aggregate(
	results []providerfanout.Result,
) Envelope {
	envelope := Envelope{
		TotalCount: len(results),
		Successes: make(
			[]Success,
			0,
			len(results),
		),
		Failures: make(
			[]Failure,
			0,
			len(results),
		),
	}

	for _, result := range results {
		if result.Err != nil {
			envelope.Failures = append(
				envelope.Failures,
				Failure{
					TaskID:     result.TaskID,
					Provider:   result.Provider,
					RequestKey: result.RequestKey,
					Err:        result.Err,
				},
			)

			continue
		}

		envelope.Successes = append(
			envelope.Successes,
			Success{
				TaskID:     result.TaskID,
				Provider:   result.Provider,
				RequestKey: result.RequestKey,
				Value:      result.Value,
				Shared:     result.Shared,
			},
		)
	}

	envelope.SuccessCount = len(
		envelope.Successes,
	)

	envelope.FailureCount = len(
		envelope.Failures,
	)

	envelope.Status = resolveBatchStatus(
		envelope.TotalCount,
		envelope.SuccessCount,
		envelope.FailureCount,
	)

	return envelope
}

func resolveBatchStatus(
	totalCount int,
	successCount int,
	failureCount int,
) BatchStatus {
	if totalCount == 0 {
		return BatchStatusEmpty
	}

	if successCount == totalCount {
		return BatchStatusSucceeded
	}

	if failureCount == totalCount {
		return BatchStatusFailed
	}

	return BatchStatusPartial
}
