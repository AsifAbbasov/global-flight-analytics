package providerhealth

import (
	"errors"
	"math"
	"math/big"
)

var ErrBasisPointsInvalid = errors.New("provider health basis points must be between zero and 10000")

type BasisPoints uint16

const FullScaleBasisPoints BasisPoints = 10_000

func NewBasisPoints(value int) (BasisPoints, error) {
	if value < 0 || value > int(FullScaleBasisPoints) {
		return 0, ErrBasisPointsInvalid
	}
	return BasisPoints(value), nil
}

func BasisPointsFromRatio(value float64) (BasisPoints, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) ||
		value < 0 || value > 1 {
		return 0, ErrBasisPointsInvalid
	}
	return NewBasisPoints(
		int(math.Round(value * float64(FullScaleBasisPoints))),
	)
}

func MustBasisPointsFromRatio(value float64) BasisPoints {
	basisPoints, err := BasisPointsFromRatio(value)
	if err != nil {
		panic(err)
	}
	return basisPoints
}

func (value BasisPoints) Validate() error {
	if value > FullScaleBasisPoints {
		return ErrBasisPointsInvalid
	}
	return nil
}

func (value BasisPoints) Ratio() float64 {
	return float64(value) / float64(FullScaleBasisPoints)
}

func ratioAtLeastBasisPoints(numerator, denominator int64, threshold BasisPoints) bool {
	return compareRatioToBasisPoints(numerator, denominator, threshold) >= 0
}

func ratioExceedsBasisPoints(numerator, denominator int64, threshold BasisPoints) bool {
	return compareRatioToBasisPoints(numerator, denominator, threshold) > 0
}

func compareRatioToBasisPoints(numerator, denominator int64, threshold BasisPoints) int {
	if denominator <= 0 || threshold.Validate() != nil {
		return -1
	}
	left := new(big.Int).Mul(
		big.NewInt(numerator),
		big.NewInt(int64(FullScaleBasisPoints)),
	)
	right := new(big.Int).Mul(
		big.NewInt(denominator),
		big.NewInt(int64(threshold)),
	)
	return left.Cmp(right)
}
