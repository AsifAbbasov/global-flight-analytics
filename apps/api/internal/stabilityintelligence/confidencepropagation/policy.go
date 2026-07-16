package confidencepropagation

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrInvalidRequest  = errors.New("confidence propagation request is invalid")
	ErrInvalidPolicy   = errors.New("confidence propagation policy is invalid")
	ErrInvalidResult   = errors.New("confidence propagation result is invalid")
	ErrDependencyCycle = errors.New("confidence dependency graph contains a cycle")
)

const PolicyVersionV1 = "confidence-propagation-policy-v1-project-derived"

type Policy struct {
	Version                     string
	MaximumNodeCount            int
	LocalWeight                 float64
	DependencyWeight            float64
	RequiredDependencyAllowance float64
	EstimatedConfidenceCap      float64
	UnknownConfidenceCap        float64
	MediumThreshold             float64
	HighThreshold               float64
}

func DefaultPolicy() Policy {
	return Policy{
		Version:                     PolicyVersionV1,
		MaximumNodeCount:            200,
		LocalWeight:                 0.35,
		DependencyWeight:            0.65,
		RequiredDependencyAllowance: 0.10,
		EstimatedConfidenceCap:      0.70,
		UnknownConfidenceCap:        0.40,
		MediumThreshold:             0.55,
		HighThreshold:               0.80,
	}
}

func (policy Policy) Validate() error {
	if strings.TrimSpace(policy.Version) != PolicyVersionV1 ||
		policy.MaximumNodeCount < 1 ||
		policy.MaximumNodeCount > 10000 ||
		!unitInterval(policy.LocalWeight) ||
		!unitInterval(policy.DependencyWeight) ||
		math.Abs(policy.LocalWeight+policy.DependencyWeight-1) > 1e-9 ||
		!unitInterval(policy.RequiredDependencyAllowance) ||
		!unitInterval(policy.EstimatedConfidenceCap) ||
		!unitInterval(policy.UnknownConfidenceCap) ||
		policy.UnknownConfidenceCap > policy.EstimatedConfidenceCap ||
		!unitInterval(policy.MediumThreshold) ||
		!unitInterval(policy.HighThreshold) ||
		policy.MediumThreshold >= policy.HighThreshold {
		return fmt.Errorf("%w: thresholds", ErrInvalidPolicy)
	}
	return nil
}

func unitInterval(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}
