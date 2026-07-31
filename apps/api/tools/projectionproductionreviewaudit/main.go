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
	strict := flag.Bool("strict", false, "fail when a Projection Production review requirement is absent")
	root := flag.String("root", "", "optional path to the apps/api module root")
	flag.Parse()

	apiRoot, err := resolveAPIRoot(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Projection production review audit: %v\n", err)
		if *strict {
			os.Exit(1)
		}
		return
	}

	failures := inspectRequirements(apiRoot, reviewRequirements())
	if len(failures) == 0 {
		fmt.Println("Projection production review audit: PASS")
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(os.Stderr, "Projection production review audit: %s\n", failure)
	}
	if *strict {
		os.Exit(1)
	}
}

func reviewRequirements() []requirement {
	return []requirement{
		{
			path: "internal/projectionintelligence/projectionbaseline/baseline.go",
			fragments: []string{
				"ProjectWithPlan(",
				"request does not match authorized plan",
				"if err := plan.Validate(); err != nil",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/continuation_core.go",
			fragments: []string{
				"ProjectApprovedWithPlan(",
				"projectWithPlan(",
				"request does not match authorized plan",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/continuation_fallback.go",
			fragments: []string{
				"FallbackProjector.ProjectWithPlan(",
				`"%w: %w"`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/config.go",
			fragments: []string{
				"ProjectWithPlan(",
				"ProjectApprovedWithPlan(",
				"EstimateArrival(projectionarrival.Request)",
				"only an Estimated Arrival delta",
			},
			forbidden: []string{
				"Estimate(\n\t\tprojectionarrival.Request",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/request_snapshot.go",
			fragments: []string{
				"func (request Request) Clone() Request",
				"request.CurrentTrajectory.Clone()",
				"cloneTrajectories(request.HistoricalCandidates)",
				"request.Route.Clone()",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/evidence.go",
			fragments: []string{
				"type AuthorizedHistoricalEvidence struct",
				"selection does not match production request and plan",
				"authorized pattern does not match selection",
				"authorized freshness does not match selection, pattern, or plan",
				"authorized route-frequency evidence does not match route history and plan",
				"route identity does not match current trajectory",
				"route evidence exceeds production time boundary",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/composer.go",
			fragments: []string{
				"snapshot := request.Clone()",
				"ProjectApprovedWithPlan(",
				"KinematicProjector.ProjectWithPlan(",
				"historical_projection_authorized_with_limited_freshness",
				"historical_projection_authorized_with_limited_route_frequency",
				"requestInputFingerprint(",
				"Finalize(result)",
				`fmt.Errorf("%w: %w", sentinel, cause)`,
			},
			forbidden: []string{
				"HistoricalProjector.ProjectApproved(",
				"KinematicProjector.Project(",
				"ArrivalEstimator.Estimate(",
				`"%w: %v"`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/arrival_adapter.go",
			fragments: []string{
				"type ArrivalOutcome struct",
				"ErrArrivalProjectionMutation",
				"sameProjectionExceptArrival(before, result)",
				"return ArrivalOutcome{}, ErrArrivalProjectionMutation",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/projection_postconditions.go",
			fragments: []string{
				"projection identity does not match current trajectory",
				"projection time boundary does not match authorized plan",
				"historical projector postconditions failed",
				"ResultStatusUnavailable",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/model.go",
			fragments: []string{
				`Version                       = "projection-production-composition-v2"`,
				`RequestFingerprintVersion     = "projection-production-request-fingerprint-v2"`,
				`CompositionFingerprintVersion = "projection-production-composition-fingerprint-v2"`,
				"HorizonPlan projectionhorizon.Plan",
				"CompositionFingerprint string",
				"result.CompositionFingerprint != compositionFingerprint(result)",
				"result.Projection.Status == projectioncontract.ResultStatusUnavailable",
			},
			forbidden: []string{
				"projection-production-composition-v1",
				"projection-production-composition-fingerprint-v1",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/fingerprint.go",
			fragments: []string{
				"func requestInputFingerprint(",
				"Request:                     request.Clone()",
				"func compositionFingerprint(result Result)",
				"cloned.CompositionFingerprint = \"\"",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/integration_contract_test.go",
			fragments: []string{
				"_ KinematicProjector         = (*projectionbaseline.Baseline)(nil)",
				"_ HistoricalProjector        = (*projectioncontinuation.Baseline)(nil)",
				"_ ArrivalProjectionEstimator = (*projectionarrival.Estimator)(nil)",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/review_hardening_test.go",
			fragments: []string{
				"TestComposeUsesOneAuthorizedHorizonPlan",
				"TestComposeRejectsRouteFromAnotherTrajectory",
				"TestComposeRejectsFutureRouteEvidence",
				"TestComposeBindsSelectionPatternAndFreshness",
				"TestComposeBindsRouteHistoryAndFrequencyToPlan",
				"TestComposeRejectsHistoricalProjectionPostconditionDrift",
				"TestComposeRejectsUnavailableHistoricalProjection",
				"TestComposeDefensivelyClonesDependencyInputs",
				"TestComposePreservesUnderlyingDependencyError",
				"TestComposePublishesLimitedEvidenceNotice",
				"TestRequestFingerprintCoversCandidatesOnFallback",
				"TestCompositionFingerprintBindsPublishedProjection",
				"TestArrivalAdapterRejectsPositionProjectionMutation",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/composition.go",
			fragments: []string{
				"projectionproduction.NewArrivalAdapter(arrival)",
				"ArrivalEstimator:           arrivalAdapter",
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				"Run projection production review audit",
				"go run ./tools/projectionproductionreviewaudit -strict",
			},
		},
		{
			path: "../../docs/144_PROJECTION_PRODUCTION_REVIEW_HARDENING.md",
			fragments: []string{
				"Status: engineering implementation complete, exact Continuous Integration and formal closure pending",
				"REVIEW_BASELINE_COMMIT=298d3fdb2d11b1797ce3728b116702b0a978d870",
				"SINGLE_HORIZON_PLAN=IMPLEMENTED_PENDING_EXACT_CI",
				"CROSS_CONTRACT_EVIDENCE_BINDING=IMPLEMENTED_PENDING_EXACT_CI",
				"ARRIVAL_ONLY_MUTATION_BOUNDARY=IMPLEMENTED_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_REVIEW_STATUS=OPEN_PENDING_EXACT_CI_AND_FORMAL_CLOSURE",
			},
		},
	}
}

func inspectRequirements(apiRoot string, requirements []requirement) []string {
	failures := make([]string, 0)
	for _, requirement := range requirements {
		path := filepath.Clean(filepath.Join(apiRoot, requirement.path))
		content, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("read %s: %v", requirement.path, err))
			continue
		}
		text := string(content)
		for _, fragment := range requirement.fragments {
			if !strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s is missing %q", requirement.path, fragment))
			}
		}
		for _, forbidden := range requirement.forbidden {
			if strings.Contains(text, forbidden) {
				failures = append(failures, fmt.Sprintf("%s contains forbidden %q", requirement.path, forbidden))
			}
		}
	}
	return failures
}

func resolveAPIRoot(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Abs(explicit)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	current := workingDirectory
	for {
		if fileExists(filepath.Join(current, "go.mod")) &&
			dirExists(filepath.Join(current, "internal", "projectionintelligence")) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("apps/api root not found from %s", workingDirectory)
		}
		current = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
