package dataquality

import (
	"errors"
	"testing"
)

func TestDataQualityValidateRejectsUnknownStateAndOutOfRangeScore(t *testing.T) {
	value := DataQuality{ValidationStatus: ValidationStatus("broken"), Completeness: CompletenessLevelComplete, Confidence: ConfidenceLevelHigh, Score: 1}
	if err := value.Validate(); !errors.Is(err, ErrValidationStatusInvalid) {
		t.Fatalf("status error = %v", err)
	}
	value.ValidationStatus = ValidationStatusValid
	value.Score = 1.1
	if err := value.Validate(); !errors.Is(err, ErrScoreInvalid) {
		t.Fatalf("score error = %v", err)
	}
}

func TestDataQualityNormalizeCollectionsRemovesNilSlices(t *testing.T) {
	normalized := (DataQuality{}).NormalizeCollections()
	if normalized.MissingFields == nil || normalized.Warnings == nil {
		t.Fatalf("collections remain nil: %+v", normalized)
	}
}
