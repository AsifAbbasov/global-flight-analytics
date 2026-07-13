package ranking

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

type ActivityWeights struct {
	Movements    float64
	Routes       float64
	Observations float64
	Intensity    float64
}

type ConfidenceWeights struct {
	Coverage  float64
	Freshness float64
}

type Input struct {
	Statistics        []statistics.Statistics
	ActivityWeights   ActivityWeights
	ConfidenceWeights ConfidenceWeights
}

type AirportRank struct {
	Position              int
	ICAOCode              string
	ActivityScore         float64
	DataConfidence        float64
	MovementsComponent    float64
	RoutesComponent       float64
	ObservationsComponent float64
	IntensityComponent    float64
	CoverageScore         float64
	FreshnessScore        float64
	TotalMovements        int
	ActiveRoutes          int
	ObservedSamples       int
	ExpectedSamples       int
	MovementsPerHour      float64
	ActiveAircraft        int
}

type Result struct {
	WindowStart       time.Time
	WindowEnd         time.Time
	ActivityWeights   ActivityWeights
	ConfidenceWeights ConfidenceWeights
	Airports          []AirportRank
}
