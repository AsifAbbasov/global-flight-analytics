package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	airportRepositoryTestDatabaseURLEnvironmentVariable = "TEST_DATABASE_URL"
	airportRepositoryElevationTolerance                 = 1e-9
)

var airportRepositorySchemaCounter uint64

type airportRepositoryIntegrationFixture struct {
	pool       *pgxpool.Pool
	repository *AirportRepository
}

func TestAirportRepositoryListMapsRowsConvertsElevationAndDoesNotTruncateAtOneHundred(
	t *testing.T,
) {
	fixture := newAirportRepositoryIntegrationFixture(
		t,
	)

	ctx := context.Background()

	insertAirportRepositoryCountry(
		t,
		ctx,
		fixture.pool,
		"00000000-0000-0000-0000-000000000001",
		"Azerbaijan",
	)

	insertAirportRepositoryAirport(
		t,
		ctx,
		fixture.pool,
		airportRepositoryAirportSeed{
			ID:          "00000000-0000-0000-0000-000000000002",
			ICAOCode:    "UBBB",
			IATACode:    "GYD",
			Name:        "Heydar Aliyev International Airport",
			City:        "Baku",
			CountryID:   "00000000-0000-0000-0000-000000000001",
			Latitude:    40.4675,
			Longitude:   50.0467,
			ElevationFT: 1000,
			Timezone:    "Asia/Baku",
			Description: "Primary Azerbaijan test airport",
		},
	)

	for index := 0; index < 100; index++ {
		insertAirportRepositoryAirport(
			t,
			ctx,
			fixture.pool,
			airportRepositoryAirportSeed{
				ID: fmt.Sprintf(
					"00000000-0000-0000-0001-%012x",
					index,
				),
				ICAOCode: fmt.Sprintf(
					"T%03d",
					index,
				),
				IATACode: fmt.Sprintf(
					"%03d",
					index,
				),
				Name: fmt.Sprintf(
					"Synthetic Airport %03d",
					index,
				),
				City:        "Synthetic City",
				CountryID:   "00000000-0000-0000-0000-000000000001",
				Latitude:    40.0 + float64(index)/1000.0,
				Longitude:   49.0 + float64(index)/1000.0,
				ElevationFT: index,
				Timezone:    "UTC",
			},
		)
	}

	airports, err := fixture.repository.List(
		ctx,
	)
	if err != nil {
		t.Fatalf(
			"list airports: %v",
			err,
		)
	}

	if len(airports) != 101 {
		t.Fatalf(
			"unexpected airport count: got %d, want 101",
			len(airports),
		)
	}

	var mappedAirportFound bool

	for _, item := range airports {
		if item.ICAOCode != "UBBB" {
			continue
		}

		mappedAirportFound = true

		if item.IATACode != "GYD" {
			t.Fatalf(
				"unexpected IATA code: got %q, want %q",
				item.IATACode,
				"GYD",
			)
		}

		if item.Name != "Heydar Aliyev International Airport" {
			t.Fatalf(
				"unexpected airport name: got %q",
				item.Name,
			)
		}

		if item.City != "Baku" {
			t.Fatalf(
				"unexpected city: got %q, want %q",
				item.City,
				"Baku",
			)
		}

		if item.Country != "Azerbaijan" {
			t.Fatalf(
				"unexpected country: got %q, want %q",
				item.Country,
				"Azerbaijan",
			)
		}

		assertAirportRepositoryFloatClose(
			t,
			item.Latitude,
			40.4675,
		)

		assertAirportRepositoryFloatClose(
			t,
			item.Longitude,
			50.0467,
		)

		assertAirportRepositoryFloatClose(
			t,
			item.ElevationM,
			304.8,
		)
		if !item.ElevationAvailable {
			t.Fatal("expected mapped airport elevation to be available")
		}

		if item.Timezone != "Asia/Baku" {
			t.Fatalf(
				"unexpected timezone: got %q, want %q",
				item.Timezone,
				"Asia/Baku",
			)
		}

		if item.Description != "Primary Azerbaijan test airport" {
			t.Fatalf(
				"unexpected description: got %q",
				item.Description,
			)
		}
	}

	if !mappedAirportFound {
		t.Fatal(
			"expected UBBB airport in repository list result",
		)
	}
}

