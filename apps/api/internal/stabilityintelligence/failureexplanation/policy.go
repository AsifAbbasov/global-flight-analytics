package failureexplanation

import (
	"fmt"
	"math"
)

const PolicyVersionV1 = "failure-explanation-policy-v1"

type Policy struct {
	Version                   string
	MaximumSignalCount        int
	CompleteConfidenceMinimum float64
	SeverityWeightInformation float64
	SeverityWeightWarning     float64
	SeverityWeightBlocking    float64
	UnknownCausePriorityBoost float64
	BlockingPriorityBoost     float64
}

func DefaultPolicy() Policy {
	return Policy{
		Version:                   PolicyVersionV1,
		MaximumSignalCount:        100,
		CompleteConfidenceMinimum: 0.60,
		SeverityWeightInformation: 0.20,
		SeverityWeightWarning:     0.60,
		SeverityWeightBlocking:    0.90,
		UnknownCausePriorityBoost: 0.05,
		BlockingPriorityBoost:     0.05,
	}
}

func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 ||
		policy.MaximumSignalCount <= 0 ||
		!unitInterval(policy.CompleteConfidenceMinimum) ||
		!unitInterval(policy.SeverityWeightInformation) ||
		!unitInterval(policy.SeverityWeightWarning) ||
		!unitInterval(policy.SeverityWeightBlocking) ||
		!unitInterval(policy.UnknownCausePriorityBoost) ||
		!unitInterval(policy.BlockingPriorityBoost) ||
		policy.SeverityWeightInformation >= policy.SeverityWeightWarning ||
		policy.SeverityWeightWarning >= policy.SeverityWeightBlocking {
		return fmt.Errorf("invalid failure explanation policy")
	}
	return nil
}

func unitInterval(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}
