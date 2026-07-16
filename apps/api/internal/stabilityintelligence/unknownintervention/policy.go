package unknownintervention

import (
	"fmt"
	"math"
)

const PolicyVersionV1 = "unknown-intervention-policy-v1"

type Policy struct {
	Version                        string
	MaximumEvidenceCount           int
	AllowedConfidenceMinimum       float64
	LimitedConfidenceMinimum       float64
	RequiredEvidenceMinimum        float64
	AllowedCompletenessMinimum     float64
	LimitedCompletenessMinimum     float64
	EstimatedEvidenceConfidenceCap float64
	UnknownEvidenceConfidenceCap   float64
}

func DefaultPolicy() Policy {
	return Policy{Version: PolicyVersionV1, MaximumEvidenceCount: 100, AllowedConfidenceMinimum: 0.75, LimitedConfidenceMinimum: 0.45, RequiredEvidenceMinimum: 0.60, AllowedCompletenessMinimum: 0.80, LimitedCompletenessMinimum: 0.50, EstimatedEvidenceConfidenceCap: 0.65, UnknownEvidenceConfidenceCap: 0.35}
}
func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 || policy.MaximumEvidenceCount <= 0 ||
		!unit(policy.AllowedConfidenceMinimum) || !unit(policy.LimitedConfidenceMinimum) || policy.AllowedConfidenceMinimum <= policy.LimitedConfidenceMinimum ||
		!unit(policy.RequiredEvidenceMinimum) || !unit(policy.AllowedCompletenessMinimum) || !unit(policy.LimitedCompletenessMinimum) || policy.AllowedCompletenessMinimum <= policy.LimitedCompletenessMinimum ||
		!unit(policy.EstimatedEvidenceConfidenceCap) || !unit(policy.UnknownEvidenceConfidenceCap) || policy.EstimatedEvidenceConfidenceCap <= policy.UnknownEvidenceConfidenceCap {
		return fmt.Errorf("invalid unknown intervention policy")
	}
	return nil
}
func unit(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}
