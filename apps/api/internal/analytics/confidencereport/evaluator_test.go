package confidencereport

import (
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
)

func TestEvaluateProducesHighConfidenceReport(
	t *testing.T,
) {
	report, err := NewDefault().Evaluate(
		validConfidenceRequest(),
	)
	if err != nil {
		t.Fatalf(
			"expected confidence report, got %v",
			err,
		)
	}

	if report.BaseScore != 0.87 {
		t.Fatalf(
			"expected base score 0.87, got %f",
			report.BaseScore,
		)
	}

	if report.PenaltyScore != 0 ||
		report.Score != 0.87 {
		t.Fatalf(
			"unexpected score values: penalty=%f score=%f",
			report.PenaltyScore,
			report.Score,
		)
	}

	if report.Level !=
		analyticalresult.ConfidenceLevelHigh {
		t.Fatalf(
			"expected high confidence, got %s",
			report.Level,
		)
	}

	if len(report.Factors) != 5 ||
		len(report.Reasons) != 5 {
		t.Fatalf(
			"expected five factors and reasons, got %d and %d",
			len(report.Factors),
			len(report.Reasons),
		)
	}

	if err := report.AnalyticalConfidence().
		Validate(); err != nil {
		t.Fatalf(
			"expected valid analytical confidence, got %v",
			err,
		)
	}
}

func TestEvaluateAppliesPenaltyAndProducesMediumConfidence(
	t *testing.T,
) {
	request := validConfidenceRequest()
	request.Factors = append(
		request.Factors,
		Penalty(
			FactorCodeCoverageGapPenalty,
			0.20,
			0.50,
			"Coverage gaps reduce confidence.",
		),
	)

	report, err := NewDefault().Evaluate(request)
	if err != nil {
		t.Fatalf(
			"expected confidence report, got %v",
			err,
		)
	}

	if report.PenaltyScore != 0.10 ||
		report.Score != 0.77 {
		t.Fatalf(
			"expected penalty 0.10 and score 0.77, got %f and %f",
			report.PenaltyScore,
			report.Score,
		)
	}

	if report.Level !=
		analyticalresult.ConfidenceLevelMedium {
		t.Fatalf(
			"expected medium confidence, got %s",
			report.Level,
		)
	}

	if !report.HasPenalty() {
		t.Fatal("expected penalty to be reported")
	}

	contribution, exists := report.Factor(
		FactorCodeCoverageGapPenalty,
	)
	if !exists {
		t.Fatal("expected coverage gap penalty contribution")
	}

	if contribution.Impact != -0.10 {
		t.Fatalf(
			"expected penalty impact -0.10, got %f",
			contribution.Impact,
		)
	}
}

func TestEvaluateProducesLowAndNoneConfidenceLevels(
	t *testing.T,
) {
	evaluator := NewDefault()
	evaluatedAt := confidenceTestTime()

	low, err := evaluator.Evaluate(
		Request{
			Factors: []Factor{
				Evidence(
					FactorCodeTrajectoryQuality,
					1,
					0.40,
					"Trajectory quality is limited.",
				),
			},
			EvaluatedAt: evaluatedAt,
		},
	)
	if err != nil {
		t.Fatalf(
			"expected low confidence report, got %v",
			err,
		)
	}

	if low.Level !=
		analyticalresult.ConfidenceLevelLow ||
		low.Score != 0.40 {
		t.Fatalf(
			"expected low confidence score 0.40, got %s %f",
			low.Level,
			low.Score,
		)
	}

	none, err := evaluator.Evaluate(
		Request{
			Factors: []Factor{
				Evidence(
					FactorCodeTrajectoryQuality,
					1,
					0.40,
					"Trajectory quality is limited.",
				),
				Penalty(
					FactorCodeCoverageGapPenalty,
					1,
					1,
					"Coverage gaps remove confidence.",
				),
			},
			EvaluatedAt: evaluatedAt,
		},
	)
	if err != nil {
		t.Fatalf(
			"expected none confidence report, got %v",
			err,
		)
	}

	if none.Level !=
		analyticalresult.ConfidenceLevelNone ||
		none.Score != 0 {
		t.Fatalf(
			"expected none confidence score zero, got %s %f",
			none.Level,
			none.Score,
		)
	}
}

