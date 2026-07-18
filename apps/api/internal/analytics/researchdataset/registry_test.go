package researchdataset

import (
	"errors"
	"testing"
	"time"
)

func TestADSCDatasetIsBlockedByFixedSourceBoundary(t *testing.T) {
	profile, err := ProfileByID(IDADSC)
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if profile.Selection != SelectionBlocked {
		t.Fatalf("selection = %s", profile.Selection)
	}

	decision, err := EvaluateSourceBoundary(IDADSC)
	if err != nil {
		t.Fatalf("evaluate boundary: %v", err)
	}
	if decision.Usable() {
		t.Fatalf("decision = %#v, want blocked", decision)
	}
}

func TestManifestRejectsBlockedADSCTable(t *testing.T) {
	manifest := validManifest(IDTrinoSnapshot2026)
	manifest.SelectedTables = []string{
		"state_vectors_data4",
		"readsb_adsc_sv",
	}

	_, err := ValidateManifest(manifest)
	if !errors.Is(err, ErrBlockedTable) {
		t.Fatalf("error = %v, want %v", err, ErrBlockedTable)
	}
}

func TestBoundedWeeklyStateVectorManifestIsAllowed(t *testing.T) {
	manifest := validManifest(IDWeeklyStateVectors)
	decision, err := ValidateManifest(manifest)
	if err != nil {
		t.Fatalf("validate manifest: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("decision = %#v", decision)
	}
}

func validManifest(id ID) Manifest {
	return Manifest{
		DatasetID: id,
		Version:   "test-v1",
		Files: []File{
			{
				Name:      "sample.avro",
				Format:    "avro",
				SizeBytes: 1024,
				SHA256:    "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
		},
		TotalBytes:           1024,
		MaximumRecords:       100,
		RegionFilter:         "AZ,GE,AM,TR",
		OfflineOnly:          true,
		ProductionDependency: false,
		LicenseReviewed:      true,
		AttributionProvided:  true,
		PreparedAt:           time.Now().UTC(),
	}
}
