package trends

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/history"
)

type Direction string

const (
	DirectionIncreasing Direction = "increasing"
	DirectionDecreasing Direction = "decreasing"
	DirectionStable     Direction = "stable"
)

type Input struct {
	History     history.History
	GeneratedAt time.Time
}

type Point struct {
	WindowStart      time.Time
	WindowEnd        time.Time
	TotalMovements   int
	MovementsPerHour float64
	ActiveRoutes     int
	CoverageScore    float64
	FreshnessScore   float64
}

type Trend struct {
	ICAOCode    string
	WindowStart time.Time
	WindowEnd   time.Time

	ComparedWindows int
	WindowDuration  time.Duration
	Direction       Direction

	Baseline Point
	Current  Point
	Peak     Point

	TotalMovementsChange               int
	MovementsPerHourChange             float64
	MovementsPerHourChangePercent      float64
	MovementsPerHourChangePercentKnown bool
	ActiveRoutesChange                 int
	CoverageScoreChange                float64
	FreshnessScoreChange               float64

	GapCount         int
	GapDuration      time.Duration
	ObservedDuration time.Duration
	ContinuityScore  float64

	GeneratedAt time.Time
}
