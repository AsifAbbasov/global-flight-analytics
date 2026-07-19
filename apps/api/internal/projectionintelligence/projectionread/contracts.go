package projectionread

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
)

type Request struct {
	TrajectoryID      string
	AsOfTime          time.Time
	RequestedDuration time.Duration
}

type SnapshotRequest struct {
	TrajectoryID string
	AsOfTime     time.Time
}

type DataSource interface {
	LoadSnapshot(
		context.Context,
		SnapshotRequest,
	) (Snapshot, error)
}

type Composer interface {
	Compose(
		projectionproduction.Request,
	) (projectionproduction.Result, error)
}
