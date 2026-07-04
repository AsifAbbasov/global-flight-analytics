package ingestionrun

import "time"

type Status string

const (
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
	StatusPartial Status = "partial"
)

type Run struct {
	ID              string
	SourceName      string
	RegionID        string
	StartedAt       time.Time
	FinishedAt      *time.Time
	Status          Status
	RecordsReceived int
	RecordsInserted int
	RecordsUpdated  int
	ErrorMessage    string
	CreatedAt       time.Time
}
