package postgres

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
)

var ErrDataQualityWriteModeInvalid = errors.New(
	"data quality write mode is invalid",
)

type dataQualityWriteMode uint8

const (
	dataQualityWriteModeLive dataQualityWriteMode = iota + 1
	dataQualityWriteModeReconciled
)

type dataQualityWriteRequest struct {
	mode                 dataQualityWriteMode
	reconciliationTaskID string
	attemptCount         int
	state                flightstate.FlightState
	quality              dataquality.DataQuality
}

func newLiveDataQualityWriteRequest(
	state flightstate.FlightState,
	quality dataquality.DataQuality,
) dataQualityWriteRequest {
	return dataQualityWriteRequest{
		mode:    dataQualityWriteModeLive,
		state:   state,
		quality: quality,
	}
}

func newReconciledDataQualityWriteRequest(
	taskID string,
	attemptCount int,
	state flightstate.FlightState,
	quality dataquality.DataQuality,
) (dataQualityWriteRequest, error) {
	normalizedTaskID := reconciliation.NormalizeTaskID(taskID)
	if normalizedTaskID == "" {
		return dataQualityWriteRequest{}, reconciliation.ErrTaskIDRequired
	}
	if attemptCount <= 0 {
		return dataQualityWriteRequest{}, reconciliation.ErrAttemptCountInvalid
	}
	return dataQualityWriteRequest{
		mode:                 dataQualityWriteModeReconciled,
		reconciliationTaskID: normalizedTaskID,
		attemptCount:         attemptCount,
		state:                state,
		quality:              quality,
	}, nil
}

func (request dataQualityWriteRequest) validate() error {
	switch request.mode {
	case dataQualityWriteModeLive:
		return nil
	case dataQualityWriteModeReconciled:
		if request.reconciliationTaskID == "" {
			return reconciliation.ErrTaskIDRequired
		}
		if request.attemptCount <= 0 {
			return reconciliation.ErrAttemptCountInvalid
		}
		return nil
	default:
		return fmt.Errorf(
			"%w: %d",
			ErrDataQualityWriteModeInvalid,
			request.mode,
		)
	}
}
