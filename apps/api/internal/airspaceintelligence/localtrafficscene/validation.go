package localtrafficscene

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactionradius"
)

const timeLayout = time.RFC3339Nano

var fingerprintPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

type ValidationReport struct {
	Status ValidationStatus
	Issues []string
}

func validateRequest(request Request, policy Policy) error {
	if strings.TrimSpace(request.RegionCode) == "" ||
		request.AsOfTime.IsZero() ||
		request.GeneratedAt.IsZero() ||
		request.GeneratedAt.Before(request.AsOfTime) {
		return fmt.Errorf("%w: region and valid times are required", ErrInvalidRequest)
	}
	if err := validateBounds(request.RegionBounds); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}
	if len(request.Observations) > policy.MaximumInputObservationCount {
		return fmt.Errorf("%w: observation count exceeds policy maximum", ErrInvalidRequest)
	}
	for index, observation := range request.Observations {
		if err := validateObservation(observation); err != nil {
			return fmt.Errorf("%w: observations[%d]: %v", ErrInvalidRequest, index, err)
		}
	}
	return nil
}

func validateObservation(observation ObservationInput) error {
	if canonicalNodeID(observation) == "" ||
		strings.TrimSpace(observation.SourceName) == "" ||
		observation.ObservedAt.IsZero() {
		return fmt.Errorf("identity, source, and observed-at time are required")
	}
	if !validLatitude(observation.Latitude) ||
		!validLongitude(observation.Longitude) ||
		!nonNegativeFinite(observation.VelocityMetersPerSecond) ||
		!headingDegrees(observation.HeadingDegrees) ||
		!finite(observation.VerticalRateMetersPerSecond) ||
		!unitInterval(observation.QualityScore) ||
		!observation.AltitudeReference.IsKnown() {
		return fmt.Errorf("coordinates, motion, quality, or altitude reference")
	}
	if observation.AltitudeMeters != nil && !finite(*observation.AltitudeMeters) {
		return fmt.Errorf("altitude")
	}
	return nil
}

func Validate(result Result, policy Policy) ValidationReport {
	issues := make([]string, 0)
	if err := policy.Validate(); err != nil {
		issues = append(issues, err.Error())
	}
	if result.SchemaVersion != SchemaVersionV1 {
		issues = append(issues, "schema_version")
	}
	if !result.Status.IsKnown() {
		issues = append(issues, "status")
	}
	if strings.TrimSpace(result.RegionCode) == "" {
		issues = append(issues, "region_code")
	}
	if err := validateBounds(result.RegionBounds); err != nil {
		issues = append(issues, "region_bounds")
	}
	if result.AsOfTime.IsZero() ||
		result.GeneratedAt.IsZero() ||
		result.GeneratedAt.Before(result.AsOfTime) {
		issues = append(issues, "times")
	}
	if result.ScopeGuard != ScopeGuardResearchOnly {
		issues = append(issues, "scope_guard")
	}
	if !unitInterval(result.Confidence.Score) ||
		!result.Confidence.Level.IsKnown() ||
		len(result.Confidence.Reasons) == 0 {
		issues = append(issues, "confidence")
	}
	if len(result.Limitations) == 0 || len(result.Explanations) == 0 {
		issues = append(issues, "explainability")
	}
	if !fingerprintPattern.MatchString(result.Provenance.InputFingerprint) {
		issues = append(issues, "input_fingerprint")
	}

	nodeIDs := make(map[string]struct{}, len(result.Aircraft))
	allowedCount := 0
	limitedCount := 0
	for index, aircraft := range result.Aircraft {
		path := fmt.Sprintf("aircraft[%d]", index)
		if strings.TrimSpace(aircraft.NodeID) == "" ||
			strings.TrimSpace(aircraft.SourceName) == "" ||
			aircraft.ObservedAt.IsZero() ||
			aircraft.ObservedAt.After(result.AsOfTime) ||
			aircraft.ObservationAge < 0 ||
			!contains(result.RegionBounds, aircraft.Latitude, aircraft.Longitude) ||
			!unitInterval(aircraft.QualityScore) {
			issues = append(issues, path)
		}
		if _, exists := nodeIDs[aircraft.NodeID]; exists {
			issues = append(issues, path+".duplicate_node_id")
		}
		nodeIDs[aircraft.NodeID] = struct{}{}
		report := interactionradius.Validate(aircraft.RadiusDecision)
		if report.Status != interactionradius.ValidationStatusValid {
			issues = append(issues, path+".radius_decision")
		}
		switch aircraft.RadiusDecision.Status {
		case interactionradius.DecisionStatusAllowed:
			allowedCount++
		case interactionradius.DecisionStatusLimited:
			limitedCount++
		default:
			issues = append(issues, path+".blocked_radius_decision")
		}
	}
	for index, item := range result.ExcludedObservations {
		if !item.Reason.IsKnown() || strings.TrimSpace(item.Message) == "" {
			issues = append(issues, fmt.Sprintf("excluded_observations[%d]", index))
		}
	}

	metrics := result.Metrics
	if metrics.InputObservationCount < 0 ||
		metrics.CandidateObservationCount < 0 ||
		metrics.IncludedAircraftCount != len(result.Aircraft) ||
		metrics.AllowedAircraftCount != allowedCount ||
		metrics.LimitedAircraftCount != limitedCount ||
		metrics.ExcludedObservationCount != len(result.ExcludedObservations) ||
		!unitInterval(metrics.SceneCoverage) {
		issues = append(issues, "metrics")
	}
	expectedStatus := statusFor(metrics, policy)
	if result.Status != expectedStatus {
		issues = append(issues, "status_consistency")
	}
	if len(issues) > 0 {
		return ValidationReport{Status: ValidationStatusInvalid, Issues: issues}
	}
	return ValidationReport{Status: ValidationStatusValid}
}

func validateBounds(bounds Bounds) error {
	if !validLatitude(bounds.MinimumLatitude) ||
		!validLatitude(bounds.MaximumLatitude) ||
		!validLongitude(bounds.MinimumLongitude) ||
		!validLongitude(bounds.MaximumLongitude) ||
		bounds.MinimumLatitude >= bounds.MaximumLatitude ||
		bounds.MinimumLongitude >= bounds.MaximumLongitude {
		return fmt.Errorf("invalid region bounds")
	}
	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func nonNegativeFinite(value float64) bool {
	return finite(value) && value >= 0
}

func unitInterval(value float64) bool {
	return finite(value) && value >= 0 && value <= 1
}

func validLatitude(value float64) bool {
	return finite(value) && value >= -90 && value <= 90
}

func validLongitude(value float64) bool {
	return finite(value) && value >= -180 && value <= 180
}

func headingDegrees(value float64) bool {
	return finite(value) && value >= 0 && value <= 360
}
