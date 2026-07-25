package featurestore

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

func normalizeValidationReport(
	features *flightfeatures.FlightFeatures,
) {
	if features == nil {
		return
	}

	features.ValidationReport =
		validator.NormalizeStoredReport(
			features.ValidationReport,
			features.Quality.Status,
		)
}

func validateStoredValidationReport(
	features flightfeatures.FlightFeatures,
) error {
	return validator.ValidateStoredReport(
		features.ValidationReport,
		features.Quality.Status,
	)
}
