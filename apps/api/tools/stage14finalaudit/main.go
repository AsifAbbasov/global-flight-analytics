package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type auditFailure struct {
	Check  string
	Detail string
}

type fileRule struct {
	Name      string
	Path      string
	Required  []string
	Forbidden []string
	MaxLines  int
}

var stage14Documents = []string{
	"41_STAGE_14_1_ARCHITECTURE_CONSOLIDATION_FOUNDATION.md",
	"42_STAGE_14_2_DEAD_CODE_CLASSIFICATION_AND_REMOVAL.md",
	"43_STAGE_14_3_AIRPORT_INTELLIGENCE_PRODUCTION_INTEGRATION.md",
	"44_STAGE_14_4_FEATURE_MATERIALIZATION_AND_PROFILER_REMOVAL.md",
	"45_STAGE_14_5_MUTATION_ENDPOINT_PROTECTION.md",
	"46_STAGE_14_6_FORMULA_BENCHMARK_AND_CALIBRATION_GATE.md",
	"47_STAGE_14_7_FRONTEND_DEPENDENCY_SECURITY_REMEDIATION.md",
	"48_STAGE_14_8_SERVER_COMPOSITION_ROOT_DECOMPOSITION.md",
	"49_STAGE_14_9_HTTP_QUERY_AND_CONTRACT_BOUNDARY_HARDENING.md",
	"50_STAGE_14_10_TRANSPONDER_EVIDENCE_PRODUCTION_INTEGRATION.md",
	"51_STAGE_14_11_TARGETED_LARGE_MODULE_HARDENING.md",
	"52_STAGE_14_12_PROJECTION_READ_SNAPSHOT_CONSISTENCY.md",
	"53_STAGE_14_13_NULLABLE_TELEMETRY_INTEGRITY.md",
	"54_STAGE_14_14_COMPOSITE_HISTORICAL_PAGINATION_CURSOR.md",
	"55_STAGE_14_15_WEATHER_COMPOSITION_BOUNDARY.md",
	"56_BACKEND_FINAL_CORRECTNESS_AUDIT.md",
	"57_STAGE_14_16_END_TO_END_TELEMETRY_AVAILABILITY.md",
	"58_STAGE_14_17_POSTGRES_MIGRATION_ATOMICITY.md",
	"59_STAGE_14_18_POSTGRES_BASELINE_REMOVAL.md",
	"60_STAGE_14_19_DATA_QUALITY_PARENT_INTEGRITY.md",
	"61_STAGE_14_20_TRAJECTORY_READ_SNAPSHOT_CONSISTENCY.md",
	"62_STAGE_14_21_INGESTION_RUN_TERMINAL_INTEGRITY.md",
	"63_STAGE_14_22_TRAJECTORY_RELATIONAL_INTEGRITY.md",
	"64_STAGE_14_23_CANONICAL_MIGRATION_FILENAME_CONTRACT.md",
	"65_STAGE_14_24_EXPLICIT_ALTITUDE_INTEGER_POLICY.md",
	"66_STAGE_14_25_TRAFFIC_ALTITUDE_STATUS_SEMANTICS.md",
	"67_STAGE_14_26_AIRPORT_ELEVATION_SEMANTICS.md",
	"68_STAGE_14_27_FLIGHT_FEATURE_TIMESTAMP_CONSISTENCY.md",
	"69_STAGE_14_28_POSTGRES_TRAJECTORY_REPOSITORY_DECOMPOSITION.md",
	"70_STAGE_14_FINAL_COMPLETION_AUDIT.md",
	"71_STAGE_14_29_MIGRATION_CATALOG_INTEGRITY.md",
	"72_STAGE_14_30_POSTGRES_CORRECTNESS_HARDENING.md",
	"73_STAGE_14_31_POSTGRES_WRITE_REPOSITORY_DECOMPOSITION.md",
}

