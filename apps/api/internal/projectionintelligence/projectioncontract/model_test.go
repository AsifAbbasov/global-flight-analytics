package projectioncontract

import (
	"testing"
	"time"
)

func TestResultCloneDoesNotShareMutableState(
	t *testing.T,
) {
	result := validProjectionResult()
	cloned := result.Clone()

	*cloned.Points[0].Position.AltitudeM = 1
	*cloned.Points[0].Uncertainty.
		VerticalRadiusM = 2
	cloned.Points[0].Confidence.
		Reasons[0].Code = "changed"
	cloned.Arrival.Confidence.
		Reasons[0].Message = "changed"
	cloned.Arrival.Limitations[0].Code =
		"changed"
	cloned.Confidence.Reasons[0].
		Contribution = 999
	cloned.Limitations[0].Message =
		"changed"
	cloned.Explanations[0].Code =
		"changed"
	cloned.Provenance.Inputs[0].Name =
		"changed"

	if *result.Points[0].Position.
		AltitudeM != 11000 {
		t.Fatal(
			"Result.Clone() shared altitude pointer",
		)
	}
	if *result.Points[0].Uncertainty.
		VerticalRadiusM != 120 {
		t.Fatal(
			"Result.Clone() shared vertical uncertainty pointer",
		)
	}
	if result.Points[0].Confidence.
		Reasons[0].Code !=
		"trajectory_quality" {
		t.Fatal(
			"Result.Clone() shared point confidence reasons",
		)
	}
	if result.Arrival.Confidence.
		Reasons[0].Message !=
		"Arrival estimate uses the selected projection baseline." {
		t.Fatal(
			"Result.Clone() shared arrival confidence reasons",
		)
	}
	if result.Arrival.Limitations[0].Code !=
		"no_official_flight_plan" {
		t.Fatal(
			"Result.Clone() shared arrival limitations",
		)
	}
	if result.Confidence.Reasons[0].
		Contribution != 0.7 {
		t.Fatal(
			"Result.Clone() shared result confidence reasons",
		)
	}
	if result.Limitations[0].Message !=
		"Projection is a research estimate, not an operational instruction." {
		t.Fatal(
			"Result.Clone() shared result limitations",
		)
	}
	if result.Explanations[0].Code !=
		"short_horizon_baseline" {
		t.Fatal(
			"Result.Clone() shared explanations",
		)
	}
	if result.Provenance.Inputs[0].Name !=
		"latest_trajectory_point" {
		t.Fatal(
			"Result.Clone() shared provenance inputs",
		)
	}
}

func TestHorizonDuration(
	t *testing.T,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		15,
		17,
		0,
		0,
		0,
		time.UTC,
	)
	horizon := Horizon{
		AsOfTime: asOfTime,
		EndTime: asOfTime.Add(
			10 * time.Minute,
		),
		Step: time.Minute,
	}

	if horizon.Duration() !=
		10*time.Minute {
		t.Fatalf(
			"Duration() = %s, want %s",
			horizon.Duration(),
			10*time.Minute,
		)
	}
}

func TestKnownEnums(
	t *testing.T,
) {
	if !ResultStatusComplete.IsKnown() {
		t.Fatal(
			"complete result status must be known",
		)
	}
	if !DecisionClassPhysicsDerived.IsKnown() {
		t.Fatal(
			"physics-derived decision class must be known",
		)
	}
	if !InputClassificationObserved.IsKnown() {
		t.Fatal(
			"observed input classification must be known",
		)
	}
	if !ConfidenceLevelHigh.IsKnown() {
		t.Fatal(
			"high confidence level must be known",
		)
	}
}

func float64Pointer(
	value float64,
) *float64 {
	return &value
}
