package projectionread

import "testing"

func TestDefaultPolicyIsInternallyConsistent(
	t *testing.T,
) {
	policy := DefaultPolicy()

	if err := policy.Validate(); err != nil {
		t.Fatalf(
			"DefaultPolicy().Validate() error = %v",
			err,
		)
	}
	if policy.DataSource.
		MaximumHistoricalCandidateCount !=
		policy.Neighbors.MaximumCandidateCount {
		t.Fatalf(
			"data-source and selector candidate limits differ: source=%d selector=%d",
			policy.DataSource.
				MaximumHistoricalCandidateCount,
			policy.Neighbors.
				MaximumCandidateCount,
		)
	}
	if policy.DataSource.
		HistoricalCandidateLookback !=
		policy.Neighbors.MaximumCandidateAge {
		t.Fatal(
			"candidate lookback and selector maximum age differ",
		)
	}
	if policy.DataSource.RecentRouteWindow !=
		policy.RouteFrequency.RecentWindow {
		t.Fatal(
			"route-history and route-frequency recent windows differ",
		)
	}
}

func TestDefaultPolicyBuildsAllAlgorithmComponents(
	t *testing.T,
) {
	components, err :=
		buildAlgorithmComponents(
			DefaultPolicy(),
		)
	if err != nil {
		t.Fatalf(
			"buildAlgorithmComponents() error = %v",
			err,
		)
	}
	if components.composer == nil {
		t.Fatal(
			"production composer was not constructed",
		)
	}
}
