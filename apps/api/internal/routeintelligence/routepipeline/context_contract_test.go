package routepipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type routePipelineContextContractTrajectoryReader struct {
	calls int
}

func (
	reader *routePipelineContextContractTrajectoryReader,
) GetTrajectoryByID(
	context.Context,
	string,
) (trajectory.FlightTrajectory, error) {
	reader.calls++

	return trajectory.FlightTrajectory{}, nil
}

func TestProcessRejectsNilContextBeforeTrajectoryRead(
	t *testing.T,
) {
	reader := &routePipelineContextContractTrajectoryReader{}
	pipeline := &Pipeline{
		trajectoryReader: reader,
	}

	result, err := pipeline.Process(
		nil,
		Request{
			TrajectoryID: "trajectory-one",
		},
	)

	if !errors.Is(
		err,
		ErrRoutePipelineContextRequired,
	) {
		t.Fatalf(
			"error = %v, want route pipeline context required",
			err,
		)
	}
	if reader.calls != 0 {
		t.Fatalf(
			"trajectory reader calls = %d, want 0",
			reader.calls,
		)
	}
	if result.PipelineVersion != "" ||
		result.TrajectoryID != "" {
		t.Fatalf(
			"result = %#v, want empty result",
			result,
		)
	}
}
