package geographicalbuilder

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"

const Version = "geographical-feature-builder-v3"

const (
	GeographicalFeatureFieldCount  = flightfeatures.GeographicalRequiredFeatureFieldCount
	DefaultGeographicCellPrecision = 2
	MinimumGeographicCellPrecision = 1
	MaximumGeographicCellPrecision = 6
	contextCheckInterval           = 1024
)

const earthMeanRadiusKM = 6371.0088
