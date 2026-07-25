package flightfeatures

import "testing"

func TestFlightFeaturesCloneCopiesValidationReportIssues(
	t *testing.T,
) {
	features := FlightFeatures{
		ValidationReport: ValidationReport{
			AuditState: ValidationAuditStateComplete,
			Issues: []ValidationIssue{
				{
					Code: "original",
				},
			},
		},
	}

	cloned := features.Clone()
	cloned.ValidationReport.Issues[0].Code = "changed"

	if features.ValidationReport.Issues[0].Code !=
		"original" {
		t.Fatal(
			"FlightFeatures.Clone() shared validation issues",
		)
	}
}
