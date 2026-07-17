package opensky

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/sourceconstraints"

func SourceProfile() sourceconstraints.SourceProfile {
	return sourceconstraints.OpenSkyProfile()
}

func EvaluateCapability(
	capability sourceconstraints.Capability,
) (sourceconstraints.Decision, error) {
	return sourceconstraints.Evaluate(sourceconstraints.Request{
		Constraints: sourceconstraints.FixedProjectConstraints(),
		Source:      SourceProfile(),
		Capability:  capability,
	})
}
