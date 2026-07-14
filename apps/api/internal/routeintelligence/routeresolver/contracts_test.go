package routeresolver

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestResolutionCloneDoesNotShareMutableState(
	t *testing.T,
) {
	resolution := Resolution{
		Result: routecontract.Result{
			Origin: &routecontract.EndpointInference{
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
			},
			Limitations: []routecontract.Limitation{
				{
					Code: "original",
				},
			},
			Provenance: routecontract.Provenance{
				SourceNames: []string{
					"original",
				},
			},
		},
		Validation: routecontract.ValidationReport{
			Issues: []routecontract.ValidationIssue{
				{
					Code: "original",
				},
			},
		},
	}

	cloned := resolution.Clone()
	cloned.Result.Origin.Confidence.Reasons[0].Code =
		"changed"
	cloned.Result.Origin.Evidence[0].Attributes[0].Key =
		"changed"
	cloned.Result.Limitations[0].Code = "changed"
	cloned.Result.Provenance.SourceNames[0] =
		"changed"
	cloned.Validation.Issues[0].Code = "changed"

	if resolution.Result.Origin.Confidence.
		Reasons[0].Code != "original" ||
		resolution.Result.Origin.Evidence[0].
			Attributes[0].Key != "original" ||
		resolution.Result.Limitations[0].Code !=
			"original" ||
		resolution.Result.Provenance.
			SourceNames[0] != "original" ||
		resolution.Validation.Issues[0].Code !=
			"original" {
		t.Fatal(
			"Resolution.Clone() shared mutable state",
		)
	}
}

func TestVersionConstantsRemainStable(
	t *testing.T,
) {
	if Version != "route-resolver-v1" {
		t.Fatalf("Version = %q", Version)
	}
	if DefaultPartialConfidenceFactor != 0.50 {
		t.Fatalf(
			"DefaultPartialConfidenceFactor = %v",
			DefaultPartialConfidenceFactor,
		)
	}
	if DefaultSameAirportConfidenceFactor != 0.75 {
		t.Fatalf(
			"DefaultSameAirportConfidenceFactor = %v",
			DefaultSameAirportConfidenceFactor,
		)
	}
}
