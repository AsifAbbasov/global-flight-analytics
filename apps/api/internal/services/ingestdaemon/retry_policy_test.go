package ingestdaemon

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type retryAtTestError struct {
	retryAt time.Time
}

func (err *retryAtTestError) Error() string {
	return "retry later"
}

func (err *retryAtTestError) RetryAtTime() time.Time {
	return err.retryAt
}

func TestExponentialBackoffIsBounded(
	t *testing.T,
) {
	testCases := []struct {
		failures int
		want     time.Duration
	}{
		{failures: 1, want: 10 * time.Second},
		{failures: 2, want: 20 * time.Second},
		{failures: 3, want: 40 * time.Second},
		{failures: 4, want: time.Minute},
		{failures: 20, want: time.Minute},
	}

	for _, testCase := range testCases {
		t.Run(
			fmt.Sprintf("failures-%d", testCase.failures),
			func(t *testing.T) {
				got := exponentialBackoff(
					10*time.Second,
					time.Minute,
					testCase.failures,
				)
				if got != testCase.want {
					t.Fatalf(
						"backoff = %s, want %s",
						got,
						testCase.want,
					)
				}
			},
		)
	}
}

func TestNextDelayHonorsProviderRetryAt(
	t *testing.T,
) {
	daemon := &Daemon{
		interval:          10 * time.Second,
		maxFailureBackoff: time.Minute,
	}
	finishedAt := time.Date(
		2026,
		time.July,
		23,
		1,
		0,
		0,
		0,
		time.UTC,
	)
	retryAt := finishedAt.Add(
		90 * time.Second,
	)

	delay := daemon.nextDelay(
		finishedAt,
		errors.New("provider unavailable"),
		2,
		retryAt,
	)
	if delay != 90*time.Second {
		t.Fatalf(
			"next delay = %s, want 90s",
			delay,
		)
	}
}

func TestRetryAtIsDiscoveredThroughWrappedErrors(
	t *testing.T,
) {
	retryAt := time.Date(
		2026,
		time.July,
		23,
		1,
		1,
		0,
		0,
		time.UTC,
	)
	err := fmt.Errorf(
		"wrapped: %w",
		&retryAtTestError{
			retryAt: retryAt,
		},
	)

	got := retryAtFromError(err)
	if !got.Equal(retryAt) {
		t.Fatalf(
			"retry at = %s, want %s",
			got,
			retryAt,
		)
	}
}
