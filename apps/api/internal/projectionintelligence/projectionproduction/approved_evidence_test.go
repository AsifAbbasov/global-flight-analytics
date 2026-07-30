package projectionproduction

import "testing"

func TestComposePassesAuthorizedEvidenceToHistoricalProjector(
	t *testing.T,
) {
	fixture := newProductionFixture()
	composer, err := New(fixture.config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = composer.Compose(fixture.request)
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	if fixture.historical.calls != 1 {
		t.Fatalf(
			"historical calls = %d, want 1",
			fixture.historical.calls,
		)
	}
	if fixture.historical.evidence.
		Selection.InputFingerprint !=
		fixture.selector.result.
			InputFingerprint {
		t.Fatal(
			"historical projector received another selection",
		)
	}
	if fixture.historical.evidence.
		Pattern.InputFingerprint !=
		fixture.pattern.result.
			InputFingerprint ||
		fixture.historical.evidence.
			Pattern.
			SourceSelectionFingerprint !=
			fixture.historical.evidence.
				Selection.
				InputFingerprint {
		t.Fatal(
			"historical projector received another pattern lineage",
		)
	}
	if fixture.selector.calls != 1 ||
		fixture.pattern.calls != 1 {
		t.Fatalf(
			"production evidence calls: selector=%d pattern=%d",
			fixture.selector.calls,
			fixture.pattern.calls,
		)
	}
}
