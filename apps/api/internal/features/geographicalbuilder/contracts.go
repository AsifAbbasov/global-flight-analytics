package geographicalbuilder

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"

const Version = "geographical-feature-builder-v2"

const (
	GeographicalFeatureFieldCount  = flightfeatures.GeographicalRequiredFeatureFieldCount
	DefaultGeographicCellPrecision = 2
)

const earthMeanRadiusKM = 6371.0088
