package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const flightStateAltitudeTestDatabaseURLEnvironmentVariable = "TEST_DATABASE_URL"

var flightStateAltitudeSchemaCounter uint64

type flightStateAltitudeIntegrationFixture struct {
	pool       *pgxpool.Pool
	repository *FlightStateRepository
}

func TestFlightStateAltitudeMigrationBackfillsLegacySemantics(
	t *testing.T,
) {
	fixture := newFlightStateAltitudeIntegrationFixture(
		t,
		true,
	)

	rows, err := fixture.pool.Query(
		context.Background(),
		`
			SELECT
				icao24,
				barometric_altitude_m,
				barometric_altitude_status,
				geometric_altitude_m,
				geometric_altitude_status
			FROM flight_states
			ORDER BY icao24 ASC
		`,
	)
	if err != nil {
		t.Fatalf(
			"query migrated legacy flight states: %v",
			err,
		)
	}
	defer rows.Close()

	type expectedRow struct {
		icao24               string
		barometricValueValid bool
		barometricValue      int32
		barometricStatus     string
		geometricValueValid  bool
		geometricValue       int32
		geometricStatus      string
	}

	expected := []expectedRow{
		{
			icao24:           "LEG001",
			barometricStatus: "unavailable",
			geometricStatus:  "unavailable",
		},
		{
			icao24:               "LEG002",
			barometricValueValid: true,
			barometricValue:      0,
			barometricStatus:     "ground",
			geometricStatus:      "unknown",
		},
		{
			icao24:           "LEG003",
			barometricStatus: "unknown",
			geometricStatus:  "unknown",
		},
		{
			icao24:               "LEG004",
			barometricValueValid: true,
			barometricValue:      1000,
			barometricStatus:     "observed",
			geometricValueValid:  true,
			geometricValue:       1100,
			geometricStatus:      "observed",
		},
	}

	index := 0

	for rows.Next() {
		if index >= len(expected) {
			t.Fatal(
				"received more migrated rows than expected",
			)
		}

		var icao24 string
		var barometricValue pgtype.Int4
		var barometricStatus string
		var geometricValue pgtype.Int4
		var geometricStatus string

		if err := rows.Scan(
			&icao24,
			&barometricValue,
			&barometricStatus,
			&geometricValue,
			&geometricStatus,
		); err != nil {
			t.Fatalf(
				"scan migrated legacy flight state: %v",
				err,
			)
		}

		want := expected[index]

		if icao24 != want.icao24 {
			t.Fatalf(
				"expected icao24 %s, got %s",
				want.icao24,
				icao24,
			)
		}

		assertInt4Value(
			t,
			"barometric altitude",
			barometricValue,
			want.barometricValueValid,
			want.barometricValue,
		)

		if barometricStatus != want.barometricStatus {
			t.Fatalf(
				"expected barometric status %s, got %s",
				want.barometricStatus,
				barometricStatus,
			)
		}

		assertInt4Value(
			t,
			"geometric altitude",
			geometricValue,
			want.geometricValueValid,
			want.geometricValue,
		)

		if geometricStatus != want.geometricStatus {
			t.Fatalf(
				"expected geometric status %s, got %s",
				want.geometricStatus,
				geometricStatus,
			)
		}

		index++
	}

	if err := rows.Err(); err != nil {
		t.Fatalf(
			"iterate migrated legacy flight states: %v",
			err,
		)
	}

	if index != len(expected) {
		t.Fatalf(
			"expected %d migrated rows, got %d",
			len(expected),
			index,
		)
	}
}

