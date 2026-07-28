package projectioncontract

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestValidateRejectsNonDivisibleHorizonGrid(t *testing.T) {
	result := validProjectionResult()
	result.Horizon.EndTime = result.Horizon.AsOfTime.Add(
		11 * time.Minute,
	)

	report := Validate(result)
	if !report.HasCode(IssueHorizonGridInvalid) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueHorizonGridInvalid)
	}
}

func TestValidateRejectsProjectionPointOffGrid(t *testing.T) {
	result := validProjectionResult()
	result.Points[0].ForecastTime = result.Horizon.AsOfTime.Add(
		6 * time.Minute,
	)

	report := Validate(result)
	if !report.HasCode(IssuePointGridInvalid) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssuePointGridInvalid)
	}
}

func TestValidateLimitedStatusRequiresExplicitEvidence(t *testing.T) {
	result := validProjectionResult()
	result.Status = ResultStatusLimited

	report := Validate(result)
	if !report.HasCode(IssueLimitedContractInvalid) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueLimitedContractInvalid)
	}

	result.Limitations = append(
		result.Limitations,
		Limitation{
			Code:    "projection_altitude_unavailable",
			Message: "Altitude evidence was unavailable.",
			Scope:   "position",
		},
	)
	if report := Validate(result); report.Status != ValidationStatusValid {
		t.Fatalf("limited result with explicit evidence is invalid: %#v", report.Issues)
	}
}

func TestValidateRejectsPositiveConfidenceWithoutReasons(t *testing.T) {
	result := validProjectionResult()
	result.Confidence.Reasons = nil

	report := Validate(result)
	if !report.HasCode(IssueConfidenceReasonRequired) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueConfidenceReasonRequired)
	}
}

func TestValidateRejectsUnboundedConfidenceContribution(t *testing.T) {
	result := validProjectionResult()
	result.Confidence.Reasons[0].Contribution = 999

	report := Validate(result)
	if !report.HasCode(IssueConfidenceReasonInvalid) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueConfidenceReasonInvalid)
	}
}

func TestValidateRejectsDuplicateConfidenceReasons(t *testing.T) {
	result := validProjectionResult()
	result.Confidence.Reasons = append(
		result.Confidence.Reasons,
		result.Confidence.Reasons[0],
	)

	report := Validate(result)
	if !report.HasCode(IssueConfidenceReasonDuplicate) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueConfidenceReasonDuplicate)
	}
}

func TestValidateRejectsOverallConfidenceAboveEvidence(t *testing.T) {
	result := validProjectionResult()
	result.Confidence.Score = 0.9
	result.Confidence.Level = ConfidenceLevelHigh
	result.Confidence.Reasons[0].Contribution = 0.9

	report := Validate(result)
	if !report.HasCode(IssueConfidenceExceedsEvidence) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueConfidenceExceedsEvidence)
	}
}

func TestValidateDoesNotInventGlobalConfidenceScoreBands(t *testing.T) {
	result := validProjectionResult()
	for index := range result.Points {
		result.Points[index].Confidence.Score = 0.01
		result.Points[index].Confidence.Level = ConfidenceLevelHigh
		result.Points[index].Confidence.Reasons[0].Contribution = 0.01
	}
	result.Arrival.Confidence.Score = 0.01
	result.Arrival.Confidence.Level = ConfidenceLevelHigh
	result.Arrival.Confidence.Reasons[0].Contribution = 0.01
	result.Confidence.Score = 0.01
	result.Confidence.Level = ConfidenceLevelHigh
	result.Confidence.Reasons[0].Contribution = 0.01

	if report := Validate(result); report.Status != ValidationStatusValid {
		t.Fatalf("producer-owned confidence bands were rejected: %#v", report.Issues)
	}
}

func TestValidateRejectsMalformedFingerprint(t *testing.T) {
	result := validProjectionResult()
	result.Provenance.InputFingerprint = "sha256:projection-test"

	report := Validate(result)
	if !report.HasCode(IssueFingerprintInvalid) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueFingerprintInvalid)
	}
}

