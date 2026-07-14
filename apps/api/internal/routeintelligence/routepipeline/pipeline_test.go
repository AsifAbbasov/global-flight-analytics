package routepipeline

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/endpointevidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routestore"
)

func TestProcessBuildsAndStoresCompleteRoute(
	t *testing.T,
) {
	item, now := validPipelineTrajectory()
	reader := &fakeTrajectoryReader{
		item: item,
	}
	lister := &fakeAirportLister{
		items: validAirports(),
	}
	store := newFakeStore(now)

	pipeline := mustPipeline(
		t,
		Config{
			TrajectoryReader: reader,
			AirportLister:    lister,
			Store:            store,
			Now: func() time.Time {
				return now
			},
		},
	)

	result, err := pipeline.Process(
		context.Background(),
		Request{
			TrajectoryID: item.ID,
		},
	)
	if err != nil {
		t.Fatalf(
			"Process() error = %v",
			err,
		)
	}

	if result.PipelineVersion != Version ||
		result.TrajectoryID != item.ID ||
		result.Origin.Status !=
			endpointevidence.SelectionStatusSelected ||
		result.Destination.Status !=
			endpointevidence.SelectionStatusSelected ||
		result.Resolution.Result.Status !=
			routecontract.RouteStatusComplete ||
		result.Resolution.Result.Origin == nil ||
		result.Resolution.Result.Destination == nil ||
		result.Resolution.Result.Origin.
			Airport.ICAOCode != "UBBB" ||
		result.Resolution.Result.Destination.
			Airport.ICAOCode != "UGTB" {
		t.Fatalf(
			"unexpected pipeline result: %#v",
			result,
		)
	}
	if result.Resolution.Validation.Status !=
		routecontract.ValidationStatusValid ||
		result.Record.ID == "" ||
		!reflect.DeepEqual(
			result.Record.Result,
			result.Resolution.Result,
		) {
		t.Fatalf(
			"unexpected stored result: %#v",
			result,
		)
	}
	if result.CatalogReport.AcceptedCount != 3 ||
		result.CatalogReport.ExcludedCount != 0 ||
		lister.calls != 1 ||
		reader.calls != 1 ||
		store.puts != 1 {
		t.Fatalf(
			"unexpected execution counts: catalog=%#v reader=%d lister=%d puts=%d",
			result.CatalogReport,
			reader.calls,
			lister.calls,
			store.puts,
		)
	}

	wantSources := []string{
		"airplaneslive",
		"ourairports",
		"trajectory_endpoint",
	}
	if !reflect.DeepEqual(
		result.Resolution.Result.
			Provenance.SourceNames,
		wantSources,
	) {
		t.Fatalf(
			"sources = %#v, want %#v",
			result.Resolution.Result.
				Provenance.SourceNames,
			wantSources,
		)
	}
}

func TestProcessUsesCachedAirportCatalog(
	t *testing.T,
) {
	item, now := validPipelineTrajectory()
	lister := &fakeAirportLister{
		items: validAirports(),
	}
	store := newFakeStore(now)
	pipeline := mustPipeline(
		t,
		Config{
			TrajectoryReader: &fakeTrajectoryReader{
				item: item,
			},
			AirportLister: lister,
			Store:         store,
			Now: func() time.Time {
				return now
			},
		},
	)

	for index := 0; index < 2; index++ {
		_, err := pipeline.Process(
			context.Background(),
			Request{
				TrajectoryID: item.ID,
			},
		)
		if err != nil {
			t.Fatalf(
				"Process() iteration %d error = %v",
				index,
				err,
			)
		}
	}

	if lister.calls != 1 {
		t.Fatalf(
			"airport lister calls = %d, want 1",
			lister.calls,
		)
	}
	if store.puts != 2 ||
		len(store.records) != 1 {
		t.Fatalf(
			"store puts=%d records=%d",
			store.puts,
			len(store.records),
		)
	}
}

func TestProcessReloadsExpiredAirportCatalog(
	t *testing.T,
) {
	item, initialNow := validPipelineTrajectory()
	currentNow := initialNow
	lister := &fakeAirportLister{
		items: validAirports(),
	}
	pipeline := mustPipeline(
		t,
		Config{
			TrajectoryReader: &fakeTrajectoryReader{
				item: item,
			},
			AirportLister: lister,
			Store: newFakeStore(
				initialNow,
			),
			AirportCatalogTTL: time.Minute,
			Now: func() time.Time {
				return currentNow
			},
		},
	)

	_, err := pipeline.Process(
		context.Background(),
		Request{
			TrajectoryID: item.ID,
		},
	)
	if err != nil {
		t.Fatalf(
			"first Process() error = %v",
			err,
		)
	}

	currentNow = initialNow.Add(
		2 * time.Minute,
	)
	_, err = pipeline.Process(
		context.Background(),
		Request{
			TrajectoryID: item.ID,
		},
	)
	if err != nil {
		t.Fatalf(
			"second Process() error = %v",
			err,
		)
	}

	if lister.calls != 2 {
		t.Fatalf(
			"airport lister calls = %d, want 2",
			lister.calls,
		)
	}
}

