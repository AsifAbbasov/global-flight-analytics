package historicalread

import (
	"encoding/hex"
	"math"
	"strings"
	"time"
)

func validateFlightRecord(
	item FlightRecord,
	query Query,
	index int,
) error {
	if strings.TrimSpace(item.ID) == "" ||
		strings.TrimSpace(item.Status) == "" {
		return invalidRecord(DatasetFlights, index, "id and status are required")
	}
	if item.FirstSeenAt.IsZero() || item.LastSeenAt.IsZero() || item.UpdatedAt.IsZero() {
		return invalidRecord(DatasetFlights, index, "timestamps are required")
	}
	if item.LastSeenAt.Before(item.FirstSeenAt) {
		return invalidRecord(DatasetFlights, index, "last seen precedes first seen")
	}
	if item.UpdatedAt.After(query.Window.AsOfTime) {
		return invalidRecord(DatasetFlights, index, "updated time exceeds as-of time")
	}
	if !item.FirstSeenAt.Before(query.Window.EndTime) ||
		!item.LastSeenAt.After(query.Window.StartTime) {
		return invalidRecord(DatasetFlights, index, "record does not overlap the half-open window")
	}
	return nil
}

func validateTrajectoryRecord(
	item TrajectoryRecord,
	query Query,
	index int,
) error {
	if strings.TrimSpace(item.ID) == "" ||
		strings.TrimSpace(item.ICAO24) == "" ||
		strings.TrimSpace(item.SourceName) == "" {
		return invalidRecord(DatasetTrajectories, index, "id, ICAO24, and source are required")
	}
	if item.StartTime.IsZero() || item.EndTime.IsZero() || item.UpdatedAt.IsZero() {
		return invalidRecord(DatasetTrajectories, index, "timestamps are required")
	}
	if item.EndTime.Before(item.StartTime) {
		return invalidRecord(DatasetTrajectories, index, "end time precedes start time")
	}
	if item.UpdatedAt.After(query.Window.AsOfTime) {
		return invalidRecord(DatasetTrajectories, index, "updated time exceeds as-of time")
	}
	if !item.StartTime.Before(query.Window.EndTime) ||
		!item.EndTime.After(query.Window.StartTime) {
		return invalidRecord(DatasetTrajectories, index, "record does not overlap the half-open window")
	}
	if item.SegmentCount < 0 || item.PointCount < 0 || item.CoverageGapCount < 0 {
		return invalidRecord(DatasetTrajectories, index, "counts must be non-negative")
	}
	if math.IsNaN(item.QualityScore) || math.IsInf(item.QualityScore, 0) ||
		item.QualityScore < 0 || item.QualityScore > 1 {
		return invalidRecord(DatasetTrajectories, index, "quality score must be finite and within [0,1]")
	}
	return nil
}

func validateObservationRecord(
	item ObservationRecord,
	query Query,
	index int,
) error {
	if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.ICAO24) == "" ||
		strings.TrimSpace(item.SourceName) == "" {
		return invalidRecord(DatasetObservations, index, "id, ICAO24, and source are required")
	}
	if item.ObservedAt.IsZero() || item.CreatedAt.IsZero() {
		return invalidRecord(DatasetObservations, index, "timestamps are required")
	}
	if item.CreatedAt.After(query.Window.AsOfTime) {
		return invalidRecord(DatasetObservations, index, "created time exceeds as-of time")
	}
	if item.ObservedAt.Before(query.Window.StartTime) ||
		!item.ObservedAt.Before(query.Window.EndTime) {
		return invalidRecord(DatasetObservations, index, "observation lies outside the half-open window")
	}
	if invalidCoordinate(item.Latitude, -90, 90) ||
		invalidCoordinate(item.Longitude, -180, 180) {
		return invalidRecord(DatasetObservations, index, "coordinates are non-finite or out of range")
	}
	return nil
}

func validateRouteRecord(
	item RouteRecord,
	query Query,
	index int,
) error {
	if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.TrajectoryID) == "" ||
		!validSHA256Fingerprint(item.InputFingerprint) {
		return invalidRecord(DatasetRoutes, index, "id, trajectory id, and valid fingerprint are required")
	}
	if !knownPersistedRouteStatus(item.Status) ||
		!knownPersistedConfidenceLevel(item.ConfidenceLevel) {
		return invalidRecord(DatasetRoutes, index, "route status or confidence level is unsupported")
	}
	if item.EventStartTime.IsZero() || item.EventEndTime.IsZero() ||
		item.AsOfTime.IsZero() || item.StoredAt.IsZero() {
		return invalidRecord(DatasetRoutes, index, "event and evidence timestamps are required")
	}
	if !item.EventStartTime.Before(item.EventEndTime) ||
		!item.EventStartTime.Before(query.Window.EndTime) ||
		!item.EventEndTime.After(query.Window.StartTime) {
		return invalidRecord(DatasetRoutes, index, "route event does not overlap the half-open window")
	}
	if item.AsOfTime.After(query.Window.AsOfTime) || item.StoredAt.After(query.Window.AsOfTime) {
		return invalidRecord(DatasetRoutes, index, "route evidence exceeds as-of time")
	}
	if item.ValidationWarningCount < 0 || item.PayloadBytes < 0 {
		return invalidRecord(DatasetRoutes, index, "counts and payload size must be non-negative")
	}
	return nil
}

func validateMatchedCount(
	dataset string,
	matchedCount int64,
	loadedCount int,
) error {
	if matchedCount < 0 || matchedCount < int64(loadedCount) {
		return &RecordValidationError{
			Dataset: dataset,
			Index:   -1,
			Reason:  "matched count is inconsistent with loaded rows",
		}
	}
	return nil
}

func validSHA256Fingerprint(value string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+64 {
		return false
	}
	_, err := hex.DecodeString(value[len(prefix):])
	return err == nil
}

func knownPersistedRouteStatus(value string) bool {
	switch strings.TrimSpace(value) {
	case "unavailable", "partial", "complete":
		return true
	default:
		return false
	}
}

func knownPersistedConfidenceLevel(value string) bool {
	switch strings.TrimSpace(value) {
	case "none", "low", "medium", "high":
		return true
	default:
		return false
	}
}

func invalidCoordinate(value *float64, minimum float64, maximum float64) bool {
	if value == nil {
		return false
	}
	return math.IsNaN(*value) || math.IsInf(*value, 0) ||
		*value < minimum || *value > maximum
}

func invalidRecord(dataset string, index int, reason string) error {
	return &RecordValidationError{
		Dataset: dataset,
		Index:   index,
		Reason:  reason,
	}
}

func normalizeTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return value.UTC()
}
