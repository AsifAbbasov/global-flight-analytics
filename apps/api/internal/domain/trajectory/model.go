package trajectory

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

type SegmentStatus string

const (
	SegmentStatusObserved     SegmentStatus = "observed"
	SegmentStatusInterpolated SegmentStatus = "interpolated"
	SegmentStatusEstimated    SegmentStatus = "estimated"
	SegmentStatusInvalid      SegmentStatus = "invalid"
)

type CoverageGapReason string

const (
	CoverageGapReasonTimeGap      CoverageGapReason = "time_gap"
	CoverageGapReasonMovementJump CoverageGapReason = "movement_jump"
	CoverageGapReasonUnknown      CoverageGapReason = "unknown"
)

type FlightIdentityBasis string

const (
	FlightIdentityBasisSourceFlightID       FlightIdentityBasis = "source_flight_id"
	FlightIdentityBasisCallsignAndStartTime FlightIdentityBasis = "callsign_and_start_time"
	FlightIdentityBasisAircraftAndStartTime FlightIdentityBasis = "aircraft_and_start_time"
)

type FlightSplitReason string

const (
	FlightSplitReasonInitialObservation    FlightSplitReason = "initial_observation"
	FlightSplitReasonSourceFlightIDChanged FlightSplitReason = "source_flight_id_changed"
	FlightSplitReasonCallsignChanged       FlightSplitReason = "callsign_changed"
	FlightSplitReasonGroundCycle           FlightSplitReason = "ground_cycle"
)

type TrackPoint4D struct {
	ID                       string
	FlightStateID            string
	FlightID                 string
	AircraftID               string
	ICAO24                   string
	Callsign                 string
	Latitude                 float64
	Longitude                float64
	BarometricAltitudeM      float64
	BarometricAltitudeStatus flightstate.AltitudeStatus
	GeometricAltitudeM       float64
	GeometricAltitudeStatus  flightstate.AltitudeStatus
	VelocityMPS              float64
	HeadingDegrees           float64
	VerticalRateMPS          float64
	OnGround                 bool
	OriginCountry            string
	ObservedAt               time.Time
	SourceName               string
}

type CoverageGap struct {
	ID                string
	TrajectoryID      string
	PreviousSegmentID string
	NextSegmentID     string
	ICAO24            string
	StartTime         time.Time
	EndTime           time.Time
	DurationSeconds   int64
	DistanceKm        float64
	Reason            CoverageGapReason
	FilledBy          string
	CreatedAt         time.Time
}

type TrajectorySegment struct {
	ID              string
	TrajectoryID    string
	FlightID        string
	AircraftID      string
	ICAO24          string
	Callsign        string
	SequenceNumber  int
	Status          SegmentStatus
	QualityScore    float64
	StartTime       time.Time
	EndTime         time.Time
	DurationSeconds int64
	StartLatitude   float64
	StartLongitude  float64
	EndLatitude     float64
	EndLongitude    float64
	PointCount      int
	SourceName      string
	CreatedAt       time.Time
}

type FlightTrajectory struct {
	ID               string
	IdentityKey      string
	IdentityBasis    FlightIdentityBasis
	SplitReason      FlightSplitReason
	FlightID         string
	AircraftID       string
	ICAO24           string
	Callsign         string
	StartTime        time.Time
	EndTime          time.Time
	DurationSeconds  int64
	SegmentCount     int
	PointCount       int
	CoverageGapCount int
	QualityScore     float64
	SourceName       string
	Points           []TrackPoint4D
	Segments         []TrajectorySegment
	CoverageGaps     []CoverageGap
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
