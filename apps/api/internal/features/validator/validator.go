package validator

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type Validator struct {
	policy Policy
	now    func() time.Time
}

func New(config Config) (*Validator, error) {
	policy := DefaultPolicy()
	if config.Policy != nil {
		policy = *config.Policy
	}

	if !ratioInRange(policy.MinimumValidCompletenessScore) {
		return nil, ErrInvalidMinimumCompleteness
	}
	if !ratioInRange(policy.MinimumValidInputQualityScore) {
		return nil, ErrInvalidMinimumInputQuality
	}
	if math.IsNaN(policy.NumericTolerance) ||
		math.IsInf(policy.NumericTolerance, 0) ||
		policy.NumericTolerance <= 0 {
		return nil, ErrInvalidNumericTolerance
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Validator{
		policy: policy,
		now:    now,
	}, nil
}

func (validator *Validator) Validate(
	ctx context.Context,
	input flightfeatures.FlightFeatures,
) (
	flightfeatures.FlightFeatures,
	Report,
	error,
) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.FlightFeatures{}, Report{}, err
	}

	features := input.Clone()
	collector := issueCollector{
		tolerance: validator.policy.NumericTolerance,
	}
	expectedFieldCounts := schemaFieldCounts()

	validateIdentity(&collector, features)
	validateWindow(&collector, features)
	validateProvenance(&collector, features)
	validateGroupEvidence(
		&collector,
		flightfeatures.FeatureGroupTemporal,
		"temporal.evidence",
		features.Temporal.Evidence,
		expectedFieldCounts[flightfeatures.FeatureGroupTemporal],
	)
	validateGroupEvidence(
		&collector,
		flightfeatures.FeatureGroupGeographical,
		"geographical.evidence",
		features.Geographical.Evidence,
		expectedFieldCounts[flightfeatures.FeatureGroupGeographical],
	)
	validateGroupEvidence(
		&collector,
		flightfeatures.FeatureGroupOperational,
		"operational.evidence",
		features.Operational.Evidence,
		expectedFieldCounts[flightfeatures.FeatureGroupOperational],
	)
	validateGroupEvidence(
		&collector,
		flightfeatures.FeatureGroupTrajectory,
		"trajectory.evidence",
		features.Trajectory.Evidence,
		expectedFieldCounts[flightfeatures.FeatureGroupTrajectory],
	)
	validateGroupEvidence(
		&collector,
		flightfeatures.FeatureGroupAircraft,
		"aircraft.evidence",
		features.Aircraft.Evidence,
		expectedFieldCounts[flightfeatures.FeatureGroupAircraft],
	)

	validateTemporalFeatures(&collector, features)
	validateGeographicalFeatures(&collector, features)
	validateOperationalFeatures(&collector, features)
	validateTrajectoryFeatures(&collector, features)
	validateAircraftFeatures(&collector, features)
	validateQuality(
		&collector,
		features,
		validator.policy,
	)

	if err := ctx.Err(); err != nil {
		return flightfeatures.FlightFeatures{}, Report{}, err
	}

	collector.sort()
	status := collector.status()

	features.Quality.Status = status
	features.Quality.Limitations = mergeLimitations(
		features.Quality.Limitations,
		collector.issues,
	)

	report := Report{
		ValidatorVersion: Version,
		Status:           status,
		ErrorCount:       collector.errorCount(),
		WarningCount:     collector.warningCount(),
		Issues: append(
			[]Issue(nil),
			collector.issues...,
		),
		ValidatedAt: validator.now().UTC(),
	}

	return features.Clone(), report.Clone(), nil
}

type issueCollector struct {
	tolerance float64
	issues    []Issue
}

func (collector *issueCollector) add(
	severity IssueSeverity,
	group flightfeatures.FeatureGroup,
	path string,
	code string,
	message string,
) {
	collector.issues = append(
		collector.issues,
		Issue{
			Code:     strings.TrimSpace(code),
			Message:  strings.TrimSpace(message),
			Path:     strings.TrimSpace(path),
			Group:    group,
			Severity: severity,
		},
	)
}

func (collector *issueCollector) error(
	group flightfeatures.FeatureGroup,
	path string,
	code string,
	message string,
) {
	collector.add(
		IssueSeverityError,
		group,
		path,
		code,
		message,
	)
}

func (collector *issueCollector) warning(
	group flightfeatures.FeatureGroup,
	path string,
	code string,
	message string,
) {
	collector.add(
		IssueSeverityWarning,
		group,
		path,
		code,
		message,
	)
}

func (collector *issueCollector) sort() {
	sort.SliceStable(
		collector.issues,
		func(left int, right int) bool {
			leftIssue := collector.issues[left]
			rightIssue := collector.issues[right]

			if leftIssue.Path != rightIssue.Path {
				return leftIssue.Path < rightIssue.Path
			}
			if leftIssue.Code != rightIssue.Code {
				return leftIssue.Code < rightIssue.Code
			}
			if leftIssue.Severity != rightIssue.Severity {
				return leftIssue.Severity <
					rightIssue.Severity
			}

			return leftIssue.Message < rightIssue.Message
		},
	)
}

func (collector *issueCollector) errorCount() int {
	count := 0
	for _, issue := range collector.issues {
		if issue.Severity == IssueSeverityError {
			count++
		}
	}

	return count
}

func (collector *issueCollector) warningCount() int {
	count := 0
	for _, issue := range collector.issues {
		if issue.Severity == IssueSeverityWarning {
			count++
		}
	}

	return count
}

func (collector *issueCollector) status() flightfeatures.ValidationStatus {
	if collector.errorCount() > 0 {
		return flightfeatures.ValidationStatusInvalid
	}
	if collector.warningCount() > 0 {
		return flightfeatures.ValidationStatusLimited
	}

	return flightfeatures.ValidationStatusValid
}

func schemaFieldCounts() map[flightfeatures.FeatureGroup]int {
	counts := make(map[flightfeatures.FeatureGroup]int)

	for _, definition := range flightfeatures.CurrentSchema().Definitions {
		counts[definition.Group]++
	}

	return counts
}

func stripValidatorLimitations(
	items []flightfeatures.FeatureLimitation,
) []flightfeatures.FeatureLimitation {
	result := make(
		[]flightfeatures.FeatureLimitation,
		0,
		len(items),
	)

	for _, item := range items {
		if strings.HasPrefix(item.Code, issueCodePrefix) {
			continue
		}

		result = append(result, item)
	}

	return result
}

func mergeLimitations(
	existing []flightfeatures.FeatureLimitation,
	issues []Issue,
) []flightfeatures.FeatureLimitation {
	result := append(
		[]flightfeatures.FeatureLimitation(nil),
		existing...,
	)
	seen := make(map[string]struct{}, len(result)+len(issues))

	for _, limitation := range result {
		seen[limitation.Code+"\x00"+limitation.Message] =
			struct{}{}
	}

	for _, issue := range issues {
		limitation := flightfeatures.FeatureLimitation{
			Code:    issue.Code,
			Message: issue.Message,
		}
		key := limitation.Code + "\x00" + limitation.Message
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, limitation)
	}

	return result
}

func ratioInRange(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0 &&
		value <= 1
}
