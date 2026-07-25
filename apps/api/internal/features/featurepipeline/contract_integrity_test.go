package featurepipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

func TestNewRejectsTypedNilDependencies(t *testing.T) {
	var typedNilExtractor *fakeExtractor
	var typedNilValidator *fakeValidator
	var typedNilWriter *recordingStore

	tests := []struct {
		name   string
		config Config
		want   error
	}{
		{
			name: "extractor",
			config: Config{
				Extractor: typedNilExtractor,
				Validator: &fakeValidator{},
				Writer:    newRecordingStore(nil),
			},
			want: ErrExtractorRequired,
		},
		{
			name: "validator",
			config: Config{
				Extractor: &fakeExtractor{},
				Validator: typedNilValidator,
				Writer:    newRecordingStore(nil),
			},
			want: ErrValidatorRequired,
		},
		{
			name: "writer",
			config: Config{
				Extractor: &fakeExtractor{},
				Validator: &fakeValidator{},
				Writer:    typedNilWriter,
			},
			want: ErrWriterRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.config)
			if !errors.Is(err, test.want) {
				t.Fatalf("New() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestPipelineRejectsNilContext(t *testing.T) {
	pipeline := newTestPipeline(
		t,
		Config{
			Extractor: &fakeExtractor{},
			Validator: &fakeValidator{},
			Writer:    newRecordingStore(nil),
		},
	)

	_, err := pipeline.Process(nil, extractor.Request{})
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			ErrContextRequired,
		)
	}
}

func TestPipelineRejectsIncompleteValidationReport(t *testing.T) {
	asOfTime := time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC)

	pipeline := newTestPipeline(
		t,
		Config{
			Extractor: &fakeExtractor{
				features: storableFeatures(
					flightfeatures.ValidationStatusUnvalidated,
					"fingerprint-a",
					asOfTime,
				),
			},
			Validator: rawIncompleteValidator{
				features: storableFeatures(
					flightfeatures.ValidationStatusValid,
					"fingerprint-a",
					asOfTime,
				),
				report: validator.Report{
					Status: flightfeatures.ValidationStatusValid,
				},
			},
			Writer: newRecordingStore(nil),
		},
	)

	_, err := pipeline.Process(context.Background(), extractor.Request{})
	if !errors.Is(err, validator.ErrInvalidReport) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			validator.ErrInvalidReport,
		)
	}

	var stageErr *StageError
	if !errors.As(err, &stageErr) ||
		stageErr.Stage != StageValidation {
		t.Fatalf(
			"Process() error = %#v, want validation StageError",
			err,
		)
	}
}

func TestResultFeaturesUsesStoredNormalizedRecord(t *testing.T) {
	asOfTime := time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC)
	validatedAt := asOfTime.Add(time.Second)

	pipeline := newTestPipeline(
		t,
		Config{
			Extractor: &fakeExtractor{
				features: storableFeatures(
					flightfeatures.ValidationStatusUnvalidated,
					"fingerprint-a",
					asOfTime,
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
					features.ICAO24 = " abc123 "
					features.Quality.Status =
						flightfeatures.ValidationStatusValid

					return features,
						validator.Report{
							ValidatorVersion: validator.Version,
							Status:           flightfeatures.ValidationStatusValid,
							ValidatedAt:      validatedAt,
						},
						nil
				},
			},
			Writer: featurestore.NewMemory(
				featurestore.MemoryConfig{},
			),
		},
	)

	result, err := pipeline.Process(
		context.Background(),
		extractor.Request{},
	)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	features := result.Features()
	if features.ICAO24 != "ABC123" {
		t.Fatalf(
			"Result.Features().ICAO24 = %q, want ABC123",
			features.ICAO24,
		)
	}
	if result.Record.Features.ICAO24 != "ABC123" {
		t.Fatalf(
			"Record.Features.ICAO24 = %q, want ABC123",
			result.Record.Features.ICAO24,
		)
	}
}

type rawIncompleteValidator struct {
	features flightfeatures.FlightFeatures
	report   validator.Report
}

func (item rawIncompleteValidator) Validate(
	context.Context,
	flightfeatures.FlightFeatures,
) (
	flightfeatures.FlightFeatures,
	validator.Report,
	error,
) {
	return item.features.Clone(), item.report.Clone(), nil
}
