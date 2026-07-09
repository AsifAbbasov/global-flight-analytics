package database

import (
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

				expectedError := "postgres connect timeout must be greater than zero"

				if err.Error() != expectedError {
					t.Fatalf(
						"expected error %q, got %q",
						expectedError,
						err.Error(),
					)
				}
			},
		)
	}
}
