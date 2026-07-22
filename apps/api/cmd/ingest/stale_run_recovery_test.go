package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRecoverStaleIngestionRunsUsesDetachedBoundedContext(
	t *testing.T,
) {
	now := time.Date(2026, time.July, 22, 20, 0, 0, 0, time.UTC)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repository := &staleRunRecoveryRepositoryStub{
		recoveredCount: 2,
	}
	recoveredCount, err := recoverStaleIngestionRuns(
		ctx,
		repository,
		now,
		30*time.Minute,
		time.Second,
	)
	if err != nil {
		t.Fatalf("recover stale ingestion runs: %v", err)
	}
	if recoveredCount != 2 {
		t.Fatalf("recovered count = %d, want 2", recoveredCount)
	}
	if repository.contextErr != nil {
		t.Fatalf("recovery context was cancelled: %v", repository.contextErr)
	}
	if !repository.staleBefore.Equal(now.Add(-30 * time.Minute)) {
		t.Fatalf(
			"stale before = %s, want %s",
			repository.staleBefore,
			now.Add(-30*time.Minute),
		)
	}
	if !repository.recoveredAt.Equal(now) {
		t.Fatalf("recovered at = %s, want %s", repository.recoveredAt, now)
	}
	if repository.errorMessage != staleIngestionRunRecoveryMessage {
		t.Fatalf("recovery message = %q", repository.errorMessage)
	}
}

func TestRecoverStaleIngestionRunsValidatesConfiguration(
	t *testing.T,
) {
	now := time.Date(2026, time.July, 22, 20, 0, 0, 0, time.UTC)
	repository := &staleRunRecoveryRepositoryStub{}

	testCases := []struct {
		name       string
		ctx        context.Context
		repository staleRunRecoveryRepository
		now        time.Time
		staleAfter time.Duration
		timeout    time.Duration
		expected   error
	}{
		{
			name:       "nil context",
			repository: repository,
			now:        now,
			staleAfter: time.Minute,
			timeout:    time.Second,
			expected:   errStaleRunRecoveryContextRequired,
		},
		{
			name:       "nil repository",
			ctx:        context.Background(),
			now:        now,
			staleAfter: time.Minute,
			timeout:    time.Second,
			expected:   errStaleRunRecoveryRepositoryRequired,
		},
		{
			name:       "zero current time",
			ctx:        context.Background(),
			repository: repository,
			staleAfter: time.Minute,
			timeout:    time.Second,
			expected:   errStaleRunRecoveryTimeRequired,
		},
		{
			name:       "invalid stale threshold",
			ctx:        context.Background(),
			repository: repository,
			now:        now,
			timeout:    time.Second,
			expected:   errStaleRunRecoveryThresholdInvalid,
		},
		{
			name:       "invalid timeout",
			ctx:        context.Background(),
			repository: repository,
			now:        now,
			staleAfter: time.Minute,
			expected:   errStaleRunRecoveryTimeoutInvalid,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := recoverStaleIngestionRuns(
				testCase.ctx,
				testCase.repository,
				testCase.now,
				testCase.staleAfter,
				testCase.timeout,
			)
			if !errors.Is(err, testCase.expected) {
				t.Fatalf("expected %v, got %v", testCase.expected, err)
			}
		})
	}
}

type staleRunRecoveryRepositoryStub struct {
	recoveredCount int64
	err            error

	contextErr   error
	staleBefore  time.Time
	recoveredAt  time.Time
	errorMessage string
}

func (repository *staleRunRecoveryRepositoryStub) RecoverStaleRunning(
	ctx context.Context,
	staleBefore time.Time,
	recoveredAt time.Time,
	errorMessage string,
) (int64, error) {
	repository.contextErr = ctx.Err()
	repository.staleBefore = staleBefore
	repository.recoveredAt = recoveredAt
	repository.errorMessage = errorMessage
	return repository.recoveredCount, repository.err
}
