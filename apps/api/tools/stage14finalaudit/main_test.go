package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStage14DocumentRegisterIsContiguous(t *testing.T) {
	if len(stage14Documents) != 32 {
		t.Fatalf("document count = %d, want 32", len(stage14Documents))
	}
	for index, fileName := range stage14Documents {
		expected := index + 41
		prefix := strings.SplitN(fileName, "_", 2)[0]
		if prefix != twoDigitDocumentNumber(expected) {
			t.Fatalf("document %d has file %q", expected, fileName)
		}
	}
}

func TestAuditRepositoryAcceptsCompleteFixture(t *testing.T) {
	root := createCompleteFixture(t)
	failures := auditRepository(root, &bytes.Buffer{})
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %#v", failures)
	}
}

func TestAuditRepositoryDetectsMissingFeatureStoreIntegration(t *testing.T) {
	root := createCompleteFixture(t)
	workflowPath := filepath.Join(root, ".github", "workflows", "backend-ci.yml")
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.ReplaceAll(
		string(content),
		"./internal/features/featurestore",
		"./internal/features/omitted",
	)
	if err := os.WriteFile(workflowPath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "./internal/features/featurestore") {
		t.Fatalf("missing Feature Store integration was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsAmbiguousTrajectoryPointAggregate(t *testing.T) {
	root := createCompleteFixture(t)
	migrationPath := filepath.Join(
		root,
		"database",
		"migrations",
		"018_trajectory_relational_integrity.sql",
	)
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.ReplaceAll(
		string(content),
		"SUM(segment.point_count)",
		"SUM(point_count)",
	)
	if err := os.WriteFile(migrationPath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "SUM(segment.point_count)") {
		t.Fatalf("ambiguous trajectory aggregate was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsIncompleteFlightStateFixture(t *testing.T) {
	root := createCompleteFixture(t)
	fixturePath := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"repository",
		"postgres",
		"flightstate_altitude_integration_test.go",
	)
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.ReplaceAll(string(content), "squawk_code text", "omitted_evidence text")
	if err := os.WriteFile(fixturePath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "squawk_code text") {
		t.Fatalf("incomplete Flight State fixture was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsTerminalIngestionRunFixtureWithoutFinishTime(t *testing.T) {
	root := createCompleteFixture(t)
	fixturePath := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"repository",
		"postgres",
		"reconciliation_result_identity_integration_test.go",
	)
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.ReplaceAll(
		string(content),
		"started_at, finished_at, status",
		"started_at, status",
	)
	updated = strings.ReplaceAll(
		updated,
		"VALUES ($1, $2, $2, 'success')",
		"VALUES ($1, $2, 'success')",
	)
	if err := os.WriteFile(fixturePath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "without finished_at") {
		t.Fatalf("invalid terminal ingestion-run fixture was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsAnyIncompleteFlightStateRepositoryFixture(t *testing.T) {
	root := createCompleteFixture(t)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/additional_integration_test.go",
		"package postgres\n// CREATE TABLE flight_states (\n// squawk_code\n// special_purpose_indicator\n// position_source\n// aircraft_category\n// missing final evidence column\n// NewFlightStateRepository(\n",
	)

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "aircraft_category_available") {
		t.Fatalf("incomplete additional FlightStateRepository fixture was not detected: %#v", failures)
	}
}

func TestAuditRepositoryAllowsPurposeBuiltMinimalFlightStateFixture(t *testing.T) {
	root := createCompleteFixture(t)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/minimal_metric_integration_test.go",
		"package postgres\n// CREATE TABLE flight_states (id uuid, icao24 text);\n// NewMetricsRepository(\n",
	)

	failures := auditRepository(root, &bytes.Buffer{})
	if containsFailureCheck(failures, "Flight State repository fixture parity") {
		t.Fatalf("purpose-built minimal fixture was incorrectly rejected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsRetiredDataQualityMigrationReference(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"repository",
		"postgres",
		"data_quality_parent_integrity_test.go",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.ReplaceAll(
		string(content),
		"019_data_quality_parent_integrity.sql",
		"016_data_quality_parent_integrity.sql",
	)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "016_data_quality_parent_integrity.sql") {
		t.Fatalf("retired migration reference was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsMissingReopenedStatusSurface(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(root, "README.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.ReplaceAll(
		string(content),
		"Stage 14 remains reopened",
		"Stage 14 is closed",
	)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "Stage 14 remains reopened") {
		t.Fatalf("missing reopened status was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsDuplicateMigrationVersion(t *testing.T) {
	root := createCompleteFixture(t)
	writeFixtureFile(
		t,
		root,
		"database/migrations/016_duplicate.sql",
		"BEGIN;\nSELECT 1;\nCOMMIT;\n",
	)

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureCheck(failures, "Migration catalog unique versions") {
		t.Fatalf("duplicate migration version was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsOutdatedGoToolchain(t *testing.T) {
	root := createCompleteFixture(t)
	goModPath := filepath.Join(root, "apps", "api", "go.mod")
	if err := os.WriteFile(
		goModPath,
		[]byte("module example.com/stage14\n\ngo 1.26.2\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "go 1.26.5") {
		t.Fatalf("outdated Go toolchain was not detected: %#v", failures)
	}
}

func TestRunReturnsFailureForMissingDocument(t *testing.T) {
	root := createCompleteFixture(t)
	missing := filepath.Join(root, "docs", stage14Documents[len(stage14Documents)-1])
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{"-root", root, "-strict"},
		&stdout,
		&stderr,
	)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if !strings.Contains(stderr.String(), "missing regular file") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func createCompleteFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFixtureFile(t, root, "apps/api/go.mod", "module example.com/stage14\n\ngo 1.26.5\n")
	writeFixtureFile(t, root, "apps/api/Dockerfile", "ARG GO_IMAGE=golang:1.26.5-alpine3.24\n")
	writeFixtureFile(t, root, "apps/web/package.json", "{}\n")
	writeFixtureFile(
		t,
		root,
		"README.md",
		"<!-- STAGE-14-29-MIGRATION-CATALOG-INTEGRITY:README -->\nStage 14 remains reopened\n",
	)
	writeFixtureFile(
		t,
		root,
		"docs/25_IMPLEMENTATION_SEQUENCE.md",
		"<!-- STAGE-14-29-MIGRATION-CATALOG-INTEGRITY:IMPLEMENTATION -->\nStage 14 remains reopened\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/data_quality_parent_integrity_test.go",
		"package postgres\n// database/migrations/019_data_quality_parent_integrity.sql\n// migration 019\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/data_quality_parent_integrity_integration_test.go",
		"package postgres\n// database/migrations/019_data_quality_parent_integrity.sql\n",
	)

	var index strings.Builder
	for _, fileName := range stage14Documents {
		content := "Stage 14 document\n"
		switch fileName {
		case "56_BACKEND_FINAL_CORRECTNESS_AUDIT.md":
			content += "scripts/verify-stage-14-completion.sh\nSTAGE_14_CURRENT_SCOPE_AUDIT=PASS\n"
		case "60_STAGE_14_19_DATA_QUALITY_PARENT_INTEGRITY.md":
			content += "Migration 019\n019_data_quality_parent_integrity.sql\n"
		case "70_STAGE_14_FINAL_COMPLETION_AUDIT.md":
			content += "STAGE_14_CURRENT_SCOPE_AUDIT=PASS\nFlight Feature timestamp integration\nbackend container\nfrontend production build\n"
		}
		writeFixtureFile(t, root, filepath.Join("docs", fileName), content)
		index.WriteString(fileName)
		index.WriteByte('\n')
	}
	writeFixtureFile(t, root, "docs/DOCUMENT_INDEX.md", index.String())

	writeFixtureFile(
		t,
		root,
		"scripts/verify-stage-14-completion.sh",
		strings.Join([]string{
			"GOTOOLCHAIN=go1.26.5+auto",
			"go env GOVERSION",
			"STAGE_14_GO_TOOLCHAIN_AUDIT=PASS",
			"scripts/verify-backend-final-correctness.sh",
			"go run ./tools/stage14finalaudit -strict",
			"./internal/repository/postgres",
			"./internal/features/featurestore",
			"./internal/routeintelligence/routestore",
			"./internal/historicalintelligence/historicalaggregate",
			"go run ./cmd/migrate",
			"STAGE_14_PRODUCTION_MIGRATOR=PASS",
			"go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...",
			"pnpm run test:web-dependency-policy",
			"pnpm run verify:web-dependencies",
			"pnpm audit --prod --audit-level moderate",
			"pnpm --dir apps/web lint",
			"pnpm --dir apps/web typecheck",
			"pnpm --dir apps/web build",
			"docker compose",
			"docker build",
			"docker image inspect",
			"docker run",
			"git diff --check",
			"STAGE_14_CURRENT_SCOPE_AUDIT=PASS",
		}, "\n")+"\n",
	)
	writeFixtureFile(t, root, "package.json", `{\n  "scripts": {\n    "verify:stage14": "bash scripts/verify-stage-14-completion.sh"\n  }\n}`+"\n")
	writeFixtureFile(
		t,
		root,
		".github/workflows/backend-ci.yml",
		strings.Join([]string{
			"go-version-file: apps/api/go.mod",
			"go run ./tools/stage14finalaudit -strict",
			"./internal/repository/postgres",
			"./internal/features/featurestore",
			"./internal/routeintelligence/routestore",
			"./internal/historicalintelligence/historicalaggregate",
			"go run ./cmd/migrate",
			"MIGRATIONS_DIR",
			"go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...",
			"docker build",
			"Run container health smoke test",
		}, "\n")+"\n",
	)
	writeFixtureFile(
		t,
		root,
		".github/workflows/frontend-ci.yml",
		strings.Join([]string{
			"pnpm run test:web-dependency-policy",
			"pnpm run verify:web-dependencies",
			"pnpm audit --prod --audit-level moderate",
			"pnpm --filter web lint",
			"pnpm --filter web typecheck",
			"pnpm --filter web build",
		}, "\n")+"\n",
	)

	writeFixtureFile(t, root, "apps/api/internal/repository/postgres/trajectory_repository.go", "package postgres\ntype TrajectoryRepository struct{}\n")
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/trajectory_read_repository.go",
		"package postgres\n// withTrajectoryReadSnapshot(\n// snapshotRepository.getLatestTrajectoryByICAO24\n// snapshotRepository.getTrajectoryByID\n",
	)
	for _, path := range trajectoryOwnerFiles {
		writeFixtureFile(t, root, path, "package postgres\n")
	}
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/trajectory_repository_decomposition_test.go",
		"package postgres\n// TestTrajectoryRepositoryCoordinatorsRemainNarrow\n// TestTrajectoryRepositoryResponsibilitiesHaveDedicatedOwners\n// TestTrajectoryWriteValidationStillPrecedesTransaction\n",
	)
	writeFixtureFile(
		t,
		root,
		"database/migrations/018_trajectory_relational_integrity.sql",
		"SUM(segment.point_count)\nMIN(segment.sequence_number)\nMAX(segment.sequence_number)\nFROM trajectory_segments AS segment\nWHERE segment.trajectory_id = target_trajectory_id\n",
	)
	writeFixtureFile(
		t,
		root,
		"database/migrations/016_add_flight_state_observation_metadata.sql",
		"BEGIN;\nSELECT 16;\nCOMMIT;\n",
	)
	writeFixtureFile(
		t,
		root,
		"database/migrations/019_data_quality_parent_integrity.sql",
		"BEGIN;\nSELECT 19;\nCOMMIT;\n",
	)
	writeFixtureFile(
		t,
		root,
		"database/migrations/020_stage14_correctness_hardening.sql",
		"ingestion_runs_processed_counts_check\ningestion_runs_error_message_status_check\nflight_route_results_as_of_time_mirror_check\nflight_route_results_stored_at_mirror_check\nhistorical_aggregate_results_window_start_mirror_check\nhistorical_aggregate_results_window_end_mirror_check\nhistorical_aggregate_results_as_of_time_mirror_check\nhistorical_aggregate_results_stored_at_mirror_check\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/flightstate_altitude_integration_test.go",
		"package postgres\n// CREATE TABLE flight_states (\n// squawk_code text\n// special_purpose_indicator boolean\n// position_source text\n// aircraft_category smallint\n// aircraft_category_available boolean\n// NewFlightStateRepository(\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/flightstate_reconciliation_repository_integration_test.go",
		"package postgres\n// CREATE TABLE flight_states (\n// squawk_code text\n// special_purpose_indicator boolean\n// position_source text\n// aircraft_category smallint\n// aircraft_category_available boolean\n// NewFlightStateRepository(\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/reconciliation_result_identity_integration_test.go",
		"package postgres\n// INSERT INTO ingestion_runs (id, started_at, finished_at, status) VALUES ($1, $2, $2, 'success');\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/features/featurestore/timestamp_consistency.go",
		"package featurestore\n// postgresTimestampMirrorTolerance = time.Microsecond\n// func validateTimestampMirror(\n// delta <= -postgresTimestampMirrorTolerance\n// delta >= postgresTimestampMirrorTolerance\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/ingestionrun_completion_validation.go",
		"package postgres\n// recordsUpdated > recordsReceived-recordsInserted\n// ingestionrun.StatusSuccess\n// ingestionrun.StatusFailed, ingestionrun.StatusPartial\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/routeintelligence/routestore/timestamp_consistency.go",
		"package routestore\n// postgresTimestampMirrorTolerance = time.Microsecond\n// func validateTimestampMirror(\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/historicalintelligence/historicalaggregate/timestamp_consistency.go",
		"package historicalaggregate\n// postgresTimestampMirrorTolerance = time.Microsecond\n// func validateTimestampMirror(\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/transaction_rollback.go",
		"package postgres\n// context.WithTimeout(\n// context.Background()\n// repositoryRollbackTimeout\n",
	)
	for _, path := range []string{
		"apps/api/internal/repository/postgres/airport_import_repository.go",
		"apps/api/internal/repository/postgres/flightstate_repository.go",
		"apps/api/internal/repository/postgres/trajectory_write_repository.go",
	} {
		writeFixtureFile(t, root, path, "package postgres\n// rollbackRepositoryTransaction(tx)\n")
	}
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/stage14_correctness_constraints_integration_test.go",
		"package postgres\n// migrator.NewRunner\n// runner.ApplyPending\n// assertStage14CorrectnessPostgresCode(t, err, \"23514\")\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/domain/traffic/altitude.go",
		"package traffic\n// func ResolveCurrentAltitude(\n// AltitudeSourceGeometric\n// AltitudeSourceBarometric\n// AltitudeSourceGround\n// return nil,\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/domain/airport/elevation.go",
		"package airport\n// ElevationStatusObserved\n// ElevationStatusUnknown\n// ElevationStatusInvalid\n// func ResolveElevation(\n",
	)
	return root
}

func writeFixtureFile(t *testing.T, root string, relativePath string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func containsFailureDetail(failures []auditFailure, fragment string) bool {
	for _, failure := range failures {
		if strings.Contains(failure.Detail, fragment) {
			return true
		}
	}
	return false
}

func containsFailureCheck(failures []auditFailure, check string) bool {
	for _, failure := range failures {
		if failure.Check == check {
			return true
		}
	}
	return false
}

func twoDigitDocumentNumber(value int) string {
	if value < 10 {
		return "0" + string(rune('0'+value))
	}
	return string([]byte{byte('0' + value/10), byte('0' + value%10)})
}
