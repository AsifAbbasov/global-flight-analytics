package projectionevaluation

import (
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func TestAggregateGroupsMethodsAndCalculatesMetrics(
	t *testing.T,
) {
	evaluator := newEvaluationEvaluator(t)

	first, err := evaluator.Evaluate(
		evaluationTestRequest(true),
	)
	if err != nil {
		t.Fatalf(
			"first Evaluate() error = %v",
			err,
		)
	}

	secondRequest :=
		evaluationTestRequest(false)
	secondRequest.Projection.Method =
		projectioncontract.Method{
			Name:    "second_method",
			Version: "second-method-v1",
			DecisionClass: projectioncontract.
				DecisionClassPhysicsDerived,
		}
	secondRequest.Projection.
		Provenance.InputFingerprint =
		"sha256:" +
			strings.Repeat("c", 64)
	second, err := evaluator.Evaluate(
		secondRequest,
	)
	if err != nil {
		t.Fatalf(
			"second Evaluate() error = %v",
			err,
		)
	}

	generatedAt := evaluationTestAsOfTime().
		Add(10 * time.Minute)
	aggregate, err := Aggregate(
		[]Result{second, first},
		generatedAt,
	)
	if err != nil {
		t.Fatalf(
			"Aggregate() error = %v",
			err,
		)
	}

	if aggregate.Status != StatusComplete ||
		aggregate.EvaluationCount != 2 ||
		aggregate.MethodCount != 2 ||
		len(aggregate.Methods) != 2 {
		t.Fatalf(
			"unexpected aggregate metadata: %#v",
			aggregate,
		)
	}
	if aggregate.Methods[0].MethodName !=
		"evaluation_test_method" ||
		aggregate.Methods[1].MethodName !=
			"second_method" {
		t.Fatalf(
			"methods are not deterministically ordered: %#v",
			aggregate.Methods,
		)
	}
	if aggregate.Methods[0].
		EvaluatedPointCount != 3 ||
		aggregate.Methods[0].
			ArrivalEvaluationCount != 1 ||
		aggregate.Methods[0].
			MeanHorizontalErrorM <= 0 {
		t.Fatalf(
			"unexpected first method summary: %#v",
			aggregate.Methods[0],
		)
	}
	if err := aggregate.Validate(); err != nil {
		t.Fatalf(
			"aggregate validation error = %v",
			err,
		)
	}
}

func TestAggregateIsDeterministicAcrossInputOrder(
	t *testing.T,
) {
	evaluator := newEvaluationEvaluator(t)
	first, err := evaluator.Evaluate(
		evaluationTestRequest(true),
	)
	if err != nil {
		t.Fatalf(
			"first Evaluate() error = %v",
			err,
		)
	}

	secondRequest :=
		evaluationTestRequest(false)
	secondRequest.Projection.
		Provenance.InputFingerprint =
		"sha256:" +
			strings.Repeat("d", 64)
	second, err := evaluator.Evaluate(
		secondRequest,
	)
	if err != nil {
		t.Fatalf(
			"second Evaluate() error = %v",
			err,
		)
	}

	generatedAt := evaluationTestAsOfTime().
		Add(10 * time.Minute)
	left, err := Aggregate(
		[]Result{first, second},
		generatedAt,
	)
	if err != nil {
		t.Fatalf(
			"left Aggregate() error = %v",
			err,
		)
	}
	right, err := Aggregate(
		[]Result{second, first},
		generatedAt,
	)
	if err != nil {
		t.Fatalf(
			"right Aggregate() error = %v",
			err,
		)
	}

	if left.InputFingerprint !=
		right.InputFingerprint {
		t.Fatal(
			"aggregate fingerprint depends on input order",
		)
	}
	if len(left.Methods) !=
		len(right.Methods) ||
		left.Methods[0].MeanHorizontalErrorM !=
			right.Methods[0].
				MeanHorizontalErrorM {
		t.Fatal(
			"aggregate metrics depend on input order",
		)
	}
}

func TestAggregateMarksPartialEvaluationSet(
	t *testing.T,
) {
	evaluator := newEvaluationEvaluator(t)
	complete, err := evaluator.Evaluate(
		evaluationTestRequest(false),
	)
	if err != nil {
		t.Fatalf(
			"complete Evaluate() error = %v",
			err,
		)
	}

	partialRequest :=
		evaluationTestRequest(false)
	partialRequest.ActualTrajectory.Points =
		partialRequest.ActualTrajectory.
			Points[:2]
	partial, err := evaluator.Evaluate(
		partialRequest,
	)
	if err != nil {
		t.Fatalf(
			"partial Evaluate() error = %v",
			err,
		)
	}

	aggregate, err := Aggregate(
		[]Result{complete, partial},
		evaluationTestAsOfTime().
			Add(10*time.Minute),
	)
	if err != nil {
		t.Fatalf(
			"Aggregate() error = %v",
			err,
		)
	}

	if aggregate.Status != StatusPartial ||
		!hasEvaluationNotice(
			aggregate.Limitations,
			"aggregate_contains_partial_or_unavailable_evaluations",
		) {
		t.Fatalf(
			"partial aggregate was not reported: %#v",
			aggregate,
		)
	}
}

func TestAggregateEmptyInputReturnsUnavailable(
	t *testing.T,
) {
	result, err := Aggregate(
		nil,
		evaluationTestAsOfTime(),
	)
	if err != nil {
		t.Fatalf(
			"Aggregate() error = %v",
			err,
		)
	}

	if result.Status != StatusUnavailable ||
		result.EvaluationCount != 0 ||
		len(result.Methods) != 0 ||
		len(result.Limitations) == 0 {
		t.Fatalf(
			"unexpected empty aggregate: %#v",
			result,
		)
	}
}
