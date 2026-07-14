package endpointevidence

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestResultCloneDoesNotShareMutableState(
	t *testing.T,
) {
	result := Result{
		Endpoint: &routecontract.EndpointInference{
			Confidence: routecontract.Confidence{
				Reasons: []routecontract.ConfidenceReason{
					{
						Code: "original",
					},
				},
			},
			Evidence: []routecontract.Evidence{
				{
					Attributes: []routecontract.EvidenceAttribute{
						{
							Key: "original",
						},
					},
				},
			},
			Limitations: []routecontract.Limitation{
				{
					Code: "original",
				},
			},
		},
		Limitations: []routecontract.Limitation{
			{
				Code: "original",
			},
		},
	}

	cloned := result.Clone()
	cloned.Endpoint.Confidence.Reasons[0].Code =
		"changed"
	cloned.Endpoint.Evidence[0].Attributes[0].Key =
		"changed"
	cloned.Endpoint.Limitations[0].Code =
		"changed"
	cloned.Limitations[0].Code = "changed"

	if result.Endpoint.Confidence.Reasons[0].Code !=
		"original" ||
		result.Endpoint.Evidence[0].Attributes[0].Key !=
			"original" ||
		result.Endpoint.Limitations[0].Code !=
			"original" ||
		result.Limitations[0].Code != "original" {
		t.Fatal(
			"Result.Clone() shared mutable state",
		)
	}
}

func TestVersionConstantsRemainStable(
	t *testing.T,
) {
	if Version != "route-endpoint-evidence-v1" {
		t.Fatalf("Version = %q", Version)
	}
	if DefaultMinimumSelectionScore != 0.60 {
		t.Fatalf(
			"DefaultMinimumSelectionScore = %v",
			DefaultMinimumSelectionScore,
		)
	}
	if DefaultMinimumCandidateScoreGap != 0.05 {
		t.Fatalf(
			"DefaultMinimumCandidateScoreGap = %v",
			DefaultMinimumCandidateScoreGap,
		)
	}
}