var trajectoryOwnerFiles = []string{
	"apps/api/internal/repository/postgres/trajectory_child_read.go",
	"apps/api/internal/repository/postgres/trajectory_gap_read.go",
	"apps/api/internal/repository/postgres/trajectory_gap_write.go",
	"apps/api/internal/repository/postgres/trajectory_parent_read.go",
	"apps/api/internal/repository/postgres/trajectory_parent_write.go",
	"apps/api/internal/repository/postgres/trajectory_reconciliation_write.go",
	"apps/api/internal/repository/postgres/trajectory_segment_read.go",
	"apps/api/internal/repository/postgres/trajectory_segment_write.go",
	"apps/api/internal/repository/postgres/trajectory_write_repository.go",
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("stage14finalaudit", flag.ContinueOnError)
	flags.SetOutput(stderr)

	rootValue := flags.String(
		"root",
		"",
		"repository root; auto-detected when omitted",
	)
	strict := flags.Bool(
		"strict",
		true,
		"return a non-zero exit code when a Stage 14 completion invariant fails",
	)

	if err := flags.Parse(args); err != nil {
		return 1
	}

	root, err := resolveRepositoryRoot(*rootValue)
	if err != nil {
		fmt.Fprintf(stderr, "locate repository root: %v\n", err)
		return 1
	}

	failures := auditRepository(root, stdout)
	if len(failures) == 0 {
		fmt.Fprintln(stdout, "Stage 14 current-scope source audit: PASS")
		return 0
	}

	fmt.Fprintln(stderr, "Stage 14 current-scope source audit: FAIL")
	for _, failure := range failures {
		fmt.Fprintf(stderr, "- %s: %s\n", failure.Check, failure.Detail)
	}
	if *strict {
		return 1
	}
	return 0
}

func resolveRepositoryRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		root, err := filepath.Abs(strings.TrimSpace(explicit))
		if err != nil {
			return "", err
		}
		if err := validateRepositoryRoot(root); err != nil {
			return "", err
		}
		return root, nil
	}

	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if validateRepositoryRoot(current) == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New(
				"repository root containing apps/api/go.mod and apps/web/package.json was not found",
			)
		}
		current = parent
	}
}

func validateRepositoryRoot(root string) error {
	for _, relativePath := range []string{
		"apps/api/go.mod",
		"apps/web/package.json",
	} {
		path := filepath.Join(root, filepath.FromSlash(relativePath))
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("required repository file %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("required repository path is not a file: %s", path)
		}
	}
	return nil
}

func auditRepository(root string, output io.Writer) []auditFailure {
	groups := []struct {
		name  string
		check func(string) []auditFailure
	}{
		{name: "Stage 14 document register", check: auditDocumentRegister},
		{name: "Migration catalog integrity", check: auditMigrationCatalog},
		{name: "Go toolchain security", check: auditGoToolchainSecurity},
		{name: "Unified verification reachability", check: auditUnifiedVerification},
		{name: "Continuous integration coverage", check: auditContinuousIntegration},
		{name: "PostgreSQL closure surface", check: auditPostgresClosureSurface},
	}

	failures := make([]auditFailure, 0)
	for _, group := range groups {
		groupFailures := group.check(root)
		if len(groupFailures) == 0 {
			fmt.Fprintf(output, "%s: PASS\n", group.name)
			continue
		}
		failures = append(failures, groupFailures...)
	}

	sort.Slice(failures, func(left int, right int) bool {
		if failures[left].Check == failures[right].Check {
			return failures[left].Detail < failures[right].Detail
		}
		return failures[left].Check < failures[right].Check
	})
	return failures
}

