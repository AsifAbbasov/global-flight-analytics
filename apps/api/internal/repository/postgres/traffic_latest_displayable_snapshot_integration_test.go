package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTrafficRepositoryUsesNewestDisplayableSuccessfulRun(
	t *testing.T,
) {
	fixture := newTrafficAltitudeFixture(t)
	ctx := context.Background()
	baseTime := trafficSnapshotBaseTime()

	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		"20000000-0000-0000-0000-000000000001",
		"success",
		baseTime,
		timePointer(baseTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		"20000000-0000-0000-0000-000000000001",
		"OLD001",
		float64Pointer(40),
		float64Pointer(49),
		float64Pointer(200),
		float64Pointer(90),
		boolPointer(false),
		nil,
		nil,
		baseTime,
	)

	newerTime := baseTime.Add(10 * time.Minute)
	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		"20000000-0000-0000-0000-000000000002",
		"success",
		newerTime,
		timePointer(newerTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		"20000000-0000-0000-0000-000000000002",
		"NEW001",
		float64Pointer(41),
		float64Pointer(50),
		float64Pointer(210),
		float64Pointer(100),
		boolPointer(false),
		nil,
		nil,
		newerTime,
	)

	items, err := fixture.repository.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get current traffic: %v", err)
	}

	assertSingleTrafficSnapshotItem(t, items, "NEW001")
}

func TestTrafficRepositoryFallsBackFromNonDisplayableSuccessfulRun(
	t *testing.T,
) {
	fixture := newTrafficAltitudeFixture(t)
	ctx := context.Background()
	baseTime := trafficSnapshotBaseTime()
	oldRunID := "21000000-0000-0000-0000-000000000001"
	newRunID := "21000000-0000-0000-0000-000000000002"

	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		oldRunID,
		"success",
		baseTime,
		timePointer(baseTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		oldRunID,
		"OLD002",
		float64Pointer(40),
		float64Pointer(49),
		float64Pointer(200),
		float64Pointer(90),
		boolPointer(false),
		nil,
		nil,
		baseTime,
	)

	newerTime := baseTime.Add(10 * time.Minute)
	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		newRunID,
		"success",
		newerTime,
		timePointer(newerTime),
	)

	insertTrafficSnapshotState(
		t,
		fixture.pool,
		newRunID,
		"NOLAT",
		nil,
		float64Pointer(49),
		float64Pointer(200),
		float64Pointer(90),
		boolPointer(false),
		nil,
		nil,
		newerTime,
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		newRunID,
		"NOLON",
		float64Pointer(40),
		nil,
		float64Pointer(200),
		float64Pointer(90),
		boolPointer(false),
		nil,
		nil,
		newerTime.Add(time.Second),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		newRunID,
		"NOSPEED",
		float64Pointer(40),
		float64Pointer(49),
		nil,
		float64Pointer(90),
		boolPointer(false),
		nil,
		nil,
		newerTime.Add(2*time.Second),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		newRunID,
		"NOHEAD",
		float64Pointer(40),
		float64Pointer(49),
		float64Pointer(200),
		nil,
		boolPointer(false),
		nil,
		nil,
		newerTime.Add(3*time.Second),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		newRunID,
		"NOGROUND",
		float64Pointer(40),
		float64Pointer(49),
		float64Pointer(200),
		float64Pointer(90),
		nil,
		nil,
		nil,
		newerTime.Add(4*time.Second),
	)

	items, err := fixture.repository.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get current traffic: %v", err)
	}

	assertSingleTrafficSnapshotItem(t, items, "OLD002")
}

func TestTrafficRepositoryIgnoresFailedUnfinishedAndEmptyRuns(
	t *testing.T,
) {
	fixture := newTrafficAltitudeFixture(t)
	ctx := context.Background()
	baseTime := trafficSnapshotBaseTime()
	oldRunID := "22000000-0000-0000-0000-000000000001"

	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		oldRunID,
		"success",
		baseTime,
		timePointer(baseTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		oldRunID,
		"STABLE1",
		float64Pointer(40),
		float64Pointer(49),
		float64Pointer(200),
		float64Pointer(90),
		boolPointer(false),
		nil,
		nil,
		baseTime,
	)

	emptyTime := baseTime.Add(10 * time.Minute)
	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		"22000000-0000-0000-0000-000000000002",
		"success",
		emptyTime,
		timePointer(emptyTime),
	)

	failedTime := baseTime.Add(20 * time.Minute)
	failedRunID := "22000000-0000-0000-0000-000000000003"
	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		failedRunID,
		"failed",
		failedTime,
		timePointer(failedTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		failedRunID,
		"FAILED1",
		float64Pointer(41),
		float64Pointer(50),
		float64Pointer(210),
		float64Pointer(100),
		boolPointer(false),
		nil,
		nil,
		failedTime,
	)

	unfinishedTime := baseTime.Add(30 * time.Minute)
	unfinishedRunID := "22000000-0000-0000-0000-000000000004"
	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		unfinishedRunID,
		"success",
		unfinishedTime,
		nil,
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		unfinishedRunID,
		"OPEN001",
		float64Pointer(42),
		float64Pointer(51),
		float64Pointer(220),
		float64Pointer(110),
		boolPointer(false),
		nil,
		nil,
		unfinishedTime,
	)

	items, err := fixture.repository.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get current traffic: %v", err)
	}

	assertSingleTrafficSnapshotItem(t, items, "STABLE1")
}

