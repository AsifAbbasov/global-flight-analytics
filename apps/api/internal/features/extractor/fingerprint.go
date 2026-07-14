package extractor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const fingerprintPrefix = "sha256:"

type canonicalTrajectory struct {
	ID               string
	IdentityKey      string
	IdentityBasis    trajectory.FlightIdentityBasis
	SplitReason      trajectory.FlightSplitReason
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
	Points           []canonicalTrackPoint
	Segments         []canonicalSegment
	CoverageGaps     []canonicalCoverageGap
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type canonicalTrackPoint struct {
	ID                       string
	FlightStateID            string
	FlightID                 string
	AircraftID               string
	ICAO24                   string
	Callsign                 string
	Latitude                 float64
	Longitude                float64
	BarometricAltitudeM      float64
	BarometricAltitudeStatus string
	GeometricAltitudeM       float64
	GeometricAltitudeStatus  string
	VelocityMPS              float64
	HeadingDegrees           float64
	VerticalRateMPS          float64
	OnGround                 bool
	OriginCountry            string
	ObservedAt               time.Time
	SourceName               string
}

type canonicalSegment struct {
	ID              string
	TrajectoryID    string
	FlightID        string
	AircraftID      string
	ICAO24          string
	Callsign        string
	SequenceNumber  int
	Status          trajectory.SegmentStatus
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

type canonicalCoverageGap struct {
	ID                string
	TrajectoryID      string
	PreviousSegmentID string
	NextSegmentID     string
	ICAO24            string
	StartTime         time.Time
	EndTime           time.Time
	DurationSeconds   int64
	DistanceKM        float64
	Reason            trajectory.CoverageGapReason
	FilledBy          string
	CreatedAt         time.Time
}

func fingerprintTrajectory(
	item trajectory.FlightTrajectory,
) (string, error) {
	canonical := canonicalizeTrajectory(item)

	payload, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf(
			"marshal trajectory feature fingerprint input: %w",
			err,
		)
	}

	sum := sha256.Sum256(payload)

	return fingerprintPrefix + hex.EncodeToString(sum[:]), nil
}

func canonicalizeTrajectory(
	item trajectory.FlightTrajectory,
) canonicalTrajectory {
	points := make(
		[]canonicalTrackPoint,
		0,
		len(item.Points),
	)
	for _, point := range item.Points {
		points = append(
			points,
			canonicalTrackPoint{
				ID:                       point.ID,
				FlightStateID:            point.FlightStateID,
				FlightID:                 point.FlightID,
				AircraftID:               point.AircraftID,
				ICAO24:                   point.ICAO24,
				Callsign:                 point.Callsign,
				Latitude:                 point.Latitude,
				Longitude:                point.Longitude,
				BarometricAltitudeM:      point.BarometricAltitudeM,
				BarometricAltitudeStatus: string(point.BarometricAltitudeStatus),
				GeometricAltitudeM:       point.GeometricAltitudeM,
				GeometricAltitudeStatus:  string(point.GeometricAltitudeStatus),
				VelocityMPS:              point.VelocityMPS,
				HeadingDegrees:           point.HeadingDegrees,
				VerticalRateMPS:          point.VerticalRateMPS,
				OnGround:                 point.OnGround,
				OriginCountry:            point.OriginCountry,
				ObservedAt:               point.ObservedAt.UTC(),
				SourceName:               point.SourceName,
			},
		)
	}

	segments := make(
		[]canonicalSegment,
		0,
		len(item.Segments),
	)
	for _, segment := range item.Segments {
		segments = append(
			segments,
			canonicalSegment{
				ID:              segment.ID,
				TrajectoryID:    segment.TrajectoryID,
				FlightID:        segment.FlightID,
				AircraftID:      segment.AircraftID,
				ICAO24:          segment.ICAO24,
				Callsign:        segment.Callsign,
				SequenceNumber:  segment.SequenceNumber,
				Status:          segment.Status,
				QualityScore:    segment.QualityScore,
				StartTime:       segment.StartTime.UTC(),
				EndTime:         segment.EndTime.UTC(),
				DurationSeconds: segment.DurationSeconds,
				StartLatitude:   segment.StartLatitude,
				StartLongitude:  segment.StartLongitude,
				EndLatitude:     segment.EndLatitude,
				EndLongitude:    segment.EndLongitude,
				PointCount:      segment.PointCount,
				SourceName:      segment.SourceName,
				CreatedAt:       segment.CreatedAt.UTC(),
			},
		)
	}

	coverageGaps := make(
		[]canonicalCoverageGap,
		0,
		len(item.CoverageGaps),
	)
	for _, gap := range item.CoverageGaps {
		coverageGaps = append(
			coverageGaps,
			canonicalCoverageGap{
				ID:                gap.ID,
				TrajectoryID:      gap.TrajectoryID,
				PreviousSegmentID: gap.PreviousSegmentID,
				NextSegmentID:     gap.NextSegmentID,
				ICAO24:            gap.ICAO24,
				StartTime:         gap.StartTime.UTC(),
				EndTime:           gap.EndTime.UTC(),
				DurationSeconds:   gap.DurationSeconds,
				DistanceKM:        gap.DistanceKm,
				Reason:            gap.Reason,
				FilledBy:          gap.FilledBy,
				CreatedAt:         gap.CreatedAt.UTC(),
			},
		)
	}

	return canonicalTrajectory{
		ID:               item.ID,
		IdentityKey:      item.IdentityKey,
		IdentityBasis:    item.IdentityBasis,
		SplitReason:      item.SplitReason,
		FlightID:         item.FlightID,
		AircraftID:       item.AircraftID,
		ICAO24:           item.ICAO24,
		Callsign:         item.Callsign,
		StartTime:        item.StartTime.UTC(),
		EndTime:          item.EndTime.UTC(),
		DurationSeconds:  item.DurationSeconds,
		SegmentCount:     item.SegmentCount,
		PointCount:       item.PointCount,
		CoverageGapCount: item.CoverageGapCount,
		QualityScore:     item.QualityScore,
		SourceName:       item.SourceName,
		Points:           points,
		Segments:         segments,
		CoverageGaps:     coverageGaps,
		CreatedAt:        item.CreatedAt.UTC(),
		UpdatedAt:        item.UpdatedAt.UTC(),
	}
}