func TestAirportRepositoryGetByICAOMapsRowAndConvertsElevation(
	t *testing.T,
) {
	fixture := newAirportRepositoryIntegrationFixture(
		t,
	)

	ctx := context.Background()

	insertAirportRepositoryCountry(
		t,
		ctx,
		fixture.pool,
		"00000000-0000-0000-0000-000000000011",
		"Georgia",
	)

	insertAirportRepositoryAirport(
		t,
		ctx,
		fixture.pool,
		airportRepositoryAirportSeed{
			ID:          "00000000-0000-0000-0000-000000000012",
			ICAOCode:    "UGTB",
			IATACode:    "TBS",
			Name:        "Tbilisi International Airport",
			City:        "Tbilisi",
			CountryID:   "00000000-0000-0000-0000-000000000011",
			Latitude:    41.6692,
			Longitude:   44.9547,
			ElevationFT: 1624,
			Timezone:    "Asia/Tbilisi",
			Description: "Georgia repository path test airport",
		},
	)

	item, err := fixture.repository.GetByICAO(
		ctx,
		"UGTB",
	)
	if err != nil {
		t.Fatalf(
			"get airport by ICAO: %v",
			err,
		)
	}

	if item.ICAOCode != "UGTB" {
		t.Fatalf(
			"unexpected ICAO code: got %q, want %q",
			item.ICAOCode,
			"UGTB",
		)
	}

	if item.IATACode != "TBS" {
		t.Fatalf(
			"unexpected IATA code: got %q, want %q",
			item.IATACode,
			"TBS",
		)
	}

	if item.Country != "Georgia" {
		t.Fatalf(
			"unexpected country: got %q, want %q",
			item.Country,
			"Georgia",
		)
	}

	assertAirportRepositoryFloatClose(
		t,
		item.ElevationM,
		494.9952,
	)
	if !item.ElevationAvailable {
		t.Fatal("expected airport elevation to be available")
	}

	if item.Description != "Georgia repository path test airport" {
		t.Fatalf(
			"unexpected description: got %q",
			item.Description,
		)
	}
}

func TestAirportRepositoryDistinguishesUnknownElevationFromObservedSeaLevel(
	t *testing.T,
) {
	fixture := newAirportRepositoryIntegrationFixture(t)
	ctx := context.Background()

	for _, row := range []struct {
		id        string
		icao      string
		elevation any
	}{
		{id: "00000000-0000-0000-0000-000000000021", icao: "NULL", elevation: nil},
		{id: "00000000-0000-0000-0000-000000000022", icao: "ZERO", elevation: 0},
	} {
		_, err := fixture.pool.Exec(
			ctx,
			`INSERT INTO airports (id, icao_code, name, latitude, longitude, elevation_ft) VALUES ($1, $2, $3, 0, 0, $4)`,
			row.id,
			row.icao,
			row.icao+" airport",
			row.elevation,
		)
		if err != nil {
			t.Fatalf("insert elevation semantics airport %s: %v", row.icao, err)
		}
	}

	unknown, err := fixture.repository.GetByICAO(ctx, "NULL")
	if err != nil {
		t.Fatalf("get unknown-elevation airport: %v", err)
	}
	if unknown.ElevationAvailable || unknown.ElevationM != 0 {
		t.Fatalf("unknown elevation was not preserved: %#v", unknown)
	}

	seaLevel, err := fixture.repository.GetByICAO(ctx, "ZERO")
	if err != nil {
		t.Fatalf("get sea-level airport: %v", err)
	}
	if !seaLevel.ElevationAvailable || seaLevel.ElevationM != 0 {
		t.Fatalf("observed sea-level elevation was not preserved: %#v", seaLevel)
	}
}

func TestAirportRepositoryGetByICAOReturnsAirportNotFound(
	t *testing.T,
) {
	fixture := newAirportRepositoryIntegrationFixture(
		t,
	)

	_, err := fixture.repository.GetByICAO(
		context.Background(),
		"ZZZZ",
	)
	if err == nil {
		t.Fatal(
			"expected missing airport error",
		)
	}

	if !errors.Is(
		err,
		airport.ErrNotFound,
	) {
		t.Fatalf(
			"expected airport.ErrNotFound, got %v",
			err,
		)
	}
}

