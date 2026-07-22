package main

import (
	"flag"
	"fmt"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type auditResult struct {
	checks     []string
	violations []string
}

func main() {
	strict := flag.Bool("strict", false, "exit non-zero when a violation is found")
	root := flag.String("root", "", "repository root; auto-detected when empty")
	flag.Parse()

	repositoryRoot, err := resolveRepositoryRoot(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	result := auditPostgreSQLLayer(repositoryRoot)
	for _, check := range result.checks {
		fmt.Println(check + ": PASS")
	}
	if len(result.violations) == 0 {
		fmt.Println("PostgreSQL layer full audit closure: PASS")
		return
	}
	for _, violation := range result.violations {
		fmt.Fprintln(os.Stderr, "VIOLATION:", violation)
	}
	if *strict {
		os.Exit(1)
	}
}

func resolveRepositoryRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return filepath.Abs(explicit)
	}
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if fileExists(filepath.Join(current, "apps", "api", "go.mod")) &&
			fileExists(filepath.Join(current, "docs", "DOCUMENT_INDEX.md")) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root was not found from %s", current)
		}
		current = parent
	}
}

func auditPostgreSQLLayer(root string) auditResult {
	result := auditResult{}
	check := func(name string, condition bool, message string) {
		if condition {
			result.checks = append(result.checks, name)
			return
		}
		result.violations = append(result.violations, message)
	}

	metricsModel := readText(root, "apps/api/internal/domain/metrics/model.go", &result)
	metricsRepository := readText(root, "apps/api/internal/repository/postgres/metrics_repository.go", &result)
	check(
		"Typed metrics scope",
		!strings.Contains(metricsModel, "UseBounds") &&
			regexp.MustCompile(`\bScope\s+ActiveAircraftQueryScope\b`).MatchString(metricsModel) &&
			strings.Contains(metricsRepository, "activeAircraftGlobalStatement") &&
			strings.Contains(metricsRepository, "activeAircraftBoundedStatement"),
		"metrics query still uses a boolean SQL mode instead of an explicit scope",
	)

	dataQuality := readText(root, "apps/api/internal/repository/postgres/data_quality_repository.go", &result)
	check(
		"Typed data quality write mode",
		strings.Contains(dataQuality, "dataQualityWriteRequest") &&
			!strings.Contains(dataQuality, "reconciliationTaskID == \"\"") &&
			!strings.Contains(dataQuality, "ctx = context.Background()"),
		"data quality persistence still uses an empty task identifier or nil-context fallback as a hidden mode",
	)

	migrationAuditor := readText(root, "apps/api/internal/database/migrationaudit/auditor.go", &result)
	check(
		"Migration audit context contract",
		strings.Contains(migrationAuditor, "ErrContextRequired") &&
			!strings.Contains(migrationAuditor, "ctx = context.Background()"),
		"migration audit still hides a nil context",
	)

	productionRepositoryFiles := readProductionGoFiles(
		filepath.Join(root, "apps", "api", "internal", "repository", "postgres"),
		&result,
	)
	check(
		"Trajectory intent-oriented naming",
		!productionGoIdentifierExists(
			root,
			[]string{"apps/api/cmd", "apps/api/internal"},
			"ListTrajectoriesByEndTimeAndBounds",
			&result,
		),
		"legacy ambiguous trajectory method identifier remains after the intent-oriented rename",
	)
	trajectoryQueries := readText(root, "apps/api/internal/repository/postgres/trajectory_read_queries.go", &result)
	trajectoryRepository := readText(root, "apps/api/internal/repository/postgres/analytical_trajectory_repository.go", &result)
	trajectoryArguments := readText(root, "apps/api/internal/repository/postgres/trajectory_id_arguments.go", &result)
	check(
		"Native UUID query contract",
		!strings.Contains(trajectoryQueries, "FROM unnest($1::text[])") &&
			!strings.Contains(trajectoryQueries, "requested.id_text::uuid") &&
			regexp.MustCompile(`unnest\s*\(\s*\$1::uuid\[\]\s*\)`).MatchString(trajectoryQueries) &&
			strings.Contains(trajectoryQueries, "ON trajectory.id = requested.id") &&
			strings.Contains(trajectoryRepository+trajectoryArguments, "trajectoryUUIDArguments"),
		"trajectory identifier query still uses text UUID semantics or does not pass native ordered UUID arguments",
	)

	repositoryHelpers := readText(root, "apps/api/internal/repository/postgres/repository_helpers.go", &result)
	check(
		"Provenance contract",
		strings.Contains(repositoryHelpers, "ErrRepositorySourceNameRequired") &&
			!strings.Contains(repositoryHelpers, "return \"unknown\""),
		"repository provenance still fabricates an unknown source",
	)

	airportRepository := readText(root, "apps/api/internal/repository/postgres/airport_repository.go", &result)
	check(
		"Airport pagination contract",
		strings.Contains(airportRepository, "ListPage") &&
			strings.Contains(airportRepository, "MaximumListPageSize"),
		"airport catalogue still lacks bounded pagination",
	)

	profileTest := readText(root, "apps/api/internal/repository/postgres/trajectory_query_profile_integration_test.go", &result)
	check(
		"Query plan evidence",
		strings.Contains(profileTest, "EXPLAIN (ANALYZE, BUFFERS") &&
			strings.Contains(profileTest, "flight_trajectories_end_time_order_idx"),
		"trajectory performance claims are missing executable EXPLAIN ANALYZE evidence",
	)

	check(
		"Repository context contract",
		!containsAny(productionRepositoryFiles, "ctx = context.Background()"),
		"a PostgreSQL repository still replaces a nil context with context.Background",
	)
	check(
		"Rollback cancellation contract",
		!containsAny(productionRepositoryFiles, ".Rollback(ctx)") &&
			containsAny(productionRepositoryFiles, ".Rollback(rollbackCtx)"),
		"a PostgreSQL repository still rolls back with the potentially cancelled request context or lacks an explicit rollback context",
	)

	runtimeFiles := readSelectedTrees(root, []string{
		"apps/api/cmd/server",
		"apps/api/cmd/ingest",
		"apps/api/cmd/reconcile",
		"apps/api/internal/server",
		"apps/api/internal/services",
	}, &result)
	check(
		"Migration repair runtime isolation",
		!containsAny(runtimeFiles, "internal/database/migrationrepair"),
		"historical migration repair code is reachable from production runtime roots",
	)

	closure := readText(root, "docs/81_POSTGRESQL_LAYER_FULL_AUDIT_CLOSURE.md", &result)
	for _, token := range []string{
		"fixed",
		"not applicable",
		"deliberately rejected",
		"EXPLAIN (ANALYZE, BUFFERS)",
		"migrationrepair",
	} {
		check(
			"Closure evidence "+token,
			strings.Contains(strings.ToLower(closure), strings.ToLower(token)),
			"PostgreSQL closure document is missing classification evidence: "+token,
		)
	}

	sort.Strings(result.checks)
	sort.Strings(result.violations)
	return result
}

