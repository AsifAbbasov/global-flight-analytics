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
		"fail when a Projection Arrival review requirement is absent",
	)
	root := flag.String(
		"root",
		"",
		"optional path to the apps/api module root",
	)
	flag.Parse()

	apiRoot, err := resolveAPIRoot(*root)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Projection arrival review audit: %v\n",
			err,
		)
		if *strict {
			os.Exit(1)
		}
		return
	}

	failures := inspectRequirements(
		apiRoot,
		reviewRequirements(),
	)
	failures = append(
		failures,
		inspectExactClosureMarkers(apiRoot)...,
	)
	if len(failures) == 0 {
		fmt.Println("Projection arrival review audit: PASS")
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Projection arrival review audit: %s\n",
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
			path: "internal/projectionintelligence/projectionarrival/model.go",
			fragments: []string{
				`Version    = "estimated-arrival-boundary-v2"`,
				`FingerprintVersion            = "estimated-arrival-boundary-fingerprint-v2"`,
				`UnavailableFingerprintVersion = "estimated-arrival-unavailable-fingerprint-v2"`,
				`type positionEvidence struct`,
				`meanClosingSpeedMPS`,
				`type unavailableReason string`,
			},
			forbidden: []string{
				`estimated-arrival-boundary-v1`,
				`estimated-arrival-boundary-fingerprint-v1`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/config.go",
			fragments: []string{
				`defaultMaximumGroundSpeedMPS = 400.0`,
				`MaximumGroundSpeedMPS`,
				`config.MaximumGroundSpeedMPS <=`,
				`config.MinimumArrivalInterval >`,
				`ErrArrivalDurationPolicyInvalid`,
				`config = config.normalized()`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/speed.go",
			fragments: []string{
				`calculateClosingSpeedProfile(`,
				`closingSpeedMPS :=`,
				`(distances[index-1] - distances[index]) /`,
				`groundSpeedMPS > maximumGroundSpeedMPS`,
				`positionSamplesFingerprint(result)`,
				`currentEndpointAt(`,
				`trajectoryPointLess(`,
			},
			forbidden: []string{
				`calculateSpeedProfile(`,
				`speedMPS <`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/duration.go",
			fragments: []string{
				`durationCeilSeconds(`,
				`math.Ceil(nanoseconds)`,
				`nanoseconds >= float64(math.MaxInt64)`,
				`durationCeilFraction(`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/computation.go",
			fragments: []string{
				`radialClosingSpeedMPS :=`,
				`uncertaintyM / radialClosingSpeedMPS`,
				`conservativeClosingSpeedMPS`,
				`latestDuration >`,
				`maximumArrivalTime :=`,
				`latestTime.After(maximumArrivalTime)`,
				`durationCeilSeconds(`,
			},
			forbidden: []string{
				`time.Duration(`,
				`remainingDistanceM /\n\t\t\t\tprofile.meanMPS`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/arrival_core.go",
			fragments: []string{
				`ErrCurrentTrajectoryIDRequired`,
				`buildPositionEvidence(`,
				`unavailableSpeedOrDuration`,
				`evidence,`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/workflow.go",
			fragments: []string{
				`currentTrajectoryID == ""`,
				`currentTrajectoryID !=`,
				`"current_trajectory_arrival_endpoint"`,
				`InputClassificationObserved`,
				`arrivalFingerprint(`,
				`evidence.samples`,
			},
			forbidden: []string{
				`strings.TrimSpace(\n\t\t\trequest.CurrentTrajectory.ID,\n\t\t) != "" &&`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/fingerprint.go",
			fragments: []string{
				`positionSamplesFingerprint(samples)`,
				`"estimated-arrival-position-samples-v1"`,
				`sample.sourceID`,
				`sample.sourceName`,
				`sample.horizontalUncertaintyM`,
				`config.MaximumGroundSpeedMPS`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/arrival_confidence.go",
			fragments: []string{
				`directional_closing_speed_stability`,
				`projectionContribution :=`,
				`destinationContribution :=`,
				`speedContribution :=`,
				`projectionContribution +`,
				`extrapolation_confidence_retention`,
				`Contribution: 0`,
			},
			forbidden: []string{
				`Contribution: -estimator.config.`,
				`projected_speed_stability`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/arrival_unavailable.go",
			fragments: []string{
				`reason unavailableReason`,
				`evidence positionEvidence`,
				`evidence.fingerprint`,
				`current_trajectory_arrival_endpoint`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/review_hardening_test.go",
			fragments: []string{
				`TestEstimateWithholdsAircraftMovingAwayFromDestination`,
				`TestEstimateWithholdsPhysicallyImpossibleGroundSpeed`,
				`TestEstimateRequiresCurrentTrajectoryIdentifier`,
				`TestArrivalFingerprintBindsCurrentEndpointEvidence`,
				`TestRadiusCrossingUncertaintyUsesRadialClosingSpeed`,
				`TestExtrapolatedArrivalRejectsLatestTimeBeyondMaximumDuration`,
				`TestArrivalConfidenceReasonsReconstructFinalScore`,
				`TestNewNormalizesMaximumGroundSpeed`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/speed_test.go",
			fragments: []string{
				`TestCalculateClosingSpeedProfilePreservesSlowAndRecedingSamples`,
				`TestCalculateClosingSpeedProfileRejectsImpossibleGroundSpeed`,
				`TestDurationCeilSecondsRoundsUpAndRejectsOverflow`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/config_test.go",
			fragments: []string{
				`name: "maximum speed"`,
				`ErrMaximumGroundSpeedInvalid`,
				`name: "duration policy coherence"`,
				`ErrArrivalDurationPolicyInvalid`,
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				`- name: Run projection arrival review audit`,
				`go run ./tools/projectionarrivalreviewaudit -strict`,
			},
		},
		{
			path: "../../docs/142_PROJECTION_ARRIVAL_REVIEW_HARDENING.md",
			fragments: []string{
				`Status: closed`,
				`REVIEW_BASELINE_COMMIT=d179e6529c2ce75ac1519d49e72015622617cbc0`,
				`ENGINEERING_CLOSURE_COMMIT=65311c066aebbc278b63e2d25558f79f57584ca3`,
				`ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30614617800`,
				`ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91104833141`,
				`ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91104833127`,
				`ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91104833181`,
				`ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91105067522`,
				`DIRECTIONAL_CLOSING_SPEED=CI_CONFIRMED`,
				`POSITION_SAMPLE_FINGERPRINT_LINEAGE=CI_CONFIRMED`,
				`PERMANENT_REVIEW_AUDIT=CI_CONFIRMED`,
				`ENGINEERING_IMPLEMENTATION=COMPLETE`,
				`ENGINEERING_DEBT=CLOSED`,
				`OPEN_CONFIRMED_FINDINGS=0`,
				`UNCLASSIFIED_FINDINGS=0`,
				`DEFERRED_FINDINGS=0`,
				`ADDITIONAL_CODE_FIXES_REQUIRED=NO`,
				`FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO`,
				`PROJECTION_ARRIVAL_FORMAL_CLOSURE=COMPLETE`,
				`PROJECTION_ARRIVAL_REVIEW_STATUS=CLOSED`,
			},
			forbidden: []string{
				`Status: engineering implementation complete, exact Continuous Integration and formal closure pending`,
				`IMPLEMENTED_PENDING_EXACT_CI`,
				`ENFORCED_PENDING_EXACT_CI`,
				`COMPLETE_PENDING_EXACT_CI`,
				`0_PENDING_EXACT_CI`,
				`NO_PENDING_EXACT_CI`,
				`FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES`,
				`PROJECTION_ARRIVAL_REVIEW_STATUS=OPEN`,
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			fragments: []string{
				`## 36. Projection Arrival Engineering and Formal Review Closure`,
				`PROJECTION_ARRIVAL_VERSION=estimated-arrival-boundary-v2`,
				`PROJECTION_ARRIVAL_DIRECTIONAL_CLOSING_SPEED=ENFORCED`,
				`PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_COMMIT=65311c066aebbc278b63e2d25558f79f57584ca3`,
				`PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30614617800`,
				`PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91104833141`,
				`PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91104833127`,
				`PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91104833181`,
				`PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91105067522`,
				`PROJECTION_ARRIVAL_PERMANENT_REVIEW_AUDIT=CI_CONFIRMED`,
				`PROJECTION_ARRIVAL_ENGINEERING_IMPLEMENTATION=COMPLETE`,
				`PROJECTION_ARRIVAL_ENGINEERING_DEBT=CLOSED`,
				`PROJECTION_ARRIVAL_OPEN_CONFIRMED_FINDINGS=0`,
				`PROJECTION_ARRIVAL_UNCLASSIFIED_FINDINGS=0`,
				`PROJECTION_ARRIVAL_DEFERRED_FINDINGS=0`,
				`PROJECTION_ARRIVAL_ADDITIONAL_CODE_FIXES_REQUIRED=NO`,
				`PROJECTION_ARRIVAL_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO`,
				`PROJECTION_ARRIVAL_FORMAL_CLOSURE=COMPLETE`,
				`PROJECTION_ARRIVAL_REVIEW_STATUS=CLOSED`,
			},
			forbidden: []string{
				`PROJECTION_ARRIVAL_PERMANENT_REVIEW_AUDIT=ENFORCED_PENDING_EXACT_CI`,
				`PROJECTION_ARRIVAL_ENGINEERING_IMPLEMENTATION=COMPLETE_PENDING_EXACT_CI`,
				`PROJECTION_ARRIVAL_OPEN_CONFIRMED_FINDINGS=0_PENDING_EXACT_CI`,
				`PROJECTION_ARRIVAL_ADDITIONAL_CODE_FIXES_REQUIRED=NO_PENDING_EXACT_CI`,
				`PROJECTION_ARRIVAL_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES`,
				`PROJECTION_ARRIVAL_REVIEW_STATUS=OPEN`,
			},
		},
		{
			path: "../../docs/DOCUMENT_INDEX.md",
			fragments: []string{
				`## Document 142 — Projection Arrival Review Hardening`,
				`is the formally closed review record`,
				`exact engineering`,
				`zero open, unclassified or`,
			},
			forbidden: []string{
				`exact Continuous Integration gates still required for formal`,
			},
		},
	}
}

type exactMarkerRequirement struct {
	path    string
	markers map[string]string
}

func inspectExactClosureMarkers(
	apiRoot string,
) []string {
	requirements := []exactMarkerRequirement{
		{
			path: "../../docs/142_PROJECTION_ARRIVAL_REVIEW_HARDENING.md",
			markers: map[string]string{
				"ENGINEERING_CLOSURE_COMMIT":                        "65311c066aebbc278b63e2d25558f79f57584ca3",
				"ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN":            "30614617800",
				"ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB": "91104833141",
				"ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB":       "91104833127",
				"ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB":           "91104833181",
				"ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB":         "91105067522",
				"PERMANENT_REVIEW_AUDIT":                            "CI_CONFIRMED",
				"ENGINEERING_IMPLEMENTATION":                        "COMPLETE",
				"ENGINEERING_DEBT":                                  "CLOSED",
				"OPEN_CONFIRMED_FINDINGS":                           "0",
				"UNCLASSIFIED_FINDINGS":                             "0",
				"DEFERRED_FINDINGS":                                 "0",
				"ADDITIONAL_CODE_FIXES_REQUIRED":                    "NO",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED":             "NO",
				"PROJECTION_ARRIVAL_FORMAL_CLOSURE":                 "COMPLETE",
				"PROJECTION_ARRIVAL_REVIEW_STATUS":                  "CLOSED",
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			markers: map[string]string{
				"PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_COMMIT":                        "65311c066aebbc278b63e2d25558f79f57584ca3",
				"PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN":            "30614617800",
				"PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB": "91104833141",
				"PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB":       "91104833127",
				"PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB":           "91104833181",
				"PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB":         "91105067522",
				"PROJECTION_ARRIVAL_PERMANENT_REVIEW_AUDIT":                            "CI_CONFIRMED",
				"PROJECTION_ARRIVAL_ENGINEERING_IMPLEMENTATION":                        "COMPLETE",
				"PROJECTION_ARRIVAL_ENGINEERING_DEBT":                                  "CLOSED",
				"PROJECTION_ARRIVAL_OPEN_CONFIRMED_FINDINGS":                           "0",
				"PROJECTION_ARRIVAL_UNCLASSIFIED_FINDINGS":                             "0",
				"PROJECTION_ARRIVAL_DEFERRED_FINDINGS":                                 "0",
				"PROJECTION_ARRIVAL_ADDITIONAL_CODE_FIXES_REQUIRED":                    "NO",
				"PROJECTION_ARRIVAL_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED":             "NO",
				"PROJECTION_ARRIVAL_FORMAL_CLOSURE":                                    "COMPLETE",
				"PROJECTION_ARRIVAL_REVIEW_STATUS":                                     "CLOSED",
			},
		},
	}

	failures := make([]string, 0)
	for _, requirement := range requirements {
		path := filepath.Clean(filepath.Join(apiRoot, requirement.path))
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			failures = append(
				failures,
				fmt.Sprintf("read exact markers from %s: %v", requirement.path, err),
			)
			continue
		}

		valuesByKey := make(map[string][]string)
		for _, line := range strings.Split(string(contentBytes), "\n") {
			trimmed := strings.TrimSpace(line)
			key, value, found := strings.Cut(trimmed, "=")
			if !found || strings.TrimSpace(key) == "" {
				continue
			}
			key = strings.TrimSpace(key)
			valuesByKey[key] = append(valuesByKey[key], strings.TrimSpace(value))
		}

		for key, expected := range requirement.markers {
			values := valuesByKey[key]
			if len(values) == 0 {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s is missing exact marker %s=%s",
						requirement.path,
						key,
						expected,
					),
				)
				continue
			}
			for _, actual := range values {
				if actual == expected {
					continue
				}
				failures = append(
					failures,
					fmt.Sprintf(
						"%s marker %s has value %q, expected %q",
						requirement.path,
						key,
						actual,
						expected,
					),
				)
			}
		}
	}

	return failures
}

func inspectRequirements(
	apiRoot string,
	requirements []requirement,
) []string {
	failures := make([]string, 0)
	for _, requirement := range requirements {
		path := filepath.Clean(
			filepath.Join(apiRoot, requirement.path),
		)
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			failures = append(
				failures,
				fmt.Sprintf("read %s: %v", requirement.path, err),
			)
			continue
		}
		content := string(contentBytes)
		for _, fragment := range requirement.fragments {
			if !strings.Contains(content, fragment) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s is missing required fragment %q",
						requirement.path,
						fragment,
					),
				)
			}
		}
		for _, fragment := range requirement.forbidden {
			if strings.Contains(content, fragment) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s contains forbidden fragment %q",
						requirement.path,
						fragment,
					),
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
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for {
		if root, validationError := validateAPIRoot(current); validationError == nil {
			return root, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf(
		"apps/api root was not found from the current directory",
	)
}

func validateAPIRoot(candidate string) (string, error) {
	absolute, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve api root: %w", err)
	}
	for _, path := range []string{
		"go.mod",
		"internal/projectionintelligence/projectionarrival",
	} {
		if _, statError := os.Stat(
			filepath.Join(absolute, path),
		); statError != nil {
			return "", statError
		}
	}
	return absolute, nil
}
