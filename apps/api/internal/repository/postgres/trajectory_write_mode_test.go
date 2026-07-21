package postgres

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestLiveTrajectoryWriteRequestHasExplicitMode(t *testing.T) {
	t.Parallel()

	request := newLiveTrajectoryWriteRequest(trajectory.FlightTrajectory{})
	if request.mode != trajectoryWriteModeLive {
		t.Fatalf("mode = %d, want live", request.mode)
	}
	if request.isReconciled() {
		t.Fatal("live trajectory write was classified as reconciled")
	}
	if err := request.validate(); err != nil {
		t.Fatalf("validate live request: %v", err)
	}
}

func TestReconciledTrajectoryWriteRequestNormalizesEvidence(t *testing.T) {
	t.Parallel()

	request, err := newReconciledTrajectoryWriteRequest(
		" 11111111-1111-1111-1111-111111111111 ",
		2,
		trajectory.FlightTrajectory{},
	)
	if err != nil {
		t.Fatalf("create reconciled request: %v", err)
	}
	if !request.isReconciled() {
		t.Fatal("reconciled trajectory write lost its explicit mode")
	}
	if request.reconciliationTaskID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("task id = %q", request.reconciliationTaskID)
	}
	if request.attemptCount != 2 {
		t.Fatalf("attempt count = %d", request.attemptCount)
	}
}

func TestReconciledTrajectoryWriteRequestRejectsIncompleteEvidence(t *testing.T) {
	t.Parallel()

	_, err := newReconciledTrajectoryWriteRequest(
		"   ",
		1,
		trajectory.FlightTrajectory{},
	)
	if !errors.Is(err, reconciliation.ErrTaskIDRequired) {
		t.Fatalf("expected task id error, got %v", err)
	}

	_, err = newReconciledTrajectoryWriteRequest(
		"11111111-1111-1111-1111-111111111111",
		0,
		trajectory.FlightTrajectory{},
	)
	if !errors.Is(err, reconciliation.ErrAttemptCountInvalid) {
		t.Fatalf("expected attempt count error, got %v", err)
	}
}

func TestTrajectoryWriteRequestRejectsImplicitOrMixedMode(t *testing.T) {
	t.Parallel()

	for _, request := range []trajectoryWriteRequest{
		{},
		{
			mode:                 trajectoryWriteModeLive,
			reconciliationTaskID: "11111111-1111-1111-1111-111111111111",
			attemptCount:         1,
		},
	} {
		err := request.validate()
		if !errors.Is(err, ErrTrajectoryWriteModeInvalid) {
			t.Fatalf("request %#v: expected explicit-mode error, got %v", request, err)
		}
	}
}
