package featurepipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

func TestNewRequiresEveryDependency(t *testing.T) {
	featureExtractor := &fakeExtractor{}
	featureValidator := &fakeValidator{}
	store := newRecordingStore(nil)

	tests := []struct {
		name   string
		config Config
		want   error
	}{
		{
			name: "extractor",
			config: Config{
				Validator: featureValidator,
				Writer:    store,
			},
			want: ErrExtractorRequired,
		},
		{
			name: "validator",
			config: Config{
				Extractor: featureExtractor,
				Writer:    store,
			},
			want: ErrValidatorRequired,
		},
		{
			name: "store",
			config: Config{
				Extractor: featureExtractor,
				Validator: featureValidator,
			},
			want: ErrStoreRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.config)
			if !errors.Is(err, test.want) {
				t.Fatalf(
					"New() error = %v, want %v",
					err,
					test.want,
				)
			}
		})
	}
}

func TestPipelineProcessesInStrictOrderAndStoresValidatedCopy(
	t *testing.T,
) {
	fixedNow := time.Date(
		2026,
		time.July,
		14,
		17,
		0,
		0,
		0,
		time.UTC,
	)
	asOfTime := fixedNow.Add(-time.Minute)
	calls := make([]string, 0, 3)

	featureExtractor := &fakeExtractor{
		extract: func(
			ctx context.Context,
			request extractor.Request,
		) (flightfeatures.FlightFeatures, error) {
			calls = append(calls, "extract")
			if err := ctx.Err(); err != nil {
				return flightfeatures.FlightFeatures{}, err
			}

			request.Trajectory.Points[0].Latitude = 99

			return storableFeatures(
				flightfeatures.ValidationStatusUnvalidated,
				"fingerprint-a",
				asOfTime,
			), nil
		},
	}
	featureValidator := &fakeValidator{
		validate: func(
			ctx context.Context,
			features flightfeatures.FlightFeatures,
		) (
			flightfeatures.FlightFeatures,
			validator.Report,
			error,
		) {
			calls = append(calls, "validate")
			if err := ctx.Err(); err != nil {
				return flightfeatures.FlightFeatures{},
					validator.Report{},
					err
			}
			if features.Quality.Status !=
				flightfeatures.ValidationStatusUnvalidated {
				t.Fatalf(
					"validator input status = %q",
					features.Quality.Status,
				)
			}

			features.Quality.Status =
				flightfeatures.ValidationStatusValid
			return features,
				validator.Report{
					ValidatorVersion: validator.Version,
					Status:           flightfeatures.ValidationStatusValid,
					ValidatedAt:      fixedNow,
				},
				nil
		},
	}
	store := newRecordingStore(&calls)
	store.memory = featurestore.NewMemory(
		featurestore.MemoryConfig{
			Now: func() time.Time {
				return fixedNow
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
		Trajectory: trajectory.FlightTrajectory{
			ID:          "88e4f622-a4cb-57e4-9708-1bb0a1403624",
			IdentityKey: "identity-input",
			Points: []trajectory.TrackPoint4D{
				{
					Latitude: 40,
				},
			},
		},
		AsOfTime: asOfTime,
	}
	original := cloneRequest(request)

	result, err := pipeline.Process(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if !reflect.DeepEqual(
		calls,
		[]string{"extract", "validate", "store"},
	) {
		t.Fatalf(
			"call order = %#v",
			calls,
		)
	}
	if result.PipelineVersion != Version ||
		result.Features().Quality.Status !=
			flightfeatures.ValidationStatusValid ||
		result.ValidationReport.Status !=
			flightfeatures.ValidationStatusValid ||
		result.Record.ID == "" ||
		result.Record.Features.Quality.Status !=
			flightfeatures.ValidationStatusValid ||
		!result.Record.StoredAt.Equal(fixedNow) {
		t.Fatalf(
			"unexpected result: %#v",
			result,
		)
	}
	if !reflect.DeepEqual(request, original) {
		t.Fatalf(
			"request mutated\nrequest=%#v\noriginal=%#v",
			request,
			original,
		)
	}

	stored, err := store.Get(
		context.Background(),
		result.Record.Key,
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !reflect.DeepEqual(
		stored,
		result.Record,
	) {
		t.Fatalf(
			"stored record = %#v, result record = %#v",
			stored,
			result.Record,
		)
	}
}

func TestPipelineRejectsInvalidWithoutStoreWrite(t *testing.T) {
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		17,
		0,
		0,
		0,
		time.UTC,
	)
	store := newRecordingStore(nil)
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
				features: storableFeatures(
					flightfeatures.ValidationStatusInvalid,
					"fingerprint-a",
					asOfTime,
				),
				report: validator.Report{
					Status:     flightfeatures.ValidationStatusInvalid,
					ErrorCount: 1,
					Issues: []validator.Issue{
						{
							Code:     "validator.test",
							Severity: validator.IssueSeverityError,
						},
					},
				},
			},
			Writer: store,
		},
	)

	result, err := pipeline.Process(
		context.Background(),
		extractor.Request{},
	)
	if !errors.Is(err, ErrValidationRejected) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			ErrValidationRejected,
		)
	}
	var rejected *ValidationRejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf(
			"Process() error = %T, want *ValidationRejectedError",
			err,
		)
	}
	if rejected.Status !=
		flightfeatures.ValidationStatusInvalid ||
		rejected.Report.ErrorCount != 1 ||
		rejected.Features.Quality.Status !=
			flightfeatures.ValidationStatusInvalid ||
		result.ValidationReport.Status !=
			flightfeatures.ValidationStatusInvalid {
		t.Fatalf(
			"unexpected rejection result: result=%#v error=%#v",
			result,
			rejected,
		)
	}
	if store.putCount != 0 {
		t.Fatalf(
			"Store.Put called %d times, want 0",
			store.putCount,
		)
	}
}

