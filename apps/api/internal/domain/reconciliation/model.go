package reconciliation

import (
	"errors"
	"strings"
	"time"
)

const maximumLastErrorLength = 4000

type DerivationType string

const (
	DerivationTypeFlightStateQuality DerivationType = "flight_state_quality"
	DerivationTypeTrajectory         DerivationType = "trajectory"
	DerivationTypeCoverageGap        DerivationType = "coverage_gap"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

var (
	ErrICAO24Required        = errors.New("reconciliation task icao24 is required")
	ErrDerivationTypeInvalid = errors.New("reconciliation task derivation type is invalid")
)

type PendingDerivation struct {
	IngestionRunID string
	ICAO24         string
	DerivationType DerivationType
	ObservedFrom   time.Time
	ObservedTo     time.Time
	LastError      string
}

func (task PendingDerivation) Normalize() PendingDerivation {
	normalized := PendingDerivation{
		IngestionRunID: strings.TrimSpace(task.IngestionRunID),
		ICAO24: strings.ToLower(
			strings.TrimSpace(task.ICAO24),
		),
		DerivationType: task.DerivationType,
		ObservedFrom:   normalizeTime(task.ObservedFrom),
		ObservedTo:     normalizeTime(task.ObservedTo),
		LastError:      strings.TrimSpace(task.LastError),
	}

	if len(normalized.LastError) > maximumLastErrorLength {
		normalized.LastError = normalized.LastError[:maximumLastErrorLength]
	}

	return normalized
}

func (task PendingDerivation) Validate() error {
	normalized := task.Normalize()

	if normalized.ICAO24 == "" {
		return ErrICAO24Required
	}

	if !IsKnownDerivationType(normalized.DerivationType) {
		return ErrDerivationTypeInvalid
	}

	return nil
}

func (task PendingDerivation) DeduplicationKey() string {
	normalized := task.Normalize()

	return strings.Join(
		[]string{
			string(normalized.DerivationType),
			normalized.ICAO24,
			normalized.IngestionRunID,
			formatTimeKey(normalized.ObservedFrom),
			formatTimeKey(normalized.ObservedTo),
		},
		"|",
	)
}

func IsKnownDerivationType(
	value DerivationType,
) bool {
	switch value {
	case DerivationTypeFlightStateQuality,
		DerivationTypeTrajectory,
		DerivationTypeCoverageGap:
		return true
	default:
		return false
	}
}

func normalizeTime(
	value time.Time,
) time.Time {
	if value.IsZero() {
		return time.Time{}
	}

	return value.UTC()
}

func formatTimeKey(
	value time.Time,
) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(time.RFC3339Nano)
}
