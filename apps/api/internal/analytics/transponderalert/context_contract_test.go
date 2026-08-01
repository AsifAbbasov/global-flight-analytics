package transponderalert

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

type contextContractLatestStateReader struct {
	calls int
}

func (reader *contextContractLatestStateReader) GetLatestByICAO24(
	context.Context,
	string,
) (flightstate.FlightState, error) {
	reader.calls++

	return flightstate.FlightState{}, nil
}

func TestServiceRejectsNilContextBeforeLatestStateRead(
	t *testing.T,
) {
	reader := &contextContractLatestStateReader{}
	service, err := NewService(ServiceConfig{
		LatestStateReader: reader,
	})
	if err != nil {
		t.Fatalf(
			"create service: %v",
			err,
		)
	}

	_, err = service.GetLatest(
		nil,
		"4A001A",
	)

	if !errors.Is(
		err,
		ErrLatestEvidenceContextRequired,
	) {
		t.Fatalf(
			"error = %v, want latest evidence context required",
			err,
		)
	}
	if reader.calls != 0 {
		t.Fatalf(
			"latest state reader calls = %d, want 0",
			reader.calls,
		)
	}
}