func TestPipelineRejectsValidationStatusMismatch(
	t *testing.T,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		17,
		0,
		0,
		0,
		time.UTC,
	)
	store := newRecordingStore(nil)
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
				features: storableFeatures(
					flightfeatures.ValidationStatusValid,
					"fingerprint-a",
					asOfTime,
				),
				report: validator.Report{
					Status: flightfeatures.ValidationStatusLimited,
				},
			},
			Writer: store,
		},
	)

	_, err := pipeline.Process(
		context.Background(),
		extractor.Request{},
	)
	if !errors.Is(
		err,
		ErrValidationStatusMismatch,
	) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			ErrValidationStatusMismatch,
		)
	}
	if store.putCount != 0 {
		t.Fatalf(
			"Store.Put called %d times, want 0",
			store.putCount,
		)
	}
}

func TestPipelineWrapsTechnicalStageErrors(t *testing.T) {
	extractionErr := errors.New("extract failed")
	validationErr := errors.New("validate failed")
	storageErr := errors.New("store failed")
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		17,
		0,
		0,
		0,
		time.UTC,
	)

	tests := []struct {
		name      string
		extractor FeatureExtractor
		validator FeatureValidator
		store     *recordingStore
		wantStage Stage
		wantErr   error
	}{
		{
			name: "extraction",
			extractor: &fakeExtractor{
				err: extractionErr,
			},
			validator: &fakeValidator{},
			store:     newRecordingStore(nil),
			wantStage: StageExtraction,
			wantErr:   extractionErr,
		},
		{
			name: "validation",
			extractor: &fakeExtractor{
				features: storableFeatures(
					flightfeatures.ValidationStatusUnvalidated,
					"fingerprint-a",
					asOfTime,
				),
			},
			validator: &fakeValidator{
				err: validationErr,
			},
			store:     newRecordingStore(nil),
			wantStage: StageValidation,
			wantErr:   validationErr,
		},
		{
			name: "storage",
			extractor: &fakeExtractor{
				features: storableFeatures(
					flightfeatures.ValidationStatusUnvalidated,
					"fingerprint-a",
					asOfTime,
				),
			},
			validator: &fakeValidator{
				features: storableFeatures(
					flightfeatures.ValidationStatusValid,
					"fingerprint-a",
					asOfTime,
				),
				report: validator.Report{
					Status: flightfeatures.ValidationStatusValid,
				},
			},
			store: &recordingStore{
				putErr: storageErr,
			},
			wantStage: StageStorage,
			wantErr:   storageErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pipeline := newTestPipeline(
				t,
				Config{
					Extractor: test.extractor,
					Validator: test.validator,
					Writer:    test.store,
				},
			)

			_, err := pipeline.Process(
				context.Background(),
				extractor.Request{},
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf(
					"Process() error = %v, want wrapped %v",
					err,
					test.wantErr,
				)
			}
			var stageErr *StageError
			if !errors.As(err, &stageErr) {
				t.Fatalf(
					"Process() error = %T, want *StageError",
					err,
				)
			}
			if stageErr.Stage != test.wantStage {
				t.Fatalf(
					"stage = %q, want %q",
					stageErr.Stage,
					test.wantStage,
				)
			}
		})
	}
}

func TestPipelinePreservesContextErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	pipeline := newTestPipeline(
		t,
		Config{
			Extractor: &fakeExtractor{
				err: context.Canceled,
			},
			Validator: &fakeValidator{},
			Writer:    newRecordingStore(nil),
		},
	)

	_, err := pipeline.Process(
		ctx,
		extractor.Request{},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Process() error = %v, want context.Canceled",
			err,
		)
	}
	var stageErr *StageError
	if errors.As(err, &stageErr) {
		t.Fatal(
			"context cancellation must not be wrapped in StageError",
		)
	}
}

func TestPipelineIsIdempotentAndSurfacesSnapshotConflict(
	t *testing.T,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		17,
		0,
		0,
		0,
		time.UTC,
	)
	fingerprint := "fingerprint-a"
	featureExtractor := &fakeExtractor{
		extract: func(
			context.Context,
			extractor.Request,
		) (flightfeatures.FlightFeatures, error) {
			return storableFeatures(
				flightfeatures.ValidationStatusUnvalidated,
				fingerprint,
				asOfTime,
			), nil
		},
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
			features.Quality.Status =
				flightfeatures.ValidationStatusValid
			return features,
				validator.Report{
					Status: flightfeatures.ValidationStatusValid,
				},
				nil
		},
	}
	store := featurestore.NewMemory(
		featurestore.MemoryConfig{},
	)
	pipeline := newTestPipeline(
		t,
		Config{
			Extractor: featureExtractor,
			Validator: featureValidator,
			Writer:    store,
		},
	)

	first, err := pipeline.Process(
		context.Background(),
		extractor.Request{},
	)
	if err != nil {
		t.Fatalf("first Process() error = %v", err)
	}
	second, err := pipeline.Process(
		context.Background(),
		extractor.Request{},
	)
	if err != nil {
		t.Fatalf("second Process() error = %v", err)
	}
	if first.Record.ID != second.Record.ID {
		t.Fatalf(
			"idempotent record IDs differ: %q and %q",
			first.Record.ID,
			second.Record.ID,
		)
	}

	page, err := store.List(
		context.Background(),
		featurestore.ListQuery{
			TrajectoryID:  "a7bd4786-5c83-5ea5-9eb0-354a460eaa91",
			SchemaVersion: flightfeatures.SchemaVersionV1,
		},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(page.Records) != 1 {
		t.Fatalf(
			"record count = %d, want 1",
			len(page.Records),
		)
	}

	fingerprint = "fingerprint-b"
	_, err = pipeline.Process(
		context.Background(),
		extractor.Request{},
	)
	if !errors.Is(
		err,
		featurestore.ErrSnapshotConflict,
	) {
		t.Fatalf(
			"conflict Process() error = %v, want %v",
			err,
			featurestore.ErrSnapshotConflict,
		)
	}
	var stageErr *StageError
	if !errors.As(err, &stageErr) ||
		stageErr.Stage != StageStorage {
		t.Fatalf(
			"conflict error = %#v, want storage StageError",
			err,
		)
	}

	stored, err := store.Get(
		context.Background(),
		first.Record.Key,
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if stored.InputFingerprint != canonicalTestFingerprint("fingerprint-a") {
		t.Fatalf(
			"existing record changed to %q",
			stored.InputFingerprint,
		)
	}
}

func TestResultCloneDoesNotShareMutableSlices(t *testing.T) {
	result := Result{
		ValidationReport: validator.Report{
			Issues: []validator.Issue{
				{Code: "issue"},
			},
		},
		Record: featurestore.Record{
			Features: flightfeatures.FlightFeatures{
				Quality: flightfeatures.FeatureQuality{
					Limitations: []flightfeatures.FeatureLimitation{
						{Code: "record"},
					},
				},
				Provenance: flightfeatures.FeatureProvenance{
					SourceNames: []string{"source"},
				},
			},
		},
	}

	cloned := result.Clone()
	cloned.ValidationReport.Issues[0].Code =
		"changed"
	cloned.Record.Features.Quality.Limitations[0].Code =
		"changed"
	cloned.Record.Features.Provenance.SourceNames[0] =
		"changed"

	if result.ValidationReport.Issues[0].Code !=
		"issue" ||
		result.Record.Features.Quality.Limitations[0].Code !=
			"record" ||
		result.Record.Features.Provenance.SourceNames[0] !=
			"source" {
		t.Fatal("Result.Clone() shared mutable slices")
	}
}

type fakeExtractor struct {
	features flightfeatures.FlightFeatures
	err      error
	extract  func(
		context.Context,
		extractor.Request,
	) (flightfeatures.FlightFeatures, error)
}

func (item *fakeExtractor) Extract(
	ctx context.Context,
	request extractor.Request,
) (flightfeatures.FlightFeatures, error) {
	if item.extract != nil {
		return item.extract(ctx, request)
	}
	if item.err != nil {
		return flightfeatures.FlightFeatures{}, item.err
	}

	return item.features.Clone(), nil
}

type fakeValidator struct {
	features flightfeatures.FlightFeatures
	report   validator.Report
	err      error
	validate func(
		context.Context,
		flightfeatures.FlightFeatures,
	) (
		flightfeatures.FlightFeatures,
		validator.Report,
		error,
	)
}

func (item *fakeValidator) Validate(
	ctx context.Context,
	features flightfeatures.FlightFeatures,
) (
	flightfeatures.FlightFeatures,
	validator.Report,
	error,
) {
	var (
		validated flightfeatures.FlightFeatures
		report    validator.Report
		err       error
	)
	if item.validate != nil {
		validated, report, err = item.validate(ctx, features)
	} else if item.err != nil {
		return flightfeatures.FlightFeatures{},
			validator.Report{},
			item.err
	} else {
		validated = item.features.Clone()
		report = item.report.Clone()
	}
	if err != nil {
		return validated, report, err
	}

	if report.ValidatorVersion == "" {
		report.ValidatorVersion = validator.Version
	}
	if report.ValidatedAt.IsZero() {
		report.ValidatedAt = validated.Window.AsOfTime.UTC()
		if report.ValidatedAt.IsZero() {
			report.ValidatedAt = time.Date(
				2026,
				time.July,
				25,
				8,
				0,
				0,
				0,
				time.UTC,
			)
		}
	}
	if report.Status ==
		flightfeatures.ValidationStatusLimited &&
		len(report.Issues) == 0 {
		report.WarningCount = 1
		report.Issues = []validator.Issue{
			{
				Code:     "test.synthetic.warning",
				Severity: validator.IssueSeverityWarning,
			},
		}
	}
	if report.Status ==
		flightfeatures.ValidationStatusInvalid &&
		len(report.Issues) == 0 {
		report.Issues = []validator.Issue{
			{
				Code:     "test.synthetic.error",
				Severity: validator.IssueSeverityError,
			},
		}
	}

	report.ErrorCount = 0
	report.WarningCount = 0
	for _, issue := range report.Issues {
		switch issue.Severity {
		case validator.IssueSeverityError:
			report.ErrorCount++
		case validator.IssueSeverityWarning:
			report.WarningCount++
		}
	}

	return validated.Clone(), report.Clone(), nil
}

type recordingStore struct {
	memory   *featurestore.MemoryStore
	calls    *[]string
	putCount int
	putErr   error
}

func newRecordingStore(
	calls *[]string,
) *recordingStore {
	return &recordingStore{
		memory: featurestore.NewMemory(
			featurestore.MemoryConfig{},
		),
		calls: calls,
	}
}

func (store *recordingStore) Put(
	ctx context.Context,
	features flightfeatures.FlightFeatures,
) (featurestore.Record, error) {
	store.putCount++
	if store.calls != nil {
		*store.calls = append(
			*store.calls,
			"store",
		)
	}
	if store.putErr != nil {
		return featurestore.Record{}, store.putErr
	}
	if store.memory == nil {
		store.memory = featurestore.NewMemory(
			featurestore.MemoryConfig{},
		)
	}

	return store.memory.Put(ctx, features)
}

func (store *recordingStore) Get(
	ctx context.Context,
	key featurestore.SnapshotKey,
) (featurestore.Record, error) {
	return store.memory.Get(ctx, key)
}

func (store *recordingStore) GetLatest(
	ctx context.Context,
	trajectoryID string,
	schemaVersion flightfeatures.SchemaVersion,
) (featurestore.Record, error) {
	return store.memory.GetLatest(
		ctx,
		trajectoryID,
		schemaVersion,
	)
}

func (store *recordingStore) List(
	ctx context.Context,
	query featurestore.ListQuery,
) (featurestore.Page, error) {
	return store.memory.List(ctx, query)
}

func storableFeatures(
	status flightfeatures.ValidationStatus,
	fingerprint string,
	asOfTime time.Time,
) flightfeatures.FlightFeatures {
	features := flightfeatures.FlightFeatures{
		SchemaVersion: flightfeatures.SchemaVersionV1,
		TrajectoryID:  "a7bd4786-5c83-5ea5-9eb0-354a460eaa91",
		IdentityKey:   "identity-1",
		ICAO24:        "ABC123",
		Window: flightfeatures.FeatureWindow{
			StartTime: asOfTime.Add(-time.Hour),
			EndTime:   asOfTime.Add(-time.Minute),
			AsOfTime:  asOfTime,
		},
		Quality: flightfeatures.FeatureQuality{
			Status: status,
		},
		Provenance: flightfeatures.FeatureProvenance{
			ExtractorVersion: "test-extractor",
			InputFingerprint: canonicalTestFingerprint(fingerprint),
		},
	}
	if status == flightfeatures.ValidationStatusValid ||
		status == flightfeatures.ValidationStatusLimited {
		features.ValidationReport = flightfeatures.ValidationReport{
			AuditState:       flightfeatures.ValidationAuditStateComplete,
			ValidatorVersion: validator.Version,
			Status:           status,
			ValidatedAt:      asOfTime.UTC(),
		}
	}
	return features
}

func canonicalTestFingerprint(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func newTestPipeline(
	t *testing.T,
	config Config,
) *Pipeline {
	t.Helper()

	pipeline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return pipeline
}
