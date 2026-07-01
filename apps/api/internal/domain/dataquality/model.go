package dataquality

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

type ConfidenceLevel string

const (
	ConfidenceLevelHigh   ConfidenceLevel = "high"
	ConfidenceLevelMedium ConfidenceLevel = "medium"
	ConfidenceLevelLow    ConfidenceLevel = "low"
	ConfidenceLevelNone   ConfidenceLevel = "none"
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
