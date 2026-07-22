package providerfallback

import "time"

func (
	err *NoProviderAvailableError,
) RetryAtTime() time.Time {
	if err == nil {
		return time.Time{}
	}

	return err.Decision.RetryAt.UTC()
}

func (
	err *NoProviderAvailableError,
) ExternalRequestAttempted() bool {
	if err == nil {
		return false
	}

	if len(err.Decision.Attempts) == 0 {
		return true
	}

	for _, attempt := range err.Decision.Attempts {
		if attempt.RequestAttempted {
			return true
		}
	}

	return false
}
