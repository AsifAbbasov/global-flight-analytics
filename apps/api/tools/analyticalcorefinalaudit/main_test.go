package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyticalCoreFindingRegisterIsComplete(
	t *testing.T,
) {
	if len(analyticalCoreFindings) != 19 {
		t.Fatalf(
			"finding count = %d, want 19",
			len(analyticalCoreFindings),
		)
	}

	counts := map[findingDisposition]int{}
	for index, item := range analyticalCoreFindings {
		expectedID := "AC-" +
			twoDigit(index+1)
		if item.ID != expectedID {
			t.Fatalf(
				"finding %d has id %s, want %s",
				index,
				item.ID,
				expectedID,
			)
		}
		counts[item.Disposition]++
	}

	if counts[dispositionFixed] != 14 {
		t.Fatalf(
			"fixed count = %d, want 14",
			counts[dispositionFixed],
		)
	}
	if counts[dispositionRetained] != 3 {
		t.Fatalf(
			"retained count = %d, want 3",
			counts[dispositionRetained],
		)
	}
	if counts[dispositionRejected] != 2 {
		t.Fatalf(
			"rejected count = %d, want 2",
			counts[dispositionRejected],
		)
	}
}

func TestAuditRepositoryAcceptsCurrentRepository(
	t *testing.T,
) {
	root, err := resolveRepositoryRoot("")
	if err != nil {
		t.Skipf(
			"repository fixture is not available: %v",
			err,
		)
	}

	failures := auditRepository(
		root,
		&bytes.Buffer{},
	)
	if len(failures) != 0 {
		t.Fatalf(
			"unexpected Analytical Core audit failures: %#v",
			failures,
		)
	}
}

func TestContributorOrderingDetectsPreparationBeforeFilter(
	t *testing.T,
) {
	root := t.TempDir()
	path := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"analytics",
		"metricexecution",
		"execution.go",
	)
	writeFixture(
		t,
		path,
		`package metricexecution
func broken() {
	if preparation != nil {
		preparation(filtered.Allowed)
	}
	FilterTrajectories()
}
`,
	)

	failures := auditContributorOrdering(root)
	if !containsFailure(
		failures,
		"preparation appears before capability filtering",
	) {
		t.Fatalf(
			"ordering defect was not detected: %#v",
			failures,
		)
	}
}

func TestAuditRulesAcceptFormatterLineBreaks(
	t *testing.T,
) {
	root := t.TempDir()

	writeFixture(
		t,
		filepath.Join(
			root,
			"apps/api/internal/http/handlers/analytical_production_snapshot.go",
		),
		`package handlers
var interval = dataqualityintegration.
	DefaultExpectedObservationInterval
`,
	)
	writeFixture(
		t,
		filepath.Join(
			root,
			"apps/api/internal/http/handlers/analytical_metrics.go",
		),
		`package handlers
var staleAfter = dataqualityintegration.
	DefaultStaleAfter
`,
	)
	writeFixture(
		t,
		filepath.Join(
			root,
			"apps/web/lib/api/analytics.ts",
		),
		`searchParameters.set(
  'radius_kilometers',
  String(parameters.radiusKilometers)
)
`,
	)

	failures := auditRules(
		root,
		[]fileRule{
			{
				Name: "sampling interval",
				Path: "apps/api/internal/http/handlers/analytical_production_snapshot.go",
				RequiredCompact: []string{
					"dataqualityintegration.DefaultExpectedObservationInterval",
				},
			},
			{
				Name: "freshness threshold",
				Path: "apps/api/internal/http/handlers/analytical_metrics.go",
				RequiredCompact: []string{
					"dataqualityintegration.DefaultStaleAfter",
				},
			},
			{
				Name: "frontend radius parameter",
				Path: "apps/web/lib/api/analytics.ts",
				RequiredCompact: []string{
					"searchParameters.set('radius_kilometers'",
				},
			},
		},
	)
	if len(failures) != 0 {
		t.Fatalf(
			"formatter line breaks caused false failures: %#v",
			failures,
		)
	}
}

func TestAuditRulesDetectMissingCompactContract(
	t *testing.T,
) {
	root := t.TempDir()
	writeFixture(
		t,
		filepath.Join(
			root,
			"apps/web/lib/api/analytics.ts",
		),
		"const unrelated = true\n",
	)

	failures := auditRules(
		root,
		[]fileRule{
			{
				Name: "frontend radius parameter",
				Path: "apps/web/lib/api/analytics.ts",
				RequiredCompact: []string{
					"searchParameters.set('radius_kilometers'",
				},
			},
		},
	)
	if !containsFailure(
		failures,
		"missing required compact fragment",
	) {
		t.Fatalf(
			"missing compact contract was not detected: %#v",
			failures,
		)
	}
}

func TestStandardSortingDetectsManualQuadraticHelper(
	t *testing.T,
) {
	root := t.TempDir()
	path := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"http",
		"handlers",
		"analytical_metrics.go",
	)
	writeFixture(
		t,
		path,
		`package handlers
func sortStrings(values []string) {
	for left := 0; left < len(values); left++ {
		for right := left + 1; right < len(values); right++ {
		}
	}
}
`,
	)

	failures := auditStandardSourceSorting(root)
	if !containsFailure(
		failures,
		"func sortStrings(",
	) {
		t.Fatalf(
			"manual sort was not detected: %#v",
			failures,
		)
	}
}

func TestRunAllowsNonStrictReporting(
	t *testing.T,
) {
	root := t.TempDir()
	writeFixture(
		t,
		filepath.Join(root, "apps", "api", "go.mod"),
		"module example.com/audit\n\ngo 1.26\n",
	)
	writeFixture(
		t,
		filepath.Join(root, "apps", "web", "package.json"),
		"{}\n",
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{
			"-root",
			root,
			"-strict=false",
		},
		&stdout,
		&stderr,
	)

	if exitCode != 0 {
		t.Fatalf(
			"non-strict exit code = %d, stderr=%s",
			exitCode,
			stderr.String(),
		)
	}
	if !strings.Contains(
		stderr.String(),
		"Analytical Core final source audit: FAIL",
	) {
		t.Fatalf(
			"expected failure report, got stderr=%q",
			stderr.String(),
		)
	}
}

func twoDigit(
	value int,
) string {
	return fmt.Sprintf("%02d", value)
}

func writeFixture(
	t *testing.T,
	path string,
	content string,
) {
	t.Helper()

	if err := os.MkdirAll(
		filepath.Dir(path),
		0o755,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		path,
		[]byte(content),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
}

func containsFailure(
	failures []auditFailure,
	fragment string,
) bool {
	for _, failure := range failures {
		if strings.Contains(
			failure.Detail,
			fragment,
		) {
			return true
		}
	}
	return false
}