func auditMigrationCatalog(root string) []auditFailure {
	directory := filepath.Join(root, "database", "migrations")
	entries, err := os.ReadDir(directory)
	if err != nil {
		return []auditFailure{{
			Check:  "Migration catalog integrity",
			Detail: fmt.Sprintf("read database/migrations: %v", err),
		}}
	}

	seenVersions := make(map[string]string)
	seenFiles := make(map[string]bool)
	failures := make([]auditFailure, 0)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		fileName := entry.Name()
		seenFiles[fileName] = true
		parts := strings.SplitN(strings.TrimSuffix(fileName, ".sql"), "_", 2)
		if len(parts) != 2 || len(parts[0]) != 3 || parts[1] == "" {
			failures = append(failures, auditFailure{
				Check:  "Migration catalog canonical names",
				Detail: fmt.Sprintf("%s is not a canonical NNN_name.sql file", fileName),
			})
			continue
		}
		if _, parseErr := strconv.Atoi(parts[0]); parseErr != nil {
			failures = append(failures, auditFailure{
				Check:  "Migration catalog canonical names",
				Detail: fmt.Sprintf("%s has a non-numeric version", fileName),
			})
			continue
		}

		if previous, exists := seenVersions[parts[0]]; exists {
			failures = append(failures, auditFailure{
				Check: "Migration catalog unique versions",
				Detail: fmt.Sprintf(
					"version %s is owned by both %s and %s",
					parts[0],
					previous,
					fileName,
				),
			})
			continue
		}
		seenVersions[parts[0]] = fileName
	}

	for _, required := range []string{
		"016_add_flight_state_observation_metadata.sql",
		"019_data_quality_parent_integrity.sql",
		"020_stage14_correctness_hardening.sql",
	} {
		if !seenFiles[required] {
			failures = append(failures, auditFailure{
				Check:  "Migration catalog required ownership",
				Detail: fmt.Sprintf("missing database/migrations/%s", required),
			})
		}
	}
	if seenFiles["016_data_quality_parent_integrity.sql"] {
		failures = append(failures, auditFailure{
			Check:  "Migration catalog retired duplicate",
			Detail: "database/migrations/016_data_quality_parent_integrity.sql must not exist",
		})
	}

	failures = append(failures, auditRules(root, []fileRule{
		{
			Name: "Data Quality source test follows canonical migration ownership",
			Path: "apps/api/internal/repository/postgres/data_quality_parent_integrity_test.go",
			Required: []string{
				"database/migrations/019_data_quality_parent_integrity.sql",
				"migration 019",
			},
			Forbidden: []string{
				"database/migrations/016_data_quality_parent_integrity.sql",
			},
		},
		{
			Name: "Data Quality integration test follows canonical migration ownership",
			Path: "apps/api/internal/repository/postgres/data_quality_parent_integrity_integration_test.go",
			Required: []string{
				"database/migrations/019_data_quality_parent_integrity.sql",
			},
			Forbidden: []string{
				"database/migrations/016_data_quality_parent_integrity.sql",
			},
		},
		{
			Name: "Data Quality document follows canonical migration ownership",
			Path: "docs/60_STAGE_14_19_DATA_QUALITY_PARENT_INTEGRITY.md",
			Required: []string{
				"Migration 019",
				"019_data_quality_parent_integrity.sql",
			},
			Forbidden: []string{
				"016_data_quality_parent_integrity.sql",
			},
		},
		{
			Name: "Repository status surfaces keep Stage 14 reopened",
			Path: "README.md",
			Required: []string{
				"STAGE-14-29-MIGRATION-CATALOG-INTEGRITY:README",
				"Stage 14 remains reopened",
			},
		},
		{
			Name: "Implementation sequence keeps Stage 14 reopened",
			Path: "docs/25_IMPLEMENTATION_SEQUENCE.md",
			Required: []string{
				"STAGE-14-29-MIGRATION-CATALOG-INTEGRITY:IMPLEMENTATION",
				"Stage 14 remains reopened",
			},
		},
	})...)

	return failures
}

func auditDocumentRegister(root string) []auditFailure {
	failures := make([]auditFailure, 0)
	indexPath := filepath.Join(root, "docs", "DOCUMENT_INDEX.md")
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		return []auditFailure{{
			Check:  "Stage 14 document index",
			Detail: fmt.Sprintf("read docs/DOCUMENT_INDEX.md: %v", err),
		}}
	}
	index := string(indexContent)

	for expectedNumber, fileName := range stage14Documents {
		number := expectedNumber + 41
		prefix := strings.SplitN(fileName, "_", 2)[0]
		parsed, parseErr := strconv.Atoi(prefix)
		if parseErr != nil || parsed != number {
			failures = append(failures, auditFailure{
				Check:  "Stage 14 document sequence",
				Detail: fmt.Sprintf("%s does not represent document %d", fileName, number),
			})
		}

		path := filepath.Join(root, "docs", fileName)
		info, statErr := os.Stat(path)
		if statErr != nil || !info.Mode().IsRegular() {
			failures = append(failures, auditFailure{
				Check:  "Stage 14 document register",
				Detail: fmt.Sprintf("missing regular file docs/%s", fileName),
			})
		}

		count := strings.Count(index, fileName)
		if count != 1 {
			failures = append(failures, auditFailure{
				Check:  "Stage 14 document index",
				Detail: fmt.Sprintf("%s is indexed %d times, expected exactly once", fileName, count),
			})
		}
	}
	return failures
}

