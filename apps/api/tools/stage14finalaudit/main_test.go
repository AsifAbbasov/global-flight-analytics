package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStage14DocumentRegisterIsContiguous(t *testing.T) {
	if len(stage14Documents) != 38 {
		t.Fatalf("document count = %d, want 38", len(stage14Documents))
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

func TestAuditRepositoryDetectsMissingClosedStatusSurface(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(root, "README.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.ReplaceAll(
		string(content),
		"Stage 14 is closed",
		"Stage 14 closure is missing",
	)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "Stage 14 is closed") {
		t.Fatalf("missing closed status was not detected: %#v", failures)
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

func TestAuditRepositoryDetectsFlightStateWriteResponsibilityLeak(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"repository",
		"postgres",
		"flightstate_repository.go",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, []byte("// INSERT INTO flight_states\n")...)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "INSERT INTO flight_states") {
		t.Fatalf("Flight State write responsibility leak was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsRetiredAirportElevationSourceOwner(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"repository",
		"postgres",
		"airport_elevation_semantics_test.go",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.ReplaceAll(
		string(content),
		"airport_read_queries.go",
		"airport_repository.go",
	)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "airport_read_queries.go") {
		t.Fatalf("missing canonical Airport query owner was not detected: %#v", failures)
	}
	if !containsFailureDetail(failures, "airport_repository.go") {
		t.Fatalf("retired Airport elevation source owner was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsAirportOffsetPagination(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"repository",
		"postgres",
		"airport_read_queries.go",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, []byte("// OFFSET $4\n")...)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "OFFSET") {
		t.Fatalf("Airport offset pagination was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsInventedRepositoryContext(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"repository",
		"postgres",
		"airport_repository.go",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, []byte("// ctx = context.Background()\n")...)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "ctx = context.Background()") {
		t.Fatalf("invented repository context was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsImplicitTrajectoryWriteMode(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"repository",
		"postgres",
		"trajectory_write_repository.go",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, []byte("// reconciliationTaskID == \"\"\n")...)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureCheck(
		failures,
		"Trajectory writes preserve caller-owned context and explicit mode",
	) {
		t.Fatalf("implicit Trajectory write mode audit rule was not triggered: %#v", failures)
	}
}

func TestAuditRepositoryDetectsArtificialUnknownSourceFallback(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"repository",
		"postgres",
		"repository_helpers.go",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, []byte("// return \"unknown\"\n")...)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureCheck(
		failures,
		"Repository arguments preserve nullable and required evidence semantics",
	) {
		t.Fatalf("artificial source fallback audit rule was not triggered: %#v", failures)
	}
}

func TestAuditRepositoryDetectsHardCodedMigrationRepairSequence(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"database",
		"migrationrepair",
		"postgres.go",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, []byte("// WHERE version IN ('010', '011', '012')\n")...)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureCheck(
		failures,
		"Migration repair inspection follows the plan boundary",
	) {
		t.Fatalf("hard-coded migration sequence audit rule was not triggered: %#v", failures)
	}
}

func TestAuditRepositoryDetectsMissingTrajectoryQueryProfileMigration(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(
		root,
		"database",
		"migrations",
		"021_trajectory_query_profiles.sql",
	)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureCheck(
		failures,
		"Trajectory query profile migration owns proven index changes",
	) {
		t.Fatalf("missing trajectory profile migration was not detected: %#v", failures)
	}
}

func TestAuditRepositoryDetectsMigratorNilContextFallback(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(root, "apps", "api", "internal", "database", "migrator", "runner.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, []byte("// ctx = context.Background()\n")...)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureCheck(failures, "Migrator operations reject nil caller context") {
		t.Fatalf("migrator nil-context fallback was not detected: %#v", failures)
	}
}

func TestAuditRepositoryRejectsReopenedOverallStatus(t *testing.T) {
	root := createCompleteFixture(t)
	path := filepath.Join(root, "scripts", "verify-stage-14-completion.sh")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.ReplaceAll(
		string(content),
		"STAGE_14_OVERALL_STATUS=CLOSED",
		"STAGE_14_OVERALL_STATUS=REOPENED",
	)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	failures := auditRepository(root, &bytes.Buffer{})
	if !containsFailureDetail(failures, "STAGE_14_OVERALL_STATUS=REOPENED") {
		t.Fatalf("reopened overall status was not rejected: %#v", failures)
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
		"<!-- STAGE-14-36-FINAL-CLOSURE:README -->\nStage 14 is closed\nSTAGE_14_OVERALL_STATUS=CLOSED\n",
	)
	writeFixtureFile(
		t,
		root,
		"docs/25_IMPLEMENTATION_SEQUENCE.md",
		"<!-- STAGE-14-36-FINAL-CLOSURE:IMPLEMENTATION -->\nStage 14 is closed\nSTAGE_14_OVERALL_STATUS=CLOSED\n",
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
			content += "STAGE-14-36-FINAL-CLOSURE:DOCUMENT-70\nSTAGE_14_CURRENT_SCOPE_AUDIT=PASS\nSTAGE_14_OVERALL_STATUS=CLOSED\nFlight Feature timestamp integration\nbackend container\nfrontend production build\n"
		case "78_STAGE_14_36_FINAL_CLOSURE_AUDIT.md":
			content += "Stage 14 is closed\nscripts/verify-stage-14-completion.sh\nSTAGE_14_36_FINAL_CLOSURE_AUDIT=PASS\nSTAGE_14_OVERALL_STATUS=CLOSED\n"
		}
		writeFixtureFile(t, root, filepath.Join("docs", fileName), content)
		index.WriteString(fileName)
		index.WriteByte('\n')
	}
	index.WriteString("<!-- POST-CLOSURE-MIGRATOR-CONTEXT-HARDENING:DOCUMENT-INDEX -->\n79_POST_CLOSURE_MIGRATOR_CONTEXT_HARDENING.md\n")
	writeFixtureFile(t, root, "docs/DOCUMENT_INDEX.md", index.String())
	writeFixtureFile(
		t,
		root,
		"docs/79_POST_CLOSURE_MIGRATOR_CONTEXT_HARDENING.md",
		"Stage 14 remains closed\nErrMigrationContextRequired\nPOST_CLOSURE_MIGRATOR_CONTEXT_HARDENING=PASS\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/database/migrator/runner.go",
		"package migrator\n// ErrMigrationContextRequired\n// func requireMigrationContext(\n// func (runner *Runner) EnsureSchemaMigrations(\n// func (runner *Runner) Status(\n// func (runner *Runner) ApplyPending(\n// func (runner *Runner) withMigrationLock(\n// context.WithTimeout(\n// context.Background()\n// migrationLockReleaseTimeout\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/database/migrator/context_contract_test.go",
		"package migrator\n// TestMigratorPublicOperationsRejectNilContext\n// TestWithMigrationLockRejectsNilContextBeforePoolAccess\n// TestMigratorContextSourceContract\n// TestMigratorCleanupContextsRemainIndependentAndBounded\n",
	)

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
			"STAGE_14_31_WRITE_REPOSITORY_DECOMPOSITION=PASS",
			"STAGE_14_32_AIRPORT_PAGINATION=PASS",
			"STAGE_14_33_EXPLICIT_CONTEXT_AND_WRITE_MODE=PASS",
			"STAGE_14_34_POSTGRESQL_CONTRACT_CONSOLIDATION=PASS",
			"scripts/profile-stage-14-trajectory-queries.sh",
			"STAGE_14_TRAJECTORY_QUERY_PROFILING=PASS",
			"STAGE_14_35_TRAJECTORY_QUERY_PROFILING=PASS",
			"STAGE_14_36_FINAL_CLOSURE_AUDIT=PASS",
			"POST_CLOSURE_MIGRATOR_CONTEXT_HARDENING=PASS",
			"STAGE_14_CURRENT_SCOPE_AUDIT=PASS",
			"STAGE_14_OVERALL_STATUS=CLOSED",
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
		"apps/api/internal/repository/postgres/trajectory_read_consolidation_test.go",
		"package postgres\n// TestTrajectoryReadQueriesHaveOneCanonicalOwner\n// TestTrajectoryRowMappingHasDedicatedOwners\n// TestAllTrajectoryReadBoundariesPreserveCallerContext\n// TestTrajectoryProfileIndexesMatchProductionOrdering\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/trajectory_query_profile_integration_test.go",
		"package postgres\n// TestTrajectoryQueryProfilesUseExpectedIndexes\n// EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)\n// flight_trajectories_icao24_latest_idx\n// flight_trajectories_end_time_order_idx\n// trajectory_segments_trajectory_sequence_unique\n// coverage_gaps_trajectory_time_idx\n",
	)
	writeFixtureFile(
		t,
		root,
		"database/migrations/021_trajectory_query_profiles.sql",
		"BEGIN;\nDROP INDEX trajectory_segments_trajectory_sequence_idx;\nCREATE INDEX flight_trajectories_icao24_latest_idx ON flight_trajectories (icao24, end_time DESC, start_time DESC, created_at DESC);\nCREATE INDEX flight_trajectories_end_time_order_idx ON flight_trajectories (end_time DESC, start_time DESC, created_at DESC);\nCOMMIT;\n",
	)
	writeFixtureFile(
		t,
		root,
		"scripts/profile-stage-14-trajectory-queries.sh",
		"TEST_DATABASE_URL\nTestTrajectoryQueryProfilesUseExpectedIndexes\nSTAGE_14_TRAJECTORY_QUERY_PROFILING=PASS\n",
	)
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
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/airport_import_repository.go",
		"package postgres\n// executeAirportImport(\n// rollbackRepositoryTransaction(tx)\n// requireRepositoryContext(ctx)\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/flightstate_repository.go",
		"package postgres\n// saveFlightStateBatch(\n// rollbackRepositoryTransaction(tx)\n// requireRepositoryContext(ctx)\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/trajectory_write_repository.go",
		"package postgres\n// rollbackRepositoryTransaction(tx)\n// requireRepositoryContext(ctx)\n// newLiveTrajectoryWriteRequest(item)\n// newReconciledTrajectoryWriteRequest(\n// request.isReconciled()\n// switch request.mode\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/airport_import_staging_write.go",
		"package postgres\n// CREATE TEMP TABLE airport_import_staging\n// INSERT INTO airport_import_staging\n// func stageAirportImportRecords(\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/airport_import_merge_write.go",
		"package postgres\n// UPDATE airports AS target\n// INSERT INTO airports (\n// func updateAirportsByICAO(\n// func insertRemainingAirports(\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/flightstate_write.go",
		"package postgres\n// INSERT INTO flight_states\n// func saveFlightStateBatch(\n// func prepareFlightStateInsertArguments(\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/write_repository_decomposition_test.go",
		"package postgres\n// TestWriteRepositoryCoordinatorsRemainNarrow\n// TestWriteRepositoryResponsibilitiesHaveDedicatedOwners\n// TestWriteRepositoryCoordinatorsPreserveTransactionBoundary\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/stage14_correctness_constraints_integration_test.go",
		"package postgres\n// migrator.NewRunner\n// runner.ApplyPending\n// assertStage14CorrectnessPostgresCode(t, err, \"23514\")\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/domain/airport/pagination.go",
		"package airport\n// DefaultListPageSize\n// MaximumListPageSize\n// type ListCursor struct\n// type ListRequest struct\n// type ListPage struct\n// func NormalizeListRequest(\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/airport_repository.go",
		"package postgres\n// airport.MaximumListPageSize\n// repository.ListPage(ctx, request)\n// scanAirportRecord(\n// requireRepositoryContext(ctx)\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/airport_read_queries.go",
		"package postgres\n// ORDER BY a.name ASC, a.id ASC\n// a.name > $1\n// a.name = $1\n// a.id > $2::uuid\n// LIMIT $1\n// LIMIT $3\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/airport_pagination_read.go",
		"package postgres\n// func (repository *AirportRepository) ListPage(\n// normalized.Limit + 1\n// scanAirportRecord(rows)\n// buildAirportPage(records, normalized.Limit)\n// requireRepositoryContext(ctx)\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/airport_row_scan.go",
		"package postgres\n// func scanAirportRecord(\n// applyAirportElevationDatabaseValue(\n// elevationFeet pgtype.Int4\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/airport_elevation_semantics_test.go",
		"package postgres\n// airport_read_queries.go\n// canonical Airport select columns must own nullable elevation exactly once\n// all Airport read queries must share the canonical select columns\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/airport_pagination_contract_test.go",
		"package postgres\n// TestAirportPaginationContractRemainsKeysetBounded\n// TestAirportReadPathsShareOneRowScanner\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/airport_pagination_integration_test.go",
		"package postgres\n// TestAirportListPageUsesStableDuplicateNameCursor\n// TestAirportListLegacyAdapterCollectsBoundedPages\n// Alpha Airport\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/repository_context.go",
		"package postgres\\n// ErrRepositoryContextRequired\\n// func requireRepositoryContext(\\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/trajectory_write_mode.go",
		"package postgres\\n// trajectoryWriteModeLive\\n// trajectoryWriteModeReconciled\\n// type trajectoryWriteRequest struct\\n// func newLiveTrajectoryWriteRequest(\\n// func newReconciledTrajectoryWriteRequest(\\n// func (request trajectoryWriteRequest) validate() error\\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/repository_boundary_contract_test.go",
		"package postgres\\n// TestRepositoryOperationsDoNotInventCallerContext\\n// TestTrajectoryWriteCoordinatorUsesExplicitMode\\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/repository_helpers.go",
		"package postgres\\n// type nullableUUIDArgument struct\\n// type nullableTextArgument struct\\n// type requiredSourceNameArgument struct\\n// ErrRepositoryUUIDArgumentInvalid\\n// ErrRepositorySourceNameRequired\\n// func requiredSourceNameValue(\\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/postgres_contract_consolidation_test.go",
		"package postgres\\n// TestRepositoryArgumentsDoNotUsePointerNilOrArtificialSourceFallback\\n// TestInternalPostgresQueriesDoNotCastUUIDColumnsToTextForArrayMembership\\n// TestArtificialUnknownSourceFallbackIsAbsentFromInternalPostgresCode\\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/repository/postgres/uuid_array_query_integration_test.go",
		"package postgres\\n// TestUUIDArrayMembershipUsesTypedColumnComparison\\n// SELECT candidate::uuid\\n// unnest($1::text[])\\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/database/migrationrepair/plan.go",
		"package migrationrepair\\n// func LoadPlan(\\n// migrationfile.Parse(fileName)\\n// sha256.Sum256(content)\\n// func (plan Plan) IsLaterVersion(\\n",
	)
	writeFixtureFile(
		t,
		root,
		"apps/api/internal/database/migrationrepair/postgres.go",
		"package migrationrepair\\n// plan Plan\\n// WHERE version >= $1\\n// plan.Anchor.Version\\n",
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