func TestValidateRejectsInvalidICAO24(t *testing.T) {
	result := validProjectionResult()
	result.ICAO24 = "4K1234"

	report := Validate(result)
	if !report.HasCode(IssueICAO24Invalid) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueICAO24Invalid)
	}
}

func TestValidateRejectsInvalidAirportICAO(t *testing.T) {
	result := validProjectionResult()
	result.Arrival.AirportICAOCode = "U1BB"

	report := Validate(result)
	if !report.HasCode(IssueArrivalAirportInvalid) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueArrivalAirportInvalid)
	}
}

func TestValidateRejectsIncompleteObservedInputEvidence(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*InputReference)
		code   string
	}{
		{
			name: "missing source",
			mutate: func(input *InputReference) {
				input.SourceName = ""
			},
			code: IssueInputInvalid,
		},
		{
			name: "missing observation",
			mutate: func(input *InputReference) {
				input.ObservedAt = time.Time{}
			},
			code: IssueInputInvalid,
		},
		{
			name: "retrieval before observation",
			mutate: func(input *InputReference) {
				input.RetrievedAt = input.ObservedAt.Add(-time.Second)
			},
			code: IssueInputChronologyInvalid,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validProjectionResult()
			test.mutate(&result.Provenance.Inputs[0])
			report := Validate(result)
			if !report.HasCode(test.code) {
				t.Fatalf("issues = %#v, want %q", report.Issues, test.code)
			}
		})
	}
}

func TestValidateRejectsDuplicateProvenanceInputs(t *testing.T) {
	result := validProjectionResult()
	duplicate := result.Provenance.Inputs[0]
	duplicate.SourceName = "another_source"
	result.Provenance.Inputs = append(result.Provenance.Inputs, duplicate)

	report := Validate(result)
	if !report.HasCode(IssueInputDuplicate) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueInputDuplicate)
	}
}

func TestValidateRejectsDuplicateLimitationsAndExplanations(t *testing.T) {
	result := validProjectionResult()
	result.Limitations = append(
		result.Limitations,
		result.Limitations[0],
	)
	result.Explanations = append(
		result.Explanations,
		result.Explanations[0],
	)

	report := Validate(result)
	if !report.HasCode(IssueLimitationDuplicate) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueLimitationDuplicate)
	}
	if !report.HasCode(IssueExplanationDuplicate) {
		t.Fatalf("issues = %#v, want %q", report.Issues, IssueExplanationDuplicate)
	}
}

func TestResultValidateReturnsTypedError(t *testing.T) {
	result := validProjectionResult()
	result.Provenance.InputFingerprint = "invalid"

	err := result.Validate()
	if !errors.Is(err, ErrResultInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrResultInvalid)
	}

	var validationError *ResultValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("error type = %T, want *ResultValidationError", err)
	}
	if len(validationError.Issues) == 0 {
		t.Fatal("typed validation error did not preserve issues")
	}

	cloned := validationError.Clone()
	cloned.Issues[0].Code = "changed"
	if validationError.Issues[0].Code == "changed" {
		t.Fatal("ResultValidationError.Clone() shared issues")
	}
}

func TestValidationReportCarriesVersion(t *testing.T) {
	report := Validate(validProjectionResult())
	if report.Version != ValidationVersion {
		t.Fatalf("version = %q, want %q", report.Version, ValidationVersion)
	}
	if !strings.HasSuffix(report.Version, "-v2") {
		t.Fatalf("validation version = %q, want v2", report.Version)
	}
}

func TestValidateAcceptsProducerQualifiedInputNames(t *testing.T) {
	result := validProjectionResult()
	result.Provenance.Inputs[1].Name =
		"historical_neighbor:trajectory-123"

	if report := Validate(result); report.Status != ValidationStatusValid {
		t.Fatalf(
			"producer-qualified provenance name was rejected: %#v",
			report.Issues,
		)
	}
}