func auditGoToolchainSecurity(root string) []auditFailure {
	return auditRules(root, []fileRule{
		{
			Name: "Go module pins the patched standard library",
			Path: "apps/api/go.mod",
			Required: []string{
				"go 1.26.5",
			},
			Forbidden: []string{
				"go 1.26.2",
			},
		},
		{
			Name: "Backend image uses the patched Go builder",
			Path: "apps/api/Dockerfile",
			Required: []string{
				"ARG GO_IMAGE=golang:1.26.5-alpine3.24",
			},
			Forbidden: []string{
				"golang:1.26.2",
			},
		},
		{
			Name: "Stage 14 audit selects and verifies the patched toolchain",
			Path: "scripts/verify-stage-14-completion.sh",
			Required: []string{
				"GOTOOLCHAIN=go1.26.5+auto",
				"go env GOVERSION",
				"STAGE_14_GO_TOOLCHAIN_AUDIT=PASS",
			},
		},
		{
			Name: "Backend continuous integration derives Go from go.mod",
			Path: ".github/workflows/backend-ci.yml",
			Required: []string{
				"go-version-file: apps/api/go.mod",
			},
		},
	})
}

func auditUnifiedVerification(root string) []auditFailure {
	return auditRules(root, []fileRule{
		{
			Name: "Stage 14 current-scope script covers every enforced boundary",
			Path: "scripts/verify-stage-14-completion.sh",
			Required: []string{
				"GOTOOLCHAIN=go1.26.5+auto",
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
				"STAGE_14_CURRENT_SCOPE_AUDIT=PASS",
			},
			MaxLines: 320,
		},
		{
			Name: "Root package exposes the Stage 14 current-scope audit",
			Path: "package.json",
			Required: []string{
				`"verify:stage14": "bash scripts/verify-stage-14-completion.sh"`,
			},
		},
		{
			Name: "Backend audit document points to the cross-stack current-scope gate",
			Path: "docs/56_BACKEND_FINAL_CORRECTNESS_AUDIT.md",
			Required: []string{
				"scripts/verify-stage-14-completion.sh",
				"STAGE_14_CURRENT_SCOPE_AUDIT=PASS",
			},
		},
		{
			Name: "Reopened Stage 14 document defines the current-scope marker",
			Path: "docs/70_STAGE_14_FINAL_COMPLETION_AUDIT.md",
			Required: []string{
				"STAGE_14_CURRENT_SCOPE_AUDIT=PASS",
				"Flight Feature timestamp integration",
				"backend container",
				"frontend production build",
			},
		},
	})
}

func auditContinuousIntegration(root string) []auditFailure {
	return auditRules(root, []fileRule{
		{
			Name: "Backend continuous integration runs the Stage 14 source audit",
			Path: ".github/workflows/backend-ci.yml",
			Required: []string{
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
			},
		},
		{
			Name: "Frontend continuous integration preserves security and production gates",
			Path: ".github/workflows/frontend-ci.yml",
			Required: []string{
				"pnpm run test:web-dependency-policy",
				"pnpm run verify:web-dependencies",
				"pnpm audit --prod --audit-level moderate",
				"pnpm --filter web lint",
				"pnpm --filter web typecheck",
				"pnpm --filter web build",
			},
		},
	})
}

