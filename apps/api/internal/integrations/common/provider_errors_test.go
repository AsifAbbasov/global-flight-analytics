package common

import (
	"errors"
	"net/http"
	"testing"
)

func TestClassifyProviderStatus(
	t *testing.T,
) {
	tests := []struct {
		name       string
		status     int
		expected   error
		expectNone bool
	}{
		{
			name:       "success status has no provider error",
			status:     http.StatusOK,
			expectNone: true,
		},
		{
			name:       "redirect status has no provider error",
			status:     http.StatusMultipleChoices,
			expectNone: true,
		},
		{
			name:     "rate limited status",
			status:   http.StatusTooManyRequests,
			expected: ErrProviderRateLimited,
		},
		{
			name:     "unauthorized status",
			status:   http.StatusUnauthorized,
			expected: ErrProviderUnauthorized,
		},
		{
			name:     "forbidden status",
			status:   http.StatusForbidden,
			expected: ErrProviderUnauthorized,
		},
		{
			name:     "client status",
			status:   http.StatusNotFound,
			expected: ErrProviderClient,
		},
		{
			name:     "server status",
			status:   http.StatusBadGateway,
			expected: ErrProviderServer,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				actual := ClassifyProviderStatus(
					test.status,
				)

				if test.expectNone {
					if actual != nil {
						t.Fatalf(
							"expected no provider error, got %v",
							actual,
						)
					}

					return
				}

				if !errors.Is(
					actual,
					test.expected,
				) {
					t.Fatalf(
						"expected error %v, got %v",
						test.expected,
						actual,
					)
				}
			},
		)
	}
}

func TestProviderStatusErrorWrapsClassifiedStatus(
	t *testing.T,
) {
	err := ProviderStatusError(
		http.StatusTooManyRequests,
	)
	if !errors.Is(
		err,
		ErrProviderRateLimited,
	) {
		t.Fatalf(
			"expected rate limited provider error, got %v",
			err,
		)
	}
}

func TestProviderStatusErrorReturnsNilForSuccessfulStatus(
	t *testing.T,
) {
	err := ProviderStatusError(
		http.StatusNoContent,
	)
	if err != nil {
		t.Fatalf(
			"expected no provider error, got %v",
			err,
		)
	}
}