func TestProcessWithoutUsableSegmentsSkipsAirportCatalog(
	t *testing.T,
) {
	item, now := validPipelineTrajectory()
	item.Segments = []trajectory.TrajectorySegment{
		{
			ID:             "invalid-segment",
			SequenceNumber: 1,
			Status:         trajectory.SegmentStatusInvalid,
			StartTime:      item.StartTime,
			EndTime:        item.EndTime,
			StartLatitude:  40,
			StartLongitude: 50,
			EndLatitude:    41,
			EndLongitude:   45,
			PointCount:     5,
		},
	}
	lister := &fakeAirportLister{
		err: errors.New(
			"airport database unavailable",
		),
	}
	pipeline := mustPipeline(
		t,
		Config{
			TrajectoryReader: &fakeTrajectoryReader{
				item: item,
			},
			AirportLister: lister,
			Store:         newFakeStore(now),
			Now: func() time.Time {
				return now
			},
		},
	)

	result, err := pipeline.Process(
		context.Background(),
		Request{
			TrajectoryID: item.ID,
		},
	)
	if err != nil {
		t.Fatalf(
			"Process() error = %v",
			err,
		)
	}

	if lister.calls != 0 ||
		result.Origin.Status !=
			endpointevidence.SelectionStatusUnavailable ||
		result.Destination.Status !=
			endpointevidence.SelectionStatusUnavailable ||
		result.Resolution.Result.Status !=
			routecontract.RouteStatusUnavailable ||
		result.CatalogReport.InputCount != 0 {
		t.Fatalf(
			"unexpected unavailable result: %#v calls=%d",
			result,
			lister.calls,
		)
	}
}

func TestProcessUsesSortedUsableSegments(
	t *testing.T,
) {
	item, now := validPipelineTrajectory()
	first := item.Segments[0]
	last := item.Segments[1]

	item.Segments = []trajectory.TrajectorySegment{
		last,
		{
			ID:             "ignored-invalid",
			SequenceNumber: 0,
			Status:         trajectory.SegmentStatusInvalid,
			StartTime: item.StartTime.Add(
				-time.Minute,
			),
			EndTime:        item.StartTime,
			StartLatitude:  0,
			StartLongitude: 0,
			EndLatitude:    0,
			EndLongitude:   0,
			PointCount:     9,
		},
		first,
	}

	pipeline := mustPipeline(
		t,
		Config{
			TrajectoryReader: &fakeTrajectoryReader{
				item: item,
			},
			AirportLister: &fakeAirportLister{
				items: validAirports(),
			},
			Store: newFakeStore(now),
			Now: func() time.Time {
				return now
			},
		},
	)

	result, err := pipeline.Process(
		context.Background(),
		Request{
			TrajectoryID: item.ID,
		},
	)
	if err != nil {
		t.Fatalf(
			"Process() error = %v",
			err,
		)
	}

	if result.Resolution.Result.Origin.
		Airport.ICAOCode != "UBBB" ||
		result.Resolution.Result.Destination.
			Airport.ICAOCode != "UGTB" {
		t.Fatalf(
			"unexpected sorted endpoint result: %#v",
			result.Resolution.Result,
		)
	}
}

