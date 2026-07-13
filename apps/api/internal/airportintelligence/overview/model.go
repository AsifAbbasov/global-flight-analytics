package overview

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/passport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/ranking"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

type RankingSummary struct {
	Position              int
	TotalAirports         int
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

type Overview struct {
	Passport    passport.Passport
	Statistics  statistics.Statistics
	Ranking     RankingSummary
	GeneratedAt time.Time
}

type Input struct {
	Passport      passport.Passport
	Statistics    statistics.Statistics
	RankingResult ranking.Result
	GeneratedAt   time.Time
}
