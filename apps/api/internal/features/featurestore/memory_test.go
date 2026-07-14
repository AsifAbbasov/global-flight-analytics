package featurestore

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestMemoryStorePutAndGet(t *testing.T) {
	storedAt := time.Date(
		2026,
		time.July,
		14,
		12,
		30,
		0,
		0,
		time.UTC,
	)
	store := NewMemory(MemoryConfig{
		Now: func() time.Time {
			return storedAt
		},
	})
	features := validStoredFeatures(
		"trajectory-one",
		time.Date(
			2026,
			time.July,
			14,
			10,
			0,
			0,
			0,
			time.UTC,
		),
		"a",
	)

	record, err := store.Put(
		context.Background(),
		features,
	)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	if !strings.HasPrefix(record.ID, recordIDPrefix) ||
		len(record.ID) != len(recordIDPrefix)+64 {
		t.Fatalf("unexpected record id: %q", record.ID)
	}
	if record.Key.TrajectoryID != "trajectory-one" ||
		record.Key.SchemaVersion !=
			flightfeatures.SchemaVersionV1 ||
		!record.Key.AsOfTime.Equal(
			features.Window.AsOfTime,
		) {
		t.Fatalf("unexpected record key: %#v", record.Key)
	}
	if record.InputFingerprint !=
		features.Provenance.InputFingerprint {
		t.Fatalf(
			"fingerprint = %q, want %q",
			record.InputFingerprint,
			features.Provenance.InputFingerprint,
		)
	}
	if !record.StoredAt.Equal(storedAt) {
		t.Fatalf(
			"StoredAt = %v, want %v",
			record.StoredAt,
			storedAt,
		)
	}

	loaded, err := store.Get(
		context.Background(),
		record.Key,
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !reflect.DeepEqual(record, loaded) {
		t.Fatalf(
			"loaded record differs\nstored=%#v\nloaded=%#v",
			record,
			loaded,
		)
	}
}

func TestMemoryStorePutIsIdempotentForSameFingerprint(
	t *testing.T,
) {
	nowCalls := 0
	store := NewMemory(MemoryConfig{
		Now: func() time.Time {
			nowCalls++
			return time.Date(
				2026,
				time.July,
				14,
				12,
				nowCalls,
				0,
				0,
				time.UTC,
			)
		},
	})
	features := validStoredFeatures(
		"trajectory-one",
		time.Date(
			2026,
			time.July,
			14,
			10,
			0,
			0,
			0,
			time.UTC,
		),
		"a",
	)

	first, err := store.Put(
		context.Background(),
		features,
	)
	if err != nil {
		t.Fatalf("first Put() error = %v", err)
	}
	second, err := store.Put(
		context.Background(),
		features,
	)
	if err != nil {
		t.Fatalf("second Put() error = %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf(
			"idempotent writes differ\nfirst=%#v\nsecond=%#v",
			first,
			second,
		)
	}
	if nowCalls != 1 {
		t.Fatalf(
			"Now() calls = %d, want 1",
			nowCalls,
		)
	}
}

func TestMemoryStoreRejectsConflictingSnapshotEvidence(
	t *testing.T,
) {
	store := NewMemory(MemoryConfig{})
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		10,
		0,
		0,
		0,
		time.UTC,
	)
	first := validStoredFeatures(
		"trajectory-one",
		asOfTime,
		"a",
	)
	second := validStoredFeatures(
		"trajectory-one",
		asOfTime,
		"b",
	)

	if _, err := store.Put(
		context.Background(),
		first,
	); err != nil {
		t.Fatalf("first Put() error = %v", err)
	}
	if _, err := store.Put(
		context.Background(),
		second,
	); !errors.Is(err, ErrSnapshotConflict) {
		t.Fatalf(
			"second Put() error = %v, want %v",
			err,
			ErrSnapshotConflict,
		)
	}

	loaded, err := store.Get(
		context.Background(),
		SnapshotKey{
			TrajectoryID:  "trajectory-one",
			SchemaVersion: flightfeatures.SchemaVersionV1,
			AsOfTime:      asOfTime,
		},
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if loaded.InputFingerprint !=
		first.Provenance.InputFingerprint {
		t.Fatalf(
			"stored fingerprint changed to %q",
			loaded.InputFingerprint,
		)
	}
}

func TestMemoryStoreRejectsUnstorableFeatures(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*flightfeatures.FlightFeatures)
		wantErr error
	}{
		{
			name: "trajectory id",
			mutate: func(features *flightfeatures.FlightFeatures) {
				features.TrajectoryID = ""
			},
			wantErr: ErrTrajectoryIDRequired,
		},
		{
			name: "schema version",
			mutate: func(features *flightfeatures.FlightFeatures) {
				features.SchemaVersion = "future"
			},
			wantErr: ErrUnsupportedSchemaVersion,
		},
		{
			name: "as-of time",
			mutate: func(features *flightfeatures.FlightFeatures) {
				features.Window.AsOfTime = time.Time{}
			},
			wantErr: ErrAsOfTimeRequired,
		},
		{
			name: "fingerprint",
			mutate: func(features *flightfeatures.FlightFeatures) {
				features.Provenance.InputFingerprint = ""
			},
			wantErr: ErrInputFingerprintRequired,
		},
		{
			name: "unvalidated",
			mutate: func(features *flightfeatures.FlightFeatures) {
				features.Quality.Status =
					flightfeatures.ValidationStatusUnvalidated
			},
			wantErr: ErrFeaturesUnvalidated,
		},
		{
			name: "invalid",
			mutate: func(features *flightfeatures.FlightFeatures) {
				features.Quality.Status =
					flightfeatures.ValidationStatusInvalid
			},
			wantErr: ErrFeaturesInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMemory(MemoryConfig{})
			features := validStoredFeatures(
				"trajectory-one",
				time.Date(
					2026,
					time.July,
					14,
					10,
					0,
					0,
					0,
					time.UTC,
				),
				"a",
			)
			test.mutate(&features)

			_, err := store.Put(
				context.Background(),
				features,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf(
					"Put() error = %v, want %v",
					err,
					test.wantErr,
				)
			}
		})
	}
}

