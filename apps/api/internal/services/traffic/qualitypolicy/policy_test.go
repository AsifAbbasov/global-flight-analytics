package qualitypolicy

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
)

func TestClampScore(t *testing.T) {
	cases := []struct {
		name     string
		score    float64
		expected float64
	}{
		{name: "below minimum", score: -0.5, expected: 0},
		{name: "inside range", score: 0.75, expected: 0.75},
		{name: "above maximum", score: 1.5, expected: 1},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			actual := ClampScore(item.score)
			if actual != item.expected {
				t.Fatalf("expected %f, got %f", item.expected, actual)
			}
		})
	}
}

func TestConfidenceFromScore(t *testing.T) {
	cases := []struct {
		name     string
		score    float64
		expected dataquality.ConfidenceLevel
	}{
		{name: "high boundary", score: HighConfidenceMinimumScore, expected: dataquality.ConfidenceLevelHigh},
		{name: "medium boundary", score: MediumConfidenceMinimumScore, expected: dataquality.ConfidenceLevelMedium},
		{name: "low", score: 0.10, expected: dataquality.ConfidenceLevelLow},
		{name: "none", score: 0, expected: dataquality.ConfidenceLevelNone},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			actual := ConfidenceFromScore(item.score)
			if actual != item.expected {
				t.Fatalf("expected %s, got %s", item.expected, actual)
			}
		})
	}
}

func TestApplyPenalty(t *testing.T) {
	actual := ApplyPenalty(StartingScore, InvalidICAO24Penalty)
	expected := 0.5

	if actual != expected {
		t.Fatalf("expected %f, got %f", expected, actual)
	}
}
