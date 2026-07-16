package airspaceregionanalytics

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
)

func TestValidateRejectsTamperedFingerprint(t *testing.T) {
	start := time.Date(2026, 7, 17, 16, 0, 0, 0, time.UTC)
	result, err := Build(testRequest(start, testSnapshot(start.Add(10*time.Second), []localtrafficscene.Aircraft{
		testAircraft("A", 40.1, 49.1, float64Pointer(9000), 90, 220, 0.9),
	}, nil, nil, "validation")), DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	result.Provenance.InputFingerprint = "tampered"
	report := Validate(result, DefaultPolicy())
	if report.Status != ValidationStatusInvalid {
		t.Fatalf("Status = %q, want %q", report.Status, ValidationStatusInvalid)
	}
}

func TestPolicyValidateRejectsWeightDrift(t *testing.T) {
	policy := DefaultPolicy()
	policy.ComplexityWeights.Density += 0.1
	if err := policy.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want invalid complexity weights")
	}
}

func TestBuildRejectsBrokenProvenanceChain(t *testing.T) {
	start := time.Date(2026, 7, 17, 17, 0, 0, 0, time.UTC)
	snapshot := testSnapshot(start.Add(10*time.Second), []localtrafficscene.Aircraft{
		testAircraft("A", 40.1, 49.1, float64Pointer(9000), 90, 220, 0.9),
	}, nil, nil, "broken")
	snapshot.Risk.Provenance.ScanFingerprint = "wrong"
	_, err := Build(testRequest(start, snapshot), DefaultPolicy())
	if err == nil {
		t.Fatal("Build() error = nil, want broken provenance chain")
	}
}
