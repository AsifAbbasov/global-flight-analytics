package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
)

type projectionIntelligenceApplicationReaderStub struct {
	result  projectionproduction.Result
	err     error
	request projectionread.Request
	calls   int
}

func (
	reader *projectionIntelligenceApplicationReaderStub,
) Get(
	_ context.Context,
	request projectionread.Request,
) (projectionproduction.Result, error) {
	reader.calls++
	reader.request = request

	return reader.result.Clone(),
		reader.err
}

func TestProjectionIntelligenceReaderAdapterMapsRequest(
	t *testing.T,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		16,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	expectedResult :=
		projectionproduction.Result{
			Version: projectionproduction.Version,
		}
	reader :=
		&projectionIntelligenceApplicationReaderStub{
			result: expectedResult,
		}
	adapter :=
		projectionIntelligenceReaderAdapter{
			reader: reader,
		}

	result, err :=
		adapter.GetProjectionIntelligence(
			context.Background(),
			handlers.
				ProjectionIntelligenceReadRequest{
				TrajectoryID:      "73aa02ab-7061-4e9e-a238-d32710371ee3",
				AsOfTime:          asOfTime,
				RequestedDuration: 5 * time.Minute,
			},
		)
	if err != nil {
		t.Fatalf(
			"GetProjectionIntelligence() error = %v",
			err,
		)
	}
	if reader.calls != 1 ||
		reader.request.TrajectoryID !=
			"73aa02ab-7061-4e9e-a238-d32710371ee3" ||
		!reader.request.AsOfTime.Equal(
			asOfTime,
		) ||
		reader.request.RequestedDuration !=
			5*time.Minute ||
		result.Version !=
			projectionproduction.Version {
		t.Fatalf(
			"unexpected adapter result or request: result=%#v request=%#v calls=%d",
			result,
			reader.request,
			reader.calls,
		)
	}
}

func TestProjectionIntelligenceReaderAdapterMapsErrors(
	t *testing.T,
) {
	tests := []struct {
		name  string
		input error
		want  error
	}{
		{
			name: "trajectory not found",
			input: projectionread.
				ErrTrajectoryNotFound,
			want: handlers.
				ErrProjectionIntelligenceNotFound,
		},
		{
			name: "service unavailable",
			input: projectionread.
				ErrServiceUnavailable,
			want: handlers.
				ErrProjectionIntelligenceServiceUnavailable,
		},
		{
			name: "invalid request",
			input: projectionread.
				ErrInvalidRequest,
			want: handlers.
				ErrProjectionIntelligenceInvalidRequest,
		},
		{
			name:  "context timeout preserved",
			input: context.DeadlineExceeded,
			want:  context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				reader :=
					&projectionIntelligenceApplicationReaderStub{
						err: test.input,
					}
				adapter :=
					projectionIntelligenceReaderAdapter{
						reader: reader,
					}

				_, err :=
					adapter.GetProjectionIntelligence(
						context.Background(),
						handlers.
							ProjectionIntelligenceReadRequest{
							TrajectoryID:      "73aa02ab-7061-4e9e-a238-d32710371ee3",
							AsOfTime:          time.Now().UTC(),
							RequestedDuration: time.Minute,
						},
					)
				if !errors.Is(
					err,
					test.want,
				) {
					t.Fatalf(
						"error = %v, want %v",
						err,
						test.want,
					)
				}
			},
		)
	}
}

func TestNewProjectionIntelligencePostgresReaderRejectsNilPool(
	t *testing.T,
) {
	_, err :=
		newProjectionIntelligencePostgresReader(
			nil,
		)
	if err == nil {
		t.Fatal(
			"expected nil PostgreSQL pool to be rejected",
		)
	}
}