func auditPostgresClosureSurface(root string) []auditFailure {
	failures := auditRules(root, []fileRule{
		{
			Name: "Trajectory repository root remains state-only",
			Path: "apps/api/internal/repository/postgres/trajectory_repository.go",
			Forbidden: []string{
				"BeginTx(",
				"INSERT INTO",
				"SELECT ",
				"saveTrajectory(",
			},
			MaxLines: 80,
		},
		{
			Name: "Trajectory public read file remains coordinator-only",
			Path: "apps/api/internal/repository/postgres/trajectory_read_repository.go",
			Required: []string{
				"withTrajectoryReadSnapshot(",
				"snapshotRepository.getLatestTrajectoryByICAO24",
				"snapshotRepository.getTrajectoryByID",
			},
			Forbidden: []string{
				"SELECT ",
				"queryTrajectory(",
				"ListTrajectorySegments(",
				"ListCoverageGaps(",
			},
			MaxLines: 150,
		},
		{
			Name: "Trajectory decomposition ownership test remains permanent",
			Path: "apps/api/internal/repository/postgres/trajectory_repository_decomposition_test.go",
			Required: []string{
				"TestTrajectoryRepositoryCoordinatorsRemainNarrow",
				"TestTrajectoryRepositoryResponsibilitiesHaveDedicatedOwners",
				"TestTrajectoryWriteValidationStillPrecedesTransaction",
			},
		},
		{
			Name: "Trajectory relational migration uses qualified aggregate columns",
			Path: "database/migrations/018_trajectory_relational_integrity.sql",
			Required: []string{
				"SUM(segment.point_count)",
				"MIN(segment.sequence_number)",
				"MAX(segment.sequence_number)",
				"FROM trajectory_segments AS segment",
				"WHERE segment.trajectory_id = target_trajectory_id",
			},
			Forbidden: []string{
				"SUM(point_count)",
			},
		},
		{
			Name: "Flight State altitude integration fixture matches repository evidence columns",
			Path: "apps/api/internal/repository/postgres/flightstate_altitude_integration_test.go",
			Required: []string{
				"squawk_code text",
				"special_purpose_indicator boolean",
				"position_source text",
				"aircraft_category smallint",
				"aircraft_category_available boolean",
			},
		},
		{
			Name: "Flight State reconciliation integration fixture matches repository evidence columns",
			Path: "apps/api/internal/repository/postgres/flightstate_reconciliation_repository_integration_test.go",
			Required: []string{
				"squawk_code text",
				"special_purpose_indicator boolean",
				"position_source text",
				"aircraft_category smallint",
				"aircraft_category_available boolean",
			},
		},
		{
			Name: "Flight Feature timestamp mirror remains fail-closed",
			Path: "apps/api/internal/features/featurestore/timestamp_consistency.go",
			Required: []string{
				"postgresTimestampMirrorTolerance = time.Microsecond",
				"func validateTimestampMirror(",
				"delta <= -postgresTimestampMirrorTolerance",
				"delta >= postgresTimestampMirrorTolerance",
			},
		},
		{
			Name: "Airport import write remains coordinator-only",
			Path: "apps/api/internal/repository/postgres/airport_import_repository.go",
			Required: []string{
				"executeAirportImport(",
				"rollbackRepositoryTransaction(tx)",
			},
			Forbidden: []string{
				"CREATE TEMP TABLE",
				"UPDATE airports AS target",
				"INSERT INTO airports (",
				"stageAirportImportRecords(",
			},
			MaxLines: 110,
		},
		{
			Name: "Flight State write delegates preparation and SQL",
			Path: "apps/api/internal/repository/postgres/flightstate_repository.go",
			Required: []string{
				"saveFlightStateBatch(",
				"rollbackRepositoryTransaction(tx)",
			},
			Forbidden: []string{
				"INSERT INTO flight_states",
				"flightstate.NormalizeSquawkCode(",
				"flightstate.ValidateAircraftCategory(",
			},
		},
		{
			Name: "Airport import staging has a dedicated owner",
			Path: "apps/api/internal/repository/postgres/airport_import_staging_write.go",
			Required: []string{
				"CREATE TEMP TABLE airport_import_staging",
				"INSERT INTO airport_import_staging",
				"func stageAirportImportRecords(",
			},
		},
		{
			Name: "Airport import merge has a dedicated owner",
			Path: "apps/api/internal/repository/postgres/airport_import_merge_write.go",
			Required: []string{
				"UPDATE airports AS target",
				"INSERT INTO airports (",
				"func updateAirportsByICAO(",
				"func insertRemainingAirports(",
			},
		},
		{
			Name: "Flight State preparation has a dedicated owner",
			Path: "apps/api/internal/repository/postgres/flightstate_write.go",
			Required: []string{
				"INSERT INTO flight_states",
				"func saveFlightStateBatch(",
				"func prepareFlightStateInsertArguments(",
			},
		},
		{
			Name: "Write repository decomposition test remains permanent",
			Path: "apps/api/internal/repository/postgres/write_repository_decomposition_test.go",
			Required: []string{
				"TestWriteRepositoryCoordinatorsRemainNarrow",
				"TestWriteRepositoryResponsibilitiesHaveDedicatedOwners",
				"TestWriteRepositoryCoordinatorsPreserveTransactionBoundary",
			},
		},
		{
			Name: "Stage 14 correctness migration owns database invariants",
			Path: "database/migrations/020_stage14_correctness_hardening.sql",
			Required: []string{
				"ingestion_runs_processed_counts_check",
				"ingestion_runs_error_message_status_check",
				"flight_route_results_as_of_time_mirror_check",
				"flight_route_results_stored_at_mirror_check",
				"historical_aggregate_results_window_start_mirror_check",
				"historical_aggregate_results_window_end_mirror_check",
				"historical_aggregate_results_as_of_time_mirror_check",
				"historical_aggregate_results_stored_at_mirror_check",
			},
		},
		{
			Name: "Ingestion Run completion validation remains fail-fast",
			Path: "apps/api/internal/repository/postgres/ingestionrun_completion_validation.go",
			Required: []string{
				"recordsUpdated > recordsReceived-recordsInserted",
				"ingestionrun.StatusSuccess",
				"ingestionrun.StatusFailed, ingestionrun.StatusPartial",
			},
		},
		{
			Name: "Route timestamp mirrors remain fail-closed",
			Path: "apps/api/internal/routeintelligence/routestore/timestamp_consistency.go",
			Required: []string{
				"postgresTimestampMirrorTolerance = time.Microsecond",
				"func validateTimestampMirror(",
			},
		},
		{
			Name: "Historical timestamp mirrors remain fail-closed",
			Path: "apps/api/internal/historicalintelligence/historicalaggregate/timestamp_consistency.go",
			Required: []string{
				"postgresTimestampMirrorTolerance = time.Microsecond",
				"func validateTimestampMirror(",
			},
		},
		{
			Name: "Repository rollback uses an independent bounded context",
			Path: "apps/api/internal/repository/postgres/transaction_rollback.go",
			Required: []string{
				"context.WithTimeout(",
				"context.Background()",
				"repositoryRollbackTimeout",
			},
		},
		{
			Name:     "Airport import delegates rollback ownership",
			Path:     "apps/api/internal/repository/postgres/airport_import_repository.go",
			Required: []string{"rollbackRepositoryTransaction(tx)"},
		},
		{
			Name:     "Flight State write delegates rollback ownership",
			Path:     "apps/api/internal/repository/postgres/flightstate_repository.go",
			Required: []string{"rollbackRepositoryTransaction(tx)"},
		},
		{
			Name:     "Trajectory write delegates rollback ownership",
			Path:     "apps/api/internal/repository/postgres/trajectory_write_repository.go",
			Required: []string{"rollbackRepositoryTransaction(tx)"},
		},
		{
			Name: "Correctness constraints use the production migration catalog",
			Path: "apps/api/internal/repository/postgres/stage14_correctness_constraints_integration_test.go",
			Required: []string{
				"migrator.NewRunner",
				"runner.ApplyPending",
				"assertStage14CorrectnessPostgresCode(t, err, \"23514\")",
			},
		},
		{
			Name: "Traffic altitude remains typed and nullable",
			Path: "apps/api/internal/domain/traffic/altitude.go",
			Required: []string{
				"func ResolveCurrentAltitude(",
				"AltitudeSourceGeometric",
				"AltitudeSourceBarometric",
				"AltitudeSourceGround",
				"return nil,",
			},
		},
		{
			Name: "Airport elevation remains explicit",
			Path: "apps/api/internal/domain/airport/elevation.go",
			Required: []string{
				"ElevationStatusObserved",
				"ElevationStatusUnknown",
				"ElevationStatusInvalid",
				"func ResolveElevation(",
			},
		},
	})

	failures = append(
		failures,
		auditTerminalIngestionRunFixtures(root)...,
	)
	failures = append(
		failures,
		auditFlightStateRepositoryFixtureParity(root)...,
	)

	for _, relativePath := range trajectoryOwnerFiles {
		path := filepath.Join(root, filepath.FromSlash(relativePath))
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			failures = append(failures, auditFailure{
				Check:  "Trajectory repository responsibility owners",
				Detail: fmt.Sprintf("missing regular file %s", relativePath),
			})
		}
	}
	return failures
}

