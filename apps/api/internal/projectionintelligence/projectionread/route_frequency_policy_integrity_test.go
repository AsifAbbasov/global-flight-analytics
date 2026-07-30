package projectionread

import "testing"

func TestPolicyValidateRequiresRouteFrequencyHistoryWindowAlignment(t *testing.T) {
	policy := DefaultPolicy()
	policy.RouteFrequency.HistoryWindow--
	if err := policy.Validate(); err == nil {
		t.Fatal("Validate() accepted mismatched route-frequency history window")
	}
}
