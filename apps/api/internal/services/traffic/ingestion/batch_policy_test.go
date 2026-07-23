package ingestion

import (
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerbatch"
)

func TestProviderBatchPartialFailure(t *testing.T) {
	partial, message := providerBatchPartialFailure(
		providerbatch.Evidence{
			Received:          3,
			Accepted:          2,
			RejectedMalformed: 1,
		},
	)
	if !partial {
		t.Fatal("expected partial provider batch")
	}
	if !strings.Contains(message, "rejected=1") {
		t.Fatalf("unexpected partial message: %q", message)
	}
}

func TestProviderBatchCompleteSuccess(t *testing.T) {
	partial, message := providerBatchPartialFailure(
		providerbatch.AcceptedOnly(2),
	)
	if partial || message != "" {
		t.Fatalf(
			"partial=%t message=%q, want complete success",
			partial,
			message,
		)
	}
}