func TestProcessWrapsStageFailures(
	t *testing.T,
) {
	item, now := validPipelineTrajectory()
	backendErr := errors.New(
		"backend failure",
	)

	tests := []struct {
		name      string
		config    Config
		wantStage Stage
	}{
		{
			name: "trajectory load",
			config: Config{
				TrajectoryReader: &fakeTrajectoryReader{
					err: backendErr,
				},
				AirportLister: &fakeAirportLister{
					items: validAirports(),
				},
				Store: newFakeStore(now),
				Now: func() time.Time {
					return now
				},
			},
			wantStage: StageTrajectoryLoad,
		},
		{
			name: "airport catalog",
			config: Config{
				TrajectoryReader: &fakeTrajectoryReader{
					item: item,
				},
				AirportLister: &fakeAirportLister{
					err: backendErr,
				},
				Store: newFakeStore(now),
				Now: func() time.Time {
					return now
				},
			},
			wantStage: StageAirportCatalog,
		},
		{
			name: "storage",
			config: Config{
				TrajectoryReader: &fakeTrajectoryReader{
					item: item,
				},
				AirportLister: &fakeAirportLister{
					items: validAirports(),
				},
				Store: &fakeStore{
					err: backendErr,
				},
				Now: func() time.Time {
					return now
				},
			},
			wantStage: StageStorage,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				pipeline := mustPipeline(
					t,
					test.config,
				)

				_, err := pipeline.Process(
					context.Background(),
					Request{
						TrajectoryID: item.ID,
					},
				)
				if !errors.Is(
					err,
					backendErr,
				) {
					t.Fatalf(
						"Process() error = %v, want wrapped backend error",
						err,
					)
				}

				var stageErr *StageError
				if !errors.As(
					err,
					&stageErr,
				) ||
					stageErr.Stage !=
						test.wantStage {
					t.Fatalf(
						"stage error = %#v, want %s",
						stageErr,
						test.wantStage,
					)
				}
			},
		)
	}
}

func TestProcessPreservesContextCancellation(
	t *testing.T,
) {
	item, now := validPipelineTrajectory()
	pipeline := mustPipeline(
		t,
		Config{
			TrajectoryReader: &fakeTrajectoryReader{
				item: item,
			},
			AirportLister: &fakeAirportLister{
				items: validAirports(),
			},
			Store: newFakeStore(now),
			Now: func() time.Time {
				return now
			},
		},
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := pipeline.Process(
		ctx,
		Request{
			TrajectoryID: item.ID,
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Process() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestProcessRejectsMismatchedTrajectoryIdentity(
	t *testing.T,
) {
	item, now := validPipelineTrajectory()
	readerItem := cloneTrajectory(item)
	readerItem.ID =
		"77fa5c4d-9251-42ec-aef3-1868c739b01e"

	pipeline := mustPipeline(
		t,
		Config{
			TrajectoryReader: &fakeTrajectoryReader{
				item: readerItem,
			},
			AirportLister: &fakeAirportLister{
				items: validAirports(),
			},
			Store: newFakeStore(now),
			Now: func() time.Time {
				return now
			},
		},
	)

	_, err := pipeline.Process(
		context.Background(),
		Request{
			TrajectoryID: item.ID,
		},
	)
	if !errors.Is(
		err,
		ErrTrajectoryIdentityMismatch,
	) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			ErrTrajectoryIdentityMismatch,
		)
	}
}

func TestProcessRequiresAnalyticalAsOfTime(
	t *testing.T,
) {
	item, now := validPipelineTrajectory()
	item.StartTime = time.Time{}
	item.EndTime = time.Time{}
	item.CreatedAt = time.Time{}
	item.UpdatedAt = time.Time{}
	item.Segments = nil

	pipeline := mustPipeline(
		t,
		Config{
			TrajectoryReader: &fakeTrajectoryReader{
				item: item,
			},
			AirportLister: &fakeAirportLister{
				items: validAirports(),
			},
			Store: newFakeStore(now),
			Now: func() time.Time {
				return now
			},
		},
	)

	_, err := pipeline.Process(
		context.Background(),
		Request{
			TrajectoryID: item.ID,
		},
	)
	if !errors.Is(
		err,
		ErrNoAnalyticalAsOfTime,
	) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			ErrNoAnalyticalAsOfTime,
		)
	}
}

func TestNewRejectsInvalidConfiguration(
	t *testing.T,
) {
	item, now := validPipelineTrajectory()
	reader := &fakeTrajectoryReader{
		item: item,
	}
	lister := &fakeAirportLister{
		items: validAirports(),
	}
	store := newFakeStore(now)

	tests := []struct {
		name   string
		config Config
		want   error
	}{
		{
			name: "trajectory reader",
			config: Config{
				AirportLister: lister,
				Store:         store,
			},
			want: ErrTrajectoryReaderRequired,
		},
		{
			name: "airport lister",
			config: Config{
				TrajectoryReader: reader,
				Store:            store,
			},
			want: ErrAirportListerRequired,
		},
		{
			name: "store",
			config: Config{
				TrajectoryReader: reader,
				AirportLister:    lister,
			},
			want: ErrStoreRequired,
		},
		{
			name: "catalog ttl",
			config: Config{
				TrajectoryReader:  reader,
				AirportLister:     lister,
				Store:             store,
				AirportCatalogTTL: -time.Second,
			},
			want: ErrInvalidAirportCatalogTTL,
		},
		{
			name: "endpoint policy",
			config: Config{
				TrajectoryReader: reader,
				AirportLister:    lister,
				Store:            store,
				EndpointEvidence: endpointevidence.Config{
					MinimumSelectionScore: 2,
				},
			},
			want: endpointevidence.
				ErrInvalidMinimumSelectionScore,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				_, err := New(test.config)
				if !errors.Is(
					err,
					test.want,
				) {
					t.Fatalf(
						"New() error = %v, want %v",
						err,
						test.want,
					)
				}
			},
		)
	}
}

