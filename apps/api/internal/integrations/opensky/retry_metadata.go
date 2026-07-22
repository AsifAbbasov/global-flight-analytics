package opensky

import (
	"fmt"
	"time"
)

type PollingTooSoonError struct {
	RetryAt time.Time
}

func (
	err *PollingTooSoonError,
) Error() string {
	if err == nil ||
		err.RetryAt.IsZero() {
		return ErrPollingTooSoon.Error()
	}

	return fmt.Sprintf(
		"%s: retry_at=%s",
		ErrPollingTooSoon,
		err.RetryAt.UTC().Format(
			time.RFC3339Nano,
		),
	)
}

func (
	err *PollingTooSoonError,
) Unwrap() error {
	return ErrPollingTooSoon
}

func (
	err *PollingTooSoonError,
) RetryAtTime() time.Time {
	if err == nil {
		return time.Time{}
	}

	return err.RetryAt.UTC()
}

func (
	err *PollingTooSoonError,
) ExternalRequestAttempted() bool {
	return false
}
