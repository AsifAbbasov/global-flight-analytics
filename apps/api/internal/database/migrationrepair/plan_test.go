package migrationrepair

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database/migrationfile"
)

func TestLoadPlanDerivesIdentityAndChecksumFromRepositoryFile(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	content := []byte("BEGIN;\nSELECT 10;\nCOMMIT;\n")
	path := filepath.Join(directory, DefaultRepairAnchorFileName)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	plan, err := LoadPlan(directory, "")
	if err != nil {
		t.Fatalf("LoadPlan() error = %v", err)
	}
	digest := sha256.Sum256(content)
	if plan.Anchor.FileName != DefaultRepairAnchorFileName ||
		plan.Anchor.Version != "010" ||
		plan.AnchorChecksum != hex.EncodeToString(digest[:]) {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}

func TestLoadPlanRejectsMissingDirectoryAndAnchor(t *testing.T) {
	t.Parallel()

	_, err := LoadPlan("   ", DefaultRepairAnchorFileName)
	if !errors.Is(err, ErrMigrationsDirectoryRequired) {
		t.Fatalf("expected migrations-directory error, got %v", err)
	}

	_, err = LoadPlan(t.TempDir(), DefaultRepairAnchorFileName)
	if !errors.Is(err, ErrRepairPlanInvalid) {
		t.Fatalf("expected repair-plan error, got %v", err)
	}
}

func TestPlanClassifiesEveryLaterCanonicalVersion(t *testing.T) {
	t.Parallel()

	plan := Plan{
		Anchor:         migrationfile.Identity{Version: "010", Name: "anchor", FileName: "010_anchor.sql"},
		AnchorChecksum: "checksum",
	}
	for _, version := range []string{"011", "012", "020", "999"} {
		if !plan.IsLaterVersion(version) {
			t.Fatalf("version %s was not classified as later", version)
		}
	}
	for _, version := range []string{"001", "009", "010"} {
		if plan.IsLaterVersion(version) {
			t.Fatalf("version %s was incorrectly classified as later", version)
		}
	}
}