func readText(root string, relative string, result *auditResult) string {
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil {
		result.violations = append(result.violations, relative+": "+err.Error())
		return ""
	}
	return string(content)
}

func readGoFiles(root string, result *auditResult) []string {
	return readGoFilesMatching(root, false, result)
}

func readProductionGoFiles(root string, result *auditResult) []string {
	return readGoFilesMatching(root, true, result)
}

func readGoFilesMatching(root string, productionOnly bool, result *auditResult) []string {
	values := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		if productionOnly && strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		values = append(values, string(content))
		return nil
	})
	if err != nil {
		result.violations = append(result.violations, root+": "+err.Error())
	}
	return values
}

func readSelectedTrees(root string, relatives []string, result *auditResult) []string {
	values := make([]string, 0)
	for _, relative := range relatives {
		values = append(values, readProductionGoFiles(filepath.Join(root, filepath.FromSlash(relative)), result)...)
	}
	return values
}

func containsAny(values []string, token string) bool {
	for _, value := range values {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func containsPattern(values []string, pattern string) bool {
	expression := regexp.MustCompile(pattern)
	for _, value := range values {
		if expression.MatchString(value) {
			return true
		}
	}
	return false
}

func productionGoIdentifierExists(
	root string,
	relatives []string,
	identifier string,
	result *auditResult,
) bool {
	for _, relative := range relatives {
		searchRoot := filepath.Join(root, filepath.FromSlash(relative))
		found := false
		err := filepath.WalkDir(searchRoot, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var lexicalScanner scanner.Scanner
			lexicalScanner.Init(
				token.NewFileSet().AddFile(path, -1, len(content)),
				content,
				nil,
				0,
			)
			for {
				_, lexicalToken, literal := lexicalScanner.Scan()
				if lexicalToken == token.EOF {
					break
				}
				if lexicalToken == token.IDENT && literal == identifier {
					found = true
					return nil
				}
			}
			return nil
		})
		if err != nil {
			result.violations = append(result.violations, searchRoot+": "+err.Error())
		}
		if found {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
