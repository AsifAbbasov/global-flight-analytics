package dataquality

import domainconfidence "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/confidence"

type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusPartial ValidationStatus = "partial"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

type CompletenessLevel string

const (
	CompletenessLevelComplete     CompletenessLevel = "complete"
	CompletenessLevelPartial      CompletenessLevel = "partial"
	CompletenessLevelPositionOnly CompletenessLevel = "position_only"
	CompletenessLevelInsufficient CompletenessLevel = "insufficient"
)

type ConfidenceLevel = domainconfidence.Level

const (
	ConfidenceLevelHigh   = domainconfidence.LevelHigh
	ConfidenceLevelMedium = domainconfidence.LevelMedium
	ConfidenceLevelLow    = domainconfidence.LevelLow
	ConfidenceLevelNone   = domainconfidence.LevelNone
)

type Warning struct {
	Code    string
	Message string
	Field   string
}

type DataQuality struct {
	ValidationStatus ValidationStatus
	Completeness     CompletenessLevel
	Confidence       ConfidenceLevel
	Score            float64
	MissingFields    []string
	Warnings         []Warning
}

// STAGE-14-1-ARCHITECTURE-CONSOLIDATION-V1-1
