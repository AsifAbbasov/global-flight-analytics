package etareliability

import (
	"fmt"
	"time"
)

type Policy struct {
	MaximumCandidateCount int
	MinimumSampleCount int
	CompleteSampleCount int
	MinimumLead time.Duration
	MaximumLead time.Duration
	LeadTolerance time.Duration
	EndpointRadiusKM float64
	MinimumPrefixPointCount int
}

func DefaultPolicy() Policy {
	return Policy{
		MaximumCandidateCount: 8,
		MinimumSampleCount: 3,
		CompleteSampleCount: 6,
		MinimumLead: 5 * time.Minute,
		MaximumLead: 90 * time.Minute,
		LeadTolerance: 10 * time.Minute,
		EndpointRadiusKM: 15,
		MinimumPrefixPointCount: 5,
	}
}

func (policy Policy) Validate() error {
	if policy.MaximumCandidateCount < 1 {
		return fmt.Errorf("maximum ETA reliability candidate count must be positive")
	}
	if policy.MinimumSampleCount < 1 || policy.MinimumSampleCount > policy.MaximumCandidateCount {
		return fmt.Errorf("minimum ETA reliability sample count is invalid")
	}
	if policy.CompleteSampleCount < policy.MinimumSampleCount || policy.CompleteSampleCount > policy.MaximumCandidateCount {
		return fmt.Errorf("complete ETA reliability sample count is invalid")
	}
	if policy.MinimumLead <= 0 || policy.MaximumLead < policy.MinimumLead || policy.LeadTolerance <= 0 {
		return fmt.Errorf("ETA reliability lead-time policy is invalid")
	}
	if !finitePositive(policy.EndpointRadiusKM) {
		return fmt.Errorf("ETA reliability endpoint radius is invalid")
	}
	if policy.MinimumPrefixPointCount < 2 {
		return fmt.Errorf("ETA reliability minimum prefix point count is invalid")
	}
	return nil
}
