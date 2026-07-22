package providerhealth

import (
	"math"
	"math/big"
)

type BasisPoints uint16

const FullScaleBasisPoints BasisPoints = 10_000

func ratioToBasisPoints(value float64) BasisPoints {
	return BasisPoints(math.Round(value * float64(FullScaleBasisPoints)))
}

func ratioAtLeastBasisPoints(numerator, denominator int64, threshold BasisPoints) bool {
	return compareRatioToBasisPoints(numerator, denominator, threshold) >= 0
}

func ratioExceedsBasisPoints(numerator, denominator int64, threshold BasisPoints) bool {
	return compareRatioToBasisPoints(numerator, denominator, threshold) > 0
}

func compareRatioToBasisPoints(numerator, denominator int64, threshold BasisPoints) int {
	if denominator <= 0 {
		return -1
	}
	left := new(big.Int).Mul(big.NewInt(numerator), big.NewInt(int64(FullScaleBasisPoints)))
	right := new(big.Int).Mul(big.NewInt(denominator), big.NewInt(int64(threshold)))
	return left.Cmp(right)
}
