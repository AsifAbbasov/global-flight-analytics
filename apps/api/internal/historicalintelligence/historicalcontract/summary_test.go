package historicalcontract

import (
	"reflect"
	"testing"
)

func TestSummarizeIgnoresUnavailablePoints(
	t *testing.T,
) {
	result := validCompleteResult()
	points := []Point{
		result.Points[0],
		{
			Status: BucketStatusUnavailable,
			Value:  999,
		},
		result.Points[1],
		result.Points[2],
	}

	got := Summarize(points)
	want := Summary{
		PointCount: 3,
		Total:      60,
		Minimum:    10,
		Maximum:    30,
		Average:    20,
		Median:     20,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"Summarize() = %#v, want %#v",
			got,
			want,
		)
	}
}

func TestSummarizeEvenMedian(
	t *testing.T,
) {
	points := []Point{
		{
			Status: BucketStatusComplete,
			Value:  40,
		},
		{
			Status: BucketStatusComplete,
			Value:  10,
		},
		{
			Status: BucketStatusComplete,
			Value:  20,
		},
		{
			Status: BucketStatusComplete,
			Value:  30,
		},
	}

	got := Summarize(points)
	if got.Median != 25 ||
		got.Average != 25 ||
		got.Total != 100 {
		t.Fatalf(
			"unexpected summary: %#v",
			got,
		)
	}
}

func TestSummarizeEmpty(
	t *testing.T,
) {
	if got := Summarize(nil); !reflect.DeepEqual(got, Summary{}) {
		t.Fatalf(
			"Summarize(nil) = %#v",
			got,
		)
	}
}
