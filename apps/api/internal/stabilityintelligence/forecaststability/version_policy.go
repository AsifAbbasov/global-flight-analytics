package forecaststability

import (
	"fmt"
	"strings"
)

const VersionPolicyVersionV1 = "forecast-version-policy-v1"

type VersionPolicy struct {
	Version                     string
	MaximumProjectionPointCount int
	MaximumMetadataLength       int
	RequireSameTrajectory       bool
	RequireChronologicalParent  bool
}

func DefaultVersionPolicy() VersionPolicy {
	return VersionPolicy{
		Version:                     VersionPolicyVersionV1,
		MaximumProjectionPointCount: 240,
		MaximumMetadataLength:       200,
		RequireSameTrajectory:       true,
		RequireChronologicalParent:  true,
	}
}

func (policy VersionPolicy) Validate() error {
	if strings.TrimSpace(policy.Version) != VersionPolicyVersionV1 {
		return fmt.Errorf("%w: version", ErrInvalidVersionPolicy)
	}
	if policy.MaximumProjectionPointCount < 1 || policy.MaximumMetadataLength < 1 {
		return fmt.Errorf("%w: capacity", ErrInvalidVersionPolicy)
	}
	return nil
}
