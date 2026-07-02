package dto

import (
	"time"

	domaintrajectory "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type Trajectory struct {
	ID               string              `json:"id"`
	FlightID         string              `json:"flight_id"`
	AircraftID       string              `json:"aircraft_id"`
	ICAO24           string              `json:"icao24"`
	Callsign         string              `json:"callsign"`
	StartTime        time.Time           `json:"start_time"`
	EndTime          time.Time           `json:"end_time"`
	DurationSeconds  int64               `json:"duration_seconds"`
	SegmentCount     int                 `json:"segment_count"`
	PointCount       int                 `json:"point_count"`
	CoverageGapCount int                 `json:"coverage_gap_count"`
	QualityScore     float64             `json:"quality_score"`
	SourceName       string              `json:"source_name"`
	Segments         []TrajectorySegment `json:"segments"`
	CoverageGaps     []CoverageGap       `json:"coverage_gaps"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type TrajectorySegment struct {
	ID              string    `json:"id"`
	TrajectoryID    string    `json:"trajectory_id"`
	FlightID        string    `json:"flight_id"`
	AircraftID      string    `json:"aircraft_id"`
	ICAO24          string    `json:"icao24"`
	Callsign        string    `json:"callsign"`
	SequenceNumber  int       `json:"sequence_number"`
	Status          string    `json:"status"`
	QualityScore    float64   `json:"quality_score"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationSeconds int64     `json:"duration_seconds"`
	StartLatitude   float64   `json:"start_latitude"`
	StartLongitude  float64   `json:"start_longitude"`
	EndLatitude     float64   `json:"end_latitude"`
	EndLongitude    float64   `json:"end_longitude"`
	PointCount      int       `json:"point_count"`
	SourceName      string    `json:"source_name"`
	CreatedAt       time.Time `json:"created_at"`
}

type CoverageGap struct {
	ID                string    `json:"id"`
	TrajectoryID      string    `json:"trajectory_id"`
	PreviousSegmentID string    `json:"previous_segment_id"`
	NextSegmentID     string    `json:"next_segment_id"`
	ICAO24            string    `json:"icao24"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	DurationSeconds   int64     `json:"duration_seconds"`
	DistanceKm        float64   `json:"distance_km"`
	Reason            string    `json:"reason"`
	FilledBy          string    `json:"filled_by"`
	CreatedAt         time.Time `json:"created_at"`
}

func ToTrajectory(item domaintrajectory.FlightTrajectory) Trajectory {
	return Trajectory{
		ID:               item.ID,
		FlightID:         item.FlightID,
		AircraftID:       item.AircraftID,
		ICAO24:           item.ICAO24,
		Callsign:         item.Callsign,
		StartTime:        item.StartTime,
		EndTime:          item.EndTime,
		DurationSeconds:  item.DurationSeconds,
		SegmentCount:     item.SegmentCount,
		PointCount:       item.PointCount,
		CoverageGapCount: item.CoverageGapCount,
		QualityScore:     item.QualityScore,
		SourceName:       item.SourceName,
		Segments:         ToTrajectorySegments(item.Segments),
		CoverageGaps:     ToCoverageGaps(item.CoverageGaps),
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}

func ToTrajectorySegments(items []domaintrajectory.TrajectorySegment) []TrajectorySegment {
	result := make([]TrajectorySegment, 0, len(items))

	for _, item := range items {
		result = append(result, TrajectorySegment{
			ID:              item.ID,
			TrajectoryID:    item.TrajectoryID,
			FlightID:        item.FlightID,
			AircraftID:      item.AircraftID,
			ICAO24:          item.ICAO24,
			Callsign:        item.Callsign,
			SequenceNumber:  item.SequenceNumber,
			Status:          string(item.Status),
			QualityScore:    item.QualityScore,
			StartTime:       item.StartTime,
			EndTime:         item.EndTime,
			DurationSeconds: item.DurationSeconds,
			StartLatitude:   item.StartLatitude,
			StartLongitude:  item.StartLongitude,
			EndLatitude:     item.EndLatitude,
			EndLongitude:    item.EndLongitude,
			PointCount:      item.PointCount,
			SourceName:      item.SourceName,
			CreatedAt:       item.CreatedAt,
		})
	}

	return result
}

func ToCoverageGaps(items []domaintrajectory.CoverageGap) []CoverageGap {
	result := make([]CoverageGap, 0, len(items))

	for _, item := range items {
		result = append(result, CoverageGap{
			ID:                item.ID,
			TrajectoryID:      item.TrajectoryID,
			PreviousSegmentID: item.PreviousSegmentID,
			NextSegmentID:     item.NextSegmentID,
			ICAO24:            item.ICAO24,
			StartTime:         item.StartTime,
			EndTime:           item.EndTime,
			DurationSeconds:   item.DurationSeconds,
			DistanceKm:        item.DistanceKm,
			Reason:            string(item.Reason),
			FilledBy:          item.FilledBy,
			CreatedAt:         item.CreatedAt,
		})
	}

	return result
}
