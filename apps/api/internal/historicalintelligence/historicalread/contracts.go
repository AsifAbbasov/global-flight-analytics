package historicalread

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

const (
	Version = "historical-read-repository-v2"

	DefaultDatasetLimit = 10_000
	MaximumDatasetLimit = 100_000

	DefaultRoutePayloadByteLimit int64 = 16 * 1024 * 1024
	MaximumRoutePayloadByteLimit int64 = 64 * 1024 * 1024

	QualityScoreDecimalPlaces = 12
	CoordinateDecimalPlaces   = 8

	SnapshotIsolationRepeatableRead    = "repeatable_read"
	SnapshotIsolationCallerTransaction = "caller_transaction"
)

const (
	DatasetFlights      = "flights"
	DatasetTrajectories = "flight_trajectories"
	DatasetObservations = "flight_states"
	DatasetRoutes       = "flight_route_results"
)

type Query struct {
	Window historicalcontract.TimeWindow
	Limit  int

	RoutePayloadByteLimit int64
}

func (query Query) Equal(other Query) bool {
	return query.Window.StartTime.UTC().Equal(
		other.Window.StartTime.UTC(),
	) &&
		query.Window.EndTime.UTC().Equal(
			other.Window.EndTime.UTC(),
		) &&
		query.Window.AsOfTime.UTC().Equal(
			other.Window.AsOfTime.UTC(),
		) &&
		query.Limit == other.Limit &&
		query.RoutePayloadByteLimit ==
			other.RoutePayloadByteLimit
}

type FlightRecord struct {
	ID string

	AircraftID          string
	AircraftIDAvailable bool
	Callsign            string
	CallsignAvailable   bool

	Status      string
	FirstSeenAt time.Time
	LastSeenAt  time.Time
	UpdatedAt   time.Time
}

type TrajectoryRecord struct {
	ID string

	FlightID            string
	FlightIDAvailable   bool
	AircraftID          string
	AircraftIDAvailable bool
	ICAO24              string
	Callsign            string
	CallsignAvailable   bool

	StartTime time.Time
	EndTime   time.Time

	SegmentCount     int
	PointCount       int
	CoverageGapCount int
	QualityScore     float64
	SourceName       string
	UpdatedAt        time.Time
}

type ObservationRecord struct {
	ID string

	FlightID            string
	FlightIDAvailable   bool
	AircraftID          string
	AircraftIDAvailable bool
	ICAO24              string
	Callsign            string
	CallsignAvailable   bool

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
	EventStartTime         time.Time
	EventEndTime           time.Time
	AsOfTime               time.Time
	InputFingerprint       string
	Status                 string
	ConfidenceLevel        string
	ValidationWarningCount int
	StoredAt               time.Time
	PayloadBytes           int64
	PayloadFingerprint     string

	Result          routecontract.Result
	ResultAvailable bool

	// Deprecated: RouteJSON is retained only for source compatibility with older
	// test fixtures. Production PostgreSQL reads decode route payloads inside
	// this repository and downstream builders use ResultAt instead of raw JSON.
	RouteJSON []byte
}

type Snapshot struct {
	Version        string
	IsolationLevel string
	Query          Query

	Flights      []FlightRecord
	Trajectories []TrajectoryRecord
	Observations []ObservationRecord
	Routes       []RouteRecord

	FlightMatchedCount      int64
	TrajectoryMatchedCount  int64
	ObservationMatchedCount int64
	RouteMatchedCount       int64

	RoutePayloadBytes      int64
	RouteTotalPayloadBytes int64

	FlightLimitReached      bool
	TrajectoryLimitReached  bool
	ObservationLimitReached bool
	RouteLimitReached       bool
	RouteByteLimitReached   bool
}

func (snapshot Snapshot) Clone() Snapshot {
	cloned := snapshot
	cloned.Flights = append([]FlightRecord(nil), snapshot.Flights...)
	cloned.Trajectories = append([]TrajectoryRecord(nil), snapshot.Trajectories...)
	cloned.Observations = cloneObservations(snapshot.Observations)
	cloned.Routes = cloneRoutes(snapshot.Routes)

	return cloned
}

func (snapshot Snapshot) TotalForSource(sourceName string) int64 {
	switch strings.TrimSpace(sourceName) {
	case DatasetFlights:
		return inferredMatchedCount(
			snapshot.FlightMatchedCount,
			len(snapshot.Flights),
			snapshot.FlightLimitReached,
		)
	case DatasetTrajectories:
		return inferredMatchedCount(
			snapshot.TrajectoryMatchedCount,
			len(snapshot.Trajectories),
			snapshot.TrajectoryLimitReached,
		)
	case DatasetObservations:
		return inferredMatchedCount(
			snapshot.ObservationMatchedCount,
			len(snapshot.Observations),
			snapshot.ObservationLimitReached,
		)
	case DatasetRoutes, "route_intelligence":
		return inferredMatchedCount(
			snapshot.RouteMatchedCount,
			len(snapshot.Routes),
			snapshot.RouteLimitReached,
		)
	default:
		return 0
	}
}

func inferredMatchedCount(
	explicit int64,
	loaded int,
	limited bool,
) int64 {
	if explicit > 0 || loaded == 0 {
		return explicit
	}
	if limited {
		return int64(loaded) + 1
	}
	return int64(loaded)
}

func RepresentedCoverage(representedCount int, totalCount int64) float64 {
	if representedCount < 0 || totalCount < 0 {
		return 0
	}
	if totalCount == 0 {
		if representedCount == 0 {
			return 1
		}
		return 0
	}

	ratio := float64(representedCount) / float64(totalCount)
	if math.IsNaN(ratio) || math.IsInf(ratio, 0) || ratio < 0 {
		return 0
	}
	if ratio > 1 {
		return 1
	}
	return ratio
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
		copied.Result = item.Result.Clone()
		copied.RouteJSON = append([]byte(nil), item.RouteJSON...)
		cloned = append(cloned, copied)
	}

	return cloned
}

type PeriodQueries struct {
	Previous Query
	Current  Query
}

type PeriodSnapshots struct {
	Previous Snapshot
	Current  Snapshot
}

func (snapshots PeriodSnapshots) Clone() PeriodSnapshots {
	return PeriodSnapshots{
		Previous: snapshots.Previous.Clone(),
		Current:  snapshots.Current.Clone(),
	}
}

type Repository interface {
	Read(context.Context, Query) (Snapshot, error)
}

type PeriodRepository interface {
	ReadPeriods(
		context.Context,
		PeriodQueries,
	) (PeriodSnapshots, error)
}
