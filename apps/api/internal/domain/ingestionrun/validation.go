package ingestionrun

import (
	"errors"
	"strings"
)

var (
	ErrIngestionSourceRequired    = errors.New("ingestion source name is required")
	ErrIngestionStartedAtRequired = errors.New("ingestion start timestamp is required")
	ErrIngestionStatusInvalid     = errors.New("ingestion status is invalid")
	ErrIngestionCountersInvalid   = errors.New("ingestion counters are invalid")
	ErrIngestionFinishedAtInvalid = errors.New("ingestion finish timestamp is invalid")
)

func (value Run) Validate() error {
	if strings.TrimSpace(value.SourceName) == "" {
		return ErrIngestionSourceRequired
	}
	if value.StartedAt.IsZero() {
		return ErrIngestionStartedAtRequired
	}
	if !isKnownIngestionRunStatus(value.Status) {
		return ErrIngestionStatusInvalid
	}
	if value.RecordsReceived < 0 || value.RecordsInserted < 0 || value.RecordsUpdated < 0 ||
		value.RecordsInserted > value.RecordsReceived ||
		value.RecordsUpdated > value.RecordsReceived-value.RecordsInserted {
		return ErrIngestionCountersInvalid
	}
	if value.Status == StatusRunning {
		if value.FinishedAt != nil {
			return ErrIngestionFinishedAtInvalid
		}
		return nil
	}
	if value.FinishedAt == nil || value.FinishedAt.Before(value.StartedAt) {
		return ErrIngestionFinishedAtInvalid
	}
	return nil
}

func isKnownIngestionRunStatus(value Status) bool {
	switch value {
	case StatusRunning, StatusSuccess, StatusFailed, StatusPartial:
		return true
	default:
		return false
	}
}