func TestFlightStateRepositoryRoundTripsAltitudeSemantics(
	t *testing.T,
) {
	fixture := newFlightStateAltitudeIntegrationFixture(
		t,
		false,
	)

	flightID := "11111111-1111-1111-1111-111111111111"
	baseTime := time.Date(
		2026,
		time.July,
		9,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	states := []flightstate.FlightState{
		makeAltitudePersistenceState(
			flightID,
			"ABC123",
			baseTime,
			0,
			flightstate.AltitudeStatusObserved,
			0,
			flightstate.AltitudeStatusObserved,
			false,
		),
		makeAltitudePersistenceState(
			flightID,
			"ABC123",
			baseTime.Add(time.Second),
			0,
			flightstate.AltitudeStatusGround,
			0,
			flightstate.AltitudeStatusUnavailable,
			true,
		),
		makeAltitudePersistenceState(
			flightID,
			"ABC123",
			baseTime.Add(2*time.Second),
			0,
			flightstate.AltitudeStatusUnknown,
			0,
			flightstate.AltitudeStatusUnknown,
			false,
		),
		makeAltitudePersistenceState(
			flightID,
			"ABC123",
			baseTime.Add(3*time.Second),
			0,
			flightstate.AltitudeStatusUnavailable,
			0,
			flightstate.AltitudeStatusUnavailable,
			false,
		),
		makeAltitudePersistenceState(
			flightID,
			"ABC123",
			baseTime.Add(4*time.Second),
			0,
			flightstate.AltitudeStatusInvalid,
			0,
			flightstate.AltitudeStatusInvalid,
			false,
		),
	}

	if err := fixture.repository.SaveFlightStates(
		context.Background(),
		states,
	); err != nil {
		t.Fatalf(
			"save altitude semantic flight states: %v",
			err,
		)
	}

	loaded, err := fixture.repository.ListByFlightID(
		context.Background(),
		flightID,
	)
	if err != nil {
		t.Fatalf(
			"list altitude semantic flight states: %v",
			err,
		)
	}

	if len(loaded) != len(states) {
		t.Fatalf(
			"expected %d loaded states, got %d",
			len(states),
			len(loaded),
		)
	}

	for index := range states {
		if loaded[index].BarometricAltitudeStatus !=
			states[index].BarometricAltitudeStatus {
			t.Fatalf(
				"state %d expected barometric status %q, got %q",
				index,
				states[index].BarometricAltitudeStatus,
				loaded[index].BarometricAltitudeStatus,
			)
		}

		if loaded[index].GeometricAltitudeStatus !=
			states[index].GeometricAltitudeStatus {
			t.Fatalf(
				"state %d expected geometric status %q, got %q",
				index,
				states[index].GeometricAltitudeStatus,
				loaded[index].GeometricAltitudeStatus,
			)
		}
	}

	latest, err := fixture.repository.GetLatestByICAO24(
		context.Background(),
		"ABC123",
	)
	if err != nil {
		t.Fatalf(
			"get latest altitude semantic flight state: %v",
			err,
		)
	}

	if latest.BarometricAltitudeStatus != flightstate.AltitudeStatusInvalid {
		t.Fatalf(
			"expected latest barometric status %q, got %q",
			flightstate.AltitudeStatusInvalid,
			latest.BarometricAltitudeStatus,
		)
	}

	if latest.GeometricAltitudeStatus != flightstate.AltitudeStatusInvalid {
		t.Fatalf(
			"expected latest geometric status %q, got %q",
			flightstate.AltitudeStatusInvalid,
			latest.GeometricAltitudeStatus,
		)
	}

	assertPersistedAltitudeNullability(
		t,
		fixture.pool,
	)
}

func TestFlightStateRepositoryRejectsUnsupportedAltitudeStatus(
	t *testing.T,
) {
	fixture := newFlightStateAltitudeIntegrationFixture(
		t,
		false,
	)

	state := makeAltitudePersistenceState(
		"22222222-2222-2222-2222-222222222222",
		"DEF456",
		time.Date(
			2026,
			time.July,
			9,
			13,
			0,
			0,
			0,
			time.UTC,
		),
		1000,
		flightstate.AltitudeStatus("unsupported"),
		1100,
		flightstate.AltitudeStatusObserved,
		false,
	)

	err := fixture.repository.SaveFlightStates(
		context.Background(),
		[]flightstate.FlightState{
			state,
		},
	)
	if err == nil {
		t.Fatal(
			"expected unsupported altitude status error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"unsupported altitude status",
	) {
		t.Fatalf(
			"expected unsupported altitude status error, got %v",
			err,
		)
	}
}

func newFlightStateAltitudeIntegrationFixture(
	t *testing.T,
	seedLegacyRows bool,
) *flightStateAltitudeIntegrationFixture {
	t.Helper()

	databaseURL := strings.TrimSpace(
		os.Getenv(
			flightStateAltitudeTestDatabaseURLEnvironmentVariable,
		),
	)
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			flightStateAltitudeTestDatabaseURLEnvironmentVariable,
		)
	}

	setupContext, cancelSetup := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancelSetup()

	bootstrapConnection, err := pgx.Connect(
		setupContext,
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"connect to PostgreSQL test database: %v",
			err,
		)
	}

	schemaName := fmt.Sprintf(
		"flight_state_altitude_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(
			&flightStateAltitudeSchemaCounter,
			1,
		),
	)

	quotedSchemaName := pgx.Identifier{
		schemaName,
	}.Sanitize()

	if _, err := bootstrapConnection.Exec(
		setupContext,
		"CREATE SCHEMA "+quotedSchemaName,
	); err != nil {
		_ = bootstrapConnection.Close(
			setupContext,
		)

		t.Fatalf(
			"create PostgreSQL test schema: %v",
			err,
		)
	}

	poolConfig, err := pgxpool.ParseConfig(
		databaseURL,
	)
	if err != nil {
		dropFlightStateAltitudeTestSchema(
			t,
			bootstrapConnection,
			quotedSchemaName,
		)

		_ = bootstrapConnection.Close(
			setupContext,
		)

		t.Fatalf(
			"parse PostgreSQL test pool config: %v",
			err,
		)
	}

	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(
			map[string]string,
		)
	}

	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName

	pool, err := pgxpool.NewWithConfig(
		setupContext,
		poolConfig,
	)
	if err != nil {
		dropFlightStateAltitudeTestSchema(
			t,
			bootstrapConnection,
			quotedSchemaName,
		)

		_ = bootstrapConnection.Close(
			setupContext,
		)

		t.Fatalf(
			"create PostgreSQL test pool: %v",
			err,
		)
	}

	if err := pool.Ping(
		setupContext,
	); err != nil {
		pool.Close()

		dropFlightStateAltitudeTestSchema(
			t,
			bootstrapConnection,
			quotedSchemaName,
		)

		_ = bootstrapConnection.Close(
			setupContext,
		)

		t.Fatalf(
			"ping PostgreSQL test pool: %v",
			err,
		)
	}

	createLegacyFlightStatesTable(
		t,
		setupContext,
		pool,
	)

	if seedLegacyRows {
		seedLegacyAltitudeRows(
			t,
			setupContext,
			pool,
		)
	}

	applyFlightStateAltitudeMigration(
		t,
		setupContext,
		pool,
	)

	t.Cleanup(
		func() {
			pool.Close()

			cleanupContext, cancelCleanup := context.WithTimeout(
				context.Background(),
				30*time.Second,
			)
			defer cancelCleanup()

			if _, err := bootstrapConnection.Exec(
				cleanupContext,
				"DROP SCHEMA IF EXISTS "+
					quotedSchemaName+
					" CASCADE",
			); err != nil {
				t.Errorf(
					"drop PostgreSQL test schema: %v",
					err,
				)
			}

			if err := bootstrapConnection.Close(
				cleanupContext,
			); err != nil {
				t.Errorf(
					"close PostgreSQL bootstrap connection: %v",
					err,
				)
			}
		},
	)

	return &flightStateAltitudeIntegrationFixture{
		pool: pool,
		repository: NewFlightStateRepository(
			pool,
		),
	}
}