func TestAirportRepositoryListPropagatesDatabaseError(
	t *testing.T,
) {
	fixture := newAirportRepositoryIntegrationFixture(
		t,
	)

	ctx := context.Background()

	if _, err := fixture.pool.Exec(
		ctx,
		"DROP TABLE airports CASCADE",
	); err != nil {
		t.Fatalf(
			"drop airports table: %v",
			err,
		)
	}

	_, err := fixture.repository.List(
		ctx,
	)
	if err == nil {
		t.Fatal(
			"expected repository list database error",
		)
	}
}

func newAirportRepositoryIntegrationFixture(
	t *testing.T,
) *airportRepositoryIntegrationFixture {
	t.Helper()

	databaseURL := strings.TrimSpace(
		os.Getenv(
			airportRepositoryTestDatabaseURLEnvironmentVariable,
		),
	)
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			airportRepositoryTestDatabaseURLEnvironmentVariable,
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
		"airport_repository_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(
			&airportRepositorySchemaCounter,
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
		dropAirportRepositoryTestSchema(
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
		dropAirportRepositoryTestSchema(
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

		dropAirportRepositoryTestSchema(
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

	createAirportRepositoryTestTables(
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

	return &airportRepositoryIntegrationFixture{
		pool: pool,
		repository: NewAirportRepository(
			pool,
		),
	}
}

func createAirportRepositoryTestTables(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
) {
	t.Helper()

	statements := []string{
		`
			CREATE TABLE countries (
				id uuid PRIMARY KEY,
				name text NOT NULL
			)
		`,
		`
			CREATE TABLE airports (
				id uuid PRIMARY KEY,
				icao_code varchar(4) UNIQUE,
				iata_code varchar(3),
				name text NOT NULL,
				city text,
				country_id uuid REFERENCES countries(id) ON DELETE SET NULL,
				latitude numeric NOT NULL,
				longitude numeric NOT NULL,
				elevation_ft integer,
				timezone text
			)
		`,
		`
			CREATE TABLE airport_profiles (
				airport_id uuid PRIMARY KEY REFERENCES airports(id) ON DELETE CASCADE,
				description text
			)
		`,
	}

	for index, statement := range statements {
		if _, err := pool.Exec(
			ctx,
			statement,
		); err != nil {
			t.Fatalf(
				"create airport repository test table at statement %d: %v",
				index,
				err,
			)
		}
	}
}

func dropAirportRepositoryTestSchema(
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

type airportRepositoryAirportSeed struct {
	ID          string
	ICAOCode    string
	IATACode    string
	Name        string
	City        string
	CountryID   string
	Latitude    float64
	Longitude   float64
	ElevationFT int
	Timezone    string
	Description string
}

func insertAirportRepositoryCountry(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
	name string,
) {
	t.Helper()

	if _, err := pool.Exec(
		ctx,
		`
			INSERT INTO countries (
				id,
				name
			)
			VALUES ($1, $2)
		`,
		id,
		name,
	); err != nil {
		t.Fatalf(
			"insert airport repository test country: %v",
			err,
		)
	}
}

func insertAirportRepositoryAirport(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	seed airportRepositoryAirportSeed,
) {
	t.Helper()

	if _, err := pool.Exec(
		ctx,
		`
			INSERT INTO airports (
				id,
				icao_code,
				iata_code,
				name,
				city,
				country_id,
				latitude,
				longitude,
				elevation_ft,
				timezone
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
		seed.ID,
		seed.ICAOCode,
		seed.IATACode,
		seed.Name,
		seed.City,
		seed.CountryID,
		seed.Latitude,
		seed.Longitude,
		seed.ElevationFT,
		seed.Timezone,
	); err != nil {
		t.Fatalf(
			"insert airport repository test airport %q: %v",
			seed.ICAOCode,
			err,
		)
	}

	if seed.Description == "" {
		return
	}

	if _, err := pool.Exec(
		ctx,
		`
			INSERT INTO airport_profiles (
				airport_id,
				description
			)
			VALUES ($1, $2)
		`,
		seed.ID,
		seed.Description,
	); err != nil {
		t.Fatalf(
			"insert airport repository test profile %q: %v",
			seed.ICAOCode,
			err,
		)
	}
}

func assertAirportRepositoryFloatClose(
	t *testing.T,
	actual float64,
	expected float64,
) {
	t.Helper()

	if math.Abs(actual-expected) >
		airportRepositoryElevationTolerance {
		t.Fatalf(
			"unexpected float value: got %.12f, want %.12f",
			actual,
			expected,
		)
	}
}