func auditTerminalIngestionRunFixtures(root string) []auditFailure {
	directory := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"repository",
		"postgres",
	)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return []auditFailure{{
			Check:  "Terminal ingestion-run integration fixtures",
			Detail: fmt.Sprintf("read %s: %v", directory, err),
		}}
	}

	failures := make([]auditFailure, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_integration_test.go") {
			continue
		}

		path := filepath.Join(directory, entry.Name())
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			failures = append(failures, auditFailure{
				Check:  "Terminal ingestion-run integration fixtures",
				Detail: fmt.Sprintf("read %s: %v", path, readErr),
			})
			continue
		}

		text := string(content)
		searchOffset := 0
		for {
			relativeStart := strings.Index(
				text[searchOffset:],
				"INSERT INTO ingestion_runs (",
			)
			if relativeStart < 0 {
				break
			}
			start := searchOffset + relativeStart
			relativeEnd := strings.Index(text[start:], ");")
			if relativeEnd < 0 {
				failures = append(failures, auditFailure{
					Check:  "Terminal ingestion-run integration fixtures",
					Detail: fmt.Sprintf("%s contains an unterminated ingestion_runs insert", entry.Name()),
				})
				break
			}

			end := start + relativeEnd + 2
			statement := text[start:end]
			terminal := strings.Contains(statement, "'success'") ||
				strings.Contains(statement, "'failed'") ||
				strings.Contains(statement, "'partial'")
			if terminal && !strings.Contains(statement, "finished_at") {
				failures = append(failures, auditFailure{
					Check: "Terminal ingestion-run integration fixtures",
					Detail: fmt.Sprintf(
						"%s inserts a terminal ingestion run without finished_at",
						entry.Name(),
					),
				})
			}
			searchOffset = end
		}
	}

	return failures
}