func TestEvaluateCapsPenaltyAndScalesPenaltyImpacts(
	t *testing.T,
) {
	config := DefaultConfig()
	config.MaximumPenalty = 0.20

	evaluator, err := New(config)
	if err != nil {
		t.Fatalf(
			"expected evaluator, got %v",
			err,
		)
	}

	report, err := evaluator.Evaluate(
		Request{
			Factors: []Factor{
				Evidence(
					FactorCodeTrajectoryQuality,
					1,
					1,
					"Trajectory quality is complete.",
				),
				Penalty(
					FactorCodeCoverageGapPenalty,
					0.50,
					1,
					"Coverage gaps reduce confidence.",
				),
				Penalty(
					FactorCodeProviderDegradationPenalty,
					0.50,
					1,
					"Provider degradation reduces confidence.",
				),
			},
			EvaluatedAt: confidenceTestTime(),
		},
	)
	if err != nil {
		t.Fatalf(
			"expected capped confidence report, got %v",
			err,
		)
	}

	if report.PenaltyScore != 0.20 ||
		report.Score != 0.80 {
		t.Fatalf(
			"expected capped penalty 0.20 and score 0.80, got %f and %f",
			report.PenaltyScore,
			report.Score,
		)
	}

	first, _ := report.Factor(
		FactorCodeCoverageGapPenalty,
	)
	second, _ := report.Factor(
		FactorCodeProviderDegradationPenalty,
	)

	if first.Impact != -0.10 ||
		second.Impact != -0.10 {
		t.Fatalf(
			"expected scaled penalty impacts, got %f and %f",
			first.Impact,
			second.Impact,
		)
	}
}

func TestEvaluateIsDeterministicAndDoesNotMutateRequest(
	t *testing.T,
) {
	request := validConfidenceRequest()
	request.Factors[0], request.Factors[4] =
		request.Factors[4], request.Factors[0]
	request.Warnings = []analyticalresult.Notice{
		{
			Code:    "z_warning",
			Message: "Second warning.",
		},
		{
			Code:    "a_warning",
			Message: "First warning.",
		},
	}

	original := cloneRequest(request)
	evaluator := NewDefault()

	first, err := evaluator.Evaluate(request)
	if err != nil {
		t.Fatalf(
			"expected first report, got %v",
			err,
		)
	}

	second, err := evaluator.Evaluate(request)
	if err != nil {
		t.Fatalf(
			"expected second report, got %v",
			err,
		)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf(
			"expected deterministic reports, got %#v and %#v",
			first,
			second,
		)
	}

	if !reflect.DeepEqual(request, original) {
		t.Fatal("expected request not to be mutated")
	}

	expectedFactorCodes := []string{
		FactorCodeDataFreshness,
		FactorCodeIdentityReliability,
		FactorCodeObservationCoverage,
		FactorCodeSourceCoverage,
		FactorCodeTrajectoryQuality,
	}

	for index, expected := range expectedFactorCodes {
		if first.Factors[index].Code != expected {
			t.Fatalf(
				"expected factor %s at index %d, got %s",
				expected,
				index,
				first.Factors[index].Code,
			)
		}
	}

	if first.Warnings[0].Code != "a_warning" ||
		first.Warnings[1].Code != "z_warning" {
		t.Fatalf(
			"expected sorted warnings, got %#v",
			first.Warnings,
		)
	}
}

func TestReportCloneAndAnalyticalConfidenceCopySlices(
	t *testing.T,
) {
	report, err := NewDefault().Evaluate(
		validConfidenceRequest(),
	)
	if err != nil {
		t.Fatalf(
			"expected report, got %v",
			err,
		)
	}

	clone := report.Clone()
	clone.Factors[0].Code = "mutated_factor"
	clone.Reasons[0].Code = "mutated_reason"

	confidence := report.AnalyticalConfidence()
	confidence.Reasons[0].Code =
		"mutated_confidence_reason"

	if report.Factors[0].Code ==
		"mutated_factor" {
		t.Fatal("expected factor slice to be copied")
	}

	if report.Reasons[0].Code ==
		"mutated_reason" ||
		report.Reasons[0].Code ==
			"mutated_confidence_reason" {
		t.Fatal("expected reason slices to be copied")
	}
}

func validConfidenceRequest() Request {
	return Request{
		Factors: []Factor{
			Evidence(
				FactorCodeTrajectoryQuality,
				0.40,
				0.90,
				"Trajectory quality strongly supports this result.",
			),
			Evidence(
				FactorCodeIdentityReliability,
				0.20,
				1.00,
				"Flight identity is reliable.",
			),
			Evidence(
				FactorCodeDataFreshness,
				0.15,
				0.80,
				"Observations are sufficiently fresh.",
			),
			Evidence(
				FactorCodeObservationCoverage,
				0.15,
				0.80,
				"Observation coverage is sufficient.",
			),
			Evidence(
				FactorCodeSourceCoverage,
				0.10,
				0.70,
				"Source coverage supports the result.",
			),
		},
		EvaluatedAt: confidenceTestTime(),
	}
}

func confidenceTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		13,
		17,
		0,
		0,
		0,
		time.UTC,
	)
}

func cloneRequest(
	request Request,
) Request {
	result := request
	result.Factors = append(
		[]Factor(nil),
		request.Factors...,
	)
	result.Warnings = append(
		[]analyticalresult.Notice(nil),
		request.Warnings...,
	)
	result.Limitations = append(
		[]analyticalresult.Notice(nil),
		request.Limitations...,
	)

	return result
}
