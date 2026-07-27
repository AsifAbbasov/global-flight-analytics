package trajectorybuilder

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type coordinate struct {
	latitude  float64
	longitude float64
}

func (value coordinate) equal(other coordinate) bool {
	return value.latitude == other.latitude && value.longitude == other.longitude
}

type timeInterval struct {
	start time.Time
	end   time.Time
}

type canonicalPoint struct {
	point               trajectory.TrackPoint4D
	observedAt          time.Time
	coordinate          coordinate
	coordinateAvailable bool
	inputIndex          int
}

type canonicalEvidence struct {
	windowStart     time.Time
	windowEnd       time.Time
	windowAvailable bool

	points   []canonicalPoint
	segments []trajectory.TrajectorySegment
	gaps     []trajectory.CoverageGap

	pointCount            int
	pointCountAvailable   bool
	segmentCount          int
	segmentCountAvailable bool
	gapCount              int
	gapCountAvailable     bool
	supportingPointCount  int

	limitations []flightfeatures.FeatureLimitation
}

type samplingMetrics struct {
	available      bool
	meanSeconds    float64
	maximumSeconds float64
}

type ratioMetric struct {
	available bool
	value     float64
}

type segmentStatusSummary struct {
	available         bool
	observedCount     int
	interpolatedCount int
	estimatedCount    int
	invalidCount      int
	limitations       []flightfeatures.FeatureLimitation
}

type pathPart struct {
	coordinates []coordinate
}
