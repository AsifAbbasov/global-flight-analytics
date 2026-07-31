package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type requirement struct {
	path      string
	fragments []string
	forbidden []string
}

func main() {
	strict := flag.Bool(
		"strict",
		false,
		"fail when a Projection Read review requirement is absent",
	)
	root := flag.String(
		"root",
		"",
		"optional path to the apps/api module root",
	)
	flag.Parse()

	apiRoot, err := resolveAPIRoot(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Projection read review audit: %v\n", err)
		if *strict {
			os.Exit(1)
		}
		return
	}

	failures := inspectRequirements(apiRoot, reviewRequirements())
	if len(failures) == 0 {
		fmt.Println("Projection read review audit: PASS")
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(os.Stderr, "Projection read review audit: %s\n", failure)
	}
	if *strict {
		os.Exit(1)
	}
}

func reviewRequirements() []requirement {
	return []requirement{
		{
			path: "internal/projectionintelligence/projectionread/contracts.go",
			fragments: []string{
				"LoadSnapshot(",
				"context.Context,",
				"SnapshotRequest,",
				") (Snapshot, error)",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/postgres_snapshot.go",
			fragments: []string{
				"IsoLevel:   pgx.RepeatableRead",
				"AccessMode: pgx.ReadOnly",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/context.go",
			fragments: []string{
				"return ErrContextRequired",
				"return ctx.Err()",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/service.go",
			fragments: []string{
				"validateReadContext(ctx)",
				"canonicalRequestedDuration := request.RequestedDuration",
				"service.policy.Horizon.DefaultDuration",
				"service.dataSource.LoadSnapshot(",
				"validateSnapshotPostconditions(",
				"generatedAt := service.now().UTC()",
				"validateComposedResult(",
			},
			forbidden: []string{
				"context.Background()",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/postconditions.go",
			fragments: []string{
				"func validateSnapshotPostconditions(",
				"routecontract.Validate(route)",
				"historical candidate does not belong to the authorized snapshot",
				"func validateComposedResult(",
				"if err := result.Validate(); err != nil",
				"composed projection identity does not match the snapshot trajectory",
				"composed projection does not match the requested as-of time",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/route_snapshot.go",
			fragments: []string{
				"rowAsOfTimeUnixNano",
				"rowInputFingerprint",
				"rowRouteStatus",
				"routecontract.Validate(result)",
				"route payload does not match persisted row metadata",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/postgres_queries.go",
			fragments: []string{
				"as_of_time_unix_nano",
				"input_fingerprint",
				"route_status",
				"latitude IS NOT NULL",
				"longitude IS NOT NULL",
				"velocity_mps IS NOT NULL",
				"heading_degrees IS NOT NULL",
				"vertical_rate_mps IS NOT NULL",
				"on_ground IS NOT NULL",
				"route_result.trajectory_id::text <> $7",
				"trajectory.flight_id::text <> $8",
				"route_record_id",
				"input_fingerprint",
				"as_of_time_unix_nano",
				"jsonb_agg(",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/historical_candidate_backfill.go",
			fragments: []string{
				"const historicalCandidateScanMultiplier = 4",
				"math.MaxInt/historicalCandidateScanMultiplier",
				"return maximumCandidateCount *",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/postgres_source.go",
			fragments: []string{
				"historicalCandidateScanLimit(",
				"MaximumHistoricalCandidateCount",
				"if len(result) ==",
				"decodeRouteHistoryEvidence(",
				"routeHistoryEvidenceFingerprint(evidence)",
				"routeHistoryFingerprint(",
				"item.UpdatedAt.Before(item.EndTime)",
			},
			forbidden: []string{
				"func nonNilContext(",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/route_history_lineage.go",
			fragments: []string{
				`routeHistoryEvidenceFingerprintVersion = "projection-route-history-evidence-v1"`,
				"type routeHistoryEvidence struct",
				"RouteRecordID",
				"InputFingerprint",
				"AsOfTimeUnixNano",
				"route-history evidence aggregate mirrors are invalid",
				"sortRouteHistoryEvidence(items)",
				"func routeHistoryEvidenceFingerprint(",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/service_postconditions_test.go",
			fragments: []string{
				"TestServiceRejectsInvalidComposerOutput",
				"TestServiceRejectsValidResultForAnotherTrajectory",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/route_snapshot_test.go",
			fragments: []string{
				"TestLoadRouteRejectsPersistedFingerprintMismatch",
				"TestLoadRouteRejectsInvalidPayloadContract",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/residual_hardening_test.go",
			fragments: []string{
				"TestLoadHistoricalCandidatesBackfillsRejectedIdentifiers",
				"TestRouteHistoryFingerprintBindsContributingRecords",
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				"Run projection read review audit",
				"go run ./tools/projectionreadreviewaudit -strict",
			},
		},
		{
			path: "../../docs/145_PROJECTION_READ_REVIEW_HARDENING.md",
			fragments: []string{
				"# Projection Read Review Hardening",
				"Status: engineering complete; formal closure pending permanent-audit exact Continuous Integration",
				"CONTRACT_HARDENING_COMMIT=4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37",
				"CONTRACT_HARDENING_GITHUB_ACTIONS_RUN=30638188394",
				"EVIDENCE_HARDENING_COMMIT=9dda4b102497028b59280143b86bf84564afb136",
				"EVIDENCE_HARDENING_GITHUB_ACTIONS_RUN=30648605652",
				"ATOMIC_REPEATABLE_READ_SNAPSHOT=CI_CONFIRMED",
				"COMPOSER_OUTPUT_POSTCONDITIONS=CI_CONFIRMED",
				"ROUTE_ROW_PAYLOAD_BINDING=CI_CONFIRMED",
				"HISTORICAL_CANDIDATE_BACKFILL=CI_CONFIRMED",
				"ROUTE_HISTORY_RECORD_LINEAGE=CI_CONFIRMED",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
				"PERMANENT_AUDIT_COMMIT=PENDING",
				"PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=PENDING",
				"PROJECTION_READ_FORMAL_CLOSURE=OPEN_PENDING_EXACT_CI",
				"PROJECTION_READ_REVIEW_STATUS=OPEN_PENDING_EXACT_CI_AND_FORMAL_CLOSURE",
			},
			forbidden: []string{
				"PROJECTION_READ_FORMAL_CLOSURE=COMPLETE",
				"PROJECTION_READ_REVIEW_STATUS=CLOSED",
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			fragments: []string{
				"## 39. Projection Read Engineering Completion and Permanent Audit",
				"PROJECTION_READ_CONTRACT_HARDENING_COMMIT=4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37",
				"PROJECTION_READ_CONTRACT_HARDENING_GITHUB_ACTIONS_RUN=30638188394",
				"PROJECTION_READ_EVIDENCE_HARDENING_COMMIT=9dda4b102497028b59280143b86bf84564afb136",
				"PROJECTION_READ_EVIDENCE_HARDENING_GITHUB_ACTIONS_RUN=30648605652",
				"PROJECTION_READ_PERMANENT_REVIEW_AUDIT=IMPLEMENTED_PENDING_EXACT_CI",
				"PROJECTION_READ_ENGINEERING_DEBT=CLOSED",
				"PROJECTION_READ_OPEN_CONFIRMED_FINDINGS=0",
				"PROJECTION_READ_FORMAL_CLOSURE=OPEN_PENDING_EXACT_CI",
				"PROJECTION_READ_REVIEW_STATUS=OPEN_PENDING_EXACT_CI_AND_FORMAL_CLOSURE",
			},
		},
		{
			path: "../../docs/DOCUMENT_INDEX.md",
			fragments: []string{
				"## Document 145 — Projection Read Review Hardening",
				"atomic repeatable-read snapshot",
				"route-row payload metadata binding",
				"historical candidate backfill",
				"record-level route-history lineage",
				"formal closure pending permanent-audit exact Continuous Integration",
			},
		},
	}
}

func inspectRequirements(
	apiRoot string,
	requirements []requirement,
) []string {
	failures := make([]string, 0)
	for _, item := range requirements {
		path := filepath.Clean(filepath.Join(apiRoot, item.path))
		content, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("read %s: %v", item.path, err))
			continue
		}
		text := string(content)
		for _, fragment := range item.fragments {
			if !strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s is missing %q", item.path, fragment),
				)
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s contains forbidden %q", item.path, fragment),
				)
			}
		}
	}
	return failures
}

func resolveAPIRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return validateAPIRoot(explicit)
	}

	current, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve current directory: %w", err)
	}
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
	return "", fmt.Errorf("apps/api module root was not found")
}

func validateAPIRoot(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve API root: %w", err)
	}
	info, err := os.Stat(filepath.Join(absolute, "go.mod"))
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("%s is not the apps/api module root", absolute)
	}
	return absolute, nil
}
