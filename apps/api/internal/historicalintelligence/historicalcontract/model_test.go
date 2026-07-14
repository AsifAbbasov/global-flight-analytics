package historicalcontract

import (
	"reflect"
	"testing"
)

func TestResultCloneDoesNotShareMutableState(
	t *testing.T,
) {
	result := validCompleteResult()
	cloned := result.Clone()

	cloned.Points[0].Confidence.Reasons[0].Code =
		"changed"
	cloned.Points[0].Limitations = append(
		cloned.Points[0].Limitations,
		Limitation{
			Code:    "changed",
			Message: "Changed.",
			Scope:   "point",
		},
	)
	cloned.Confidence.Reasons[0].Message =
		"Changed."
	cloned.Limitations[0].Code = "changed"
	cloned.Provenance.SourceNames[0] =
		"changed"
	*cloned.Comparison.PercentageChange = 999

	if result.Points[0].Confidence.
		Reasons[0].Code !=
		"observed_samples" {
		t.Fatal(
			"Result.Clone() shared point confidence state",
		)
	}
	if len(result.Points[0].Limitations) != 0 {
		t.Fatal(
			"Result.Clone() shared point limitations",
		)
	}
	if result.Confidence.Reasons[0].Message !=
		"Historical source coverage is complete." {
		t.Fatal(
			"Result.Clone() shared series confidence state",
		)
	}
	if result.Limitations[0].Code !=
		"historical_observation_only" {
		t.Fatal(
			"Result.Clone() shared result limitations",
		)
	}
	if result.Provenance.SourceNames[0] !=
		"flight_trajectories" {
		t.Fatal(
			"Result.Clone() shared provenance sources",
		)
	}
	if *result.Comparison.PercentageChange != 50 {
		t.Fatal(
			"Result.Clone() shared comparison percentage pointer",
		)
	}
}

func TestConfidenceLevelForScore(
	t *testing.T,
) {
	tests := []struct {
		score float64
		want  ConfidenceLevel
	}{
		{score: 0, want: ConfidenceLevelNone},
		{score: 0.1, want: ConfidenceLevelLow},
		{score: 0.599, want: ConfidenceLevelLow},
		{score: 0.6, want: ConfidenceLevelMedium},
		{score: 0.799, want: ConfidenceLevelMedium},
		{score: 0.8, want: ConfidenceLevelHigh},
		{score: 1, want: ConfidenceLevelHigh},
	}

	for _, test := range tests {
		if got := ConfidenceLevelForScore(
			test.score,
		); got != test.want {
			t.Fatalf(
				"ConfidenceLevelForScore(%v) = %q, want %q",
				test.score,
				got,
				test.want,
			)
		}
	}
}

func TestTrendDirectionForChange(
	t *testing.T,
) {
	tests := []struct {
		change float64
		want   TrendDirection
	}{
		{
			change: -10,
			want:   TrendDirectionDown,
		},
		{
			change: 0,
			want:   TrendDirectionFlat,
		},
		{
			change: 1e-10,
			want:   TrendDirectionFlat,
		},
		{
			change: 10,
			want:   TrendDirectionUp,
		},
	}

	for _, test := range tests {
		if got := TrendDirectionForChange(
			test.change,
		); got != test.want {
			t.Fatalf(
				"TrendDirectionForChange(%v) = %q, want %q",
				test.change,
				got,
				test.want,
			)
		}
	}
}

func TestSupportedMetricNamesAreSortedAndDefensive(
	t *testing.T,
) {
	first := SupportedMetricNames()
	second := SupportedMetricNames()

	if len(first) != 20 {
		t.Fatalf(
			"supported metric count = %d, want 20",
			len(first),
		)
	}
	for index := 1; index < len(first); index++ {
		if first[index] <= first[index-1] {
			t.Fatalf(
				"metric names are not strictly sorted: %#v",
				first,
			)
		}
	}

	first[0] = "changed"
	if reflect.DeepEqual(first, second) ||
		second[0] == "changed" {
		t.Fatal(
			"SupportedMetricNames() returned shared state",
		)
	}
}

func TestTimeWindowDuration(
	t *testing.T,
) {
	result := validCompleteResult()

	if got := result.Window.Duration(); got != 72*60*60*1e9 {
		t.Fatalf(
			"Duration() = %s",
			got,
		)
	}
}
