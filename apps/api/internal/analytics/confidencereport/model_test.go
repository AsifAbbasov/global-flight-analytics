package confidencereport

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
)

func TestFactorKindsAndStandardConstructors(
	t *testing.T,
) {
	if !FactorKindEvidence.IsKnown() ||
		!FactorKindPenalty.IsKnown() {
		t.Fatal("expected known factor kinds")
	}

	if FactorKind("unknown").IsKnown() {
		t.Fatal("expected unknown factor kind rejection")
	}

	evidence := Evidence(
		FactorCodeRecentContinuity,
		0.25,
		0.80,
		"Recent continuity supports confidence.",
	)
	penalty := Penalty(
		FactorCodeFallbackSourcePenalty,
		0.10,
		0.50,
		"Fallback source usage reduces confidence.",
	)

	if evidence.Kind != FactorKindEvidence ||
		penalty.Kind != FactorKindPenalty {
		t.Fatal("expected standard factor constructors")
	}
}

func TestAnalyticalConfidenceUsesReportValues(
	t *testing.T,
) {
	report := Report{
		Score: 0.75,
		Level: analyticalresult.
			ConfidenceLevelMedium,
		Reasons: []analyticalresult.Notice{
			{
				Code:    "confidence_evidence_test",
				Message: "Test evidence supports confidence.",
			},
		},
	}

	confidence := report.AnalyticalConfidence()
	if confidence.Score != 0.75 ||
		confidence.Level !=
			analyticalresult.ConfidenceLevelMedium {
		t.Fatalf(
			"unexpected analytical confidence: %#v",
			confidence,
		)
	}
}
