package reconciliation

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

const maximumLastErrorLength = 4000

type DerivationType string

const (
	DerivationTypeFlightStateQuality DerivationType = "flight_state_quality"
	DerivationTypeTrajectory         DerivationType = "trajectory"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

var (
	ErrICAO24Required         = errors.New("reconciliation task icao24 is required")
	ErrDerivationTypeInvalid  = errors.New("reconciliation task derivation type is invalid")
	ErrObservedRangeRequired  = errors.New("reconciliation task observed range is required")
	ErrObservedRangeInvalid   = errors.New("reconciliation task observed range is invalid")
	ErrTaskIDRequired         = errors.New("reconciliation task id is required")
	ErrAttemptCountInvalid    = errors.New("reconciliation attempt count must be greater than zero")
	ErrNextAttemptAtRequired  = errors.New("reconciliation next attempt time is required")
	ErrStaleBeforeRequired    = errors.New("reconciliation stale-before time is required")
	ErrNoTaskAvailable        = errors.New("no reconciliation task is available")
	ErrTaskTransitionRejected = errors.New("reconciliation task transition was rejected")
	ErrIngestionRunIDInvalid  = errors.New("reconciliation ingestion run identifier is invalid")
)

type PendingDerivation struct {
	IngestionRunID string
	ICAO24         string
	DerivationType DerivationType
	ObservedFrom   time.Time
	ObservedTo     time.Time
	LastError      string
}

type Task struct {
	ID                   string
	DeduplicationKey     string
	IngestionRunID       string
	ICAO24               string
	DerivationType       DerivationType
	Status               TaskStatus
	ObservedFrom         time.Time
	ObservedTo           time.Time
	AttemptCount         int
	SignalVersion        int64
	ClaimedSignalVersion int64
	LastError            string
	NextAttemptAt        time.Time
	ProcessingStartedAt  *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
	CompletedAt          *time.Time
}

func (task PendingDerivation) Normalize() PendingDerivation {
	return PendingDerivation{
		IngestionRunID: NormalizeTaskID(
			task.IngestionRunID,
		),
		ICAO24: strings.ToLower(
			strings.TrimSpace(task.ICAO24),
		),
		DerivationType: DerivationType(
			strings.TrimSpace(
				string(task.DerivationType),
			),
		),
		ObservedFrom: normalizeTime(task.ObservedFrom),
		ObservedTo:   normalizeTime(task.ObservedTo),
		LastError:    NormalizeLastError(task.LastError),
	}
}

func (task PendingDerivation) Validate() error {
	normalized := task.Normalize()

	if normalized.ICAO24 == "" {
		return ErrICAO24Required
	}
	if strings.Contains(normalized.IngestionRunID, "|") ||
		!utf8.ValidString(normalized.IngestionRunID) {
		return ErrIngestionRunIDInvalid
	}

	if !IsKnownDerivationType(normalized.DerivationType) {
		return ErrDerivationTypeInvalid
	}

	if normalized.ObservedFrom.IsZero() || normalized.ObservedTo.IsZero() {
		return ErrObservedRangeRequired
	}

	if normalized.ObservedFrom.After(normalized.ObservedTo) {
		return ErrObservedRangeInvalid
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
		DerivationTypeTrajectory:
		return true
	default:
		return false
	}
}

func NormalizeTaskID(
	value string,
) string {
	return strings.TrimSpace(value)
}

func NormalizeLastError(
	value string,
) string {
	normalized := strings.TrimSpace(value)

	normalized = strings.ToValidUTF8(normalized, "\uFFFD")
	if len(normalized) <= maximumLastErrorLength {
		return normalized
	}

	cut := maximumLastErrorLength
	for cut > 0 && !utf8.RuneStart(normalized[cut]) {
		cut--
	}

	return normalized[:cut]
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