func TestTrafficRepositoryFiltersNonDisplayableRowsWithinSelectedRun(
	t *testing.T,
) {
	fixture := newTrafficAltitudeFixture(t)
	ctx := context.Background()
	baseTime := trafficSnapshotBaseTime()
	oldRunID := "23000000-0000-0000-0000-000000000001"
	newRunID := "23000000-0000-0000-0000-000000000002"

	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		oldRunID,
		"success",
		baseTime,
		timePointer(baseTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		oldRunID,
		"OLD003",
		float64Pointer(40),
		float64Pointer(49),
		float64Pointer(200),
		float64Pointer(90),
		boolPointer(false),
		nil,
		nil,
		baseTime,
	)

	newerTime := baseTime.Add(10 * time.Minute)
	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		newRunID,
		"success",
		newerTime,
		timePointer(newerTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		newRunID,
		"MIXED1",
		float64Pointer(41),
		float64Pointer(50),
		float64Pointer(210),
		float64Pointer(100),
		boolPointer(false),
		nil,
		nil,
		newerTime,
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		newRunID,
		"MIXED2",
		float64Pointer(42),
		float64Pointer(51),
		nil,
		float64Pointer(110),
		boolPointer(false),
		nil,
		nil,
		newerTime.Add(time.Second),
	)

	items, err := fixture.repository.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get current traffic: %v", err)
	}

	assertSingleTrafficSnapshotItem(t, items, "MIXED1")
}

func TestTrafficRepositoryBoundsUseGloballySelectedSnapshot(
	t *testing.T,
) {
	fixture := newTrafficAltitudeFixture(t)
	ctx := context.Background()
	baseTime := trafficSnapshotBaseTime()
	oldRunID := "24000000-0000-0000-0000-000000000001"
	newRunID := "24000000-0000-0000-0000-000000000002"

	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		oldRunID,
		"success",
		baseTime,
		timePointer(baseTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		oldRunID,
		"INBOUND",
		float64Pointer(40),
		float64Pointer(49),
		float64Pointer(200),
		float64Pointer(90),
		boolPointer(false),
		nil,
		nil,
		baseTime,
	)

	newerTime := baseTime.Add(10 * time.Minute)
	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		newRunID,
		"success",
		newerTime,
		timePointer(newerTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		newRunID,
		"OUTSIDE",
		float64Pointer(10),
		float64Pointer(10),
		float64Pointer(210),
		float64Pointer(100),
		boolPointer(false),
		nil,
		nil,
		newerTime,
	)

	globalItems, err := fixture.repository.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get current traffic: %v", err)
	}
	assertSingleTrafficSnapshotItem(t, globalItems, "OUTSIDE")

	boundedItems, err := fixture.repository.GetCurrentByBounds(
		ctx,
		traffic.Bounds{
			MinLatitude:  39,
			MaxLatitude:  41,
			MinLongitude: 48,
			MaxLongitude: 50,
		},
	)
	if err != nil {
		t.Fatalf("get bounded current traffic: %v", err)
	}
	if len(boundedItems) != 0 {
		t.Fatalf(
			"bounded traffic count = %d, want 0 without stale regional fallback",
			len(boundedItems),
		)
	}
}

func TestTrafficRepositoryTreatsMissingAltitudeAsDisplayable(
	t *testing.T,
) {
	fixture := newTrafficAltitudeFixture(t)
	ctx := context.Background()
	baseTime := trafficSnapshotBaseTime()
	runID := "25000000-0000-0000-0000-000000000001"

	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		runID,
		"success",
		baseTime,
		timePointer(baseTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		runID,
		"NOALT1",
		float64Pointer(40),
		float64Pointer(49),
		float64Pointer(200),
		float64Pointer(90),
		boolPointer(false),
		nil,
		nil,
		baseTime,
	)

	items, err := fixture.repository.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get current traffic: %v", err)
	}

	assertSingleTrafficSnapshotItem(t, items, "NOALT1")
	if items[0].AltitudeM != nil {
		t.Fatalf(
			"altitude = %v, want nil optional altitude",
			*items[0].AltitudeM,
		)
	}
	if items[0].AltitudeSource != traffic.AltitudeSourceNone {
		t.Fatalf(
			"altitude source = %q, want %q",
			items[0].AltitudeSource,
			traffic.AltitudeSourceNone,
		)
	}
}

