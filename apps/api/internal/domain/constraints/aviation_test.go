package constraints

import (
	"math"
	"testing"
)

func TestIsLatitude(t *testing.T) {
	cases := []struct {
		name     string
		value    float64
		expected bool
	}{
		{name: "minimum", value: -90, expected: true},
		{name: "maximum", value: 90, expected: true},
		{name: "below minimum", value: -90.1, expected: false},
		{name: "above maximum", value: 90.1, expected: false},
		{name: "not a number", value: math.NaN(), expected: false},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			actual := IsLatitude(item.value)
			if actual != item.expected {
				t.Fatalf("expected %v, got %v", item.expected, actual)
			}
		})
	}
}

func TestIsLongitude(t *testing.T) {
	cases := []struct {
		name     string
		value    float64
		expected bool
	}{
		{name: "minimum", value: -180, expected: true},
		{name: "maximum", value: 180, expected: true},
		{name: "below minimum", value: -180.1, expected: false},
		{name: "above maximum", value: 180.1, expected: false},
		{name: "infinite", value: math.Inf(1), expected: false},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			actual := IsLongitude(item.value)
			if actual != item.expected {
				t.Fatalf("expected %v, got %v", item.expected, actual)
			}
		})
	}
}

func TestIsPercentInt(t *testing.T) {
	cases := []struct {
		name     string
		value    int
		expected bool
	}{
		{name: "zero", value: 0, expected: true},
		{name: "hundred", value: 100, expected: true},
		{name: "negative", value: -1, expected: false},
		{name: "above hundred", value: 101, expected: false},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			actual := IsPercentInt(item.value)
			if actual != item.expected {
				t.Fatalf("expected %v, got %v", item.expected, actual)
			}
		})
	}
}

func TestHeadingBounds(t *testing.T) {
	if !IsHeadingDegreesExclusive(359.9) {
		t.Fatal("expected 359.9 to be valid for exclusive heading")
	}

	if IsHeadingDegreesExclusive(360) {
		t.Fatal("expected 360 to be invalid for exclusive heading")
	}

	if !IsHeadingDegreesInclusive(360) {
		t.Fatal("expected 360 to be valid for inclusive heading")
	}

	if IsHeadingDegreesInclusive(361) {
		t.Fatal("expected 361 to be invalid for inclusive heading")
	}
}
