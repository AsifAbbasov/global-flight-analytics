package confidencereport

import (
	"errors"
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
)

func TestRequestValidationRequiresEvidence(
	t *testing.T,
) {
	request := validConfidenceRequest()
	request.Factors = []Factor{
		Penalty(
			FactorCodeCoverageGapPenalty,
			0.25,
			0.5,
			"Coverage gaps reduce confidence.",
		),
	}

	err := request.Validate()
	if !errors.Is(
		err,
		ErrEvidenceFactorsMissing,
	) {
		t.Fatalf(
			"expected evidence requirement, got %v",
			err,
		)
	}
}

func TestRequestValidationRejectsDuplicateFactorCodes(
	t *testing.T,
) {
	request := validConfidenceRequest()
	request.Factors = append(
		request.Factors,
		request.Factors[0],
	)

	err := request.Validate()
	if !errors.Is(
		err,
		ErrDuplicateFactorCode,
	) {
		t.Fatalf(
			"expected duplicate factor error, got %v",
			err,
		)
	}
}

func TestFactorValidationRejectsInvalidFields(
	t *testing.T,
) {
	testCases := []struct {
		name     string
		factor   Factor
		expected error
	}{
		{
			name: "invalid code",
			factor: Evidence(
				"Invalid Code",
				1,
				1,
				"Valid message.",
			),
			expected: ErrFactorCodeInvalid,
		},
		{
			name: "invalid kind",
			factor: Factor{
				Code:    "valid_code",
				Kind:    FactorKind("unknown"),
				Weight:  1,
				Value:   1,
				Message: "Valid message.",
			},
			expected: ErrFactorKindInvalid,
		},
		{
			name: "invalid weight",
			factor: Evidence(
				"valid_code",
				0,
				1,
				"Valid message.",
			),
			expected: ErrFactorWeightInvalid,
		},
		{
			name: "invalid value",
			factor: Evidence(
				"valid_code",
				1,
				math.Inf(1),
				"Valid message.",
			),
			expected: ErrFactorValueInvalid,
		},
		{
			name: "invalid message",
			factor: Evidence(
				"valid_code",
				1,
				1,
				" ",
			),
			expected: ErrFactorMessageInvalid,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				err := testCase.factor.Validate()

				if !errors.Is(
					err,
					testCase.expected,
				) {
					t.Fatalf(
						"expected %v, got %v",
						testCase.expected,
						err,
					)
				}
			},
		)
	}
}

func TestRequestValidationRejectsDuplicateNoticeCodes(
	t *testing.T,
) {
	request := validConfidenceRequest()
	request.Warnings = []analyticalresult.Notice{
		{
			Code:    "partial_coverage",
			Message: "Coverage is partial.",
		},
		{
			Code:    "partial_coverage",
			Message: "Coverage remains partial.",
		},
	}

	err := request.Validate()
	if !errors.Is(
		err,
		ErrDuplicateNoticeCode,
	) {
		t.Fatalf(
			"expected duplicate notice error, got %v",
			err,
		)
	}
}
