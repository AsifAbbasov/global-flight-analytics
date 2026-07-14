package historicalread

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

const (
	Version = "historical-read-repository-v1"

	DefaultDatasetLimit = 10_000
	MaximumDatasetLimit = 100_000
)

type Query struct {
	Window historicalcontract.TimeWindow
	Limit  int
}

type FlightRecord struct {
	ID          string
	AircraftID  string
	Callsign    string
	Status      string
	FirstSeenAt time.Time
	LastSeenAt  time.Time
	UpdatedAt   time.Time
}

type TrajectoryRecord struct {
	ID               string
	FlightID         string
	AircraftID       string
	ICAO24           string
	Callsign         string
	StartTime        time.Time
	EndTime          time.Time
	SegmentCount     int
	PointCount       int
	CoverageGapCount int
	QualityScore     float64
	SourceName       string
	UpdatedAt        time.Time
}

type ObservationRecord struct {
	ID         string
	FlightID   string
	AircraftID string
	ICAO24     string
	Callsign   string
	Latitude   *float64
	Longitude  *float64
	OnGround   *bool
	ObservedAt time.Time
	SourceName string
	CreatedAt  time.Time
}

type RouteRecord struct {
	ID                     string
	TrajectoryID           string
	AsOfTime               time.Time
	InputFingerprint       string
	Status                 string
	ConfidenceLevel        string
	ValidationWarningCount int
	RouteJSON              []byte
	StoredAt               time.Time
}

type Snapshot struct {
	Version string
	Query   Query

	Flights      []FlightRecord
	Trajectories []TrajectoryRecord
	Observations []ObservationRecord
	Routes       []RouteRecord

	FlightLimitReached      bool
	TrajectoryLimitReached  bool
	ObservationLimitReached bool
	RouteLimitReached       bool
}

func (snapshot Snapshot) Clone() Snapshot {
	cloned := snapshot
	cloned.Flights = append([]FlightRecord(nil), snapshot.Flights...)
	cloned.Trajectories = append([]TrajectoryRecord(nil), snapshot.Trajectories...)
	cloned.Observations = cloneObservations(snapshot.Observations)
	cloned.Routes = cloneRoutes(snapshot.Routes)

	return cloned
}

func cloneObservations(items []ObservationRecord) []ObservationRecord {
	cloned := make([]ObservationRecord, 0, len(items))
	for _, item := range items {
		copied := item
		if item.Latitude != nil {
			value := *item.Latitude
			copied.Latitude = &value
		}
		if item.Longitude != nil {
			value := *item.Longitude
			copied.Longitude = &value
		}
		if item.OnGround != nil {
			value := *item.OnGround
			copied.OnGround = &value
		}
		cloned = append(cloned, copied)
	}

	return cloned
}

func cloneRoutes(items []RouteRecord) []RouteRecord {
	cloned := make([]RouteRecord, 0, len(items))
	for _, item := range items {
		copied := item
		copied.RouteJSON = append([]byte(nil), item.RouteJSON...)
		cloned = append(cloned, copied)
	}

	return cloned
}

type Repository interface {
	Read(context.Context, Query) (Snapshot, error)
}
