package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDataQualityRepositoryRequiresPersistedParentForCanonicalReport(
	t *testing.T,
) {
	t.Parallel()

	source := readDataQualityRepositorySource(t)

	requiredTokens := []string{
		"ErrDataQualityFlightStateNotPersisted",
		"INSERT INTO data_quality_reports",
		"FROM flight_states AS persisted_state",
		"WHERE persisted_state.id = $1",
		"RETURNING id::text",
		"errors.Is(err, pgx.ErrNoRows)",
	}

	for _, token := range requiredTokens {
		if !strings.Contains(source, token) {
			t.Fatalf("data quality repository is missing parent-integrity token %q", token)
		}
	}
}

func TestDataQualityRepositorySeparatesRejectedObservationEvidence(
	t *testing.T,
) {
	t.Parallel()

	source := readDataQualityRepositorySource(t)

	requiredTokens := []string{
		"quality.ValidationStatus == dataquality.ValidationStatusInvalid",
		"insertRejectedFlightStateQuality",
		"INSERT INTO rejected_flight_state_quality_reports",
	}

	for _, token := range requiredTokens {
		if !strings.Contains(source, token) {
			t.Fatalf("data quality repository is missing rejected-evidence token %q", token)
		}
	}
}

func TestDataQualityParentIntegrityMigrationClosesOrphanPath(
	t *testing.T,
) {
	t.Parallel()

	migration := readDataQualityParentIntegrityMigration(t)

	requiredTokens := []string{
		"CREATE TABLE rejected_flight_state_quality_reports",
		"WHERE flight_state_id IS NULL",
		"DELETE FROM data_quality_reports",
		"ALTER COLUMN state_id SET NOT NULL",
		"ALTER COLUMN flight_state_id SET NOT NULL",
		"ON DELETE CASCADE",
		"CHECK (state_id = flight_state_id)",
	}

	for _, token := range requiredTokens {
		if !strings.Contains(migration, token) {
			t.Fatalf("migration 019 is missing parent-integrity token %q", token)
		}
	}
}

func readDataQualityRepositorySource(t *testing.T) string {
	t.Helper()

	content, err := os.ReadFile("data_quality_repository.go")
	if err != nil {
		t.Fatalf("read data_quality_repository.go: %v", err)
	}

	return string(content)
}

func readDataQualityParentIntegrityMigration(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve data quality parent integrity test path")
	}

	migrationPath := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"../../../../../database/migrations/019_data_quality_parent_integrity.sql",
		),
	)

	content, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration 019: %v", err)
	}

	return string(content)
}
