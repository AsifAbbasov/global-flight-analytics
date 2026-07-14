package analyticalresult

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
)

var (
	ErrStatusInvalid = errors.New(
		"analytical result status is invalid",
	)
	ErrCalculatedAtMissing = errors.New(
		"analytical result calculation time is required",
	)
	ErrDataQualityEvaluationAfterCalculation = errors.New(
		"data quality evaluation time must not be after analytical result calculation time",
	)
	ErrValueRequired = errors.New(
		"analytical result value is required",
	)
	ErrValueForbidden = errors.New(
		"analytical result value is forbidden",
	)
	ErrConfidenceLevelInvalid = errors.New(
		"analytical result confidence level is invalid",
	)
	ErrConfidenceScoreInvalid = errors.New(
		"analytical result confidence score must be finite and between zero and one",
	)
	ErrConfidenceNoneScoreInvalid = errors.New(
		"none confidence must have a zero score",
	)
	ErrConfidenceRequired = errors.New(
		"complete or limited analytical result requires confidence",
	)
	ErrEligibilityCapabilityInvalid = errors.New(
		"analytical result eligibility capability is invalid",
	)
	ErrEligibilityTimeMissing = errors.New(
		"analytical result eligibility evaluation time is required",
	)
	ErrAllowedEligibilityReasonsPresent = errors.New(
		"allowed eligibility cannot contain denial reasons",
	)
	ErrDeniedEligibilityReasonsMissing = errors.New(
		"denied eligibility requires at least one reason",
	)
	ErrDeniedEligibilityRequired = errors.New(
		"denied analytical result requires denied eligibility",
	)
	ErrDeniedEligibilityForNonDeniedStatus = errors.New(
		"non-denied analytical result cannot contain denied eligibility",
	)
	ErrFailureRequired = errors.New(
		"failed analytical result requires failure metadata",
	)
	ErrFailureForbidden = errors.New(
		"non-failed analytical result cannot contain failure metadata",
	)
	ErrLimitedExplanationRequired = errors.New(
		"limited analytical result requires a warning or limitation",
	)
	ErrMachineCodeInvalid = errors.New(
		"machine-readable code is invalid",
	)
	ErrNoticeMessageMissing = errors.New(
		"notice message is required",
	)
	ErrDuplicateNoticeCode = errors.New(
		"notice codes must be unique within a collection",
	)
	ErrFailureMessageMissing = errors.New(
		"failure message is required",
	)
	ErrSourceNameInvalid = errors.New(
		"analytical source name is invalid",
	)
	ErrSourceRoleInvalid = errors.New(
		"analytical source role is invalid",
	)
	ErrSourceObservationRangeIncomplete = errors.New(
		"analytical source observation range must be entirely empty or entirely populated",
	)
	ErrSourceObservationRangeInvalid = errors.New(
		"analytical source observation end must not precede its start",
	)
)

func (result Result[T]) Validate() error {
	if !result.Status.IsKnown() {
		return fmt.Errorf(
			"%w: %q",
			ErrStatusInvalid,
			result.Status,
		)
	}

	if result.CalculatedAt.IsZero() {
		return ErrCalculatedAtMissing
	}

	if err := result.Confidence.Validate(); err != nil {
		return fmt.Errorf(
			"validate analytical confidence: %w",
			err,
		)
	}

	if result.DataQuality != nil {
		if err := result.DataQuality.Validate(); err != nil {
			return fmt.Errorf(
				"validate analytical data quality: %w",
				err,
			)
		}
		if result.DataQuality.EvaluatedAt.After(
			result.CalculatedAt,
		) {
			return fmt.Errorf(
				"%w: data_quality=%s calculated_at=%s",
				ErrDataQualityEvaluationAfterCalculation,
				result.DataQuality.EvaluatedAt.Format(
					time.RFC3339Nano,
				),
				result.CalculatedAt.Format(
					time.RFC3339Nano,
				),
			)
		}
	}

	if result.Eligibility != nil {
		if err := result.Eligibility.Validate(); err != nil {
			return fmt.Errorf(
				"validate analytical eligibility: %w",
				err,
			)
		}
	}

	if err := validateSources(result.Sources); err != nil {
		return err
	}

	if err := validateNotices("warnings", result.Warnings); err != nil {
		return err
	}

	if err := validateNotices("limitations", result.Limitations); err != nil {
		return err
	}

	switch result.Status {
	case StatusComplete:
		if !result.HasValue {
			return ErrValueRequired
		}
		if result.Confidence.Level == ConfidenceLevelNone {
			return ErrConfidenceRequired
		}
		if result.Eligibility != nil && !result.Eligibility.Allowed {
			return ErrDeniedEligibilityForNonDeniedStatus
		}
		if len(result.Warnings) > 0 || len(result.Limitations) > 0 {
			return ErrLimitedExplanationRequired
		}
		if result.Failure != nil {
			return ErrFailureForbidden
		}

	case StatusLimited:
		if !result.HasValue {
			return ErrValueRequired
		}
		if result.Confidence.Level == ConfidenceLevelNone {
			return ErrConfidenceRequired
		}
		if result.Eligibility != nil && !result.Eligibility.Allowed {
			return ErrDeniedEligibilityForNonDeniedStatus
		}
		if len(result.Warnings) == 0 && len(result.Limitations) == 0 {
			return ErrLimitedExplanationRequired
		}
		if result.Failure != nil {
			return ErrFailureForbidden
		}

	case StatusDenied:
		if result.HasValue {
			return ErrValueForbidden
		}
		if result.Eligibility == nil || result.Eligibility.Allowed {
			return ErrDeniedEligibilityRequired
		}
		if result.Confidence.Level != ConfidenceLevelNone {
			return ErrConfidenceNoneScoreInvalid
		}
		if result.Failure != nil {
			return ErrFailureForbidden
		}

	case StatusFailed:
		if result.HasValue {
			return ErrValueForbidden
		}
		if result.Eligibility != nil && !result.Eligibility.Allowed {
			return ErrDeniedEligibilityForNonDeniedStatus
		}
		if result.Confidence.Level != ConfidenceLevelNone {
			return ErrConfidenceNoneScoreInvalid
		}
		if result.Failure == nil {
			return ErrFailureRequired
		}
		if err := result.Failure.Validate(); err != nil {
			return fmt.Errorf(
				"validate analytical failure: %w",
				err,
			)
		}
	}

	return nil
}

