package datasetprofiler

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestNumericAccumulatorProfilesDistribution(t *testing.T) {
	accumulator := numericAccumulator{}
	for _, value := range []float64{4, 1, 3, 2} {
		accumulator.add(value, true)
	}
	accumulator.add(math.NaN(), true)
	accumulator.add(5, false)

	profile := accumulator.profile()
	if profile.Count != 4 ||
		profile.InvalidCount != 2 ||
		profile.Minimum != 1 ||
		profile.Maximum != 4 ||
		profile.Mean != 2.5 ||
		profile.Median != 2.5 ||
		math.Abs(profile.Percentile95-3.85) > 1e-12 {
		t.Fatalf("unexpected numeric profile: %#v", profile)
	}
}

func TestFrequencyAccumulatorSortsDeterministically(t *testing.T) {
	accumulator := newFrequencyAccumulator()
	accumulator.addRecordValues([]string{"beta", "alpha", "alpha"})
	accumulator.addRecordValues([]string{"gamma", "alpha"})
	accumulator.addRecordValues([]string{"beta"})

	want := []FrequencyProfile{
		{
			Value:               "alpha",
			OccurrenceCount:     3,
			AffectedRecordCount: 2,
		},
		{
			Value:               "beta",
			OccurrenceCount:     2,
			AffectedRecordCount: 2,
		},
		{
			Value:               "gamma",
			OccurrenceCount:     1,
			AffectedRecordCount: 1,
		},
	}
	if got := accumulator.profiles(); !reflect.DeepEqual(got, want) {
		t.Fatalf("profiles = %#v, want %#v", got, want)
	}
}

func TestTimeHelpersNormalizeToUTC(t *testing.T) {
	location := time.FixedZone("UTC+04", 4*60*60)
	first := time.Date(2026, time.July, 14, 12, 0, 0, 0, location)
	second := first.Add(time.Hour)

	earliest := earlierNonZero(time.Time{}, second)
	earliest = earlierNonZero(earliest, first)
	latest := laterNonZero(time.Time{}, first)
	latest = laterNonZero(latest, second)

	if earliest.Location() != time.UTC ||
		latest.Location() != time.UTC ||
		!earliest.Equal(first.UTC()) ||
		!latest.Equal(second.UTC()) {
		t.Fatalf("unexpected UTC normalization: earliest=%v latest=%v", earliest, latest)
	}
}

func TestPercentileBoundaries(t *testing.T) {
	values := []float64{1, 2, 3}
	if percentile(values, 0) != 1 ||
		percentile(values, 1) != 3 ||
		percentile(values, 0.5) != 2 ||
		percentile(nil, 0.5) != 0 {
		t.Fatal("unexpected percentile boundary behavior")
	}
}
