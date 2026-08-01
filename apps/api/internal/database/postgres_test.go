package database

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewPostgresPoolRejectsNonPositiveConnectTimeout(
	t *testing.T,
) {
	tests := []struct {
		name           string
		connectTimeout time.Duration
	}{
		{
			name:           "zero timeout",
			connectTimeout: 0,
		},
		{
			name:           "negative timeout",
			connectTimeout: -1 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				pool, err := NewPostgresPool(
					"postgres://example.invalid/database",
					test.connectTimeout,
				)

				if err == nil {
					if pool != nil {
						pool.Close()
					}

					t.Fatal(
						"expected postgres pool initialization error, got nil",
					)
				}

				if pool != nil {
					pool.Close()

					t.Fatal(
						"expected nil postgres pool on invalid connect timeout",
					)
				}

				if !errors.Is(
					err,
					errPostgresConnectTimeoutInvalid,
				) {
					t.Fatalf(
						"expected invalid timeout classification, got %v",
						err,
					)
				}
			},
		)
	}
}

func TestNewPostgresPoolContextRequiresContext(
	t *testing.T,
) {
	pool, err := NewPostgresPoolContext(
		nil,
		"postgres://example.invalid/database",
		time.Second,
	)

	if pool != nil {
		pool.Close()
		t.Fatal(
			"expected nil postgres pool when context is missing",
		)
	}
	if !errors.Is(
		err,
		errPostgresContextRequired,
	) {
		t.Fatalf(
			"expected context-required error, got %v",
			err,
		)
	}
}

func TestNewPostgresPoolContextPreservesCallerCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	pool, err := NewPostgresPoolContext(
		ctx,
		"postgres://postgres:postgres@127.0.0.1:5432/global_flight_analytics?sslmode=disable",
		time.Second,
	)

	if pool != nil {
		pool.Close()
		t.Fatal(
			"expected nil postgres pool after caller cancellation",
		)
	}
	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected caller cancellation, got %v",
			err,
		)
	}
}