func TestMemoryStoreAcceptsLimitedFeatures(t *testing.T) {
	store := NewMemory(MemoryConfig{})
	features := validStoredFeatures(
		"trajectory-one",
		time.Date(
			2026,
			time.July,
			14,
			10,
			0,
			0,
			0,
			time.UTC,
		),
		"a",
	)
	features.Quality.Status =
		flightfeatures.ValidationStatusLimited

	record, err := store.Put(
		context.Background(),
		features,
	)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if record.Features.Quality.Status !=
		flightfeatures.ValidationStatusLimited {
		t.Fatalf(
			"stored status = %q",
			record.Features.Quality.Status,
		)
	}
}

func TestMemoryStoreGetLatestAndListUseDescendingAsOfTime(
	t *testing.T,
) {
	store := NewMemory(MemoryConfig{})
	base := time.Date(
		2026,
		time.July,
		14,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	for index, suffix := range []string{"a", "b", "c"} {
		_, err := store.Put(
			context.Background(),
			validStoredFeatures(
				"trajectory-one",
				base.Add(
					time.Duration(index)*time.Hour,
				),
				suffix,
			),
		)
		if err != nil {
			t.Fatalf(
				"Put(%d) error = %v",
				index,
				err,
			)
		}
	}

	latest, err := store.GetLatest(
		context.Background(),
		"trajectory-one",
		flightfeatures.SchemaVersionV1,
	)
	if err != nil {
		t.Fatalf("GetLatest() error = %v", err)
	}
	if !latest.Key.AsOfTime.Equal(
		base.Add(2 * time.Hour),
	) {
		t.Fatalf(
			"latest AsOfTime = %v",
			latest.Key.AsOfTime,
		)
	}

	page, err := store.List(
		context.Background(),
		ListQuery{
			TrajectoryID:  "trajectory-one",
			SchemaVersion: flightfeatures.SchemaVersionV1,
			Limit:         2,
		},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(page.Records) != 2 || !page.HasMore {
		t.Fatalf("unexpected page: %#v", page)
	}
	if !page.Records[0].Key.AsOfTime.Equal(
		base.Add(2*time.Hour),
	) || !page.Records[1].Key.AsOfTime.Equal(
		base.Add(time.Hour),
	) {
		t.Fatalf(
			"unexpected list order: %#v",
			page.Records,
		)
	}

	nextPage, err := store.List(
		context.Background(),
		ListQuery{
			TrajectoryID:   "trajectory-one",
			SchemaVersion:  flightfeatures.SchemaVersionV1,
			BeforeAsOfTime: page.Records[1].Key.AsOfTime,
			Limit:          2,
		},
	)
	if err != nil {
		t.Fatalf("next List() error = %v", err)
	}
	if len(nextPage.Records) != 1 ||
		nextPage.HasMore ||
		!nextPage.Records[0].Key.AsOfTime.Equal(base) {
		t.Fatalf(
			"unexpected next page: %#v",
			nextPage,
		)
	}
}

func TestMemoryStoreIsolatesTrajectories(t *testing.T) {
	store := NewMemory(MemoryConfig{})
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	for _, trajectoryID := range []string{
		"trajectory-one",
		"trajectory-two",
	} {
		_, err := store.Put(
			context.Background(),
			validStoredFeatures(
				trajectoryID,
				asOfTime,
				trajectoryID,
			),
		)
		if err != nil {
			t.Fatalf(
				"Put(%s) error = %v",
				trajectoryID,
				err,
			)
		}
	}

	page, err := store.List(
		context.Background(),
		ListQuery{
			TrajectoryID:  "trajectory-one",
			SchemaVersion: flightfeatures.SchemaVersionV1,
		},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(page.Records) != 1 ||
		page.Records[0].Key.TrajectoryID !=
			"trajectory-one" {
		t.Fatalf("unexpected isolated page: %#v", page)
	}
}

func TestMemoryStoreReturnsDefensiveCopies(t *testing.T) {
	store := NewMemory(MemoryConfig{})
	features := validStoredFeatures(
		"trajectory-one",
		time.Date(
			2026,
			time.July,
			14,
			10,
			0,
			0,
			0,
			time.UTC,
		),
		"a",
	)
	features.Quality.Limitations =
		[]flightfeatures.FeatureLimitation{
			{
				Code:    "original",
				Message: "Original limitation.",
			},
		}

	record, err := store.Put(
		context.Background(),
		features,
	)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	features.Quality.Limitations[0].Code = "input-changed"
	record.Features.Quality.Limitations[0].Code =
		"record-changed"

	loaded, err := store.Get(
		context.Background(),
		record.Key,
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if loaded.Features.Quality.Limitations[0].Code !=
		"original" {
		t.Fatalf(
			"stored limitation changed: %#v",
			loaded.Features.Quality.Limitations,
		)
	}

	loaded.Features.Quality.Limitations[0].Code =
		"loaded-changed"
	again, err := store.Get(
		context.Background(),
		record.Key,
	)
	if err != nil {
		t.Fatalf("second Get() error = %v", err)
	}
	if again.Features.Quality.Limitations[0].Code !=
		"original" {
		t.Fatal("Get() returned shared feature slices")
	}
}

func TestMemoryStoreConcurrentIdempotentPut(t *testing.T) {
	store := NewMemory(MemoryConfig{})
	features := validStoredFeatures(
		"trajectory-one",
		time.Date(
			2026,
			time.July,
			14,
			10,
			0,
			0,
			0,
			time.UTC,
		),
		"a",
	)

	const workerCount = 32
	results := make(chan Record, workerCount)
	errorsChannel := make(chan error, workerCount)
	var waitGroup sync.WaitGroup

	for worker := 0; worker < workerCount; worker++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			record, err := store.Put(
				context.Background(),
				features,
			)
			if err != nil {
				errorsChannel <- err
				return
			}
			results <- record
		}()
	}

	waitGroup.Wait()
	close(results)
	close(errorsChannel)

	for err := range errorsChannel {
		t.Fatalf("concurrent Put() error = %v", err)
	}

	ids := make([]string, 0, workerCount)
	for record := range results {
		ids = append(ids, record.ID)
	}
	if len(ids) != workerCount {
		t.Fatalf(
			"result count = %d, want %d",
			len(ids),
			workerCount,
		)
	}
	sort.Strings(ids)
	for index := 1; index < len(ids); index++ {
		if ids[index] != ids[0] {
			t.Fatalf(
				"concurrent ids differ: %#v",
				ids,
			)
		}
	}

	page, err := store.List(
		context.Background(),
		ListQuery{
			TrajectoryID:  "trajectory-one",
			SchemaVersion: flightfeatures.SchemaVersionV1,
		},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(page.Records) != 1 {
		t.Fatalf(
			"stored record count = %d, want 1",
			len(page.Records),
		)
	}
}

func TestMemoryStoreValidatesQueriesAndContext(t *testing.T) {
	store := NewMemory(MemoryConfig{})

	if _, err := store.Get(
		context.Background(),
		SnapshotKey{},
	); !errors.Is(err, ErrTrajectoryIDRequired) {
		t.Fatalf(
			"Get() error = %v, want trajectory id error",
			err,
		)
	}
	if _, err := store.GetLatest(
		context.Background(),
		"trajectory-one",
		"future",
	); !errors.Is(err, ErrUnsupportedSchemaVersion) {
		t.Fatalf(
			"GetLatest() error = %v",
			err,
		)
	}
	if _, err := store.List(
		context.Background(),
		ListQuery{
			TrajectoryID:  "trajectory-one",
			SchemaVersion: flightfeatures.SchemaVersionV1,
			Limit:         MaximumListLimit + 1,
		},
	); !errors.Is(err, ErrInvalidListLimit) {
		t.Fatalf(
			"List() error = %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.Put(
		ctx,
		validStoredFeatures(
			"trajectory-one",
			time.Now().UTC(),
			"a",
		),
	); !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Put() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestMemoryStoreReturnsNotFound(t *testing.T) {
	store := NewMemory(MemoryConfig{})
	key := SnapshotKey{
		TrajectoryID:  "trajectory-one",
		SchemaVersion: flightfeatures.SchemaVersionV1,
		AsOfTime: time.Date(
			2026,
			time.July,
			14,
			10,
			0,
			0,
			0,
			time.UTC,
		),
	}

	if _, err := store.Get(
		context.Background(),
		key,
	); !errors.Is(err, ErrSnapshotNotFound) {
		t.Fatalf(
			"Get() error = %v, want %v",
			err,
			ErrSnapshotNotFound,
		)
	}
	if _, err := store.GetLatest(
		context.Background(),
		"trajectory-one",
		flightfeatures.SchemaVersionV1,
	); !errors.Is(err, ErrSnapshotNotFound) {
		t.Fatalf(
			"GetLatest() error = %v, want %v",
			err,
			ErrSnapshotNotFound,
		)
	}
}

func validStoredFeatures(
	trajectoryID string,
	asOfTime time.Time,
	fingerprintSuffix string,
) flightfeatures.FlightFeatures {
	fingerprintCharacter := fingerprintSuffix
	if fingerprintCharacter == "" {
		fingerprintCharacter = "a"
	}
	fingerprintCharacter =
		string([]rune(fingerprintCharacter)[0])

	return flightfeatures.FlightFeatures{
		SchemaVersion: flightfeatures.SchemaVersionV1,
		TrajectoryID:  trajectoryID,
		IdentityKey:   "flight-identity-example",
		FlightID:      "flight-one",
		AircraftID:    "aircraft-one",
		ICAO24:        "ABC123",
		Callsign:      "TEST123",
		Window: flightfeatures.FeatureWindow{
			StartTime: asOfTime.Add(-time.Hour),
			EndTime:   asOfTime,
			AsOfTime:  asOfTime,
		},
		ExtractedAt: asOfTime.Add(time.Minute),
		Temporal: flightfeatures.TemporalFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:               flightfeatures.AvailabilityStatusAvailable,
				AvailableFieldCount:  8,
				TotalFieldCount:      8,
				SupportingPointCount: 4,
			},
		},
		Geographical: flightfeatures.GeographicalFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:               flightfeatures.AvailabilityStatusAvailable,
				AvailableFieldCount:  11,
				TotalFieldCount:      11,
				SupportingPointCount: 4,
			},
		},
		Operational: flightfeatures.OperationalFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:               flightfeatures.AvailabilityStatusAvailable,
				AvailableFieldCount:  11,
				TotalFieldCount:      11,
				SupportingPointCount: 4,
			},
		},
		Trajectory: flightfeatures.TrajectoryFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:               flightfeatures.AvailabilityStatusAvailable,
				AvailableFieldCount:  16,
				TotalFieldCount:      16,
				SupportingPointCount: 4,
			},
			PointCount:             4,
			TrajectoryQualityScore: 0.9,
		},
		Aircraft: flightfeatures.AircraftFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:              flightfeatures.AvailabilityStatusAvailable,
				AvailableFieldCount: 6,
				TotalFieldCount:     6,
			},
		},
		Quality: flightfeatures.FeatureQuality{
			Status:               flightfeatures.ValidationStatusValid,
			CompletenessScore:    1,
			InputQualityScore:    0.9,
			SupportingPointCount: 4,
		},
		Provenance: flightfeatures.FeatureProvenance{
			ExtractorVersion: "flight-feature-extractor-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat(
					fingerprintCharacter,
					64,
				),
			TrajectoryUpdatedAt: asOfTime,
			SourceNames:         []string{"open-sky"},
		},
	}
}
