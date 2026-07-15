package projectionarrival

import (
	"math"
	"testing"
	"time"
)

func TestCalculateSpeedProfileUsesLatestBoundedSamples(
	t *testing.T,
) {
	start := arrivalTestAsOfTime()
	samples := []positionSample{
		{
			timeValue: start,
			latitude:  0,
			longitude: 0,
		},
		{
			timeValue: start.Add(time.Minute),
			latitude:  0,
			longitude: 0.01,
		},
		{
			timeValue: start.Add(2 * time.Minute),
			latitude:  0,
			longitude: 0.02,
		},
		{
			timeValue: start.Add(3 * time.Minute),
			latitude:  0,
			longitude: 0.04,
		},
	}

	profile, valid :=
		calculateSpeedProfile(
			samples,
			5,
			2,
		)
	if !valid {
		t.Fatal(
			"calculateSpeedProfile() returned invalid",
		)
	}
	if profile.sampleCount != 2 {
		t.Fatalf(
			"sample count = %d, want 2",
			profile.sampleCount,
		)
	}
	if profile.meanMPS <= 0 ||
		profile.stdDevMPS <= 0 ||
		profile.maximumMPS <=
			profile.minimumMPS {
		t.Fatalf(
			"unexpected profile: %#v",
			profile,
		)
	}
}

func TestEnforceMinimumArrivalInterval(
	t *testing.T,
) {
	asOfTime := arrivalTestAsOfTime()
	estimatedTime :=
		asOfTime.Add(10 * time.Minute)

	earliest, estimated, latest :=
		enforceMinimumArrivalInterval(
			asOfTime,
			estimatedTime,
			estimatedTime,
			estimatedTime,
			4*time.Minute,
		)

	if !estimated.Equal(estimatedTime) ||
		!earliest.Equal(
			estimatedTime.Add(
				-2*time.Minute,
			),
		) ||
		!latest.Equal(
			estimatedTime.Add(
				2*time.Minute,
			),
		) {
		t.Fatalf(
			"unexpected interval: %s %s %s",
			earliest,
			estimated,
			latest,
		)
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
