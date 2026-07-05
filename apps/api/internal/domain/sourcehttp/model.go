package sourcehttp

import (
	"strings"
	"time"
)

type Validator struct {
	SourceName   string
	ResourceURL  string
	ETag         string
	LastModified string

	// ObservedAt is the time when this validator state was observed.
	//
	// It is not the time of the latest HTTP check.
	// Repeated HTTP 304 Not Modified responses with unchanged validators
	// must not advance this value only to record that another check happened.
	//
	// HTTP check telemetry belongs to the provider retrieval lifecycle,
	// where CheckedAt represents the actual time of a source check.
	ObservedAt time.Time
}

func (validator Validator) HasValidators() bool {
	return strings.TrimSpace(validator.ETag) != "" ||
		strings.TrimSpace(validator.LastModified) != ""
}
