package stabilityproduction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
)

type contextContractProjectionReader struct {
	calls int
}

func (
	reader *contextContractProjectionReader,
) ReadProjection(
	context.Context,
	ProjectionRequest,
) (projectionproduction.Result, error) {
	reader.calls++

	return projectionproduction.Result{}, nil
}

func TestGetRejectsNilContextBeforeProjectionReads(
	t *testing.T,
) {
	reader := &contextContractProjectionReader{}
	service, err := New(Config{
		ProjectionReader: reader,
		Now: func() time.Time {
			return time.Date(
				2026,
				time.August,
				1,
				12,
				0,
				0,
				0,
				time.UTC,
			)
		},
	})
	if err != nil {
		t.Fatalf(
			"create service: %v",
			err,
		)
	}

	result, err := service.Get(
		nil,
		Request{},
	)

	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf(
			"error = %v, want stability production context required",
			err,
		)
	}
	if reader.calls != 0 {
		t.Fatalf(
			"projection reader calls = %d, want 0",
			reader.calls,
		)
	}
	if result.Version != "" ||
		result.TrajectoryID != "" ||
		len(result.Projections) != 0 ||
		result.InputFingerprint != "" {
		t.Fatalf(
			"result = %#v, want empty result",
			result,
		)
	}
}
