package trajectorybuilder

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"

const Version = "trajectory-feature-builder-v2"

const TrajectoryFeatureFieldCount = flightfeatures.TrajectoryRequiredFeatureFieldCount

const (
	earthMeanRadiusKM    = 6371.0088
	contextCheckInterval = 1024
	pathRatioTolerance   = 1e-12
)
