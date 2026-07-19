package dto

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/transponderalert"
)

func TestToTransponderEvidenceResponsePreservesSafetySemantics(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		19,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	result := transponderalert.LatestEvidence{
		Evidence: transponderalert.Evidence{
			SchemaVersion: transponderalert.SchemaVersion,
			Fingerprint:   "sha256:test",
			ICAO24:        "4A001A",
			Callsign:      "AHY101",
			SquawkCode:    "7700",
			Kind: transponderalert.
				KindGeneralEmergencyCode,
			Label: "Observed general emergency transponder code",
			Strength: transponderalert.
				StrengthSingleObservation,
			FirstObservedAt:  now.Add(-30 * time.Second),
			LastObservedAt:   now.Add(-30 * time.Second),
			AsOfTime:         now,
			ObservationCount: 1,
			SourceNames: []string{
				"opensky",
			},
			MaximumClaimStrength: "observed_transponder_code_only",
			Limitations: []string{
				"research only",
			},
		},
		FreshnessStatus: transponderalert.FreshnessRecent,
		Age:             30 * time.Second,
		MaximumFreshAge: 5 * time.Minute,
		Confidence: transponderalert.Confidence{
			Level: transponderalert.ConfidenceLimited,
			Reasons: []string{
				"single observation",
			},
		},
		EvidenceOnly:       true,
		ConfirmedEmergency: false,
		OperationalAlert:   false,
	}

	response := ToTransponderEvidenceResponse(result)

	if !response.EvidenceOnly {
		t.Fatal("evidence-only flag is false")
	}
	if response.ConfirmedEmergency {
		t.Fatal("emergency was incorrectly confirmed")
	}
	if response.OperationalAlert {
		t.Fatal("operational alert was incorrectly produced")
	}
	if response.ObservedTransponderCode != "7700" {
		t.Fatalf(
			"observed code = %q",
			response.ObservedTransponderCode,
		)
	}
	if response.Freshness.AgeSeconds != 30 {
		t.Fatalf(
			"age seconds = %d",
			response.Freshness.AgeSeconds,
		)
	}
	if response.Freshness.MaximumFreshAgeSeconds != 300 {
		t.Fatalf(
			"maximum fresh age seconds = %d",
			response.Freshness.MaximumFreshAgeSeconds,
		)
	}
	if response.MaximumClaimStrength !=
		"observed_transponder_code_only" {
		t.Fatalf(
			"claim strength = %q",
			response.MaximumClaimStrength,
		)
	}
}
