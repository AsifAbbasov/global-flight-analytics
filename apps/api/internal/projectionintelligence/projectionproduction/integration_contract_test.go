package projectionproduction

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionarrival"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontinuation"
)

var (
	_ KinematicProjector         = (*projectionbaseline.Baseline)(nil)
	_ HistoricalProjectionSource = (*projectioncontinuation.Baseline)(nil)
	_ HistoricalProjector        = (*HistoricalProjectionAdapter)(nil)
	_ ArrivalProjectionEstimator = (*projectionarrival.Estimator)(nil)
)
