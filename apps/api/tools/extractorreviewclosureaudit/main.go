package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type rule struct {
	path      string
	required  []string
	forbidden []string
}

func main() {
	root := flag.String("root", "", "repository root")
	strict := flag.Bool("strict", true, "fail on contract violation")
	flag.Parse()

	resolved, err := resolveRoot(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rules := []rule{
		{
			path: "apps/api/internal/domain/aircraft/identity.go",
			required: []string{
				"func CanonicalICAO24(",
				"func NormalizeICAO24(",
				"func IsValidICAO24(",
				"func IsCanonicalICAO24(",
			},
			forbidden: []string{"regexp.MustCompile"},
		},
		{
			path: "apps/api/internal/domain/trajectory/clone.go",
			required: []string{
				"func (item FlightTrajectory) Clone() FlightTrajectory",
				"cloned.Points = append",
				"cloned.Segments = append",
				"cloned.CoverageGaps = append",
			},
		},
		{
			path:     "apps/api/internal/features/flightfeatures/requirements.go",
			required: []string{"func CurrentGroupFieldCount("},
		},
		{
			path: "apps/api/internal/features/extractor/extractor.go",
			required: []string{
				"request.Trajectory.Clone()",
				"aircraft.CanonicalICAO24(",
				"aircraft.IsValidICAO24(",
				"flightfeatures.CurrentGroupFieldCount(flightfeatures.FeatureGroupAircraft)",
			},
			forbidden: []string{
				"func cloneTrajectory(",
				"aircraftFeatureFieldCount",
				"icao24Pattern",
			},
		},
		{
			path: "apps/api/internal/features/aircraftprovider/provider.go",
			required: []string{
				"aircraft.NormalizeICAO24(",
				"returnedICAO24, valid := aircraft.NormalizeICAO24(",
				"flightfeatures.CurrentGroupFieldCount(flightfeatures.FeatureGroupAircraft)",
			},
			forbidden: []string{
				"func normalizeICAO24(",
				"icao24Pattern",
				"AircraftFeatureFieldCount",
			},
		},
		{
			path:      "apps/api/internal/features/aircraftprovider/contracts.go",
			forbidden: []string{"AircraftFeatureFieldCount"},
		},
		{
			path:      "apps/api/internal/features/validator/rules.go",
			required:  []string{"aircraft.IsCanonicalICAO24("},
			forbidden: []string{"icao24Pattern"},
		},
		{
			path:      "apps/api/internal/features/extractor/fingerprint.go",
			required:  []string{"aircraft.CanonicalICAO24("},
			forbidden: []string{"func normalizeFingerprintICAO24("},
		},
		{
			path: "apps/api/internal/features/extractor/fingerprint_contract_test.go",
			required: []string{
				"TestCanonicalFingerprintInputFieldOrderRemainsExplicit",
				"TestCanonicalFingerprintEvidenceMirrorsDomainContracts",
				"TestTrajectoryClonePreservesFingerprintInput",
				"reflect.TypeOf(trajectory.FlightTrajectory{})",
				`"DistanceKm": "DistanceKM"`,
			},
		},
		{
			path:     "apps/api/internal/features/extractor/contracts.go",
			required: []string{`const Version = "flight-feature-extractor-v6"`},
		},
		{
			path:     "apps/api/internal/features/extractorcomposition/contracts.go",
			required: []string{`const Version = "flight-feature-extractor-composition-v6"`},
		},
		{
			path:     "apps/api/internal/features/featurepipeline/contracts.go",
			required: []string{`const Version = "flight-feature-processing-pipeline-v10"`},
		},
		{
			path:     "apps/api/internal/features/validator/contracts.go",
			required: []string{`const Version = "flight-feature-validator-v4"`},
		},
		{
			path: ".github/workflows/backend-ci.yml",
			required: []string{
				"Run extractor review closure audit",
				"go run ./tools/extractorreviewclosureaudit -strict",
			},
		},
		{
			path: "docs/114_EXTRACTOR_REVIEW_FINAL_CLOSURE.md",
			required: []string{
				"EXTRACTOR_ICAO24_CONTRACT=CENTRALIZED",
				"EXTRACTOR_TRAJECTORY_CLONE_CONTRACT=CENTRALIZED",
				"EXTRACTOR_AIRCRAFT_FIELD_COUNT=SCHEMA_DERIVED",
				"EXTRACTOR_FINGERPRINT_MIRROR_GUARD=ENFORCED",
				"EXTRACTOR_PROCESSING_GENERATION=v5",
				"EXTRACTOR_REVIEW_STATUS=CLOSED",
				"OPEN_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
			},
		},
	}

	failures := make([]string, 0)
	for _, item := range rules {
		content, readErr := os.ReadFile(filepath.Join(resolved, item.path))
		if readErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.path, readErr))
			continue
		}
		text := string(content)
		for _, fragment := range item.required {
			if !strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s missing %q", item.path, fragment))
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s contains forbidden %q", item.path, fragment))
			}
		}
	}

	failures = append(failures, rejectDuplicateContracts(resolved)...)
	if len(failures) == 0 {
		fmt.Println("Extractor review closure audit: PASS")
		return
	}

	fmt.Fprintln(os.Stderr, "Extractor review closure audit: FAIL")
	for _, failure := range failures {
		fmt.Fprintln(os.Stderr, "-", failure)
	}
	if *strict {
		os.Exit(1)
	}
}

func rejectDuplicateContracts(root string) []string {
	patterns := []struct {
		roots    []string
		fragment string
		allowed  string
	}{
		{
			roots: []string{
				"apps/api/internal/domain/aircraft",
				"apps/api/internal/features/extractor",
				"apps/api/internal/features/aircraftprovider",
				"apps/api/internal/features/validator",
			},
			fragment: "func CanonicalICAO24(",
			allowed:  "apps/api/internal/domain/aircraft/identity.go",
		},
		{
			roots: []string{
				"apps/api/internal/domain/aircraft",
				"apps/api/internal/features/extractor",
				"apps/api/internal/features/aircraftprovider",
				"apps/api/internal/features/validator",
			},
			fragment: "func NormalizeICAO24(",
			allowed:  "apps/api/internal/domain/aircraft/identity.go",
		},
		{
			roots: []string{
				"apps/api/internal/domain/trajectory",
				"apps/api/internal/features/extractor",
			},
			fragment: "func (item FlightTrajectory) Clone() FlightTrajectory",
			allowed:  "apps/api/internal/domain/trajectory/clone.go",
		},
	}

	failures := make([]string, 0)
	for _, pattern := range patterns {
		for _, relativeRoot := range pattern.roots {
			searchRoot := filepath.Join(root, relativeRoot)
			walkErr := filepath.Walk(searchRoot, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() || filepath.Ext(path) != ".go" {
					return nil
				}
				content, readErr := os.ReadFile(path)
				if readErr != nil {
					return readErr
				}
				if !strings.Contains(string(content), pattern.fragment) {
					return nil
				}
				relative, relErr := filepath.Rel(root, path)
				if relErr != nil {
					return relErr
				}
				relative = filepath.ToSlash(relative)
				if relative != pattern.allowed {
					failures = append(failures, fmt.Sprintf("duplicate %q in %s", pattern.fragment, relative))
				}
				return nil
			})
			if walkErr != nil {
				failures = append(failures, walkErr.Error())
			}
		}
	}
	return failures
}

func resolveRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return filepath.Abs(strings.TrimSpace(explicit))
	}
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, statErr := os.Stat(filepath.Join(current, "apps/api/go.mod")); statErr == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root was not found")
		}
		current = parent
	}
}
