package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var legacyCoverageFieldPattern = regexp.MustCompile(
	`(?m)^\s*DataCoverageRatio\s*:`,
)

type requirement struct {
	path      string
	fragments []string
	patterns  []*regexp.Regexp
	forbidden []string
}

func main() {
	strict := flag.Bool(
		"strict",
		false,
		"fail when a Historical Series hardening contract is absent",
	)
	flag.Parse()

	requirements := []requirement{
		{
			path: "internal/historicalintelligence/historicalseries/contracts.go",
			fragments: []string{
				"historical-series-builder-v2",
				"type CoverageEvidence struct",
				"type DatasetCoverage struct",
			},
			patterns: []*regexp.Regexp{
				regexp.MustCompile(
					`(?m)^\s*Coverage\s+CoverageEvidence\s*$`,
				),
			},
			forbidden: []string{
				"DataCoverageRatio",
			},
		},
		{
			path: "internal/historicalintelligence/historicalseries/coverage.go",
			fragments: []string{
				"BindDatasetCoverage",
				"DatasetReadIncomplete",
				"CoverageStateUnavailable",
				"CoverageStatePartial",
				"CoverageStateComplete",
			},
		},
		{
			path: "internal/historicalintelligence/historicalseries/builder.go",
			fragments: []string{
				"CanonicalizePlan",
				"ErrLatestSourceTimeRequired",
				"ErrGeneratedAtRequired",
				"checkedSampleCount",
				"ErrLimitationDuplicate",
				"deriveSeriesStatus",
				"temporalCoverageScore",
			},
			forbidden: []string{
				"latestSourceUpdatedAt = window.EndTime",
				"generatedAt = window.AsOfTime",
				"request.DataCoverageRatio",
			},
		},
		{
			path: "internal/historicalintelligence/historicaltraffic/builder.go",
			fragments: []string{
				"BindDatasetCoverage",
				"DatasetReadIncomplete",
			},
			forbidden: []string{
				"DataCoverageRatio:",
				"relevantCount",
				"func conservativeCoverage(",
			},
		},
		{
			path: "internal/historicalintelligence/historicalairport/builder.go",
			fragments: []string{
				"BindDatasetCoverage",
				"DatasetReadIncomplete",
			},
			forbidden: []string{
				"DataCoverageRatio:",
			},
		},
		{
			path: "internal/historicalintelligence/historicalroute/builder.go",
			fragments: []string{
				"BindDatasetCoverage",
				"DatasetReadIncomplete",
			},
			forbidden: []string{
				"DataCoverageRatio:",
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
		for _, pattern := range item.patterns {
			if !pattern.MatchString(text) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s misses pattern %q",
						item.path,
						pattern.String(),
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

	walkErr := filepath.Walk(
		".",
		func(
			path string,
			info os.FileInfo,
			walkError error,
		) error {
			if walkError != nil {
				return walkError
			}
			if info.IsDir() ||
				filepath.Ext(path) != ".go" {
				return nil
			}

			content, readError := os.ReadFile(
				filepath.Clean(path),
			)
			if readError != nil {
				return readError
			}
			if legacyCoverageFieldPattern.Match(
				content,
			) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s retains legacy DataCoverageRatio field assignment",
						path,
					),
				)
			}
			return nil
		},
	)
	if walkErr != nil {
		failures = append(
			failures,
			fmt.Sprintf(
				"scan Go sources for legacy coverage fields: %v",
				walkErr,
			),
		)
	}

	if len(failures) == 0 {
		fmt.Println(
			"Historical series review audit: PASS",
		)
		return
	}

	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Historical series review audit: %s\n",
			failure,
		)
	}
	if *strict {
		os.Exit(1)
	}
}
