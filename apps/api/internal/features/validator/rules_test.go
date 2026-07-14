package validator

import (
	"math"
	"reflect"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestSchemaFieldCountsMatchVersionOneContract(t *testing.T) {
	got := schemaFieldCounts()
	want := map[flightfeatures.FeatureGroup]int{
		flightfeatures.FeatureGroupTemporal:     8,
		flightfeatures.FeatureGroupGeographical: 11,
		flightfeatures.FeatureGroupOperational:  11,
		flightfeatures.FeatureGroupTrajectory:   16,
		flightfeatures.FeatureGroupAircraft:     6,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"schema field counts = %#v, want %#v",
			got,
			want,
		)
	}
}

func TestApproximatelyEqualUsesRelativeTolerance(t *testing.T) {
	if !approximatelyEqual(1_000_000, 1_000_000.5, 1e-6) {
		t.Fatal("expected large values within relative tolerance")
	}
	if approximatelyEqual(1, 1.01, 1e-6) {
		t.Fatal("expected values outside tolerance")
	}
	if approximatelyEqual(math.NaN(), 1, 1e-6) {
		t.Fatal("expected non-finite value to be rejected")
	}
}

func TestMergeLimitationsDeduplicatesValidationIssues(
	t *testing.T,
) {
	existing := []flightfeatures.FeatureLimitation{
		{
			Code:    "domain",
			Message: "Domain limitation.",
		},
	}
	issues := []Issue{
		{
			Code:    issueCodePrefix + "one",
			Message: "Validation issue.",
		},
		{
			Code:    issueCodePrefix + "one",
			Message: "Validation issue.",
		},
	}

	result := mergeLimitations(existing, issues)
	if len(result) != 2 {
		t.Fatalf(
			"mergeLimitations() = %#v, want two items",
			result,
		)
	}
}

func TestStripValidatorLimitationsPreservesDomainLimitations(
	t *testing.T,
) {
	input := []flightfeatures.FeatureLimitation{
		{
			Code:    "domain",
			Message: "Domain limitation.",
		},
		{
			Code:    issueCodePrefix + "old",
			Message: "Old validation issue.",
		},
	}

	result := stripValidatorLimitations(input)
	if len(result) != 1 || result[0].Code != "domain" {
		t.Fatalf(
			"stripValidatorLimitations() = %#v",
			result,
		)
	}
}
