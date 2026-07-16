package forecastanalysis

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrInvalidRequest = errors.New("forecast stability analysis request is invalid")
	ErrInvalidPolicy  = errors.New("forecast stability analysis policy is invalid")
	ErrInvalidResult  = errors.New("forecast stability analysis result is invalid")
)

const PolicyVersionV1 = "forecast-stability-analysis-policy-v1-experimental"

type Policy struct {
	Version                        string
	MinimumVersionCount            int
	MaximumVersionCount            int
	MinimumComparableShare         float64
	StableHealthShare              float64
	UnstableHealthShare            float64
	MaterialChangeShareForUnstable float64
	TrendScoreDelta                float64
	VolatileScoreStandardDeviation float64
	MinimumTrendTransitions        int
}

func DefaultPolicy() Policy {
	return Policy{
		Version:                        PolicyVersionV1,
		MinimumVersionCount:            2,
		MaximumVersionCount:            120,
		MinimumComparableShare:         0.70,
		StableHealthShare:              0.80,
		UnstableHealthShare:            0.50,
		MaterialChangeShareForUnstable: 0.25,
		TrendScoreDelta:                0.10,
		VolatileScoreStandardDeviation: 0.25,
		MinimumTrendTransitions:        3,
	}
}

func (policy Policy) Validate() error {
	if strings.TrimSpace(policy.Version) != PolicyVersionV1 ||
		policy.MinimumVersionCount < 2 ||
		policy.MaximumVersionCount < policy.MinimumVersionCount ||
		policy.MaximumVersionCount > 1000 ||
		!unitInterval(policy.MinimumComparableShare) ||
		!unitInterval(policy.StableHealthShare) ||
		!unitInterval(policy.UnstableHealthShare) ||
		policy.UnstableHealthShare >= policy.StableHealthShare ||
		!unitInterval(policy.MaterialChangeShareForUnstable) ||
		!positiveFinite(policy.TrendScoreDelta) ||
		!positiveFinite(policy.VolatileScoreStandardDeviation) ||
		policy.MinimumTrendTransitions < 2 {
		return fmt.Errorf("%w: thresholds", ErrInvalidPolicy)
	}
	return nil
}

func unitInterval(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}

func positiveFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value > 0
}
