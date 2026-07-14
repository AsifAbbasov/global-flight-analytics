package routecontext

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

type ConfidenceLevel string

const (
	ConfidenceLevelNone   ConfidenceLevel = "none"
	ConfidenceLevelLow    ConfidenceLevel = "low"
	ConfidenceLevelMedium ConfidenceLevel = "medium"
	ConfidenceLevelHigh   ConfidenceLevel = "high"
)

type Notice struct {
	Code    string
	Message string
}

type Confidence struct {
	Score   float64
	Level   ConfidenceLevel
	Reasons []Notice
}

type AirportCandidate struct {
	Airport    airport.Airport
	DistanceKM float64
	Confidence Confidence
}

type Context struct {
	ICAO24       string
	TrajectoryID string
	Origin       *AirportCandidate
	Destination  *AirportCandidate
	Confidence   Confidence
	Limitations  []Notice
	GeneratedAt  time.Time
}
