package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const auditVersion = "backend-observability-audit-v1"

type finding struct {
	path    string
	message string
}

func main() {
	strict := flag.Bool("strict", false, "return a non-zero status when findings exist")
	flag.Parse()

	root, err := repositoryRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "locate repository root: %v\n", err)
		os.Exit(1)
	}

	findings := runAudit(root)
	fmt.Printf("Audit version: %s\n", auditVersion)
	fmt.Printf("Observability findings: %d\n", len(findings))
	if len(findings) == 0 {
		fmt.Println("Backend observability audit: PASS")
		return
	}

	fmt.Println("Backend observability audit: FAIL")
	for _, current := range findings {
		fmt.Printf("- %s: %s\n", current.path, current.message)
	}
	if *strict {
		os.Exit(1)
	}
}

func runAudit(
	root string,
) []finding {
	requirements := []struct {
		path     string
		literals []string
	}{
		{
			path: "apps/api/internal/observability/registry.go",
			literals: []string{
				"global_flight_analytics",
				"ObserveProviderRequest",
				"ObserveIngestionCycle",
				"normalizeRoute",
			},
		},
		{
			path: "apps/api/internal/observability/http.go",
			literals: []string{
				"METRICS_AUTHENTICATION_REQUIRED",
				"ExpectedDigest.MatchesCandidate",
				"ctx.Route()",
			},
		},
		{
			path: "apps/api/internal/observability/postgres_collector.go",
			literals: []string{
				"ingestion_runs",
				"derived_reconciliation_tasks",
				"postgres_pool_acquired_connections",
			},
		},
		{
			path: "apps/api/internal/server/server.go",
			literals: []string{
				"observability.HTTPMiddleware",
				"observability.MetricsPath",
				"observability.FiberHandler",
			},
		},
		{
			path: "apps/api/cmd/ingest/main.go",
			literals: []string{
				"observability.StartMetricsServer",
				"observability.NewProviderRecorder",
			},
		},
		{
			path: "apps/api/cmd/ingest/cycle_observer.go",
			literals: []string{
				"metricsRegistry.ObserveIngestionCycle",
			},
		},
		{
			path: "docs/168_BACKEND_OBSERVABILITY_AND_SLO_CLOSURE.md",
			literals: []string{
				"99.5%",
				"p95",
				"high-cardinality",
				"reconciliation",
			},
		},
		{
			path: ".github/workflows/backend-ci.yml",
			literals: []string{
				"Run backend observability audit",
				"go run ./tools/backendobservabilityaudit -strict",
			},
		},
		{
			path: "scripts/verify-release.sh",
			literals: []string{
				"backendobservabilityaudit",
			},
		},
	}

	findings := make([]finding, 0)
	for _, requirement := range requirements {
		content, err := os.ReadFile(filepath.Join(root, requirement.path))
		if err != nil {
			findings = append(findings, finding{
				path:    requirement.path,
				message: "required observability file is missing",
			})
			continue
		}
		text := string(content)
		for _, literal := range requirement.literals {
			if !strings.Contains(text, literal) {
				findings = append(findings, finding{
					path:    requirement.path,
					message: fmt.Sprintf("required literal %q is missing", literal),
				})
			}
		}
	}

	registryPath := "apps/api/internal/observability/exposition.go"
	registryContent, err := os.ReadFile(filepath.Join(root, registryPath))
	if err == nil {
		text := string(registryContent)
		for _, forbidden := range []string{
			`name: "request_id"`,
			`name: "ip"`,
			`name: "icao24"`,
			`name: "trajectory_id"`,
			`name: "error"`,
			`name: "url"`,
		} {
			if strings.Contains(text, forbidden) {
				findings = append(findings, finding{
					path:    registryPath,
					message: fmt.Sprintf("forbidden high-cardinality or sensitive label %s", forbidden),
				})
			}
		}
	}

	return findings
}

func repositoryRoot() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := workingDirectory
	for {
		if _, err := os.Stat(filepath.Join(current, "apps", "api", "go.mod")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root was not found from %s", workingDirectory)
		}
		current = parent
	}
}
