package projectionarrival

import (
	"math"
	"testing"
	"time"
)

func TestCalculateClosingSpeedProfileUsesLatestBoundedSamples(
	t *testing.T,
) {
	start := arrivalTestAsOfTime()
	samples := []positionSample{
		{timeValue: start, latitude: 0, longitude: 0},
		{timeValue: start.Add(time.Minute), latitude: 0, longitude: 0.01},
		{timeValue: start.Add(2 * time.Minute), latitude: 0, longitude: 0.02},
		{timeValue: start.Add(3 * time.Minute), latitude: 0, longitude: 0.04},
	}
	distances, valid := arrivalDistances(samples, 0, 0.10)
	if !valid {
		t.Fatal("arrivalDistances() returned invalid")
	}

	profile, valid := calculateClosingSpeedProfile(
		samples,
		distances,
		400,
		2,
	)
	if !valid {
		t.Fatal("calculateClosingSpeedProfile() returned invalid")
	}
	if profile.sampleCount != 2 {
		t.Fatalf("sample count = %d, want 2", profile.sampleCount)
	}
	if profile.meanClosingSpeedMPS <= 0 ||
		profile.closingSpeedStdDevMPS <= 0 ||
		profile.maximumClosingSpeedMPS <=
			profile.minimumClosingSpeedMPS {
		t.Fatalf("unexpected profile: %#v", profile)
	}
}

func TestCalculateClosingSpeedProfilePreservesSlowAndRecedingSamples(
	t *testing.T,
) {
	start := arrivalTestAsOfTime()
	samples := []positionSample{
		{timeValue: start, latitude: 0, longitude: 0},
		{timeValue: start.Add(time.Minute), latitude: 0, longitude: 0.001},
		{timeValue: start.Add(2 * time.Minute), latitude: 0, longitude: 0.0011},
		{timeValue: start.Add(3 * time.Minute), latitude: 0, longitude: 0.0010},
	}
	distances, valid := arrivalDistances(samples, 0, 1)
	if !valid {
		t.Fatal("arrivalDistances() returned invalid")
	}

	profile, valid := calculateClosingSpeedProfile(
		samples,
		distances,
		400,
		4,
	)
	if !valid {
		t.Fatal("calculateClosingSpeedProfile() returned invalid")
	}
	if profile.sampleCount != 3 {
		t.Fatalf(
			"sample count = %d, want all 3 slow/directional samples",
			profile.sampleCount,
		)
	}
	if profile.minimumClosingSpeedMPS >= 0 {
		t.Fatalf(
			"minimum closing speed = %f, want receding sample below zero",
			profile.minimumClosingSpeedMPS,
		)
	}
}

func TestCalculateClosingSpeedProfileRejectsImpossibleGroundSpeed(
	t *testing.T,
) {
	start := arrivalTestAsOfTime()
	samples := []positionSample{
		{timeValue: start, latitude: 0, longitude: 0},
		{timeValue: start.Add(time.Second), latitude: 0, longitude: 1},
	}
	distances, valid := arrivalDistances(samples, 0, 2)
	if !valid {
		t.Fatal("arrivalDistances() returned invalid")
	}

	_, valid = calculateClosingSpeedProfile(
		samples,
		distances,
		400,
		4,
	)
	if valid {
		t.Fatal("impossible ground speed was accepted")
	}
}

func TestEnforceMinimumArrivalInterval(
	t *testing.T,
) {
	asOfTime := arrivalTestAsOfTime()
	estimatedTime := asOfTime.Add(10 * time.Minute)

	earliest, estimated, latest := enforceMinimumArrivalInterval(
		asOfTime,
		estimatedTime,
		estimatedTime,
		estimatedTime,
		4*time.Minute,
	)

	if !estimated.Equal(estimatedTime) ||
		!earliest.Equal(estimatedTime.Add(-2*time.Minute)) ||
		!latest.Equal(estimatedTime.Add(2*time.Minute)) {
		t.Fatalf(
			"unexpected interval: %s %s %s",
			earliest,
			estimated,
			latest,
		)
	}
}

func TestDurationCeilSecondsRoundsUpAndRejectsOverflow(
	t *testing.T,
) {
	duration, valid := durationCeilSeconds(0.0000000001)
	if !valid || duration != time.Nanosecond {
		t.Fatalf("duration = %s valid=%t, want 1ns", duration, valid)
	}

	if _, valid := durationCeilSeconds(math.MaxFloat64); valid {
		t.Fatal("overflowing duration was accepted")
	}
}

func TestGreatCircleDistanceAcrossDateline(
	t *testing.T,
) {
	distanceM := greatCircleDistanceM(
		0,
		179.9,
		0,
		-179.9,
	)
	if math.Abs(distanceM-22239) > 100 {
		t.Fatalf(
			"distance = %f, want approximately 22239",
			distanceM,
		)
	}
}