func TestTrafficRepositoryBoundsFallBackFromNonDisplayableSuccessfulRun(
	t *testing.T,
) {
	fixture := newTrafficAltitudeFixture(t)
	ctx := context.Background()
	baseTime := trafficSnapshotBaseTime()
	oldRunID := "26000000-0000-0000-0000-000000000001"
	newRunID := "26000000-0000-0000-0000-000000000002"

	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		oldRunID,
		"success",
		baseTime,
		timePointer(baseTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		oldRunID,
		"BOUNDS1",
		float64Pointer(40),
		float64Pointer(49),
		float64Pointer(200),
		float64Pointer(90),
		boolPointer(false),
		nil,
		nil,
		baseTime,
	)

	newerTime := baseTime.Add(10 * time.Minute)
	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		newRunID,
		"success",
		newerTime,
		timePointer(newerTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		newRunID,
		"INVALID1",
		float64Pointer(40),
		float64Pointer(49),
		nil,
		float64Pointer(90),
		boolPointer(false),
		nil,
		nil,
		newerTime,
	)

	items, err := fixture.repository.GetCurrentByBounds(
		ctx,
		traffic.Bounds{
			MinLatitude:  39,
			MaxLatitude:  41,
			MinLongitude: 48,
			MaxLongitude: 50,
		},
	)
	if err != nil {
		t.Fatalf("get bounded current traffic: %v", err)
	}

	assertSingleTrafficSnapshotItem(t, items, "BOUNDS1")
}

func TestTrafficRepositoryPreservesZeroTelemetryValues(
	t *testing.T,
) {
	fixture := newTrafficAltitudeFixture(t)
	ctx := context.Background()
	baseTime := trafficSnapshotBaseTime()
	runID := "27000000-0000-0000-0000-000000000001"

	insertTrafficSnapshotRun(
		t,
		fixture.pool,
		runID,
		"success",
		baseTime,
		timePointer(baseTime),
	)
	insertTrafficSnapshotState(
		t,
		fixture.pool,
		runID,
		"ZERO001",
		float64Pointer(40),
		float64Pointer(49),
		float64Pointer(0),
		float64Pointer(0),
		boolPointer(false),
		nil,
		nil,
		baseTime,
	)

	items, err := fixture.repository.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get current traffic: %v", err)
	}

	assertSingleTrafficSnapshotItem(t, items, "ZERO001")
	if items[0].VelocityMPS != 0 {
		t.Fatalf(
			"velocity = %v, want 0 preserved as observed telemetry",
			items[0].VelocityMPS,
		)
	}
	if items[0].HeadingDegrees != 0 {
		t.Fatalf(
			"heading = %v, want 0 preserved as observed telemetry",
			items[0].HeadingDegrees,
		)
	}
	if items[0].OnGround {
		t.Fatal("on_ground = true, want false preserved as observed telemetry")
	}
}

func insertTrafficSnapshotRun(
	t *testing.T,
	pool *pgxpool.Pool,
	runID string,
	status string,
	createdAt time.Time,
	finishedAt *time.Time,
) {
	t.Helper()

	mustExecTrafficAltitudeSQL(
		t,
		pool,
		`
			INSERT INTO ingestion_runs (
				id,
				finished_at,
				status,
				created_at
			)
			VALUES ($1, $2, $3, $4);
		`,
		runID,
		finishedAt,
		status,
		createdAt,
	)
}

func insertTrafficSnapshotState(
	t *testing.T,
	pool *pgxpool.Pool,
	runID string,
	icao24 string,
	latitude *float64,
	longitude *float64,
	velocity *float64,
	heading *float64,
	onGround *bool,
	geometricAltitude *int,
	barometricAltitude *int,
	observedAt time.Time,
) {
	t.Helper()

	geometricStatus := "unknown"
	if geometricAltitude != nil {
		geometricStatus = "observed"
	}
	barometricStatus := "unavailable"
	if barometricAltitude != nil {
		barometricStatus = "observed"
	}

	mustExecTrafficAltitudeSQL(
		t,
		pool,
		`
			INSERT INTO flight_states (
				ingestion_run_id,
				icao24,
				callsign,
				latitude,
				longitude,
				geometric_altitude_m,
				geometric_altitude_status,
				barometric_altitude_m,
				barometric_altitude_status,
				velocity_mps,
				heading_degrees,
				on_ground,
				observed_at,
				origin_country
			)
			VALUES (
				$1,
				$2,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9,
				$10,
				$11,
				$12,
				'Azerbaijan'
			)
		`,
		runID,
		icao24,
		latitude,
		longitude,
		geometricAltitude,
		geometricStatus,
		barometricAltitude,
		barometricStatus,
		velocity,
		heading,
		onGround,
		observedAt,
	)
}

func assertSingleTrafficSnapshotItem(
	t *testing.T,
	items []traffic.CurrentTrafficItem,
	expectedICAO24 string,
) {
	t.Helper()

	if len(items) != 1 {
		t.Fatalf(
			"current traffic count = %d, want 1",
			len(items),
		)
	}
	if items[0].ICAO24 != expectedICAO24 {
		t.Fatalf(
			"icao24 = %q, want %q",
			items[0].ICAO24,
			expectedICAO24,
		)
	}
}

func trafficSnapshotBaseTime() time.Time {
	return time.Date(
		2026,
		time.August,
		7,
		0,
		0,
		0,
		0,
		time.UTC,
	)
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func float64Pointer(value float64) *float64 {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}
