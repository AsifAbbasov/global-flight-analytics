package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	strict := flag.Bool("strict", false, "fail on any historical contract review violation")
	flag.Parse()
	root, err := repositoryRoot()
	if err != nil {
		fail(err)
	}
	checks := []struct {
		path     string
		contains []string
		excludes []string
	}{
		{path: "apps/api/internal/historicalintelligence/historicalcontract/model.go", contains: []string{"historical-intelligence-contract-v2", "MetricNamePeakActivity"}, excludes: []string{"var supportedMetricNames"}},
		{path: "apps/api/internal/historicalintelligence/historicalcontract/metric_catalog.go", contains: []string{"type MetricSpec struct", "MetricValueKindCount", "MaximumExactCountValue", "MetricSpecFor"}},
		{path: "apps/api/internal/historicalintelligence/historicalcontract/historical_validation_core.go", contains: []string{"historical-intelligence-contract-validation-v2", "validateMetric(result.Metric, result.Scope"}},
		{path: "apps/api/internal/historicalintelligence/historicalcontract/historical_validation_identity.go", contains: []string{"metric_scope_mismatch", "metric_unit_mismatch", "MetricSpecFor"}},
		{path: "apps/api/internal/historicalintelligence/historicalcontract/historical_validation_series.go", contains: []string{"partial_series_without_represented_bucket", "unavailable_bucket_confidence_invalid", "partial_bucket_without_limitation"}},
		{path: "apps/api/internal/historicalintelligence/historicalcontract/historical_validation_summary.go", contains: []string{"comparison_current_summary_mismatch", "comparisonValueForMetric"}},
		{path: "apps/api/internal/historicalintelligence/historicalcontract/historical_validation_evidence.go", contains: []string{"confidence_reason_contribution_mismatch", "contributionCompensation"}},
		{path: "apps/api/internal/historicalintelligence/historicalcontract/schema.go", contains: []string{"schema_version", "points.confidence", "comparison.current_value", "provenance.source_names"}},
		{path: "apps/api/internal/historicalintelligence/historicalseries/builder.go", contains: []string{"representedPointCount", "historical_data_unavailable"}},
		{path: "apps/api/internal/historicalintelligence/historicalseries/builder_test.go", contains: []string{"TestBuildZeroCoverageSeriesIsUnavailable"}},
		{path: "apps/api/internal/historicalintelligence/historicalmaterialization/materializer.go", contains: []string{"MetricSpecFor", "metricScopeAllowed"}, excludes: []string{"func scopeAllowed("}},
		{path: "apps/api/internal/historicalintelligence/historicalairport/builder.go", contains: []string{"MetricSpecFor"}, excludes: []string{"airportMetricDefinition", "type airportMetricSpec"}},
		{path: "apps/api/internal/historicalintelligence/historicalroute/builder.go", contains: []string{"MetricSpecFor"}, excludes: []string{"routeMetricDefinition", "type routeMetricSpec"}},
		{path: "apps/api/internal/historicalintelligence/historicalaggregate/helpers.go", contains: []string{"RegionCode: strings.ToLower", "normalized.Scope.RegionCode = strings.ToLower"}},
		{path: "apps/api/internal/http/handlers/historical_intelligence.go", contains: []string{"MetricSpecFor", "specification.AllowsScope", "regionCode := strings.ToLower"}},
		{path: "apps/api/cmd/verify-postgres-historical-http-api/fixture.go", contains: []string{"MetricNameRouteObservations", "Unit:        \"route_results\""}, excludes: []string{"Unit:        \"routes\""}},
		{path: "apps/api/internal/historicalintelligence/historicalcontract/historicalcontract_review_hardening_test.go", contains: []string{"TestValidateRejectsFractionalCountValue", "TestValidateAcceptsZeroEventCompleteBucket", "TestValidateBindsComparisonToCurrentSummary"}},
		{path: ".github/workflows/backend-ci.yml", contains: []string{"go run ./tools/historicalcontractreviewaudit -strict"}},
		{path: "docs/124_HISTORICAL_CONTRACT_REVIEW_HARDENING.md", contains: []string{"OPEN_CONFIRMED_FINDINGS=0", "UNCLASSIFIED_FINDINGS=0", "DEFERRED_FINDINGS=0"}},
	}
	violations := 0
	for _, check := range checks {
		data, readErr := os.ReadFile(filepath.Join(root, check.path))
		if readErr != nil {
			fmt.Printf("FAIL %s: %v\n", check.path, readErr)
			violations++
			continue
		}
		value := string(data)
		for _, expected := range check.contains {
			if !strings.Contains(value, expected) {
				fmt.Printf("FAIL %s: missing %q\n", check.path, expected)
				violations++
			}
		}
		for _, forbidden := range check.excludes {
			if strings.Contains(value, forbidden) {
				fmt.Printf("FAIL %s: contains forbidden %q\n", check.path, forbidden)
				violations++
			}
		}
	}
	if violations > 0 {
		fmt.Printf("Historical contract review audit: FAIL (%d violations)\n", violations)
		if *strict {
			os.Exit(1)
		}
		return
	}
	fmt.Println("Historical contract review audit: PASS")
}

func repositoryRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "apps", "api", "go.mod")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root not found")
		}
		current = parent
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
