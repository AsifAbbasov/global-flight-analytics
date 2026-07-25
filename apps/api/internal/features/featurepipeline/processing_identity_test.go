package featurepipeline

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

func TestPipelineStampsConfiguredProcessingVersion(t *testing.T) {
	asOf := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)
	pipeline := newTestPipeline(
		t,
		Config{
			Extractor: &fakeExtractor{
				features: storableFeatures(
					flightfeatures.ValidationStatusUnvalidated,
					"fingerprint-a",
					asOf,
				),
			},
			Validator: &fakeValidator{
				validate: func(
					_ context.Context,
					features flightfeatures.FlightFeatures,
				) (
					flightfeatures.FlightFeatures,
					validator.Report,
					error,
				) {
					features.Quality.Status =
						flightfeatures.ValidationStatusValid
					return features,
						validator.Report{
							ValidatorVersion: validator.Version,
							Status:           flightfeatures.ValidationStatusValid,
							ValidatedAt:      asOf,
						},
						nil
				},
			},
			Writer:            newRecordingStore(nil),
			ProcessingVersion: "flight-feature-processing-pipeline-test-v2",
		},
	)

	result, err := pipeline.Process(
		context.Background(),
		extractor.Request{},
	)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if result.Record.Key.ProcessingVersion !=
		"flight-feature-processing-pipeline-test-v2" {
		t.Fatalf(
			"key processing version = %q",
			result.Record.Key.ProcessingVersion,
		)
	}
}

func TestCurrentProcessingVersionMatchesPipelineVersion(t *testing.T) {
	if flightfeatures.CurrentProcessingVersion !=
		flightfeatures.ProcessingVersion(Version) {
		t.Fatalf(
			"current processing version = %q, pipeline version = %q",
			flightfeatures.CurrentProcessingVersion,
			Version,
		)
	}
}
