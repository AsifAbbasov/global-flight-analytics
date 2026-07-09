package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const airportImportRepositoryTestDatabaseURLEnvironmentVariable = "TEST_DATABASE_URL"

var airportImportRepositorySchemaCounter uint64

type airportImportRepositoryIntegrationFixture struct {
	pool       *pgxpool.Pool
	repository *AirportRepository
}

func TestAirportRepositoryUpsertImportedInsertsSourceBackedAirportAndResolvesCountry(
	t *testing.T,
) {
	fixture := newAirportImportRepositoryIntegrationFixture(
		t,
	)

	ctx := context.Background()

	insertAirportImportCountry(
		t,
		ctx,
		fixture.pool,
		"00000000-0000-0000-0000-000000000101",
		"AZ",
		"Azerbaijan",
	)

	lastSyncedAt := time.Date(
		2026,
		time.July,
		9,
		8,
		30,
		0,
		0,
		time.UTC,
	)

	elevationFeet := 10

	reconciledCount, err := fixture.repository.UpsertImported(
		ctx,
		[]airport.ImportRecord{
			{
				SourceIdent:       "ourairports-ubbb",
				ICAOCode:          "UBBB",
				IATACode:          "GYD",
				Name:              "Heydar Aliyev International Airport",
				City:              "Baku",
				SourceCountryCode: "AZ",
				Latitude:          40.4675,
				Longitude:         50.0467,
				ElevationFT:       &elevationFeet,
				SourceName:        "ourairports",
				LastSyncedAt:      lastSyncedAt,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"upsert imported airport: %v",
			err,
		)
	}

	if reconciledCount != 1 {
		t.Fatalf(
			"unexpected reconciled count: got %d, want 1",
			reconciledCount,
		)
	}

	var (
		sourceIdent       string
		icaoCode          string
		iataCode          string
		name              string
		city              string
		countryID         string
		sourceCountryCode string
		latitude          float64
		longitude         float64
		elevationFT       int
		sourceName        string
		actualSyncedAt    time.Time
	)

	err = fixture.pool.QueryRow(
		ctx,
		`
			SELECT
				source_ident,
				COALESCE(icao_code, ''),
				COALESCE(iata_code, ''),
				name,
				COALESCE(city, ''),
				COALESCE(country_id::text, ''),
				COALESCE(source_country_code, ''),
				latitude,
				longitude,
				COALESCE(elevation_ft, 0),
				source_name,
				last_synced_at
			FROM airports
			WHERE source_name = $1
				AND source_ident = $2
		`,
		"ourairports",
		"ourairports-ubbb",
	).Scan(
		&sourceIdent,
		&icaoCode,
		&iataCode,
		&name,
		&city,
		&countryID,
		&sourceCountryCode,
		&latitude,
		&longitude,
		&elevationFT,
		&sourceName,
		&actualSyncedAt,
	)
	if err != nil {
		t.Fatalf(
			"query inserted imported airport: %v",
			err,
		)
	}

	if sourceIdent != "ourairports-ubbb" {
		t.Fatalf(
			"unexpected source identity: got %q",
			sourceIdent,
		)
	}

	if icaoCode != "UBBB" {
		t.Fatalf(
			"unexpected ICAO code: got %q",
			icaoCode,
		)
	}

	if iataCode != "GYD" {
		t.Fatalf(
			"unexpected IATA code: got %q",
			iataCode,
		)
	}

	if name != "Heydar Aliyev International Airport" {
		t.Fatalf(
			"unexpected airport name: got %q",
			name,
		)
	}

	if city != "Baku" {
		t.Fatalf(
			"unexpected airport city: got %q",
			city,
		)
	}

	if countryID != "00000000-0000-0000-0000-000000000101" {
		t.Fatalf(
			"unexpected resolved country identifier: got %q",
			countryID,
		)
	}

	if sourceCountryCode != "AZ" {
		t.Fatalf(
			"unexpected source country code: got %q",
			sourceCountryCode,
		)
	}

	if latitude != 40.4675 {
		t.Fatalf(
			"unexpected latitude: got %f",
			latitude,
		)
	}

	if longitude != 50.0467 {
		t.Fatalf(
			"unexpected longitude: got %f",
			longitude,
		)
	}

	if elevationFT != 10 {
		t.Fatalf(
			"unexpected elevation: got %d",
			elevationFT,
		)
	}

	if sourceName != "ourairports" {
		t.Fatalf(
			"unexpected source name: got %q",
			sourceName,
		)
	}

	if !actualSyncedAt.Equal(
		lastSyncedAt,
	) {
		t.Fatalf(
			"unexpected last_synced_at: got %s, want %s",
			actualSyncedAt,
			lastSyncedAt,
		)
	}
}

func TestAirportRepositoryUpsertImportedUpdatesBySourceIdentityWithoutCreatingDuplicate(
	t *testing.T,
) {
	fixture := newAirportImportRepositoryIntegrationFixture(
		t,
	)

	ctx := context.Background()

	initialSyncedAt := time.Date(
		2026,
		time.July,
		9,
		9,
		0,
		0,
		0,
		time.UTC,
	)

	initialElevation := 100

	_, err := fixture.repository.UpsertImported(
		ctx,
		[]airport.ImportRecord{
			{
				SourceIdent:       "stable-source-id",
				ICAOCode:          "UBBB",
				IATACode:          "GYD",
				Name:              "Initial Airport Name",
				City:              "Initial City",
				SourceCountryCode: "AZ",
				Latitude:          40.0,
				Longitude:         49.0,
				ElevationFT:       &initialElevation,
				SourceName:        "ourairports",
				LastSyncedAt:      initialSyncedAt,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"insert initial imported airport: %v",
			err,
		)
	}

	updatedSyncedAt := initialSyncedAt.Add(
		time.Hour,
	)

	reconciledCount, err := fixture.repository.UpsertImported(
		ctx,
		[]airport.ImportRecord{
			{
				SourceIdent:       "stable-source-id",
				ICAOCode:          "UBBF",
				IATACode:          "NEW",
				Name:              "Updated Airport Name",
				City:              "Updated City",
				SourceCountryCode: "GE",
				Latitude:          41.5,
				Longitude:         44.8,
				ElevationFT:       nil,
				SourceName:        "ourairports",
				LastSyncedAt:      updatedSyncedAt,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"update imported airport by source identity: %v",
			err,
		)
	}

	if reconciledCount != 1 {
		t.Fatalf(
			"unexpected reconciled count: got %d, want 1",
			reconciledCount,
		)
	}

	assertAirportImportRepositoryRowCount(
		t,
		ctx,
		fixture.pool,
		1,
	)

	var (
		icaoCode       string
		iataCode       string
		name           string
		city           string
		latitude       float64
		longitude      float64
		elevationIsNil bool
		actualSyncedAt time.Time
	)

	err = fixture.pool.QueryRow(
		ctx,
		`
			SELECT
				COALESCE(icao_code, ''),
				COALESCE(iata_code, ''),
				name,
				COALESCE(city, ''),
				latitude,
				longitude,
				elevation_ft IS NULL,
				last_synced_at
			FROM airports
			WHERE source_name = $1
				AND source_ident = $2
		`,
		"ourairports",
		"stable-source-id",
	).Scan(
		&icaoCode,
		&iataCode,
		&name,
		&city,
		&latitude,
		&longitude,
		&elevationIsNil,
		&actualSyncedAt,
	)
	if err != nil {
		t.Fatalf(
			"query updated imported airport: %v",
			err,
		)
	}

	if icaoCode != "UBBF" {
		t.Fatalf(
			"unexpected updated ICAO code: got %q",
			icaoCode,
		)
	}

	if iataCode != "NEW" {
		t.Fatalf(
			"unexpected updated IATA code: got %q",
			iataCode,
		)
	}

	if name != "Updated Airport Name" {
		t.Fatalf(
			"unexpected updated airport name: got %q",
			name,
		)
	}

	if city != "Updated City" {
		t.Fatalf(
			"unexpected updated airport city: got %q",
			city,
		)
	}

	if latitude != 41.5 {
		t.Fatalf(
			"unexpected updated latitude: got %f",
			latitude,
		)
	}

	if longitude != 44.8 {
		t.Fatalf(
			"unexpected updated longitude: got %f",
			longitude,
		)
	}

	if !elevationIsNil {
		t.Fatal(
			"expected nil elevation after source-identity update",
		)
	}

	if !actualSyncedAt.Equal(
		updatedSyncedAt,
	) {
		t.Fatalf(
			"unexpected updated last_synced_at: got %s, want %s",
			actualSyncedAt,
			updatedSyncedAt,
		)
	}
}

func TestAirportRepositoryUpsertImportedReconcilesExistingAirportByICAOAndPreservesRowIdentity(
	t *testing.T,
) {
	fixture := newAirportImportRepositoryIntegrationFixture(
		t,
	)

	ctx := context.Background()

	const existingAirportID = "00000000-0000-0000-0000-000000000201"

	insertAirportImportExistingAirport(
		t,
		ctx,
		fixture.pool,
		existingAirportID,
		"legacy-source-id",
		"UBBB",
		"OLD",
		"Legacy Airport Name",
		"Legacy City",
		40.1,
		49.1,
		"legacy",
		time.Date(
			2026,
			time.July,
			8,
			8,
			0,
			0,
			0,
			time.UTC,
		),
	)

	reconciledCount, err := fixture.repository.UpsertImported(
		ctx,
		[]airport.ImportRecord{
			{
				SourceIdent:       "ourairports-ubbb",
				ICAOCode:          "UBBB",
				IATACode:          "GYD",
				Name:              "Reconciled Airport Name",
				City:              "Baku",
				SourceCountryCode: "AZ",
				Latitude:          40.4675,
				Longitude:         50.0467,
				SourceName:        "ourairports",
				LastSyncedAt: time.Date(
					2026,
					time.July,
					9,
					10,
					0,
					0,
					0,
					time.UTC,
				),
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"reconcile existing airport by ICAO: %v",
			err,
		)
	}

	if reconciledCount != 1 {
		t.Fatalf(
			"unexpected reconciled count: got %d, want 1",
			reconciledCount,
		)
	}

	assertAirportImportRepositoryRowCount(
		t,
		ctx,
		fixture.pool,
		1,
	)

	var (
		actualID    string
		sourceName  string
		sourceIdent string
		iataCode    string
		name        string
	)

	err = fixture.pool.QueryRow(
		ctx,
		`
			SELECT
				id::text,
				source_name,
				source_ident,
				COALESCE(iata_code, ''),
				name
			FROM airports
			WHERE icao_code = $1
		`,
		"UBBB",
	).Scan(
		&actualID,
		&sourceName,
		&sourceIdent,
		&iataCode,
		&name,
	)
	if err != nil {
		t.Fatalf(
			"query reconciled airport: %v",
			err,
		)
	}

	if actualID != existingAirportID {
		t.Fatalf(
			"airport row identity changed during ICAO reconciliation: got %q, want %q",
			actualID,
			existingAirportID,
		)
	}

	if sourceName != "ourairports" {
		t.Fatalf(
			"unexpected reconciled source name: got %q",
			sourceName,
		)
	}

	if sourceIdent != "ourairports-ubbb" {
		t.Fatalf(
			"unexpected reconciled source identity: got %q",
			sourceIdent,
		)
	}

	if iataCode != "GYD" {
		t.Fatalf(
			"unexpected reconciled IATA code: got %q",
			iataCode,
		)
	}

	if name != "Reconciled Airport Name" {
		t.Fatalf(
			"unexpected reconciled airport name: got %q",
			name,
		)
	}
}

func TestAirportRepositoryUpsertImportedIsIdempotentForRepeatedSourceIdentity(
	t *testing.T,
) {
	fixture := newAirportImportRepositoryIntegrationFixture(
		t,
	)

	ctx := context.Background()

	record := airport.ImportRecord{
		SourceIdent:       "repeatable-source-id",
		ICAOCode:          "UGTB",
		IATACode:          "TBS",
		Name:              "Tbilisi International Airport",
		City:              "Tbilisi",
		SourceCountryCode: "GE",
		Latitude:          41.6692,
		Longitude:         44.9547,
		SourceName:        "ourairports",
		LastSyncedAt: time.Date(
			2026,
			time.July,
			9,
			11,
			0,
			0,
			0,
			time.UTC,
		),
	}

	for attempt := 0; attempt < 2; attempt++ {
		reconciledCount, err := fixture.repository.UpsertImported(
			ctx,
			[]airport.ImportRecord{
				record,
			},
		)
		if err != nil {
			t.Fatalf(
				"repeated upsert attempt %d: %v",
				attempt+1,
				err,
			)
		}

		if reconciledCount != 1 {
			t.Fatalf(
				"unexpected reconciled count on attempt %d: got %d, want 1",
				attempt+1,
				reconciledCount,
			)
		}
	}

	assertAirportImportRepositoryRowCount(
		t,
		ctx,
		fixture.pool,
		1,
	)
}

func TestAirportRepositoryUpsertImportedDuplicateSourceIdentityBatchRollsBackAtomically(
	t *testing.T,
) {
	fixture := newAirportImportRepositoryIntegrationFixture(
		t,
	)

	ctx := context.Background()

	_, err := fixture.repository.UpsertImported(
		ctx,
		[]airport.ImportRecord{
			{
				SourceIdent:  "duplicate-source-id",
				ICAOCode:     "UBBB",
				IATACode:     "GYD",
				Name:         "First Duplicate Candidate",
				Latitude:     40.4675,
				Longitude:    50.0467,
				SourceName:   "ourairports",
				LastSyncedAt: time.Now().UTC(),
			},
			{
				SourceIdent:  "duplicate-source-id",
				ICAOCode:     "UGTB",
				IATACode:     "TBS",
				Name:         "Second Duplicate Candidate",
				Latitude:     41.6692,
				Longitude:    44.9547,
				SourceName:   "ourairports",
				LastSyncedAt: time.Now().UTC(),
			},
		},
	)
	if err == nil {
		t.Fatal(
			"expected duplicate source identity batch error",
		)
	}

	assertAirportImportRepositoryRowCount(
		t,
		ctx,
		fixture.pool,
		0,
	)
}

func TestAirportRepositoryUpsertImportedCoordinateConstraintFailureRollsBackWholeBatch(
	t *testing.T,
) {
	fixture := newAirportImportRepositoryIntegrationFixture(
		t,
	)

	ctx := context.Background()

	_, err := fixture.repository.UpsertImported(
		ctx,
		[]airport.ImportRecord{
			{
				SourceIdent:  "valid-before-invalid",
				ICAOCode:     "UBBB",
				IATACode:     "GYD",
				Name:         "Valid Airport Candidate",
				Latitude:     40.4675,
				Longitude:    50.0467,
				SourceName:   "ourairports",
				LastSyncedAt: time.Now().UTC(),
			},
			{
				SourceIdent:  "invalid-coordinate",
				ICAOCode:     "TEST",
				IATACode:     "BAD",
				Name:         "Invalid Coordinate Candidate",
				Latitude:     91,
				Longitude:    49,
				SourceName:   "ourairports",
				LastSyncedAt: time.Now().UTC(),
			},
		},
	)
	if err == nil {
		t.Fatal(
			"expected airport coordinate constraint error",
		)
	}

	assertAirportImportRepositoryRowCount(
		t,
		ctx,
		fixture.pool,
		0,
	)
}

func TestAirportRepositoryUpsertImportedBlankSourceIdentityConstraintRollsBack(
	t *testing.T,
) {
	fixture := newAirportImportRepositoryIntegrationFixture(
		t,
	)

	ctx := context.Background()

	_, err := fixture.repository.UpsertImported(
		ctx,
		[]airport.ImportRecord{
			{
				SourceIdent:  "   ",
				ICAOCode:     "UBBB",
				IATACode:     "GYD",
				Name:         "Blank Source Identity Candidate",
				Latitude:     40.4675,
				Longitude:    50.0467,
				SourceName:   "ourairports",
				LastSyncedAt: time.Now().UTC(),
			},
		},
	)
	if err == nil {
		t.Fatal(
			"expected blank source identity constraint error",
		)
	}

	assertAirportImportRepositoryRowCount(
		t,
		ctx,
		fixture.pool,
		0,
	)
}

func newAirportImportRepositoryIntegrationFixture(
	t *testing.T,
) *airportImportRepositoryIntegrationFixture {
	t.Helper()

	databaseURL := strings.TrimSpace(
		os.Getenv(
			airportImportRepositoryTestDatabaseURLEnvironmentVariable,
		),
	)
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			airportImportRepositoryTestDatabaseURLEnvironmentVariable,
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
		"airport_import_repository_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(
			&airportImportRepositorySchemaCounter,
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
		dropAirportImportRepositoryTestSchema(
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
		dropAirportImportRepositoryTestSchema(
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

		dropAirportImportRepositoryTestSchema(
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

	createAirportImportRepositoryTestTables(
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

	return &airportImportRepositoryIntegrationFixture{
		pool: pool,
		repository: NewAirportRepository(
			pool,
		),
	}
}

func createAirportImportRepositoryTestTables(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
) {
	t.Helper()

	statements := []string{
		`
			CREATE TABLE countries (
				id uuid PRIMARY KEY,
				iso2 varchar(2) NOT NULL UNIQUE,
				name text NOT NULL
			)
		`,
		`
			CREATE TABLE airports (
				id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
				icao_code varchar(4) UNIQUE,
				iata_code varchar(3),
				name text NOT NULL,
				city text,
				country_id uuid REFERENCES countries(id) ON DELETE SET NULL,
				latitude numeric NOT NULL,
				longitude numeric NOT NULL,
				elevation_ft integer,
				timezone text,
				source_name text NOT NULL,
				last_synced_at timestamptz,
				created_at timestamptz NOT NULL DEFAULT now(),
				updated_at timestamptz NOT NULL DEFAULT now(),
				source_ident text,
				source_country_code varchar(2),

				CONSTRAINT airports_coordinates_check
					CHECK (
						latitude >= -90
						AND latitude <= 90
						AND longitude >= -180
						AND longitude <= 180
					),

				CONSTRAINT airports_identifier_check
					CHECK (
						source_ident IS NOT NULL
						OR icao_code IS NOT NULL
						OR iata_code IS NOT NULL
					),

				CONSTRAINT airports_source_ident_check
					CHECK (
						source_ident IS NULL
						OR btrim(source_ident) <> ''
					),

				CONSTRAINT airports_source_country_code_check
					CHECK (
						source_country_code IS NULL
						OR char_length(source_country_code) = 2
					)
			)
		`,
		`
			CREATE UNIQUE INDEX airports_source_identity_unique
				ON airports (source_name, source_ident)
				WHERE source_ident IS NOT NULL
		`,
		`
			CREATE INDEX airports_source_country_code_idx
				ON airports (source_country_code)
				WHERE source_country_code IS NOT NULL
		`,
	}

	for index, statement := range statements {
		if _, err := pool.Exec(
			ctx,
			statement,
		); err != nil {
			t.Fatalf(
				"create airport import repository test schema at statement %d: %v",
				index,
				err,
			)
		}
	}
}

func dropAirportImportRepositoryTestSchema(
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
			"drop PostgreSQL test schema after setup failure: %v",
			err,
		)
	}
}

func insertAirportImportCountry(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
	iso2 string,
	name string,
) {
	t.Helper()

	if _, err := pool.Exec(
		ctx,
		`
			INSERT INTO countries (
				id,
				iso2,
				name
			)
			VALUES ($1, $2, $3)
		`,
		id,
		iso2,
		name,
	); err != nil {
		t.Fatalf(
			"insert airport import test country: %v",
			err,
		)
	}
}

func insertAirportImportExistingAirport(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
	sourceIdent string,
	icaoCode string,
	iataCode string,
	name string,
	city string,
	latitude float64,
	longitude float64,
	sourceName string,
	lastSyncedAt time.Time,
) {
	t.Helper()

	if _, err := pool.Exec(
		ctx,
		`
			INSERT INTO airports (
				id,
				source_ident,
				icao_code,
				iata_code,
				name,
				city,
				latitude,
				longitude,
				source_name,
				last_synced_at
			)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9,
				$10
			)
		`,
		id,
		sourceIdent,
		icaoCode,
		iataCode,
		name,
		city,
		latitude,
		longitude,
		sourceName,
		lastSyncedAt,
	); err != nil {
		t.Fatalf(
			"insert existing airport import test row: %v",
			err,
		)
	}
}

func assertAirportImportRepositoryRowCount(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	expected int,
) {
	t.Helper()

	var actual int

	if err := pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM airports",
	).Scan(
		&actual,
	); err != nil {
		t.Fatalf(
			"count airport import test rows: %v",
			err,
		)
	}

	if actual != expected {
		t.Fatalf(
			"unexpected airport row count: got %d, want %d",
			actual,
			expected,
		)
	}
}