func createLegacyFlightStatesTable(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
) {
	t.Helper()

	_, err := pool.Exec(
		ctx,
		`
			CREATE TABLE flight_states (
				id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
				flight_id uuid,
				aircraft_id uuid,
				icao24 varchar(10) NOT NULL,
				callsign text,
				latitude numeric,
				longitude numeric,
				barometric_altitude_m integer,
				geometric_altitude_m integer,
				velocity_mps numeric,
				heading_degrees numeric,
				vertical_rate_mps numeric,
				on_ground boolean,
				origin_country text,
				squawk_code text,
				special_purpose_indicator boolean,
				position_source text,
				aircraft_category smallint,
				aircraft_category_available boolean,
				observed_at timestamptz NOT NULL,
				source_name text NOT NULL,
				ingestion_run_id uuid,
				created_at timestamptz NOT NULL DEFAULT now()
			)
		`,
	)
	if err != nil {
		t.Fatalf(
			"create legacy flight_states table: %v",
			err,
		)
	}
}

func seedLegacyAltitudeRows(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
) {
	t.Helper()

	_, err := pool.Exec(
		ctx,
		`
			INSERT INTO flight_states (
				icao24,
				barometric_altitude_m,
				geometric_altitude_m,
				on_ground,
				observed_at,
				source_name
			)
			VALUES
				(
					'LEG001',
					NULL,
					NULL,
					false,
					'2026-07-09T10:00:00Z',
					'legacy'
				),
				(
					'LEG002',
					0,
					0,
					true,
					'2026-07-09T10:00:01Z',
					'legacy'
				),
				(
					'LEG003',
					0,
					0,
					false,
					'2026-07-09T10:00:02Z',
					'legacy'
				),
				(
					'LEG004',
					1000,
					1100,
					false,
					'2026-07-09T10:00:03Z',
					'legacy'
				)
		`,
	)
	if err != nil {
		t.Fatalf(
			"seed legacy altitude rows: %v",
			err,
		)
	}
}

