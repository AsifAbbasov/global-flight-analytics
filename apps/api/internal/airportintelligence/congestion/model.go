// Package congestion derives a bounded, research-only airport activity pressure proxy
// from existing Airport Intelligence history. It does not model airport capacity,
// queues, slots, delays, runway occupancy, or official operational congestion.
package congestion

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/history"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

const (
	Version    = "airport-congestion-intelligence-v1"
	ScopeGuard = "relative_observed_activity_only_not_airport_capacity_delay_queue_slot_or_runway_congestion"
)

type Status string

const (
	StatusAvailable   Status = "available"
	StatusUnavailable Status = "unavailable"
)

type Input struct {
	History     history.History
	WindowStart time.Time
	WindowEnd   time.Time
	GeneratedAt time.Time
}

type Result struct {
	Version string
	Status  Status

	ICAOCode    string
	WindowStart time.Time
	WindowEnd   time.Time

	Current statistics.Statistics

	ObservedWindowCount           int
	ExpectedWindowCount           int
	GapWindowCount                int
	TrailingGapWindowCount        int
	CurrentWindowIsLatestExpected bool
	EvidenceCoverage              float64
	EvidenceSupport               float64

	BaselineWindowCount              int
	BaselineMedianMovementsPerHour   float64
	PriorPeakMovementsPerHour        float64
	CurrentToBaselineRatio           float64
	CurrentToBaselineRatioKnown      bool
	CurrentToPriorPeakRatio          float64
	CurrentToPriorPeakRatioKnown     bool
	CongestionScore                  float64
	CongestionScoreKnown             bool
	ExceedsPriorObservedActivityPeak bool

	ScoreSemantics string
	ScopeGuard     string
	Explanation    string
	GeneratedAt    time.Time
}
