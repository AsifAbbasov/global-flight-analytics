// Package airportproduction composes Airport Intelligence domain packages
// into a read-only production service backed by project PostgreSQL data.
package airportproduction

import (
	"context"
	"errors"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/history"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/overview"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/ranking"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/trends"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

const (
	Version = "airport-intelligence-production-v1"

	DefaultWindowDays = 30
	MinimumWindowDays = 1
	MaximumWindowDays = 365
)

var (
	ErrInvalidConfiguration = errors.New("Airport Intelligence production configuration is invalid")
	ErrInvalidRequest       = errors.New("Airport Intelligence request is invalid")
	ErrObservationsNotFound = errors.New("Airport Intelligence observations were not found")
	ErrInsufficientHistory  = errors.New("Airport Intelligence history is insufficient")
	ErrPostgresPoolRequired = errors.New("Airport Intelligence PostgreSQL pool is required")
)

type WindowRequest struct {
	AsOfTime time.Time
	Days     int
}

type Window struct {
	StartTime     time.Time
	EndTime       time.Time
	AsOfTime      time.Time
	CompletedDays int
}

type DailyQuery struct {
	ICAOCode    string
	WindowStart time.Time
	WindowEnd   time.Time
}

type DailyObservation struct {
	ICAOCode       string
	WindowStart    time.Time
	WindowEnd      time.Time
	Arrivals       int
	Departures     int
	ActiveAircraft int
	ActiveRoutes   int
	ObservedAt     time.Time
}

type ObservationReader interface {
	ListDaily(context.Context, DailyQuery) ([]DailyObservation, error)
}

type Limitation struct {
	Code    string
	Message string
}

type OverviewResult struct {
	Version     string
	Window      Window
	Overview    overview.Overview
	Limitations []Limitation
	GeneratedAt time.Time
}

type HistoryResult struct {
	Version     string
	Window      Window
	History     history.History
	Limitations []Limitation
	GeneratedAt time.Time
}

type TrendsResult struct {
	Version     string
	Window      Window
	Trends      trends.Trend
	Limitations []Limitation
	GeneratedAt time.Time
}

type RankingResult struct {
	Version     string
	Window      Window
	Ranking     ranking.Result
	Airports    map[string]airport.Airport
	Limitations []Limitation
	GeneratedAt time.Time
}

type ReadService interface {
	GetOverview(context.Context, string, WindowRequest) (OverviewResult, error)
	GetHistory(context.Context, string, WindowRequest) (HistoryResult, error)
	GetTrends(context.Context, string, WindowRequest) (TrendsResult, error)
	GetRanking(context.Context, WindowRequest) (RankingResult, error)
}
