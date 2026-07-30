package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type requirement struct {
	path      string
	function  string
	fragments []string
	forbidden []string
}

func main() {
	strict := flag.Bool(
		"strict",
		false,
		"fail when a Projection Route Frequency review requirement is absent",
	)
	root := flag.String(
		"root",
		"",
		"optional path to the apps/api module root",
	)
	flag.Parse()

	apiRoot, err := resolveAPIRoot(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Projection route frequency review audit: %v\n", err)
		if *strict {
			os.Exit(1)
		}
		return
	}

	failures := inspectRequirements(apiRoot, reviewRequirements())
	if len(failures) == 0 {
		fmt.Println("Projection route frequency review audit: PASS")
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(os.Stderr, "Projection route frequency review audit: %s\n", failure)
	}
	if *strict {
		os.Exit(1)
	}
}

func reviewRequirements() []requirement {
	return []requirement{
		{
			path: "internal/projectionintelligence/projectionroutefrequency/model.go",
			fragments: []string{
				`Version            = "projection-low-frequency-route-guard-v3"`,
				`FingerprintVersion = "projection-low-frequency-route-fingerprint-v3"`,
				`summary.DistinctFlightCount > summary.ObservationCount`,
				`summary.DistinctDayCount > summary.ObservationCount`,
				`summary.DistinctDayCount > summary.DistinctFlightCount`,
				`summary.RecentObservationCount > summary.DistinctFlightCount`,
				`if len(summary.SourceNames) == 0 {`,
				`component.Weight <= 0`,
				`weightedScore += component.Score * component.Weight`,
				`route-frequency score does not match weighted components`,
				`normalizedLimitations := normalizeNotices(result.Limitations)`,
				`route-frequency limitations must be deterministically ordered`,
				`allowed route-frequency result must be usable and limitation-free`,
			},
			forbidden: []string{
				`projection-low-frequency-route-guard-v1`,
				`projection-low-frequency-route-guard-v2`,
				`projection-low-frequency-route-fingerprint-v1`,
				`projection-low-frequency-route-fingerprint-v2`,
				`component.Weight < 0`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionroutefrequency/config.go",
			fragments: []string{
				`HistoryWindow                 time.Duration`,
				`route distinct-day counts must satisfy zero < minimum <= target <= observation target`,
				`recent route observation counts must satisfy zero < minimum <= target <= observation target`,
				`recent route-history window must be positive and not exceed the full history window`,
				`route-frequency score thresholds must satisfy zero < minimum <= complete <= one`,
				`route-frequency component weights must be finite, positive, and sum to one`,
				`config.TargetDistinctDayCount >`,
				`config.TargetRecentObservationCount >`,
				`config.RecentWindow > config.HistoryWindow`,
				`config.MinimumRouteConfidenceScore <= 0`,
				`config.MinimumUsableScore <= 0`,
				`config.CompleteScoreMinimum <= 0`,
				`weight <= 0`,
			},
			forbidden: []string{
				`route-frequency component weights must be finite, non-negative, and sum to one`,
				`weight < 0`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionroutefrequency/evaluator.go",
			fragments: []string{
				`ErrRouteHistoryWindowMismatch`,
				`ErrRouteHistoryRecentWindowMismatch`,
				`expectedWindowStart := history.AsOfTime.UTC().Add(`,
				`-evaluator.config.HistoryWindow`,
				`expectedRecentWindowStart := history.AsOfTime.UTC().Add(`,
				`-evaluator.config.RecentWindow`,
				`history.DistinctFlightCount,`,
				`history.DistinctFlightCount <`,
				`route_confidence_below_minimum`,
				`route_observation_count_below_minimum`,
				`route_distinct_day_count_below_minimum`,
				`recent_route_observation_count_below_minimum`,
				`latest_route_observation_too_old`,
				`route_frequency_score_below_minimum`,
				`if err := result.Validate(); err != nil {`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionroutefrequency/fingerprint.go",
			fragments: []string{
				`writeFingerprintString(digest, FingerprintVersion)`,
				`writeFingerprintString(digest, string(route.SchemaVersion))`,
				`writeFingerprintString(digest, string(route.Status))`,
				`writeFingerprintString(digest, strings.TrimSpace(route.TrajectoryID))`,
				`writeFingerprintString(digest, strings.TrimSpace(route.FlightID))`,
				`writeFingerprintBool(digest, routeAvailable)`,
				`writeFingerprintString(digest, resolvedKey)`,
				`writeFingerprintBool(digest, route.Summary.SameAirport)`,
				`writeFingerprintFloat(digest, route.Confidence.Score)`,
				`strings.ToUpper(strings.TrimSpace(history.RouteKey))`,
				`writeFingerprintTime(digest, history.WindowStart)`,
				`writeFingerprintTime(digest, history.WindowEnd)`,
				`writeFingerprintTime(digest, history.RecentWindowStart)`,
				`config.HistoryWindow`,
				`config.RecentWindow`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/postgres_queries.go",
			fragments: []string{
				`route_result.trajectory_id::text <> $7`,
				`trajectory.flight_id::text <> $8`,
				`) AS evidence_id`,
				`latest_route_per_evidence AS (`,
				`SELECT DISTINCT ON (evidence_id)`,
				`as_of_time AT TIME ZONE 'UTC'`,
				`WHERE as_of_time >= $6`,
				`FROM latest_route_per_evidence;`,
			},
			forbidden: []string{
				`FROM latest_route_per_trajectory;`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/policy.go",
			fragments: []string{
				`HistoryWindow                 time.Duration`,
				`RouteHistoryWindow:              180 * 24 * time.Hour`,
				`RecentRouteWindow:               30 * 24 * time.Hour`,
				`HistoryWindow:                 180 * 24 * time.Hour`,
				`RecentWindow:                  30 * 24 * time.Hour`,
				`policy.RouteFrequency.HistoryWindow !=`,
				`policy.DataSource.RouteHistoryWindow`,
				`policy.RouteFrequency.RecentWindow !=`,
				`policy.DataSource.RecentRouteWindow`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionroutefrequency/decision_integrity_test.go",
			fragments: []string{
				`TestConfigValidateRejectsIncoherentPolicyTargets`,
				`TestEvaluateRejectsHistoryWindowMismatch`,
				`TestEvaluateReportsAllBlockingReasons`,
				`TestResultValidateRejectsWeightedScoreMismatch`,
				`TestResultValidateRejectsZeroComponentWeight`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/route_frequency_policy_integrity_test.go",
			fragments: []string{
				`TestPolicyValidateRequiresRouteFrequencyHistoryWindowAlignment`,
			},
		},
		{
			path:     "internal/projectionintelligence/projectionproduction/fixtures_test.go",
			function: "validProductionFrequency",
			fragments: []string{
				`Score: 0.94`,
			},
			forbidden: []string{
				`Score: 0.85`,
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			fragments: []string{
				"## 32. Projection Route Frequency Review Hardening",
				"PROJECTION_ROUTE_FREQUENCY_VERSION=projection-low-frequency-route-guard-v3",
				"PROJECTION_ROUTE_FREQUENCY_FINGERPRINT_VERSION=projection-low-frequency-route-fingerprint-v3",
				"PROJECTION_ROUTE_FREQUENCY_EVIDENCE_ISOLATION=ENFORCED",
				"PROJECTION_ROUTE_FREQUENCY_LOGICAL_FLIGHT_DEDUPLICATION=ENFORCED",
				"PROJECTION_ROUTE_FREQUENCY_FIXED_HISTORY_WINDOW=ENFORCED",
				"PROJECTION_ROUTE_FREQUENCY_ALL_HARD_VIOLATIONS=REPORTED",
				"PROJECTION_ROUTE_FREQUENCY_WEIGHTED_SCORE_RECONSTRUCTION=ENFORCED",
				"PROJECTION_ROUTE_FREQUENCY_EVIDENCE_ISOLATION_COMMIT=c6fff15f8d0c770197db40a69d54f8856044d8d2",
				"PROJECTION_ROUTE_FREQUENCY_POLICY_DECISION_INTEGRITY_COMMIT=ee7c79bc8213dc030ce0d98f13d1065c9bb96275",
				"PROJECTION_ROUTE_FREQUENCY_POLICY_DECISION_INTEGRITY_GITHUB_ACTIONS_RUN=PENDING",
				"PROJECTION_ROUTE_FREQUENCY_PERMANENT_AUDIT_COMMIT=PENDING",
				"PROJECTION_ROUTE_FREQUENCY_REVIEW_STATUS=OPEN",
			},
			forbidden: []string{
				"PROJECTION_ROUTE_FREQUENCY_REVIEW_STATUS=CLOSED",
			},
		},
		{
			path: "../../docs/140_PROJECTION_ROUTE_FREQUENCY_REVIEW_HARDENING.md",
			fragments: []string{
				"# Projection Route Frequency Review Hardening",
				"Status: open",
				"EVIDENCE_ISOLATION_COMMIT=c6fff15f8d0c770197db40a69d54f8856044d8d2",
				"EVIDENCE_ISOLATION_GITHUB_ACTIONS_RUN=30534243693",
				"POLICY_DECISION_INTEGRITY_COMMIT=ee7c79bc8213dc030ce0d98f13d1065c9bb96275",
				"POLICY_DECISION_INTEGRITY_GITHUB_ACTIONS_RUN=PENDING",
				"PERMANENT_AUDIT_COMMIT=PENDING",
				"OPEN_CONFIRMED_FINDINGS=3",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES",
				"PROJECTION_ROUTE_FREQUENCY_REVIEW_STATUS=OPEN",
				"apps/api/tools/projectionroutefrequencyreviewaudit",
				"Run projection route frequency review audit",
			},
			forbidden: []string{
				"Status: closed",
				"OPEN_CONFIRMED_FINDINGS=0",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO",
				"PROJECTION_ROUTE_FREQUENCY_REVIEW_STATUS=CLOSED",
			},
		},
		{
			path: "../../docs/DOCUMENT_INDEX.md",
			fragments: []string{
				"## Document 140 — Projection Route Frequency Review Hardening",
				"140_PROJECTION_ROUTE_FREQUENCY_REVIEW_HARDENING.md",
				"current-flight evidence isolation",
				"logical-flight deduplication",
				"fixed full and recent exposure windows",
				"review remains open pending exact Continuous Integration and formal closure",
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				`- name: Run projection route frequency review audit`,
				`run: go run ./tools/projectionroutefrequencyreviewaudit -strict`,
			},
		},
	}
}

func inspectRequirements(root string, requirements []requirement) []string {
	failures := make([]string, 0)
	for _, item := range requirements {
		path := filepath.Clean(filepath.Join(root, filepath.FromSlash(item.path)))
		content, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("read %s: %v", item.path, err))
			continue
		}
		text, err := requirementText(path, content, item.function)
		if err != nil {
			failures = append(failures, fmt.Sprintf("inspect %s: %v", requirementLabel(item), err))
			continue
		}
		normalizedText := normalizeWhitespace(text)
		for _, fragment := range item.fragments {
			if !strings.Contains(
				normalizedText,
				normalizeWhitespace(fragment),
			) {
				failures = append(
					failures,
					fmt.Sprintf("%s is missing required fragment %q", requirementLabel(item), fragment),
				)
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(
				normalizedText,
				normalizeWhitespace(fragment),
			) {
				failures = append(
					failures,
					fmt.Sprintf("%s contains forbidden fragment %q", requirementLabel(item), fragment),
				)
			}
		}
	}
	sort.Strings(failures)
	return failures
}

func requirementText(path string, content []byte, functionName string) (string, error) {
	if filepath.Ext(path) != ".go" {
		if strings.TrimSpace(functionName) != "" {
			return "", fmt.Errorf("function scope is only supported for Go files")
		}
		return string(content), nil
	}

	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, path, content, parser.AllErrors)
	if err != nil {
		return "", fmt.Errorf("parse Go source: %w", err)
	}
	if strings.TrimSpace(functionName) == "" {
		return string(content), nil
	}

	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name == nil || function.Name.Name != functionName {
			continue
		}
		var rendered bytes.Buffer
		if err := printer.Fprint(&rendered, fileSet, function); err != nil {
			return "", fmt.Errorf("render function %s: %w", functionName, err)
		}
		return rendered.String(), nil
	}
	return "", fmt.Errorf("function %s was not found", functionName)
}

func requirementLabel(item requirement) string {
	if strings.TrimSpace(item.function) == "" {
		return item.path
	}
	return item.path + "#" + item.function
}

func normalizeWhitespace(value string) string {
	normalized := strings.Join(strings.Fields(value), "")
	return strings.ReplaceAll(normalized, ",)", ")")
}

func resolveAPIRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return validateAPIRoot(explicit)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	current := workingDirectory
	for {
		if root, err := validateAPIRoot(current); err == nil {
			return root, nil
		}
		candidate := filepath.Join(current, "apps", "api")
		if root, err := validateAPIRoot(candidate); err == nil {
			return root, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", fmt.Errorf("apps/api module root was not found from %s", workingDirectory)
}

func validateAPIRoot(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve API root: %w", err)
	}
	if _, err := os.Stat(filepath.Join(absolute, "go.mod")); err != nil {
		return "", fmt.Errorf("API root %s does not contain go.mod", absolute)
	}
	if _, err := os.Stat(filepath.Join(absolute, "internal", "projectionintelligence")); err != nil {
		return "", fmt.Errorf("API root %s does not contain Projection Intelligence", absolute)
	}
	return absolute, nil
}
