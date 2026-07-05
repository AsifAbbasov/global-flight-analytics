package sharedsnapshot

import (
	"errors"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
)

var (
	ErrCycleStartedAtRequired = errors.New(
		"shared snapshot cycle start time is required",
	)

	ErrAssembledAtRequired = errors.New(
		"shared snapshot assembled time is required",
	)

	ErrAssembledBeforeCycleStart = errors.New(
		"shared snapshot assembled time cannot precede cycle start time",
	)
)

type Snapshot struct {
	// CycleStartedAt is the time when the provider collection cycle
	// represented by this snapshot started.
	//
	// Publication ordering must use this value rather than AssembledAt.
	// An older slower cycle may finish after a newer faster cycle.
	CycleStartedAt time.Time

	// AssembledAt is the time when the shared snapshot was assembled.
	//
	// It is not provider observation time.
	// It is not the ordering key for competing collection cycles.
	AssembledAt time.Time

	Status providerfanin.BatchStatus

	TotalCount   int
	SuccessCount int
	FailureCount int

	Successes []providerfanin.Success
	Failures  []providerfanin.Failure
}

func FromEnvelope(
	assembledAt time.Time,
	envelope providerfanin.Envelope,
) (Snapshot, error) {
	if assembledAt.IsZero() {
		return Snapshot{}, ErrAssembledAtRequired
	}

	return FromEnvelopeForCycle(
		assembledAt,
		assembledAt,
		envelope,
	)
}

func FromEnvelopeForCycle(
	cycleStartedAt time.Time,
	assembledAt time.Time,
	envelope providerfanin.Envelope,
) (Snapshot, error) {
	if cycleStartedAt.IsZero() {
		return Snapshot{}, ErrCycleStartedAtRequired
	}

	if assembledAt.IsZero() {
		return Snapshot{}, ErrAssembledAtRequired
	}

	if assembledAt.Before(cycleStartedAt) {
		return Snapshot{}, ErrAssembledBeforeCycleStart
	}

	return Snapshot{
		CycleStartedAt: cycleStartedAt.UTC(),
		AssembledAt:    assembledAt.UTC(),

		Status: envelope.Status,

		TotalCount:   envelope.TotalCount,
		SuccessCount: envelope.SuccessCount,
		FailureCount: envelope.FailureCount,

		Successes: cloneSuccesses(
			envelope.Successes,
		),
		Failures: cloneFailures(
			envelope.Failures,
		),
	}, nil
}

func (snapshot Snapshot) Clone() Snapshot {
	return Snapshot{
		CycleStartedAt: snapshot.CycleStartedAt,
		AssembledAt:    snapshot.AssembledAt,

		Status: snapshot.Status,

		TotalCount:   snapshot.TotalCount,
		SuccessCount: snapshot.SuccessCount,
		FailureCount: snapshot.FailureCount,

		Successes: cloneSuccesses(
			snapshot.Successes,
		),
		Failures: cloneFailures(
			snapshot.Failures,
		),
	}
}

func snapshotOrderTime(
	snapshot Snapshot,
) time.Time {
	if !snapshot.CycleStartedAt.IsZero() {
		return snapshot.CycleStartedAt
	}

	return snapshot.AssembledAt
}

func cloneSuccesses(
	successes []providerfanin.Success,
) []providerfanin.Success {
	if successes == nil {
		return nil
	}

	clonedSuccesses := make(
		[]providerfanin.Success,
		len(successes),
	)

	copy(
		clonedSuccesses,
		successes,
	)

	return clonedSuccesses
}

func cloneFailures(
	failures []providerfanin.Failure,
) []providerfanin.Failure {
	if failures == nil {
		return nil
	}

	clonedFailures := make(
		[]providerfanin.Failure,
		len(failures),
	)

	copy(
		clonedFailures,
		failures,
	)

	return clonedFailures
}
