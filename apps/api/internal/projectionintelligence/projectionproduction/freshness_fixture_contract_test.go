package projectionproduction

import "testing"

func TestProductionFreshnessFixtureMatchesHardenedContracts(t *testing.T) {
	fixture := newProductionFixture()

	if err := fixture.selector.result.Validate(); err != nil {
		t.Fatalf("selection fixture validation error = %v", err)
	}
	if fixture.pattern.result.SourceSelectionFingerprint !=
		fixture.selector.result.InputFingerprint {
		t.Fatalf(
			"pattern source selection fingerprint = %q, want %q",
			fixture.pattern.result.SourceSelectionFingerprint,
			fixture.selector.result.InputFingerprint,
		)
	}
	if err := fixture.pattern.result.Validate(); err != nil {
		t.Fatalf("pattern fixture validation error = %v", err)
	}
	if fixture.freshness.result.SourceSelectionFingerprint !=
		fixture.selector.result.InputFingerprint {
		t.Fatalf(
			"freshness source selection fingerprint = %q, want %q",
			fixture.freshness.result.SourceSelectionFingerprint,
			fixture.selector.result.InputFingerprint,
		)
	}
	if fixture.freshness.result.SourcePatternFingerprint !=
		fixture.pattern.result.InputFingerprint {
		t.Fatalf(
			"freshness source pattern fingerprint = %q, want %q",
			fixture.freshness.result.SourcePatternFingerprint,
			fixture.pattern.result.InputFingerprint,
		)
	}
	if err := fixture.freshness.result.Policy.Validate(); err != nil {
		t.Fatalf("freshness policy fixture validation error = %v", err)
	}
	if err := fixture.freshness.result.Validate(); err != nil {
		t.Fatalf("freshness fixture validation error = %v", err)
	}
}