func (confidence Confidence) Validate() error {
	if !confidence.Level.IsKnown() {
		return fmt.Errorf(
			"%w: %q",
			ErrConfidenceLevelInvalid,
			confidence.Level,
		)
	}

	if math.IsNaN(confidence.Score) ||
		math.IsInf(confidence.Score, 0) ||
		confidence.Score < 0 ||
		confidence.Score > 1 {
		return fmt.Errorf(
			"%w: %f",
			ErrConfidenceScoreInvalid,
			confidence.Score,
		)
	}

	if confidence.Level == ConfidenceLevelNone && confidence.Score != 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrConfidenceNoneScoreInvalid,
			confidence.Score,
		)
	}

	if confidence.Level != ConfidenceLevelNone && confidence.Score == 0 {
		return fmt.Errorf(
			"%w: non-none confidence cannot have zero score",
			ErrConfidenceScoreInvalid,
		)
	}

	return validateNotices(
		"confidence reasons",
		confidence.Reasons,
	)
}

func (eligibility Eligibility) Validate() error {
	if !isKnownCapability(eligibility.Capability) {
		return fmt.Errorf(
			"%w: %q",
			ErrEligibilityCapabilityInvalid,
			eligibility.Capability,
		)
	}

	if eligibility.EvaluatedAt.IsZero() {
		return ErrEligibilityTimeMissing
	}

	if eligibility.Allowed && len(eligibility.Reasons) > 0 {
		return ErrAllowedEligibilityReasonsPresent
	}

	if !eligibility.Allowed && len(eligibility.Reasons) == 0 {
		return ErrDeniedEligibilityReasonsMissing
	}

	return nil
}

func (failure Failure) Validate() error {
	if err := validateMachineCode(failure.Code); err != nil {
		return fmt.Errorf(
			"validate failure code: %w",
			err,
		)
	}

	if strings.TrimSpace(failure.Message) == "" ||
		failure.Message != strings.TrimSpace(failure.Message) {
		return ErrFailureMessageMissing
	}

	return nil
}

func validateSources(sources []Source) error {
	for index, source := range sources {
		if err := source.Validate(); err != nil {
			return fmt.Errorf(
				"validate analytical source at index %d: %w",
				index,
				err,
			)
		}
	}

	return nil
}

func (source Source) Validate() error {
	if strings.TrimSpace(source.Name) == "" ||
		source.Name != strings.TrimSpace(source.Name) {
		return ErrSourceNameInvalid
	}

	if !source.Role.IsKnown() {
		return fmt.Errorf(
			"%w: %q",
			ErrSourceRoleInvalid,
			source.Role,
		)
	}

	fromMissing := source.ObservedFrom.IsZero()
	toMissing := source.ObservedTo.IsZero()

	if fromMissing != toMissing {
		return ErrSourceObservationRangeIncomplete
	}

	if !fromMissing && source.ObservedTo.Before(source.ObservedFrom) {
		return ErrSourceObservationRangeInvalid
	}

	return validateNotices(
		"source limitations",
		source.Limitations,
	)
}

func validateNotices(name string, notices []Notice) error {
	seen := make(
		map[string]struct{},
		len(notices),
	)

	for index, notice := range notices {
		if err := notice.Validate(); err != nil {
			return fmt.Errorf(
				"validate %s notice at index %d: %w",
				name,
				index,
				err,
			)
		}

		if _, exists := seen[notice.Code]; exists {
			return fmt.Errorf(
				"%w: collection=%s code=%q",
				ErrDuplicateNoticeCode,
				name,
				notice.Code,
			)
		}

		seen[notice.Code] = struct{}{}
	}

	return nil
}

func (notice Notice) Validate() error {
	if err := validateMachineCode(notice.Code); err != nil {
		return err
	}

	if strings.TrimSpace(notice.Message) == "" ||
		notice.Message != strings.TrimSpace(notice.Message) {
		return ErrNoticeMessageMissing
	}

	return nil
}

func validateMachineCode(value string) error {
	if value == "" || value != strings.TrimSpace(value) {
		return fmt.Errorf(
			"%w: %q",
			ErrMachineCodeInvalid,
			value,
		)
	}

	for index, character := range value {
		if index == 0 && !unicode.IsLower(character) {
			return fmt.Errorf(
				"%w: %q",
				ErrMachineCodeInvalid,
				value,
			)
		}

		if unicode.IsLower(character) ||
			unicode.IsDigit(character) ||
			character == '_' ||
			character == '-' ||
			character == '.' {
			continue
		}

		return fmt.Errorf(
			"%w: %q",
			ErrMachineCodeInvalid,
			value,
		)
	}

	return nil
}

func isKnownCapability(capability trajectoryeligibility.Capability) bool {
	switch capability {
	case trajectoryeligibility.CapabilityTrafficMetrics,
		trajectoryeligibility.CapabilityAirportActivity,
		trajectoryeligibility.CapabilityRouteInference,
		trajectoryeligibility.CapabilityHistoricalAggregation,
		trajectoryeligibility.CapabilityProjection:
		return true
	default:
		return false
	}
}
