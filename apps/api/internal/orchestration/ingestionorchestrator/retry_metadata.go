package ingestionorchestrator

import "time"

func (
	err *AccessDeniedError,
) RetryAtTime() time.Time {
	if err == nil {
		return time.Time{}
	}

	return err.RetryAt.UTC()
}

func (
	err *AccessDeniedError,
) ExternalRequestAttempted() bool {
	return false
}
