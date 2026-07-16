package forecaststability

import (
	"testing"
	"time"
)

func TestRegisterVersionInitialAndIdenticalReplay(t *testing.T) {
	projection := testProjection()
	registeredAt := projection.GeneratedAt.Add(time.Second)
	initial, err := RegisterVersion(ForecastVersionRequest{
		Projection:            projection,
		PolicyVersion:         "projection-production-policy-v1",
		ImplementationVersion: "build-001",
		RegisteredAt:          registeredAt,
	}, DefaultVersionPolicy())
	if err != nil {
		t.Fatalf("initial registration: %v", err)
	}
	if initial.Decision != RegistrationDecisionInitial || initial.Record.Ordinal != 1 {
		t.Fatalf("initial result = %#v", initial)
	}

	replay, err := RegisterVersion(ForecastVersionRequest{
		Projection:            projection,
		PolicyVersion:         "projection-production-policy-v1",
		ImplementationVersion: "build-001",
		Previous:              &initial.Record,
		RegisteredAt:          registeredAt.Add(time.Second),
	}, DefaultVersionPolicy())
	if err != nil {
		t.Fatalf("replay registration: %v", err)
	}
	if replay.Decision != RegistrationDecisionReused || replay.Record.VersionID != initial.Record.VersionID || replay.Record.Ordinal != 1 {
		t.Fatalf("replay result = %#v", replay)
	}
}

func TestRegisterVersionCreatesSuccessorAndClassifiesChanges(t *testing.T) {
	projection := testProjection()
	initial, err := RegisterVersion(ForecastVersionRequest{
		Projection:            projection,
		PolicyVersion:         "projection-production-policy-v1",
		ImplementationVersion: "build-001",
		RegisteredAt:          projection.GeneratedAt.Add(time.Second),
	}, DefaultVersionPolicy())
	if err != nil {
		t.Fatal(err)
	}
	changed := projection.Clone()
	changed.Provenance.InputFingerprint = fingerprintOf("input-2")
	changed.Points[0].Position.Longitude += 0.10
	changed.GeneratedAt = projection.GeneratedAt.Add(time.Minute)
	successor, err := RegisterVersion(ForecastVersionRequest{
		Projection:            changed,
		PolicyVersion:         "projection-production-policy-v2",
		ImplementationVersion: "build-002",
		Previous:              &initial.Record,
		RegisteredAt:          changed.GeneratedAt.Add(time.Second),
	}, DefaultVersionPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if successor.Decision != RegistrationDecisionCreated || successor.Record.Ordinal != 2 || successor.Record.ParentVersionID != initial.Record.VersionID {
		t.Fatalf("successor = %#v", successor)
	}
	kinds := map[VersionChangeKind]bool{}
	for _, change := range successor.Changes {
		kinds[change.Kind] = true
	}
	for _, expected := range []VersionChangeKind{VersionChangePolicy, VersionChangeImplementation, VersionChangeInput, VersionChangeOutput} {
		if !kinds[expected] {
			t.Fatalf("missing change %q in %#v", expected, successor.Changes)
		}
	}
}

func TestProjectionFingerprintIgnoresNarrativeOrdering(t *testing.T) {
	left := testProjection()
	right := left.Clone()
	right.Limitations[0], right.Limitations[1] = right.Limitations[1], right.Limitations[0]
	right.Explanations[0], right.Explanations[1] = right.Explanations[1], right.Explanations[0]
	right.Provenance.Inputs[0], right.Provenance.Inputs[1] = right.Provenance.Inputs[1], right.Provenance.Inputs[0]
	if projectionOutputFingerprint(left) != projectionOutputFingerprint(right) {
		t.Fatal("canonical fingerprint changed under reorder")
	}
}