func applyFlightStateAltitudeMigration(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(
		0,
	)
	if !ok {
		t.Fatal(
			"resolve integration test file path",
		)
	}

	migrationFilenames := []string{
		"006_flight_state_altitude_semantics.sql",
		"023_ingestion_durability_replay_partial.sql",
	}

	for _, migrationFilename := range migrationFilenames {
		migrationPath := filepath.Clean(
			filepath.Join(
				filepath.Dir(currentFile),
				"../../../../../database/migrations",
				migrationFilename,
			),
		)
		sqlBytes, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Fatalf("read flight state migration %s: %v", migrationPath, err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			t.Fatalf("apply flight state migration %s: %v", migrationFilename, err)
		}
	}
}

func dropFlightStateAltitudeTestSchema(
	t *testing.T,
	connection *pgx.Conn,
	quotedSchemaName string,
) {
	t.Helper()

	cleanupContext, cancelCleanup := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancelCleanup()

	if _, err := connection.Exec(
		cleanupContext,
		"DROP SCHEMA IF EXISTS "+
			quotedSchemaName+
			" CASCADE",
	); err != nil {
		t.Errorf(
			"drop PostgreSQL test schema: %v",
			err,
		)
	}
}

func makeAltitudePersistenceState(
	flightID string,
	icao24 string,
	observedAt time.Time,
	barometricValue float64,
	barometricStatus flightstate.AltitudeStatus,
	geometricValue float64,
	geometricStatus flightstate.AltitudeStatus,
	onGround bool,
) flightstate.FlightState {
	return flightstate.FlightState{
		FlightID:                 flightID,
		ICAO24:                   icao24,
		Callsign:                 "AHY101",
		Latitude:                 40.4093,
		Longitude:                49.8671,
		BarometricAltitudeM:      barometricValue,
		BarometricAltitudeStatus: barometricStatus,
		GeometricAltitudeM:       geometricValue,
		GeometricAltitudeStatus:  geometricStatus,
		VelocityMPS:              220,
		HeadingDegrees:           90,
		VerticalRateMPS:          0,
		OnGround:                 onGround,
		OriginCountry:            "Azerbaijan",
		ObservedAt:               observedAt,
		SourceName:               "airplanes.live",
	}
}

func assertPersistedAltitudeNullability(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	rows, err := pool.Query(
		context.Background(),
		`
			SELECT
				barometric_altitude_m,
				barometric_altitude_status,
				geometric_altitude_m,
				geometric_altitude_status
			FROM flight_states
			WHERE icao24 = 'ABC123'
			ORDER BY observed_at ASC
		`,
	)
	if err != nil {
		t.Fatalf(
			"query persisted altitude nullability: %v",
			err,
		)
	}
	defer rows.Close()

	expectedBarometricValid := []bool{
		true,
		true,
		false,
		false,
		false,
	}
	expectedGeometricValid := []bool{
		true,
		false,
		false,
		false,
		false,
	}

	index := 0

	for rows.Next() {
		if index >= len(expectedBarometricValid) {
			t.Fatal(
				"received more persisted altitude rows than expected",
			)
		}

		var barometricValue pgtype.Int4
		var barometricStatus string
		var geometricValue pgtype.Int4
		var geometricStatus string

		if err := rows.Scan(
			&barometricValue,
			&barometricStatus,
			&geometricValue,
			&geometricStatus,
		); err != nil {
			t.Fatalf(
				"scan persisted altitude nullability: %v",
				err,
			)
		}

		if barometricValue.Valid != expectedBarometricValid[index] {
			t.Fatalf(
				"row %d expected barometric valid=%v, got %v",
				index,
				expectedBarometricValid[index],
				barometricValue.Valid,
			)
		}

		if geometricValue.Valid != expectedGeometricValid[index] {
			t.Fatalf(
				"row %d expected geometric valid=%v, got %v",
				index,
				expectedGeometricValid[index],
				geometricValue.Valid,
			)
		}

		index++
	}

	if err := rows.Err(); err != nil {
		t.Fatalf(
			"iterate persisted altitude nullability rows: %v",
			err,
		)
	}

	if index != len(expectedBarometricValid) {
		t.Fatalf(
			"expected %d persisted rows, got %d",
			len(expectedBarometricValid),
			index,
		)
	}
}

func assertInt4Value(
	t *testing.T,
	field string,
	actual pgtype.Int4,
	expectedValid bool,
	expectedValue int32,
) {
	t.Helper()

	if actual.Valid != expectedValid {
		t.Fatalf(
			"%s expected valid=%v, got %v",
			field,
			expectedValid,
			actual.Valid,
		)
	}

	if expectedValid &&
		actual.Int32 != expectedValue {
		t.Fatalf(
			"%s expected value %d, got %d",
			field,
			expectedValue,
			actual.Int32,
		)
	}
}
