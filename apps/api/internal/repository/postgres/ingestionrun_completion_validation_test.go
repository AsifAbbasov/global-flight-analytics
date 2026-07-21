package postgres

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
)

func TestValidateIngestionRunCompletionAcceptsConsistentTerminalEvidence(
	t *testing.T,
) {
	t.Parallel()

	tests := []struct {
		name         string
		status       ingestionrun.Status
		received     int
		inserted     int
		updated      int
		errorMessage string
		wantMessage  string
	}{
		{
			name:     "success without error",
			status:   ingestionrun.StatusSuccess,
			received: 10,
			inserted: 7,
			updated:  3,
		},
		{
			name:         "failed with normalized error",
			status:       ingestionrun.StatusFailed,
			received:     4,
			inserted:     1,
			updated:      0,
			errorMessage: "  provider unavailable  ",
			wantMessage:  "provider unavailable",
		},
		{
			name:         "partial with explanation",
			status:       ingestionrun.StatusPartial,
			received:     8,
			inserted:     5,
			updated:      1,
			errorMessage: "two records rejected",
			wantMessage:  "two records rejected",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			message, err := validateIngestionRunCompletion(
				test.status,
				test.received,
				test.inserted,
				test.updated,
				test.errorMessage,
			)
			if err != nil {
				t.Fatalf("validate completion: %v", err)
			}
			if message != test.wantMessage {
				t.Fatalf("message = %q, want %q", message, test.wantMessage)
			}
		})
	}
}

func TestValidateIngestionRunCompletionRejectsImpossibleCounts(t *testing.T) {
	t.Parallel()

	for _, counts := range [][3]int{
		{-1, 0, 0},
		{1, -1, 0},
		{1, 0, -1},
		{1, 2, 0},
		{2, 1, 2},
	} {
		_, err := validateIngestionRunCompletion(
			ingestionrun.StatusSuccess,
			counts[0],
			counts[1],
			counts[2],
			"",
		)
		if !errors.Is(err, ErrIngestionRunCountsInvalid) {
			t.Fatalf("counts %v returned %v", counts, err)
		}
	}
}

func TestValidateIngestionRunCompletionRejectsStatusErrorMismatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status       ingestionrun.Status
		errorMessage string
	}{
		{status: ingestionrun.StatusSuccess, errorMessage: "unexpected error"},
		{status: ingestionrun.StatusFailed, errorMessage: ""},
		{status: ingestionrun.StatusPartial, errorMessage: "   "},
		{status: ingestionrun.StatusRunning, errorMessage: ""},
	}

	for _, test := range tests {
		_, err := validateIngestionRunCompletion(
			test.status,
			1,
			1,
			0,
			test.errorMessage,
		)
		if !errors.Is(err, ErrIngestionRunErrorMessageInvalid) {
			t.Fatalf("status %q returned %v", test.status, err)
		}
	}
}
