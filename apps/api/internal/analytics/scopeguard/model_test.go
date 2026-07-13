package scopeguard

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestDecisionHasReason(t *testing.T) {
	decision := Decision{
		Reasons: []trajectoryeligibility.ReasonCode{
			trajectoryeligibility.ReasonLowQualityScore,
		},
	}

	if !decision.HasReason(trajectoryeligibility.ReasonLowQualityScore) {
		t.Fatal("expected decision to contain low quality reason")
	}
	if decision.HasReason(trajectoryeligibility.ReasonMissingIdentity) {
		t.Fatal("expected decision not to contain missing identity reason")
	}
}

func TestDeniedErrorUnwrapsAndFormatsDeterministically(t *testing.T) {
	evaluatedAt := time.Date(2026, time.July, 13, 15, 0, 0, 0, time.UTC)
	err := &DeniedError{
		Capability: trajectoryeligibility.CapabilityRouteInference,
		Reasons: []trajectoryeligibility.ReasonCode{
			trajectoryeligibility.ReasonLowQualityScore,
			trajectoryeligibility.ReasonMissingIdentity,
		},
		IdentityKey: "flight-identity-test",
		ICAO24:      "ABC123",
		EvaluatedAt: evaluatedAt,
	}

	if !errors.Is(err, ErrDenied) {
		t.Fatal("expected denied error to unwrap to ErrDenied")
	}

	message := err.Error()
	for _, expected := range []string{
		"capability=route_inference",
		`identity_key="flight-identity-test"`,
		`icao24="ABC123"`,
		"reasons=low_quality_score,missing_identity",
		"evaluated_at=2026-07-13T15:00:00Z",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected error to contain %q, got %q", expected, message)
		}
	}
}

func TestFilterResultCounts(t *testing.T) {
	result := FilterResult{
		Allowed: make([]trajectory.FlightTrajectory, 2),
		Denied:  make([]DeniedTrajectory, 3),
	}

	if result.AllowedCount() != 2 {
		t.Fatalf("expected 2 allowed items, got %d", result.AllowedCount())
	}
	if result.DeniedCount() != 3 {
		t.Fatalf("expected 3 denied items, got %d", result.DeniedCount())
	}
}
