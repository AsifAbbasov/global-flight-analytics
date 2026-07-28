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
		"fail when a Historical Replay review contract is absent",
	)
	flag.Parse()

	requirements := []requirement{
		{
			path: "internal/historicalintelligence/historicalreplay/contracts.go",
			fragments: []string{
				`const Version = "historical-replay-v2"`,
				`const FingerprintVersion = "historical-replay-input-fingerprint-v1"`,
				"StatusComplete Status = \"complete\"",
				"StatusPartial  Status = \"partial\"",
				"StatusFailed   Status = \"failed\"",
				"PlannedWindowCount   int",
				"CompletedWindowCount int",
				"HasFailure bool",
				"InputFingerprint string",
				"func (result Result) Validate() error",
			},
		},
		{
			path: "internal/historicalintelligence/historicalreplay/request.go",
			fragments: []string{
				"historicalcontract.MetricSpecFor(",
				"specification.AllowsScope(scope.Type)",
				"ErrDatasetLimitInvalid",
				"ErrGeneratedAtBeforeAsOfTime",
				"ErrGeneratedAtAfterStartTime",
				"PlanningMaximumBucketCount",
				"MaximumBucketCount: 1",
			},
		},
		{
			path: "internal/historicalintelligence/historicalreplay/runner.go",
			fragments: []string{
				"return Result{}, ErrContextRequired",
				"validateOutcome(",
				"ErrReplayContinuityMismatch",
				"FailureCodeContinuityMismatch",
				"StatusPartial",
				"result.Validate()",
			},
			forbidden: []string{
				"ctx = context.Background()",
			},
		},
		{
			path: "internal/historicalintelligence/historicalreplay/outcome_validation.go",
			fragments: []string{
				"historicalmaterialization.Version",
				"historicalwindow.ValidatePlan(",
				"validateReadSummary(",
				"historicalcontract.Validate(result)",
				"strings.TrimSpace(record.ID)",
				"record.InputFingerprint",
				"current_result.comparison.previous_value",
				"reflect.DeepEqual(",
			},
		},
		{
			path: "internal/historicalintelligence/historicalreplay/validation.go",
			fragments: []string{
				"historicalwindow.ValidatePlan(",
				"PlannedWindowCount",
				"CompletedWindowCount",
				"validateResultStatus(",
				"validateResultFailure(",
				"replayInputFingerprint(",
				"ErrReplayContinuityMismatch",
			},
		},
		{
			path: "internal/historicalintelligence/historicalreplay/runner_test.go",
			fragments: []string{
				"TestRunReturnsValidatedCompleteReplay",
				"TestRunReturnsSelfContainedPartialResult",
				"TestRunRejectsInvalidMaterializationOutcome",
				"TestRunValidatesGlobalRequestBeforeReplay",
				"TestRunUsesBoundedPlanningLimits",
				"TestRunRejectsNilAndCanceledContext",
				"TestRunReturnsPartialContextCancellation",
				"TestRunRejectsOverlappingReadContinuityMismatch",
				"TestRunReturnsFailedResultWhenNoWindowExists",
				"TestRunIsDeterministicAndCloneIsolated",
				"TestResultValidateRejectsTampering",
			},
		},
		{
			path: "cmd/materialize-historical-intelligence/operation.go",
			fragments: []string{
				"errCommandContextRequired",
				"if err != nil && replayed.Version == \"\"",
				"report := reportFromReplay(",
				"return report, err",
			},
		},
		{
			path: "cmd/materialize-historical-intelligence/main.go",
			fragments: []string{
				"writeCommandOutcome(",
				"if report.Version != \"\"",
				"encoder.Encode(report)",
				"if executeErr != nil",
			},
		},
		{
			path: "cmd/materialize-historical-intelligence/contracts.go",
			fragments: []string{
				`const commandVersion = "historical-intelligence-production-runner-v2"`,
				"Status  string `json:\"status\"`",
				"PlannedReplayWindowCount",
				"CompletedReplayWindowCount",
				"ReplayInputFingerprint",
				"ReplayFailure",
			},
		},
		{
			path: "cmd/materialize-historical-intelligence/replay_partial_test.go",
			fragments: []string{
				"TestCommandOperationPreservesPartialReplayReport",
				"TestWriteCommandOutcomeEmitsPrefixBeforeFailure",
				"TestCommandOperationRejectsNilContext",
			},
		},
		{
			path: "../../docs/133_HISTORICAL_REPLAY_REVIEW_HARDENING.md",
			fragments: []string{
				"Status: implemented review remediation",
				"MATERIALIZATION_OUTCOME_VALIDATION=ENFORCED",
				"REPLAY_RESULT_STATUS=EXPLICIT",
				"PARTIAL_PROGRESS_REPORTING=PRESERVED",
				"OVERLAPPING_PERIOD_CONTINUITY=VERIFIED",
				"REPLAY_REQUEST_VALIDATION=EARLY",
				"REPLAY_PLANNING_LIMITS=BOUNDED",
				"NIL_CONTEXT_REJECTED=YES",
				"REPLAY_INPUT_FINGERPRINT=BOUND",
				"HISTORICAL_REPLAY_ENGINEERING_REMEDIATION=IMPLEMENTED",
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				"go run ./tools/historicalreplayreviewaudit -strict",
			},
		},
	}

	failures := make([]string, 0)
	for _, item := range requirements {
		content, err := os.ReadFile(
			filepath.Clean(item.path),
		)
		if err != nil {
			failures = append(
				failures,
				fmt.Sprintf(
					"read %s: %v",
					item.path,
					err,
				),
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

	if len(failures) == 0 {
		fmt.Println(
			"Historical replay review audit: PASS",
		)
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Historical replay review audit: %s\n",
			failure,
		)
	}
	if *strict {
		os.Exit(1)
	}
}
