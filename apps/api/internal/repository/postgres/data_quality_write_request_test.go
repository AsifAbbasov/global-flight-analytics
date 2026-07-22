package postgres

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
)

func TestDataQualityWriteRequestUsesExplicitModes(t *testing.T) {
	live := newLiveDataQualityWriteRequest(
		flightstate.FlightState{},
		dataquality.DataQuality{},
	)
	if live.mode != dataQualityWriteModeLive || live.validate() != nil {
		t.Fatalf("live request = %#v", live)
	}

	reconciled, err := newReconciledDataQualityWriteRequest(
		" 11111111-1111-1111-1111-111111111111 ",
		1,
		flightstate.FlightState{},
		dataquality.DataQuality{},
	)
	if err != nil {
		t.Fatalf("reconciled request error = %v", err)
	}
	if reconciled.mode != dataQualityWriteModeReconciled ||
		reconciled.reconciliationTaskID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("reconciled request = %#v", reconciled)
	}
}

func TestDataQualityWriteRequestRejectsSentinelInputs(t *testing.T) {
	_, err := newReconciledDataQualityWriteRequest(
		" ",
		1,
		flightstate.FlightState{},
		dataquality.DataQuality{},
	)
	if !errors.Is(err, reconciliation.ErrTaskIDRequired) {
		t.Fatalf("blank task error = %v", err)
	}

	err = (dataQualityWriteRequest{}).validate()
	if !errors.Is(err, ErrDataQualityWriteModeInvalid) {
		t.Fatalf("zero mode error = %v", err)
	}
}
