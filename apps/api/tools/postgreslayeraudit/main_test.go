package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditAcceptsClosureFixture(t *testing.T) {
	root := createAuditFixture(t)
	result := auditPostgreSQLLayer(root)
	if len(result.violations) != 0 {
		t.Fatalf("closure fixture violations = %v", result.violations)
	}
}

func TestAuditRejectsLegacyMetricsFlag(t *testing.T) {
	root := createAuditFixture(t)
	writeFixture(t, root, "apps/api/internal/domain/metrics/model.go", "type Q struct { UseBounds bool }\n")

	result := auditPostgreSQLLayer(root)
	if len(result.violations) == 0 {
		t.Fatal("legacy metrics flag was not rejected")
	}
}

func createAuditFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"apps/api/go.mod":                                                                    "module fixture\n",
		"docs/DOCUMENT_INDEX.md":                                                             "index\n",
		"apps/api/internal/domain/metrics/model.go":                                          "package metrics\ntype ActiveAircraftQuery struct {\n\tScope        ActiveAircraftQueryScope\n}\n",
		"apps/api/internal/repository/postgres/metrics_repository.go":                        "activeAircraftGlobalStatement activeAircraftBoundedStatement\n",
		"apps/api/internal/repository/postgres/data_quality_repository.go":                   "dataQualityWriteRequest\n",
		"apps/api/internal/database/migrationaudit/auditor.go":                               "ErrContextRequired\n",
		"apps/api/internal/repository/postgres/trajectory_read_queries.go":                   "FROM unnest($1::uuid[]) ON trajectory.id = requested.id\n",
		"apps/api/internal/repository/postgres/analytical_trajectory_repository.go":          "trajectoryUUIDArguments\n",
		"apps/api/internal/repository/postgres/trajectory_id_arguments.go":                   "trajectoryUUIDArguments\n",
		"apps/api/internal/repository/postgres/repository_helpers.go":                        "ErrRepositorySourceNameRequired\n",
		"apps/api/internal/repository/postgres/airport_repository.go":                        "ListPage MaximumListPageSize\n",
		"apps/api/internal/repository/postgres/trajectory_query_profile_integration_test.go": "EXPLAIN (ANALYZE, BUFFERS flight_trajectories_end_time_order_idx\n",
		"apps/api/internal/repository/postgres/trajectory_read_snapshot.go":                  "package postgres\nfunc rollback() { tx.Rollback(rollbackCtx) }\n",
		"docs/81_POSTGRESQL_LAYER_FULL_AUDIT_CLOSURE.md":                                     "fixed not applicable deliberately rejected EXPLAIN (ANALYZE, BUFFERS) migrationrepair\n",
		"apps/api/cmd/server/main.go":                                                        "package main\n",
		"apps/api/cmd/ingest/main.go":                                                        "package main\n",
		"apps/api/cmd/reconcile/main.go":                                                     "package main\n",
		"apps/api/internal/server/stub.go":                                                   "package server\n",
		"apps/api/internal/services/stub.go":                                                 "package services\n",
	}
	for path, content := range files {
		writeFixture(t, root, path, content)
	}
	return root
}

func writeFixture(t *testing.T, root string, relative string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestAuditRejectsDataQualitySentinel(t *testing.T) {
	root := createAuditFixture(t)
	writeFixture(
		t,
		root,
		"apps/api/internal/repository/postgres/data_quality_repository.go",
		"dataQualityWriteRequest\nif reconciliationTaskID == \"\" {}\n",
	)

	result := auditPostgreSQLLayer(root)
	if len(result.violations) == 0 {
		t.Fatal("Data Quality sentinel was not rejected")
	}
}

func TestAuditRejectsTextTrajectoryIdentifiers(t *testing.T) {
	root := createAuditFixture(t)
	writeFixture(
		t,
		root,
		"apps/api/internal/repository/postgres/trajectory_read_queries.go",
		"FROM unnest($1::text[]) ON trajectory.id = requested.id_text::uuid\n",
	)

	result := auditPostgreSQLLayer(root)
	if len(result.violations) == 0 {
		t.Fatal("text trajectory identifiers were not rejected")
	}
}

func TestAuditRejectsRepositoryContextFallback(t *testing.T) {
	root := createAuditFixture(t)
	writeFixture(
		t,
		root,
		"apps/api/internal/repository/postgres/context_fallback.go",
		"package postgres\nfunc fallback() { ctx = context.Background() }\n",
	)

	result := auditPostgreSQLLayer(root)
	if len(result.violations) == 0 {
		t.Fatal("repository context fallback was not rejected")
	}
}

func TestAuditRejectsMigrationRepairRuntimeImport(t *testing.T) {
	root := createAuditFixture(t)
	writeFixture(
		t,
		root,
		"apps/api/cmd/server/main.go",
		"package main\nimport _ \"fixture/internal/database/migrationrepair\"\n",
	)

	result := auditPostgreSQLLayer(root)
	if len(result.violations) == 0 {
		t.Fatal("migration repair runtime import was not rejected")
	}
}

func TestAuditRejectsRequestContextRollback(t *testing.T) {
	root := createAuditFixture(t)
	writeFixture(
		t,
		root,
		"apps/api/internal/repository/postgres/request_rollback.go",
		"package postgres\nfunc rollback() { tx.Rollback(ctx) }\n",
	)

	result := auditPostgreSQLLayer(root)
	if len(result.violations) == 0 {
		t.Fatal("request context rollback was not rejected")
	}
}

func TestAuditIgnoresLegacyTrajectoryNameInsideStringLiteral(t *testing.T) {
	root := createAuditFixture(t)
	writeFixture(
		t,
		root,
		"apps/api/internal/trajectory/string_evidence.go",
		"package trajectory\nconst historicalName = \"ListTrajectoriesByEndTimeAndBounds\"\n",
	)

	result := auditPostgreSQLLayer(root)
	if len(result.violations) != 0 {
		t.Fatalf("string evidence produced violations = %v", result.violations)
	}
}

func TestAuditRejectsLegacyTrajectoryMethodIdentifier(t *testing.T) {
	root := createAuditFixture(t)
	writeFixture(
		t,
		root,
		"apps/api/internal/trajectory/legacy_method.go",
		"package trajectory\nfunc ListTrajectoriesByEndTimeAndBounds() {}\n",
	)

	result := auditPostgreSQLLayer(root)
	if len(result.violations) == 0 {
		t.Fatal("legacy trajectory method identifier was not rejected")
	}
}
func TestAuditSourceTargetsLegacyIdentifierWithoutRejectingCurrentName(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read audit source: %v", err)
	}
	source := string(content)
	if !strings.Contains(source, `"ListTrajectoriesByEndTimeAndBounds"`) {
		t.Fatal("audit source no longer targets the legacy trajectory identifier")
	}
	if strings.Contains(
		source,
		`productionGoIdentifierExists(
			root,
			[]string{"apps/api/cmd", "apps/api/internal"},
			"ListTrajectoriesWithinBounds"`,
	) {
		t.Fatal("audit source incorrectly rejects the current trajectory identifier")
	}
}
