package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type sourceCheck struct {
	path        string
	mustContain []string
	mustExclude []string
}

func main() {
	strict := flag.Bool(
		"strict",
		false,
		"fail on every Historical Window review invariant",
	)
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "unexpected positional arguments")
		os.Exit(1)
	}
	_ = strict

	checks := []sourceCheck{
		{
			path: "internal/historicalintelligence/historicalwindow/contracts.go",
			mustContain: []string{
				`FingerprintVersion = "historical-time-window-fingerprint-v2"`,
				`ValidationVersion  = "historical-time-window-validation-v1"`,
				`!bucket.StartTime.Before(bucket.EndTime)`,
			},
		},
		{
			path: "internal/historicalintelligence/historicalwindow/planner.go",
			mustContain: []string{
				`return Plan{}, ErrContextRequired`,
				`if len(buckets) >= maximum`,
				`func previousStartForSpan(`,
				`return ctx.Err()`,
			},
			mustExclude: []string{
				`effectiveEnd.Sub(effectiveStart)`,
				`sequence%1_024`,
				`setEffectiveAndPreviousWindows`,
			},
		},
		{
			path: "internal/historicalintelligence/historicalwindow/boundaries.go",
			mustContain: []string{
				`func granularityPolicyFor(`,
				`if !floor.Equal(normalized)`,
				`return time.Time{}, ErrBoundarySequenceInvalid`,
			},
		},
		{
			path: "internal/historicalintelligence/historicalwindow/fingerprint.go",
			mustContain: []string{
				`plan.PreviousWindow`,
				`for _, bucket := range plan.Buckets`,
				`bucket.Key`,
			},
			mustExclude: []string{
				`plan.MaximumBucketCount`,
			},
		},
		{
			path: "internal/historicalintelligence/historicalwindow/validation.go",
			mustContain: []string{
				`func CanonicalizePlan(`,
				`func ValidatePlan(`,
				`ErrPlanIntegrityInvalid`,
			},
		},
		{
			path: "internal/historicalintelligence/historicalseries/builder.go",
			mustContain: []string{
				`historicalwindow.CanonicalizePlan(`,
				`request.Values[index].Bucket = planned`,
			},
		},
		{
			path: "internal/historicalintelligence/historicaltraffic/builder.go",
			mustContain: []string{
				`historicalwindow.CanonicalizePlan(`,
				`request.Plan = canonicalPlan`,
			},
		},
		{
			path: "internal/historicalintelligence/historicalairport/builder.go",
			mustContain: []string{
				`historicalwindow.CanonicalizePlan(`,
				`request.Plan = canonicalPlan`,
			},
		},
		{
			path: "internal/historicalintelligence/historicalroute/builder.go",
			mustContain: []string{
				`historicalwindow.CanonicalizePlan(`,
				`request.Plan = canonicalPlan`,
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			mustContain: []string{
				`Run historical window review audit`,
				`go run ./tools/historicalwindowreviewaudit -strict`,
			},
		},
	}

	for _, check := range checks {
		data, err := os.ReadFile(check.path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", check.path, err)
			os.Exit(1)
		}
		text := string(data)
		for _, required := range check.mustContain {
			if !strings.Contains(text, required) {
				fmt.Fprintf(
					os.Stderr,
					"%s is missing required marker %q\n",
					check.path,
					required,
				)
				os.Exit(1)
			}
		}
		for _, prohibited := range check.mustExclude {
			if strings.Contains(text, prohibited) {
				fmt.Fprintf(
					os.Stderr,
					"%s contains prohibited marker %q\n",
					check.path,
					prohibited,
				)
				os.Exit(1)
			}
		}
	}

	fmt.Println("Historical window review audit: PASS")
}
