package postgres

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

var ErrTrajectoryWriteModeInvalid = errors.New(
	"trajectory write mode is invalid",
)

type trajectoryWriteMode uint8

const (
	trajectoryWriteModeLive trajectoryWriteMode = iota + 1
	trajectoryWriteModeReconciled
)

type trajectoryWriteRequest struct {
	mode                 trajectoryWriteMode
	reconciliationTaskID string
	attemptCount         int
	item                 trajectory.FlightTrajectory
}

func newLiveTrajectoryWriteRequest(
	item trajectory.FlightTrajectory,
) trajectoryWriteRequest {
	return trajectoryWriteRequest{
		mode: trajectoryWriteModeLive,
		item: item,
	}
}

func newReconciledTrajectoryWriteRequest(
	taskID string,
	attemptCount int,
	item trajectory.FlightTrajectory,
) (trajectoryWriteRequest, error) {
	normalizedTaskID := reconciliation.NormalizeTaskID(taskID)
	if normalizedTaskID == "" {
		return trajectoryWriteRequest{}, reconciliation.ErrTaskIDRequired
	}
	if attemptCount <= 0 {
		return trajectoryWriteRequest{}, reconciliation.ErrAttemptCountInvalid
	}

	request := trajectoryWriteRequest{
		mode:                 trajectoryWriteModeReconciled,
		reconciliationTaskID: normalizedTaskID,
		attemptCount:         attemptCount,
		item:                 item,
	}
	if err := request.validate(); err != nil {
		return trajectoryWriteRequest{}, err
	}
	return request, nil
}

func (request trajectoryWriteRequest) validate() error {
	switch request.mode {
	case trajectoryWriteModeLive:
		if request.reconciliationTaskID != "" || request.attemptCount != 0 {
			return fmt.Errorf(
				"%w: live write contains reconciliation metadata",
				ErrTrajectoryWriteModeInvalid,
			)
		}
	case trajectoryWriteModeReconciled:
		if reconciliation.NormalizeTaskID(request.reconciliationTaskID) == "" {
			return reconciliation.ErrTaskIDRequired
		}
		if request.attemptCount <= 0 {
			return reconciliation.ErrAttemptCountInvalid
		}
	default:
		return fmt.Errorf(
			"%w: got %d",
			ErrTrajectoryWriteModeInvalid,
			request.mode,
		)
	}
	return nil
}

func (request trajectoryWriteRequest) isReconciled() bool {
	return request.mode == trajectoryWriteModeReconciled
}
