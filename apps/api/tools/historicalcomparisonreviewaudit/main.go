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
		"fail when a Historical Comparison review contract is absent",
	)
	flag.Parse()

	requirements := []requirement{
		{
			path: "internal/historicalintelligence/historicalcomparison/contracts.go",
			fragments: []string{
				"historical-period-comparison-v2",
				"type periodValues struct",
			},
			forbidden: []string{
				"type Values struct",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcomparison/engine.go",
			fragments: []string{
				"validateSourceResults",
				"validateCompatibility",
				"selectPeriodValues",
				"buildPeriodComparison",
				"assembleComparedResult",
				"validateComparedResult",
			},
			forbidden: []string{
				"reflect",
				"DeepEqual",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcomparison/compatibility.go",
			fragments: []string{
				"current.Scope.Equal(previous.Scope)",
				"validateCoverageCompatibility",
				"coverageEqualityTolerance",
				"ErrCoverageMismatch",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcomparison/arithmetic.go",
			fragments: []string{
				"calculatePercentageChange",
				"math.MaxFloat64/100",
				"ArithmeticError",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcomparison/errors.go",
			fragments: []string{
				"ErrComparisonArithmeticInvalid",
				"ErrComparisonResultInvalid",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcomparison/quality.go",
			fragments: []string{
				"comparisonLimitations",
				"historical_comparison_period_quality",
				"historical_comparison_previous_period_limitations",
				"historical_comparison_matched_partial_coverage",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcomparison/provenance.go",
			fragments: []string{
				"comparisonFingerprint",
				"appendResultIdentity",
				"mergeSourceNames",
				"\"current\"",
				"\"previous\"",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcomparison/values.go",
			fragments: []string{
				"comparisonValueSelectors",
				"AggregationCount",
				"AggregationSum",
				"AggregationMinimum",
				"AggregationMaximum",
				"AggregationAverage",
				"AggregationMedian",
				"AggregationRatio",
			},
			forbidden: []string{
				"switch current.Metric.Aggregation",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcomparison/validation.go",
			fragments: []string{
				"ErrNestedComparisonUnsupported",
				"ErrComparisonResultInvalid",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcontract/scope.go",
			fragments: []string{
				"func (scope Scope) Equal",
				"scope.DestinationICAOCode",
			},
			forbidden: []string{
				"reflect.DeepEqual",
				"reflect.",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcomparison/engine_test.go",
			fragments: []string{
				"TestAttachRejectsUnequalCoverageProfiles",
				"TestAttachCarriesMatchedPartialCoverageQuality",
				"TestValidateCompatibilityClassifiesEveryMismatch",
				"TestSelectPeriodValuesSupportsEveryAggregation",
				"TestAttachRejectsPercentageOverflowWithArithmeticError",
				"TestComparisonFingerprintChangesWithPreviousEvidence",
				"TestAttachRejectsNestedComparison",
			},
		},
		{
			path: "../../docs/129_HISTORICAL_COMPARISON_REVIEW_HARDENING.md",
			fragments: []string{
				"Status: closed",
				"COMPARISON_COVERAGE_PROFILE_MATCH=ENFORCED",
				"PREVIOUS_PERIOD_QUALITY=BOUND",
				"COMPARISON_PROVENANCE=ATOMIC",
				"COMPARISON_FINGERPRINT_BOTH_PERIODS=BOUND",
				"COMPARISON_ARITHMETIC_FINITE=ENFORCED",
				"SCOPE_EQUALITY=EXPLICIT",
				"PERCENTAGE_ZERO_BASE=UNDEFINED_OPTIONAL",
				"TEMPORAL_SUMMARY_SEMANTICS=DOCUMENTED",
				"HISTORICAL_COMPARISON_ENGINEERING_REMEDIATION=IMPLEMENTED",
				"d60af19d87fbbb234bab72fb4389a8d503ae06b9",
				"21734b85b9f50ae717dca031c798866161895989",
				"30341011740",
				"Backend Quality=SUCCESS",
				"Backend Quality Job=90216363225",
				"PostgreSQL 16 Integration=SUCCESS",
				"PostgreSQL 16 Integration Job=90216363216",
				"Backend Race Safety=SUCCESS",
				"Backend Race Safety Job=90216363189",
				"Backend Container=SUCCESS",
				"Backend Container Job=90216611574",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
				"HISTORICAL_COMPARISON_ENGINEERING_DEBT=CLOSED",
				"HISTORICAL_COMPARISON_ADDITIONAL_CODE_FIXES_REQUIRED=NO",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO",
				"HISTORICAL_COMPARISON_REVIEW_STATUS=CLOSED",
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
			"Historical comparison review audit: PASS",
		)
		return
	}

	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Historical comparison review audit: %s\n",
			failure,
		)
	}
	if *strict {
		os.Exit(1)
	}
}
