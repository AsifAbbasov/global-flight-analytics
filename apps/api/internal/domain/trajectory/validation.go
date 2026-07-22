package trajectory

import (
	"errors"
	"math"
	"time"
)

var (
	ErrTrajectoryCoordinatesInvalid = errors.New("trajectory coordinates are invalid")
	ErrTrajectoryTimeRangeInvalid   = errors.New("trajectory time range is invalid")
	ErrTrajectoryDurationInvalid    = errors.New("trajectory duration is inconsistent")
	ErrTrajectoryCountInvalid       = errors.New("trajectory count is inconsistent")
	ErrTrajectoryQualityInvalid     = errors.New("trajectory quality score is invalid")
	ErrTrajectoryDistanceInvalid    = errors.New("trajectory distance is invalid")
)

func (value TrackPoint4D) Validate() error {
	if !isValidTrajectoryCoordinates(value.Latitude, value.Longitude) {
		return ErrTrajectoryCoordinatesInvalid
	}
	if value.ObservedAt.IsZero() {
		return ErrTrajectoryTimeRangeInvalid
	}
	return nil
}

func (value CoverageGap) Validate() error {
	if err := validateTimeRange(value.StartTime, value.EndTime, value.DurationSeconds); err != nil {
		return err
	}
	if math.IsNaN(value.DistanceKm) || math.IsInf(value.DistanceKm, 0) || value.DistanceKm < 0 {
		return ErrTrajectoryDistanceInvalid
	}
	return nil
}

func (value TrajectorySegment) Validate() error {
	if err := validateTimeRange(value.StartTime, value.EndTime, value.DurationSeconds); err != nil {
		return err
	}
	if !isValidTrajectoryCoordinates(value.StartLatitude, value.StartLongitude) ||
		!isValidTrajectoryCoordinates(value.EndLatitude, value.EndLongitude) {
		return ErrTrajectoryCoordinatesInvalid
	}
	if value.PointCount < 0 {
		return ErrTrajectoryCountInvalid
	}
	if !isValidTrajectoryQuality(value.QualityScore) {
		return ErrTrajectoryQualityInvalid
	}
	return nil
}

func (value FlightTrajectory) Validate() error {
	if err := validateTimeRange(value.StartTime, value.EndTime, value.DurationSeconds); err != nil {
		return err
	}
	if value.PointCount < 0 || value.SegmentCount < 0 || value.CoverageGapCount < 0 {
		return ErrTrajectoryCountInvalid
	}
	if value.Points != nil && value.PointCount != len(value.Points) {
		return ErrTrajectoryCountInvalid
	}
	if value.Segments != nil && value.SegmentCount != len(value.Segments) {
		return ErrTrajectoryCountInvalid
	}
	if value.CoverageGaps != nil && value.CoverageGapCount != len(value.CoverageGaps) {
		return ErrTrajectoryCountInvalid
	}
	if !isValidTrajectoryQuality(value.QualityScore) {
		return ErrTrajectoryQualityInvalid
	}
	for _, point := range value.Points {
		if err := point.Validate(); err != nil {
			return err
		}
	}
	for _, segment := range value.Segments {
		if err := segment.Validate(); err != nil {
			return err
		}
	}
	for _, gap := range value.CoverageGaps {
		if err := gap.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateTimeRange(start, end time.Time, durationSeconds int64) error {
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return ErrTrajectoryTimeRangeInvalid
	}
	if durationSeconds < 0 || durationSeconds != int64(end.Sub(start)/time.Second) {
		return ErrTrajectoryDurationInvalid
	}
	return nil
}

func isValidTrajectoryCoordinates(latitude, longitude float64) bool {
	return !math.IsNaN(latitude) && !math.IsInf(latitude, 0) && latitude >= -90 && latitude <= 90 &&
		!math.IsNaN(longitude) && !math.IsInf(longitude, 0) && longitude >= -180 && longitude <= 180
}

func isValidTrajectoryQuality(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}
