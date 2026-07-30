package projectionfreshness

import "time"

// Policy is the immutable decision-policy snapshot published with every
// freshness result. It makes the result independently reviewable and allows
// Result.Validate to reconstruct the expected components and decision.
type Policy struct {
	MaximumNewestNeighborAge time.Duration
	MaximumMeanNeighborAge   time.Duration
	MaximumOldestNeighborAge time.Duration

	RecentNeighborAgeLimit     time.Duration
	MinimumRecentNeighborCount int
	TargetRecentNeighborCount  int

	MinimumUsableScore   float64
	CompleteScoreMinimum float64

	NewestAgeWeight     float64
	MeanAgeWeight       float64
	OldestAgeWeight     float64
	RecentSupportWeight float64
}

func (config Config) policySnapshot() Policy {
	return Policy{
		MaximumNewestNeighborAge: config.MaximumNewestNeighborAge,
		MaximumMeanNeighborAge:   config.MaximumMeanNeighborAge,
		MaximumOldestNeighborAge: config.MaximumOldestNeighborAge,

		RecentNeighborAgeLimit:     config.RecentNeighborAgeLimit,
		MinimumRecentNeighborCount: config.MinimumRecentNeighborCount,
		TargetRecentNeighborCount:  config.TargetRecentNeighborCount,

		MinimumUsableScore:   config.MinimumUsableScore,
		CompleteScoreMinimum: config.CompleteScoreMinimum,

		NewestAgeWeight:     config.NewestAgeWeight,
		MeanAgeWeight:       config.MeanAgeWeight,
		OldestAgeWeight:     config.OldestAgeWeight,
		RecentSupportWeight: config.RecentSupportWeight,
	}
}

func (policy Policy) config() Config {
	return Config{
		MaximumNewestNeighborAge: policy.MaximumNewestNeighborAge,
		MaximumMeanNeighborAge:   policy.MaximumMeanNeighborAge,
		MaximumOldestNeighborAge: policy.MaximumOldestNeighborAge,

		RecentNeighborAgeLimit:     policy.RecentNeighborAgeLimit,
		MinimumRecentNeighborCount: policy.MinimumRecentNeighborCount,
		TargetRecentNeighborCount:  policy.TargetRecentNeighborCount,

		MinimumUsableScore:   policy.MinimumUsableScore,
		CompleteScoreMinimum: policy.CompleteScoreMinimum,

		NewestAgeWeight:     policy.NewestAgeWeight,
		MeanAgeWeight:       policy.MeanAgeWeight,
		OldestAgeWeight:     policy.OldestAgeWeight,
		RecentSupportWeight: policy.RecentSupportWeight,
	}
}

func (policy Policy) Validate() error {
	return policy.config().Validate()
}
