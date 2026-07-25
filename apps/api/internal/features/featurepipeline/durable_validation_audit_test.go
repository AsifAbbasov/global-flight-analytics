package featurepipeline

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

func TestPipelinePersistsAndReplaysStoredValidationReport(
	t *testing.T,
) {
	firstValidatedAt := time.Date(
		2026,
		time.July,
		25,
		13,
		0,
		0,
		123456789,
		time.UTC,
	)
	asOfTime := firstValidatedAt.Add(-time.Minute)
	validationCall := 0

	featureExtractor := &fakeExtractor{
		features: storableFeatures(
			flightfeatures.ValidationStatusUnvalidated,
			"fingerprint-validation-audit",
			asOfTime,
		),
	}
	featureValidator := &fakeValidator{
		validate: func(
			_ context.Context,
			features flightfeatures.FlightFeatures,
		) (
			flightfeatures.FlightFeatures,
			validator.Report,
			error,
		) {
			validationCall++
			features.Quality.Status =
				flightfeatures.ValidationStatusLimited

			return features,
				validator.Report{
					AuditState:       validator.AuditStateComplete,
					ValidatorVersion: validator.Version,
					Status: flightfeatures.
						ValidationStatusLimited,
					WarningCount: 1,
					Issues: []validator.Issue{
						{
							Code:    "validation.audit.warning",
							Message: "durable warning",
							Path:    "quality",
							Group: flightfeatures.
								FeatureGroupTemporal,
							Severity: validator.
								IssueSeverityWarning,
						},
					},
					ValidatedAt: firstValidatedAt.Add(
						time.Duration(
							validationCall-1,
						) * time.Second,
					),
				},
				nil
		},
	}
	store := featurestore.NewMemory(
		featurestore.MemoryConfig{
			Now: func() time.Time {
				return firstValidatedAt.Add(time.Minute)
			},
		},
	)
	pipeline := newTestPipeline(
		t,
		Config{
			Extractor: featureExtractor,
			Validator: featureValidator,
			Writer:    store,
		},
	)

	request := extractor.Request{
		AsOfTime: asOfTime,
	}
	first, err := pipeline.Process(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf("first Process() error = %v", err)
	}
	second, err := pipeline.Process(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf("second Process() error = %v", err)
	}

	storedReport := first.Record.Features.ValidationReport
	if storedReport.AuditState !=
		validator.AuditStateComplete ||
		storedReport.ValidatorVersion != validator.Version ||
		!storedReport.ValidatedAt.Equal(firstValidatedAt) ||
		len(storedReport.Issues) != 1 {
		t.Fatalf("stored report = %#v", storedReport)
	}
	if !reflect.DeepEqual(
		first.ValidationReport,
		storedReport,
	) {
		t.Fatalf(
			"first result report differs from stored report",
		)
	}
	if !reflect.DeepEqual(
		second.ValidationReport,
		storedReport,
	) {
		t.Fatalf(
			"replay returned transient report: %#v",
			second.ValidationReport,
		)
	}
	if !reflect.DeepEqual(first.Record, second.Record) {
		t.Fatalf(
			"idempotent records differ\nfirst=%#v\nsecond=%#v",
			first.Record,
			second.Record,
		)
	}

	loaded, err := store.Get(
		context.Background(),
		first.Record.Key,
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !reflect.DeepEqual(
		loaded.Features.ValidationReport,
		storedReport,
	) {
		t.Fatalf(
			"loaded report = %#v, want %#v",
			loaded.Features.ValidationReport,
			storedReport,
		)
	}
}