func validPipelineTrajectory() (
	trajectory.FlightTrajectory,
	time.Time,
) {
	startTime := time.Date(
		2026,
		time.July,
		14,
		18,
		0,
		0,
		123456789,
		time.UTC,
	)
	endTime := startTime.Add(
		90 * time.Minute,
	)
	updatedAt := endTime.Add(
		time.Second,
	)
	now := updatedAt.Add(
		time.Second,
	)

	item := trajectory.FlightTrajectory{
		ID: "8a3d6e20-2c68-4b35-a512-7d91e6a90c31",
		IdentityKey: "flight-identity-" +
			strings.Repeat("c", 64),
		IdentityBasis: trajectory.
			FlightIdentityBasisSourceFlightID,
		SplitReason: trajectory.
			FlightSplitReasonInitialObservation,
		FlightID:         "7b41fb0b-3b20-4ed6-acdd-df9f9f905644",
		AircraftID:       "d5a760b3-2af6-4520-b07d-84d087bd026c",
		ICAO24:           "ABC123",
		Callsign:         "J2001",
		StartTime:        startTime,
		EndTime:          endTime,
		DurationSeconds:  5400,
		SegmentCount:     2,
		PointCount:       10,
		CoverageGapCount: 0,
		QualityScore:     0.90,
		SourceName:       "airplaneslive",
		CreatedAt:        updatedAt,
		UpdatedAt:        updatedAt,
		Segments: []trajectory.TrajectorySegment{
			{
				ID:             "segment-origin",
				TrajectoryID:   "8a3d6e20-2c68-4b35-a512-7d91e6a90c31",
				ICAO24:         "ABC123",
				Callsign:       "J2001",
				SequenceNumber: 1,
				Status: trajectory.
					SegmentStatusObserved,
				QualityScore: 0.95,
				StartTime:    startTime,
				EndTime: startTime.Add(
					40 *
						time.Minute,
				),
				DurationSeconds: 2400,
				StartLatitude:   40.4675,
				StartLongitude:  50.0467,
				EndLatitude:     40.8,
				EndLongitude:    48.0,
				PointCount:      5,
				SourceName:      "airplaneslive",
				CreatedAt:       updatedAt,
			},
			{
				ID:             "segment-destination",
				TrajectoryID:   "8a3d6e20-2c68-4b35-a512-7d91e6a90c31",
				ICAO24:         "ABC123",
				Callsign:       "J2001",
				SequenceNumber: 2,
				Status: trajectory.
					SegmentStatusObserved,
				QualityScore: 0.90,
				StartTime: startTime.Add(
					50 *
						time.Minute,
				),
				EndTime:         endTime,
				DurationSeconds: 2400,
				StartLatitude:   41.0,
				StartLongitude:  46.0,
				EndLatitude:     41.6692,
				EndLongitude:    44.9547,
				PointCount:      5,
				SourceName:      "airplaneslive",
				CreatedAt:       updatedAt,
			},
		},
	}

	return item, now
}

func validAirports() []airport.Airport {
	return []airport.Airport{
		{
			ICAOCode:   "UBBB",
			IATACode:   "GYD",
			Name:       "Heydar Aliyev International Airport",
			City:       "Baku",
			Country:    "Azerbaijan",
			Latitude:   40.4675,
			Longitude:  50.0467,
			ElevationM: 3,
			Timezone:   "Asia/Baku",
		},
		{
			ICAOCode:   "UGTB",
			IATACode:   "TBS",
			Name:       "Tbilisi International Airport",
			City:       "Tbilisi",
			Country:    "Georgia",
			Latitude:   41.6692,
			Longitude:  44.9547,
			ElevationM: 495,
			Timezone:   "Asia/Tbilisi",
		},
		{
			ICAOCode:   "LTFM",
			IATACode:   "IST",
			Name:       "Istanbul Airport",
			City:       "Istanbul",
			Country:    "Turkey",
			Latitude:   41.2753,
			Longitude:  28.7519,
			ElevationM: 99,
			Timezone:   "Europe/Istanbul",
		},
	}
}

