package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
		"fail when a Projection Horizon review requirement is absent",
	)
	root := flag.String(
		"root",
		"",
		"optional path to the apps/api module root",
	)
	flag.Parse()

	apiRoot, err := resolveAPIRoot(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Projection horizon review audit: %v\n", err)
		if *strict {
			os.Exit(1)
		}
		return
	}

	failures := inspectRequirements(apiRoot, reviewRequirements())
	if len(failures) == 0 {
		fmt.Println("Projection horizon review audit: PASS")
		return
	}

	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Projection horizon review audit: %s\n",
			failure,
		)
	}
	if *strict {
		os.Exit(1)
	}
}

func reviewRequirements() []requirement {
	return []requirement{
		{
			path: "internal/projectionintelligence/projectionhorizon/policy.go",
			fragments: []string{
				`Version                    = "projection-horizon-policy-v2"`,
				"MaximumSupportedPointCount = 10000",
				"func (config Config) Validate() error",
				"normalizedName := strings.TrimSpace(config.Name)",
				"config.MinimumDuration%config.Step != 0",
				"config.DefaultDuration%config.Step != 0",
				"config.MaximumDuration%config.Step != 0",
				"requiredPointCountValue := config.MaximumDuration / config.Step",
				"func (config Config) ResolveRequestedDuration(",
				"ErrRequestedDurationGridInvalid",
				"func (policy *Policy) BuildDefault(",
				"return Plan{}, ErrPolicyUnavailable",
				"pointCount := exactForecastPointCount(duration, step)",
				"for index := range result",
			},
			forbidden: []string{
				`Version = "projection-horizon-policy-v1"`,
				"offset += step",
				"forecastTimes = append(forecastTimes, endTime)",
				"return Plan{}, ErrMaximumPointCountInvalid\n\t}\n\tif request.AsOfTime.IsZero()",
			},
		},
		{
			path: "internal/projectionintelligence/projectionhorizon/plan.go",
			fragments: []string{
				"type Plan struct",
				"Fingerprint string",
				"func (plan Plan) Clone() Plan",
				"func FinalizePlan(plan Plan) (Plan, error)",
				"func (plan Plan) Validate() error",
				"plan.EffectiveDuration%plan.Step != 0",
				"time.Duration(index+1) * plan.Step",
				"plan.ForecastTimes[len(plan.ForecastTimes)-1].Equal(",
				"ErrPlanFingerprintInvalid",
				"expectedFingerprint := calculatePlanFingerprint(plan)",
			},
		},
		{
			path: "internal/projectionintelligence/projectionhorizon/fingerprint.go",
			fragments: []string{
				`FingerprintVersion = "projection-horizon-plan-fingerprint-v1"`,
				`fingerprintPrefix  = "sha256:"`,
				"sha256.New()",
				"plan.Version",
				"plan.PolicyName",
				"plan.AsOfTime",
				"plan.EndTime",
				"plan.Step",
				"plan.RequestedDuration",
				"plan.EffectiveDuration",
				"plan.Truncated",
				"plan.TruncationReason",
				"len(plan.ForecastTimes)",
				"for _, forecastTime := range plan.ForecastTimes",
			},
		},
		{
			path: "internal/projectionintelligence/projectionhorizon/hardening_test.go",
			fragments: []string{
				"TestNewRejectsNonNormalizedPolicyName",
				"TestNewRejectsNonDivisibleConfiguredGrid",
				"TestNewRejectsStepAboveMinimumDuration",
				"TestNewRejectsUnsafeMaximumPointCount",
				"TestNewRejectsExtremeConfigurationBeforeScheduleAllocation",
				"TestNilPolicyReturnsLifecycleError",
				"TestBuildDefaultUsesExplicitDefaultPath",
				"TestBuildNormalizesNonUTCTime",
				"TestPlanFingerprintCoversRequestedAndTruncationEvidence",
				"TestPlanValidationRejectsManualInvariantCorruption",
				"TestPlanValidationRejectsFingerprintCorruption",
			},
		},
		{
			path: "internal/projectionintelligence/projectionhorizon/policy_test.go",
			fragments: []string{
				"TestBuildRejectsUnalignedRequestedDuration",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/baseline.go",
			fragments: []string{
				"ErrHorizonPlanInvalid",
				"if err := plan.Validate(); err != nil",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/fingerprint.go",
			fragments: []string{
				"plan.Fingerprint",
			},
			forbidden: []string{
				"plan.AsOfTime,",
				"plan.EndTime,",
				"plan.Step,",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/continuation_core.go",
			fragments: []string{
				"ErrHorizonPlanInvalid",
				"if err := plan.Validate(); err != nil",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/fingerprint.go",
			fragments: []string{
				"plan.Fingerprint",
			},
			forbidden: []string{
				"plan.AsOfTime,",
				"plan.EndTime,",
				"plan.Step,",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/composer.go",
			fragments: []string{
				"ErrHorizonPlanInvalid",
				"if err := plan.Validate(); err != nil",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/policy.go",
			fragments: []string{
				"func (policy Policy) horizonConfig() projectionhorizon.Config",
				"func (policy Policy) validateRequestedDuration(",
				"policy.horizonConfig().ResolveRequestedDuration(",
				"policy.horizonConfig().Validate()",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/service.go",
			fragments: []string{
				"request.RequestedDuration < 0",
				"service.policy.validateRequestedDuration(",
			},
			forbidden: []string{
				"request.RequestedDuration <= 0",
			},
		},
		{
			path: "internal/http/handlers/projection_intelligence.go",
			fragments: []string{
				"if normalized == \"\" {",
				"return 0, nil",
			},
		},
		{
			path: "internal/http/handlers/projection_intelligence_default_duration_test.go",
			fragments: []string{
				"TestProjectionIntelligenceHandlerAllowsDefaultDuration",
				"reader.request.RequestedDuration != 0",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/horizon_plan_validation_test.go",
			fragments: []string{
				"TestProjectRejectsInvalidHorizonPlannerResult",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/horizon_plan_validation_test.go",
			fragments: []string{
				"TestProjectRejectsInvalidHorizonPlannerResult",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/horizon_plan_validation_test.go",
			fragments: []string{
				"TestComposeRejectsInvalidHorizonPlannerResult",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/service_horizon_test.go",
			fragments: []string{
				"TestServiceAllowsDefaultProjectionDuration",
				"TestServiceRejectsInvalidHorizonDurationBeforeLoadingSnapshot",
				`name:     "negative"`,
				`name:     "below minimum"`,
				`name:     "off fixed grid"`,
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			fragments: []string{
				"PROJECTION_HORIZON_POLICY_VERSION=projection-horizon-policy-v2",
				"PROJECTION_HORIZON_FIXED_STEP_GRID=ENFORCED",
				"PROJECTION_HORIZON_PLAN_VALIDATION=ENFORCED",
				"PROJECTION_HORIZON_PLAN_FINGERPRINT=SHA256",
				"PROJECTION_HORIZON_DEFAULT_HTTP_DURATION=SUPPORTED",
				"PROJECTION_HORIZON_POINT_COUNT_BOUNDED=ENFORCED",
				"apps/api/tools/projectionhorizonreviewaudit",
			},
		},
		{
			path: "../../docs/135_PROJECTION_HORIZON_REVIEW_HARDENING.md",
			fragments: []string{
				"Status: closed",
				"AUTHORITATIVE_BASELINE_COMMIT=88532c3dbe1b943544ce15ea741315f6dad138e7",
				"ENGINEERING_IMPLEMENTATION_COMMIT=7249aa7625dd306bbd769dade6ce3262edca01ab",
				"PERMANENT_AUDIT_COMMIT=d2bc87b07ea0eb6a0b9b25f0a1e3cb2cbc52cd1b",
				"PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30402249129",
				"Backend Race Safety=SUCCESS",
				"Backend Race Safety Job=90419597222",
				"Backend Quality=SUCCESS",
				"Backend Quality Job=90419597284",
				"PostgreSQL 16 Integration=SUCCESS",
				"PostgreSQL 16 Integration Job=90419597312",
				"Backend Container=SUCCESS",
				"Backend Container Job=90419817088",
				"PROJECTION_HORIZON_POLICY_VERSION=projection-horizon-policy-v2",
				"PROJECTION_HORIZON_FIXED_STEP_GRID=ENFORCED",
				"PROJECTION_HORIZON_PLAN_VALIDATION=ENFORCED",
				"PROJECTION_HORIZON_PLAN_FINGERPRINT=SHA256",
				"PROJECTION_HORIZON_DEFAULT_HTTP_DURATION=SUPPORTED",
				"PROJECTION_HORIZON_POINT_COUNT_BOUNDED=ENFORCED",
				"PROJECTION_HORIZON_POLICY_NAME=VALIDATED",
				"PROJECTION_HORIZON_NIL_POLICY_ERROR=TYPED",
				"PROJECTION_HORIZON_CONSUMER_PLAN_VALIDATION=ENFORCED",
				"PUBLIC_PLAN_STRUCT=RETAINED_WITH_VALIDATE",
				"FAILED_CONSTRUCTOR_NIL_RESULT=IDIOMATIC_GO_RETAINED",
				"INTERNAL_ZERO_DEFAULT_SENTINEL=RETAINED_WITH_EXPLICIT_BUILD_DEFAULT",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
				"PROJECTION_HORIZON_ENGINEERING_DEBT=CLOSED",
				"PROJECTION_HORIZON_ADDITIONAL_CODE_FIXES_REQUIRED=NO",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO",
				"PROJECTION_HORIZON_REVIEW_STATUS=CLOSED",
			},
		},
		{
			path: "../../docs/DOCUMENT_INDEX.md",
			fragments: []string{
				"## Document 135 — Projection Horizon Review Hardening",
				"135_PROJECTION_HORIZON_REVIEW_HARDENING.md",
				"GitHub Actions run `30402249129`",
				"formal closure with zero open, unclassified, or deferred findings",
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				"go run ./tools/projectionhorizonreviewaudit -strict",
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
			failures = append(
				failures,
				fmt.Sprintf("read %s: %v", item.path, err),
			)
			continue
		}

		text := string(content)
		for _, fragment := range item.fragments {
			if !strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s misses %q",
						item.path,
						fragment,
					),
				)
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s retains forbidden %q",
						item.path,
						fragment,
					),
				)
			}
		}
	}

	sort.Strings(failures)
	return failures
}

func resolveAPIRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		root, err := filepath.Abs(explicit)
		if err != nil {
			return "", fmt.Errorf("resolve explicit API root: %w", err)
		}
		if isAPIRoot(root) {
			return root, nil
		}
		return "", fmt.Errorf("%s is not the apps/api module root", root)
	}

	current, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	current, err = filepath.Abs(current)
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		if isAPIRoot(current) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf("apps/api module root was not found")
}

func isAPIRoot(path string) bool {
	goModulePath := filepath.Join(path, "go.mod")
	horizonPath := filepath.Join(
		path,
		"internal",
		"projectionintelligence",
		"projectionhorizon",
	)

	goModuleInfo, goModuleErr := os.Stat(goModulePath)
	horizonInfo, horizonErr := os.Stat(horizonPath)
	return goModuleErr == nil &&
		!goModuleInfo.IsDir() &&
		horizonErr == nil &&
		horizonInfo.IsDir()
}