func auditFlightStateRepositoryFixtureParity(root string) []auditFailure {
	directory := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"repository",
		"postgres",
	)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return []auditFailure{{
			Check:  "Flight State repository fixture parity",
			Detail: fmt.Sprintf("read %s: %v", directory, err),
		}}
	}

	required := []string{
		"squawk_code",
		"special_purpose_indicator",
		"position_source",
		"aircraft_category",
		"aircraft_category_available",
	}
	failures := make([]auditFailure, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_integration_test.go") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			failures = append(failures, auditFailure{
				Check:  "Flight State repository fixture parity",
				Detail: fmt.Sprintf("read %s: %v", path, readErr),
			})
			continue
		}
		text := string(content)
		if !strings.Contains(text, "CREATE TABLE flight_states (") ||
			!strings.Contains(text, "NewFlightStateRepository(") {
			continue
		}
		for _, column := range required {
			if !strings.Contains(text, column) {
				failures = append(failures, auditFailure{
					Check: "Flight State repository fixture parity",
					Detail: fmt.Sprintf(
						"%s instantiates FlightStateRepository with flight_states missing %s",
						entry.Name(),
						column,
					),
				})
			}
		}
	}
	return failures
}

func auditRules(root string, rules []fileRule) []auditFailure {
	failures := make([]auditFailure, 0)
	for _, rule := range rules {
		path := filepath.Join(root, filepath.FromSlash(rule.Path))
		content, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, auditFailure{
				Check:  rule.Name,
				Detail: fmt.Sprintf("read %s: %v", rule.Path, err),
			})
			continue
		}
		text := string(content)
		for _, fragment := range rule.Required {
			if !strings.Contains(text, fragment) {
				failures = append(failures, auditFailure{
					Check:  rule.Name,
					Detail: fmt.Sprintf("%s is missing required fragment %q", rule.Path, fragment),
				})
			}
		}
		for _, fragment := range rule.Forbidden {
			if strings.Contains(text, fragment) {
				failures = append(failures, auditFailure{
					Check:  rule.Name,
					Detail: fmt.Sprintf("%s contains forbidden fragment %q", rule.Path, fragment),
				})
			}
		}
		if rule.MaxLines > 0 {
			lineCount := len(strings.Split(strings.TrimSuffix(text, "\n"), "\n"))
			if lineCount > rule.MaxLines {
				failures = append(failures, auditFailure{
					Check:  rule.Name,
					Detail: fmt.Sprintf("%s has %d lines, maximum is %d", rule.Path, lineCount, rule.MaxLines),
				})
			}
		}
	}
	return failures
}