type fakeTrajectoryReader struct {
	item  trajectory.FlightTrajectory
	err   error
	calls int
}

func (reader *fakeTrajectoryReader) GetTrajectoryByID(
	ctx context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
	reader.calls++

	if err := ctx.Err(); err != nil {
		return trajectory.FlightTrajectory{},
			err
	}
	if reader.err != nil {
		return trajectory.FlightTrajectory{},
			reader.err
	}

	return cloneTrajectory(reader.item), nil
}

type fakeAirportLister struct {
	items []airport.Airport
	err   error
	calls int
}

func (lister *fakeAirportLister) List(
	ctx context.Context,
) ([]airport.Airport, error) {
	lister.calls++

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if lister.err != nil {
		return nil, lister.err
	}

	return append(
		[]airport.Airport(nil),
		lister.items...,
	), nil
}

type fakeStore struct {
	now     time.Time
	err     error
	puts    int
	records map[string]routestore.Record
}

func newFakeStore(
	now time.Time,
) *fakeStore {
	return &fakeStore{
		now: now,
		records: make(
			map[string]routestore.Record,
		),
	}
}

func (store *fakeStore) Put(
	ctx context.Context,
	result routecontract.Result,
) (routestore.Record, error) {
	store.puts++

	if err := ctx.Err(); err != nil {
		return routestore.Record{}, err
	}
	if store.err != nil {
		return routestore.Record{},
			store.err
	}
	if store.records == nil {
		store.records = make(
			map[string]routestore.Record,
		)
	}

	key := routestore.ResultKey{
		TrajectoryID:  result.TrajectoryID,
		SchemaVersion: result.SchemaVersion,
		AsOfTime:      result.Window.AsOfTime.UTC(),
	}
	encodedKey := fmt.Sprintf(
		"%s|%s|%d",
		key.TrajectoryID,
		key.SchemaVersion,
		key.AsOfTime.UnixNano(),
	)

	if existing, exists :=
		store.records[encodedKey]; exists {
		if existing.InputFingerprint !=
			result.Provenance.
				InputFingerprint {
			return routestore.Record{},
				routestore.
					ErrResultConflict
		}

		return existing.Clone(), nil
	}

	record := routestore.Record{
		ID: "route-record-" +
			strings.Repeat("f", 64),
		Key: key,
		InputFingerprint: result.Provenance.
			InputFingerprint,
		Result:   result.Clone(),
		StoredAt: store.now.UTC(),
	}
	store.records[encodedKey] =
		record.Clone()

	return record.Clone(), nil
}

func (store *fakeStore) Get(
	ctx context.Context,
	key routestore.ResultKey,
) (routestore.Record, error) {
	if err := ctx.Err(); err != nil {
		return routestore.Record{}, err
	}

	encodedKey := fmt.Sprintf(
		"%s|%s|%d",
		key.TrajectoryID,
		key.SchemaVersion,
		key.AsOfTime.UTC().UnixNano(),
	)
	record, exists := store.records[encodedKey]
	if !exists {
		return routestore.Record{},
			routestore.ErrResultNotFound
	}

	return record.Clone(), nil
}

func (store *fakeStore) GetLatest(
	ctx context.Context,
	trajectoryID string,
	schemaVersion routecontract.SchemaVersion,
) (routestore.Record, error) {
	if err := ctx.Err(); err != nil {
		return routestore.Record{}, err
	}

	var selected routestore.Record
	for _, record := range store.records {
		if record.Key.TrajectoryID !=
			trajectoryID ||
			record.Key.SchemaVersion !=
				schemaVersion {
			continue
		}
		if selected.ID == "" ||
			record.Key.AsOfTime.After(
				selected.Key.AsOfTime,
			) {
			selected = record.Clone()
		}
	}
	if selected.ID == "" {
		return routestore.Record{},
			routestore.ErrResultNotFound
	}

	return selected.Clone(), nil
}

func (store *fakeStore) List(
	ctx context.Context,
	query routestore.ListQuery,
) (routestore.Page, error) {
	if err := ctx.Err(); err != nil {
		return routestore.Page{}, err
	}

	records := make(
		[]routestore.Record,
		0,
		len(store.records),
	)
	for _, record := range store.records {
		if record.Key.TrajectoryID ==
			query.TrajectoryID &&
			record.Key.SchemaVersion ==
				query.SchemaVersion {
			records = append(
				records,
				record.Clone(),
			)
		}
	}

	return routestore.Page{
		Records: records,
	}.Clone(), nil
}

func mustPipeline(
	t *testing.T,
	config Config,
) *Pipeline {
	t.Helper()

	pipeline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	return pipeline
}
