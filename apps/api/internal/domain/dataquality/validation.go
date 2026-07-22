package dataquality

import (
	"errors"
	"math"
)

var (
	ErrValidationStatusInvalid = errors.New("data quality validation status is invalid")
	ErrCompletenessInvalid     = errors.New("data quality completeness level is invalid")
	ErrConfidenceInvalid       = errors.New("data quality confidence level is invalid")
	ErrScoreInvalid            = errors.New("data quality score must be finite and between zero and one")
)

func (value DataQuality) Validate() error {
	switch value.ValidationStatus {
	case ValidationStatusValid, ValidationStatusPartial, ValidationStatusInvalid:
	default:
		return ErrValidationStatusInvalid
	}
	switch value.Completeness {
	case CompletenessLevelComplete, CompletenessLevelPartial, CompletenessLevelPositionOnly, CompletenessLevelInsufficient:
	default:
		return ErrCompletenessInvalid
	}
	if value.Confidence.Validate() != nil {
		return ErrConfidenceInvalid
	}
	if math.IsNaN(value.Score) || math.IsInf(value.Score, 0) || value.Score < 0 || value.Score > 1 {
		return ErrScoreInvalid
	}
	return nil
}

func (value DataQuality) NormalizeCollections() DataQuality {
	if value.MissingFields == nil {
		value.MissingFields = make([]string, 0)
	}
	if value.Warnings == nil {
		value.Warnings = make([]Warning, 0)
	}
	return value
}
