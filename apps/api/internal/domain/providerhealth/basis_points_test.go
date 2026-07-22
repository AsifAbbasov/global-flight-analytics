package providerhealth

import (
	"errors"
	"math"
	"testing"
)

func TestBasisPointsFromRatioBuildsTypedThreshold(t *testing.T) {
	value, err := BasisPointsFromRatio(0.95)
	if err != nil {
		t.Fatalf("BasisPointsFromRatio() error = %v", err)
	}
	if value != 9_500 || value.Ratio() != 0.95 {
		t.Fatalf("basis points = %d ratio=%v", value, value.Ratio())
	}
}

func TestBasisPointsRejectInvalidRatio(t *testing.T) {
	for _, value := range []float64{-0.01, 1.01, math.NaN(), math.Inf(1)} {
		_, err := BasisPointsFromRatio(value)
		if !errors.Is(err, ErrBasisPointsInvalid) {
			t.Fatalf("value %v error = %v", value, err)
		}
	}
}

func TestBasisPointComparisonPreservesLargeCounterPrecision(t *testing.T) {
	const denominator int64 = 9_007_199_254_740_993
	numerator := denominator - 1
	if !ratioAtLeastBasisPoints(
		numerator,
		denominator,
		MustBasisPointsFromRatio(0.9999),
	) {
		t.Fatal("large exact ratio should satisfy 99.99 percent threshold")
	}
}
